package middleware

import (
	"yflow/internal/api/response"
	"yflow/internal/domain"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// RequireRole 要求用户具有指定角色
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取当前用户信息
		userRole, exists := ctx.Get("userRole")
		if !exists {
			response.Forbidden(ctx, "无法获取用户角色信息")
			ctx.Abort()
			return
		}

		// 检查角色权限
		if !hasRolePermission(userRole.(string), requiredRole) {
			response.Forbidden(ctx, "权限不足")
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// RequireAdminRole 要求管理员角色
func RequireAdminRole() gin.HandlerFunc {
	return RequireRole("admin")
}

// RequireProjectPermission 要求项目权限
func RequireProjectPermission(requiredRole string, projectMemberService domain.ProjectMemberService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取当前用户ID
		userID, exists := ctx.Get("userID")
		if !exists {
			response.Unauthorized(ctx, "用户未登录")
			ctx.Abort()
			return
		}

		// 获取当前用户角色
		userRole, exists := ctx.Get("userRole")
		if !exists {
			response.Forbidden(ctx, "无法获取用户角色信息")
			ctx.Abort()
			return
		}

		// 管理员拥有所有权限
		if userRole.(string) == "admin" {
			ctx.Next()
			return
		}

		// 获取项目ID
		projectIDStr := ctx.Param("project_id")
		if projectIDStr == "" {
			projectIDStr = ctx.Param("id") // 兼容不同的路由参数名
		}

		if projectIDStr == "" {
			response.ValidationError(ctx, "缺少项目ID参数")
			ctx.Abort()
			return
		}

		projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
		if err != nil {
			response.ValidationError(ctx, "无效的项目ID")
			ctx.Abort()
			return
		}

		// 检查项目权限
		hasPermission, err := projectMemberService.CheckPermission(
			ctx.Request.Context(),
			userID.(uint64),
			uint64(projectID),
			requiredRole,
		)
		if err != nil {
			response.InternalServerError(ctx, "权限检查失败")
			ctx.Abort()
			return
		}

		if !hasPermission {
			response.Forbidden(ctx, "项目权限不足")
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// RequireProjectOwner 要求项目所有者权限
func RequireProjectOwner(projectMemberService domain.ProjectMemberService) gin.HandlerFunc {
	return RequireProjectPermission("owner", projectMemberService)
}

// RequireProjectEditor 要求项目编辑权限
func RequireProjectEditor(projectMemberService domain.ProjectMemberService) gin.HandlerFunc {
	return RequireProjectPermission("editor", projectMemberService)
}

// RequireProjectViewer 要求项目查看权限
func RequireProjectViewer(projectMemberService domain.ProjectMemberService) gin.HandlerFunc {
	return RequireProjectPermission("viewer", projectMemberService)
}

// RequireSelfOrAdmin 要求是本人或管理员
func RequireSelfOrAdmin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取当前用户信息
		currentUserID, exists := ctx.Get("userID")
		if !exists {
			response.Unauthorized(ctx, "用户未登录")
			ctx.Abort()
			return
		}

		currentUserRole, exists := ctx.Get("userRole")
		if !exists {
			response.Forbidden(ctx, "无法获取用户角色信息")
			ctx.Abort()
			return
		}

		// 管理员拥有所有权限
		if currentUserRole.(string) == "admin" {
			ctx.Next()
			return
		}

		// 获取目标用户ID
		targetUserIDStr := ctx.Param("id")
		if targetUserIDStr == "" {
			targetUserIDStr = ctx.Param("user_id")
		}

		if targetUserIDStr == "" {
			response.ValidationError(ctx, "缺少用户ID参数")
			ctx.Abort()
			return
		}

		targetUserID, err := strconv.ParseUint(targetUserIDStr, 10, 32)
		if err != nil {
			response.ValidationError(ctx, "无效的用户ID")
			ctx.Abort()
			return
		}

		// 检查是否是本人
		if currentUserID.(uint64) != targetUserID {
			response.Forbidden(ctx, "只能操作自己的账户")
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// hasRolePermission 检查角色权限
func hasRolePermission(userRole, requiredRole string) bool {
	// 角色权限层级：admin > member > viewer
	roleLevel := map[string]int{
		"viewer": 1,
		"member": 2,
		"admin":  3,
	}

	userLevel, exists := roleLevel[strings.ToLower(userRole)]
	if !exists {
		return false
	}

	requiredLevel, exists := roleLevel[strings.ToLower(requiredRole)]
	if !exists {
		return false
	}

	return userLevel >= requiredLevel
}
