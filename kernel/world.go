// Package kernel 实现最小内核 LPC VM 的通用世界状态存储
package kernel

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"

	"github.com/twinsant/sugarland/lpc"
)

// WorldConfig 世界配置参数
// 内核不解释任何字段：它只是传给 LPC mudlib 的键值数据。
// 不同游戏可在其中放任意参数；网格拓扑键 width/height 除外——
// 环面坐标归一属于内核机制。
type WorldConfig map[string]interface{}

// DefaultConfig 返回内核拓扑所需的最小参数（网格尺寸）
// 游戏参数默认值由 mudlib create() 通过 has_config/set_config 声明，
// 内核不携带任何游戏默认值
func DefaultConfig() WorldConfig {
	return WorldConfig{
		"width":  50,
		"height": 50,
	}
}

// Int 读取整数配置项（JSON 解码后是 float64，字面量则是 int）
func (c WorldConfig) Int(key string) int {
	switch v := c[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	}
	return 0
}

// Float 读取浮点配置项
func (c WorldConfig) Float(key string) float64 {
	switch v := c[key].(type) {
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case float64:
		return v
	}
	return 0
}

// Bool 读取布尔配置项
func (c WorldConfig) Bool(key string) bool {
	v, _ := c[key].(bool)
	return v
}

// World 表示整个世界
type World struct {
	Config      WorldConfig        `json:"config"`
	Timestep    int                `json:"timestep"`
	Rooms       [][]Room           `json:"-"` // 二维网格
	Citizens    map[int]*Citizen   `json:"-"` // 实体字典（ID -> Citizen）
	NextID      int                `json:"-"` // 下一个实体 ID
	rng         *rand.Rand         `json:"-"`
	HeartBeats  *HeartBeatManager  `json:"-"` // Heart Beat 管理器
	ObjManager  *lpc.ObjectManager `json:"-"` // LPC 对象管理器
	WorldObj    *lpc.Object        `json:"-"` // 世界 LPC 对象（承载全部仿真规则）
	WorldSource string             `json:"-"` // 当前世界脚本源码
	WorldCustom bool               `json:"-"` // 世界脚本是否为用户自定义（非默认 mudlib）
	Scoreboard  map[string]any     `json:"scoreboard"`

	// 最近一步产生的事件与统计，供快照接口消费
	LastDeaths    []DeathEvent       `json:"-"`
	LastBirths    []MateResult       `json:"-"`
	LastHarvests  map[int]float64    `json:"-"` // 格子索引 (y*Width+x) -> 本轮采集量
	stepHarvested float64            // 本轮总采集量
	stepConsumed  int                // 本轮总消耗（由 mudlib record_consumed 记账）
	gameStats     map[string]float64 // mudlib emit_stat 上报的游戏指标

	mu sync.Mutex // HTTP handler 与 auto-step 并发保护
}

// Width/Height 网格拓扑尺寸（内核环面归一使用）
func (w *World) Width() int  { return w.Config.Int("width") }
func (w *World) Height() int { return w.Config.Int("height") }

// NewWorld 创建一个新世界并初始化
func NewWorld(config WorldConfig) *World {
	w := &World{
		Config:       config,
		Timestep:     0,
		NextID:       1,
		rng:          rand.New(rand.NewSource(42)),
		HeartBeats:   NewHeartBeatManager(),
		ObjManager:   lpc.NewObjectManager(),
		LastHarvests: make(map[int]float64),
		Citizens:     make(map[int]*Citizen),
		gameStats:    make(map[string]float64),
	}
	ResetGlobalStats()
	w.initRooms()
	w.loadDefaultWorld()
	// 世界生成（地形与初始人口）由 LPC 世界对象的 create() 完成
	w.callWorldVoid("create")
	w.Scoreboard = w.UpdateScoreboard()
	return w
}

// initRooms 分配网格并为每格绑定默认 cell.lpc 蓝图对象
// 格子内容全部由 mudlib create() 填充；蓝图 AST 只解析一次，2500 个对象共享
func (w *World) initRooms() {
	ww, hh := w.Width(), w.Height()
	w.Rooms = make([][]Room, hh)
	for y := 0; y < hh; y++ {
		w.Rooms[y] = make([]Room, ww)
		for x := 0; x < ww; x++ {
			w.Rooms[y][x] = *NewRoom(x, y)
		}
	}
	p := lpc.NewParser(defaultCellSource)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		panic(fmt.Sprintf("default cell blueprint failed to load: %v", p.Errors()))
	}
	for y := 0; y < hh; y++ {
		for x := 0; x < ww; x++ {
			w.bindRoomObject(&w.Rooms[y][x], prog)
		}
	}
}

// bindRoomObject 为格子绑定 LPC 对象（共享 AST），注册格子 efun 并调用 create()
// 调用方需已持有 w.mu
func (w *World) bindRoomObject(c *Room, prog *lpc.ProgramAST) {
	obj := lpc.NewObject(fmt.Sprintf("cell_%d_%d", c.X, c.Y), "Room")
	obj.LoadProgram(prog)
	w.ObjManager.Add(obj.ID, obj)
	c.LPCObj = obj
	w.registerRoomEfuns(obj.VM, c)
	w.callRoomMethod(c, "create")
}

// bindRoomSource 用 LPC 源码为格子绑定对象（bind_cell efun 的实现）
// 调用方需已持有 w.mu
func (w *World) bindRoomSource(c *Room, source string) error {
	p := lpc.NewParser(source)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		return fmt.Errorf("parse errors: %v", p.Errors())
	}
	w.bindRoomObject(c, prog)
	return nil
}

// torusDist 计算环面上的曼哈顿距离
func torusDist(x1, y1, x2, y2, w, h int) float64 {
	dx := abs(x1 - x2)
	dy := abs(y1 - y2)
	if dx > w/2 {
		dx = w - dx
	}
	if dy > h/2 {
		dy = h - dy
	}
	return float64(dx + dy)
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// GetRoom 获取指定坐标的 Room（环面拓扑包裹）
func (w *World) GetRoom(x, y int) *Room {
	wx := ((x % w.Width()) + w.Width()) % w.Width()
	wy := ((y % w.Height()) + w.Height()) % w.Height()
	return &w.Rooms[wy][wx]
}

// beginStep 清空上一轮的事件、统计与实体 scratch 标记
func (w *World) beginStep() {
	w.LastDeaths = nil
	w.LastBirths = nil
	w.LastHarvests = make(map[int]float64)
	w.stepHarvested = 0
	w.stepConsumed = 0
	w.gameStats = make(map[string]float64)
	for _, c := range w.Citizens {
		c.Flags = nil
	}
}

// recordHarvest 记录一次采集（供快照与统计使用）
// 坐标先按环面归一：citizen 的 X/Y 本身可能越界（move 只改坐标，GetRoom 才包裹）
func (w *World) recordHarvest(x, y int, amount float64) {
	wx := ((x % w.Width()) + w.Width()) % w.Width()
	wy := ((y % w.Height()) + w.Height()) % w.Height()
	w.LastHarvests[wy*w.Width()+wx] += amount
	w.stepHarvested += amount
}

// takeRoomResource 取走指定格某属性的全部存量并记账（内核记账原语）
func (w *World) takeRoomResource(x, y int, name string) float64 {
	got := w.GetRoom(x, y).Take(name)
	w.recordHarvest(x, y, got)
	return got
}

// moveCitizenTo 将实体移动到指定坐标并同步维护两端 room 的 contents
// 坐标在写入侧按环面归一（contents 一致性要求）
func (w *World) moveCitizenTo(c *Citizen, x, y int) {
	ox, oy := c.X, c.Y
	delete(w.GetRoom(ox, oy).Contents, c.ID)
	c.X = ((x % w.Width()) + w.Width()) % w.Width()
	c.Y = ((y % w.Height()) + w.Height()) % w.Height()
	w.GetRoom(c.X, c.Y).Contents[c.ID] = true
	// 记账：本次移动的环面距离写入 last_move_dist，
	// 是否计费（移动消耗）由 mudlib 决定
	c.SetAttr("last_move_dist", int(torusDist(ox, oy, c.X, c.Y, w.Width(), w.Height())))
}

// walkCitizen 沿实体所在 room 的命名 exit 移动，返回是否成功
func (w *World) walkCitizen(c *Citizen, dir string) bool {
	dest, ok := w.GetRoom(c.X, c.Y).Exits[dir]
	if !ok {
		return false
	}
	w.moveCitizenTo(c, dest[0], dest[1])
	return true
}

// Step 推进仿真一步
// 内核只负责清场、计时与记分板；全部规则由世界 LPC 对象的 heart_beat 编排
func (w *World) Step() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.beginStep()
	w.Timestep++
	w.lpcWorldHeartBeat()
	w.Scoreboard = w.UpdateScoreboard()
}

// StepWithPause 推进一步，返回需要 agent 决策的 citizen ID 列表
func (w *World) StepWithPause() []int {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.beginStep()
	w.Timestep++
	w.lpcWorldHeartBeat()
	w.Scoreboard = w.UpdateScoreboard()

	var waitingAgents []int
	for _, c := range w.Citizens {
		if c.IsAgentControlled {
			waitingAgents = append(waitingAgents, c.ID)
		}
	}
	return waitingAgents
}

// ExecuteCommand 执行 agent 对 citizen 的指令
func (w *World) ExecuteCommand(citizenID int, action string, args map[string]interface{}) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.executeCommand(citizenID, action, args)
}

// executeCommand 执行 agent 对 citizen 的指令
// 命令是纯原语：move 只写坐标；harvest 取走格子资源（args.resource 指定属性名，
// 缺省取走全部属性；args.into 指定存入的实体属性名）；stay 不动。
// 环境后果由 mudlib 在世界节拍中统一结算
func (w *World) executeCommand(citizenID int, action string, args map[string]interface{}) error {
	citizen, ok := w.Citizens[citizenID]
	if !ok {
		return fmt.Errorf("citizen %d not found", citizenID)
	}
	if !citizen.IsAgentControlled {
		return fmt.Errorf("citizen %d is not agent-controlled", citizenID)
	}

	switch action {
	case "move":
		x, xOk := args["x"].(float64)
		y, yOk := args["y"].(float64)
		if !xOk || !yOk {
			return fmt.Errorf("move requires x and y args")
		}
		w.moveCitizenTo(citizen, int(x), int(y))
	case "harvest":
		resource, _ := args["resource"].(string)
		into, _ := args["into"].(string)
		names := []string{resource}
		if resource == "" {
			names = names[:0]
			for name := range w.GetRoom(citizen.X, citizen.Y).Attrs {
				names = append(names, name)
			}
		}
		for _, name := range names {
			got := w.takeRoomResource(citizen.X, citizen.Y, name)
			if into != "" {
				citizen.SetAttr(into, citizen.AttrFloat(into)+got)
			}
		}
	case "walk":
		dir, _ := args["dir"].(string)
		if !w.walkCitizen(citizen, dir) {
			return fmt.Errorf("no exit %q from (%d,%d)", dir, citizen.X, citizen.Y)
		}
	case "stay":
		// 原地不动；mudlib 在世界节拍中结算环境后果
	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	return nil
}

// Reset 重置世界
// ApplyConfig 合并配置键（调用方显式传入的值优先于 mudlib declare_defaults 的缺省）
// 通用机制：内核不解释键的含义
func (w *World) ApplyConfig(cfg map[string]interface{}) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for k, v := range cfg {
		w.Config[k] = v
	}
}

func (w *World) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.Timestep = 0
	w.NextID = 1
	w.HeartBeats = NewHeartBeatManager()
	w.ObjManager = lpc.NewObjectManager()
	w.WorldObj = nil
	w.LastDeaths = nil
	w.LastBirths = nil
	w.LastHarvests = make(map[int]float64)
	w.stepHarvested = 0
	w.stepConsumed = 0
	w.gameStats = make(map[string]float64)
	ResetGlobalStats()
	w.initRooms()
	w.Citizens = make(map[int]*Citizen)
	w.loadDefaultWorld()
	w.callWorldVoid("create")
	w.Scoreboard = w.UpdateScoreboard()
}

// AttachAgent 绑定 LPC agent 到指定实体
func (w *World) AttachAgent(citizenID int, lpcSource string, interval int) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.attachAgent(citizenID, lpcSource, interval)
}

func (w *World) attachAgent(citizenID int, lpcSource string, interval int) error {
	citizen, ok := w.Citizens[citizenID]
	if !ok {
		return fmt.Errorf("citizen %d not found", citizenID)
	}

	err := citizen.LoadScript(lpcSource)
	if err != nil {
		return err
	}

	objID := fmt.Sprintf("citizen_%d", citizenID)
	w.ObjManager.Add(objID, citizen.LPCObj)
	w.HeartBeats.Register(objID, citizenID, interval)
	citizen.HasLPC = true

	return nil
}

// bindCitizenObject 为实体加载 LPC 对象（蓝图或脚本）并注册到对象管理器
// 每个实体天生是 LPC object：spawn 时绑定默认蓝图，attach 替换为用户脚本
func (w *World) bindCitizenObject(c *Citizen, source string) error {
	if err := c.LoadScript(source); err != nil {
		return err
	}
	w.ObjManager.Add(fmt.Sprintf("citizen_%d", c.ID), c.LPCObj)
	return nil
}

// DetachAgent 解绑实体的 LPC agent
func (w *World) DetachAgent(citizenID int) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.detachAgent(citizenID)
}

func (w *World) detachAgent(citizenID int) error {
	citizen, ok := w.Citizens[citizenID]
	if !ok {
		return fmt.Errorf("citizen %d not found", citizenID)
	}

	objID := fmt.Sprintf("citizen_%d", citizenID)
	w.HeartBeats.Unregister(objID)
	w.ObjManager.Destroy(objID)
	citizen.LPCObj = nil
	citizen.HasLPC = false
	// 恢复默认蓝图对象：实体回到 mudlib 定义的默认行为
	if err := w.bindCitizenObject(citizen, defaultCitizenSource); err != nil {
		fmt.Printf("[LPC] restore blueprint for citizen %d: %v\n", citizenID, err)
	}

	return nil
}

// AttachAIAgent 将 AI agent 绑定到实体（接管模式）
func (w *World) AttachAIAgent(citizenID int, agentName string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	citizen, ok := w.Citizens[citizenID]
	if !ok {
		return fmt.Errorf("citizen %d not found", citizenID)
	}
	citizen.IsAgentControlled = true
	citizen.AgentName = agentName
	return nil
}

// DetachAIAgent 解绑 AI agent
func (w *World) DetachAIAgent(citizenID int) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	citizen, ok := w.Citizens[citizenID]
	if !ok {
		return fmt.Errorf("citizen %d not found", citizenID)
	}
	citizen.IsAgentControlled = false
	citizen.AgentName = ""
	return nil
}

// AgentContext 表示 agent 上下文信息
// Attrs 平铺到顶层键；不同游戏实体的属性集合由 mudlib 决定
type AgentContext struct {
	CitizenID        int                      `json:"citizen_id"`
	AgentName        string                   `json:"agent_name"`
	X                int                      `json:"x"`
	Y                int                      `json:"y"`
	Attrs            map[string]any           `json:"-"`
	Timestep         int                      `json:"timestep"`
	AvailableActions []string                 `json:"available_actions"`
	Neighbors        []NeighborInfo           `json:"neighbors"`
	NearbyCells      []map[string]interface{} `json:"nearby_cells"`
}

// MarshalJSON 将 Attrs 平铺到顶层键
func (ctx *AgentContext) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"citizen_id":        ctx.CitizenID,
		"agent_name":        ctx.AgentName,
		"x":                 ctx.X,
		"y":                 ctx.Y,
		"timestep":          ctx.Timestep,
		"available_actions": ctx.AvailableActions,
		"neighbors":         ctx.Neighbors,
		"nearby_cells":      ctx.NearbyCells,
	}
	for k, v := range ctx.Attrs {
		m[k] = v
	}
	return json.Marshal(m)
}

// NeighborInfo 相邻实体信息（Attrs 平铺）
type NeighborInfo struct {
	ID    int            `json:"id"`
	X     int            `json:"x"`
	Y     int            `json:"y"`
	Attrs map[string]any `json:"-"`
}

// MarshalJSON 将 Attrs 平铺到顶层键
func (n NeighborInfo) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{"id": n.ID, "x": n.X, "y": n.Y}
	for k, v := range n.Attrs {
		m[k] = v
	}
	return json.Marshal(m)
}

// GetAgentContext 获取实体的完整 agent 上下文
// radius < 0 表示自动（读约定属性 "vision"，默认 mudlib 的约定）；
// radius >= 0 显式指定视野半径，与游戏属性解耦
func (w *World) GetAgentContext(citizenID int, radius int) (*AgentContext, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	citizen, ok := w.Citizens[citizenID]
	if !ok {
		return nil, fmt.Errorf("citizen %d not found", citizenID)
	}
	// 视野半径：显式传入优先；否则回退读约定属性 "vision"（默认 mudlib 的约定，
	// 其他游戏可在 API 层显式传 radius）
	vision := radius
	if vision < 0 {
		vision = citizen.AttrInt("vision")
	}

	ctx := &AgentContext{
		CitizenID:        citizenID,
		AgentName:        citizen.AgentName,
		X:                citizen.X,
		Y:                citizen.Y,
		Attrs:            citizen.Attrs,
		Timestep:         w.Timestep,
		AvailableActions: []string{"move", "walk", "harvest", "stay"},
	}

	// 收集视野内邻居
	for _, c := range w.Citizens {
		if c.ID == citizenID || !c.Alive {
			continue
		}
		dist := torusDist(citizen.X, citizen.Y, c.X, c.Y, w.Width(), w.Height())
		if dist <= float64(vision) {
			ctx.Neighbors = append(ctx.Neighbors, NeighborInfo{
				ID: c.ID, X: c.X, Y: c.Y, Attrs: c.Attrs,
			})
		}
	}

	// 收集视野内格子
	for dy := -vision; dy <= vision; dy++ {
		for dx := -vision; dx <= vision; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			if abs(dx)+abs(dy) > vision {
				continue
			}
			cell := w.GetRoom(citizen.X+dx, citizen.Y+dy)
			info := map[string]interface{}{"x": cell.X, "y": cell.Y}
			for k, v := range cell.Attrs {
				info[k] = v
			}
			ctx.NearbyCells = append(ctx.NearbyCells, info)
		}
	}

	return ctx, nil
}

// WorldState 返回世界状态快照（用于 JSON 序列化）
// total_<attr> 键按格子属性动态生成（如 total_sugar）
type WorldState struct {
	Config     WorldConfig
	Timestep   int
	Population int
	TotalCells int
	Totals     map[string]float64
}

// MarshalJSON 输出 config/timestep/population/total_cells 及各属性 total_<name>
func (s WorldState) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"config":      s.Config,
		"timestep":    s.Timestep,
		"population":  s.Population,
		"total_cells": s.TotalCells,
	}
	for name, v := range s.Totals {
		m["total_"+name] = v
	}
	return json.Marshal(m)
}

// GetState 获取世界状态
func (w *World) GetState() WorldState {
	w.mu.Lock()
	defer w.mu.Unlock()
	totals := make(map[string]float64)
	for y := 0; y < w.Height(); y++ {
		for x := 0; x < w.Width(); x++ {
			for name, v := range w.Rooms[y][x].Attrs {
				totals[name] += v
			}
		}
	}
	return WorldState{
		Config:     w.Config,
		Timestep:   w.Timestep,
		Population: len(w.Citizens),
		TotalCells: w.Width() * w.Height(),
		Totals:     totals,
	}
}

// GetRoomsJSON 返回所有 Room 的 JSON 快照
func (w *World) GetRoomsJSON() ([]byte, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	cells := make([]Room, 0, w.Width()*w.Height())
	for y := 0; y < w.Height(); y++ {
		for x := 0; x < w.Width(); x++ {
			cells = append(cells, w.Rooms[y][x])
		}
	}
	return json.Marshal(cells)
}

// GetCitizensList 返回所有实体列表
func (w *World) GetCitizensList() []*Citizen {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.getCitizensListLocked()
}

// GetCitizenByID 按 ID 获取实体
func (w *World) GetCitizenByID(id int) *Citizen {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.Citizens[id]
}
