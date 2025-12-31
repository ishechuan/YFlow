package middleware

import (
	"fmt"
	"yflow/internal/api/response"
	log_utils "yflow/utils"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SQLSecurityConfig SQL安全配置
type SQLSecurityConfig struct {
	MaxQueryLength    int      // 最大查询长度
	AllowedSortFields []string // 允许的排序字段
	AllowedOperators  []string // 允许的操作符
	ForbiddenKeywords []string // 禁止的关键词
}

// DefaultSQLSecurityConfig 默认SQL安全配置
func DefaultSQLSecurityConfig() SQLSecurityConfig {
	return SQLSecurityConfig{
		MaxQueryLength: 1000,
		AllowedSortFields: []string{
			"id", "name", "created_at", "updated_at", "status",
			"username", "email", "project_id", "language_id",
			"key_name", "value", "context",
		},
		AllowedOperators: []string{"=", "!=", ">", "<", ">=", "<=", "LIKE", "IN"},
		ForbiddenKeywords: []string{
			"DROP", "DELETE", "TRUNCATE", "ALTER", "CREATE", "INSERT",
			"UPDATE", "EXEC", "EXECUTE", "UNION", "SCRIPT", "DECLARE",
			"CAST", "CONVERT", "SUBSTRING", "CHAR", "ASCII", "WAITFOR",
			"BENCHMARK", "SLEEP", "LOAD_FILE", "INTO OUTFILE", "INTO DUMPFILE",
		},
	}
}

// SQLSecurityMiddleware SQL安全中间件
func SQLSecurityMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return SQLSecurityMiddlewareWithConfig(logger, DefaultSQLSecurityConfig())
}

// SQLSecurityMiddlewareWithConfig 带配置的SQL安全中间件
func SQLSecurityMiddlewareWithConfig(logger *zap.Logger, config SQLSecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证查询参数
		if err := validateQueryParams(c, config, logger); err != nil {
			response.BadRequest(c, fmt.Sprintf("查询参数验证失败: %s", err.Error()))
			return
		}

		// 验证路径参数
		if err := validatePathParams(c, config, logger); err != nil {
			response.BadRequest(c, fmt.Sprintf("路径参数验证失败: %s", err.Error()))
			return
		}

		c.Next()
	}
}

// validateQueryParams 验证查询参数
func validateQueryParams(c *gin.Context, config SQLSecurityConfig, logger *zap.Logger) error {
	queryParams := c.Request.URL.Query()

	for key, values := range queryParams {
		for _, value := range values {
			// 检查参数长度
			if len(value) > config.MaxQueryLength {
				return fmt.Errorf("参数 %s 长度超过限制", key)
			}

			// 检查危险关键词
			if containsForbiddenKeywords(value, config.ForbiddenKeywords) {
				logger.Error("Suspicious query parameter detected",
					zap.String("param", key),
					zap.String("value", log_utils.SanitizeLogValue(value)),
					zap.String("ip", c.ClientIP()),
					zap.String("path", c.Request.URL.Path),
				)
				return fmt.Errorf("参数 %s 包含不允许的内容", key)
			}

			// 特殊参数验证
			switch key {
			case "sort", "order_by":
				if !isAllowedSortField(value, config.AllowedSortFields) {
					return fmt.Errorf("不允许的排序字段: %s", value)
				}
			case "limit":
				if !isValidLimit(value) {
					return fmt.Errorf("无效的限制参数: %s", value)
				}
			case "offset", "page":
				if !isValidOffset(value) {
					return fmt.Errorf("无效的偏移参数: %s", value)
				}
			}
		}
	}

	return nil
}

// validatePathParams 验证路径参数
func validatePathParams(c *gin.Context, config SQLSecurityConfig, logger *zap.Logger) error {
	// 获取路径参数
	params := c.Params

	for _, param := range params {
		// 检查参数长度
		if len(param.Value) > config.MaxQueryLength {
			return fmt.Errorf("路径参数 %s 长度超过限制", param.Key)
		}

		// 检查危险关键词
		if containsForbiddenKeywords(param.Value, config.ForbiddenKeywords) {
			logger.Error("Suspicious path parameter detected",
				zap.String("param", param.Key),
				zap.String("value", log_utils.SanitizeLogValue(param.Value)),
				zap.String("ip", c.ClientIP()),
				zap.String("path", c.Request.URL.Path),
			)
			return fmt.Errorf("路径参数 %s 包含不允许的内容", param.Key)
		}

		// ID参数特殊验证
		if param.Key == "id" || strings.HasSuffix(param.Key, "_id") {
			if !isValidID(param.Value) {
				return fmt.Errorf("无效的ID参数: %s", param.Value)
			}
		}
	}

	return nil
}

// containsForbiddenKeywords 检查是否包含禁止的关键词
func containsForbiddenKeywords(input string, keywords []string) bool {
	inputUpper := strings.ToUpper(input)

	for _, keyword := range keywords {
		// 使用单词边界匹配，避免误判
		pattern := `\b` + regexp.QuoteMeta(strings.ToUpper(keyword)) + `\b`
		if matched, _ := regexp.MatchString(pattern, inputUpper); matched {
			return true
		}
	}

	// 检查SQL注入常见模式
	sqlInjectionPatterns := []string{
		`'.*OR.*'.*'`,
		`'.*AND.*'.*'`,
		`'.*UNION.*SELECT`,
		`'.*;\s*(DROP|DELETE|INSERT|UPDATE)`,
		`--.*`,
		`/\*.*\*/`,
		`'.*'.*=.*'.*'`,
	}

	for _, pattern := range sqlInjectionPatterns {
		if matched, _ := regexp.MatchString(pattern, inputUpper); matched {
			return true
		}
	}

	return false
}

// isAllowedSortField 检查是否为允许的排序字段
func isAllowedSortField(field string, allowedFields []string) bool {
	// 处理带方向的排序字段 (例如: "name DESC", "id ASC")
	parts := strings.Fields(strings.ToLower(field))
	if len(parts) > 2 {
		return false
	}

	fieldName := parts[0]

	// 检查字段名是否在白名单中
	for _, allowed := range allowedFields {
		if fieldName == strings.ToLower(allowed) {
			// 如果有排序方向，验证是否为 ASC 或 DESC
			if len(parts) == 2 {
				direction := parts[1]
				return direction == "asc" || direction == "desc"
			}
			return true
		}
	}

	return false
}

// isValidLimit 验证限制参数
func isValidLimit(limit string) bool {
	// 使用正则表达式验证是否为正整数，且不超过1000
	matched, _ := regexp.MatchString(`^[1-9]\d{0,2}$|^1000$`, limit)
	return matched
}

// isValidOffset 验证偏移参数
func isValidOffset(offset string) bool {
	// 验证是否为非负整数
	matched, _ := regexp.MatchString(`^\d+$`, offset)
	return matched
}

// isValidID 验证ID参数
func isValidID(id string) bool {
	// 验证是否为正整数
	matched, _ := regexp.MatchString(`^[1-9]\d*$`, id)
	return matched
}

// WhitelistQueryMiddleware 查询白名单中间件
func WhitelistQueryMiddleware(logger *zap.Logger, allowedParams []string) gin.HandlerFunc {
	allowedMap := make(map[string]bool)
	for _, param := range allowedParams {
		allowedMap[param] = true
	}

	return func(c *gin.Context) {
		queryParams := c.Request.URL.Query()

		for key := range queryParams {
			if !allowedMap[key] {
				logger.Warn("Unauthorized query parameter detected",
					zap.String("param", key),
					zap.String("ip", c.ClientIP()),
					zap.String("path", c.Request.URL.Path),
				)
				response.BadRequest(c, fmt.Sprintf("不允许的查询参数: %s", key))
				return
			}
		}

		c.Next()
	}
}

// DatabaseQueryLogMiddleware 数据库查询日志中间件
func DatabaseQueryLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录查询开始时间
		c.Set("query_start_time", time.Now())

		c.Next()

		// 这里可以添加查询结束后的日志记录
		// 例如：查询耗时、影响的行数等
	}
}

// SQLInjectionDetectionMiddleware SQL注入检测中间件
func SQLInjectionDetectionMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查所有输入参数
		suspiciousPatterns := []string{
			`'.*OR.*1.*=.*1`,
			`'.*OR.*'.*'.*=.*'.*'`,
			`'.*UNION.*SELECT`,
			`'.*;\s*DROP`,
			`'.*;\s*DELETE`,
			`'.*;\s*INSERT`,
			`'.*;\s*UPDATE`,
			`WAITFOR\s+DELAY`,
			`BENCHMARK\s*\(`,
			`SLEEP\s*\(`,
		}

		// 检查查询参数
		queryParams := c.Request.URL.Query()
		for key, values := range queryParams {
			for _, value := range values {
				for _, pattern := range suspiciousPatterns {
					if matched, _ := regexp.MatchString("(?i)"+pattern, value); matched {
						logger.Error("SQL injection attempt detected",
							zap.String("param", key),
							zap.String("value", log_utils.SanitizeLogValue(value)),
							zap.String("pattern", pattern),
							zap.String("ip", c.ClientIP()),
							zap.String("path", c.Request.URL.Path),
							zap.String("method", c.Request.Method),
						)
						response.BadRequest(c, "检测到恶意请求")
						return
					}
				}
			}
		}

		c.Next()
	}
}
