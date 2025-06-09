package app

import (
	"context"
	"net"
	"net/http"
	"sync"

	"github.com/costa92/go-protoc/pkg/log"
	"github.com/gorilla/mux"
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

// App 是应用程序的框架，负责管理 gRPC 和 HTTP 服务器。
type App struct {
	httpServer *HTTPServer
	grpcServer *grpc.Server
	opts       *Options
	wg         sync.WaitGroup
}

// NewApp 创建一个新的 App 实例。
func NewApp(httpAddr, grpcAddr string, opts ...ServerOption) *App {
	options := NewOptions()
	options.httpAddr = httpAddr
	options.grpcAddr = grpcAddr
	for _, o := range opts {
		o(options)
	}

	// 如果没有提供监听器，则创建一个
	if options.grpcListener == nil {
		lis, err := net.Listen("tcp", options.grpcAddr)
		if err != nil {
			log.L().Fatalf("Failed to listen: %v", err)
		}
		options.grpcListener = lis
	}

	// 创建 gRPC 服务器
	grpcServer := grpc.NewServer(
		append(
			options.grpcOptions,
			grpc.ChainUnaryInterceptor(options.grpcUnaryInterceptors...),
			grpc.ChainStreamInterceptor(options.grpcStreamInterceptors...),
		)...,
	)

	// 创建 HTTP 服务器
	httpServer := NewHTTPServer(options.httpAddr, options.httpMiddlewares...)

	return &App{
		httpServer: httpServer,
		grpcServer: grpcServer,
		opts:       options,
	}
}

// Start 启动应用程序。
func (a *App) Start(ctx context.Context) error {
	// 启动 gRPC 服务器
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		log.L().Infof("gRPC server is listening on %s", a.opts.grpcListener.Addr().String())
		if err := a.grpcServer.Serve(a.opts.grpcListener); err != nil {
			log.L().Fatalf("gRPC server failed to serve: %v", err)
		}
	}()

	// 启动 HTTP 服务器
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		log.L().Infof("HTTP server is listening on %s", a.opts.httpAddr)
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.L().Fatalf("HTTP server failed to listen and serve: %v", err)
		}
	}()

	<-ctx.Done()
	return a.Stop()
}

// Stop 优雅地停止应用程序。
func (a *App) Stop() error {
	log.L().Infof("Shutting down servers...")

	// 停止 gRPC 服务器
	a.grpcServer.GracefulStop()

	// 停止 HTTP 服务器
	if err := a.httpServer.Shutdown(context.Background()); err != nil {
		log.L().Errorf("Failed to shutdown HTTP server: %v", err)
	}

	a.wg.Wait()
	log.L().Infof("Servers are shut down.")
	return nil
}

// GetHTTPServer 返回 HTTP 服务器实例。
func (a *App) GetHTTPServer() *HTTPServer {
	return a.httpServer
}

// InstallAPIGroup 将 API 组安装到服务器。
func (a *App) InstallAPIGroup(installer APIGroupInstaller) {
	installer.Install(a.grpcServer, a.httpServer)
}

// APIGroupInstaller 定义了用于安装 API 组的接口。
type APIGroupInstaller interface {
	Install(grpcServer *grpc.Server, httpServer *HTTPServer)
}

// HTTPServer 是对 http.Server 的包装。
type HTTPServer struct {
	*http.Server
	router *mux.Router
}

// NewHTTPServer 创建一个新的 HTTPServer 实例。
func NewHTTPServer(addr string, middlewares ...mux.MiddlewareFunc) *HTTPServer {
	router := mux.NewRouter()
	for _, mw := range middlewares {
		router.Use(mw)
	}

	return &HTTPServer{
		Server: &http.Server{
			Addr:    addr,
			Handler: router,
		},
		router: router,
	}
}

// Router 返回 mux.Router 实例。
func (s *HTTPServer) Router() *mux.Router {
	return s.router
}
