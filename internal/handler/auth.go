package handler

import (
	"net/http"
	"docufiller-update-server/internal/config"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	cfg *config.Config
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{cfg: cfg}
}

// Login 处理登录请求
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 从配置验证凭据
	if req.Username != h.cfg.Admin.Username || req.Password != h.cfg.Admin.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 设置 Session
	session := sessions.Default(c)
	session.Set("authenticated", true)
	session.Set("username", req.Username)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Logout 处理登出请求
func (h *AuthHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// AuthMiddleware 管理后台认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		authenticated := session.Get("authenticated")
		if authenticated != true {
			// 未登录，返回 401 或重定向到登录页
			if c.Request.Header.Get("Content-Type") == "application/json" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			} else {
				c.Redirect(http.StatusFound, "/admin/login")
			}
			c.Abort()
			return
		}

		c.Next()
	}
}
