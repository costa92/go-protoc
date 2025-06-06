package app

import (
	"context"
	"log"
	"net/http"
	"time"

	// "github.com/grpc-ecosystem/go-grpc-middleware/providers/zap/v2"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type ServiceRegisterFunc func(srv *grpc.Server, logger *zap.Logger)

// GatewayRegisterFunc 是一个函数类型，用于封装注册 gRPC-Gateway 处理器的逻辑
type GatewayRegisterFunc func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error

type AppOption func(*App)

type App struct {
	serviceRegisterFuncs []ServiceRegisterFunc
	logger               *zap.Logger
	gatewayRegisterFuncs []GatewayRegisterFunc
}

func WithServiceRegisterFunc(serviceRegisterFunc ServiceRegisterFunc) AppOption {
	return func(a *App) {
		a.serviceRegisterFuncs = append(a.serviceRegisterFuncs, serviceRegisterFunc)
	}
}

func WithGatewayRegisterFuncs(gatewayRegisterFuncs ...GatewayRegisterFunc) AppOption {
	return func(a *App) {
		a.gatewayRegisterFuncs = append(a.gatewayRegisterFuncs, gatewayRegisterFuncs...)
	}
}

func WithLogger(logger *zap.Logger) AppOption {
	return func(a *App) {
		a.logger = logger
	}
}

func NewApp(options ...AppOption) *App {
	app := &App{}
	for _, option := range options {
		option(app)
	}
	return app
}

func (a *App) Run() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	eg, ctx := errgroup.WithContext(ctx)

	// 启动 gRPC 服务器
	eg.Go(func() error {
		a.logger.Info("starting gRPC server")
		grpc := NewGRPCServer(a.logger, a.serviceRegisterFuncs...)
		if _, err := grpc.RunGRPCServer(":8100"); err != nil {
			a.logger.Error("failed to run gRPC server", zap.Error(err))
			return err
		}
		return nil
	})

	// 启动 HTTP 服务器
	eg.Go(func() error {
		httpServer := NewHTTPServer(
			WithAddr(":8080"),
			WithHTTPServerLogger(a.logger),
			WithHTTPServerGatewayRegisterFuncs(a.gatewayRegisterFuncs...),
		)

		server, err := httpServer.RunHTTPServer()
		if err != nil {
			a.logger.Error("failed to create HTTP server", zap.Error(err))
			return err
		}

		// 在一个新的 goroutine 中启动服务器
		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				a.logger.Error("HTTP server error", zap.Error(err))
			}
		}()

		a.logger.Info("HTTP server started successfully")

		// 等待上下文取消
		<-ctx.Done()

		// 优雅关闭服务器
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			a.logger.Error("failed to shutdown HTTP server", zap.Error(err))
			return err
		}

		return nil
	})

	if err := eg.Wait(); err != nil {
		a.logger.Error("application error", zap.Error(err))
		log.Fatal(err)
	}
}
