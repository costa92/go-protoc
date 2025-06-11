package handlers

import (
	"context"
	"fmt"

	"github.com/costa92/go-protoc/pkg/app"
	"github.com/costa92/go-protoc/pkg/logger"

	helloworldv1 "github.com/costa92/go-protoc/pkg/api/helloworld/v1"
	helloworldv2 "github.com/costa92/go-protoc/pkg/api/helloworld/v2"
)

// GreeterV1Registration 是 v1 版本 Greeter 服务的注册
type GreeterV1Registration struct {
	service helloworldv1.GreeterServer
	logger  logger.Logger
}

// NewGreeterV1Registration 创建一个新的 GreeterV1Registration
func NewGreeterV1Registration(service helloworldv1.GreeterServer, logger logger.Logger) *GreeterV1Registration {
	return &GreeterV1Registration{
		service: service,
		logger:  logger,
	}
}

// Name 返回服务的名称
func (r *GreeterV1Registration) Name() string {
	return "greeter.v1"
}

// Register 注册服务到 gRPC 和 HTTP 服务器
func (r *GreeterV1Registration) Register(grpcServer *app.GRPCServer, httpServer *app.HTTPServer) error {
	// 注册 gRPC 服务
	helloworldv1.RegisterGreeterServer(grpcServer.Server(), r.service)
	r.logger.Infow("已注册 v1 gRPC 服务")

	// 注册 gRPC-Gateway 处理器
	gwmux := httpServer.GatewayMux()
	if gwmux == nil {
		return fmt.Errorf("GatewayMux is nil")
	}

	// 注册 v1 版本的 HTTP 处理器
	if err := helloworldv1.RegisterGreeterHandlerServer(context.Background(), gwmux, r.service); err != nil {
		r.logger.Errorw("注册 v1 greeter handler server 失败", "error", err)
		return err
	}
	r.logger.Infow("已成功注册 v1 gRPC-Gateway 处理器")

	return nil
}

// GreeterV2Registration 是 v2 版本 Greeter 服务的注册
type GreeterV2Registration struct {
	service helloworldv2.GreeterServer
	logger  logger.Logger
}

// NewGreeterV2Registration 创建一个新的 GreeterV2Registration
func NewGreeterV2Registration(service helloworldv2.GreeterServer, logger logger.Logger) *GreeterV2Registration {
	return &GreeterV2Registration{
		service: service,
		logger:  logger,
	}
}

// Name 返回服务的名称
func (r *GreeterV2Registration) Name() string {
	return "greeter.v2"
}

// Register 注册服务到 gRPC 和 HTTP 服务器
func (r *GreeterV2Registration) Register(grpcServer *app.GRPCServer, httpServer *app.HTTPServer) error {
	// 注册 gRPC 服务
	helloworldv2.RegisterGreeterServer(grpcServer.Server(), r.service)
	r.logger.Infow("已注册 v2 gRPC 服务")

	// 注册 gRPC-Gateway 处理器
	gwmux := httpServer.GatewayMux()
	if gwmux == nil {
		return fmt.Errorf("GatewayMux is nil")
	}

	// 注册 v2 版本的 HTTP 处理器
	if err := helloworldv2.RegisterGreeterHandlerServer(context.Background(), gwmux, r.service); err != nil {
		r.logger.Errorw("注册 v2 greeter handler server 失败", "error", err)
		return err
	}
	r.logger.Infow("已成功注册 v2 gRPC-Gateway 处理器")

	return nil
}
