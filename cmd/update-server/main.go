package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"docufiller-update-server/internal/config"
	"docufiller-update-server/internal/database"
	"docufiller-update-server/internal/handler"
	"docufiller-update-server/internal/logger"
	"docufiller-update-server/internal/middleware"
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
	if err := database.AutoMigrate(db); err != nil {
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

	// 初始化服务
	storageService := service.NewStorageService(cfg.Storage.BasePath)
	programService := service.NewProgramService(db)
	versionService := service.NewVersionService(db, storageService)
	setupService := service.NewSetupService(db)
	clientPackagerService := service.NewClientPackager(programService)

	// 初始化 handlers
	setupHandler := handler.NewSetupHandler(setupService)
	authHandler := handler.NewAuthHandler(db)

	adminHandler := handler.NewAdminHandler(
		programService,
		versionService,
		tokenSvc,
		setupService,
		clientPackagerService,
	)

	// 检查初始化状态
	initialized, err := setupService.IsInitialized()
	if err != nil {
		logger.Fatalf("Failed to check initialization status: %v", err)
	}

	// 设置根路由 - 如果未初始化，重定向到 setup 页面
	if !initialized {
		r.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusFound, "/setup")
		})
	} else {
		// 已初始化，设置管理后台首页
		r.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusFound, "/admin")
		})
	}

	// Setup 页面路由 - 未初始化时可用
	setupGroup := r.Group("/setup")
	{
		// Setup 页面
		setupGroup.GET("", func(c *gin.Context) {
			c.HTML(http.StatusOK, "setup.html", gin.H{
				"title": "初始化向导",
			})
		})

		// 认证相关 - 初始化页面内需要登录
		setupGroup.POST("/login", authHandler.Login)
		setupGroup.GET("/login", func(c *gin.Context) {
			c.HTML(http.StatusOK, "login.html", gin.H{
				"title": "管理员登录",
				"action": "/setup/login",
			})
		})
	}

	// Admin 页面路由 - 需要认证
	adminGroup := r.Group("/admin")
	if initialized {
		adminGroup.Use(handler.AuthMiddleware())
	}
	{
		adminGroup.GET("", func(c *gin.Context) {
			if initialized {
				c.HTML(http.StatusOK, "admin.html", gin.H{
					"title": "管理后台",
				})
			} else {
				c.Redirect(http.StatusFound, "/setup")
			}
		})

		adminGroup.GET("/login", func(c *gin.Context) {
			c.HTML(http.StatusOK, "login.html", gin.H{
				"title": "管理员登录",
				"action": "/admin/login",
			})
		})

		adminGroup.POST("/login", authHandler.Login)
		adminGroup.POST("/logout", authHandler.Logout)
	}

	// Setup API 路由 - 未初始化时可用
	setupAPI := r.Group("/api/setup")
	if !initialized {
		setupAPI.Use(func(c *gin.Context) {
			initialized, _ := setupService.IsInitialized()
			if initialized {
				c.JSON(http.StatusForbidden, gin.H{"error": "服务器已初始化"})
				c.Abort()
				return
			}
			c.Next()
		})
	}
	{
		setupAPI.GET("/status", setupHandler.CheckInitStatus)
		setupAPI.POST("/initialize", setupHandler.Initialize)
	}

	// Admin API 路由 - 需要认证
	adminAPI := r.Group("/api/admin")
	adminAPI.Use(handler.AuthMiddleware())
	{
		// 统计信息
		adminAPI.GET("/stats", adminHandler.GetStats)

		// 程序管理
		adminAPI.GET("/programs", adminHandler.ListPrograms)
		adminAPI.POST("/programs", adminHandler.CreateProgram)
		adminAPI.GET("/programs/:programId", adminHandler.GetProgramDetail)
		adminAPI.DELETE("/programs/:programId", adminHandler.DeleteProgram)

		// 版本管理
		adminAPI.GET("/programs/:programId/versions", adminHandler.ListVersions)
		adminAPI.DELETE("/programs/:programId/versions/:version", adminHandler.DeleteVersion)

		// 客户端包
		adminAPI.GET("/programs/:programId/client/publish", adminHandler.DownloadPublishClient)
		adminAPI.GET("/programs/:programId/client/update", adminHandler.DownloadUpdateClient)

		// Token 管理
		adminAPI.POST("/programs/:programId/tokens/regenerate", adminHandler.RegenerateToken)

		// 加密密钥管理
		adminAPI.POST("/programs/:programId/encryption/regenerate", adminHandler.RegenerateEncryptionKey)
	}

	// 公开 API 路由
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

	// 静态文件服务 - 作为 fallback 处理
	r.NoRoute(func(c *gin.Context) {
		// 尝试提供静态文件
		fileServer := http.FileServer(http.Dir("./web"))
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Infof("Server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}