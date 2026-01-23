package client

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Default(t *testing.T) {
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.Server.URL != "http://localhost:8080" {
		t.Errorf("Expected default URL, got %s", cfg.Server.URL)
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  url: "http://test-server:9000"
  timeout: 60
program:
  id: "test-app"
  current_version: "1.0.0"
auth:
  token: "test-token"
download:
  save_path: "./test-updates"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.Server.URL != "http://test-server:9000" {
		t.Errorf("Expected test-server URL, got %s", cfg.Server.URL)
	}

	if cfg.Program.ID != "test-app" {
		t.Errorf("Expected test-app, got %s", cfg.Program.ID)
	}

	// Test backward compatibility methods
	if cfg.GetProgramID() != "test-app" {
		t.Errorf("GetProgramID() = %s, want test-app", cfg.GetProgramID())
	}
}
