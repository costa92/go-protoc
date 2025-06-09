package helloworld

import (
	"context"

	"github.com/costa92/go-protoc/internal/apiserver"
	"github.com/costa92/go-protoc/pkg/app"

	// "github.com/costa92/go-protoc/pkg/auth"
	"github.com/costa92/go-protoc/pkg/log"
	"github.com/costa92/go-protoc/pkg/response"

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
func (i *Installer) Install(grpcServer *app.GRPCServer, httpServer *app.HTTPServer) error {
	s := NewGreeterServer()
	helloworldv1.RegisterGreeterServer(grpcServer.Server(), s)

	// 创建一个带有自定义 marshalers 的 ServeMux
	gwmux := runtime.NewServeMux()

	// 配置统一响应系统
	response.Setup(gwmux)

	// 注册 gRPC-Gateway 处理器
	if err := helloworldv1.RegisterGreeterHandlerServer(context.Background(), gwmux, s); err != nil {
		log.L().Errorf("Failed to register greeter handler server: %v", err)
		return err
	}

	// 注册直接的 HTTP 处理器
	router := httpServer.Router()
	router.HandleFunc("/health", HealthCheckHandler).Methods("GET")
	router.HandleFunc("/error-example", SimpleErrorHandler).Methods("GET")
	router.HandleFunc("/download-example", FileDownloadHandler).Methods("GET")

	// 注册 gRPC 网关路由，处理其他请求
	router.PathPrefix("/").Handler(gwmux)
	return nil
}

// 在初始化时自动注册
func init() {
	apiserver.RegisterAPIGroup(NewInstaller())
}

var _ apiserver.APIGroupInstaller = &Installer{}
