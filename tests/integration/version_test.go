package integration

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

	"github.com/stretchr/testify/assert"

	"docufiller-update-server/tests/helpers"
)

// TestUploadVersion tests uploading a new version
func TestUploadVersion(t *testing.T) {
	srv := helpers.SetupTestServerWithAdmin(t)
	defer srv.Close()

	// Create a test program
	programID := helpers.CreateTestProgram(t, srv, "UploadTestApp", "For upload testing")

	// Get upload token
	uploadToken, _ := helpers.GetProgramTokens(t, srv, programID)

	// Create a test ZIP file
	zipPath := helpers.CreateTestVersionZip(t, "UploadTestApp", "1.0.0")
	defer os.Remove(zipPath)

	// Create multipart form request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	file, _ := os.Open(zipPath)
	defer file.Close()

	part, _ := writer.CreateFormFile("file", filepath.Base(zipPath))
	_, _ = io.Copy(part, file)

	// Add form fields
	_ = writer.WriteField("channel", "stable")
	_ = writer.WriteField("version", "1.0.0")
	_ = writer.WriteField("notes", "First release")
	_ = writer.WriteField("mandatory", "false")
	writer.Close()

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/programs/%s/versions", programID), body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+uploadToken)

	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Response structure: {message: "...", version: {...}}
	versionObj, ok := response["version"].(map[string]interface{})
	assert.True(t, ok, "version should be in response")

	assert.Equal(t, "stable", versionObj["channel"])
	assert.Equal(t, "1.0.0", versionObj["version"])
	assert.NotEmpty(t, versionObj["fileName"])
}

// TestListVersions tests listing versions for a program
func TestListVersions(t *testing.T) {
	srv := helpers.SetupTestServerWithAdmin(t)
	defer srv.Close()

	programID := helpers.CreateTestProgram(t, srv, "ListTestApp", "For list testing")

	// Create some test versions directly in DB
	helpers.CreateTestVersion(t, srv.DB, programID, "stable", "1.0.0")
	helpers.CreateTestVersion(t, srv.DB, programID, "stable", "1.1.0")
	helpers.CreateTestVersion(t, srv.DB, programID, "beta", "2.0.0")

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/admin/programs/%s/versions", programID), nil)
	req.Header.Set("Authorization", "Bearer "+srv.AdminToken)

	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.GreaterOrEqual(t, len(response), 3)
}

// TestGetLatestVersion tests getting the latest version
func TestGetLatestVersion(t *testing.T) {
	srv := helpers.SetupTestServerWithAdmin(t)
	defer srv.Close()

	programID := helpers.CreateTestProgram(t, srv, "LatestTestApp", "For latest testing")

	// Create test versions
	helpers.CreateTestVersion(t, srv.DB, programID, "stable", "1.0.0")
	helpers.CreateTestVersion(t, srv.DB, programID, "stable", "1.1.0")

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/programs/%s/versions/latest?channel=stable", programID), nil)

	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "stable", response["channel"])
}

// TestDeleteVersion tests deleting a version
func TestDeleteVersion(t *testing.T) {
	srv := helpers.SetupTestServerWithAdmin(t)
	defer srv.Close()

	programID := helpers.CreateTestProgram(t, srv, "DeleteTestApp", "For delete testing")

	// Create a test version
	_ = helpers.CreateTestVersion(t, srv.DB, programID, "stable", "1.0.0")

	// Delete the version
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/admin/programs/%s/versions/%s", programID, "1.0.0"), nil)
	req.Header.Set("Authorization", "Bearer "+srv.AdminToken)

	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestUploadValidation tests validation for upload requests
func TestUploadValidation(t *testing.T) {
	srv := helpers.SetupTestServerWithAdmin(t)
	defer srv.Close()

	programID := helpers.CreateTestProgram(t, srv, "ValidationApp", "For validation testing")
	uploadToken, _ := helpers.GetProgramTokens(t, srv, programID)

	tests := []struct {
		name       string
		token      string
		channel    string
		version    string
		wantStatus int
	}{
		{
			name:       "missing file",
			token:      uploadToken,
			channel:    "stable",
			version:    "1.0.0",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid token",
			token:      "invalid-token",
			channel:    "stable",
			version:    "1.0.0",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			_ = writer.WriteField("channel", tt.channel)
			_ = writer.WriteField("version", tt.version)
			writer.Close()

			req := httptest.NewRequest("POST", fmt.Sprintf("/api/programs/%s/versions", programID), body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req.Header.Set("Authorization", "Bearer "+tt.token)

			w := httptest.NewRecorder()
			srv.Router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// TestGetVersionDetail tests getting a specific version's details
func TestGetVersionDetail(t *testing.T) {
	srv := helpers.SetupTestServerWithAdmin(t)
	defer srv.Close()

	programID := helpers.CreateTestProgram(t, srv, "DetailVersionApp", "For detail testing")

	// Create a test version
	helpers.CreateTestVersion(t, srv.DB, programID, "stable", "1.0.0")

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/programs/%s/versions/stable/1.0.0", programID), nil)

	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "stable", response["channel"])
	assert.Equal(t, "1.0.0", response["version"])
}

// TestDownloadTokenAuth tests download token authentication
func TestDownloadTokenAuth(t *testing.T) {
	srv := helpers.SetupTestServerWithAdmin(t)
	defer srv.Close()

	programID := helpers.CreateTestProgram(t, srv, "DownloadTestApp", "For download testing")

	// Create a test version
	helpers.CreateTestVersion(t, srv.DB, programID, "stable", "1.0.0")

	// Try to download without token
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/programs/%s/download/stable/1.0.0", programID), nil)

	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	// Should fail without authentication
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
