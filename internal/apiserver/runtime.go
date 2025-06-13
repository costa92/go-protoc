// internal/apiserver/apiserver.go

package apiserver

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	apiserver_options "github.com/costa92/go-protoc/internal/apiserver/options"
	"github.com/costa92/go-protoc/pkg/app"
	"github.com/costa92/go-protoc/pkg/logger"
	"github.com/costa92/go-protoc/pkg/metrics"
	"github.com/costa92/go-protoc/pkg/middleware"
	pkg_options "github.com/costa92/go-protoc/pkg/options"
	"github.com/costa92/go-protoc/pkg/telemetry"
)

// RunFunc 定义运行 API 服务器的函数类型
type RunFunc func(opts *apiserver_options.Options) error

// APIServer 封装 API 服务器的运行时
type APIServer struct {
	name              string
	logger            logger.Logger
	apiInstaller      app.APIGroupInstaller // 1. 新增此行: 接收 API 安装器
	middlewareManager *middleware.Manager   // 新增: 中间件管理器
}

// NewAPIServer 创建一个新的 API 服务器实例
func NewAPIServer(
	name string,
	logger logger.Logger,
	apiInstaller app.APIGroupInstaller, // 2. 新增此参数: 注入 API 安装器
) *APIServer {
	return &APIServer{
		name:         name,
		logger:       logger,
		apiInstaller: apiInstaller, // 3. 新增此行: 赋值
	}
}

// Run 运行 API 服务器
func (s *APIServer) Run(opts *apiserver_options.Options) error {
	// ... (从 "完成选项配置" 到 "添加 metrics 路由" 的代码保持不变) ...
	if err := opts.Complete(); err != nil {
		return err
	}

	s.logger.Infow("服务器配置信息",
		"name", s.name,
		"grpc_addr", opts.GetGRPCOptions().Addr,
		"http_addr", opts.GetHTTPOptions().Addr,
		"log_level", opts.GetLogOptions().Level,
		"metrics_enabled", opts.Metrics.Enabled,
		"tracing_enabled", opts.Tracing.Enabled,
		"middleware_timeout", opts.Middleware.Timeout,
		"rate_limit_enabled", opts.Middleware.RateLimit.Enable,
		"rate_limit_limit", opts.Middleware.RateLimit.Limit,
	)

	// ===================== 新增: 初始化中间件系统 =====================
	s.logger.Infow("初始化可插拔中间件系统...")

	// 从配置中构建中间件配置
	middlewareConfig := s.buildMiddlewareConfig(opts)

	// 初始化中间件管理器
	var err error
	s.middlewareManager, err = middleware.InitializeMiddleware(middlewareConfig)
	if err != nil {
		s.logger.Errorw("初始化中间件管理器失败", "error", err)
		return err
	}

	s.logger.Infow("中间件系统初始化完成")
	// ================================================================

	grpcOpts := opts.GetGRPCOptions()
	httpOpts := opts.GetHTTPOptions()

	grpcListener, err := net.Listen("tcp", grpcOpts.Addr)
	if err != nil {
		s.logger.Errorw("创建 gRPC 监听器失败", "error", err)
		return err
	}

	grpcServer := app.NewGRPCServer(s.name, grpcListener)
	httpServer := app.NewHTTPServer(s.name, httpOpts.Addr)

	// ===================== 新增: 应用中间件链 =====================
	s.logger.Infow("应用中间件链到服务器...")

	// 将中间件链应用到 HTTP 服务器
	if err := s.middlewareManager.ApplyToHTTPServer(httpServer); err != nil {
		s.logger.Errorw("应用 HTTP 中间件失败", "error", err)
		return err
	}

	// 将中间件链应用到 gRPC 服务器
	if err := s.middlewareManager.ApplyToGRPCServer(grpcServer); err != nil {
		s.logger.Errorw("应用 gRPC 中间件失败", "error", err)
		return err
	}

	s.logger.Infow("中间件链应用完成")
	// ================================================================

	if opts.Tracing.Enabled {
		s.logger.Infow("启用 tracing 功能",
			"service_name", opts.Tracing.ServiceName,
			"otlp_endpoint", opts.Tracing.OTLPEndpoint)
		grpcServer.AddUnaryServerInterceptors(telemetry.UnaryServerInterceptor())
		httpServer.AddMiddleware(telemetry.TracingMiddleware)
	}

	if opts.Metrics.Enabled {
		httpServer.AddRoute(opts.Metrics.Path, metrics.PrometheusHandler().ServeHTTP)
	}

	// ===================== 核心修改点 =====================
	s.logger.Infow("开始自动安装 API 组...")

	// 4. 注册所有通过 wire 注入的 API 安装器
	//    我们在这里只注册 greeter，未来有新服务时，可以在 wire.go 中为 NewAPIServer 增加参数，并在此处注册
	app.RegisterAPIGroup(s.apiInstaller)

	// 5. 从注册表中获取所有 API 安装器，并循环执行它们的 Install 方法
	for _, installer := range app.GetAPIGroups() {
		if err := installer.Install(grpcServer, httpServer); err != nil {
			s.logger.Errorw("安装 API 组失败", "installer", installer, "error", err)
			return err
		}
	}
	// =======================================================

	// 确保路由最终化
	httpServer.FinalizeRoutes()

	// ... (从 "创建上下文用于优雅关闭" 到文件结尾的代码保持不变) ...
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 2)

	go func() {
		s.logger.Infow("启动 gRPC 服务器", "address", grpcOpts.Addr)
		if err := grpcServer.Start(ctx); err != nil {
			errCh <- err
		}
	}()

	go func() {
		s.logger.Infow("启动 HTTP 服务器", "address", httpOpts.Addr)
		if err := httpServer.Start(ctx); err != nil {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		s.logger.Infow("收到信号，开始优雅关闭", "signal", sig)
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), grpcOpts.Timeout)
		defer shutdownCancel()

		// ===================== 新增: 关闭中间件管理器 =====================
		if s.middlewareManager != nil {
			if err := s.middlewareManager.Shutdown(shutdownCtx); err != nil {
				s.logger.Errorw("关闭中间件管理器失败", "error", err)
			}
		}
		// ================================================================

		if err := grpcServer.Stop(shutdownCtx); err != nil {
			s.logger.Errorw("关闭 gRPC 服务器失败", "error", err)
		}
		if err := httpServer.Stop(shutdownCtx); err != nil {
			s.logger.Errorw("关闭 HTTP 服务器失败", "error", err)
		}

		if opts.Tracing.Enabled && pkg_options.TracerShutdownFunc != nil {
			s.logger.Infow("关闭 tracer")
			timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := pkg_options.TracerShutdownFunc(timeoutCtx); err != nil {
				s.logger.Errorw("关闭 tracer 失败", "error", err)
			}
		}

		s.logger.Infow("服务器已关闭")
		return nil

	case err := <-errCh:
		s.logger.Errorw("服务器发生错误", "error", err)
		return err
	}
}

// buildMiddlewareConfig 根据配置选项构建中间件配置
func (s *APIServer) buildMiddlewareConfig(opts *apiserver_options.Options) map[string]map[string]interface{} {
	config := make(map[string]map[string]interface{})

	// 基础配置
	config["recovery"] = map[string]interface{}{
		"enabled":  true,
		"priority": 10,
	}

	config["logging"] = map[string]interface{}{
		"enabled":    true,
		"priority":   100,
		"skip_paths": []interface{}{"/health", "/metrics"},
	}

	// CORS 配置 - 根据需要启用
	config["cors"] = map[string]interface{}{
		"enabled":       false, // 可以根据配置文件决定是否启用
		"priority":      50,
		"allow_origins": []interface{}{"*"},
		"allow_methods": []interface{}{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		"allow_headers": []interface{}{"Content-Type", "Authorization"},
	}

	// 可以根据 opts.Middleware 配置添加更多的中间件配置
	// 比如限流、认证等

	return config
}
