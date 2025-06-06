package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/costa92/go-protoc/internal/helloworld"
	grpcmiddleware "github.com/costa92/go-protoc/pkg/middleware/grpc"
	httpmiddleware "github.com/costa92/go-protoc/pkg/middleware/http"
	"github.com/costa92/go-protoc/pkg/server"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	// 创建日志器
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 创建服务器配置
	config := &server.Config{
		HTTPAddr: ":8090",
		GRPCAddr: ":8091",
		Logger:   logger,
		// 添加 HTTP 中间件
		HTTPMiddlewares: []mux.MiddlewareFunc{
			httpmiddleware.LoggingMiddleware(logger),
			httpmiddleware.RecoveryMiddleware(logger),
			httpmiddleware.TimeoutMiddleware(5 * time.Second),
		},
		// 添加 gRPC 拦截器
		GRPCUnaryInterceptors: []grpc.UnaryServerInterceptor{
			grpcmiddleware.UnaryLoggingInterceptor(logger),
			grpcmiddleware.UnaryRecoveryInterceptor(logger),
		},
		GRPCStreamInterceptors: []grpc.StreamServerInterceptor{
			grpcmiddleware.StreamLoggingInterceptor(logger),
			grpcmiddleware.StreamRecoveryInterceptor(logger),
		},
	}

	// 创建 API 服务器
	apiServer := server.NewGenericAPIServer(config)

	// 创建并安装 helloworld API 组
	helloworldInstaller := helloworld.NewInstaller(logger, config.GRPCAddr)
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
	if err := apiServer.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
