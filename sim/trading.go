// Package sim 实现 Sugarscape 中的易货贸易系统
package sim

// TradingStrategy 表示贸易策略
type TradingStrategy int

const (
	StrategyBull TradingStrategy = iota // Bull 策略：高频小额
	StrategyBear                        // Bear 策略：低频大额
)

// TradeResult 表示一次交易结果
type TradeResult struct {
	CitizenAID int     `json:"citizen_a_id"`
	CitizenBID int     `json:"citizen_b_id"`
	AmountA    int     `json:"amount_a"` // A 给出的糖
	AmountB    int     `json:"amount_b"` // B 给出的糖
	Timestep   int     `json:"timestep"`
}

// ExecuteTrading 执行一轮贸易
// 遍历所有存活的 Citizen，对相邻的进行易货贸易
func (w *World) ExecuteTrading() []TradeResult {
	if !w.Config.EnableTrading {
		return nil
	}

	var results []TradeResult
	alive := w.GetAliveCitizens()

	for i := 0; i < len(alive); i++ {
		for j := i + 1; j < len(alive); j++ {
			a := alive[i]
			b := alive[j]
			// 检查是否相邻（曼哈顿距离 <= 视觉范围交集）
			if !w.areAdjacent(a, b) {
				continue
			}
			// 根据边际效用计算是否值得交易
			result := w.tryTrade(a, b)
			if result != nil {
				result.Timestep = w.Timestep
				results = append(results, *result)
			}
		}
	}

	return results
}

// areAdjacent 判断两个 citizen 是否相邻（距离 <= min(双方视觉)）
func (w *World) areAdjacent(a, b *Citizen) bool {
	dx := torusDist(a.X, a.Y, b.X, b.Y, w.Config.Width, w.Config.Height)
	maxDist := a.Vision
	if b.Vision < maxDist {
		maxDist = b.Vision
	}
	return dx <= float64(maxDist)
}

// tryTrade 尝试在两个 citizen 之间进行交易
// 基于边际效用：如果 A 的糖对 B 的边际效用 > B 的糖对 A 的边际效用，则交易
func (w *World) tryTrade(a, b *Citizen) *TradeResult {
	// 计算边际效用
	muA := w.marginalUtility(a)
	muB := w.marginalUtility(b)

	// 如果双方的边际效用差异不足以驱动交易，跳过
	if muA <= 0 && muB <= 0 {
		return nil
	}

	// 确定策略
	strategyA := w.getStrategy(a)
	strategyB := w.getStrategy(b)

	var amountA, amountB int

	if strategyA == StrategyBull && strategyB == StrategyBull {
		// 双方 Bull：小额高频
		amountA = 1
		amountB = 1
	} else if strategyA == StrategyBear && strategyB == StrategyBear {
		// 双方 Bear：大额低频
		amountA = a.Wealth / 4
		amountB = b.Wealth / 4
	} else {
		// 混合：折中
		amountA = 2
		amountB = 2
	}

	// 根据边际效用调整交换比例
	if muA > muB && amountB > 0 {
		// A 更需要糖，B 少给一点
		ratio := muA / (muA + muB + 0.001)
		amountB = max(1, int(float64(amountB)*(1.0-ratio)))
	} else if muB > muA && amountA > 0 {
		ratio := muB / (muA + muB + 0.001)
		amountA = max(1, int(float64(amountA)*(1.0-ratio)))
	}

	// 确保双方有足够的糖
	if a.Wealth < amountA || b.Wealth < amountB {
		return nil
	}
	if amountA <= 0 && amountB <= 0 {
		return nil
	}

	// 执行交易
	a.Wealth -= amountA
	a.Wealth += amountB
	b.Wealth -= amountB
	b.Wealth += amountA

	return &TradeResult{
		CitizenAID: a.ID,
		CitizenBID: b.ID,
		AmountA:    amountA,
		AmountB:    amountB,
	}
}

// marginalUtility 计算 citizen 的糖边际效用
// 财富越少，边际效用越高
func (w *World) marginalUtility(c *Citizen) float64 {
	if c.Wealth <= 0 {
		return 100.0 // 濒死状态，效用极高
	}
	// 效用 = 1 / (wealth + 1)，财富越少效用越高
	return 1.0 / float64(c.Wealth+1)
}

// getStrategy 根据 citizen 属性判断贸易策略
func (w *World) getStrategy(c *Citizen) TradingStrategy {
	// 高代谢率 → Bull（需要频繁补充）
	// 低代谢率 → Bear（可以等待）
	if c.Metabolism >= 3 {
		return StrategyBull
	}
	return StrategyBear
}

// GetAliveCitizens 获取所有存活的 citizen 列表
func (w *World) GetAliveCitizens() []*Citizen {
	alive := make([]*Citizen, 0, len(w.Citizens))
	for _, c := range w.Citizens {
		if c.Alive {
			alive = append(alive, c)
		}
	}
	return alive
}
