package app

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GRPCServer 是 gRPC 服务器的封装
type GRPCServer struct {
	addr               string
	server             *grpc.Server
	logger             *zap.Logger
	options            []grpc.ServerOption
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
}

// NewGRPCServer 创建一个新的 gRPC 服务器
func NewGRPCServer(addr string, logger *zap.Logger, opts ...ServerOption) *GRPCServer {
	s := &GRPCServer{
		addr:               addr,
		logger:             logger,
		options:            make([]grpc.ServerOption, 0),
		unaryInterceptors:  make([]grpc.UnaryServerInterceptor, 0),
		streamInterceptors: make([]grpc.StreamServerInterceptor, 0),
	}

	// 应用选项
	for _, opt := range opts {
		opt(s)
	}

	// 创建服务器选项
	if len(s.unaryInterceptors) > 0 {
		s.options = append(s.options, grpc.ChainUnaryInterceptor(s.unaryInterceptors...))
	}
	if len(s.streamInterceptors) > 0 {
		s.options = append(s.options, grpc.ChainStreamInterceptor(s.streamInterceptors...))
	}

	// 创建 gRPC 服务器
	s.server = grpc.NewServer(s.options...)

	return s
}

// Server 返回底层的 gRPC 服务器
func (s *GRPCServer) Server() *grpc.Server {
	return s.server
}

// Start 启动 gRPC 服务器
func (s *GRPCServer) Start(ctx context.Context) error {
	// 注册反射服务
	s.logger.Info("registering gRPC reflection service")
	reflection.Register(s.server)

	// 创建错误通道
	errChan := make(chan error, 1)

	// 启动 gRPC 服务器
	go func() {
		s.logger.Info("starting gRPC server", zap.String("addr", s.addr))
		listener, err := net.Listen("tcp", s.addr)
		if err != nil {
			errChan <- fmt.Errorf("failed to listen on %s: %w", s.addr, err)
			return
		}
		if err := s.server.Serve(listener); err != nil {
			errChan <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	// 等待上下文取消或错误
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		s.logger.Info("shutting down gRPC server")
		s.server.GracefulStop()
		return nil
	}
}
