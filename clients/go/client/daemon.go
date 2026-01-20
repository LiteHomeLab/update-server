package client

import (
	"sync"
)

// DaemonState 下载状态
type DaemonState struct {
	State    string         `json:"state"`     // idle | downloading | completed | error
	Version  string         `json:"version"`
	File     string         `json:"file"`
	Progress *ProgressInfo  `json:"progress,omitempty"`
	Error    string         `json:"error,omitempty"`
	mu       sync.RWMutex
}

// ProgressInfo 进度信息（用于 JSON 输出）
type ProgressInfo struct {
	Downloaded int64   `json:"downloaded"`
	Total      int64   `json:"total"`
	Percentage float64 `json:"percentage"`
	Speed      int64   `json:"speed"`
}

// NewDaemonState 创建新的状态管理器
func NewDaemonState(version string) *DaemonState {
	return &DaemonState{
		State:   "idle",
		Version: version,
	}
}

// SetState 设置状态
func (d *DaemonState) SetState(state string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.State = state
}

// GetState 获取状态（线程安全）
func (d *DaemonState) GetState() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.State
}

// SetProgress 设置进度
func (d *DaemonState) SetProgress(downloaded, total int64, speed float64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Progress = &ProgressInfo{
		Downloaded: downloaded,
		Total:      total,
		Percentage: float64(downloaded) / float64(total) * 100,
		Speed:      int64(speed),
	}
}

// SetCompleted 设置完成状态
func (d *DaemonState) SetCompleted(filePath string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.State = "completed"
	d.File = filePath
}

// SetError 设置错误状态
func (d *DaemonState) SetError(err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.State = "error"
	d.Error = err.Error()
}

// ToJSON 转换为 JSON（用于 HTTP 响应）
func (d *DaemonState) ToJSON() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := map[string]interface{}{
		"state":   d.State,
		"version": d.Version,
	}

	if d.File != "" {
		result["file"] = d.File
	}
	if d.Progress != nil {
		result["progress"] = d.Progress
	}
	if d.Error != "" {
		result["error"] = d.Error
	}

	return result
}
