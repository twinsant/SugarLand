// Package sim 实现 Sugarscape 仿真核心模型和规则
package sim

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"

	"github.com/twinsant/sugarland/lpc"
)

// WorldConfig 世界配置参数
type WorldConfig struct {
	Width           int     `json:"width"`            // 网格宽度
	Height          int     `json:"height"`           // 网格高度
	PeakX1          int     `json:"peak_x1"`          // 第一个糖高峰中心 X
	PeakY1          int     `json:"peak_y1"`          // 第一个糖高峰中心 Y
	PeakX2          int     `json:"peak_x2"`          // 第二个糖高峰中心 X
	PeakY2          int     `json:"peak_y2"`          // 第二个糖高峰中心 Y
	PeakCapacity    float64 `json:"peak_capacity"`    // 峰值容量 (c=4)
	GrowthRate      float64 `json:"growth_rate"`      // 生长率 α
	InitPopulation  int     `json:"init_population"`  // 初始人口
	EnableTrading   bool    `json:"enable_trading"`   // 是否启用贸易
	EnableMating    bool    `json:"enable_mating"`    // 是否启用繁殖
	EnablePollution bool    `json:"enable_pollution"` // 是否启用污染
	SeasonInterval  int     `json:"season_interval"`  // 季节切换间隔
}

// DefaultConfig 返回默认配置
func DefaultConfig() WorldConfig {
	return WorldConfig{
		Width:           50,
		Height:          50,
		PeakX1:          15,
		PeakY1:          15,
		PeakX2:          35,
		PeakY2:          35,
		PeakCapacity:    4.0,
		GrowthRate:      1.0,
		InitPopulation:  400,
		EnableTrading:   true,
		EnableMating:    true,
		EnablePollution: false,
		SeasonInterval:  100,
	}
}

// World 表示整个 Sugarscape 世界
type World struct {
	Config      WorldConfig         `json:"config"`
	Timestep    int                 `json:"timestep"`
	Cells       [][]Cell            `json:"-"` // 二维网格
	Citizens    map[int]*Citizen    `json:"-"` // 公民字典（ID -> Citizen）
	NextID      int                 `json:"-"` // 下一个公民 ID
	rng         *rand.Rand          `json:"-"`
	HeartBeats  *HeartBeatManager   `json:"-"` // Heart Beat 管理器
	ObjManager  *lpc.ObjectManager  `json:"-"` // LPC 对象管理器
	Scoreboard  Scoreboard          `json:"scoreboard"`
}

// NewWorld 创建一个新世界并初始化
func NewWorld(config WorldConfig) *World {
	w := &World{
		Config:     config,
		Timestep:   0,
		NextID:     1,
		rng:        rand.New(rand.NewSource(42)),
		HeartBeats: NewHeartBeatManager(),
		ObjManager: lpc.NewObjectManager(),
	}
	ResetGlobalStats()
	w.initCells()
	w.initCitizens()
	w.Scoreboard = w.UpdateScoreboard()
	return w
}

// initCells 初始化细胞空间：双峰糖资源地形 + 环面拓扑
func (w *World) initCells() {
	c := w.Config
	w.Cells = make([][]Cell, c.Height)
	for y := 0; y < c.Height; y++ {
		w.Cells[y] = make([]Cell, c.Width)
		for x := 0; x < c.Width; x++ {
			cell := NewCell(x, y)
			dist1 := torusDist(x, y, c.PeakX1, c.PeakY1, c.Width, c.Height)
			dist2 := torusDist(x, y, c.PeakX2, c.PeakY2, c.Width, c.Height)
			minDist := math.Min(dist1, dist2)
			cap := c.PeakCapacity - minDist*0.08
			if cap < 0 {
				cap = 0
			}
			cell.Capacity = math.Round(cap*100) / 100
			cell.Sugar = cell.Capacity
			cell.Growth = c.GrowthRate
			w.Cells[y][x] = *cell
		}
	}
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// initCitizens 初始化公民（均匀分布）
func (w *World) initCitizens() {
	c := w.Config
	w.Citizens = make(map[int]*Citizen)

	var validCells [][2]int
	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			if w.Cells[y][x].Capacity > 0 {
				validCells = append(validCells, [2]int{x, y})
			}
		}
	}

	for i := 0; i < c.InitPopulation; i++ {
		pos := validCells[w.rng.Intn(len(validCells))]
		citizen := NewCitizen(w.NextID, pos[0], pos[1], w.rng)
		w.Citizens[w.NextID] = citizen
		w.NextID++
	}
}

// GetCell 获取指定坐标的 Cell（环面拓扑包裹）
func (w *World) GetCell(x, y int) *Cell {
	wx := ((x % w.Config.Width) + w.Config.Width) % w.Config.Width
	wy := ((y % w.Config.Height) + w.Config.Height) % w.Config.Height
	return &w.Cells[wy][wx]
}

// Step 推进仿真一步（执行 G → M → Trade → Mate → R 规则）
func (w *World) Step() {
	w.Timestep++
	w.ruleG()      // 生长
	w.ruleM()      // 移动
	w.ExecuteTrading() // 贸易
	w.ExecuteMating()  // 繁殖
	w.ruleR()      // 更替
	w.Scoreboard = w.UpdateScoreboard()
}

// StepWithPause 推进一步，遇到 agent-controlled 的 citizen 时暂停
// 返回需要 agent 决策的 citizen ID 列表
func (w *World) StepWithPause() []int {
	w.Timestep++
	w.ruleG()

	// 移动阶段：检查 agent-controlled citizen
	var waitingAgents []int
	alive := w.GetAliveCitizens()
	rand.Shuffle(len(alive), func(i, j int) {
		alive[i], alive[j] = alive[j], alive[i]
	})
	for _, c := range alive {
		if !c.Alive {
			continue
		}
		if c.IsAgentControlled {
			if !c.PendingCommand {
				// 需要 agent 决策
				waitingAgents = append(waitingAgents, c.ID)
				continue
			}
			// 有待执行命令，由 Command API 处理
			c.PendingCommand = false
			continue
		}
		// 非 agent-controlled，正常执行
		if c.HasHeartBeat() {
			w.lpcHeartBeat(c)
		} else {
			w.moveAndHarvest(c)
		}
	}

	w.ExecuteTrading()
	w.ExecuteMating()
	w.ruleR()
	w.Scoreboard = w.UpdateScoreboard()

	return waitingAgents
}

// ExecuteCommand 执行 agent 对 citizen 的指令
func (w *World) ExecuteCommand(citizenID int, action string, args map[string]interface{}) error {
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
		citizen.X = int(x)
		citizen.Y = int(y)
		// 执行收割和代谢
		cell := w.GetCell(citizen.X, citizen.Y)
		harvested := cell.Harvest(cell.Sugar)
		citizen.Wealth += int(harvested)
		citizen.Wealth -= citizen.Metabolism
		citizen.Age++
		if citizen.Wealth <= 0 {
			citizen.Alive = false
		}
		if citizen.Age >= citizen.MaxAge {
			citizen.Alive = false
		}
	case "harvest":
		cell := w.GetCell(citizen.X, citizen.Y)
		harvested := cell.Harvest(cell.Sugar)
		citizen.Wealth += int(harvested)
	case "stay":
		// 不移动，只收割和代谢
		cell := w.GetCell(citizen.X, citizen.Y)
		harvested := cell.Harvest(cell.Sugar)
		citizen.Wealth += int(harvested)
		citizen.Wealth -= citizen.Metabolism
		citizen.Age++
		if citizen.Wealth <= 0 {
			citizen.Alive = false
		}
		if citizen.Age >= citizen.MaxAge {
			citizen.Alive = false
		}
	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	return nil
}

// Reset 重置世界
func (w *World) Reset() {
	w.Timestep = 0
	w.NextID = 1
	w.HeartBeats = NewHeartBeatManager()
	w.ObjManager = lpc.NewObjectManager()
	ResetGlobalStats()
	w.initCells()
	w.initCitizens()
	w.Scoreboard = w.UpdateScoreboard()
}

// AttachAgent 绑定 LPC agent 到指定公民
func (w *World) AttachAgent(citizenID int, lpcSource string, interval int) error {
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

	return nil
}

// DetachAgent 解绑公民的 LPC agent
func (w *World) DetachAgent(citizenID int) error {
	citizen, ok := w.Citizens[citizenID]
	if !ok {
		return fmt.Errorf("citizen %d not found", citizenID)
	}

	objID := fmt.Sprintf("citizen_%d", citizenID)
	w.HeartBeats.Unregister(objID)
	w.ObjManager.Destroy(objID)
	citizen.LPCObj = nil

	return nil
}

// AttachAIAgent 将 AI agent 绑定到 citizen（接管模式）
func (w *World) AttachAIAgent(citizenID int, agentName string) error {
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
	citizen, ok := w.Citizens[citizenID]
	if !ok {
		return fmt.Errorf("citizen %d not found", citizenID)
	}
	citizen.IsAgentControlled = false
	citizen.AgentName = ""
	citizen.PendingCommand = false
	return nil
}

// AgentContext 表示 agent 上下文信息
type AgentContext struct {
	CitizenID      int              `json:"citizen_id"`
	AgentName      string           `json:"agent_name"`
	X              int              `json:"x"`
	Y              int              `json:"y"`
	Wealth         int              `json:"wealth"`
	Vision         int              `json:"vision"`
	Metabolism     int              `json:"metabolism"`
	Age            int              `json:"age"`
	MaxAge         int              `json:"max_age"`
	Sugar          int              `json:"sugar"`
	Timestep       int              `json:"timestep"`
	AvailableActions []string       `json:"available_actions"`
	Neighbors      []NeighborInfo   `json:"neighbors"`
	NearbyCells    []CellInfo       `json:"nearby_cells"`
}

// NeighborInfo 相邻 citizen 信息
type NeighborInfo struct {
	ID     int `json:"id"`
	X      int `json:"x"`
	Y      int `json:"y"`
	Wealth int `json:"wealth"`
	Age    int `json:"age"`
}

// CellInfo 附近格子信息
type CellInfo struct {
	X     int     `json:"x"`
	Y     int     `json:"y"`
	Sugar float64 `json:"sugar"`
}

// GetAgentContext 获取公民的完整 agent 上下文
func (w *World) GetAgentContext(citizenID int) (*AgentContext, error) {
	citizen, ok := w.Citizens[citizenID]
	if !ok {
		return nil, fmt.Errorf("citizen %d not found", citizenID)
	}
	cell := w.GetCell(citizen.X, citizen.Y)

	ctx := &AgentContext{
		CitizenID:  citizenID,
		AgentName:  citizen.AgentName,
		X:          citizen.X,
		Y:          citizen.Y,
		Wealth:     citizen.Wealth,
		Vision:     citizen.Vision,
		Metabolism: citizen.Metabolism,
		Age:        citizen.Age,
		MaxAge:     citizen.MaxAge,
		Sugar:      int(cell.Sugar),
		Timestep:   w.Timestep,
		AvailableActions: []string{"move", "harvest", "stay"},
	}

	// 收集邻居信息
	for _, c := range w.Citizens {
		if c.ID == citizenID || !c.Alive {
			continue
		}
		dist := torusDist(citizen.X, citizen.Y, c.X, c.Y, w.Config.Width, w.Config.Height)
		if dist <= float64(citizen.Vision) {
			ctx.Neighbors = append(ctx.Neighbors, NeighborInfo{
				ID:     c.ID,
				X:      c.X,
				Y:      c.Y,
				Wealth: c.Wealth,
				Age:    c.Age,
			})
		}
	}

	// 收集视野内格子信息
	for dy := -citizen.Vision; dy <= citizen.Vision; dy++ {
		for dx := -citizen.Vision; dx <= citizen.Vision; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			if abs(dx)+abs(dy) > citizen.Vision {
				continue
			}
			c := w.GetCell(citizen.X+dx, citizen.Y+dy)
			ctx.NearbyCells = append(ctx.NearbyCells, CellInfo{
				X:     c.X,
				Y:     c.Y,
				Sugar: c.Sugar,
			})
		}
	}

	return ctx, nil
}

// WorldState 返回世界状态快照（用于 JSON 序列化）
type WorldState struct {
	Config     WorldConfig `json:"config"`
	Timestep   int         `json:"timestep"`
	Population int         `json:"population"`
	TotalSugar float64     `json:"total_sugar"`
	TotalCells int         `json:"total_cells"`
}

// GetState 获取世界状态
func (w *World) GetState() WorldState {
	totalSugar := 0.0
	for y := 0; y < w.Config.Height; y++ {
		for x := 0; x < w.Config.Width; x++ {
			totalSugar += w.Cells[y][x].Sugar
		}
	}
	return WorldState{
		Config:     w.Config,
		Timestep:   w.Timestep,
		Population: len(w.Citizens),
		TotalSugar: totalSugar,
		TotalCells: w.Config.Width * w.Config.Height,
	}
}

// GetCellsJSON 返回所有 Cell 的 JSON 快照
func (w *World) GetCellsJSON() ([]byte, error) {
	cells := make([]Cell, 0, w.Config.Width*w.Config.Height)
	for y := 0; y < w.Config.Height; y++ {
		for x := 0; x < w.Config.Width; x++ {
			cells = append(cells, w.Cells[y][x])
		}
	}
	return json.Marshal(cells)
}

// GetCitizensList 返回所有公民列表
func (w *World) GetCitizensList() []*Citizen {
	list := make([]*Citizen, 0, len(w.Citizens))
	for _, c := range w.Citizens {
		list = append(list, c)
	}
	return list
}

// GetCitizenByID 按 ID 获取公民
func (w *World) GetCitizenByID(id int) *Citizen {
	return w.Citizens[id]
}
