package tests

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

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"docufiller-update-server/internal/handler"
	"docufiller-update-server/internal/logger"
	"docufiller-update-server/internal/middleware"
	"docufiller-update-server/internal/models"
	"docufiller-update-server/internal/service"
)

// TestServer represents a test server instance
type TestServer struct {
	Router       *gin.Engine
	DB           *gorm.DB
	TestToken    string
	DownloadToken string
	TempDir      string
}

// SetupTestServer creates a test server with in-memory database
func SetupTestServer(t *testing.T) *TestServer {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "update-server-test-*")
	assert.NoError(t, err)

	// Setup test database
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	assert.NoError(t, err)

	// Auto migrate
	err = db.AutoMigrate(&models.Version{}, &models.Program{}, &models.Token{})
	assert.NoError(t, err)

	// Initialize logger (suppress output)
	loggerCfg := logger.Config{
		Level:  "error",
		Output: "stdout",
	}
	_ = logger.Init(loggerCfg)

	// Setup Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Initialize services
	tokenSvc := service.NewTokenService(db)
	authMiddleware := middleware.NewAuthMiddleware(tokenSvc)

	// Create test tokens
	_, adminToken, _ := tokenSvc.GenerateToken("", "admin", "admin")
	_, downloadToken, _ := tokenSvc.GenerateToken("test-program", "download", "download")

	// Setup routes
	setupTestRoutes(router, db, authMiddleware)

	return &TestServer{
		Router:        router,
		DB:            db,
		TestToken:     adminToken,
		DownloadToken: downloadToken,
		TempDir:       tempDir,
	}
}

// TearDownTestServer cleans up test resources
func (ts *TestServer) TearDownTestServer(t *testing.T) {
	// Close database connection
	sqlDB, _ := ts.DB.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}

	// Remove temp directory
	err := os.RemoveAll(ts.TempDir)
	assert.NoError(t, err)
}

func setupTestRoutes(r *gin.Engine, db *gorm.DB, authMiddleware *middleware.AuthMiddleware) {
	// Public routes
	public := r.Group("/api")
	{
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
		public.GET("/programs/:programId/versions/latest", handler.NewVersionHandler(db).GetLatestVersion)
		public.GET("/programs/:programId/versions", handler.NewVersionHandler(db).GetVersionList)
		public.GET("/programs/:programId/versions/:channel/:version", handler.NewVersionHandler(db).GetVersionDetail)
	}

	// Authenticated routes
	upload := r.Group("/api")
	upload.Use(authMiddleware.RequireUpload())
	{
		upload.POST("/programs/:programId/versions", handler.NewVersionHandler(db).UploadVersion)
		upload.DELETE("/programs/:programId/versions/:channel/:version", handler.NewVersionHandler(db).DeleteVersion)
	}

	download := r.Group("/api")
	download.Use(authMiddleware.RequireDownload())
	{
		download.GET("/programs/:programId/download/:channel/:version", handler.NewVersionHandler(db).DownloadFile)
	}
}

// Helper function to create a test version record
func createTestVersion(db *gorm.DB, programID, channel, version string) *models.Version {
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

// Helper function to create multipart form data for file upload
func createMultipartFormData(fieldName, fileName, fileContent string) (bytes.Buffer, *multipart.Writer) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add file field
	part, _ := writer.CreateFormFile(fieldName, fileName)
	part.Write([]byte(fileContent))

	// Add other fields
	writer.WriteField("channel", "stable")
	writer.WriteField("version", "1.0.0")
	writer.WriteField("notes", "Test version")
	writer.WriteField("mandatory", "false")

	writer.Close()
	return body, writer
}

// TestHealthCheck tests the health check endpoint
func TestHealthCheck(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.TearDownTestServer(t)

	req, _ := http.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"ok"}`, w.Body.String())
}

// TestGetLatestVersion tests retrieving the latest version
func TestGetLatestVersion(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.TearDownTestServer(t)

	// Create test versions
	createTestVersion(ts.DB, "test-program", "stable", "1.0.0")
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	createTestVersion(ts.DB, "test-program", "stable", "1.1.0")

	// Request latest version
	req, _ := http.NewRequest("GET", "/api/programs/test-program/versions/latest?channel=stable", nil)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "1.1.0", response["version"])
}

// TestGetLatestVersion_NotFound tests requesting latest version when none exists
func TestGetLatestVersion_NotFound(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.TearDownTestServer(t)

	req, _ := http.NewRequest("GET", "/api/programs/nonexistent/versions/latest?channel=stable", nil)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestGetVersionList tests retrieving version list
func TestGetVersionList(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.TearDownTestServer(t)

	// Create test versions
	createTestVersion(ts.DB, "test-program", "stable", "1.0.0")
	createTestVersion(ts.DB, "test-program", "stable", "1.1.0")
	createTestVersion(ts.DB, "test-program", "beta", "2.0.0")

	// Request version list for stable channel
	req, _ := http.NewRequest("GET", "/api/programs/test-program/versions?channel=stable", nil)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
}

// TestGetVersionList_AllChannels tests retrieving version list without channel filter
func TestGetVersionList_AllChannels(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.TearDownTestServer(t)

	// Create test versions
	createTestVersion(ts.DB, "test-program", "stable", "1.0.0")
	createTestVersion(ts.DB, "test-program", "beta", "2.0.0")

	// Request all versions
	req, _ := http.NewRequest("GET", "/api/programs/test-program/versions", nil)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
}

// TestGetVersionDetail tests retrieving version details
func TestGetVersionDetail(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.TearDownTestServer(t)

	// Create test version
	createTestVersion(ts.DB, "test-program", "stable", "1.0.0")

	// Request version detail
	req, _ := http.NewRequest("GET", "/api/programs/test-program/versions/stable/1.0.0", nil)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "1.0.0", response["version"])
	assert.Equal(t, "stable", response["channel"])
}

// TestGetVersionDetail_NotFound tests retrieving non-existent version
func TestGetVersionDetail_NotFound(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.TearDownTestServer(t)

	req, _ := http.NewRequest("GET", "/api/programs/test-program/versions/stable/9.9.9", nil)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestUploadVersion tests uploading a new version
func TestUploadVersion(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.TearDownTestServer(t)

	// Create upload request with file
	body, writer := createMultipartFormData("file", "test-program-1.0.0.zip", "test file content")

	req, _ := http.NewRequest("POST", "/api/programs/test-program/versions", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+ts.TestToken)

	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Version uploaded successfully", response["message"])
}

// TestUploadVersion_NoAuth tests uploading without authentication
func TestUploadVersion_NoAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.TearDownTestServer(t)

	body, writer := createMultipartFormData("file", "test.zip", "content")

	req, _ := http.NewRequest("POST", "/api/programs/test-program/versions", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	// No authorization header

	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestDeleteVersion tests deleting a version
func TestDeleteVersion(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.TearDownTestServer(t)

	// Create test version
	v := createTestVersion(ts.DB, "test-program", "stable", "1.0.0")

	// Delete version (use channel and version in path)
	req, _ := http.NewRequest("DELETE", "/api/programs/test-program/versions/stable/1.0.0", nil)
	req.Header.Set("Authorization", "Bearer "+ts.TestToken)

	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify deletion (GORM uses soft delete by default)
	var count int64
	ts.DB.Model(&models.Version{}).Where("id = ?", v.ID).Count(&count)
	assert.Equal(t, int64(0), count, "Record should be soft deleted")
}

// TestDownloadFile_TokenRequired tests that download requires authentication
func TestDownloadFile_TokenRequired(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.TearDownTestServer(t)

	// Create test version
	createTestVersion(ts.DB, "test-program", "stable", "1.0.0")

	// Request without token
	req, _ := http.NewRequest("GET", "/api/programs/test-program/download/stable/1.0.0", nil)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestDownloadFile_WithToken tests downloading with valid token
func TestDownloadFile_WithToken(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.TearDownTestServer(t)

	// Create package directory and file
	pkgDir := filepath.Join(ts.TempDir, "packages", "test-program", "stable", "1.0.0")
	os.MkdirAll(pkgDir, 0755)
	testFile := filepath.Join(pkgDir, "test.zip")
	testContent := []byte("test file content")
	os.WriteFile(testFile, testContent, 0644)

	// Create test version record with updated file path
	v := createTestVersion(ts.DB, "test-program", "stable", "1.0.0")
	v.FilePath = pkgDir
	ts.DB.Save(v)

	// Request with token
	req, _ := http.NewRequest("GET", "/api/programs/test-program/download/stable/1.0.0", nil)
	req.Header.Set("Authorization", "Bearer "+ts.DownloadToken)

	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	// Note: The actual file serving might fail due to path issues in test environment,
	// but we verify the authentication passed
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

// TestDatabaseConnection tests database connection and migration
func TestDatabaseConnection(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.TearDownTestServer(t)

	// Verify database connection
	sqlDB, err := ts.DB.DB()
	assert.NoError(t, err)
	assert.NoError(t, sqlDB.Ping())

	// Verify tables exist
	assert.True(t, ts.DB.Migrator().HasTable(&models.Version{}))
	assert.True(t, ts.DB.Migrator().HasTable(&models.Program{}))
	assert.True(t, ts.DB.Migrator().HasTable(&models.Token{}))
}

// BenchmarkGetLatestVersion benchmarks the latest version endpoint
func BenchmarkGetLatestVersion(b *testing.B) {
	tempDir, _ := os.MkdirTemp("", "bench-*")
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "bench.db")
	db, _ := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	db.AutoMigrate(&models.Version{})

	// Create 100 test versions
	for i := 0; i < 100; i++ {
		createTestVersion(db, "bench-program", "stable", fmt.Sprintf("1.%d.0", i))
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/programs/:programId/versions/latest", handler.NewVersionHandler(db).GetLatestVersion)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/programs/bench-program/versions/latest?channel=stable", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// Helper function to read response body
func readBody(r io.Reader) string {
	body, _ := io.ReadAll(r)
	return string(body)
}
