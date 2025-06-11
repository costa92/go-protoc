package handlers

import (
	"github.com/costa92/go-protoc/pkg/app"
	"github.com/costa92/go-protoc/pkg/logger"

	"github.com/google/wire"
)

// ServiceRegister 定义服务注册接口
type ServiceRegister interface {
	RegisterGRPC(server *app.GRPCServer) error
	RegisterHTTP(mux *app.HTTPServer) error
}

// Installer 实现了 app.APIGroupInstaller 接口
type Installer struct {
	registry *ServiceRegistry
	logger   logger.Logger
}

// NewInstaller 创建一个新的 Installer 实例
func NewInstaller(
	registry *ServiceRegistry,
	greeterV1Registration *GreeterV1Registration,
	greeterV2Registration *GreeterV2Registration,
	logger logger.Logger,
) *Installer {
	// 注册所有服务
	registry.Register(greeterV1Registration)
	registry.Register(greeterV2Registration)

	// 在这里添加更多服务注册
	// registry.Register(newUserServiceRegistration)
	// registry.Register(newOrderServiceRegistration)
	// 等等...

	return &Installer{
		registry: registry,
		logger:   logger,
	}
}

// Install 将 API 组安装到 gRPC 和 HTTP 服务器
func (i *Installer) Install(grpcServer *app.GRPCServer, httpServer *app.HTTPServer) error {
	i.logger.Infow("开始安装 API 组")

	// 注册所有服务
	if err := i.registry.RegisterAll(grpcServer, httpServer); err != nil {
		return err
	}

	i.logger.Infow("所有 API 组安装完成")
	return nil
}

// 确保 Installer 实现了 app.APIGroupInstaller 接口
var _ app.APIGroupInstaller = &Installer{}

// ProviderSet 是 handlers 的 Wire 提供者集合
var ProviderSet = wire.NewSet(
	NewInstaller,
	NewServiceRegistry,
	NewGreeterV1Registration,
	NewGreeterV2Registration,
	wire.Bind(new(app.APIGroupInstaller), new(*Installer)),
)
