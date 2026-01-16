package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"docufiller-update-server/internal/logger"
)

var uploadToken string

func InitAuth(token string) {
	uploadToken = token
	logger.Info("Auth middleware initialized")
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只对上传和删除操作进行认证
		if !strings.Contains(c.Request.URL.Path, "/upload") &&
			!strings.Contains(c.Request.URL.Path, "/delete") {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warnf("Upload request without authorization from %s", c.ClientIP())
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token != uploadToken {
			logger.Warnf("Invalid upload token from %s", c.ClientIP())
			c.JSON(403, gin.H{"error": "Forbidden"})
			c.Abort()
			return
		}

		logger.Debugf("Authorized request from %s", c.ClientIP())
		c.Next()
	}
}
