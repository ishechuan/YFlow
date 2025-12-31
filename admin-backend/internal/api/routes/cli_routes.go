package routes

import (
	"yflow/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

// setupCLIRoutes 设置CLI相关路由
func (r *Router) setupCLIRoutes(rg *gin.RouterGroup) {
	// CLI路由使用API Key认证和API限流
	cliRoutes := rg.Group("/cli")
	cliRoutes.Use(r.middlewareFactory.APIKeyAuthMiddleware())
	cliRoutes.Use(middleware.TollboothAPIRateLimitMiddleware())
	{
		// CLI身份验证
		cliRoutes.GET("/auth", r.CLIHandler.Auth)

		// 获取翻译数据
		cliRoutes.GET("/translations", r.CLIHandler.GetTranslations)
	}

	// 推送翻译键（批量操作，应用批量操作限流）
	batchCliRoutes := rg.Group("/cli")
	batchCliRoutes.Use(r.middlewareFactory.APIKeyAuthMiddleware())
	batchCliRoutes.Use(middleware.TollboothBatchOperationRateLimitMiddleware())
	{
		batchCliRoutes.POST("/keys", r.CLIHandler.PushKeys)
	}
}
