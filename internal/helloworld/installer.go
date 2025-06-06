package helloworld

import (
	"context"
	"fmt"
	"net/http"

	"github.com/costa92/go-protoc/internal/helloworld/service"
	v1 "github.com/costa92/go-protoc/pkg/api/helloworld/v1"
	v2 "github.com/costa92/go-protoc/pkg/api/helloworld/v2"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// APIGroupInstaller 实现了 server.Installer 接口
type APIGroupInstaller struct {
	logger       *zap.Logger
	grpcEndpoint string
	v1Service    v1.GreeterServer
	v2Service    v2.GreeterServer
}

// NewInstaller 创建一个新的 APIGroupInstaller
func NewInstaller(logger *zap.Logger, grpcEndpoint string) *APIGroupInstaller {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &APIGroupInstaller{
		logger:       logger,
		grpcEndpoint: grpcEndpoint,
		v1Service:    service.NewGreeterV1Server(),
		v2Service:    service.NewGreeterV2Server(),
	}
}

// Install 实现 server.Installer 接口
func (i *APIGroupInstaller) Install(router *mux.Router) error {
	// 创建 gRPC-Gateway 的 mux
	gwmux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// 注册 v1 和 v2 的 HTTP Endpoints
	ctx := context.Background()
	if err := v1.RegisterGreeterHandlerFromEndpoint(ctx, gwmux, i.grpcEndpoint, opts); err != nil {
		return fmt.Errorf("failed to register v1 handler: %w", err)
	}
	if err := v2.RegisterGreeterHandlerFromEndpoint(ctx, gwmux, i.grpcEndpoint, opts); err != nil {
		return fmt.Errorf("failed to register v2 handler: %w", err)
	}

	// 将 gateway mux 挂载到主路由器
	router.PathPrefix("/").Handler(gwmux)

	// 添加健康检查路由
	router.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	return nil
}

// RegisterGRPC 实现 server.Installer 接口
func (i *APIGroupInstaller) RegisterGRPC(srv *grpc.Server) error {
	// 注册 v1 和 v2 的 gRPC 服务
	v1.RegisterGreeterServer(srv, i.v1Service)
	v2.RegisterGreeterServer(srv, i.v2Service)
	return nil
}
