package sim

import (
	"fmt"
	"math/rand"

	"github.com/twinsant/sugarland/lpc"
)

// 核心规则集：按顺序执行 G → M → R

// ruleG 生长规则：每个 Cell 根据生长率恢复糖资源，直到达到容量上限
func (w *World) ruleG() {
	for y := 0; y < w.Config.Height; y++ {
		for x := 0; x < w.Config.Width; x++ {
			w.Cells[y][x].Grow()
		}
	}
}

// ruleM 移动规则：所有 Citizen 随机顺序执行
// 在视觉范围内找糖最多的格子，移动、收割、扣除代谢消耗
func (w *World) ruleM() {
	// 收集存活的公民
	alive := make([]*Citizen, 0, len(w.Citizens))
	for _, c := range w.Citizens {
		if c.Alive {
			alive = append(alive, c)
		}
	}
	// 随机打乱顺序
	rand.Shuffle(len(alive), func(i, j int) {
		alive[i], alive[j] = alive[j], alive[i]
	})
	// 按随机顺序执行移动
	for _, c := range alive {
		if !c.Alive {
			continue
		}
		// 如果有 LPC heart_beat，先调用它
		if c.HasHeartBeat() {
			w.lpcHeartBeat(c)
		} else {
			w.moveAndHarvest(c)
		}
	}
}

// lpcHeartBeat 通过 LPC heart_beat 方法执行公民行为
func (w *World) lpcHeartBeat(c *Citizen) {
	obj := c.LPCObj
	if obj == nil || obj.VM == nil {
		w.moveAndHarvest(c)
		return
	}

	// 注册当前 world context 的 efun
	vm := obj.VM
	vm.RegisterEfun("query_x", func(args []lpc.Value) lpc.Value {
		return lpc.IntValue(c.X)
	})
	vm.RegisterEfun("query_y", func(args []lpc.Value) lpc.Value {
		return lpc.IntValue(c.Y)
	})
	vm.RegisterEfun("query_sugar", func(args []lpc.Value) lpc.Value {
		cell := w.GetCell(c.X, c.Y)
		return lpc.IntValue(int(cell.Sugar))
	})
	vm.RegisterEfun("move", func(args []lpc.Value) lpc.Value {
		if len(args) >= 2 {
			c.X = args[0].IntVal
			c.Y = args[1].IntVal
		}
		return lpc.Null()
	})
	vm.RegisterEfun("query_wealth", func(args []lpc.Value) lpc.Value {
		return lpc.IntValue(c.Wealth)
	})
	vm.RegisterEfun("query_vision", func(args []lpc.Value) lpc.Value {
		return lpc.IntValue(c.Vision)
	})
	vm.RegisterEfun("harvest", func(args []lpc.Value) lpc.Value {
		cell := w.GetCell(c.X, c.Y)
		harvested := cell.Harvest(cell.Sugar)
		c.Wealth += int(harvested)
		return lpc.IntValue(int(harvested))
	})
	vm.RegisterEfun("query_cell_sugar", func(args []lpc.Value) lpc.Value {
		if len(args) >= 2 {
			cell := w.GetCell(args[0].IntVal, args[1].IntVal)
			return lpc.IntValue(int(cell.Sugar))
		}
		return lpc.IntValue(0)
	})
	vm.RegisterEfun("random", func(args []lpc.Value) lpc.Value {
		if len(args) > 0 && args[0].IntVal > 0 {
			return lpc.IntValue(rand.Intn(args[0].IntVal))
		}
		return lpc.IntValue(0)
	})
	vm.RegisterEfun("write", func(args []lpc.Value) lpc.Value {
		if len(args) > 0 {
			fmt.Printf("[LPC Citizen#%d] %s\n", c.ID, args[0].String())
		}
		return lpc.Null()
	})

	// 调用 heart_beat
	_, err := vm.CallFunc("heart_beat", []lpc.Value{})
	if err != nil {
		fmt.Printf("[LPC] Citizen#%d heart_beat error: %v\n", c.ID, err)
	}

	// LPC 执行完后，仍然执行标准的代谢和衰老
	cell := w.GetCell(c.X, c.Y)
	harvested := cell.Harvest(cell.Sugar)
	c.Wealth += int(harvested)
	c.Wealth -= c.Metabolism
	c.Age++
	if c.Wealth <= 0 {
		c.Alive = false
	}
	if c.Age >= c.MaxAge {
		c.Alive = false
	}
}

// moveAndHarvest 执行单个公民的移动和收割
func (w *World) moveAndHarvest(c *Citizen) {
	// 在视觉范围内找到糖最多的格子
	bestX, bestY := c.X, c.Y
	bestSugar := w.GetCell(c.X, c.Y).Sugar

	for dy := -c.Vision; dy <= c.Vision; dy++ {
		for dx := -c.Vision; dx <= c.Vision; dx++ {
			if dx == 0 && dy == 0 {
				continue // 跳过当前位置
			}
			// 曼哈顿距离限制（视觉范围是菱形）
			if abs(dx)+abs(dy) > c.Vision {
				continue
			}
			cell := w.GetCell(c.X+dx, c.Y+dy)
			if cell.Sugar > bestSugar {
				bestSugar = cell.Sugar
				bestX = c.X + dx
				bestY = c.Y + dy
			}
		}
	}
	// 移动到最佳位置
	c.X = bestX
	c.Y = bestY
	// 收割糖（在新的环面坐标下）
	cell := w.GetCell(c.X, c.Y)
	harvested := cell.Harvest(cell.Sugar) // 收割所有可用糖
	c.Wealth += int(harvested)
	// 扣除代谢消耗
	c.Wealth -= c.Metabolism
	// 增加年龄
	c.Age++
	// 检查是否饿死
	if c.Wealth <= 0 {
		c.Alive = false
	}
	// 检查是否老死
	if c.Age >= c.MaxAge {
		c.Alive = false
	}
}

// ruleR 更替规则：移除死亡公民，在随机空位生成新公民
func (w *World) ruleR() {
	// 收集死亡公民的 ID
	var deadIDs []int
	for id, c := range w.Citizens {
		if c.IsDead() {
			deadIDs = append(deadIDs, id)
		}
	}
	// 移除死亡公民
	for _, id := range deadIDs {
		delete(w.Citizens, id)
	}
	// 在随机位置生成新公民（替换死亡的）
	for range deadIDs {
		// 随机选一个位置
		x := rand.Intn(w.Config.Width)
		y := rand.Intn(w.Config.Height)
		citizen := NewCitizen(w.NextID, x, y, rand.New(rand.NewSource(int64(w.NextID+w.Timestep))))
		w.Citizens[w.NextID] = citizen
		w.NextID++
	}
}
