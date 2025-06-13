package service

import (
	"context"
	"fmt"

	helloworldv1 "github.com/costa92/go-protoc/pkg/api/helloworld/v1"
	helloworldv2 "github.com/costa92/go-protoc/pkg/api/helloworld/v2"
	userv1 "github.com/costa92/go-protoc/pkg/api/user/v1"
	"github.com/costa92/go-protoc/pkg/app"
	"github.com/google/wire"
)

// ProviderSet 是 service 的 Wire 提供者集合
var ProviderSet = wire.NewSet(
	// 1. 提供具体的 v1, v2 服务实现
	NewGreeterV1Service,
	NewGreeterV2Service,
	NewUserService,
	// 2. 将具体的服务实现绑定到它们所实现的 gRPC Server 接口
	// 这是为了让 wire 知道谁满足了 NewGreeterInstaller 的接口依赖
	wire.Bind(new(helloworldv1.GreeterServer), new(*GreeterV1Service)),
	wire.Bind(new(helloworldv2.GreeterServer), new(*GreeterV2Service)),
	wire.Bind(new(userv1.UserServiceServer), new(*UserService)),

	// 3. 提供我们的安装器
	NewGreeterInstaller,

	// 4. 将具体的安装器实现绑定到应用所依赖的 APIGroupInstaller 接口
	// 这是为了让 wire 知道谁满足了 newApp 函数的接口依赖
	// wire.Bind(new(app.APIGroupInstaller), new(*greeterInstaller)),
)

// greeterInstaller 封装了 Greeter 服务的注册逻辑, 现在依赖 v1 和 v2 两个服务接口。

type greeterInstaller struct {
	v1Svc   helloworldv1.GreeterServer
	v2Svc   helloworldv2.GreeterServer
	userSvc userv1.UserServiceServer
}

// NewGreeterInstaller 创建一个新的 greeter API 安装器。
// 它的依赖项变成了 v1 和 v2 的服务接口。
func NewGreeterInstaller(v1Svc helloworldv1.GreeterServer, v2Svc helloworldv2.GreeterServer, userSvc userv1.UserServiceServer) app.APIGroupInstaller {
	return &greeterInstaller{
		v1Svc:   v1Svc,
		v2Svc:   v2Svc,
		userSvc: userSvc,
	}
}

// Install 负责将 greeter gRPC 服务和 HTTP Gateway 注册到服务器
func (i *greeterInstaller) Install(grpcServer *app.GRPCServer, httpServer *app.HTTPServer) error {
	// 1. 分别注册 v1 和 v2 的 gRPC 服务
	helloworldv1.RegisterGreeterServer(grpcServer.Server(), i.v1Svc)
	helloworldv2.RegisterGreeterServer(grpcServer.Server(), i.v2Svc)
	userv1.RegisterUserServiceServer(grpcServer.Server(), i.userSvc)

	// 2. 分别注册 v1 和 v2 的 HTTP Gateway 处理器
	// 网关通过 grpcServer.ClientConn 连接到 gRPC 服务，扮演客户端的角色
	// 注册 gRPC-Gateway 处理器
	gwmux := httpServer.GatewayMux()
	if gwmux == nil {
		return fmt.Errorf("GatewayMux is nil")
	}
	if err := helloworldv1.RegisterGreeterHandlerServer(context.Background(), gwmux, i.v1Svc); err != nil {
		return err
	}
	if err := helloworldv2.RegisterGreeterHandlerServer(context.Background(), gwmux, i.v2Svc); err != nil {
		return err
	}
	if err := userv1.RegisterUserServiceHandlerServer(context.Background(), gwmux, i.userSvc); err != nil {
		return err
	}
	return nil
}
