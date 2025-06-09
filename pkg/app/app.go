package app

import (
	"context"
	"fmt"

	"github.com/costa92/go-protoc/pkg/log"
	"golang.org/x/sync/errgroup"
)

// Server 是服务器接口，定义了启动和停止方法
type Server interface {
	// Start 启动服务器并阻塞直到上下文取消
	Start(ctx context.Context) error
	// Stop 优雅地关闭服务器
	Stop(ctx context.Context) error
}

// App 是应用程序的框架，负责管理各种类型的服务
type App struct {
	servers []Server
	name    string
}

// NewApp 创建一个新的 App 实例
func NewApp(name string, servers ...Server) *App {
	return &App{
		servers: servers,
		name:    name,
	}
}

// AddServer 向 App 添加一个服务
func (a *App) AddServer(server Server) {
	a.servers = append(a.servers, server)
}

// Start 启动应用程序中的所有服务
func (a *App) Start(ctx context.Context) error {
	log.Infof("启动应用 %s", a.name)

	// 创建一个错误组，用于并发管理服务
	g, ctx := errgroup.WithContext(ctx)

	// 并发启动所有服务
	for _, srv := range a.servers {
		srv := srv // 创建闭包变量副本
		g.Go(func() error {
			log.Infof("正在启动服务 %T", srv)
			return srv.Start(ctx)
		})
	}

	// 等待所有服务完成或出现错误
	return g.Wait()
}

// Stop 优雅地停止应用程序中的所有服务
func (a *App) Stop(ctx context.Context) error {
	log.Infof("正在关闭应用 %s", a.name)

	// 创建一个错误组，用于并发管理服务关闭
	g, ctx := errgroup.WithContext(ctx)

	// 并发停止所有服务
	for _, srv := range a.servers {
		srv := srv // 创建闭包变量副本
		g.Go(func() error {
			return srv.Stop(ctx)
		})
	}

	// 等待所有服务关闭或超时
	if err := g.Wait(); err != nil {
		return fmt.Errorf("关闭服务时发生错误: %v", err)
	}

	log.Infof("应用 %s 已成功关闭", a.name)
	return nil
}
