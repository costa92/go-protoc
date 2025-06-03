package app

import (
	"net/http"

	"google.golang.org/grpc"
)

type App struct {
	grpcServer *grpc.Server
	httpServer *http.Server
}

func NewApp(grpcServer *grpc.Server, httpServer *http.Server) *App {
	return &App{
		grpcServer: grpcServer,
		httpServer: httpServer,
	}
}

func (a *App) Run() {

	// 启动 gRPC 服务器

	go func() {
		// 调用 RunGRPCServer 函数
		grpc := NewGRPCServer(WithGRPCServer(a.grpcServer))
		if _, err := grpc.RunGRPCServer(":8100"); err != nil {
			panic(err)
		}
	}()

	// 调用 RunHTTPServer 函数
	httpServer := RunHTTPServer(":8080", nil, nil)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
