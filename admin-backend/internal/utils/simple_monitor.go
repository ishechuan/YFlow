package utils

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// SimpleMonitor 简单监控器
type SimpleMonitor struct {
	startTime     time.Time
	requestCount  int64
	errorCount    int64
	slowRequests  int64
	lastErrorTime time.Time
	db            *gorm.DB
	redisClient   *redis.Client
}

// MonitorStats 监控统计信息
type MonitorStats struct {
	Status        string    `json:"status"`
	Uptime        string    `json:"uptime"`
	UptimeSeconds int64     `json:"uptime_seconds"`
	RequestCount  int64     `json:"request_count"`
	ErrorCount    int64     `json:"error_count"`
	SlowRequests  int64     `json:"slow_requests"`
	ErrorRate     string    `json:"error_rate"`
	LastErrorTime string    `json:"last_error_time,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
	Version       string    `json:"version"`
	Database      string    `json:"database"`
	Redis         string    `json:"redis"`
}

// NewSimpleMonitor 创建简单监控器实例
func NewSimpleMonitor(db *gorm.DB, redisClient *redis.Client) *SimpleMonitor {
	return &SimpleMonitor{
		startTime:   time.Now(),
		db:          db,
		redisClient: redisClient,
	}
}

// RecordRequest 记录请求
func (m *SimpleMonitor) RecordRequest() {
	atomic.AddInt64(&m.requestCount, 1)
}

// RecordError 记录错误
func (m *SimpleMonitor) RecordError() {
	atomic.AddInt64(&m.errorCount, 1)
	m.lastErrorTime = time.Now()
}

// RecordSlowRequest 记录慢请求
func (m *SimpleMonitor) RecordSlowRequest() {
	atomic.AddInt64(&m.slowRequests, 1)
}

// GetStats 获取统计信息
func (m *SimpleMonitor) GetStats() MonitorStats {
	uptime := time.Since(m.startTime)
	requestCount := atomic.LoadInt64(&m.requestCount)
	errorCount := atomic.LoadInt64(&m.errorCount)
	slowRequests := atomic.LoadInt64(&m.slowRequests)

	var errorRate string
	if requestCount > 0 {
		rate := float64(errorCount) / float64(requestCount) * 100
		errorRate = fmt.Sprintf("%.2f%%", rate)
	} else {
		errorRate = "0.00%"
	}

	var lastErrorTimeStr string
	if !m.lastErrorTime.IsZero() {
		lastErrorTimeStr = m.lastErrorTime.Format("2006-01-02 15:04:05")
	}

	// 检查服务状态
	status := "healthy"
	dbStatus := m.checkDatabase()
	redisStatus := m.checkRedis()

	if !dbStatus {
		status = "unhealthy"
	}

	return MonitorStats{
		Status:        status,
		Uptime:        uptime.String(),
		UptimeSeconds: int64(uptime.Seconds()),
		RequestCount:  requestCount,
		ErrorCount:    errorCount,
		SlowRequests:  slowRequests,
		ErrorRate:     errorRate,
		LastErrorTime: lastErrorTimeStr,
		Timestamp:     time.Now(),
		Version:       "1.0.0",
		Database:      m.getDatabaseStatus(dbStatus),
		Redis:         m.getRedisStatus(redisStatus),
	}
}

// HealthCheck 健康检查端点
func (m *SimpleMonitor) HealthCheck(c *gin.Context) {
	stats := m.GetStats()

	// 根据状态返回不同的HTTP状态码
	if stats.Status == "healthy" {
		c.JSON(200, stats)
	} else {
		c.JSON(503, stats)
	}
}

// SimpleStats 简单统计端点
func (m *SimpleMonitor) SimpleStats(c *gin.Context) {
	stats := m.GetStats()
	c.JSON(200, stats)
}

// DetailedStats 详细统计端点
func (m *SimpleMonitor) DetailedStats(c *gin.Context) {
	stats := m.GetStats()

	// 添加更多详细信息
	detailed := gin.H{
		"basic_stats": stats,
		"system_info": m.getSystemInfo(),
		"performance": m.getPerformanceMetrics(),
	}

	c.JSON(200, detailed)
}

// checkDatabase 检查数据库连接
func (m *SimpleMonitor) checkDatabase() bool {
	if m.db == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	sqlDB, err := m.db.DB()
	if err != nil {
		// 这里不能直接使用 utils 包的函数，因为会造成循环导入
		return false
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		// 这里不能直接使用 utils 包的函数，因为会造成循环导入
		// 简单记录到标准错误输出
		return false
	}

	return true
}

// checkRedis 检查Redis连接
func (m *SimpleMonitor) checkRedis() bool {
	if m.redisClient == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.redisClient.Ping(ctx).Result()
	if err != nil {
		// 这里不能直接使用 utils 包的函数，因为会造成循环导入
		return false
	}

	return true
}

// getDatabaseStatus 获取数据库状态描述
func (m *SimpleMonitor) getDatabaseStatus(isHealthy bool) string {
	if !isHealthy {
		return "down"
	}

	if m.db == nil {
		return "not_configured"
	}

	// 获取连接池信息
	sqlDB, err := m.db.DB()
	if err != nil {
		return "error"
	}

	stats := sqlDB.Stats()
	return fmt.Sprintf("healthy (open: %d, idle: %d)", stats.OpenConnections, stats.Idle)
}

// getRedisStatus 获取Redis状态描述
func (m *SimpleMonitor) getRedisStatus(isHealthy bool) string {
	if !isHealthy {
		return "down"
	}

	if m.redisClient == nil {
		return "not_configured"
	}

	return "healthy"
}

// getSystemInfo 获取系统信息
func (m *SimpleMonitor) getSystemInfo() gin.H {
	return gin.H{
		"go_version":   "go1.23",
		"service_name": "yflow-backend",
		"environment":  getEnv("ENV", "development"),
		"log_level":    getEnv("LOG_LEVEL", "info"),
	}
}

// getPerformanceMetrics 获取性能指标
func (m *SimpleMonitor) getPerformanceMetrics() gin.H {
	requestCount := atomic.LoadInt64(&m.requestCount)
	uptime := time.Since(m.startTime)

	var avgRequestsPerSecond float64
	if uptime.Seconds() > 0 {
		avgRequestsPerSecond = float64(requestCount) / uptime.Seconds()
	}

	return gin.H{
		"avg_requests_per_second": fmt.Sprintf("%.2f", avgRequestsPerSecond),
		"uptime_hours":            fmt.Sprintf("%.2f", uptime.Hours()),
	}
}

// getEnv 获取环境变量
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
