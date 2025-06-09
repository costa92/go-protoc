package apiserver

import (
	"context"
	"net/http"
	"os"
	"path/filepath"

	"github.com/costa92/go-protoc/internal/helloworld"
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
)

// Server represents the API server
type Server struct {
	app *app.App
	tp  *sdktrace.TracerProvider
}

// NewServer creates a new API server instance
func NewServer(configPath string) (*Server, error) {
	// 加载配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	// 初始化日志记录器
	if initErr := log.Init(cfg.Log); initErr != nil {
		return nil, initErr
	}

	log.Infow("成功加载配置文件来自", "path", configPath)

	// 初始化 OpenTelemetry Tracer
	tp, err := tracing.InitTracer(&cfg.Observability.Tracing)
	if err != nil {
		return nil, err
	}

	// 为 OpenTelemetry HTTP 追踪创建一个中间件
	otelHTTPMiddleware := func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, "http-server")
	}

	// 创建 gRPC 统计处理器
	otelGrpcHandler := otelgrpc.NewServerHandler(otelgrpc.WithTracerProvider(tp))

	// 创建应用实例
	apiServer := app.NewApp(
		cfg.Server.HTTP.Addr,
		cfg.Server.GRPC.Addr,
		// 添加 HTTP 中间件
		app.WithHTTPMiddlewares(
			otelHTTPMiddleware,
			httpmiddleware.LoggingMiddleware(),
			httpmiddleware.RecoveryMiddleware(),
			httpmiddleware.TimeoutMiddleware(cfg),
			httpmiddleware.CORSMiddleware(cfg),
			httpmiddleware.RateLimitMiddleware(cfg),
			httpmiddleware.ValidationMiddleware(),
		),
		// 添加 gRPC 拦截器
		app.WithGRPCUnaryInterceptors(
			grpcmiddleware.UnaryLoggingInterceptor(),
			grpcmiddleware.UnaryRecoveryInterceptor(),
			grpcmiddleware.ValidationUnaryServerInterceptor(),
		),
		app.WithGRPCStreamInterceptors(
			grpcmiddleware.StreamLoggingInterceptor(),
			grpcmiddleware.StreamRecoveryInterceptor(),
			grpcmiddleware.ValidationStreamServerInterceptor(),
		),
		// 添加 gRPC 服务器选项 - 使用 StatsHandler 替代拦截器
		app.WithGRPCOptions(
			grpc.StatsHandler(otelGrpcHandler),
		),
	)

	// 创建并安装 helloworld API 组
	helloworldInstaller := helloworld.NewInstaller()
	apiServer.InstallAPIGroup(helloworldInstaller)

	// 添加指标路由（如果启用）
	if cfg.Observability.Metrics.Enabled {
		log.Infof("启用Prometheus指标，路径: %s", cfg.Observability.Metrics.Path)
		apiServer.GetHTTPServer().Router().Handle(cfg.Observability.Metrics.Path, metrics.PrometheusHandler())
	}

	return &Server{
		app: apiServer,
		tp:  tp,
	}, nil
}

// Start starts the server
func (s *Server) Start(ctx context.Context) error {
	return s.app.Start(ctx)
}

// Stop stops the server
func (s *Server) Stop() {
	if s.tp != nil {
		if err := s.tp.Shutdown(context.Background()); err != nil {
			log.Errorf("关闭追踪器失败: %v", err)
		}
	}
	if err := s.app.Stop(); err != nil {
		log.Errorf("关闭应用失败: %v", err)
	}
	log.Sync() // 忽略错误，因为这是在关闭时的清理操作
}

// GetConfigPath returns the configuration file path
func GetConfigPath() string {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = filepath.Join("configs", "config.yaml")
	}
	return configPath
}
