package handler

import (
	"net/http"
	"docufiller-update-server/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 管理员登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var admin models.AdminUser
	if err := h.db.Where("username = ?", req.Username).First(&admin).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	if !admin.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 设置 session
	c.SetCookie("admin_session", admin.Username, 3600*24, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Logout 退出登录
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("admin_session", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
}

// AuthMiddleware 管理员认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := c.Cookie("admin_session")
		if err != nil || session == "" {
			if c.Request.URL.Path != "/login" {
				c.Redirect(http.StatusFound, "/login")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}