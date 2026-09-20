// Package kernel 实现最小内核 LPC VM 的通用实体存储
package kernel

import (
	"encoding/json"
	"fmt"

	"github.com/twinsant/sugarland/lpc"
)

// Citizen 表示世界中的一个实体（agent slot）
// 内核只拥有身份、坐标与生命周期；游戏属性（wealth/vision/metabolism
// /age/max_age 等）全部放在 Attrs，由 LPC mudlib 定义含义
type Citizen struct {
	ID                int            `json:"id"`
	X                 int            `json:"x"`
	Y                 int            `json:"y"`
	Alive             bool           `json:"alive"`
	Attrs             map[string]any `json:"-"` // 游戏自定义属性（数值/字符串/数组，含义由 mudlib 定义）
	LPCObj            *lpc.Object    `json:"-"` // 关联的 LPC 对象（living object 行为面，spawn 时绑定蓝图）
	AgentName         string         `json:"agent_name"`
	IsAgentControlled bool           `json:"is_agent_controlled"`
	HasLPC            bool           `json:"has_lpc"`
	Flags             map[string]int `json:"-"` // 每步清零的 scratch 标记（mudlib 记簿用）
}

// NewCitizen 创建一个空白实体（属性为空，仅标记存活）
// 基因与初始属性由 mudlib 通过 set_citizen 决定
func NewCitizen(id, x, y int) *Citizen {
	return &Citizen{
		ID:    id,
		X:     x,
		Y:     y,
		Alive: true,
		Attrs: make(map[string]any),
	}
}

// Attr 读取实体属性（不存在返回 nil）
func (c *Citizen) Attr(name string) any {
	return c.Attrs[name]
}

// AttrInt 读取实体属性并转为 int（不存在或非数值返回 0）
func (c *Citizen) AttrInt(name string) int {
	switch v := c.Attrs[name].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	}
	return 0
}

// AttrFloat 读取实体属性并转为 float64（不存在或非数值返回 0）
func (c *Citizen) AttrFloat(name string) float64 {
	switch v := c.Attrs[name].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	}
	return 0
}

// SetAttr 写入实体属性
func (c *Citizen) SetAttr(name string, v any) {
	c.Attrs[name] = v
}

// MarshalJSON 将 Attrs 平铺到顶层键，保持 wire 格式与客户端兼容
func (c *Citizen) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"id":                  c.ID,
		"x":                   c.X,
		"y":                   c.Y,
		"alive":               c.Alive,
		"agent_name":          c.AgentName,
		"is_agent_controlled": c.IsAgentControlled,
		"has_lpc":             c.HasLPC,
	}
	for k, v := range c.Attrs {
		m[k] = v
	}
	return json.Marshal(m)
}

// HasHeartBeat 检查实体是否有 LPC heart_beat 方法
func (c *Citizen) HasHeartBeat() bool {
	if c.LPCObj == nil {
		return false
	}
	_, ok := c.LPCObj.FindMethod("heart_beat")
	return ok
}

// LoadScript 加载 LPC 脚本到实体
func (c *Citizen) LoadScript(source string) error {
	p := lpc.NewParser(source)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		return fmt.Errorf("parse errors: %v", p.Errors())
	}
	obj := lpc.NewObject(fmt.Sprintf("citizen_%d", c.ID), fmt.Sprintf("Citizen#%d", c.ID))
	obj.LoadProgram(prog)
	c.LPCObj = obj
	return nil
}

func (c *Citizen) String() string {
	return fmt.Sprintf("Citizen#%d(%d,%d) alive=%v attrs=%v", c.ID, c.X, c.Y, c.Alive, c.Attrs)
}
