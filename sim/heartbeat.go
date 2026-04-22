// Package sim 实现 Sugarscape 仿真核心模型和规则
package sim

// HeartBeatEntry 注册了 heart beat 的对象
type HeartBeatEntry struct {
	ObjectID    string // LPC 对象 ID
	CitizenID   int    // 对应的公民 ID
	Interval    int    // 每 N 步调用一次
	LastCall    int    // 上次调用的步数
}

// HeartBeatManager 管理所有 heart beat 回调
type HeartBeatManager struct {
	entries []*HeartBeatEntry
}

// NewHeartBeatManager 创建新的 heart beat 管理器
func NewHeartBeatManager() *HeartBeatManager {
	return &HeartBeatManager{
		entries: make([]*HeartBeatEntry, 0),
	}
}

// Register 注册一个 heart beat 回调
func (h *HeartBeatManager) Register(objectID string, citizenID, interval int) {
	h.entries = append(h.entries, &HeartBeatEntry{
		ObjectID:  objectID,
		CitizenID: citizenID,
		Interval:  interval,
		LastCall:  0,
	})
}

// Unregister 取消注册
func (h *HeartBeatManager) Unregister(objectID string) {
	for i, e := range h.entries {
		if e.ObjectID == objectID {
			h.entries = append(h.entries[:i], h.entries[i+1:]...)
			return
		}
	}
}

// GetEntries 获取所有注册的条目
func (h *HeartBeatManager) GetEntries() []*HeartBeatEntry {
	return h.entries
}

// ShouldCall 判断当前步是否应该调用该条目的 heart_beat
func (e *HeartBeatEntry) ShouldCall(timestep int) bool {
	if e.Interval <= 0 {
		return true
	}
	return timestep-e.LastCall >= e.Interval
}

// MarkCalled 标记已调用
func (e *HeartBeatEntry) MarkCalled(timestep int) {
	e.LastCall = timestep
}
