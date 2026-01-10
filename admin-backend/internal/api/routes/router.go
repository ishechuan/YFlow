package routes

import (
	"yflow/internal/api/handlers"
	"yflow/internal/api/middleware"
	"yflow/internal/api/response"
	"yflow/internal/domain"
	internal_utils "yflow/internal/utils"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Router 路由器
type Router struct {
	UserHandler               *handlers.UserHandler
	ProjectHandler            *handlers.ProjectHandler
	LanguageHandler           *handlers.LanguageHandler
	TranslationHandler        *handlers.TranslationHandler
	TranslationHistoryHandler *handlers.TranslationHistoryHandler
	DashboardHandler          *handlers.DashboardHandler
	ProjectMemberHandler      *handlers.ProjectMemberHandler
	CLIHandler                *handlers.CLIHandler
	InvitationHandler         *handlers.InvitationHandler
	middlewareFactory         *middleware.MiddlewareFactory
	Logger                    *zap.Logger
}

// RouterDeps 定义 Router 的依赖（用于 fx.In）
type RouterDeps struct {
	fx.In
	UserHandler               *handlers.UserHandler
	ProjectHandler            *handlers.ProjectHandler
	LanguageHandler           *handlers.LanguageHandler
	TranslationHandler        *handlers.TranslationHandler
	TranslationHistoryHandler *handlers.TranslationHistoryHandler
	DashboardHandler          *handlers.DashboardHandler
	ProjectMemberHandler      *handlers.ProjectMemberHandler
	CLIHandler                *handlers.CLIHandler
	InvitationHandler         *handlers.InvitationHandler
	AuthService               domain.AuthService
	UserService               domain.UserService
	ProjectMemberService      domain.ProjectMemberService
	Logger                    *zap.Logger
}

// NewRouter 创建路由器
func NewRouter(deps RouterDeps) *Router {
	return &Router{
		UserHandler:               deps.UserHandler,
		ProjectHandler:            deps.ProjectHandler,
		LanguageHandler:           deps.LanguageHandler,
		TranslationHandler:        deps.TranslationHandler,
		TranslationHistoryHandler: deps.TranslationHistoryHandler,
		DashboardHandler:          deps.DashboardHandler,
		ProjectMemberHandler:      deps.ProjectMemberHandler,
		CLIHandler:                deps.CLIHandler,
		InvitationHandler:         deps.InvitationHandler,
		middlewareFactory: middleware.NewMiddlewareFactory(
			deps.AuthService,
			deps.UserService,
			deps.ProjectMemberService,
		),
		Logger: deps.Logger,
	}
}

// SetupRoutes 设置路由
func (r *Router) SetupRoutes(engine *gin.Engine, monitor *internal_utils.SimpleMonitor) {
	// 基本路由
	engine.GET("/", func(c *gin.Context) {
		response.Success(c, gin.H{"message": "Hello, World!"})
	})

	// 监控端点
	r.setupMonitoringRoutes(engine, monitor)

	// Swagger 文档
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API 路由组
	api := engine.Group("/api")
	{
		r.setupPublicRoutes(api)
		r.setupPublicInvitationRoutes(api)
		r.setupPublicRegisterRoutes(api)
		r.setupAuthenticatedRoutes(api)
		r.setupCLIRoutes(api)
	}
}

// setupAuthenticatedRoutes 设置需要认证的路由
func (r *Router) setupAuthenticatedRoutes(rg *gin.RouterGroup) {
	// 应用JWT认证中间件和API限流中间件
	authRoutes := rg.Group("")
	authRoutes.Use(r.middlewareFactory.JWTAuthMiddleware())
	authRoutes.Use(middleware.TollboothAPIRateLimitMiddleware())

	// 用户相关路由
	r.setupUserRoutes(authRoutes)

	// 项目相关路由
	r.setupProjectRoutes(authRoutes)

	// 语言相关路由
	r.setupLanguageRoutes(authRoutes)

	// 翻译相关路由
	r.setupTranslationRoutes(authRoutes)

	// 仪表板相关路由
	r.setupDashboardRoutes(authRoutes)

	// 邀请管理路由
	r.setupInvitationRoutes(authRoutes)
}

// RouterModule 定义路由模块
var RouterModule = fx.Module("router",
	fx.Provide(NewRouter),
)
