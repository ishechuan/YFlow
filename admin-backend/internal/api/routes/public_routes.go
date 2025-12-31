package routes

import (
	"yflow/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

// setupPublicRoutes 设置公开路由
func (r *Router) setupPublicRoutes(rg *gin.RouterGroup) {
	// 登录路由组（应用登录限流中间件）
	loginRoutes := rg.Group("")
	loginRoutes.Use(middleware.TollboothLoginRateLimitMiddleware())
	{
		// 公开的认证路由（每秒5个请求，突发10个）
		loginRoutes.POST("/login", r.UserHandler.Login)
		loginRoutes.POST("/refresh", r.UserHandler.RefreshToken)
	}
}
