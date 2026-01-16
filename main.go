package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"docufiller-update-server/internal/config"
	"docufiller-update-server/internal/database"
	"docufiller-update-server/internal/handler"
	"docufiller-update-server/internal/logger"
	"docufiller-update-server/internal/middleware"
	"docufiller-update-server/internal/models"
	"docufiller-update-server/internal/service"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		panic(err)
	}

	// 初始化日志
	loggerCfg := logger.Config{
		Level:      cfg.Logger.Level,
		Output:     cfg.Logger.Output,
		FilePath:   cfg.Logger.FilePath,
		MaxSize:    cfg.Logger.MaxSize,
		MaxBackups: cfg.Logger.MaxBackups,
		MaxAge:     cfg.Logger.MaxAge,
		Compress:   cfg.Logger.Compress,
	}
	if err := logger.Init(loggerCfg); err != nil {
		panic(err)
	}
	logger.Info("Starting DocuFiller Update Server...")

	// 初始化数据库
	db, err := database.NewGORM(cfg.Database.Path)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(&models.Version{}, &models.Program{}, &models.Token{}); err != nil {
		logger.Fatalf("Failed to migrate database: %v", err)
	}

	// 初始化认证中间件
	tokenSvc := service.NewTokenService(db)
	authMiddleware := middleware.NewAuthMiddleware(tokenSvc)

	// 初始化加密服务
	cryptoSvc := service.NewCryptoService(cfg.Crypto.MasterKey)
	cryptoMiddleware := middleware.NewCryptoMiddleware(cryptoSvc)

	// 设置 Gin
	if cfg.Logger.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// 注册加密中间件
	r.Use(cryptoMiddleware.Process())

	// 注册路由
	setupRoutes(r, db, authMiddleware)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Infof("Server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes(r *gin.Engine, db *gorm.DB, authMiddleware *middleware.AuthMiddleware) {
	// 公开路由
	public := r.Group("/api")
	{
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
		public.GET("/programs/:programId/versions/latest", handler.NewVersionHandler(db).GetLatestVersion)
		public.GET("/programs/:programId/versions", handler.NewVersionHandler(db).GetVersionList)
		public.GET("/programs/:programId/versions/:channel/:version", handler.NewVersionHandler(db).GetVersionDetail)
	}

	// 认证路由 - 下载
	download := r.Group("/api")
	download.Use(authMiddleware.RequireDownload())
	{
		download.GET("/programs/:programId/download/:channel/:version", handler.NewVersionHandler(db).DownloadFile)
	}

	// 认证路由 - 上传
	upload := r.Group("/api")
	upload.Use(authMiddleware.RequireUpload())
	{
		upload.POST("/programs/:programId/versions", handler.NewVersionHandler(db).UploadVersion)
		upload.DELETE("/programs/:programId/versions/:version", handler.NewVersionHandler(db).DeleteVersion)

		// 程序管理路由
		programHandler := handler.NewProgramHandler(service.NewProgramService(db))
		upload.POST("/programs", programHandler.CreateProgram)
		upload.GET("/programs", programHandler.ListPrograms)
		upload.GET("/programs/:programId", programHandler.GetProgram)
	}

	// 向后兼容路由 - 映射到 docufiller
	// 保留旧 API 端点以确保向后兼容性
	deprecationMiddleware := middleware.DeprecationWarning()

	legacy := r.Group("/api/version")
	legacy.Use(deprecationMiddleware)
	{
		// GET /api/version/latest?channel=stable -> /api/programs/docufiller/versions/latest?channel=stable
		legacy.GET("/latest", func(c *gin.Context) {
			c.Request.URL.Path = "/api/programs/docufiller/versions/latest"
			r.HandleContext(c)
		})

		// GET /api/version/list?channel=stable -> /api/programs/docufiller/versions?channel=stable
		legacy.GET("/list", func(c *gin.Context) {
			c.Request.URL.Path = "/api/programs/docufiller/versions"
			r.HandleContext(c)
		})

		// POST /api/version/upload -> /api/programs/docufiller/versions
		legacy.POST("/upload", func(c *gin.Context) {
			c.Request.URL.Path = "/api/programs/docufiller/versions"
			r.HandleContext(c)
		})
	}

	// 兼容下载路由
	// GET /api/download/:channel/:version -> /api/programs/docufiller/download/:channel/:version
	r.GET("/api/download/:channel/:version", deprecationMiddleware, func(c *gin.Context) {
		c.Request.URL.Path = "/api/programs/docufiller/download/" + c.Param("channel") + "/" + c.Param("version")
		r.HandleContext(c)
	})
}
