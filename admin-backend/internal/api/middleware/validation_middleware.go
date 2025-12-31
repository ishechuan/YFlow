package middleware

import (
	"fmt"
	"yflow/internal/api/response"
	"yflow/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequestValidationMiddleware 请求验证中间件
// 验证HTTP请求的基本格式和Content-Type
func RequestValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查Content-Type（对于POST、PUT请求）
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut {
			contentType := c.GetHeader("Content-Type")
			if contentType != "" && contentType != "application/json" && contentType != "multipart/form-data" {
				response.BadRequest(c, fmt.Sprintf("不支持的Content-Type: %s", contentType))
				return
			}
		}

		c.Next()
	}
}

// RequestSizeLimitMiddleware 请求大小限制中间件
// 限制请求体的最大大小（默认32MB）
func RequestSizeLimitMiddleware(maxSize int64) gin.HandlerFunc {
	if maxSize <= 0 {
		maxSize = 32 << 20 // 默认32MB
	}

	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			response.BadRequest(c, fmt.Sprintf("请求体过大，最大支持 %d bytes", maxSize))
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}

// PaginationValidationMiddleware 分页参数验证中间件
// 验证和规范化分页参数
func PaginationValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只对GET请求进行分页验证
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		page := c.Query("page")
		pageSize := c.Query("page_size")

		// 验证page参数 (允许范围: 1-10000)
		if page != "" {
			if !utils.IsValidInteger(page) {
				response.BadRequest(c, "page参数必须是有效的整数")
				return
			}
			pageNum := utils.ParseInt(page, 1)
			if pageNum <= 0 {
				response.BadRequest(c, "page参数必须大于0")
				return
			}
			if pageNum > 10000 {
				response.BadRequest(c, "page参数不能超过10000")
				return
			}
		}

		// 验证page_size参数 (允许范围: 1-100)
		if pageSize != "" {
			if !utils.IsValidInteger(pageSize) {
				response.BadRequest(c, "page_size参数必须是有效的整数")
				return
			}
			size := utils.ParseInt(pageSize, 10)
			if size <= 0 {
				response.BadRequest(c, "page_size参数必须大于0")
				return
			}
			if size > 100 {
				response.BadRequest(c, "page_size参数不能超过100")
				return
			}
		}

		c.Next()
	}
}
