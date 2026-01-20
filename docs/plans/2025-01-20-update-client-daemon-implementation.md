# Update Client Daemon Mode Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为 update-client 添加 Daemon 模式，允许调用程序通过 HTTP API 获取下载进度

**Architecture:**
- 新增 Daemon 服务器模块，使用 `net/http` 实现轻量级 HTTP API（/status 和 /shutdown）
- 线程安全的下载状态管理（使用 sync.RWMutex）
- 父进程监控机制（Windows 平台使用 API 检测父进程存活）
- 集成到现有的 download 命令，通过 `--daemon` 和 `--port` 参数启用

**Tech Stack:** Go (net/http, sync, os/exec), Cobra CLI

---

## Task 1: 创建 Daemon 状态管理数据结构

**Files:**
- Create: `clients/go/client/daemon.go`

**Step 1: 创建 daemon.go 文件，定义状态管理结构**

```go
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
```

**Step 2: 运行 go build 验证代码编译**

Run: `go build ./clients/go/client`
Expected: 成功编译，无错误

**Step 3: 提交**

```bash
git add clients/go/client/daemon.go
git commit -m "feat: add daemon state management structure"
```

---

## Task 2: 创建 Daemon HTTP 服务器

**Files:**
- Modify: `clients/go/client/daemon.go` (添加 HTTP 服务器代码)

**Step 1: 在 daemon.go 末尾添加 HTTP 服务器实现**

```go
package client

import (
	// ... 现有 imports
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// DaemonServer Daemon HTTP 服务器
type DaemonServer struct {
	port       int
	server     *http.Server
	state      *DaemonState
	done       chan struct{}
	shutdownReq bool
	mu         sync.RWMutex
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
		Addr:    fmt.Sprintf(":%d", d.port),
		Handler: mux,
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d.state.ToJSON())
}

// handleShutdown 处理 /shutdown 请求
func (d *DaemonServer) handleShutdown(w http.ResponseWriter, r *http.Request) {
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
	go d Shutdown()
}

// Shutdown 关闭服务器
func (d *DaemonServer) Shutdown() {
	select {
	case <-d.done:
		return // 已经关闭
	default:
	}

	close(d.done)
	d.server.Close()
}
```

**Step 2: 添加缺失的 import (net)**

在 `daemon.go` 文件顶部的 import 部分添加：

```go
import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
)
```

**Step 3: 运行 go build 验证**

Run: `go build ./clients/go/client`
Expected: 成功编译

**Step 4: 提交**

```bash
git add clients/go/client/daemon.go
git commit -m "feat: add daemon HTTP server with /status and /shutdown endpoints"
```

---

## Task 3: 创建父进程监控模块（Windows 平台）

**Files:**
- Create: `clients/go/client/parent_windows.go`

**Step 1: 创建 Windows 平台的父进程监控实现**

```go
//go:build windows

package client

import (
	"log"
	"syscall"
	"time"
	"unsafe"
)

var (
	modkernel32           = syscall.NewLazyDLL("kernel32.dll")
	procQueryFullProcessImageName = modkernel32.NewProc("QueryFullProcessImageNameW")
)

// monitorParentProcess 监控父进程存活状态
func (d *DaemonServer) monitorParentProcess(parentPID int) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !d.isParentAlive(parentPID) {
				log.Println("Parent process died, shutting down")
				d.Shutdown()
				return
			}
		case <-d.done:
			return
		}
	}
}

// isParentAlive 检查父进程是否存活
func (d *DaemonServer) isParentAlive(pid int) bool {
	if pid == 0 {
		return true // 无法获取父进程 PID，假设存活
	}

	handle, err := syscall.OpenProcess(syscall.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle)

	var exitCode uint32
	err = syscall.GetExitCodeProcess(handle, &exitCode)
	return err == nil && exitCode == 259 // STILL_ACTIVE = 259
}

// GetParentPID 获取父进程 PID
func GetParentPID() int {
	// Windows: 获取当前进程的父进程 PID
	// 使用 NtQueryInformationProcess 或其他 API
	// 简化实现：返回 0 表示未实现
	return 0
}
```

**Step 2: 创建非 Windows 平台的占位实现**

**File:** Create: `clients/go/client/parent_other.go`

```go
//go:build !windows

package client

import (
	"log"
	"time"
)

// monitorParentProcess 监控父进程存活状态（非 Windows 平台）
func (d *DaemonServer) monitorParentProcess(parentPID int) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !d.isParentAlive(parentPID) {
				log.Println("Parent process died, shutting down")
				d.Shutdown()
				return
			}
		case <-d.done:
			return
		}
	}
}

// isParentAlive 检查父进程是否存活（非 Windows）
func (d *DaemonServer) isParentAlive(pid int) bool {
	if pid == 0 {
		return true
	}
	// Unix: 检查进程是否存在
	// 简化实现
	return true
}

// GetParentPID 获取父进程 PID（非 Windows）
func GetParentPID() int {
	return 0
}
```

**Step 3: 运行 go build 验证**

Run: `go build ./clients/go/client`
Expected: 成功编译

**Step 4: 提交**

```bash
git add clients/go/client/parent_windows.go clients/go/client/parent_other.go
git commit -m "feat: add parent process monitoring for daemon mode"
```

---

## Task 4: 修改 downloader.go 支持状态更新回调

**Files:**
- Modify: `clients/go/client/downloader.go`

**Step 1: 在 downloadOnce 函数中添加状态回调支持**

找到 `func (c *UpdateChecker) downloadOnce(...)` 函数，修改进度回调部分以支持状态更新。

首先在 `UpdateChecker` 结构体中添加 `daemonState` 字段（如果不存在）：

查看 `clients/go/client/checker.go` 文件：
```go
type UpdateChecker struct {
	config       *Config
	httpClient   *http.Client
	jsonOutput   bool
	maxRetries   int
	daemonState  *DaemonState // 新增：Daemon 状态管理器
}
```

**Step 2: 修改 downloader.go 的 downloadOnce 函数**

在进度回调部分，更新 daemonState：

找到 `downloadOnce` 函数中的进度回调代码块（约第 210-224 行），修改为：

```go
			// 调用进度回调
			if callback != nil && total > 0 {
				elapsed := time.Since(startTime).Seconds()
				speed := float64(downloaded) / elapsed
				if elapsed == 0 {
					speed = 0
				}

				progress := DownloadProgress{
					Version:    version,
					Downloaded: downloaded,
					Total:      total,
					Percentage: float64(downloaded) / float64(total) * 100,
					Speed:      speed,
				}

				callback(progress)

				// 更新 Daemon 状态（如果存在）
				if c.daemonState != nil {
					c.daemonState.SetProgress(downloaded, total, speed)
				}
			}
```

**Step 3: 在下载开始前设置状态为 downloading**

在 `downloadOnce` 函数开始处添加：

```go
func (c *UpdateChecker) downloadOnce(version string, destPath string, callback ProgressCallback) error {
	// 设置状态为 downloading
	if c.daemonState != nil {
		c.daemonState.SetState("downloading")
	}

	// ... 原有代码
```

**Step 4: 在下载成功后设置 completed 状态**

在 `downloadOnce` 函数返回 `nil` 之前添加：

```go
	// 下载成功
	if c.daemonState != nil {
		c.daemonState.SetCompleted(destPath)
	}
	return nil
```

**Step 5: 运行 go build 验证**

Run: `go build ./clients/go/client`
Expected: 成功编译

**Step 6: 提交**

```bash
git add clients/go/client/downloader.go
git commit -m "feat: integrate daemon state updates in download progress"
```

---

## Task 5: 修改 DownloadWithOutput 函数支持错误状态

**Files:**
- Modify: `clients/go/client/downloader.go`

**Step 1: 修改 DownloadWithOutput 函数以设置错误状态**

找到 `DownloadWithOutput` 函数，添加错误处理：

```go
// DownloadWithOutput 下载更新并输出结果
func (c *UpdateChecker) DownloadWithOutput(version string, outputPath string) error {
	// 设置初始状态
	if c.daemonState != nil {
		c.daemonState.SetState("idle")
	}

	if outputPath == "" {
		outputPath = c.generateOutputPath(version)
	}

	// Download
	if !c.jsonOutput {
		fmt.Printf("✓ Starting download: %s\n", filepath.Base(outputPath))
	}
	info, err := c.CheckUpdate("")
	if err != nil {
		if c.daemonState != nil {
			c.daemonState.SetError(err)
		}
		return c.outputError(err)
	}
	if info != nil && !c.jsonOutput {
		fmt.Printf("  Size: %.1f MB\n", float64(info.FileSize)/1024/1024)
	}

	if err := c.DownloadUpdate(version, outputPath, c.progressCallback); err != nil {
		if c.daemonState != nil {
			c.daemonState.SetError(err)
		}
		return c.outputError(err)
	}

	// Verify
	verified := true
	if info != nil && info.FileHash != "" {
		verified, _ = c.VerifyFile(outputPath, info.FileHash)
		if !verified {
			err := fmt.Errorf("verification failed")
			if c.daemonState != nil {
				c.daemonState.SetError(err)
			}
			return c.outputError(err)
		}
	}

	// Decrypt if key is available
	decrypted := false
	if c.config.Auth.EncryptionKey != "" {
		decryptor, err := NewDecryptor(c.config.Auth.EncryptionKey)
		if err == nil {
			if err := decryptor.DecryptFile(outputPath, outputPath); err == nil {
				decrypted = true
			}
		}
	}

	return c.outputDownloadResult(outputPath, decrypted, verified)
}
```

**Step 2: 运行 go build 验证**

Run: `go build ./clients/go/client`
Expected: 成功编译

**Step 3: 提交**

```bash
git add clients/go/client/downloader.go
git commit -m "feat: add error state handling in DownloadWithOutput"
```

---

## Task 6: 修改 main.go 添加 Daemon 模式参数和逻辑

**Files:**
- Modify: `cmd/update-client/main.go`

**Step 1: 添加 Daemon 模式标志变量**

在文件顶部的变量声明区域添加：

```go
var (
	cfgFile    string
	jsonOutput bool
	daemonMode bool
	daemonPort int
)
```

**Step 2: 在 downloadCmd 中添加 Daemon 标志**

修改 `init()` 函数中的 downloadCmd 部分：

```go
func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "update-config.yaml", "config file")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output JSON format")

	checkCmd.Flags().String("current-version", "", "current version (overrides config)")

	downloadCmd.Flags().String("output", "", "output file path")
	downloadCmd.Flags().String("version", "", "version to download")
	downloadCmd.Flags().BoolVar(&daemonMode, "daemon", false, "enable daemon mode (HTTP server)")
	downloadCmd.Flags().IntVar(&daemonPort, "port", 0, "HTTP server port (required with --daemon)")
	downloadCmd.MarkFlagRequired("version")

	rootCmd.AddCommand(checkCmd, downloadCmd)
}
```

**Step 3: 修改 runDownload 函数支持 Daemon 模式**

完全替换 `runDownload` 函数：

```go
func runDownload(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := client.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get flags
	version, _ := cmd.Flags().GetString("version")
	outputPath, _ := cmd.Flags().GetString("output")
	daemon, _ := cmd.Flags().GetBool("daemon")
	port, _ := cmd.Flags().GetInt("port")

	// 验证 Daemon 模式参数
	if daemon && port == 0 {
		return fmt.Errorf("--port is required when using --daemon")
	}

	// Daemon 模式
	if daemon {
		return runDaemonDownload(cfg, version, outputPath, port)
	}

	// 普通模式
	checker := client.NewUpdateChecker(cfg, jsonOutput)
	return checker.DownloadWithOutput(version, outputPath)
}
```

**Step 4: 添加 runDaemonDownload 函数**

在 `main.go` 文件末尾添加：

```go
func runDaemonDownload(cfg *client.Config, version, outputPath string, port int) error {
	// 创建状态管理器
	state := client.NewDaemonState(version)

	// 创建并启动 Daemon 服务器
	server := client.NewDaemonServer(port, state)

	// 创建 checker 并设置 daemonState
	checker := client.NewUpdateChecker(cfg, false)
	checker.SetDaemonState(state)

	// 启动父进程监控
	parentPID := client.GetParentPID()
	go server.monitorParentProcess(parentPID)

	// 在后台启动 HTTP 服务器
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Start()
	}()

	// 等待服务器启动或失败
	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("failed to start daemon server: %w", err)
		}
	case <-time.After(500 * time.Millisecond):
		// 服务器启动成功
	}

	// 执行下载
	downloadErr := checker.DownloadWithOutput(version, outputPath)

	// 等待 shutdown 信号
	<-server.Done()

	return downloadErr
}
```

**Step 5: 修改 checker.go 添加 SetDaemonState 方法**

**File:** Modify: `clients/go/client/checker.go`

在 `UpdateChecker` 结构体中添加 `daemonState` 字段，并添加 setter 方法：

```go
// UpdateChecker 更新检查器
type UpdateChecker struct {
	config       *Config
	httpClient   *http.Client
	jsonOutput   bool
	maxRetries   int
	daemonState  *DaemonState // Daemon 状态管理器
}

// SetDaemonState 设置 Daemon 状态管理器
func (c *UpdateChecker) SetDaemonState(state *DaemonState) {
	c.daemonState = state
}
```

**Step 6: 运行 go build 验证**

Run: `go build ./cmd/update-client`
Expected: 成功编译

**Step 7: 提交**

```bash
git add cmd/update-client/main.go clients/go/client/checker.go
git commit -m "feat: add daemon mode CLI flags and integration"
```

---

## Task 7: 在 daemon.go 中添加 Done() 方法

**Files:**
- Modify: `clients/go/client/daemon.go`

**Step 1: 在 DaemonServer 结构体中添加 Done 方法**

在 `daemon.go` 中的 `Shutdown()` 方法后添加：

```go
// Done 返回关闭信号 channel
func (d *DaemonServer) Done() <-chan struct{} {
	return d.done
}
```

**Step 2: 运行 go build 验证**

Run: `go build ./clients/go/client`
Expected: 成功编译

**Step 3: 提交**

```bash
git add clients/go/client/daemon.go
git commit -m "feat: add Done() method to DaemonServer"
```

---

## Task 8: 编写 Daemon 模式的单元测试

**Files:**
- Create: `clients/go/client/daemon_test.go`

**Step 1: 创建测试文件**

```go
package client

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestDaemonState(t *testing.T) {
	state := NewDaemonState("1.0.0")

	// 测试初始状态
	if state.GetState() != "idle" {
		t.Errorf("Expected initial state 'idle', got '%s'", state.GetState())
	}

	// 测试状态转换
	state.SetState("downloading")
	if state.GetState() != "downloading" {
		t.Errorf("Expected state 'downloading', got '%s'", state.GetState())
	}

	// 测试进度更新
	state.SetProgress(1024, 2048, 1024.0)
	jsonData := state.ToJSON()
	if progress, ok := jsonData["progress"].(map[string]interface{}); ok {
		if progress["percentage"].(float64) != 50.0 {
			t.Errorf("Expected percentage 50.0, got %v", progress["percentage"])
		}
	}

	// 测试完成状态
	state.SetCompleted("/path/to/file.zip")
	if state.GetState() != "completed" {
		t.Errorf("Expected state 'completed', got '%s'", state.GetState())
	}

	// 测试错误状态
	state.SetError(os.ErrExist)
	if state.GetState() != "error" {
		t.Errorf("Expected state 'error', got '%s'", state.GetState())
	}
}

func TestDaemonServer(t *testing.T) {
	state := NewDaemonState("1.0.0")
	server := NewDaemonServer(19876, state)

	// 启动服务器
	go func() {
		if err := server.Start(); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(500 * time.Millisecond)

	// 测试 /status 端点
	resp, err := http.Get("http://localhost:19876/status")
	if err != nil {
		t.Fatalf("Failed to call /status: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	if result["state"] != "idle" {
		t.Errorf("Expected state 'idle', got '%v'", result["state"])
	}

	if result["version"] != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%v'", result["version"])
	}

	// 测试 /shutdown 端点
	resp2, err := http.Post("http://localhost:19876/shutdown", "application/json", nil)
	if err != nil {
		t.Fatalf("Failed to call /shutdown: %v", err)
	}
	defer resp2.Body.Close()

	var shutdownResult map[string]interface{}
	if err := json.NewDecoder(resp2.Body).Decode(&shutdownResult); err != nil {
		t.Fatalf("Failed to decode shutdown JSON: %v", err)
	}

	if shutdownResult["success"] != true {
		t.Errorf("Expected success=true, got %v", shutdownResult["success"])
	}

	// 等待服务器关闭
	<-server.Done()
	time.Sleep(100 * time.Millisecond)
}
```

**Step 2: 运行测试**

Run: `go test ./clients/go/client -v -run TestDaemon`
Expected: 测试通过

**Step 3: 提交**

```bash
git add clients/go/client/daemon_test.go
git commit -m "test: add unit tests for daemon state and server"
```

---

## Task 9: 构建 update-client 并进行手动测试

**Files:**
- Build: `cmd/update-client`

**Step 1: 构建可执行文件**

Run: `go build -o bin/update-client.exe ./cmd/update-client`
Expected: 成功生成 `bin/update-client.exe`

**Step 2: 测试普通模式（确保未破坏现有功能）**

Run: `./bin/update-client.exe check --json`
Expected: 返回 JSON 格式的检查结果

**Step 3: 测试 Daemon 模式启动**

Run: `./bin/update-client.exe download --daemon --port 19876 --version 1.0.0`
Expected: 输出 "✓ Daemon mode started on port 19876" 并保持运行

**Step 4: 测试 /status 端点**

在另一个终端运行：
```bash
curl http://localhost:19876/status
```
Expected: 返回 JSON 状态，包含 `"state": "downloading"` 或 `"state": "idle"`

**Step 5: 测试 /shutdown 端点**

Run: `curl -X POST http://localhost:19876/shutdown`
Expected: 返回 `{"success": true, "message": "Server shutting down"}`

**Step 6: 测试端口冲突**

启动第一个实例：`./bin/update-client.exe download --daemon --port 19876 --version 1.0.0`

在另一个终端启动第二个实例：`./bin/update-client.exe download --daemon --port 19876 --version 1.0.0`
Expected: 第二个实例立即退出，显示端口占用错误

**Step 7: 提交构建脚本（如果需要）**

创建 `build-client.bat`：
```batch
@echo off
echo Building update-client...
go build -o bin/update-client.exe ./cmd/update-client
if errorlevel 1 (
    echo Build failed
    exit /b 1
)
echo Build successful: bin/update-client.exe
```

```bash
git add build-client.bat
git commit -m "build: add update-client build script"
```

---

## Task 10: 更新使用文档

**Files:**
- Modify: `docs/UPDATE_CLIENT_GUIDE.md`

**Step 1: 验证文档中的 Daemon 模式章节**

确认 `docs/UPDATE_CLIENT_GUIDE.md` 文件中已包含：
- Daemon 模式概述
- HTTP API 文档（/status 和 /shutdown）
- 命令行参数说明（--daemon 和 --port）
- 集成示例（C#、Go、Python）

文档已在之前的会话中更新，此任务为验证任务。

**Step 2: 如有需要，添加或更新内容**

如果需要更新，确保包含：
```markdown
## Daemon 模式（后台进度监控）

当需要实时监控下载进度时，使用 `--daemon` 参数启动独立的 HTTP 服务器。

### 启动 Daemon 模式

```bash
update-client.exe download --daemon --port 19876 --version 1.2.0
```

输出：
```
✓ Daemon mode started on port 19876
✓ Downloading: app-v1.2.0.zip
✓ Monitoring parent process (PID: 12345)
```

### HTTP API

**GET /status** - 获取下载状态

```json
{
  "state": "downloading",
  "version": "1.2.0",
  "file": "./updates/app-v1.2.0.zip",
  "progress": {
    "downloaded": 52428800,
    "total": 104857600,
    "percentage": 50.0,
    "speed": 8912896
  },
  "error": ""
}
```

**POST /shutdown** - 关闭服务器

```bash
curl -X POST http://localhost:19876/shutdown
```

响应：
```json
{
  "success": true,
  "message": "Server shutting down"
}
```
```

**Step 3: 提交文档更新（如需要）**

```bash
git add docs/UPDATE_CLIENT_GUIDE.md
git commit -m "docs: verify daemon mode documentation is complete"
```

---

## 验证清单

完成所有任务后，验证以下功能：

- [ ] 普通模式下载仍然正常工作
- [ ] Daemon 模式可以启动 HTTP 服务器
- [ ] /status 端点返回正确的 JSON 格式状态
- [ ] /shutdown 端点可以正确关闭服务器
- [ ] 端口冲突时显示错误并退出
- [ ] 下载进度实时更新到状态
- [ ] 下载完成后状态变为 completed
- [ ] 下载失败时状态变为 error 并包含错误信息
- [ ] 父进程退出时 Daemon 自动关闭（Windows 平台）
- [ ] 所有单元测试通过
- [ ] 文档完整且准确
