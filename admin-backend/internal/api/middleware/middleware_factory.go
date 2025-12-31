package middleware

import (
	"yflow/internal/domain"

	"github.com/gin-gonic/gin"
)

// MiddlewareFactory 中间件工厂
// 负责管理需要依赖注入的中间件
type MiddlewareFactory struct {
	authService          domain.AuthService
	userService          domain.UserService
	projectMemberService domain.ProjectMemberService
}

// NewMiddlewareFactory 创建中间件工厂
func NewMiddlewareFactory(
	authService domain.AuthService,
	userService domain.UserService,
	projectMemberService domain.ProjectMemberService,
) *MiddlewareFactory {
	return &MiddlewareFactory{
		authService:          authService,
		userService:          userService,
		projectMemberService: projectMemberService,
	}
}

// JWTAuthMiddleware 返回配置好的JWT认证中间件
func (f *MiddlewareFactory) JWTAuthMiddleware() gin.HandlerFunc {
	return JWTAuthMiddleware(f.authService, f.userService)
}

// RequireAdminRole 返回要求管理员角色的中间件
func (f *MiddlewareFactory) RequireAdminRole() gin.HandlerFunc {
	return RequireAdminRole()
}

// RequireRole 返回要求指定角色的中间件
func (f *MiddlewareFactory) RequireRole(role string) gin.HandlerFunc {
	return RequireRole(role)
}

// RequireProjectOwner 返回要求项目所有者权限的中间件
func (f *MiddlewareFactory) RequireProjectOwner() gin.HandlerFunc {
	return RequireProjectOwner(f.projectMemberService)
}

// RequireProjectEditor 返回要求项目编辑权限的中间件
func (f *MiddlewareFactory) RequireProjectEditor() gin.HandlerFunc {
	return RequireProjectEditor(f.projectMemberService)
}

// RequireProjectViewer 返回要求项目查看权限的中间件
func (f *MiddlewareFactory) RequireProjectViewer() gin.HandlerFunc {
	return RequireProjectViewer(f.projectMemberService)
}

// RequireSelfOrAdmin 返回要求是本人或管理员的中间件
func (f *MiddlewareFactory) RequireSelfOrAdmin() gin.HandlerFunc {
	return RequireSelfOrAdmin()
}
