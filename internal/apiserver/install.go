package apiserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/costa92/go-protoc/internal/apiserver/service"
	"github.com/costa92/go-protoc/pkg/app"

	// "github.com/costa92/go-protoc/pkg/auth"
	"github.com/costa92/go-protoc/pkg/log"

	helloworldv1 "github.com/costa92/go-protoc/pkg/api/helloworld/v1"
	helloworldv2 "github.com/costa92/go-protoc/pkg/api/helloworld/v2"
)

// Installer 实现了 APIGroupInstaller 接口
type Installer struct{}

// NewInstaller 创建一个新的 Installer
func NewInstaller() *Installer {
	return &Installer{}
}

// Install 将 helloworld API 组安装到 gRPC 和 HTTP 服务器。
func (i *Installer) Install(grpcServer *app.GRPCServer, httpServer *app.HTTPServer) error {
	log.L().Infow("开始安装 helloworld API 组")

	s := service.NewGreeterV1Server()
	// 注册 gRPC 服务
	helloworldv1.RegisterGreeterServer(grpcServer.Server(), s)
	log.L().Infow("已注册 gRPC 服务")

	sv2 := service.NewGreeterV2Server()
	helloworldv2.RegisterGreeterServer(grpcServer.Server(), sv2)
	log.L().Infow("已注册 gRPC 服务")
	// 注册 gRPC-Gateway 处理器
	log.L().Infow("开始注册 gRPC-Gateway 处理器")
	gwmux := httpServer.GatewayMux()
	if gwmux == nil {
		return fmt.Errorf("GatewayMux is nil")
	}

	gwmux.HandlePath("GET", "/test", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		w.Write([]byte("test "))
	})

	log.L().Infow("开始调用 RegisterGreeterHandlerServer")
	if err := helloworldv1.RegisterGreeterHandlerServer(context.Background(), gwmux, s); err != nil {
		log.L().Errorf("Failed to register greeter handler server: %v", err)
		return err
	}
	log.L().Infow("已成功注册 gRPC-Gateway 处理器")

	if err := helloworldv2.RegisterGreeterHandlerServer(context.Background(), gwmux, sv2); err != nil {
		log.L().Errorf("Failed to register greeter handler server: %v", err)
		return err
	}
	log.L().Infow("已成功注册 gRPC-Gateway 处理器")

	return nil
}

// 在初始化时自动注册
func init() {
	log.L().Infow("Registering helloworld API group")
	app.RegisterAPIGroup(NewInstaller())
}

var _ app.APIGroupInstaller = &Installer{}
