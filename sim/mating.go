// Package sim 实现 Sugarscape 中的繁殖系统
package sim



// MateResult 表示一次繁殖结果
type MateResult struct {
	ParentAID int `json:"parent_a_id"`
	ParentBID int `json:"parent_b_id"`
	ChildID   int `json:"child_id"`
	Timestep  int `json:"timestep"`
}

// ExecuteMating 执行一轮繁殖
// 条件：双亲都适龄、财富超过初始禀赋、相邻
func (w *World) ExecuteMating() []MateResult {
	if !w.Config.EnableMating {
		return nil
	}

	var results []MateResult
	alive := w.GetAliveCitizens()
	mated := make(map[int]bool) // 标记已参与繁殖的 citizen

	for i := 0; i < len(alive); i++ {
		a := alive[i]
		if mated[a.ID] {
			continue
		}
		// 检查 A 的繁殖条件
		if !w.canMate(a) {
			continue
		}

		for j := i + 1; j < len(alive); j++ {
			b := alive[j]
			if mated[b.ID] {
				continue
			}
			if !w.canMate(b) {
				continue
			}
			if !w.areAdjacent(a, b) {
				continue
			}

			// 执行繁殖
			child := w.reproduce(a, b)
			if child != nil {
				mated[a.ID] = true
				mated[b.ID] = true
				results = append(results, MateResult{
					ParentAID: a.ID,
					ParentBID: b.ID,
					ChildID:   child.ID,
					Timestep:  w.Timestep,
				})
				break // 每个 citizen 每轮只能繁殖一次
			}
		}
	}

	return results
}

// canMate 检查 citizen 是否满足繁殖条件
func (w *World) canMate(c *Citizen) bool {
	if !c.Alive {
		return false
	}
	// 年龄条件：至少 10 步，距离最大年龄还有 10 步以上
	if c.Age < 10 || c.Age > c.MaxAge-10 {
		return false
	}
	// 财富条件：超过初始禀赋的上界 (25)
	if c.Wealth <= 25 {
		return false
	}
	return true
}

// reproduce 从两个亲代产生一个后代
func (w *World) reproduce(a, b *Citizen) *Citizen {
	// 遗传：后代的 vision 和 metabolism 从双亲随机融合
	rng := w.rng

	vision := a.Vision
	if rng.Intn(2) == 0 {
		vision = b.Vision
	}
	// 小概率变异
	if rng.Intn(10) == 0 {
		vision = rng.Intn(6) + 1
	}

	metabolism := a.Metabolism
	if rng.Intn(2) == 0 {
		metabolism = b.Metabolism
	}
	if rng.Intn(10) == 0 {
		metabolism = rng.Intn(4) + 1
	}

	maxAge := a.MaxAge
	if rng.Intn(2) == 0 {
		maxAge = b.MaxAge
	}
	if rng.Intn(10) == 0 {
		maxAge = rng.Intn(41) + 60
	}

	// 创建后代
	child := &Citizen{
		ID:         w.NextID,
		X:          a.X,
		Y:          a.Y,
		Vision:     vision,
		Metabolism: metabolism,
		MaxAge:     maxAge,
		Age:        0,
		Wealth:     0, // 先设为 0，下面继承遗产
		Alive:      true,
	}
	w.NextID++

	// 遗产继承：双方各转移初始禀赋的一半给子代
	inheritanceA := 13 // 25 / 2 + 1
	inheritanceB := 13
	if a.Wealth < inheritanceA {
		inheritanceA = a.Wealth
	}
	if b.Wealth < inheritanceB {
		inheritanceB = b.Wealth
	}
	child.Wealth = inheritanceA + inheritanceB
	a.Wealth -= inheritanceA
	b.Wealth -= inheritanceB

	// 如果后代出生位置已有太多人，随机偏移
	child.X = (child.X + rng.Intn(3) - 1 + w.Config.Width) % w.Config.Width
	child.Y = (child.Y + rng.Intn(3) - 1 + w.Config.Height) % w.Config.Height

	w.Citizens[child.ID] = child
	return child
}
