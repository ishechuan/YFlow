package routes

import (
	"yflow/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

// setupTranslationRoutes 设置翻译相关路由
func (r *Router) setupTranslationRoutes(authRoutes *gin.RouterGroup) {
	translationRoutes := authRoutes.Group("/translations")
	{
		// 需要项目查看权限的操作
		translationViewRoutes := translationRoutes.Group("")
		translationViewRoutes.Use(r.middlewareFactory.RequireProjectViewer())
		{
			translationViewRoutes.GET("/by-project/:project_id", r.TranslationHandler.GetByProjectID)
			translationViewRoutes.GET("/matrix/by-project/:project_id", r.TranslationHandler.GetMatrix)
			translationViewRoutes.GET("/:id", r.TranslationHandler.GetByID)
		}

		// 需要项目编辑权限的操作
		translationEditRoutes := translationRoutes.Group("")
		translationEditRoutes.Use(r.middlewareFactory.RequireProjectEditor())
		{
			translationEditRoutes.POST("", r.TranslationHandler.Create)
			translationEditRoutes.PUT("/:id", r.TranslationHandler.Update)
			translationEditRoutes.DELETE("/:id", r.TranslationHandler.Delete)
		}
	}

	// 批量操作路由组（应用批量操作限流中间件和项目编辑权限）
	batchRoutes := authRoutes.Group("/translations")
	batchRoutes.Use(middleware.TollboothBatchOperationRateLimitMiddleware())
	batchRoutes.Use(r.middlewareFactory.RequireProjectEditor())
	{
		batchRoutes.POST("/batch", r.TranslationHandler.CreateBatch)
		batchRoutes.POST("/batch-delete", r.TranslationHandler.DeleteBatch)
	}

	// 导出路由（应用批量操作限流中间件和项目查看权限）
	exportRoutes := authRoutes.Group("/exports")
	exportRoutes.Use(middleware.TollboothBatchOperationRateLimitMiddleware())
	exportRoutes.Use(r.middlewareFactory.RequireProjectViewer()) // 导出只需要查看权限
	{
		exportRoutes.GET("/project/:project_id", r.TranslationHandler.Export)
	}

	// 导入路由（应用批量操作限流中间件和项目编辑权限）
	importRoutes := authRoutes.Group("/imports")
	importRoutes.Use(middleware.TollboothBatchOperationRateLimitMiddleware())
	importRoutes.Use(r.middlewareFactory.RequireProjectEditor()) // 导入需要编辑权限
	{
		importRoutes.POST("/project/:project_id", r.TranslationHandler.Import)
	}

	// 机器翻译路由（应用限流中间件和项目编辑权限）
	machineTranslateRoutes := authRoutes.Group("/translations/machine-translate")
	machineTranslateRoutes.Use(middleware.TollboothBatchOperationRateLimitMiddleware())
	machineTranslateRoutes.Use(r.middlewareFactory.RequireProjectEditor())
	{
		machineTranslateRoutes.GET("/languages", r.TranslationHandler.GetSupportedLanguages)
		machineTranslateRoutes.GET("/health", r.TranslationHandler.HealthCheck)
	}

	// 自动填充语言路由
	autoFillRoutes := authRoutes.Group("/projects")
	autoFillRoutes.Use(middleware.TollboothBatchOperationRateLimitMiddleware())
	autoFillRoutes.Use(r.middlewareFactory.RequireProjectEditor())
	{
		autoFillRoutes.POST("/:project_id/auto-fill-language", r.TranslationHandler.AutoFillLanguage)
	}

	// 翻译历史路由
	// 单个翻译历史（需要项目查看权限）
	translationHistoryRoutes := translationRoutes.Group("")
	translationHistoryRoutes.Use(r.middlewareFactory.RequireProjectViewer())
	{
		translationHistoryRoutes.GET("/:id/history", r.TranslationHistoryHandler.GetByTranslationID)
	}

	// 项目翻译历史路由（需要项目查看权限）
	projectHistoryRoutes := authRoutes.Group("/projects")
	projectHistoryRoutes.Use(r.middlewareFactory.RequireProjectViewer())
	{
		projectHistoryRoutes.GET("/:project_id/translation-history", r.TranslationHistoryHandler.GetByProjectID)
	}

	// 用户翻译历史路由（需要认证）
	userHistoryRoutes := authRoutes.Group("/users")
	{
		userHistoryRoutes.GET("/:id/translation-history", r.TranslationHistoryHandler.GetByUserID)
	}
}
