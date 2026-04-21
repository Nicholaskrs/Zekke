package middleware

import (
	"net/http"
	"template-go/util/config"

	"github.com/gin-gonic/gin"
)

func InternalMiddleware(config config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKeyConfig := config.ApiKey
		apiKeyHeader := c.GetHeader("X-API-Key")

		if apiKeyHeader == "" || apiKeyHeader != apiKeyConfig {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": http.StatusUnauthorized, "message": "Unauthorized access, Invalid/Missing token"})
			return
		}

		c.Next()
	}
}
