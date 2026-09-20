// Package kernel 实现最小内核的聚合快照
package kernel

import "encoding/json"

// MateResult 表示一次繁殖事件（由 mudlib 通过 emit_birth efun 记录）
type MateResult struct {
	ParentAID int `json:"parent_a_id"`
	ParentBID int `json:"parent_b_id"`
	ChildID   int `json:"child_id"`
	Timestep  int `json:"timestep"`
}

// DeathEvent 记录一步之内的实体死亡事件（供快照接口消费）
// Attrs 是死亡时的属性快照，序列化时平铺到顶层键
type DeathEvent struct {
	ID     int            `json:"id"`
	X      int            `json:"x"`
	Y      int            `json:"y"`
	Reason string         `json:"reason"`
	Attrs  map[string]any `json:"-"`
}

// MarshalJSON 将 Attrs 平铺到顶层键
func (d DeathEvent) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{"id": d.ID, "x": d.X, "y": d.Y, "reason": d.Reason}
	for k, v := range d.Attrs {
		m[k] = v
	}
	return json.Marshal(m)
}

// Snapshot 是面向外部渲染端（如 Godot 客户端）的聚合世界快照
// Resources 按格子属性名各生成一个扁平数组（索引 y*Width+x），
// 序列化时平铺到顶层键——mudlib 定义了 sugar/capacity 就出现对应数组
type Snapshot struct {
	Timestep   int                  `json:"timestep"`
	Config     WorldConfig          `json:"config"`
	Resources  map[string][]float64 `json:"-"`
	Citizens   []*Citizen           `json:"citizens"`
	Scoreboard map[string]any       `json:"scoreboard"`
	Deaths     []DeathEvent         `json:"deaths"`
	Births     []MateResult         `json:"births"`
	Harvests   map[int]float64      `json:"harvests"`
	Harvested  float64              `json:"harvested"`
	Consumed   int                  `json:"consumed"`
	WorldAgent bool                 `json:"world_agent"`
}

// MarshalJSON 输出固定字段 + 每个格子属性一个顶层数组键
func (s Snapshot) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"timestep":    s.Timestep,
		"config":      s.Config,
		"citizens":    s.Citizens,
		"scoreboard":  s.Scoreboard,
		"deaths":      s.Deaths,
		"births":      s.Births,
		"harvests":    s.Harvests,
		"harvested":   s.Harvested,
		"consumed":    s.Consumed,
		"world_agent": s.WorldAgent,
	}
	for name, arr := range s.Resources {
		m[name] = arr
	}
	return json.Marshal(m)
}

// GetSnapshot 生成当前世界的聚合快照
func (w *World) GetSnapshot() Snapshot {
	w.mu.Lock()
	defer w.mu.Unlock()

	size := w.Width() * w.Height()
	resources := make(map[string][]float64)
	for y := 0; y < w.Height(); y++ {
		for x := 0; x < w.Width(); x++ {
			i := y*w.Width() + x
			for name, v := range w.Rooms[y][x].Attrs {
				arr, ok := resources[name]
				if !ok {
					arr = make([]float64, size)
					resources[name] = arr
				}
				arr[i] = v
			}
		}
	}

	deaths := w.LastDeaths
	if deaths == nil {
		deaths = []DeathEvent{}
	}
	births := w.LastBirths
	if births == nil {
		births = []MateResult{}
	}
	harvests := w.LastHarvests
	if harvests == nil {
		harvests = map[int]float64{}
	}

	return Snapshot{
		Timestep:   w.Timestep,
		Config:     w.Config,
		Resources:  resources,
		Citizens:   w.getCitizensListLocked(),
		Scoreboard: w.Scoreboard,
		Deaths:     deaths,
		Births:     births,
		Harvests:   harvests,
		Harvested:  w.stepHarvested,
		Consumed:   w.stepConsumed,
		WorldAgent: w.WorldCustom,
	}
}

// getCitizensListLocked 返回所有实体列表（调用方需已持有锁）
func (w *World) getCitizensListLocked() []*Citizen {
	list := make([]*Citizen, 0, len(w.Citizens))
	for _, c := range w.Citizens {
		list = append(list, c)
	}
	return list
}
