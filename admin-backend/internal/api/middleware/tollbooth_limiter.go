package middleware

import (
	"fmt"
	"yflow/internal/api/response"
	"net"
	"time"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth/v7/limiter"
	"github.com/gin-gonic/gin"
)

// TollboothLimitMiddleware 使用 tollbooth 的通用限流中间件
func TollboothLimitMiddleware(max float64, ttl time.Duration, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	lmt := tollbooth.NewLimiter(max, &limiter.ExpirableOptions{
		DefaultExpirationTTL: ttl,
	})

	// 设置 IP 提取函数
	if keyFunc != nil {
		lmt.SetIPLookups([]string{"X-Real-IP", "X-Forwarded-For", "RemoteAddr"})
	}

	return func(c *gin.Context) {
		// 获取限流键（通常是 IP）
		var key string
		if keyFunc != nil {
			key = keyFunc(c)
		} else {
			key = getClientIP(c)
		}

		// 检查限流
		err := tollbooth.LimitByKeys(lmt, []string{key})
		if err != nil {
			response.ErrorWithDetails(c, 429, "RATE_LIMIT_EXCEEDED",
				"请求过于频繁，请稍后再试",
				fmt.Sprintf("Rate limit exceeded for: %s", key))
			c.Abort()
			return
		}

		c.Next()
	}
}

// TollboothGlobalRateLimitMiddleware 全局限流中间件
func TollboothGlobalRateLimitMiddleware() gin.HandlerFunc {
	// 每秒100个请求，5分钟过期
	return TollboothLimitMiddleware(100, 5*time.Minute, nil)
}

// TollboothLoginRateLimitMiddleware 登录限流中间件
func TollboothLoginRateLimitMiddleware() gin.HandlerFunc {
	// 每秒5个请求，10分钟过期（防止暴力破解）
	return TollboothLimitMiddleware(5, 10*time.Minute, nil)
}

// TollboothAPIRateLimitMiddleware API限流中间件
func TollboothAPIRateLimitMiddleware() gin.HandlerFunc {
	// 每秒50个请求，5分钟过期
	return TollboothLimitMiddleware(50, 5*time.Minute, nil)
}

// TollboothBatchOperationRateLimitMiddleware 批量操作限流中间件
func TollboothBatchOperationRateLimitMiddleware() gin.HandlerFunc {
	// 每秒20个请求，10分钟过期（CLI批量导入用）
	return TollboothLimitMiddleware(20, 10*time.Minute, nil)
}

// TollboothCustomRateLimitMiddleware 自定义限流中间件
func TollboothCustomRateLimitMiddleware(max float64, ttl time.Duration) gin.HandlerFunc {
	return TollboothLimitMiddleware(max, ttl, nil)
}

// TollboothUserBasedRateLimitMiddleware 基于用户的限流中间件
func TollboothUserBasedRateLimitMiddleware(max float64, ttl time.Duration) gin.HandlerFunc {
	return TollboothLimitMiddleware(max, ttl, func(c *gin.Context) string {
		// 优先使用用户ID，如果没有则使用IP
		if userID, exists := c.Get("userID"); exists {
			return fmt.Sprintf("user:%v", userID)
		}
		return fmt.Sprintf("ip:%s", getClientIP(c))
	})
}

// getClientIP 获取客户端真实IP地址
func getClientIP(c *gin.Context) string {
	// 优先检查X-Real-IP头
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	// 检查X-Forwarded-For头（可能包含多个IP）
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		// 取第一个IP
		if firstIP := net.ParseIP(ip); firstIP != nil {
			return ip
		}
	}

	// 使用Gin的ClientIP方法作为后备
	return c.ClientIP()
}
