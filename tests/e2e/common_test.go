package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
)

// TestPlaywrightBasic verifies Playwright is properly installed
func TestPlaywrightBasic(t *testing.T) {
	pw, err := playwright.Run()
	if err != nil {
		t.Skip("Playwright not installed or not available:", err)
		return
	}
	defer pw.Stop()

	// Just verify we can access chromium
	_ = pw.Chromium
	t.Log("Playwright is available")
}
