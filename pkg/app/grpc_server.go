package app

import (
	"context"
	"fmt"
	"net"

	"github.com/costa92/go-protoc/pkg/log"
	"google.golang.org/grpc"
)

// GRPCServer 是对 grpc.Server 的包装，实现了 Server 接口
type GRPCServer struct {
	server   *grpc.Server
	listener net.Listener
	name     string
}

// NewGRPCServer 创建一个新的 GRPCServer 实例
func NewGRPCServer(name string, listener net.Listener, opts ...grpc.ServerOption) *GRPCServer {
	return &GRPCServer{
		server:   grpc.NewServer(opts...),
		listener: listener,
		name:     name,
	}
}

// Server 返回底层的 grpc.Server 实例
func (s *GRPCServer) Server() *grpc.Server {
	return s.server
}

// Start 实现 Server 接口的 Start 方法
func (s *GRPCServer) Start(ctx context.Context) error {
	log.Infof("gRPC 服务器 %s 正在监听 %s", s.name, s.listener.Addr().String())

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
	log.Infof("正在关闭 gRPC 服务器 %s", s.name)

	// 创建一个通道来跟踪 GracefulStop 的完成
	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	// 等待 gRPC 服务器关闭或超时
	select {
	case <-ctx.Done():
		log.Warnf("gRPC 服务器 %s 优雅关闭超时，强制停止", s.name)
		s.server.Stop()
		return fmt.Errorf("gRPC 服务器 %s 优雅关闭超时", s.name)
	case <-done:
		log.Infof("gRPC 服务器 %s 已成功关闭", s.name)
		return nil
	}
}
