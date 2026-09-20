// Package kernel 实现最小内核 LPC VM 的通用世界状态存储
package kernel

import (
	"encoding/json"
	"fmt"

	"github.com/twinsant/sugarland/lpc"
)

// Room 表示世界中的一个房间（原 Cell）
// 内核只拥有坐标寻址、属性表、exits 连接表与 contents；
// 属性含义与拓扑语义全部由 LPC mudlib 定义。
// 每个 room 天生绑定一个 LPC 对象（默认 cell.lpc 蓝图），行为面归 LPC
type Room struct {
	X        int                `json:"x"`
	Y        int                `json:"y"`
	Attrs    map[string]float64 `json:"-"` // 游戏自定义属性（sugar/capacity/growth 等）
	Exits    map[string][2]int  `json:"-"` // 方向名 -> 目标坐标（mudlib 维护的连接表）
	Contents map[int]bool       `json:"-"` // 站内实体 ID 集合（内核在移动时维护）
	LPCObj   *lpc.Object        `json:"-"` // 关联的 LPC 对象（room 对象）
}

// NewRoom 创建一个空白 room
func NewRoom(x, y int) *Room {
	return &Room{
		X:        x,
		Y:        y,
		Attrs:    make(map[string]float64),
		Exits:    make(map[string][2]int),
		Contents: make(map[int]bool),
	}
}

// Get 读取 room 属性（不存在返回 0）
func (r *Room) Get(name string) float64 {
	return r.Attrs[name]
}

// Set 写入 room 属性
func (r *Room) Set(name string, v float64) {
	r.Attrs[name] = v
}

// Take 取走该 room 某属性的全部存量，返回取走的量
func (r *Room) Take(name string) float64 {
	v := r.Attrs[name]
	if v <= 0 {
		return 0
	}
	r.Attrs[name] = 0
	return v
}

// MarshalJSON 将 Attrs 平铺到顶层键，保持 wire 格式与客户端兼容
func (r Room) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{"x": r.X, "y": r.Y}
	for k, v := range r.Attrs {
		m[k] = v
	}
	return json.Marshal(m)
}

func (r *Room) String() string {
	return fmt.Sprintf("Room(%d,%d) attrs=%v exits=%d contents=%d", r.X, r.Y, r.Attrs, len(r.Exits), len(r.Contents))
}
