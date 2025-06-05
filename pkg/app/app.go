package app

import (
	"log"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type App struct {
}

func NewApp(options ...func(*App)) *App {
	return &App{}
}

func (a *App) Run() {
	eg := errgroup.Group{}
	// 启动 gRPC 服务器
	eg.Go(func() error {
		grpc := NewGRPCServer(WithGRPCServer(grpc.NewServer()))
		if _, err := grpc.RunGRPCServer(":8100"); err != nil {
			return err
		}
		return nil
	})

	eg.Go(func() error {
		httpServer := NewHTTPServer(WithAddr(":8080"))
		if _, err := httpServer.RunHTTPServer(nil, nil); err != nil {
			return err
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}
}
