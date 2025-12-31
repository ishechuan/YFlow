package middleware

import (
	"bytes"
	internal_utils "yflow/internal/utils"
	log_utils "yflow/utils"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggingOptions 日志中间件选项
type LoggingOptions struct {
	Monitor              *internal_utils.SimpleMonitor // 监控实例（可选）
	LogRequestBody       bool                          // 是否记录请求体
	SlowRequestThreshold time.Duration                 // 慢请求阈值
}

// DefaultLoggingOptions 默认选项
var defaultLoggingOptions = LoggingOptions{
	LogRequestBody:       false,
	SlowRequestThreshold: time.Second,
}

// LoggingMiddleware 统一日志中间件
// 支持通过 LoggingOptions 配置:
//   - Monitor: 监控实例，用于记录请求指标
//   - LogRequestBody: 是否记录请求体（默认 false）
//   - SlowRequestThreshold: 慢请求阈值（默认 1秒）
func LoggingMiddleware(logger *zap.Logger, opts ...LoggingOptions) gin.HandlerFunc {
	options := defaultLoggingOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	return func(c *gin.Context) {
		start := time.Now()

		// 包装响应写入器
		rw := &ResponseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
			statusCode:     200,
		}
		c.Writer = rw

		// 读取请求体（如果需要）
		var requestBody []byte
		if options.LogRequestBody && c.Request.Body != nil {
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				logger.Warn("Failed to read request body", zap.Error(err))
			} else {
				requestBody = body
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 处理请求
		c.Next()

		// 计算处理时间
		duration := time.Since(start)
		isSlowRequest := duration > options.SlowRequestThreshold

		// 记录监控指标（如果配置了监控）
		if options.Monitor != nil {
			options.Monitor.RecordRequest()
			if isSlowRequest {
				options.Monitor.RecordSlowRequest()
			}
			if rw.statusCode >= 400 {
				options.Monitor.RecordError()
			}
		}

		// 收集日志字段
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", log_utils.SanitizeLogValue(c.Request.URL.RawQuery)),
			zap.String("user_agent", log_utils.SanitizeLogValue(c.Request.UserAgent())),
			zap.String("client_ip", log_utils.SanitizeLogValue(c.ClientIP())),
			zap.Int("status_code", rw.statusCode),
			zap.Int("response_size", rw.size),
			zap.Duration("duration", duration),
			zap.String("request_id", GetRequestID(c)),
		}

		// 添加用户信息（如果存在）
		if userID, exists := c.Get("userID"); exists {
			fields = append(fields, zap.Any("user_id", userID))
		}
		if username, exists := c.Get("username"); exists {
			if s, ok := username.(string); ok {
				fields = append(fields, zap.String("username", log_utils.SanitizeLogValue(s)))
			}
		}

		// 添加请求体（如果启用且非敏感路径）
		if options.LogRequestBody && ShouldLogRequestBody(c.Request.URL.Path) && len(requestBody) > 0 && len(requestBody) < 1024 {
			fields = append(fields, zap.String("request_body", log_utils.SanitizeLogValue(string(requestBody))))
		}

		// 跳过日志记录的路径
		if _, skip := c.Get("skip_logging"); skip {
			return
		}

		// 记录到访问日志
		logger.Info("HTTP Request", fields...)

		// 慢请求警告日志
		if isSlowRequest {
			logger.Warn("Slow request detected",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()),
				zap.Duration("duration", duration),
				zap.Int("status", rw.statusCode),
			)
		}

		// 注意：所有请求已在第113行记录（包括status_code），无需重复记录错误
	}
}

// MonitoringStatsMiddleware 监控统计中间件（轻量级版本）
func MonitoringStatsMiddleware(monitor *internal_utils.SimpleMonitor) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		// 简单的统计收集
		duration := time.Since(start)
		status := c.Writer.Status()

		// 记录基础指标
		monitor.RecordRequest()

		if duration > time.Second {
			monitor.RecordSlowRequest()
		}

		if status >= 400 {
			monitor.RecordError()
		}

		// 设置响应头，方便调试
		c.Header("X-Response-Time", duration.String())
	}
}

// RequestIDMiddleware 请求ID中间件
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = GenerateRequestID()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// SkipLoggingMiddleware 跳过日志的中间件（用于健康检查等）
// paths: 需要跳过日志的路径列表
func SkipLoggingMiddleware(paths ...string) gin.HandlerFunc {
	skipPaths := make(map[string]bool)
	for _, path := range paths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		if skipPaths[c.Request.URL.Path] {
			c.Set("skip_logging", true)
		}
		// 跳过/swagger/路径的日志（Swagger UI）
		if IsSwaggerPath(c) {
			c.Set("skip_logging", true)
		}
		c.Next()
	}
}

