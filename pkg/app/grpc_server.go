package app

import (
	"net"

	"google.golang.org/grpc"
)

type GRPCServerOption func(*GRPCServer) func(*grpc.Server)

func WithGRPCServer(grpcServer *grpc.Server) GRPCServerOption {
	return func(s *GRPCServer) func(*grpc.Server) {
		return func(gs *grpc.Server) {
			// 这里可以添加自定义的 grpc.Server 配置逻辑
		}
	}
}

func NewGRPCServer(options ...GRPCServerOption) *GRPCServer {
	return &GRPCServer{
		options: options,
	}
}

// 定义结构体来保存服务实例
type GRPCServer struct {
	options []GRPCServerOption
}

// RunGRPCServer runs gRPC service to publish ToDo service
func (s *GRPCServer) RunGRPCServer(grpcAddr string) (*grpc.Server, error) {
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return nil, err
	}
	grpcServer := grpc.NewServer()
	// 应用选项
	for _, opt := range s.options {
		apply := opt(s)
		apply(grpcServer)
	}

	if err := grpcServer.Serve(lis); err != nil {
		return nil, err
	}

	return grpcServer, nil
}
