// Package sim 实现 Sugarscape 的记分板统计系统
package sim

import (
	"sort"
)

// Scoreboard 记录世界仿真统计信息
type Scoreboard struct {
	Timestep        int     `json:"timestep"`
	Population      int     `json:"population"`
	BirthCount      int     `json:"birth_count"`
	DeathCount      int     `json:"death_count"`
	DeathByAge      int     `json:"death_by_age"`
	DeathByHunger   int     `json:"death_by_hunger"`
	AvgVision       float64 `json:"avg_vision"`
	AvgMetabolism   float64 `json:"avg_metabolism"`
	AvgAge          float64 `json:"avg_age"`
	AvgWealth       float64 `json:"avg_wealth"`
	GiniCoefficient float64 `json:"gini_coefficient"`
	BullCount       int     `json:"bull_count"`
	BearCount       int     `json:"bear_count"`
	TotalSugar      float64 `json:"total_sugar"`
}

// World 中的累计统计
var (
	globalBirthCount    int
	globalDeathCount    int
	globalDeathByAge    int
	globalDeathByHunger int
)

// UpdateScoreboard 从当前世界状态更新记分板
func (w *World) UpdateScoreboard() Scoreboard {
	alive := w.GetAliveCitizens()
	pop := len(alive)

	sb := Scoreboard{
		Timestep:     w.Timestep,
		Population:   pop,
		BirthCount:   globalBirthCount,
		DeathCount:   globalDeathCount,
		DeathByAge:   globalDeathByAge,
		DeathByHunger: globalDeathByHunger,
	}

	if pop == 0 {
		return sb
	}

	// 计算平均值
	var totalVision, totalMetabolism, totalAge, totalWealth float64
	wealths := make([]int, pop)

	for i, c := range alive {
		totalVision += float64(c.Vision)
		totalMetabolism += float64(c.Metabolism)
		totalAge += float64(c.Age)
		totalWealth += float64(c.Wealth)
		wealths[i] = c.Wealth

		// 统计策略分布
		if w.getStrategy(c) == StrategyBull {
			sb.BullCount++
		} else {
			sb.BearCount++
		}
	}

	sb.AvgVision = totalVision / float64(pop)
	sb.AvgMetabolism = totalMetabolism / float64(pop)
	sb.AvgAge = totalAge / float64(pop)
	sb.AvgWealth = totalWealth / float64(pop)
	sb.GiniCoefficient = calculateGini(wealths)

	// 总糖量
	totalSugar := 0.0
	for y := 0; y < w.Config.Height; y++ {
		for x := 0; x < w.Config.Width; x++ {
			totalSugar += w.Cells[y][x].Sugar
		}
	}
	sb.TotalSugar = totalSugar

	return sb
}

// calculateGini 计算基尼系数
// 基尼系数 = (2 * Σ(i * wealth_i)) / (n * Σwealth_i) - (n + 1) / n
func calculateGini(wealths []int) float64 {
	n := len(wealths)
	if n <= 1 {
		return 0.0
	}

	sort.Ints(wealths)

	var sum float64
	var totalWealth float64
	for i, w := range wealths {
		sum += float64(i+1) * float64(w)
		totalWealth += float64(w)
	}

	if totalWealth == 0 {
		return 0.0
	}

	gini := (2.0*sum)/(float64(n)*totalWealth) - (float64(n)+1.0)/float64(n)
	if gini < 0 {
		gini = 0
	}
	return gini
}

// RecordBirth 记录一次出生
func RecordBirth() {
	globalBirthCount++
}

// RecordDeath 记录一次死亡
func RecordDeath(byAge bool) {
	globalDeathCount++
	if byAge {
		globalDeathByAge++
	} else {
		globalDeathByHunger++
	}
}

// ResetGlobalStats 重置全局统计
func ResetGlobalStats() {
	globalBirthCount = 0
	globalDeathCount = 0
	globalDeathByAge = 0
	globalDeathByHunger = 0
}
