package app

import (
	"context"
	"fmt"
	"net"

	"github.com/costa92/go-protoc/pkg/logger"
	"google.golang.org/grpc"
)

// GRPCServer 是对 grpc.Server 的包装，实现了 Server 接口
type GRPCServer struct {
	server             *grpc.Server
	listener           net.Listener
	name               string
	baseOpts           []grpc.ServerOption // 基础服务器选项
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
	isBuilt            bool // 标记服务器是否已经构建完成
}

// NewGRPCServer 创建一个新的 GRPCServer 实例
func NewGRPCServer(name string, listener net.Listener, opts ...grpc.ServerOption) *GRPCServer {
	return &GRPCServer{
		listener:           listener,
		name:               name,
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
	s.isBuilt = true
	logger.Infow("gRPC 服务器构建完成",
		"unary_interceptors", len(s.unaryInterceptors),
		"stream_interceptors", len(s.streamInterceptors))
}

// AddUnaryServerInterceptors 添加 gRPC 一元拦截器
func (s *GRPCServer) AddUnaryServerInterceptors(interceptors ...grpc.UnaryServerInterceptor) {
	if s.isBuilt {
		logger.Warnw("gRPC 服务器已构建，无法添加拦截器")
		return
	}

	logger.Infow("添加 gRPC 一元拦截器", "count", len(interceptors))
	s.unaryInterceptors = append(s.unaryInterceptors, interceptors...)
}

// AddStreamServerInterceptors 添加 gRPC 流式拦截器
func (s *GRPCServer) AddStreamServerInterceptors(interceptors ...grpc.StreamServerInterceptor) {
	if s.isBuilt {
		logger.Warnw("gRPC 服务器已构建，无法添加拦截器")
		return
	}

	logger.Infow("添加 gRPC 流式拦截器", "count", len(interceptors))
	s.streamInterceptors = append(s.streamInterceptors, interceptors...)
}

// chainUnaryInterceptors 将多个一元拦截器链接在一起
func chainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return buildChain(interceptors, handler)(ctx, req, info)
	}
}

// buildChain 构建拦截器链
func buildChain(interceptors []grpc.UnaryServerInterceptor, handler grpc.UnaryHandler) func(context.Context, interface{}, *grpc.UnaryServerInfo) (interface{}, error) {
	if len(interceptors) == 0 {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo) (interface{}, error) {
			return handler(ctx, req)
		}
	}

	// 从最后一个拦截器开始构建链
	chainedHandler := handler
	for i := len(interceptors) - 1; i >= 0; i-- {
		interceptor := interceptors[i]
		currentHandler := chainedHandler
		chainedHandler = func(ctx context.Context, req interface{}) (interface{}, error) {
			return interceptor(ctx, req, nil, currentHandler)
		}
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo) (interface{}, error) {
		// 重新包装最外层拦截器以传递正确的 info
		return interceptors[0](ctx, req, info, func(ctx context.Context, req interface{}) (interface{}, error) {
			if len(interceptors) == 1 {
				return handler(ctx, req)
			}
			return buildChain(interceptors[1:], handler)(ctx, req, info)
		})
	}
}

// chainStreamInterceptors 将多个流式拦截器链接在一起
func chainStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return buildStreamChain(interceptors, handler)(srv, ss, info)
	}
}

// buildStreamChain 构建流式拦截器链
func buildStreamChain(interceptors []grpc.StreamServerInterceptor, handler grpc.StreamHandler) func(interface{}, grpc.ServerStream, *grpc.StreamServerInfo) error {
	if len(interceptors) == 0 {
		return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo) error {
			return handler(srv, ss)
		}
	}

	// 从最后一个拦截器开始构建链
	chainedHandler := handler
	for i := len(interceptors) - 1; i >= 0; i-- {
		interceptor := interceptors[i]
		currentHandler := chainedHandler
		chainedHandler = func(srv interface{}, ss grpc.ServerStream) error {
			return interceptor(srv, ss, nil, currentHandler)
		}
	}

	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo) error {
		// 重新包装最外层拦截器以传递正确的 info
		return interceptors[0](srv, ss, info, func(srv interface{}, ss grpc.ServerStream) error {
			if len(interceptors) == 1 {
				return handler(srv, ss)
			}
			return buildStreamChain(interceptors[1:], handler)(srv, ss, info)
		})
	}
}

// Start 实现 Server 接口的 Start 方法
func (s *GRPCServer) Start(ctx context.Context) error {
	// 确保服务器已构建
	if !s.isBuilt {
		s.buildServer()
	}

	logger.Infof("gRPC 服务器 %s 正在监听 %s", s.name, s.listener.Addr().String())

	// 创建一个 channel 用于接收服务器退出信号
	errCh := make(chan error, 1)

	// 在后台启动 gRPC 服务器
	go func() {
		if err := s.server.Serve(s.listener); err != nil {
			errCh <- fmt.Errorf("gRPC 服务器 %s 失败: %v", s.name, err)
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
	logger.Infof("正在关闭 gRPC 服务器 %s", s.name)

	// 创建一个通道来跟踪 GracefulStop 的完成
	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	// 等待 gRPC 服务器关闭或超时
	select {
	case <-ctx.Done():
		logger.Warnf("gRPC 服务器 %s 优雅关闭超时，强制停止", s.name)
		s.server.Stop()
		return fmt.Errorf("gRPC 服务器 %s 优雅关闭超时", s.name)
	case <-done:
		logger.Infof("gRPC 服务器 %s 已成功关闭", s.name)
		return nil
	}
}
