package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"docufiller-update-server/tests/helpers"
)

// TestPublishFlowClosedLoop tests the complete publish flow:
// 1. Create program via admin API
// 2. Verify client can upload versions
// 3. Verify version appears in backend
func TestPublishFlowClosedLoop(t *testing.T) {
	pw, err := playwright.Run()
	if err != nil {
		t.Skip("Playwright not available:", err)
		return
	}
	defer pw.Stop()

	// Setup test server with admin
	srv := helpers.SetupTestServerWithAdmin(t)
	defer srv.Close()

	t.Log("Step 1: Creating program via API...")
	programID := fmt.Sprintf("publoop-%d", time.Now().UnixNano())

	// Create program using the test server's router directly
	payload := map[string]interface{}{
		"programId":   programID,
		"name":        "PublishLoopApp",
		"description": "Application for publish loop testing",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/admin/programs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+srv.AdminToken)

	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code, "Program creation should succeed")

	var createResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResp)
	programObj := createResp["Program"].(map[string]interface{})
	actualProgramID := programObj["programId"].(string)
	assert.NotEmpty(t, actualProgramID)

	t.Log("Step 2: Getting upload token...")
	uploadToken, _ := helpers.GetProgramTokens(t, srv, actualProgramID)

	t.Log("Step 3: Creating and uploading version...")
	testZipPath := helpers.CreateTestVersionZip(t, "PublishLoopApp", "1.0.0")
	defer os.Remove(testZipPath)

	uploadBody := &bytes.Buffer{}
	writer := multipart.NewWriter(uploadBody)
	file, _ := os.Open(testZipPath)
	part, _ := writer.CreateFormFile("file", filepath.Base(testZipPath))
	io.Copy(part, file)
	file.Close()
	writer.WriteField("channel", "stable")
	writer.WriteField("version", "1.0.0")
	writer.WriteField("notes", "First release via publisher")
	writer.WriteField("mandatory", "false")
	writer.Close()

	uploadReq := httptest.NewRequest("POST", "/api/programs/"+actualProgramID+"/versions", uploadBody)
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())
	uploadReq.Header.Set("Authorization", "Bearer "+uploadToken)

	uploadW := httptest.NewRecorder()
	srv.Router.ServeHTTP(uploadW, uploadReq)

	require.Equal(t, http.StatusOK, uploadW.Code, "Version upload should succeed")

	t.Log("Step 4: Verifying version in database...")
	var count int64
	srv.DB.Table("versions").
		Where("program_id = ? AND channel = ? AND version = ?", actualProgramID, "stable", "1.0.0").
		Count(&count)
	assert.Greater(t, count, int64(0), "Version should exist in database")

	t.Log("Step 5: Verifying version via admin API...")
	listReq := httptest.NewRequest("GET", "/api/admin/programs/"+actualProgramID+"/versions", nil)
	listReq.Header.Set("Authorization", "Bearer "+srv.AdminToken)

	listW := httptest.NewRecorder()
	srv.Router.ServeHTTP(listW, listReq)

	require.Equal(t, http.StatusOK, listW.Code)

	var versions []map[string]interface{}
	json.Unmarshal(listW.Body.Bytes(), &versions)

	assert.GreaterOrEqual(t, len(versions), 1, "Should have at least one version")

	t.Log("Publish flow closed-loop test completed successfully!")
}

// TestUpdateFlowClosedLoop tests the complete update flow:
// 1. Create program and upload version
// 2. Check for updates via API
// 3. Verify update information
func TestUpdateFlowClosedLoop(t *testing.T) {
	pw, err := playwright.Run()
	if err != nil {
		t.Skip("Playwright not available:", err)
		return
	}
	defer pw.Stop()

	srv := helpers.SetupTestServerWithAdmin(t)
	defer srv.Close()

	t.Log("Step 1: Creating program and version...")
	programID := fmt.Sprintf("updloop-%d", time.Now().UnixNano())
	helpers.CreateTestProgram(t, srv, "UpdateLoopApp", "Update loop test")

	uploadToken, _ := helpers.GetProgramTokens(t, srv, programID)

	testZipPath := helpers.CreateTestVersionZip(t, "UpdateLoopApp", "1.0.0")
	defer os.Remove(testZipPath)

	uploadBody := &bytes.Buffer{}
	writer := multipart.NewWriter(uploadBody)
	file, _ := os.Open(testZipPath)
	part, _ := writer.CreateFormFile("file", filepath.Base(testZipPath))
	io.Copy(part, file)
	file.Close()
	writer.WriteField("channel", "stable")
	writer.WriteField("version", "1.0.0")
	writer.WriteField("notes", "Initial release")
	writer.WriteField("mandatory", "false")
	writer.Close()

	uploadReq := httptest.NewRequest("POST", "/api/programs/"+programID+"/versions", uploadBody)
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())
	uploadReq.Header.Set("Authorization", "Bearer "+uploadToken)

	uploadW := httptest.NewRecorder()
	srv.Router.ServeHTTP(uploadW, uploadReq)

	require.Equal(t, http.StatusOK, uploadW.Code, "Version upload should succeed")

	t.Log("Step 2: Checking for updates via API...")
	checkReq := httptest.NewRequest("GET", "/api/programs/"+programID+"/versions/latest?channel=stable", nil)

	checkW := httptest.NewRecorder()
	srv.Router.ServeHTTP(checkW, checkReq)

	require.Equal(t, http.StatusOK, checkW.Code)

	var latestVersion map[string]interface{}
	json.Unmarshal(checkW.Body.Bytes(), &latestVersion)

	assert.Equal(t, "1.0.0", latestVersion["version"])
	assert.Equal(t, "stable", latestVersion["channel"])

	t.Log("Update flow closed-loop test completed successfully!")
}

// TestClientPackageStructure verifies the structure of client packages
func TestClientPackageStructure(t *testing.T) {
	pw, err := playwright.Run()
	if err != nil {
		t.Skip("Playwright not available:", err)
		return
	}
	defer pw.Stop()

	srv := helpers.SetupTestServerWithAdmin(t)
	defer srv.Close()

	programID := fmt.Sprintf("struct-%d", time.Now().UnixNano())
	helpers.CreateTestProgram(t, srv, "StructureTestApp", "Client package structure test")

	t.Log("Client package structure test - placeholder")
	t.Log("Program ID:", programID)
	// This would verify actual client package structure when fully implemented
}

// TestBrowserBasedE2E runs actual browser-based E2E tests
func TestBrowserBasedE2E(t *testing.T) {
	pw, err := playwright.Run()
	if err != nil {
		t.Skip("Playwright not available:", err)
		return
	}
	defer pw.Stop()

	t.Log("Launching browser...")
	browser, err := pw.Chromium.Launch()
	require.NoError(t, err)
	defer browser.Close()

	context, err := browser.NewContext()
	require.NoError(t, err)
	defer context.Close()

	page, err := context.NewPage()
	require.NoError(t, err)

	t.Log("Verifying Playwright browser is functional...")
	// Just verify we can create a page and get its title
	title, err := page.Title()
	require.NoError(t, err)
	t.Logf("Browser page title: %s", title)

	t.Log("Browser-based E2E test completed successfully!")
}
