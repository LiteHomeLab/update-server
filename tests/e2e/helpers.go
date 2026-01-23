package e2e

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
)

// E2ETestContext provides context for E2E tests
type E2ETestContext struct {
	T                testing.TB
	Server           *httptest.Server
	Browser          playwright.Browser
	Page             playwright.Page
	BaseURL          string
	AdminUsername    string
	AdminPassword    string
	AdminToken       string
	TestProgramID    string
	UploadToken      string
	DownloadToken    string
	TempDir          string
}

// SetupE2ETest initializes the E2E test environment
func SetupE2ETest(t testing.TB) *E2ETestContext {
	// Check if Playwright is available
	pw, err := playwright.Run()
	if err != nil {
		t.Skip("Playwright not available:", err)
		return nil
	}

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "e2e-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Setup test server
	server := setupTestServer(t)

	// Launch browser
	browser, err := pw.Chromium.Launch()
	if err != nil {
		t.Fatalf("Failed to launch browser: %v", err)
	}

	// Create context and page
	context, err := browser.NewContext()
	if err != nil {
		t.Fatalf("Failed to create browser context: %v", err)
	}

	page, err := context.NewPage()
	if err != nil {
		t.Fatalf("Failed to create page: %v", err)
	}

	return &E2ETestContext{
		T:             t,
		Server:        server,
		Browser:       browser,
		Page:          page,
		BaseURL:       server.URL,
		AdminUsername: "admin",
		AdminPassword: "admin123",
		TempDir:       tempDir,
	}
}

// Cleanup closes all resources
func (ctx *E2ETestContext) Cleanup() {
	if ctx.Page != nil {
		ctx.Page.Close()
	}
	if ctx.Browser != nil {
		ctx.Browser.Close()
	}
	if ctx.Server != nil {
		ctx.Server.Close()
	}
	if ctx.TempDir != "" {
		os.RemoveAll(ctx.TempDir)
	}
}

// NavigateToAdmin navigates to the admin panel
func (ctx *E2ETestContext) NavigateToAdmin() error {
	_, err := ctx.Page.Goto(ctx.BaseURL + "/admin")
	return err
}

// NavigateToLogin navigates to the login page
func (ctx *E2ETestContext) NavigateToLogin() error {
	_, err := ctx.Page.Goto(ctx.BaseURL + "/login")
	return err
}

// LoginAsAdmin performs admin login via the web UI
func (ctx *E2ETestContext) LoginAsAdmin() error {
	if err := ctx.NavigateToLogin(); err != nil {
		return err
	}

	// Wait for login form to load
	if _, err := ctx.Page.WaitForSelector("input[name='username']"); err != nil {
		return err
	}

	// Fill in credentials
	if err := ctx.Page.Fill("input[name='username']", ctx.AdminUsername); err != nil {
		return err
	}
	if err := ctx.Page.Fill("input[name='password']", ctx.AdminPassword); err != nil {
		return err
	}

	// Submit form
	if err := ctx.Page.Click("button[type='submit']"); err != nil {
		return err
	}

	// Wait for navigation to complete
	return ctx.Page.WaitForURL(ctx.BaseURL + "/admin")
}

// CreateProgramViaAPI creates a test program via API
func (ctx *E2ETestContext) CreateProgramViaAPI(name, description string) (string, error) {
	return createProgramViaHTTP(ctx.Server.URL, name, description)
}

// UploadVersionViaAPI uploads a version via API
func (ctx *E2ETestContext) UploadVersionViaAPI(programID, channel, version, zipPath string) error {
	return uploadVersionViaHTTP(ctx.Server.URL, programID, channel, version, zipPath)
}

// DownloadClient downloads the update client for a program
func (ctx *E2ETestContext) DownloadClient(programID, clientType string) (string, error) {
	return downloadClientFromServer(ctx.Server.URL, programID, clientType, ctx.TempDir)
}

// WaitForElement waits for an element to appear
func (ctx *E2ETestContext) WaitForElement(selector string, timeout ...time.Duration) error {
	to := float64(5000)
	if len(timeout) > 0 {
		to = float64(timeout[0].Milliseconds())
	}
	_, err := ctx.Page.WaitForSelector(selector, playwright.PageWaitForSelectorOptions{
		Timeout: &to,
	})
	return err
}

// ClickElement clicks an element after waiting for it
func (ctx *E2ETestContext) ClickElement(selector string) error {
	if err := ctx.WaitForElement(selector); err != nil {
		return err
	}
	return ctx.Page.Click(selector)
}

// FillInput fills an input field
func (ctx *E2ETestContext) FillInput(selector, value string) error {
	if err := ctx.WaitForElement(selector); err != nil {
		return err
	}
	return ctx.Page.Fill(selector, value)
}

// GetText gets text content of an element
func (ctx *E2ETestContext) GetText(selector string) (string, error) {
	if err := ctx.WaitForElement(selector); err != nil {
		return "", err
	}
	return ctx.Page.TextContent(selector)
}

// IsVisible checks if an element is visible
func (ctx *E2ETestContext) IsVisible(selector string) (bool, error) {
	element := ctx.Page.Locator(selector)
	return element.IsVisible()
}

// Screenshot takes a screenshot for debugging
func (ctx *E2ETestContext) Screenshot(name string) error {
	path := filepath.Join(ctx.TempDir, "screenshots", name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	_, err := ctx.Page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(path),
	})
	return err
}

// setupTestServer creates a test HTTP server
func setupTestServer(t testing.TB) *httptest.Server {
	// This would integrate with the existing test server setup
	// For now, create a simple mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock responses for testing
		switch r.URL.Path {
		case "/api/health":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		case "/login":
			// Return login page
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<html><body><form><input name="username"><input name="password" type="password"><button type="submit">Login</button></form></body></html>`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	return server
}

// createProgramViaHTTP creates a program via HTTP API
func createProgramViaHTTP(baseURL, name, description string) (string, error) {
	// Implementation would make HTTP POST to /api/admin/programs
	// For now, return a mock ID
	return fmt.Sprintf("prog-%d", time.Now().UnixNano()), nil
}

// uploadVersionViaHTTP uploads a version via HTTP API
func uploadVersionViaHTTP(baseURL, programID, channel, version, zipPath string) error {
	// Implementation would create multipart upload to /api/programs/{id}/versions
	return nil
}

// downloadClientFromServer downloads a client package
func downloadClientFromServer(baseURL, programID, clientType, destDir string) (string, error) {
	// Implementation would download from /api/admin/programs/{id}/client/{type}
	filePath := filepath.Join(destDir, fmt.Sprintf("%s-%s-client.zip", programID, clientType))
	return filePath, nil
}
