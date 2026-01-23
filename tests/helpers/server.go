package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"docufiller-update-server/internal/config"
	"docufiller-update-server/internal/handler"
	"docufiller-update-server/internal/logger"
	"docufiller-update-server/internal/middleware"
	"docufiller-update-server/internal/models"
	"docufiller-update-server/internal/service"
)

// TestServer represents a test server instance
type TestServer struct {
	Router            *gin.Engine
	DB                *gorm.DB
	URL               string
	TempDir           string
	TokenService      *service.TokenService
	ProgramService    *service.ProgramService
	VersionService    *service.VersionService
	StorageBasePath   string
	AdminToken        string
	TestProgramID     string
	TestUploadToken   string
	TestDownloadToken string
}

// setupTestServer creates a basic test server
func setupTestServer(t TestingT) *TestServer {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "update-server-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Setup test database with busy timeout for Windows
	dbPath := filepath.Join(tempDir, "test.db")
	dsn := "file:" + dbPath + "?_busy_timeout=30000&_journal_mode=WAL&_foreign_keys=1&_synchronous=NORMAL"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Configure SQLite to use single connection to avoid locking
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	// Auto migrate all models
	err = db.AutoMigrate(
		&models.Version{},
		&models.Program{},
		&models.Token{},
		&models.EncryptionKey{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize logger (suppress output)
	loggerCfg := logger.Config{
		Level:  "error",
		Output: "stdout",
	}
	_ = logger.Init(loggerCfg)

	// Setup test config
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 18080,
			Host: "127.0.0.1",
		},
		Crypto: config.CryptoConfig{
			MasterKey: "test-master-key-for-testing-32bytes!!",
		},
		Admin: config.AdminConfig{
			Username: "admin",
			Password: "test-password",
		},
	}

	// Setup Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add Session middleware
	store := cookie.NewStore([]byte(cfg.Crypto.MasterKey))
	router.Use(sessions.Sessions("admin-session", store))

	// Initialize services
	tokenSvc := service.NewTokenService(db)
	authMiddleware := middleware.NewAuthMiddleware(tokenSvc)

	// Initialize crypto service
	cryptoSvc := service.NewCryptoService(cfg.Crypto.MasterKey)
	cryptoMiddleware := middleware.NewCryptoMiddleware(cryptoSvc)

	// Storage base path
	storageBasePath := filepath.Join(tempDir, "packages")

	// Initialize services
	storageService := service.NewStorageService(storageBasePath)
	programService := service.NewProgramService(db)
	versionService := service.NewVersionService(db, storageService)
	clientPackagerService := service.NewClientPackager(programService)

	// Get a test server port
	serverURL := "http://127.0.0.1:18080"

	// Create test server instance
	ts := &TestServer{
		Router:          router,
		DB:              db,
		URL:             serverURL,
		TempDir:         tempDir,
		TokenService:    tokenSvc,
		ProgramService:  programService,
		VersionService:  versionService,
		StorageBasePath: storageBasePath,
	}

	// Setup routes
	setupTestRoutes(router, db, cfg, tokenSvc, authMiddleware, cryptoMiddleware, programService, versionService, clientPackagerService, storageBasePath)

	return ts
}

// SetupTestServer creates a test server with in-memory database
func SetupTestServer(t TestingT) *TestServer {
	return setupTestServer(t)
}

// setupTestServerWithAdmin creates a test server with an admin token
func setupTestServerWithAdmin(t TestingT) *TestServer {
	srv := setupTestServer(t)

	// Generate admin token (for API authentication)
	_, adminToken, _ := srv.TokenService.GenerateToken("", "admin", "system")

	srv.AdminToken = adminToken

	return srv
}

// SetupTestServerWithAdmin creates a test server with an admin user
func SetupTestServerWithAdmin(t TestingT) *TestServer {
	return setupTestServerWithAdmin(t)
}

// setupTestServerWithProgram creates a test server with admin token and a test program
func setupTestServerWithProgram(t TestingT) *TestServer {
	srv := setupTestServerWithAdmin(t)

	// Create test program
	programName := fmt.Sprintf("test-app-%d", time.Now().Unix())
	program := &models.Program{
		ProgramID:     fmt.Sprintf("prog-%d", time.Now().Unix()),
		Name:          programName,
		Description:   "Test application",
		EncryptionKey: "test-encryption-key-32-bytes-long!",
		IsActive:      true,
	}

	if err := srv.DB.Create(program).Error; err != nil {
		t.Fatalf("Failed to create test program: %v", err)
	}

	// Generate tokens for the program
	_, uploadTokenValue, _ := srv.TokenService.GenerateToken(program.ProgramID, "upload", "system")
	_, downloadTokenValue, _ := srv.TokenService.GenerateToken(program.ProgramID, "download", "system")

	srv.TestProgramID = program.ProgramID
	srv.TestUploadToken = uploadTokenValue
	srv.TestDownloadToken = downloadTokenValue

	return srv
}

// SetupTestServerWithProgram creates a test server with admin user and a test program
func SetupTestServerWithProgram(t TestingT) *TestServer {
	return setupTestServerWithProgram(t)
}

func setupTestRoutes(
	r *gin.Engine,
	db *gorm.DB,
	cfg *config.Config,
	tokenSvc *service.TokenService,
	authMiddleware *middleware.AuthMiddleware,
	cryptoMiddleware *middleware.CryptoMiddleware,
	programService *service.ProgramService,
	versionService *service.VersionService,
	clientPackagerService *service.ClientPackager,
	storageBasePath string,
) {
	// Register crypto middleware
	r.Use(cryptoMiddleware.Process())

	// Initialize handlers
	authHandler := handler.NewAuthHandler(cfg)

	adminHandler := handler.NewAdminHandler(
		programService,
		versionService,
		tokenSvc,
		clientPackagerService,
	)

	versionHandler := handler.NewVersionHandler(db)

	// Admin API routes
	adminAPI := r.Group("/api/admin")
	{
		adminAPI.POST("/login", authHandler.Login)
		adminAPI.GET("/stats", adminHandler.GetStats)
		adminAPI.GET("/programs", adminHandler.ListPrograms)
		adminAPI.POST("/programs", adminHandler.CreateProgram)
		adminAPI.GET("/programs/:programId", adminHandler.GetProgramDetail)
		adminAPI.DELETE("/programs/:programId", adminHandler.DeleteProgram)
		adminAPI.GET("/programs/:programId/versions", adminHandler.ListVersions)
		adminAPI.DELETE("/programs/:programId/versions/:version", adminHandler.DeleteVersion)
		adminAPI.GET("/programs/:programId/client/publish", adminHandler.DownloadPublishClient)
		adminAPI.GET("/programs/:programId/client/update", adminHandler.DownloadUpdateClient)
	}

	// Public API routes
	public := r.Group("/api")
	{
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
		public.GET("/programs/:programId/versions/latest", versionHandler.GetLatestVersion)
		public.GET("/programs/:programId/versions", versionHandler.GetVersionList)
		public.GET("/programs/:programId/versions/:channel/:version", versionHandler.GetVersionDetail)
	}

	// Authenticated upload routes
	upload := r.Group("/api")
	upload.Use(authMiddleware.RequireUpload())
	{
		upload.POST("/programs/:programId/versions", versionHandler.UploadVersion)
		upload.DELETE("/programs/:programId/versions/:channel/:version", versionHandler.DeleteVersion)
	}

	// Authenticated download routes
	download := r.Group("/api")
	download.Use(authMiddleware.RequireDownload())
	{
		download.GET("/programs/:programId/download/:channel/:version", versionHandler.DownloadFile)
	}
}

// Close cleans up test resources
func (srv *TestServer) Close() error {
	// Close database connection
	if srv.DB != nil {
		sqlDB, _ := srv.DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	// Remove temp directory
	if srv.TempDir != "" {
		return os.RemoveAll(srv.TempDir)
	}
	return nil
}

// TestingT is an interface that *testing.T implements
type TestingT interface {
	Helper()
	Fatalf(format string, args ...interface{})
}
