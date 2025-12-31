package middleware

import (
	"yflow/internal/api/response"
	"os"

	"github.com/gin-gonic/gin"
)

// APIKeyAuthMiddleware API Key认证中间件
func (f *MiddlewareFactory) APIKeyAuthMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 从环境变量获取API Key
		expectedAPIKey := os.Getenv("CLI_API_KEY")
		if expectedAPIKey == "" {
			// 如果没有设置环境变量，使用默认值（开发环境）
			expectedAPIKey = "yflow-cli-default-key"
		}

		// 从请求头获取API Key
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			response.Unauthorized(c, "API Key is required")
			c.Abort()
			return
		}

		// 验证API Key
		if apiKey != expectedAPIKey {
			response.Unauthorized(c, "Invalid API Key")
			c.Abort()
			return
		}

		// 验证通过，继续处理请求
		c.Next()
	})
}