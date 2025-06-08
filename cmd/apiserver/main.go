package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/costa92/go-protoc/internal/helloworld"
	"github.com/costa92/go-protoc/pkg/app"
	grpcmiddleware "github.com/costa92/go-protoc/pkg/middleware/grpc"
	httpmiddleware "github.com/costa92/go-protoc/pkg/middleware/http"
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

	// 初始化 OpenTelemetry Tracer
	shutdown, err := tracing.InitTracer("go-protoc-service") //
	if err != nil {
		logger.Fatal("failed to initialize OpenTelemetry Tracer", zap.Error(err))
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			logger.Error("failed to shutdown tracer", zap.Error(err))
		}
	}()

	// 为 OpenTelemetry HTTP 追踪创建一个中间件
	otelHTTPMiddleware := func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, "http-server")
	}

	// 创建 gRPC 统计处理器
	otelGrpcHandler := otelgrpc.NewServerHandler()

	// 创建应用实例
	apiServer := app.NewApp(":8090", ":8091", logger,
		// 添加 HTTP 中间件
		app.WithHTTPMiddlewares(
			otelHTTPMiddleware,
			httpmiddleware.LoggingMiddleware(logger),
			httpmiddleware.RecoveryMiddleware(logger),
			httpmiddleware.TimeoutMiddleware(5*time.Second),
		),
		// 添加 gRPC 拦截器
		app.WithGRPCUnaryInterceptors(
			grpcmiddleware.UnaryLoggingInterceptor(logger),
			grpcmiddleware.UnaryRecoveryInterceptor(logger),
		),
		app.WithGRPCStreamInterceptors(
			grpcmiddleware.StreamLoggingInterceptor(logger),
			grpcmiddleware.StreamRecoveryInterceptor(logger),
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
		logger.Fatal("failed to install helloworld API group", zap.Error(err))
	}

	// 创建一个带取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Info("received signal", zap.String("signal", sig.String()))
		cancel()
	}()

	// 启动服务器
	if err := apiServer.Start(ctx); err != nil {
		log.Fatal(err)
	}
}
