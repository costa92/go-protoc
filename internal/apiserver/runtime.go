package apiserver

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/costa92/go-protoc/internal/apiserver/handlers"
	"github.com/costa92/go-protoc/internal/apiserver/options"
	"github.com/costa92/go-protoc/pkg/app"
	"github.com/costa92/go-protoc/pkg/logger"
)

// RunFunc 定义运行 API 服务器的函数类型
type RunFunc func(opts *options.Options) error

// APIServer 封装 API 服务器的运行时
type APIServer struct {
	name      string
	logger    logger.Logger
	installer *handlers.Installer
}

// NewAPIServer 创建一个新的 API 服务器实例
func NewAPIServer(
	name string,
	logger logger.Logger,
	installer *handlers.Installer,
) *APIServer {
	return &APIServer{
		name:      name,
		logger:    logger,
		installer: installer,
	}
}

// Run 运行 API 服务器
func (s *APIServer) Run(opts *options.Options) error {
	// 完成选项配置，从配置文件加载配置
	if err := opts.Complete(); err != nil {
		return err
	}

	// 打印配置信息用于验证
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

	// 获取 gRPC 和 HTTP 选项
	grpcOpts := opts.GetGRPCOptions()
	httpOpts := opts.GetHTTPOptions()

	// 创建 gRPC 监听器
	grpcListener, err := net.Listen("tcp", grpcOpts.Addr)
	if err != nil {
		s.logger.Errorw("创建 gRPC 监听器失败", "error", err)
		return err
	}

	// 创建服务器
	grpcServer := app.NewGRPCServer(s.name, grpcListener)
	httpServer := app.NewHTTPServer(s.name, httpOpts.Addr)

	// 使用installer安装API
	s.logger.Infow("使用installer安装API组")
	if err := s.installer.Install(grpcServer, httpServer); err != nil {
		s.logger.Errorw("安装 API 组失败", "error", err)
		return err
	}

	// 确保路由最终化
	httpServer.FinalizeRoutes()

	// 创建上下文用于优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建错误通道
	errCh := make(chan error, 2)

	// 启动 gRPC 服务器
	go func() {
		s.logger.Infow("启动 gRPC 服务器", "address", grpcOpts.Addr)
		if err := grpcServer.Start(ctx); err != nil {
			errCh <- err
		}
	}()

	// 启动 HTTP 服务器
	go func() {
		s.logger.Infow("启动 HTTP 服务器", "address", httpOpts.Addr)
		if err := httpServer.Start(ctx); err != nil {
			errCh <- err
		}
	}()

	// 处理信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号或错误
	select {
	case sig := <-sigCh:
		s.logger.Infow("收到信号，开始优雅关闭", "signal", sig)

		// 创建关闭上下文
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), grpcOpts.Timeout)
		defer shutdownCancel()

		// 优雅关闭服务器
		if err := grpcServer.Stop(shutdownCtx); err != nil {
			s.logger.Errorw("关闭 gRPC 服务器失败", "error", err)
		}

		if err := httpServer.Stop(shutdownCtx); err != nil {
			s.logger.Errorw("关闭 HTTP 服务器失败", "error", err)
		}

		s.logger.Infow("服务器已关闭")
		return nil

	case err := <-errCh:
		s.logger.Errorw("服务器发生错误", "error", err)
		return err
	}
}
