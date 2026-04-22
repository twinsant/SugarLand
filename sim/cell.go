// Package sim 实现 Sugarscape 仿真核心模型和规则
package sim

import (
	"fmt"
	"math"
)

// Cell 表示细胞空间中的单个格子
type Cell struct {
	X        int     `json:"x"`
	Y        int     `json:"y"`
	Sugar    float64 `json:"sugar"`    // 当前糖存量
	Capacity float64 `json:"capacity"` // 糖容量上限（最大糖量）
	Pollution float64 `json:"pollution"` // 污染值
	Growth   float64 `json:"growth"`   // 生长率 α
}

// NewCell 创建一个新的 Cell
func NewCell(x, y int) *Cell {
	return &Cell{
		X:      x,
		Y:      y,
		Growth: 1.0, // 生长率 α = 1
	}
}

// Grow 执行生长规则：糖资源恢复，直到达到容量上限
func (c *Cell) Grow() {
	c.Sugar = math.Min(c.Sugar+c.Growth, c.Capacity)
}

// Harvest 收割指定数量的糖，返回实际收割量
func (c *Cell) Harvest(amount float64) float64 {
	harvested := math.Min(amount, c.Sugar)
	c.Sugar -= harvested
	return harvested
}

func (c *Cell) String() string {
	return fmt.Sprintf("Cell(%d,%d) sugar=%.1f/%.1f pollution=%.1f", c.X, c.Y, c.Sugar, c.Capacity, c.Pollution)
}
