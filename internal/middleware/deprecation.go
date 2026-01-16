package middleware

import "github.com/gin-gonic/gin"

// DeprecationWarning adds deprecation warning headers to API responses
func DeprecationWarning() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-API-Deprecation", "This API is deprecated. Use /api/programs/* instead.")
		c.Next()
	}
}
