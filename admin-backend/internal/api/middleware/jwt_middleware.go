package middleware

import (
	"yflow/internal/api/response"
	"yflow/internal/domain"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware JWT鉴权中间件
// 接受authService和userService作为参数，支持依赖注入
func JWTAuthMiddleware(authService domain.AuthService, userService domain.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Authorization头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "未提供Authorization头")
			return
		}

		// Bearer token格式检查
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.BadRequest(c, "Authorization格式错误，应为'Bearer token'")
			return
		}

		// 验证token
		tokenString := parts[1]
		user, err := authService.ValidateToken(c.Request.Context(), tokenString)
		if err != nil {
			if strings.Contains(err.Error(), "expired") {
				response.TokenExpired(c, "token已过期")
			} else {
				response.InvalidToken(c, "无效的token")
			}
			return
		}

		// 获取完整的用户信息以获取角色
		fullUser, err := userService.GetUserInfo(c.Request.Context(), user.ID)
		if err != nil {
			response.Unauthorized(c, "用户信息获取失败")
			return
		}

		// 将用户信息存储到上下文中
		c.Set("userID", fullUser.ID)
		c.Set("username", fullUser.Username)
		c.Set("userRole", fullUser.Role)
		c.Set("userStatus", fullUser.Status)

		// 检查用户状态
		if fullUser.Status != "active" {
			response.Forbidden(c, "用户账户已被禁用")
			return
		}

		c.Next()
	}
}
