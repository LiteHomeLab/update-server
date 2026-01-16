package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewUpdateAdmin(t *testing.T) {
	admin := NewUpdateAdmin("http://localhost:8080", "test-token")

	if admin.serverURL != "http://localhost:8080" {
		t.Errorf("Expected serverURL http://localhost:8080, got %s", admin.serverURL)
	}

	if admin.token != "test-token" {
		t.Errorf("Expected token test-token, got %s", admin.token)
	}

	if admin.client == nil {
		t.Error("Expected client to be initialized")
	}

	if admin.client.Timeout != 300*time.Second {
		t.Errorf("Expected timeout 300s, got %v", admin.client.Timeout)
	}
}

func TestListVersions(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "/api/programs/test-program/versions") {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		if r.URL.Query().Get("channel") != "stable" {
			t.Errorf("Expected channel stable, got %s", r.URL.Query().Get("channel"))
		}

		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Expected Authorization header 'Bearer test-token', got %s", auth)
		}

		// Send mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"version": "1.0.0",
				"channel": "stable",
				"fileName": "app-1.0.0.zip",
				"fileSize": 1024000,
				"fileHash": "abc123",
				"releaseNotes": "Initial release",
				"publishDate": "2024-01-01T00:00:00Z",
				"mandatory": false
			},
			{
				"version": "1.0.1",
				"channel": "stable",
				"fileName": "app-1.0.1.zip",
				"fileSize": 1025000,
				"fileHash": "def456",
				"releaseNotes": "Bug fixes",
				"publishDate": "2024-01-02T00:00:00Z",
				"mandatory": true
			}
		]`))
	}))
	defer server.Close()

	// Create admin client
	admin := NewUpdateAdmin(server.URL, "test-token")

	// Test ListVersions
	versions, err := admin.ListVersions("test-program", "stable")
	if err != nil {
		t.Fatalf("ListVersions failed: %v", err)
	}

	if len(versions) != 2 {
		t.Fatalf("Expected 2 versions, got %d", len(versions))
	}

	// Verify first version
	if versions[0].Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", versions[0].Version)
	}

	if versions[0].Channel != "stable" {
		t.Errorf("Expected channel stable, got %s", versions[0].Channel)
	}

	if versions[0].FileSize != 1024000 {
		t.Errorf("Expected file size 1024000, got %d", versions[0].FileSize)
	}

	if versions[0].Mandatory {
		t.Error("Expected mandatory false, got true")
	}

	// Verify second version
	if versions[1].Version != "1.0.1" {
		t.Errorf("Expected version 1.0.1, got %s", versions[1].Version)
	}

	if !versions[1].Mandatory {
		t.Error("Expected mandatory true, got false")
	}
}

func TestListVersions_Error(t *testing.T) {
	// Mock server returning error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	admin := NewUpdateAdmin(server.URL, "test-token")

	_, err := admin.ListVersions("test-program", "stable")
	if err == nil {
		t.Error("Expected error for failed request, got nil")
	}

	if !strings.Contains(err.Error(), "status 500") {
		t.Errorf("Expected error message to contain 'status 500', got %v", err)
	}
}

func TestDeleteVersion(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}

		expectedPath := "/api/programs/test-program/versions/1.0.0"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Expected Authorization header 'Bearer test-token', got %s", auth)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	admin := NewUpdateAdmin(server.URL, "test-token")

	err := admin.DeleteVersion("test-program", "1.0.0")
	if err != nil {
		t.Fatalf("DeleteVersion failed: %v", err)
	}
}

func TestDeleteVersion_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	admin := NewUpdateAdmin(server.URL, "test-token")

	err := admin.DeleteVersion("test-program", "1.0.0")
	if err == nil {
		t.Error("Expected error for failed delete, got nil")
	}

	if !strings.Contains(err.Error(), "status 404") {
		t.Errorf("Expected error message to contain 'status 404', got %v", err)
	}
}

func TestUploadVersion(t *testing.T) {
	// Create temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-upload.zip")
	content := strings.Repeat("test content", 1000) // ~12KB
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Track progress
	var progressCalls []UploadProgress

	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		expectedPath := "/api/programs/test-program/versions"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Expected Authorization header 'Bearer test-token', got %s", auth)
		}

		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "multipart/form-data") {
			t.Errorf("Expected multipart/form-data content type, got %s", contentType)
		}

		// Parse form
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("Failed to parse multipart form: %v", err)
		}

		// Verify form fields
		if r.FormValue("channel") != "beta" {
			t.Errorf("Expected channel beta, got %s", r.FormValue("channel"))
		}

		if r.FormValue("version") != "2.0.0" {
			t.Errorf("Expected version 2.0.0, got %s", r.FormValue("version"))
		}

		if r.FormValue("notes") != "Test release notes" {
			t.Errorf("Expected notes 'Test release notes', got %s", r.FormValue("notes"))
		}

		if r.FormValue("mandatory") != "true" {
			t.Errorf("Expected mandatory true, got %s", r.FormValue("mandatory"))
		}

		// Verify file
		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("Failed to get form file: %v", err)
		}
		defer file.Close()

		if header.Filename != "test-upload.zip" {
			t.Errorf("Expected filename test-upload.zip, got %s", header.Filename)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	admin := NewUpdateAdmin(server.URL, "test-token")

	// Test upload with progress callback
	callback := func(p UploadProgress) {
		progressCalls = append(progressCalls, p)
	}

	err := admin.UploadVersion("test-program", "beta", "2.0.0", testFile, "Test release notes", true, callback)
	if err != nil {
		t.Fatalf("UploadVersion failed: %v", err)
	}

	// Verify progress was tracked
	if len(progressCalls) == 0 {
		t.Error("Expected progress callbacks, got none")
	}

	// Verify final progress
	lastProgress := progressCalls[len(progressCalls)-1]
	if lastProgress.Uploaded != lastProgress.Total {
		t.Errorf("Expected uploaded %d, got %d", lastProgress.Total, lastProgress.Uploaded)
	}

	if lastProgress.Percentage != 100.0 {
		t.Errorf("Expected 100%% progress, got %f", lastProgress.Percentage)
	}
}

func TestVerifyUpload(t *testing.T) {
	expectedHash := "abc123def456"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		expectedPath := "/api/programs/test-program/versions/stable/1.0.0"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Expected Authorization header 'Bearer test-token', got %s", auth)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"version": "1.0.0",
			"channel": "stable",
			"fileName": "app-1.0.0.zip",
			"fileSize": 1024000,
			"fileHash": "` + expectedHash + `",
			"releaseNotes": "Test release",
			"publishDate": "2024-01-01T00:00:00Z",
			"mandatory": false
		}`))
	}))
	defer server.Close()

	admin := NewUpdateAdmin(server.URL, "test-token")

	err := admin.VerifyUpload("test-program", "stable", "1.0.0", expectedHash)
	if err != nil {
		t.Fatalf("VerifyUpload failed: %v", err)
	}
}

func TestVerifyUpload_HashMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"version": "1.0.0",
			"channel": "stable",
			"fileName": "app-1.0.0.zip",
			"fileSize": 1024000,
			"fileHash": "server-hash",
			"releaseNotes": "Test release",
			"publishDate": "2024-01-01T00:00:00Z",
			"mandatory": false
		}`))
	}))
	defer server.Close()

	admin := NewUpdateAdmin(server.URL, "test-token")

	err := admin.VerifyUpload("test-program", "stable", "1.0.0", "client-hash")
	if err == nil {
		t.Error("Expected hash mismatch error, got nil")
	}

	if !strings.Contains(err.Error(), "SHA256 mismatch") {
		t.Errorf("Expected SHA256 mismatch error, got %v", err)
	}
}

func TestUploadVersionWithVerify(t *testing.T) {
	// Create temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-verify.zip")
	content := strings.Repeat("verify test", 100)
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Calculate expected hash
	expectedHash, _ := CalculateSHA256(testFile)

	uploadCalled := false
	verifyCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			// Upload endpoint
			uploadCalled = true
			w.WriteHeader(http.StatusOK)
		} else if r.Method == "GET" {
			// Verify endpoint
			verifyCalled = true
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"version": "1.5.0",
				"channel": "beta",
				"fileName": "test-verify.zip",
				"fileSize": 1200,
				"fileHash": "` + expectedHash + `",
				"releaseNotes": "Test",
				"publishDate": "2024-01-01T00:00:00Z",
				"mandatory": false
			}`))
		}
	}))
	defer server.Close()

	admin := NewUpdateAdmin(server.URL, "test-token")

	err := admin.UploadVersionWithVerify("test-program", "beta", "1.5.0", testFile, "Test", false, nil)
	if err != nil {
		t.Fatalf("UploadVersionWithVerify failed: %v", err)
	}

	if !uploadCalled {
		t.Error("Upload endpoint was not called")
	}

	if !verifyCalled {
		t.Error("Verify endpoint was not called")
	}
}

func TestUploadVersion_FileNotFound(t *testing.T) {
	admin := NewUpdateAdmin("http://localhost:8080", "test-token")

	err := admin.UploadVersion("test-program", "stable", "1.0.0", "nonexistent.zip", "Test", false, nil)
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}

	if !strings.Contains(err.Error(), "failed to open file") {
		t.Errorf("Expected file open error, got %v", err)
	}
}

func TestProgressWriter(t *testing.T) {
	var progressCalls []UploadProgress
	callback := func(p UploadProgress) {
		progressCalls = append(progressCalls, p)
	}

	var buf strings.Builder
	pw := &progressWriter{
		writer:   &buf,
		total:    100,
		callback: callback,
	}

	// Write in multiple chunks
	pw.Write([]byte("test1"))
	pw.Write([]byte("test2"))
	pw.Write([]byte("test3"))

	if len(progressCalls) != 3 {
		t.Errorf("Expected 3 progress calls, got %d", len(progressCalls))
	}

	// Verify progress tracking
	if progressCalls[0].Uploaded != 5 {
		t.Errorf("Expected uploaded 5, got %d", progressCalls[0].Uploaded)
	}

	if progressCalls[1].Uploaded != 10 {
		t.Errorf("Expected uploaded 10, got %d", progressCalls[1].Uploaded)
	}

	if progressCalls[2].Uploaded != 15 {
		t.Errorf("Expected uploaded 15, got %d", progressCalls[2].Uploaded)
	}

	// Verify percentage
	expectedPercent := float64(15) / float64(100) * 100
	if progressCalls[2].Percentage != expectedPercent {
		t.Errorf("Expected percentage %f, got %f", expectedPercent, progressCalls[2].Percentage)
	}
}
