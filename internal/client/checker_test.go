package client

import (
	"testing"
)

func TestNewUpdateChecker(t *testing.T) {
	config := DefaultConfig()
	config.ProgramID = "testapp"

	checker := NewUpdateChecker(config, false)

	if checker == nil {
		t.Fatal("NewUpdateChecker returned nil")
	}

	if checker.config.GetProgramID() != "testapp" {
		t.Errorf("ProgramID = %s, want testapp", checker.config.GetProgramID())
	}
}
