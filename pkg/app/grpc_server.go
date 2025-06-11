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
	server   *grpc.Server
	listener net.Listener
	name     string
	opts     []grpc.ServerOption // 保存服务器选项，用于动态添加拦截器
}

// NewGRPCServer 创建一个新的 GRPCServer 实例
func NewGRPCServer(name string, listener net.Listener, opts ...grpc.ServerOption) *GRPCServer {
	return &GRPCServer{
		server:   grpc.NewServer(opts...),
		listener: listener,
		name:     name,
		opts:     opts,
	}
}

// Server 返回底层的 grpc.Server 实例
func (s *GRPCServer) Server() *grpc.Server {
	return s.server
}

// AddUnaryServerInterceptors 添加 gRPC 一元拦截器
// 注意：此方法必须在 Start 方法调用前使用，否则不会生效
func (s *GRPCServer) AddUnaryServerInterceptors(interceptors ...grpc.UnaryServerInterceptor) {
	// 由于 gRPC 服务器一旦创建就无法修改拦截器，这里我们采用一种变通方法
	// 如果 server 已经启动，记录一个警告日志
	logger.Infow("添加 gRPC 一元拦截器")

	// 创建新的 gRPC 服务器，包含原有的选项和新的拦截器
	for _, interceptor := range interceptors {
		// 创建一个包装拦截器，它将调用之前的拦截器和新的拦截器
		s.opts = append(s.opts, grpc.UnaryInterceptor(interceptor))
	}

	// 重新创建服务器
	s.server = grpc.NewServer(s.opts...)
}

// AddStreamServerInterceptors 添加 gRPC 流式拦截器
// 注意：此方法必须在 Start 方法调用前使用，否则不会生效
func (s *GRPCServer) AddStreamServerInterceptors(interceptors ...grpc.StreamServerInterceptor) {
	// 与添加一元拦截器类似
	logger.Infow("添加 gRPC 流式拦截器")

	// 创建新的 gRPC 服务器，包含原有的选项和新的拦截器
	for _, interceptor := range interceptors {
		s.opts = append(s.opts, grpc.StreamInterceptor(interceptor))
	}

	// 重新创建服务器
	s.server = grpc.NewServer(s.opts...)
}

// Start 实现 Server 接口的 Start 方法
func (s *GRPCServer) Start(ctx context.Context) error {
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
