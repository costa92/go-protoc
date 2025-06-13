package app

import (
	"context"
	"fmt"
	"net"

	"github.com/costa92/go-protoc/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GRPCServer 是对 grpc.Server 的包装，实现了 Server 接口
type GRPCServer struct {
	server             *grpc.Server
	listener           net.Listener
	baseOpts           []grpc.ServerOption // 基础服务器选项
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
	isBuilt            bool // 标记服务器是否已经构建完成
}

// NewGRPCServer 创建一个新的 GRPCServer 实例
func NewGRPCServer(addr string, opts ...grpc.ServerOption) *GRPCServer {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatalf("Failed to listen: %v", err)
	}
	return &GRPCServer{
		listener:           listener,
		baseOpts:           opts,
		unaryInterceptors:  make([]grpc.UnaryServerInterceptor, 0),
		streamInterceptors: make([]grpc.StreamServerInterceptor, 0),
		isBuilt:            false,
	}
}

// Server 返回底层的 grpc.Server 实例
func (s *GRPCServer) Server() *grpc.Server {
	if !s.isBuilt {
		s.buildServer()
	}
	return s.server
}

// buildServer 构建 gRPC 服务器，组合所有拦截器
func (s *GRPCServer) buildServer() {
	if s.isBuilt {
		return
	}

	opts := make([]grpc.ServerOption, len(s.baseOpts))
	copy(opts, s.baseOpts)

	// 创建拦截器链
	if len(s.unaryInterceptors) > 0 {
		chainedUnaryInterceptor := chainUnaryInterceptors(s.unaryInterceptors...)
		opts = append(opts, grpc.UnaryInterceptor(chainedUnaryInterceptor))
	}

	if len(s.streamInterceptors) > 0 {
		chainedStreamInterceptor := chainStreamInterceptors(s.streamInterceptors...)
		opts = append(opts, grpc.StreamInterceptor(chainedStreamInterceptor))
	}

	s.server = grpc.NewServer(opts...)

	// 启用 gRPC 反射服务
	reflection.Register(s.server)

	s.isBuilt = true
	logger.Infow("gRPC 服务器构建完成",
		"unary_interceptors", len(s.unaryInterceptors),
		"stream_interceptors", len(s.streamInterceptors))
}

// AddUnaryServerInterceptors 添加一元拦截器
func (s *GRPCServer) AddUnaryServerInterceptors(interceptors ...grpc.UnaryServerInterceptor) {
	s.unaryInterceptors = append(s.unaryInterceptors, interceptors...)
}

// AddStreamServerInterceptors 添加流式拦截器
func (s *GRPCServer) AddStreamServerInterceptors(interceptors ...grpc.StreamServerInterceptor) {
	s.streamInterceptors = append(s.streamInterceptors, interceptors...)
}

// chainUnaryInterceptors 将多个一元拦截器链接成一个
func chainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	n := len(interceptors)

	if n == 0 {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	}

	if n == 1 {
		return interceptors[0]
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var next grpc.UnaryHandler
		var current int

		next = func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
			if current == n-1 {
				return handler(currentCtx, currentReq)
			}
			current++
			return interceptors[current](currentCtx, currentReq, info, next)
		}

		return interceptors[0](ctx, req, info, next)
	}
}

// chainStreamInterceptors 将多个流式拦截器链接成一个
func chainStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	n := len(interceptors)

	if n == 0 {
		return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			return handler(srv, ss)
		}
	}

	if n == 1 {
		return interceptors[0]
	}

	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var next grpc.StreamHandler
		var current int

		next = func(currentSrv interface{}, currentStream grpc.ServerStream) error {
			if current == n-1 {
				return handler(currentSrv, currentStream)
			}
			current++
			return interceptors[current](currentSrv, currentStream, info, next)
		}

		return interceptors[0](srv, ss, info, next)
	}
}

// Start 实现 Server 接口的 Start 方法
func (s *GRPCServer) Start(ctx context.Context) error {
	// 确保服务器已构建
	if !s.isBuilt {
		s.buildServer()
	}

	logger.Infow("gRPC 服务器正在监听", "addr", s.listener.Addr().String())

	// 创建一个 channel 用于接收服务器退出信号
	errCh := make(chan error, 1)

	// 在后台启动 gRPC 服务器
	go func() {
		if err := s.server.Serve(s.listener); err != nil {
			errCh <- fmt.Errorf("gRPC 服务器 %s 失败: %v", s.listener.Addr().String(), err)
		}
		close(errCh)
	}()

	// 等待上下文取消或服务器错误
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Stop 实现 Server 接口的 Stop 方法
func (s *GRPCServer) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	// 优雅停止服务器
	stopped := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		s.server.Stop()
		return ctx.Err()
	case <-stopped:
		return nil
	}
}
