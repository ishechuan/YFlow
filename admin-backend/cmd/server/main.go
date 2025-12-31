package main

import (
	_ "yflow/docs" // 导入 swagger 文档（需要初始化 SwaggerInfo）
	"yflow/internal/api/middleware"
	"yflow/internal/config"
	"yflow/internal/container"
	internal_utils "yflow/internal/utils"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @title           YFlow API
// @version         1.0
// @description     语流是一个用于管理多语言翻译的系统。
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.example.com/support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入格式: Bearer {token}
func main() {
	// 加载配置
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 使用 FX 运行应用（阻塞直到收到停止信号）
	// FX 将自动管理：
	// - 依赖注入
	// - 生命周期（启动/停止）
	// - 优雅关闭
	container.Run(cfg, setupMiddleware)
}

// setupMiddleware 设置全局中间件
func setupMiddleware(router *gin.Engine, monitor *internal_utils.SimpleMonitor, logger *zap.Logger) {
	// 请求ID中间件（最先设置，确保所有后续中间件都能使用请求ID）
	router.Use(middleware.RequestIDMiddleware())

	// 统一日志中间件（第二个设置，确保所有请求都能被记录，并包含请求ID）
	// 集成监控，用于记录请求指标
	if monitor != nil {
		router.Use(middleware.LoggingMiddleware(logger, middleware.LoggingOptions{
			Monitor:              monitor,
			LogRequestBody:       false,
			SlowRequestThreshold: time.Second,
		}))
	} else {
		router.Use(middleware.LoggingMiddleware(logger))
	}

	// 安全HTTP头中间件
	router.Use(middleware.SecurityHeadersMiddleware())

	// 全局限流中间件（使用 tollbooth，每秒100个请求）
	router.Use(middleware.TollboothGlobalRateLimitMiddleware())

	// 安全验证中间件（跳过 swagger 路径）
	router.Use(middleware.SkipForSwagger(middleware.SecurityValidationMiddleware(logger)))

	// SQL安全中间件（跳过 swagger 路径）
	router.Use(middleware.SkipForSwagger(middleware.SQLSecurityMiddleware(logger)))

	// 增强输入验证中间件（跳过 swagger 路径）
	router.Use(middleware.SkipForSwagger(middleware.EnhancedInputValidationMiddleware()))

	// XSS防护中间件
	router.Use(middleware.XSSProtectionMiddleware(logger))

	// CSP违规报告中间件
	router.Use(middleware.CSPViolationReportMiddleware(logger))

	// 跳过监控端点和 swagger 的日志记录
	router.Use(middleware.SkipLoggingMiddleware("/health", "/stats", "/metrics"))

	// 全局错误处理中间件
	router.Use(middleware.ErrorHandlerMiddleware(logger))

	// 应用程序错误处理中间件
	router.Use(middleware.AppErrorHandlerMiddleware(logger))

	// 请求大小限制中间件 (32MB)
	router.Use(middleware.RequestSizeLimitMiddleware(32 << 20))

	// 请求验证中间件（跳过 swagger 路径）
	router.Use(middleware.SkipForSwagger(middleware.RequestValidationMiddleware()))

	// 分页参数验证中间件（跳过 swagger 路径）
	router.Use(middleware.SkipForSwagger(middleware.PaginationValidationMiddleware()))

	// 允许跨域请求
	router.Use(middleware.CORSMiddleware())

	// 404处理器
	router.NoRoute(middleware.NotFoundHandler())
}
