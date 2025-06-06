package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/costa92/go-protoc/internal/helloworld"
	"github.com/costa92/go-protoc/pkg/app"
	grpcmiddleware "github.com/costa92/go-protoc/pkg/middleware/grpc"
	httpmiddleware "github.com/costa92/go-protoc/pkg/middleware/http"
	"go.uber.org/zap"
)

func main() {
	// 创建日志器
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 创建应用实例
	apiServer := app.NewApp(":8090", ":8091", logger,
		// 添加 HTTP 中间件
		app.WithHTTPMiddlewares(
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
