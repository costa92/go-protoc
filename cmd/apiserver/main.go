package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

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
	"google.golang.org/grpc"
)

func main() {
	// 加载配置
	configPath := getConfigPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志记录器
	if err := log.Init(cfg.Log); err != nil {
		log.Fatalf("初始化日志记录器失败: %v", err)
	}
	defer log.Sync()

	log.Infof("成功加载配置文件来自: %s", configPath)

	// 初始化 OpenTelemetry Tracer
	tp, err := tracing.InitTracer(&cfg.Observability.Tracing)
	if err != nil {
		log.Fatalf("初始化OpenTelemetry Tracer失败: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Errorf("关闭追踪器失败: %v", err)
		}
	}()

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

	defer apiServer.Stop()

	// 创建并安装 helloworld API 组
	helloworldInstaller := helloworld.NewInstaller()
	apiServer.InstallAPIGroup(helloworldInstaller)

	// 添加指标路由（如果启用）
	if cfg.Observability.Metrics.Enabled {
		log.Infof("启用Prometheus指标，路径: %s", cfg.Observability.Metrics.Path)
		apiServer.GetHTTPServer().Router().Handle(cfg.Observability.Metrics.Path, metrics.PrometheusHandler())
	}

	// 创建一个带取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Infof("接收到信号: %s", sig.String())
		cancel()
	}()

	// 启动服务器
	if err := apiServer.Start(ctx); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

// getConfigPath 返回配置文件路径
func getConfigPath() string {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = filepath.Join("configs", "config.yaml")
	}
	return configPath
}
