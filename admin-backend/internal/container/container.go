package container

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"yflow/internal/api/routes"
	"yflow/internal/config"
	"yflow/internal/di"
	internal_utils "yflow/internal/utils"

	"github.com/gin-gonic/gin"
)

// ServerParams 服务器启动所需的依赖
type ServerParams struct {
	fx.In

	Config          *config.Config
	Logger          *zap.Logger
	Router          *routes.Router
	Monitor         *internal_utils.SimpleMonitor
	LoggerSync      func()                                                        `name:"logger-sync"`
	SetupMiddleware func(*gin.Engine, *internal_utils.SimpleMonitor, *zap.Logger) `optional:"true"`
}

// RunServer 创建并运行 HTTP 服务器（FX 生命周期管理）
func RunServer(lc fx.Lifecycle, params ServerParams) {
	// 创建 Gin 引擎
	engine := gin.New()

	// 设置中间件（如果提供了自定义设置函数则使用，否则跳过）
	if params.SetupMiddleware != nil {
		params.SetupMiddleware(engine, params.Monitor, params.Logger)
	}

	// 设置路由
	params.Router.SetupRoutes(engine, params.Monitor)

	// 创建 HTTP 服务器
	server := &http.Server{
		Addr:    ":8080",
		Handler: engine,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			params.Logger.Info("Server starting",
				zap.String("version", "1.0.0"),
				zap.String("environment", params.Config.Env),
				zap.String("address", ":8080"),
				zap.String("docs", "http://localhost:8080/swagger/index.html"),
			)

			// 在 goroutine 中启动服务器，避免阻塞 FX
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					params.Logger.Error("Server failed", zap.Error(err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			params.Logger.Info("Server shutting down...")

			// 优雅关闭服务器
			shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			if err := server.Shutdown(shutdownCtx); err != nil {
				params.Logger.Error("Server shutdown error", zap.Error(err))
				return err
			}

			// 同步日志缓冲区
			if params.LoggerSync != nil {
				params.LoggerSync()
			}

			params.Logger.Info("Server stopped gracefully")
			return nil
		},
	})
}

// MiddlewareSetupFunc 中间件设置函数类型
type MiddlewareSetupFunc func(*gin.Engine, *internal_utils.SimpleMonitor, *zap.Logger)

// NewApp 创建 FX 应用（符合 FX 最佳实践）
func NewApp(cfg *config.Config, setupMiddleware MiddlewareSetupFunc) *fx.App {
	return fx.New(
		fx.NopLogger, // 暂时注释掉以查看错误日志

		// 通过 fx.Supply 提供配置
		fx.Supply(cfg),

		// 提供中间件设置函数（可选）
		fx.Provide(func() MiddlewareSetupFunc {
			return setupMiddleware
		}),

		// 转换为 ServerParams 需要的类型
		fx.Provide(func(fn MiddlewareSetupFunc) func(*gin.Engine, *internal_utils.SimpleMonitor, *zap.Logger) {
			return fn
		}),

		// 应用核心模块
		di.AppModule,

		// 服务器生命周期管理
		fx.Invoke(RunServer),
	)
}

// Run 运行应用（阻塞直到收到停止信号）
func Run(cfg *config.Config, setupMiddleware MiddlewareSetupFunc) {
	app := NewApp(cfg, setupMiddleware)
	app.Run()
}
