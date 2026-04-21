package sim

import (
	"encoding/json"
	"math"
	"math/rand"
)

// WorldConfig 世界配置参数
type WorldConfig struct {
	Width          int     `json:"width"`           // 网格宽度
	Height         int     `json:"height"`          // 网格高度
	PeakX1         int     `json:"peak_x1"`         // 第一个糖高峰中心 X
	PeakY1         int     `json:"peak_y1"`         // 第一个糖高峰中心 Y
	PeakX2         int     `json:"peak_x2"`         // 第二个糖高峰中心 X
	PeakY2         int     `json:"peak_y2"`         // 第二个糖高峰中心 Y
	PeakCapacity   float64 `json:"peak_capacity"`   // 峰值容量 (c=4)
	GrowthRate     float64 `json:"growth_rate"`     // 生长率 α
	InitPopulation int     `json:"init_population"` // 初始人口
}

// DefaultConfig 返回默认配置
func DefaultConfig() WorldConfig {
	return WorldConfig{
		Width:          50,
		Height:         50,
		PeakX1:         15,
		PeakY1:         15,
		PeakX2:         35,
		PeakY2:         35,
		PeakCapacity:   4.0,
		GrowthRate:     1.0,
		InitPopulation: 400,
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
}

// NewWorld 创建一个新世界并初始化
func NewWorld(config WorldConfig) *World {
	w := &World{
		Config:   config,
		Timestep: 0,
		NextID:   1,
		rng:      rand.New(rand.NewSource(42)),
	}
	w.initCells()
	w.initCitizens()
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
			// 计算到两个高峰的曼哈顿距离（环面拓扑）
			dist1 := torusDist(x, y, c.PeakX1, c.PeakY1, c.Width, c.Height)
			dist2 := torusDist(x, y, c.PeakX2, c.PeakY2, c.Width, c.Height)
			// 取较近的高峰计算容量
			minDist := math.Min(dist1, dist2)
			// 双峰地形：容量随距离递减，沙漠区域为 0
			// 峰值容量 c=4，距离每增加 1 减少约 0.08（50 格内衰减到 0）
			cap := c.PeakCapacity - minDist*0.08
			if cap < 0 {
				cap = 0
			}
			cell.Capacity = math.Round(cap*100) / 100 // 保留两位小数
			cell.Sugar = cell.Capacity                 // 初始时糖满
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

// initCitizens 初始化公民（均匀分布）
func (w *World) initCitizens() {
	c := w.Config
	w.Citizens = make(map[int]*Citizen)

	// 收集所有有糖的格子位置（避免出生在沙漠）
	var validCells [][2]int
	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			if w.Cells[y][x].Capacity > 0 {
				validCells = append(validCells, [2]int{x, y})
			}
		}
	}

	for i := 0; i < c.InitPopulation; i++ {
		// 均匀随机选一个有糖的格子
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

// Step 推进仿真一步（执行 G → M → R 规则）
func (w *World) Step() {
	w.Timestep++
	w.ruleG() // 生长
	w.ruleM() // 移动
	w.ruleR() // 更替
}

// Reset 重置世界
func (w *World) Reset() {
	w.Timestep = 0
	w.NextID = 1
	w.initCells()
	w.initCitizens()
}

// WorldState 返回世界状态快照（用于 JSON 序列化）
type WorldState struct {
	Config      WorldConfig `json:"config"`
	Timestep    int         `json:"timestep"`
	Population  int         `json:"population"`
	TotalSugar  float64     `json:"total_sugar"`
	TotalCells  int         `json:"total_cells"`
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
	// 扁平化二维数组为一维，方便前端渲染
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
