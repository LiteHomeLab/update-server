package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
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

	var percentage float64
	if total > 0 {
		percentage = float64(downloaded) / float64(total) * 100
	}

	d.Progress = &ProgressInfo{
		Downloaded: downloaded,
		Total:      total,
		Percentage: percentage,
		Speed:      int64(speed),
	}
}

// SetCompleted 设置完成状态
func (d *DaemonState) SetCompleted(filePath string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.State = "completed"
	d.File = filePath
	d.Error = "" // Clear any previous error
}

// SetError 设置错误状态
func (d *DaemonState) SetError(err error) {
	if err == nil {
		return
	}

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

// DaemonServer Daemon HTTP 服务器
type DaemonServer struct {
	port        int
	server      *http.Server
	state       *DaemonState
	done        chan struct{}
	shutdownReq bool
	shutdownOnce sync.Once
	mu          sync.RWMutex
}

// NewDaemonServer 创建 Daemon 服务器
func NewDaemonServer(port int, state *DaemonState) *DaemonServer {
	return &DaemonServer{
		port:  port,
		state: state,
		done:  make(chan struct{}),
	}
}

// Start 启动 HTTP 服务器
func (d *DaemonServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/status", d.handleStatus)
	mux.HandleFunc("/shutdown", d.handleShutdown)

	d.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", d.port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("✓ Daemon mode started on port %d\n", d.port)

	// 检查端口是否可用
	listener, err := net.Listen("tcp", d.server.Addr)
	if err != nil {
		return fmt.Errorf("port %d is already in use. Try a different port", d.port)
	}

	return d.server.Serve(listener)
}

// handleStatus 处理 /status 请求
func (d *DaemonServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d.state.ToJSON())
}

// handleShutdown 处理 /shutdown 请求
func (d *DaemonServer) handleShutdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	d.mu.Lock()
	if d.shutdownReq {
		d.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "already_shutting_down",
			"message": "Already shutting down",
		})
		return
	}
	d.shutdownReq = true
	d.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Server shutting down",
	})

	// 异步关闭服务器
	go d.Shutdown()
}

// Shutdown 关闭服务器
func (d *DaemonServer) Shutdown() {
	d.shutdownOnce.Do(func() {
		if d.server != nil {
			d.server.Close()
		}
		if d.done != nil {
			close(d.done)
		}
	})
}
