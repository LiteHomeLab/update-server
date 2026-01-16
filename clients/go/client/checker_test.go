package client

import (
	"testing"
)

func TestNewUpdateChecker(t *testing.T) {
	config := DefaultConfig()
	config.ProgramID = "testapp"

	checker := NewUpdateChecker(config)

	if checker == nil {
		t.Fatal("NewUpdateChecker returned nil")
	}

	if checker.config.ProgramID != "testapp" {
		t.Errorf("ProgramID = %s, want testapp", checker.config.ProgramID)
	}
}
