package routes

import (
	internal_utils "yflow/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// setupMonitoringRoutes 设置监控路由
func (r *Router) setupMonitoringRoutes(engine *gin.Engine, monitor *internal_utils.SimpleMonitor) {
	// 健康检查端点（替换原有的简单健康检查）
	engine.GET("/health", monitor.HealthCheck)

	// 基础统计端点
	engine.GET("/stats", monitor.SimpleStats)

	// 详细统计端点
	engine.GET("/stats/detailed", monitor.DetailedStats)

	r.Logger.Info("Monitoring endpoints configured",
		zap.String("health_check", "GET /health"),
		zap.String("basic_stats", "GET /stats"),
		zap.String("detailed_stats", "GET /stats/detailed"),
	)
}
