package middleware

import (
	"fmt"
	"yflow/internal/api/response"
	"yflow/internal/domain"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorHandlerMiddleware 创建带 logger 的错误处理中间件
func ErrorHandlerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// 获取请求信息
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("request_id", getRequestIDFromContext(c)),
		}

		// 添加用户信息（如果存在）
		if userID, exists := c.Get("userID"); exists {
			fields = append(fields, zap.Any("user_id", userID))
		}

		if err, ok := recovered.(string); ok {
			logger.Error("Panic recovered", append(fields,
				zap.String("error", err),
				zap.String("stack", string(debug.Stack())),
			)...)
			response.InternalServerError(c, "服务器发生异常")
		} else if err, ok := recovered.(error); ok {
			logger.Error("Panic recovered", append(fields,
				zap.Error(err),
				zap.String("stack", string(debug.Stack())),
			)...)
			response.InternalServerError(c, "服务器发生异常")
		} else {
			logger.Error("Panic recovered", append(fields,
				zap.Any("error", recovered),
				zap.String("stack", string(debug.Stack())),
			)...)
			response.InternalServerError(c, "服务器发生异常")
		}
		c.Abort()
	})
}

// getRequestIDFromContext 从上下文获取请求ID
func getRequestIDFromContext(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}
	return ""
}

// AppErrorHandlerMiddleware 创建带 logger 的应用程序错误处理中间件
func AppErrorHandlerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// 获取请求信息
			fields := []zap.Field{
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()),
				zap.String("request_id", getRequestIDFromContext(c)),
			}

			// 添加用户信息（如果存在）
			if userID, exists := c.Get("userID"); exists {
				fields = append(fields, zap.Any("user_id", userID))
			}

			// 检查是否为应用程序错误
			if appErr, ok := domain.IsAppError(err); ok {
				// 记录错误日志
				logger.Error("Application error", append(fields,
					zap.String("error_type", string(appErr.Type)),
					zap.String("error_code", appErr.Code),
					zap.String("error_message", appErr.Message),
					zap.Any("error_context", appErr.Context),
					zap.Error(appErr.Cause),
				)...)

				// 返回结构化错误响应
				c.JSON(appErr.HTTPStatus(), response.APIResponse{
					Success: false,
					Error: &response.ErrorInfo{
						Code:    appErr.Code,
						Message: appErr.Message,
						Details: appErr.Details,
					},
				})
				return
			}

			// 处理其他类型的错误
			logger.Error("Unhandled error", append(fields, zap.Error(err))...)
			response.InternalServerError(c, "服务器内部错误")
		}
	}
}

// NotFoundHandler 404处理器
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.NotFound(c, fmt.Sprintf("路由 %s %s 不存在", c.Request.Method, c.Request.URL.Path))
	}
}
