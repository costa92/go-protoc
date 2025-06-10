package app

import (
	"context"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/costa92/go-protoc/pkg/app"
	"github.com/costa92/go-protoc/pkg/config"
	"github.com/costa92/go-protoc/pkg/log"
	"github.com/costa92/go-protoc/pkg/metrics"
	grpcmiddleware "github.com/costa92/go-protoc/pkg/middleware/grpc"
	httpmiddleware "github.com/costa92/go-protoc/pkg/middleware/http"
	"github.com/costa92/go-protoc/pkg/tracing"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server 表示 API 服务器
type Server struct {
	app *app.App
	tp  *sdktrace.TracerProvider
}

// NewServer 创建一个新的 API 服务器实例
func NewServer(name string, cfg *config.Config, opts ...Option) (*Server, error) {
	apiServer := &Server{}
	// Apply all the options
	for _, opt := range opts {
		opt(apiServer)
	}
	// 初始化 OpenTelemetry Tracer
	tp, err := tracing.InitTracer(&cfg.Observability.Tracing)
	if err != nil {
		return nil, err
	}

	// 创建 HTTP 服务器
	httpServer := createHTTPServer(cfg)

	// 创建 gRPC 服务器
	grpcServer, err := createGRPCServer(cfg, tp)
	if err != nil {
		return nil, err
	}

	// 创建应用实例，管理所有服务
	application := app.NewApp(name, httpServer, grpcServer)

	// 安装所有已注册的 API 组
	if err := installAPIGroups(grpcServer, httpServer); err != nil {
		return nil, err
	}

	// 添加指标路由（如果启用）
	if cfg.Observability.Metrics.Enabled {
		log.Infof("启用 Prometheus 指标，路径: %s", cfg.Observability.Metrics.Path)
		httpServer.AddRoute(cfg.Observability.Metrics.Path, metrics.PrometheusHandler().ServeHTTP)
	}

	return &Server{
		app: application,
		tp:  tp,
	}, nil
}

// createHTTPServer 创建和配置 HTTP 服务器
func createHTTPServer(cfg *config.Config) *app.HTTPServer {
	// 为 OpenTelemetry HTTP 追踪创建一个中间件
	otelHTTPMiddleware := func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, "http-server")
	}

	// 创建带中间件的 HTTP 服务器
	return app.NewHTTPServer(
		"api-http",
		cfg.Server.HTTP.Addr,
		otelHTTPMiddleware,
		httpmiddleware.LoggingMiddleware(
			cfg.Observability.SkipPaths,
		),
		httpmiddleware.RecoveryMiddleware(),
		httpmiddleware.TimeoutMiddleware(time.Duration(cfg.Middleware.Timeout)),
		httpmiddleware.CORSMiddleware(
			cfg.Middleware.CORS.AllowOrigins,
			cfg.Middleware.CORS.AllowMethods,
			cfg.Middleware.CORS.AllowHeaders,
			cfg.Middleware.CORS.ExposeHeaders,
			cfg.Middleware.CORS.AllowCredentials,
			cfg.Middleware.CORS.MaxAge,
		),
		httpmiddleware.RateLimitMiddleware(
			cfg.Middleware.RateLimit.Enable,
			float64(cfg.Middleware.RateLimit.Limit),
			cfg.Middleware.RateLimit.Burst,
			cfg.Observability.SkipPaths,
		),
		httpmiddleware.ValidationMiddleware(),
	)
}

// createGRPCServer 创建和配置 gRPC 服务器
func createGRPCServer(cfg *config.Config, tp *sdktrace.TracerProvider) (*app.GRPCServer, error) {
	// 创建 gRPC 监听器
	lis, err := net.Listen("tcp", cfg.Server.GRPC.Addr)
	if err != nil {
		return nil, err
	}

	// 创建 gRPC 统计处理器
	otelGrpcHandler := otelgrpc.NewServerHandler(otelgrpc.WithTracerProvider(tp))

	// 创建带拦截器的 gRPC 服务器
	return app.NewGRPCServer(
		"api-grpc",
		lis,
		grpc.ChainUnaryInterceptor(
			grpcmiddleware.UnaryLoggingInterceptor(),
			grpcmiddleware.UnaryRecoveryInterceptor(),
			grpcmiddleware.ValidationUnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			grpcmiddleware.StreamLoggingInterceptor(),
			grpcmiddleware.StreamRecoveryInterceptor(),
			grpcmiddleware.ValidationStreamServerInterceptor(),
		),
		grpc.StatsHandler(otelGrpcHandler),
	), nil
}

// installAPIGroups 安装所有已注册的 API 组
func installAPIGroups(grpcServer *app.GRPCServer, httpServer *app.HTTPServer) error {
	log.Infow("开始安装 API 组")
	// 从注册表获取所有 API 组并安装
	installers := app.GetAPIGroups()
	log.Infow("找到 %d 个 API 组安装器", "installers_len", len(installers))

	for i, installer := range installers {
		log.Infof("安装 API 组 %d: %T", i, installer)
		if err := installer.Install(grpcServer, httpServer); err != nil {
			log.Errorf("安装 API 组 %d 失败: %v", i, err)
			return err
		}
		log.Infow("成功安装 API 组", "index", i)
	}

	// 注册 gRPC 反射服务，使 grpcurl 等工具可以自省 API
	reflection.Register(grpcServer.Server())
	log.Infow("已注册 gRPC 反射服务")

	log.Infow("所有 API 组安装完成")
	return nil
}

// Start 启动服务器
func (s *Server) Start(ctx context.Context) error {
	// 创建一个带超时的上下文，用于优雅关闭
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 设置关闭处理
	go func() {
		<-ctx.Done() // 等待取消信号
		log.Infof("接收到关闭信号，开始优雅关闭...")
		if err := s.app.Stop(shutdownCtx); err != nil {
			log.Errorf("关闭服务器时发生错误: %v", err)
		}
	}()

	// 启动应用
	return s.app.Start(ctx)
}

// Stop 停止服务器
func (s *Server) Stop(ctx context.Context) error {
	log.Infof("开始关闭服务器...")
	if s.tp != nil {
		if err := s.tp.Shutdown(ctx); err != nil {
			log.Errorf("关闭追踪器失败: %v", err)
			return err
		}
	}
	log.Sync() // 忽略错误，因为这是在关闭时的清理操作
	return nil
}

// GetConfigPath 返回配置文件路径
func GetConfigPath() string {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = filepath.Join("configs", "config.yaml")
	}
	return configPath
}
