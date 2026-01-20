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
