package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/costa92/go-protoc/internal/helloworld"
	"github.com/costa92/go-protoc/pkg/app"
	"github.com/costa92/go-protoc/pkg/config"
	grpcmiddleware "github.com/costa92/go-protoc/pkg/middleware/grpc"
	httpmiddleware "github.com/costa92/go-protoc/pkg/middleware/http"
	"github.com/costa92/go-protoc/pkg/metrics"
	"github.com/costa92/go-protoc/pkg/tracing"

	// ↓↓↓ 确保以下两个包已导入 ↓↓↓

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	// 创建日志器
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 加载配置
	configPath := getConfigPath()
	logger.Info("加载配置文件", zap.String("path", configPath))
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.Fatal("加载配置失败", zap.Error(err))
	}

	// 初始化 OpenTelemetry Tracer
	tp, err := tracing.InitTracer(&cfg.Observability.Tracing)
	if err != nil {
		logger.Fatal("初始化OpenTelemetry Tracer失败", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Error("关闭追踪器失败", zap.Error(err))
		}
	}()

	// 为 OpenTelemetry HTTP 追踪创建一个中间件
	otelHTTPMiddleware := func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, "http-server")
	}

	// 创建 gRPC 统计处理器
	otelGrpcHandler := otelgrpc.NewServerHandler(otelgrpc.WithTracerProvider(tp))

	// HTTP超时配置
	httpTimeout := time.Duration(cfg.Server.HTTP.Timeout) * time.Second

	// 创建应用实例
	apiServer := app.NewApp(cfg.Server.HTTP.Addr, cfg.Server.GRPC.Addr, logger,
		// 添加 HTTP 中间件
		app.WithHTTPMiddlewares(
			otelHTTPMiddleware,
			httpmiddleware.LoggingMiddleware(logger),
			httpmiddleware.RecoveryMiddleware(logger),
			httpmiddleware.TimeoutMiddleware(httpTimeout),
			httpmiddleware.ValidationMiddleware(logger),
		),
		// 添加 gRPC 拦截器
		app.WithGRPCUnaryInterceptors(
			grpcmiddleware.UnaryLoggingInterceptor(logger),
			grpcmiddleware.UnaryRecoveryInterceptor(logger),
			grpcmiddleware.ValidationUnaryServerInterceptor(),
		),
		app.WithGRPCStreamInterceptors(
			grpcmiddleware.StreamLoggingInterceptor(logger),
			grpcmiddleware.StreamRecoveryInterceptor(logger),
			grpcmiddleware.ValidationStreamServerInterceptor(),
		),
		// 添加 gRPC 服务器选项 - 使用 StatsHandler 替代拦截器
		app.WithGRPCOptions(
			grpc.StatsHandler(otelGrpcHandler),
		),
	)

	defer apiServer.Stop()

	// 创建并安装 helloworld API 组
	helloworldInstaller := helloworld.NewInstaller(logger)
	if err := apiServer.InstallAPIGroup(helloworldInstaller); err != nil {
		logger.Fatal("安装helloworld API组失败", zap.Error(err))
	}

	// 添加指标路由（如果启用）
	if cfg.Observability.Metrics.Enabled {
		logger.Info("启用Prometheus指标", zap.String("path", cfg.Observability.Metrics.Path))
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
		logger.Info("接收到信号", zap.String("signal", sig.String()))
		cancel()
	}()

	// 启动服务器
	if err := apiServer.Start(ctx); err != nil {
		log.Fatal(err)
	}
}

// getConfigPath 返回配置文件路径
func getConfigPath() string {
	// 从环境变量获取配置路径，如果未设置则使用默认路径
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		// 默认使用当前目录下的configs/config.yaml
		configPath = filepath.Join("configs", "config.yaml")
	}
	return configPath
}
