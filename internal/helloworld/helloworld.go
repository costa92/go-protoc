package helloworld

import (
	"context"

	"github.com/costa92/go-protoc/pkg/app"
	// "github.com/costa92/go-protoc/pkg/auth"
	"github.com/costa92/go-protoc/pkg/log"
	"google.golang.org/grpc"

	helloworldv1 "github.com/costa92/go-protoc/pkg/api/helloworld/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// GreeterServer 是 helloworld.GreeterServer 的实现
type GreeterServer struct {
	helloworldv1.UnimplementedGreeterServer
}

// NewGreeterServer 创建一个新的 GreeterServer
func NewGreeterServer() *GreeterServer {
	return &GreeterServer{}
}

// SayHello 实现 helloworld.GreeterServer 接口
func (s *GreeterServer) SayHello(ctx context.Context, req *helloworldv1.HelloRequest) (*helloworldv1.HelloReply, error) {
	log.L().Infof("Received: %v", req.GetName())
	// 示例：如何从上下文中获取用户信息
	// if userInfo, ok := auth.FromContext(ctx); ok {
	// 	log.L().WithValues("user", userInfo.GetUsername()).Infof("user is authenticated")
	// }

	return &helloworldv1.HelloReply{Message: "Hello " + req.GetName()}, nil
}

// Installer 实现了 APIGroupInstaller 接口
type Installer struct{}

// NewInstaller 创建一个新的 Installer
func NewInstaller() *Installer {
	return &Installer{}
}

// Install 将 helloworld API 组安装到 gRPC 和 HTTP 服务器。
func (i *Installer) Install(grpcServer *grpc.Server, httpServer *app.HTTPServer) {
	s := NewGreeterServer()
	helloworldv1.RegisterGreeterServer(grpcServer, s)

	gwmux := runtime.NewServeMux()
	if err := helloworldv1.RegisterGreeterHandlerServer(context.Background(), gwmux, s); err != nil {
		log.L().Fatalf("Failed to register greeter handler server: %v", err)
	}

	httpServer.Router().PathPrefix("/").Handler(gwmux)
}

var _ app.APIGroupInstaller = &Installer{}
