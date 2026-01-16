package middleware

import (
	"docufiller-update-server/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	tokenSvc *service.TokenService
}

func NewAuthMiddleware(tokenSvc *service.TokenService) *AuthMiddleware {
	return &AuthMiddleware{tokenSvc: tokenSvc}
}

// RequireAuth 需要认证（任意有效 Token）
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return m.requireAuth("")
}

// RequireAdmin 需要管理员 Token
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.requireAuth("admin")
}

// RequireUpload 需要上传权限（Admin 或 Upload Token）
func (m *AuthMiddleware) RequireUpload() gin.HandlerFunc {
	return m.requireAuthWithProgram("upload")
}

// RequireDownload 需要下载权限
func (m *AuthMiddleware) RequireDownload() gin.HandlerFunc {
	return m.requireAuthWithProgram("download")
}

func (m *AuthMiddleware) requireAuth(requiredType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			c.JSON(401, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		tokenRecord, err := m.tokenSvc.ValidateToken(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// 检查 Token 类型
		if requiredType != "" &&
			tokenRecord.TokenType != requiredType &&
			tokenRecord.TokenType != "admin" {
			c.JSON(403, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		c.Set("token", tokenRecord)
		c.Next()
	}
}

func (m *AuthMiddleware) requireAuthWithProgram(requiredType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			c.JSON(401, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		tokenRecord, err := m.tokenSvc.ValidateToken(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		programID := c.Param("programId")
		if !m.tokenSvc.HasPermission(tokenRecord, requiredType, programID) {
			c.JSON(403, gin.H{"error": "program access denied"})
			c.Abort()
			return
		}

		c.Set("token", tokenRecord)
		c.Next()
	}
}

func (m *AuthMiddleware) extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// OptionalAuth 可选认证（支持匿名访问）
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token != "" {
			if tokenRecord, err := m.tokenSvc.ValidateToken(token); err == nil {
				c.Set("token", tokenRecord)
			}
		}
		c.Next()
	}
}
