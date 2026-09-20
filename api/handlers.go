// Package api 提供 RESTful API 处理器
package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/twinsant/sugarland/kernel"
)

// Handler 封装 API 处理器，持有 World 引用
type Handler struct {
	World *kernel.World
}

// NewHandler 创建新的 API 处理器
func NewHandler(w *kernel.World) *Handler {
	return &Handler{World: w}
}

// RegisterRoutes 注册所有 API 路由
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/world", h.handleWorld)
	mux.HandleFunc("/api/world/step", h.handleWorldStep)
	mux.HandleFunc("/api/world/reset", h.handleWorldReset)
	mux.HandleFunc("/api/world/agent", h.handleWorldAgent)
	mux.HandleFunc("/api/world/agent/detach", h.handleWorldAgentDetach)
	mux.HandleFunc("/api/snapshot", h.handleSnapshot)
	mux.HandleFunc("/api/citizens", h.handleCitizens)
	mux.HandleFunc("/api/citizens/", h.handleCitizens)
	mux.HandleFunc("/api/cellspace", h.handleCells)
	mux.HandleFunc("/api/cells/", h.handleRoomByCoord)
	mux.HandleFunc("/api/agent/attach", h.handleAgentAttach)
	mux.HandleFunc("/api/agent/detach", h.handleAgentDetach)
	mux.HandleFunc("/api/agent/context", h.handleAgentContextGet)
	mux.HandleFunc("/api/agent/", h.handleAgentContext)
	mux.HandleFunc("/api/citizens/command", h.handleCitizenCommand)
	mux.HandleFunc("/api/scoreboard", h.handleScoreboard)
	mux.HandleFunc("/api/ai/attach", h.handleAIAgentAttach)
	mux.HandleFunc("/api/ai/detach", h.handleAIAgentDetach)
}

// writeJSON 写入 JSON 响应
func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(data)
}

// writeError 写入错误响应
func writeError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// GET /api/world - 世界状态
func (h *Handler) handleWorld(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, h.World.GetState())
}

// POST /api/world/step - 推进一步
func (h *Handler) handleWorldStep(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	h.World.Step()
	writeJSON(w, h.World.GetState())
}

// POST /api/world/reset - 重置世界
// 可选请求体 {"config": {...}}：先合并配置再重建世界，
// 显式传入的键优先于 mudlib declare_defaults 的缺省值
func (h *Handler) handleWorldReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Config map[string]interface{} `json:"config"`
	}
	// 空 body 容忍：解码失败按无配置处理
	_ = json.NewDecoder(r.Body).Decode(&req)
	if len(req.Config) > 0 {
		h.World.ApplyConfig(req.Config)
	}
	h.World.Reset()
	writeJSON(w, h.World.GetState())
}

// GET /api/snapshot - 聚合世界快照（Godot 等渲染端一次取全）
func (h *Handler) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, h.World.GetSnapshot())
}

// WorldAgentRequest 世界级 LPC 脚本请求
type WorldAgentRequest struct {
	LPCSource string `json:"lpc_source"`
}

// /api/world/agent - 世界规则脚本管理
// GET：返回当前生效的世界 LPC 源码（默认 mudlib 或用户自定义）
// POST：用自定义 LPC 脚本整体替换世界规则
func (h *Handler) handleWorldAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		source, custom := h.World.GetWorldScript()
		writeJSON(w, map[string]interface{}{
			"object": "world_daemon",
			"source": source,
			"custom": custom,
		})
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req WorldAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.World.AttachWorldScript(req.LPCSource); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]interface{}{
		"status": "attached",
		"object": "world_daemon",
	})
}

// POST /api/world/agent/detach - 卸载世界级 LPC 守护脚本
func (h *Handler) handleWorldAgentDetach(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	h.World.DetachWorldScript()
	writeJSON(w, map[string]string{"status": "detached"})
}

// GET /api/citizens - 公民列表
func (h *Handler) handleCitizens(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/citizens")
	if path != "" && path != "/" {
		idStr := strings.TrimPrefix(path, "/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			writeError(w, "invalid citizen id", http.StatusBadRequest)
			return
		}
		citizen := h.World.GetCitizenByID(id)
		if citizen == nil {
			writeError(w, "citizen not found", http.StatusNotFound)
			return
		}
		writeJSON(w, citizen)
		return
	}
	writeJSON(w, h.World.GetCitizensList())
}

// GET /api/cellspace - 地图快照
func (h *Handler) handleCells(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	data, err := h.World.GetRoomsJSON()
	if err != nil {
		writeError(w, "failed to serialize cells", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// GET /api/cells/:x/:y - 单个格子状态
func (h *Handler) handleRoomByCoord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/cells/"), "/")
	if len(parts) != 2 {
		writeError(w, "invalid path, expected /api/cells/:x/:y", http.StatusBadRequest)
		return
	}
	x, err1 := strconv.Atoi(parts[0])
	y, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		writeError(w, "invalid coordinates", http.StatusBadRequest)
		return
	}
	cell := h.World.GetRoom(x, y)
	writeJSON(w, cell)
}

// AgentAttachRequest 绑定 agent 请求（LPC 脚本模式）
type AgentAttachRequest struct {
	CitizenID int    `json:"citizen_id"`
	LPCSource string `json:"lpc_source"`
	Interval  int    `json:"interval"`
}

// POST /api/agent/attach - 绑定 LPC agent 到 citizen
func (h *Handler) handleAgentAttach(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req AgentAttachRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Interval <= 0 {
		req.Interval = 1
	}
	err := h.World.AttachAgent(req.CitizenID, req.LPCSource, req.Interval)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]interface{}{
		"status":     "attached",
		"citizen_id": req.CitizenID,
		"interval":   req.Interval,
	})
}

// POST /api/agent/detach - 解绑 LPC agent
func (h *Handler) handleAgentDetach(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		CitizenID int `json:"citizen_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	err := h.World.DetachAgent(req.CitizenID)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]interface{}{
		"status":     "detached",
		"citizen_id": req.CitizenID,
	})
}

// GET /api/agent/context?id=:id - 获取 agent 上下文
func (h *Handler) handleAgentContextGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		writeError(w, "missing id parameter", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, "invalid agent id", http.StatusBadRequest)
		return
	}
	radius := -1
	if rs := r.URL.Query().Get("radius"); rs != "" {
		if rv, err := strconv.Atoi(rs); err == nil {
			radius = rv
		}
	}
	ctx, err := h.World.GetAgentContext(id, radius)
	if err != nil {
		writeError(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, ctx)
}

// GET /api/agent/:id/context - 获取 agent 上下文（路径风格）
func (h *Handler) handleAgentContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/agent/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "context" {
		writeError(w, "invalid path, expected /api/agent/:id/context", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		writeError(w, "invalid agent id", http.StatusBadRequest)
		return
	}
	radius := -1
	if rs := r.URL.Query().Get("radius"); rs != "" {
		if rv, err := strconv.Atoi(rs); err == nil {
			radius = rv
		}
	}
	ctx, err := h.World.GetAgentContext(id, radius)
	if err != nil {
		writeError(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, ctx)
}

// CitizenCommandRequest citizen 指令请求
type CitizenCommandRequest struct {
	CitizenID int                    `json:"citizen_id"`
	Action    string                 `json:"action"`
	Args      map[string]interface{} `json:"args"`
}

// POST /api/citizens/command - 给被接管的 citizen 下达指令
func (h *Handler) handleCitizenCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req CitizenCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Args == nil {
		req.Args = make(map[string]interface{})
	}
	err := h.World.ExecuteCommand(req.CitizenID, req.Action, req.Args)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	citizen := h.World.GetCitizenByID(req.CitizenID)
	writeJSON(w, map[string]interface{}{
		"status":     "executed",
		"citizen_id": req.CitizenID,
		"action":     req.Action,
		"citizen":    citizen,
	})
}

// GET /api/scoreboard - 获取记分板
func (h *Handler) handleScoreboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, h.World.Scoreboard)
}

// AIAgentAttachRequest AI agent 接管请求
type AIAgentAttachRequest struct {
	CitizenID int    `json:"citizen_id"`
	AgentName string `json:"agent_name"`
}

// POST /api/ai/attach - 绑定 AI agent 到 citizen（接管模式）
func (h *Handler) handleAIAgentAttach(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req AIAgentAttachRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.AgentName == "" {
		req.AgentName = "openclaw"
	}
	err := h.World.AttachAIAgent(req.CitizenID, req.AgentName)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]interface{}{
		"status":     "attached",
		"citizen_id": req.CitizenID,
		"agent_name": req.AgentName,
	})
}

// POST /api/ai/detach - 解绑 AI agent
func (h *Handler) handleAIAgentDetach(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		CitizenID int `json:"citizen_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	err := h.World.DetachAIAgent(req.CitizenID)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]interface{}{
		"status":     "detached",
		"citizen_id": req.CitizenID,
	})
}
