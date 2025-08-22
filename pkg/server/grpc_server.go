package server

import (
	"context"
	"net"

	"github.com/costa92/go-protoc/v2/pkg/logger"
	genericoptions "github.com/costa92/go-protoc/v2/pkg/options"
	"github.com/costa92/go-protoc/v2/pkg/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// GRPCServer 代表一个 GRPC 服务器.
type GRPCServer struct {
	srv *grpc.Server
	lis net.Listener
}

// NewGRPCServer 创建一个新的 GRPC 服务器实例.
func NewGRPCServer(
	grpcOptions *genericoptions.GRPCOptions,
	tlsOptions *genericoptions.TLSOptions,
	serverOptions []grpc.ServerOption,
	registerServer func(grpc.ServiceRegistrar),
) (*GRPCServer, error) {
	lis, err := net.Listen("tcp", grpcOptions.Addr)
	if err != nil {
		logger.Errorw("Failed to listen", "error", err)
		return nil, err
	}

	if tlsOptions != nil && tlsOptions.UseTLS {
		tlsConfig := tlsOptions.MustTLSConfig()
		serverOptions = append(serverOptions, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	grpcsrv := grpc.NewServer(serverOptions...)

	registerServer(grpcsrv)
	registerHealthServer(grpcsrv)
	reflection.Register(grpcsrv)

	return &GRPCServer{
		srv: grpcsrv,
		lis: lis,
	}, nil
}

// RunOrDie 启动 GRPC 服务器并在出错时记录致命错误.
func (s *GRPCServer) RunOrDie() {
	versionInfo := version.Get()
	logger.Infow("Starting gRPC server",
		"protocol", "grpc",
		"addr", s.lis.Addr().String(),
		"service", versionInfo.ServiceName,
		"version", versionInfo.GitVersion,
		"branch", versionInfo.GitBranch,
		"commit", versionInfo.GitCommit[:8], // 显示短commit hash
		"build_date", versionInfo.BuildDate,
	)

	if err := s.srv.Serve(s.lis); err != nil {
		logger.Fatalw("Failed to start gRPC server", "err", err,
			"service", versionInfo.ServiceName,
			"protocol", "grpc",
			"addr", s.lis.Addr().String(),
		)
	}
}

// GracefulStop 优雅地关闭 GRPC 服务器.
func (s *GRPCServer) GracefulStop(ctx context.Context) {
	versionInfo := version.Get()
	logger.Infow("Gracefully stopping gRPC server",
		"service", versionInfo.ServiceName,
		"protocol", "grpc",
		"addr", s.lis.Addr().String(),
	)
	s.srv.GracefulStop()
	logger.Infow("gRPC server stopped gracefully",
		"service", versionInfo.ServiceName,
		"protocol", "grpc",
	)
}

// registerHealthServer 注册健康检查服务.
func registerHealthServer(grpcsrv *grpc.Server) {
	// 创建健康检查服务实例
	healthServer := health.NewServer()

	// 设定服务的健康状态
	healthServer.SetServingStatus("MiniBlog", grpc_health_v1.HealthCheckResponse_SERVING)

	// 注册健康检查服务
	grpc_health_v1.RegisterHealthServer(grpcsrv, healthServer)
}
