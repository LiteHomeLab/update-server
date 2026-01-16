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
	middleware.InitAuth(cfg.API.UploadToken)

	// 设置 Gin
	if cfg.Logger.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// 注册路由
	setupRoutes(r, db)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Infof("Server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes(r *gin.Engine, db *gorm.DB) {
	// 健康检查
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 版本相关路由
	versionHandler := handler.NewVersionHandler(db)
	api := r.Group("/api")
	{
		api.GET("/version/latest", versionHandler.GetLatestVersion)
		api.GET("/version/list", versionHandler.GetVersionList)
		api.GET("/version/:channel/:version", versionHandler.GetVersionDetail)
		api.POST("/version/upload", middleware.AuthMiddleware(), versionHandler.UploadVersion)
		api.DELETE("/version/:channel/:version", middleware.AuthMiddleware(), versionHandler.DeleteVersion)
		api.GET("/download/:channel/:version", versionHandler.DownloadFile)
	}
}
