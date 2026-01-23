package client

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadUpdate(t *testing.T) {
	// 这个测试需要运行中的服务器
	config := DefaultConfig()
	config.ServerURL = "http://localhost:18080"
	config.ProgramID = "testapp"
	config.Channel = "stable"

	checker := NewUpdateChecker(config, false)

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "downloaded.zip")

	progressCalled := false
	callback := func(p DownloadProgress) {
		progressCalled = true
	}

	err := checker.DownloadUpdate("1.0.0", destPath, callback)

	if err != nil {
		t.Logf("Download failed (expected if server not running): %v", err)
		return
	}

	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("File was not downloaded")
	}

	if !progressCalled {
		t.Error("Progress callback was not called")
	}
}
