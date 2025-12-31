package di

import (
	"yflow/internal/api/handlers"
	"yflow/internal/api/routes"

	"go.uber.org/fx"
)

// AppModule 定义主模块
var AppModule = fx.Module("app",
	// 数据库和缓存
	fx.Provide(NewDB),
	fx.Provide(NewRedisClient),

	// 缓存服务
	fx.Provide(NewCacheService),

	// 监控器
	fx.Provide(NewSimpleMonitor),

	// Repositories
	fx.Provide(NewUserRepository),
	fx.Provide(NewProjectRepository),
	fx.Provide(NewLanguageRepository),
	fx.Provide(NewTranslationRepository),
	fx.Provide(NewProjectMemberRepository),
	fx.Provide(NewInvitationRepository),

	// Auth Service (无缓存)
	fx.Provide(NewAuthService),

	// Services (带缓存装饰器)
	fx.Provide(NewUserService),
	fx.Provide(NewProjectService),
	fx.Provide(NewLanguageService),
	fx.Provide(NewTranslationService),
	fx.Provide(NewDashboardService),
	fx.Provide(NewProjectMemberService),
	fx.Provide(NewInvitationService),

	// Handlers
	fx.Provide(handlers.NewUserHandler),
	fx.Provide(handlers.NewProjectHandler),
	fx.Provide(handlers.NewLanguageHandler),
	fx.Provide(handlers.NewTranslationHandler),
	fx.Provide(handlers.NewProjectMemberHandler),
	fx.Provide(handlers.NewCLIHandler),
	fx.Provide(handlers.NewDashboardHandler),
	fx.Provide(handlers.NewInvitationHandler),

	// Router
	fx.Provide(routes.NewRouter),

	// Logger
	fx.Provide(NewLogger),

	// DB Security Monitor
	fx.Provide(NewDBSecurityMonitor),
)
