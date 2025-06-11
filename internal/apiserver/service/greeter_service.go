package service

import (
	"context"

	helloworldv1 "github.com/costa92/go-protoc/pkg/api/helloworld/v1"
	helloworldv2 "github.com/costa92/go-protoc/pkg/api/helloworld/v2"
	"github.com/costa92/go-protoc/pkg/logger"

	"github.com/google/wire"
)

// GreeterV1Service 直接实现 v1 版本的 GreeterServer 接口
type GreeterV1Service struct {
	helloworldv1.UnimplementedGreeterServer
	logger logger.Logger
}

// NewGreeterV1Service 创建一个新的 GreeterV1Service 实例
func NewGreeterV1Service(logger logger.Logger) *GreeterV1Service {
	return &GreeterV1Service{
		logger: logger,
	}
}

// SayHello 实现 v1 版本的 SayHello 方法
func (s *GreeterV1Service) SayHello(ctx context.Context, req *helloworldv1.HelloRequest) (*helloworldv1.HelloReply, error) {
	s.logger.Infow("收到 SayHelloV1 请求", "name", req.GetName())
	return &helloworldv1.HelloReply{Message: "V1: Hello " + req.GetName()}, nil
}

// SayHelloAgain 实现 v1 版本的 SayHelloAgain 方法
func (s *GreeterV1Service) SayHelloAgain(ctx context.Context, req *helloworldv1.HelloRequest) (*helloworldv1.HelloReply, error) {
	s.logger.Infow("收到 SayHelloAgainV1 请求", "name", req.GetName())
	return &helloworldv1.HelloReply{Message: "V1: Hello again " + req.GetName()}, nil
}

// GreeterV2Service 实现 v2 版本的 Greeter 服务
type GreeterV2Service struct {
	helloworldv2.UnimplementedGreeterServer
	logger logger.Logger
}

// NewGreeterV2Service 创建一个新的 GreeterV2Service
func NewGreeterV2Service(logger logger.Logger) *GreeterV2Service {
	return &GreeterV2Service{
		logger: logger,
	}
}

// SayHello 实现 v2 版本的 SayHello 方法
func (s *GreeterV2Service) SayHello(ctx context.Context, req *helloworldv2.HelloRequest) (*helloworldv2.HelloReply, error) {
	logger.Infow("收到 SayHelloV2 请求", "name", req.GetName())
	return &helloworldv2.HelloReply{Message: "V2: Hello " + req.GetName()}, nil
}

// SayHelloAgain 实现 v2 版本的 SayHelloAgain 方法
func (s *GreeterV2Service) SayHelloAgain(ctx context.Context, req *helloworldv2.HelloRequest) (*helloworldv2.HelloReply, error) {
	logger.Infow("收到 SayHelloAgainV2 请求", "name", req.GetName())
	return &helloworldv2.HelloReply{Message: "V2: Hello again " + req.GetName()}, nil
}

// ProviderSet 是 service 的 Wire 提供者集合
var ProviderSet = wire.NewSet(
	NewGreeterV1Service,
	NewGreeterV2Service,
	wire.Bind(new(helloworldv1.GreeterServer), new(*GreeterV1Service)),
	wire.Bind(new(helloworldv2.GreeterServer), new(*GreeterV2Service)),
)
