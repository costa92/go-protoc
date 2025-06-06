package app

import (
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type GRPCServerOption func(*GRPCServer) func(*grpc.Server)

func WithGRPCServer(grpcServer *grpc.Server) GRPCServerOption {
	return func(s *GRPCServer) func(*grpc.Server) {
		return func(gs *grpc.Server) {
			// 这里可以添加自定义的 grpc.Server 配置逻辑
		}
	}
}

func NewGRPCServer(logger *zap.Logger, serviceRegisterFuncs ...ServiceRegisterFunc) *GRPCServer {
	return &GRPCServer{
		logger:               logger,
		serviceRegisterFuncs: serviceRegisterFuncs,
	}
}

type GRPCServer struct {
	logger               *zap.Logger
	serviceRegisterFuncs []ServiceRegisterFunc
}

// RunGRPCServer runs gRPC service to publish ToDo service
func (s *GRPCServer) RunGRPCServer(grpcAddr string) (*grpc.Server, error) {
	s.logger.Info("starting gRPC server", zap.String("grpcAddr", grpcAddr))
	grpcServer := grpc.NewServer()
	// 应用选项
	for _, serviceRegisterFunc := range s.serviceRegisterFuncs {
		serviceRegisterFunc(grpcServer, s.logger)
	}

	// 注册反射服务，这对于 gRPC 客户端（如 grpcurl）发现服务很有用
	reflection.Register(grpcServer)
	s.logger.Info("registered gRPC reflection service")

	// 注册健康检查服务
	healthSrv := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthSrv)
	s.logger.Info("registered gRPC health check service")
	// 初始化时，可以将所有服务的状态都设置为 SERVING，或者更精细地管理
	// 你也可以在每个服务注册函数内部去设置其健康状态
	healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING) // "" 表示整个服务器的总体状态
	s.logger.Info("set gRPC health check service to serving")

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return nil, err
	}

	if err := grpcServer.Serve(lis); err != nil {
		return nil, err
	}
	s.logger.Info("gRPC server stopped")

	return grpcServer, nil
}
