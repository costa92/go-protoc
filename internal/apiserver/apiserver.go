package apiserver

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/costa92/go-protoc/pkg/app"
	"github.com/costa92/go-protoc/pkg/logger"
	genericoptions "github.com/costa92/go-protoc/pkg/options"
)

// Config contains application-related configurations.
type Config struct {
	GRPCOptions   *genericoptions.GRPCOptions
	HTTPOptions   *genericoptions.HTTPOptions
}

// ErrServerNotInitialized 定义了服务器未初始化的错误
var ErrServerNotInitialized = errors.New("server not initialized")

// Server 定义了 API 服务器的核心结构
type Server struct {
	app        *app.App
	config     *Config
	grpcServer *app.GRPCServer
	httpServer *app.HTTPServer
	stopCh     chan struct{}
	mu         sync.Mutex
	isRunning  bool
}





// NewServer 创建并返回一个新的 Server 实例
func (c *Config) NewServer(ctx context.Context) (*Server, error) {
	s := &Server{
		config:    c,
		isRunning: false,
	}
	return s, nil
}

// Run 启动服务器
func (s *Server) Run(ctx context.Context) error {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return errors.New("server is already running")
	}
	s.isRunning = true
	s.mu.Unlock()

	// 创建错误通道来收集潜在的错误
	errChan := make(chan error, 2)

	// 初始化 GRPC 服务器
	s.grpcServer = app.NewGRPCServer(s.config.GRPCOptions.Addr)
	// 初始化 HTTP 服务器
	s.httpServer = app.NewHTTPServer(s.config.HTTPOptions.Addr)

	// 启动 GRPC 服务
	go func() {
		logger.Infof("Starting GRPC server on %s", s.config.GRPCOptions.Addr)
		if err := s.grpcServer.Start(ctx); err != nil {
			errChan <- fmt.Errorf("failed to run grpc server: %w", err)
		}
	}()

	// 启动 HTTP 服务
	go func() {
		logger.Infof("Starting HTTP server on %s", s.config.HTTPOptions.Addr)
		if err := s.httpServer.Start(ctx); err != nil {
			errChan <- fmt.Errorf("failed to run http server: %w", err)
		}
	}()

	// 等待上下文取消或者出现错误
	select {
	case <-ctx.Done():
		return s.Shutdown()
	case err := <-errChan:
		return err
	}
}

// Shutdown 优雅地关闭服务器
func (s *Server) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	close(s.stopCh)
	s.isRunning = false

	var errs []error

	// 创建超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭 GRPC 服务器
	if s.grpcServer != nil {
		if err := s.grpcServer.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("error shutting down grpc server: %w", err))
		}
	}

	// 关闭 HTTP 服务器
	if s.httpServer != nil {
		if err := s.httpServer.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("error shutting down http server: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	
	return nil
}
