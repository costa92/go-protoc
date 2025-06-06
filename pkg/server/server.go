package server

import (
	"context"
	"fmt"

	"github.com/costa92/go-protoc/pkg/app"
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

// GenericAPIServer 是一个通用的 API 服务器
type GenericAPIServer struct {
	// 配置选项
	config *Config
	// HTTP 服务器
	httpServer *app.HTTPServer
	// gRPC 服务器
	grpcServer *app.GRPCServer
	// 日志器
	logger *zap.Logger
	// API 组安装器列表
	installers []Installer
}

// Config 包含 GenericAPIServer 的配置选项
type Config struct {
	// HTTP 服务器地址
	HTTPAddr string
	// gRPC 服务器地址
	GRPCAddr string
	// 日志器
	Logger *zap.Logger
	// HTTP 中间件列表
	HTTPMiddlewares []mux.MiddlewareFunc
	// gRPC 一元拦截器列表
	GRPCUnaryInterceptors []grpc.UnaryServerInterceptor
	// gRPC 流式拦截器列表
	GRPCStreamInterceptors []grpc.StreamServerInterceptor
}

// NewConfig 创建一个默认的配置
func NewConfig() *Config {
	logger, _ := zap.NewProduction()
	return &Config{
		HTTPAddr: ":8090",
		GRPCAddr: ":8091",
		Logger:   logger,
	}
}

// NewGenericAPIServer 创建一个新的 GenericAPIServer
func NewGenericAPIServer(config *Config) *GenericAPIServer {
	if config == nil {
		config = NewConfig()
	}

	// 创建 HTTP 服务器选项
	httpOpts := []app.ServerOption{
		app.WithHTTPMiddlewares(config.HTTPMiddlewares...),
	}

	// 创建 gRPC 服务器选项
	grpcOpts := []app.ServerOption{
		app.WithGRPCUnaryInterceptors(config.GRPCUnaryInterceptors...),
		app.WithGRPCStreamInterceptors(config.GRPCStreamInterceptors...),
	}

	// 创建 HTTP 服务器
	httpServer := app.NewHTTPServer(config.HTTPAddr, config.Logger, httpOpts...)

	// 创建 gRPC 服务器
	grpcServer := app.NewGRPCServer(config.GRPCAddr, config.Logger, grpcOpts...)

	return &GenericAPIServer{
		config:     config,
		httpServer: httpServer,
		grpcServer: grpcServer,
		logger:     config.Logger,
	}
}

// InstallAPIGroup 安装一个 API 组
func (s *GenericAPIServer) InstallAPIGroup(installer Installer) error {
	// 先注册 gRPC 服务
	if err := installer.RegisterGRPC(s.grpcServer.Server()); err != nil {
		return fmt.Errorf("failed to register gRPC service: %w", err)
	}

	// 再注册 HTTP 路由
	if err := installer.Install(s.httpServer.Router()); err != nil {
		return fmt.Errorf("failed to install API group: %w", err)
	}

	s.installers = append(s.installers, installer)
	return nil
}

// Run 启动 API 服务器
func (s *GenericAPIServer) Run(ctx context.Context) error {
	// 创建一个错误组
	errChan := make(chan error, 2)

	// 启动 gRPC 服务器
	go func() {
		if err := s.grpcServer.Start(ctx); err != nil {
			errChan <- err
		}
	}()

	// 启动 HTTP 服务器
	go func() {
		if err := s.httpServer.Start(ctx); err != nil {
			errChan <- err
		}
	}()

	// 等待错误或上下文取消
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		s.logger.Info("shutting down servers")
		return nil
	}
}
