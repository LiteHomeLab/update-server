package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"docufiller-update-server/internal/config"
	"docufiller-update-server/internal/database"
	"docufiller-update-server/internal/handler"
	"docufiller-update-server/internal/logger"
	"docufiller-update-server/internal/middleware"
	"docufiller-update-server/internal/service"
	"docufiller-update-server/web"
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

	// 添加 Session 中间件
	store := cookie.NewStore([]byte(cfg.Crypto.MasterKey))
	r.Use(sessions.Sessions("admin-session", store))

	// 加载 HTML 模板 (使用嵌入的文件系统)
	r.LoadHTMLFS(http.FS(web.Files), "*.html")

	// 注册加密中间件
	r.Use(cryptoMiddleware.Process())

	// 初始化服务
	storageService := service.NewStorageService(cfg.Storage.BasePath)
	programService := service.NewProgramService(db)
	versionService := service.NewVersionService(db, storageService)
	clientPackagerService := service.NewClientPackager(programService)

	// 初始化 handlers
	authHandler := handler.NewAuthHandler(cfg)

	adminHandler := handler.NewAdminHandler(
		programService,
		versionService,
		tokenSvc,
		clientPackagerService,
	)

	// 根路径直接重定向到管理后台
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/admin")
	})

	// Admin 登录路由 - 无需认证
	r.GET("/admin/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"title": "管理员登录",
			"action": "/admin/login",
		})
	})
	r.POST("/admin/login", authHandler.Login)

	// Admin 页面路由 - 需要认证
	adminGroup := r.Group("/admin")
	adminGroup.Use(handler.AuthMiddleware())
	{
		adminGroup.GET("", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin.html", gin.H{
				"title": "管理后台",
			})
		})

		adminGroup.POST("/logout", authHandler.Logout)
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

	// 静态文件服务 - 使用嵌入的文件系统
	fileServer := http.FileServer(http.FS(web.Files))
	r.NoRoute(func(c *gin.Context) {
		// 检查是否是静态文件请求
		path := c.Request.URL.Path
		// 跳过根路径
		if path == "/" || path == "" {
			c.Status(http.StatusNotFound)
			return
		}
		trimmedPath := strings.TrimPrefix(path, "/")
		// 跳过空路径
		if trimmedPath == "" {
			c.Status(http.StatusNotFound)
			return
		}
		if _, err := web.Files.Open(trimmedPath); err == nil {
			fileServer.ServeHTTP(c.Writer, c.Request)
		} else {
			// 不是静态文件，返回 404
			c.Status(http.StatusNotFound)
		}
	})

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Infof("Server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}