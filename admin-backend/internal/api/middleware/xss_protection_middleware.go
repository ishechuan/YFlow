package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"yflow/internal/api/response"
	log_utils "yflow/utils"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"go.uber.org/zap"
)

// XSSProtectionConfig XSS防护配置
type XSSProtectionConfig struct {
	EnableStrictMode   bool     // 是否启用严格模式（移除所有HTML）
	AllowedTags        []string // 允许的HTML标签
	AllowedAttributes  []string // 允许的HTML属性
	MaxContentLength   int      // 最大内容长度
	SanitizeResponse   bool     // 是否清理响应内容
	LogSuspiciousInput bool     // 是否记录可疑输入
}

// DefaultXSSProtectionConfig 默认XSS防护配置
func DefaultXSSProtectionConfig() XSSProtectionConfig {
	return XSSProtectionConfig{
		EnableStrictMode:   false,
		AllowedTags:        []string{"p", "br", "strong", "em", "u", "i", "b"},
		AllowedAttributes:  []string{"class", "id"},
		MaxContentLength:   50000, // 50KB
		SanitizeResponse:   false,
		LogSuspiciousInput: true,
	}
}

// XSSProtectionMiddleware XSS防护中间件
func XSSProtectionMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return XSSProtectionMiddlewareWithConfig(logger, DefaultXSSProtectionConfig())
}

// XSSProtectionMiddlewareWithConfig 带配置的XSS防护中间件
func XSSProtectionMiddlewareWithConfig(logger *zap.Logger, config XSSProtectionConfig) gin.HandlerFunc {
	// 创建HTML清理策略
	var policy *bluemonday.Policy
	if config.EnableStrictMode {
		policy = bluemonday.StrictPolicy() // 移除所有HTML
	} else {
		policy = bluemonday.NewPolicy()
		// 添加允许的标签
		for _, tag := range config.AllowedTags {
			policy.AllowElements(tag)
		}
		// 添加允许的属性
		if len(config.AllowedAttributes) > 0 {
			policy.AllowStandardAttributes()
			for _, attr := range config.AllowedAttributes {
				policy.AllowAttrs(attr).Globally()
			}
		}
	}

	// 编译XSS检测正则表达式
	xssPatterns := compileXSSPatterns()

	return func(c *gin.Context) {
		// 跳过非内容请求
		if c.Request.Method == http.MethodGet ||
			c.Request.Method == http.MethodDelete ||
			c.Request.Method == http.MethodHead ||
			c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		// 检查Content-Type
		contentType := c.GetHeader("Content-Type")
		if !strings.Contains(contentType, "application/json") &&
			!strings.Contains(contentType, "multipart/form-data") &&
			!strings.Contains(contentType, "application/x-www-form-urlencoded") {
			c.Next()
			return
		}

		// 处理JSON请求
		if strings.Contains(contentType, "application/json") {
			if err := processJSONRequest(c, policy, xssPatterns, config, logger); err != nil {
				response.BadRequest(c, fmt.Sprintf("XSS防护检查失败: %s", err.Error()))
				return
			}
		}

		c.Next()
	}
}

// processJSONRequest 处理JSON请求
func processJSONRequest(c *gin.Context, policy *bluemonday.Policy, xssPatterns []*regexp.Regexp, config XSSProtectionConfig, logger *zap.Logger) error {
	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return fmt.Errorf("无法读取请求体")
	}

	// 恢复请求体
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	// 检查内容长度
	if len(body) > config.MaxContentLength {
		return fmt.Errorf("请求内容过大")
	}

	// 解析JSON
	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		// 不是有效的JSON，跳过处理
		return nil
	}

	// 检测和清理XSS
	cleanedData, hasXSS, err := sanitizeJSONData(jsonData, policy, xssPatterns, config)
	if err != nil {
		return err
	}

	// 记录XSS尝试
	if hasXSS && config.LogSuspiciousInput {
		logger.Error("XSS attempt detected",
			zap.String("ip", c.ClientIP()),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("user_agent", log_utils.SanitizeLogValue(c.GetHeader("User-Agent"))),
		)
	}

	// 更新请求体
	cleanedBody, err := json.Marshal(cleanedData)
	if err != nil {
		return fmt.Errorf("数据序列化失败")
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(cleanedBody))
	c.Request.ContentLength = int64(len(cleanedBody))

	return nil
}

// sanitizeJSONData 递归清理JSON数据
func sanitizeJSONData(data interface{}, policy *bluemonday.Policy, xssPatterns []*regexp.Regexp, config XSSProtectionConfig) (interface{}, bool, error) {
	hasXSS := false

	switch v := data.(type) {
	case string:
		cleaned, xssDetected := sanitizeString(v, policy, xssPatterns)
		return cleaned, xssDetected, nil
	case map[string]interface{}:
		cleaned := make(map[string]interface{})
		for key, value := range v {
			// 清理键名
			cleanKey, keyXSS := sanitizeString(key, policy, xssPatterns)
			if keyXSS {
				hasXSS = true
			}

			// 递归清理值
			cleanValue, valueXSS, err := sanitizeJSONData(value, policy, xssPatterns, config)
			if err != nil {
				return nil, false, err
			}
			if valueXSS {
				hasXSS = true
			}

			cleaned[cleanKey] = cleanValue
		}
		return cleaned, hasXSS, nil
	case []interface{}:
		cleaned := make([]interface{}, len(v))
		for i, item := range v {
			cleanItem, itemXSS, err := sanitizeJSONData(item, policy, xssPatterns, config)
			if err != nil {
				return nil, false, err
			}
			if itemXSS {
				hasXSS = true
			}
			cleaned[i] = cleanItem
		}
		return cleaned, hasXSS, nil
	default:
		// 数字、布尔值等其他类型直接返回
		return data, false, nil
	}
}

// sanitizeString 清理字符串
func sanitizeString(input string, policy *bluemonday.Policy, xssPatterns []*regexp.Regexp) (string, bool) {
	hasXSS := false

	// 检测XSS模式
	for _, pattern := range xssPatterns {
		if pattern.MatchString(strings.ToLower(input)) {
			hasXSS = true
			break
		}
	}

	// HTML清理
	cleaned := policy.Sanitize(input)

	// 如果清理后内容发生变化，说明包含了HTML内容
	if cleaned != input {
		hasXSS = true
	}

	return cleaned, hasXSS
}

// compileXSSPatterns 编译XSS检测模式
func compileXSSPatterns() []*regexp.Regexp {
	patterns := []string{
		// Script标签
		`<script[^>]*>.*?</script>`,
		`<script[^>]*>`,

		// 事件处理器
		`on\w+\s*=`,
		`javascript:`,
		`vbscript:`,

		// 危险标签
		`<iframe[^>]*>`,
		`<object[^>]*>`,
		`<embed[^>]*>`,
		`<applet[^>]*>`,
		`<meta[^>]*>`,
		`<link[^>]*>`,
		`<style[^>]*>`,

		// CSS表达式
		`expression\s*\(`,
		`url\s*\(.*javascript:`,
		`@import`,

		// 数据URI
		`data:\s*text/html`,
		`data:\s*application/javascript`,

		// 其他危险模式
		`<\s*\w+[^>]*\s+src\s*=\s*["']?\s*javascript:`,
		`<\s*\w+[^>]*\s+href\s*=\s*["']?\s*javascript:`,
	}

	var compiledPatterns []*regexp.Regexp
	for _, pattern := range patterns {
		if re, err := regexp.Compile("(?i)" + pattern); err == nil {
			compiledPatterns = append(compiledPatterns, re)
		}
	}

	return compiledPatterns
}

// ResponseXSSProtectionMiddleware 响应XSS防护中间件
func ResponseXSSProtectionMiddleware() gin.HandlerFunc {
	policy := bluemonday.UGCPolicy()

	return func(c *gin.Context) {
		// 创建自定义ResponseWriter
		writer := &xssResponseWriter{
			ResponseWriter: c.Writer,
			policy:         policy,
		}
		c.Writer = writer

		c.Next()
	}
}

// xssResponseWriter 自定义ResponseWriter，用于清理响应内容
type xssResponseWriter struct {
	gin.ResponseWriter
	policy *bluemonday.Policy
}

func (w *xssResponseWriter) Write(data []byte) (int, error) {
	// 只处理HTML/JSON响应
	contentType := w.Header().Get("Content-Type")
	if strings.Contains(contentType, "text/html") {
		// 清理HTML内容
		cleaned := w.policy.SanitizeBytes(data)
		return w.ResponseWriter.Write(cleaned)
	} else if strings.Contains(contentType, "application/json") {
		// 对JSON响应进行基础清理
		cleaned := html.EscapeString(string(data))
		return w.ResponseWriter.Write([]byte(cleaned))
	}

	// 其他类型直接写入
	return w.ResponseWriter.Write(data)
}

// HTMLEscapeMiddleware HTML转义中间件（轻量级选项）
func HTMLEscapeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过非内容请求
		if c.Request.Method == http.MethodGet ||
			c.Request.Method == http.MethodDelete {
			c.Next()
			return
		}

		// 简单的HTML转义处理
		if strings.Contains(c.GetHeader("Content-Type"), "application/json") {
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.Next()
				return
			}

			// 对常见的XSS字符进行转义
			escaped := html.EscapeString(string(body))
			c.Request.Body = io.NopCloser(strings.NewReader(escaped))
			c.Request.ContentLength = int64(len(escaped))
		}

		c.Next()
	}
}

// CSPViolationReportMiddleware CSP违规报告中间件
func CSPViolationReportMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 处理CSP违规报告
		if c.Request.URL.Path == "/csp-report" && c.Request.Method == http.MethodPost {
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无法读取报告"})
				return
			}

			// 记录CSP违规
			logger.Warn("CSP violation report",
				zap.String("ip", c.ClientIP()),
				zap.String("user_agent", log_utils.SanitizeLogValue(c.GetHeader("User-Agent"))),
				zap.String("report", log_utils.SanitizeLogValue(string(body))),
			)

			c.JSON(http.StatusOK, gin.H{"status": "received"})
			return
		}

		c.Next()
	}
}
