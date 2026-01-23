package helpers

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"

	"docufiller-update-server/internal/models"
)

// CreateTestProgram creates a test program in the database
func CreateTestProgram(t TestingT, srv *TestServer, name, description string) string {
	// Use a combination of timestamp and random bytes for uniqueness
	program := &models.Program{
		ProgramID:     fmt.Sprintf("prog-%d-%s", time.Now().UnixNano(), generateTestEncryptionKey()[:8]),
		Name:          name,
		Description:   description,
		EncryptionKey: generateTestEncryptionKey(),
		IsActive:      true,
	}

	if err := srv.DB.Create(program).Error; err != nil {
		t.Fatalf("Failed to create test program: %v", err)
	}

	return program.ProgramID
}

// CreateTestVersion creates a test version record in the database
func CreateTestVersion(t TestingT, db *gorm.DB, programID, channel, version string) *models.Version {
	v := &models.Version{
		ProgramID:     programID,
		Version:       version,
		Channel:       channel,
		FileName:      fmt.Sprintf("%s-%s-%s.zip", programID, channel, version),
		FilePath:      filepath.Join("./data/packages", programID, channel, version),
		FileSize:      1024000,
		FileHash:      "abc123def456",
		ReleaseNotes:  "Test release notes",
		PublishDate:   time.Now(),
		DownloadCount: 0,
		Mandatory:     false,
	}
	db.Create(v)
	return v
}

// CreateTestZip creates a test ZIP file for testing
func CreateTestZip(t TestingT, programName, version string) string {
	tempDir := os.TempDir()
	zipPath := filepath.Join(tempDir, fmt.Sprintf("%s-%s.zip", programName, version))

	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add a test file to the ZIP
	writer, err := zipWriter.Create(fmt.Sprintf("%s.exe", programName))
	if err != nil {
		t.Fatalf("Failed to create zip entry: %v", err)
	}

	// Write some test content
	testContent := []byte(fmt.Sprintf("Test application %s version %s", programName, version))
	_, err = writer.Write(testContent)
	if err != nil {
		t.Fatalf("Failed to write to zip: %v", err)
	}

	return zipPath
}

// CreateTestVersionZip creates a version ZIP file with proper content
func CreateTestVersionZip(t TestingT, programName, version string) string {
	return CreateTestZip(t, programName, version)
}

// getAdminToken performs login and returns the admin token
func getAdminToken(t TestingT, srv *TestServer) string {
	if srv.AdminToken != "" {
		return srv.AdminToken
	}

	// Perform login to get token (using test credentials from config)
	loginBody := map[string]interface{}{
		"username": "admin",
		"password": "test-password",
	}
	jsonBody, _ := json.Marshal(loginBody)

	req := httptest.NewRequest("POST", "/api/admin/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to login: status %d, body: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Login now returns {"success": true} instead of a token
	// The admin token is generated during server setup
	if srv.AdminToken == "" {
		t.Fatalf("Admin token not set in test server")
	}

	return srv.AdminToken
}

// GetProgramTokens returns the upload and download tokens for a program
func GetProgramTokens(t TestingT, srv *TestServer, programID string) (upload, download string) {
	uploadToken, err := srv.TokenService.GetToken(programID, "upload", "system")
	if err != nil {
		t.Fatalf("Failed to get upload token: %v", err)
	}

	downloadToken, err := srv.TokenService.GetToken(programID, "download", "system")
	if err != nil {
		t.Fatalf("Failed to get download token: %v", err)
	}

	return uploadToken.TokenValue, downloadToken.TokenValue
}

// createMultipartUploadRequest creates a multipart form upload request
func createMultipartUploadRequest(t TestingT, programID, channel, version, notes string, filePath string) (*http.Request, *bytes.Buffer) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	io.Copy(part, file)

	// Add other fields
	writer.WriteField("channel", channel)
	writer.WriteField("version", version)
	writer.WriteField("notes", notes)
	writer.WriteField("mandatory", "false")
	writer.Close()

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/programs/%s/versions", programID), body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, body
}

// generateTestEncryptionKey generates a test encryption key
func generateTestEncryptionKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// cleanupTestFile removes a test file
func cleanupTestFile(path string) {
	if path != "" {
		os.Remove(path)
	}
}
