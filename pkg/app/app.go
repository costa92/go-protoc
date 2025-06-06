package app

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// APIGroupInfo 包含一个 API 组的所有信息
type APIGroupInfo struct {
	// API 组的名称，例如 "helloworld"
	GroupName string
	// API 组的版本，例如 "v1", "v2"
	Versions []string
	// 该 API 组的 gRPC 服务注册函数
	GRPCServiceRegister func(*grpc.Server)
	// 该 API 组的 HTTP 处理器注册函数
	HTTPHandlerRegister func(context.Context, *mux.Router, string, []grpc.DialOption) error
}

// Installer 负责将 API 组的路由安装到给定的路由器上
type Installer interface {
	// Install 将 API 组的路由安装到给定的路由器上
	Install(router *mux.Router) error
	// RegisterGRPC 注册 gRPC 服务
	RegisterGRPC(srv *grpc.Server) error
}

// App 是一个应用服务器，管理 HTTP 和 gRPC 服务
type App struct {
	// HTTP 服务器地址
	httpAddr string
	// gRPC 服务器地址
	grpcAddr string
	// HTTP 服务器
	httpServer *HTTPServer
	// gRPC 服务器
	grpcServer *GRPCServer
	// 日志器
	logger *zap.Logger
	// API 组安装器列表
	installers []Installer
	// HTTP 中间件列表
	httpMiddlewares []mux.MiddlewareFunc
	// gRPC 一元拦截器列表
	grpcUnaryInterceptors []grpc.UnaryServerInterceptor
	// gRPC 流式拦截器列表
	grpcStreamInterceptors []grpc.StreamServerInterceptor
}

// NewApp 创建一个新的 App 实例
func NewApp(httpAddr, grpcAddr string, logger *zap.Logger, opts ...ServerOption) *App {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}

	app := &App{
		httpAddr:               httpAddr,
		grpcAddr:               grpcAddr,
		logger:                 logger,
		httpMiddlewares:        make([]mux.MiddlewareFunc, 0),
		grpcUnaryInterceptors:  make([]grpc.UnaryServerInterceptor, 0),
		grpcStreamInterceptors: make([]grpc.StreamServerInterceptor, 0),
	}

	// 应用所有选项
	for _, opt := range opts {
		opt(app)
	}

	// 创建 HTTP 服务器
	httpOpts := []ServerOption{WithHTTPMiddlewares(app.httpMiddlewares...)}
	app.httpServer = NewHTTPServer(app.httpAddr, app.logger, httpOpts...)

	// 创建 gRPC 服务器
	grpcOpts := []ServerOption{
		WithGRPCUnaryInterceptors(app.grpcUnaryInterceptors...),
		WithGRPCStreamInterceptors(app.grpcStreamInterceptors...),
	}
	app.grpcServer = NewGRPCServer(app.grpcAddr, app.logger, grpcOpts...)

	return app
}

// InstallAPIGroup 安装一个 API 组
func (a *App) InstallAPIGroup(installer Installer) error {
	// 先注册 gRPC 服务
	if err := installer.RegisterGRPC(a.grpcServer.Server()); err != nil {
		return fmt.Errorf("failed to register gRPC service: %w", err)
	}

	// 再注册 HTTP 路由
	if err := installer.Install(a.httpServer.Router()); err != nil {
		return fmt.Errorf("failed to install API group: %w", err)
	}

	a.installers = append(a.installers, installer)
	return nil
}

// Start 启动应用服务器
func (a *App) Start(ctx context.Context) error {
	// 创建一个错误通道
	errChan := make(chan error, 2)

	// 启动 gRPC 服务器
	go func() {
		if err := a.grpcServer.Start(ctx); err != nil {
			errChan <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	// 启动 HTTP 服务器
	go func() {
		if err := a.httpServer.Start(ctx); err != nil {
			errChan <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// 等待错误或上下文取消
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return a.Stop()
	}
}

// Stop 停止应用服务器
func (a *App) Stop() error {
	// 创建一个带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 先停止 gRPC 服务器
	if a.grpcServer != nil && a.grpcServer.server != nil {
		a.logger.Info("stopping gRPC server")
		done := make(chan struct{})
		go func() {
			a.grpcServer.server.GracefulStop()
			close(done)
		}()

		select {
		case <-done:
			a.logger.Info("gRPC server stopped")
		case <-ctx.Done():
			a.logger.Warn("gRPC server shutdown timeout, forcing stop")
			a.grpcServer.server.Stop()
		}
	}

	// 再停止 HTTP 服务器
	if a.httpServer != nil && a.httpServer.server != nil {
		a.logger.Info("stopping HTTP server")
		if err := a.httpServer.server.Shutdown(ctx); err != nil {
			a.logger.Error("HTTP server shutdown error", zap.Error(err))
			return fmt.Errorf("HTTP server shutdown error: %w", err)
		}
		a.logger.Info("HTTP server stopped")
	}

	return nil
}
