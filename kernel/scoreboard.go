// Package kernel 实现最小内核的记分板记账
// 内核只记录生命周期计数（出生/死亡/人口）；游戏指标（基尼、均值、
// 策略分布、资源总量）由 LPC mudlib 通过 emit_stat 上报并合并进来
package kernel

// World 中的累计统计（生命周期计数属内核记账）
var (
	globalBirthCount     int
	globalDeathCount     int
	globalDeathsByReason = map[string]int{}
)

// GetAliveCitizens 获取所有存活实体列表
func (w *World) GetAliveCitizens() []*Citizen {
	alive := make([]*Citizen, 0, len(w.Citizens))
	for _, c := range w.Citizens {
		if c.Alive {
			alive = append(alive, c)
		}
	}
	return alive
}

// UpdateScoreboard 生成记分板：内核计数 + 本轮 mudlib emit_stat 上报的指标
func (w *World) UpdateScoreboard() map[string]any {
	reasons := make(map[string]int, len(globalDeathsByReason))
	for k, v := range globalDeathsByReason {
		reasons[k] = v
	}
	sb := map[string]any{
		"timestep":         w.Timestep,
		"population":       len(w.GetAliveCitizens()),
		"birth_count":      globalBirthCount,
		"death_count":      globalDeathCount,
		"deaths_by_reason": reasons,
	}
	for k, v := range w.gameStats {
		sb[k] = v
	}
	return sb
}

// RecordBirth 记录一次出生
func RecordBirth() {
	globalBirthCount++
}

// RecordDeath 记录一次死亡（reason 为任意 tag，内核不解释语义）
func RecordDeath(reason string) {
	globalDeathCount++
	globalDeathsByReason[reason]++
}

// ResetGlobalStats 重置全局统计
func ResetGlobalStats() {
	globalBirthCount = 0
	globalDeathCount = 0
	globalDeathsByReason = map[string]int{}
}
