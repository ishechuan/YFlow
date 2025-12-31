package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"yflow/internal/api/response"
	log_utils "yflow/utils"
	"io"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"go.uber.org/zap"
)

// InputValidationConfig 输入验证配置
type InputValidationConfig struct {
	MaxStringLength   int      // 字符串最大长度
	MaxJSONSize       int64    // JSON最大大小
	AllowedFileTypes  []string // 允许的文件类型
	ForbiddenPatterns []string // 禁止的模式
	EnableHTMLClean   bool     // 是否启用HTML清理
	EnableXSSProtect  bool     // 是否启用XSS防护
}

// DefaultInputValidationConfig 默认配置
func DefaultInputValidationConfig() InputValidationConfig {
	return InputValidationConfig{
		MaxStringLength:   10000,                                                                  // 10KB
		MaxJSONSize:       1 << 20,                                                                // 1MB
		AllowedFileTypes:  []string{".json", ".csv", ".xlsx"},                                     // 允许的文件类型
		ForbiddenPatterns: []string{"<script", "javascript:", "vbscript:", "onload=", "onerror="}, // 危险模式
		EnableHTMLClean:   true,
		EnableXSSProtect:  true,
	}
}

// EnhancedInputValidationMiddleware 增强的输入验证中间件
func EnhancedInputValidationMiddleware() gin.HandlerFunc {
	return EnhancedInputValidationMiddlewareWithConfig(DefaultInputValidationConfig())
}

// EnhancedInputValidationMiddlewareWithConfig 带配置的增强输入验证中间件
func EnhancedInputValidationMiddlewareWithConfig(config InputValidationConfig) gin.HandlerFunc {
	// 创建HTML清理策略
	policy := bluemonday.UGCPolicy() // 用户生成内容策略
	if !config.EnableHTMLClean {
		policy = bluemonday.StrictPolicy() // 严格策略，移除所有HTML
	}

	// 编译危险模式正则表达式
	var forbiddenRegexps []*regexp.Regexp
	for _, pattern := range config.ForbiddenPatterns {
		if re, err := regexp.Compile("(?i)" + regexp.QuoteMeta(pattern)); err == nil {
			forbiddenRegexps = append(forbiddenRegexps, re)
		}
	}

	return func(c *gin.Context) {
		// 跳过非JSON请求
		if c.Request.Method == http.MethodGet ||
			c.Request.Method == http.MethodDelete ||
			!strings.Contains(c.GetHeader("Content-Type"), "application/json") {
			c.Next()
			return
		}

		// 检查请求大小
		if c.Request.ContentLength > config.MaxJSONSize {
			response.BadRequest(c, fmt.Sprintf("请求体过大，最大支持 %d bytes", config.MaxJSONSize))
			return
		}

		// 读取请求体
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			response.BadRequest(c, "无法读取请求体")
			return
		}

		// 恢复请求体供后续处理
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// 验证JSON格式
		var jsonData interface{}
		if err := json.Unmarshal(body, &jsonData); err != nil {
			response.BadRequest(c, "无效的JSON格式")
			return
		}

		// 递归验证和清理JSON数据
		cleanedData, err := validateAndCleanJSON(jsonData, config, policy, forbiddenRegexps)
		if err != nil {
			response.BadRequest(c, fmt.Sprintf("输入验证失败: %s", err.Error()))
			return
		}

		// 将清理后的数据重新序列化
		cleanedBody, err := json.Marshal(cleanedData)
		if err != nil {
			response.InternalServerError(c, "数据处理失败")
			return
		}

		// 更新请求体
		c.Request.Body = io.NopCloser(bytes.NewBuffer(cleanedBody))
		c.Request.ContentLength = int64(len(cleanedBody))

		c.Next()
	}
}

// validateAndCleanJSON 递归验证和清理JSON数据
func validateAndCleanJSON(data interface{}, config InputValidationConfig, policy *bluemonday.Policy, forbiddenRegexps []*regexp.Regexp) (interface{}, error) {
	switch v := data.(type) {
	case string:
		return validateAndCleanString(v, config, policy, forbiddenRegexps)
	case map[string]interface{}:
		cleaned := make(map[string]interface{})
		for key, value := range v {
			// 验证键名
			cleanKey, err := validateAndCleanString(key, config, policy, forbiddenRegexps)
			if err != nil {
				return nil, fmt.Errorf("无效的键名 '%s': %v", key, err)
			}

			// 递归验证值
			cleanValue, err := validateAndCleanJSON(value, config, policy, forbiddenRegexps)
			if err != nil {
				return nil, err
			}

			cleaned[cleanKey.(string)] = cleanValue
		}
		return cleaned, nil
	case []interface{}:
		cleaned := make([]interface{}, len(v))
		for i, item := range v {
			cleanItem, err := validateAndCleanJSON(item, config, policy, forbiddenRegexps)
			if err != nil {
				return nil, err
			}
			cleaned[i] = cleanItem
		}
		return cleaned, nil
	default:
		// 数字、布尔值等其他类型直接返回
		return data, nil
	}
}

// validateAndCleanString 验证和清理字符串
func validateAndCleanString(s string, config InputValidationConfig, policy *bluemonday.Policy, forbiddenRegexps []*regexp.Regexp) (interface{}, error) {
	// 检查字符串长度
	if len(s) > config.MaxStringLength {
		return nil, fmt.Errorf("字符串长度超过限制 (%d)", config.MaxStringLength)
	}

	// 检查UTF-8编码有效性
	if !utf8.ValidString(s) {
		return nil, fmt.Errorf("无效的UTF-8编码")
	}

	// 检查危险模式
	for _, re := range forbiddenRegexps {
		if re.MatchString(s) {
			return nil, fmt.Errorf("包含危险内容")
		}
	}

	// HTML清理（如果启用）
	if config.EnableHTMLClean {
		s = policy.Sanitize(s)
	}

	// 基础清理
	s = strings.TrimSpace(s)

	return s, nil
}

// ValidateEmailFormat 验证邮箱格式
func ValidateEmailFormat(email string) bool {
	return govalidator.IsEmail(email)
}

// ValidateURLFormat 验证URL格式
func ValidateURLFormat(url string) bool {
	return govalidator.IsURL(url)
}

// ValidateAlphanumeric 验证字母数字格式
func ValidateAlphanumeric(s string) bool {
	return govalidator.IsAlphanumeric(s)
}

// ValidateLength 验证字符串长度范围
func ValidateLength(s string, min, max int) bool {
	return govalidator.IsByteLength(s, min, max)
}

// SpecificFieldValidationMiddleware 特定字段验证中间件
func SpecificFieldValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 根据路径进行特定验证
		path := c.Request.URL.Path

		switch {
		case strings.Contains(path, "/login"):
			// 登录请求的特殊验证
			validateLoginRequest(c)
		case strings.Contains(path, "/projects"):
			// 项目请求的特殊验证
			validateProjectRequest(c)
		case strings.Contains(path, "/translations"):
			// 翻译请求的特殊验证
			validateTranslationRequest(c)
		}

		c.Next()
	}
}

// validateLoginRequest 验证登录请求
func validateLoginRequest(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		return
	}

	// 这里可以添加登录特定的验证逻辑
	// 例如：用户名格式、密码复杂度等
}

// validateProjectRequest 验证项目请求
func validateProjectRequest(c *gin.Context) {
	if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut {
		return
	}

	// 项目名称、描述等的特定验证
}

// validateTranslationRequest 验证翻译请求
func validateTranslationRequest(c *gin.Context) {
	if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut {
		return
	}

	// 翻译键名、值等的特定验证
}

// SecurityValidationMiddleware 安全验证中间件
func SecurityValidationMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查User-Agent
		userAgent := c.GetHeader("User-Agent")
		if userAgent == "" {
			logger.Warn("Missing User-Agent header",
				zap.String("ip", c.ClientIP()),
				zap.String("path", c.Request.URL.Path),
			)
		}

		// 检查可疑的请求头
		suspiciousHeaders := []string{
			"X-Forwarded-For",
			"X-Real-IP",
			"X-Originating-IP",
		}

		for _, header := range suspiciousHeaders {
			if value := c.GetHeader(header); value != "" {
				// 记录可能的代理或伪造IP
				logger.Info("Proxy header detected",
					zap.String("header", header),
					zap.String("value", log_utils.SanitizeLogValue(value)),
					zap.String("ip", c.ClientIP()),
					zap.String("path", c.Request.URL.Path),
				)
			}
		}

		c.Next()
	}
}
