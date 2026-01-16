package main

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// 设置环境变量
	os.Setenv("UPDATE_SERVER_URL", "http://test-server:8080")
	os.Setenv("UPDATE_TOKEN", "test-token")

	defer os.Unsetenv("UPDATE_SERVER_URL")
	defer os.Unsetenv("UPDATE_TOKEN")

	cfg := LoadConfig("http://override:9090", "override-token", "myapp")

	if cfg.ServerURL != "http://override:9090" {
		t.Errorf("Expected override URL, got %s", cfg.ServerURL)
	}
	if cfg.Token != "override-token" {
		t.Errorf("Expected override token, got %s", cfg.Token)
	}
	if cfg.ProgramID != "myapp" {
		t.Errorf("Expected program ID, got %s", cfg.ProgramID)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	os.Setenv("UPDATE_SERVER_URL", "http://env-server:8080")
	os.Setenv("UPDATE_TOKEN", "env-token")

	defer os.Unsetenv("UPDATE_SERVER_URL")
	defer os.Unsetenv("UPDATE_TOKEN")

	cfg := LoadConfig("", "", "myapp")

	if cfg.ServerURL != "http://env-server:8080" {
		t.Errorf("Expected env URL, got %s", cfg.ServerURL)
	}
	if cfg.Token != "env-token" {
		t.Errorf("Expected env token, got %s", cfg.Token)
	}
}
