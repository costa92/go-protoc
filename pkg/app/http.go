package app

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// HTTPServer 是 HTTP 服务器的封装
type HTTPServer struct {
	addr        string
	handler     http.Handler
	router      *mux.Router
	mainHandler http.Handler
	logger      *zap.Logger
	middlewares []mux.MiddlewareFunc
}

// NewHTTPServer 创建一个新的 HTTP 服务器
func NewHTTPServer(addr string, logger *zap.Logger, opts ...ServerOption) *HTTPServer {
	s := &HTTPServer{
		addr:        addr,
		router:      mux.NewRouter(),
		logger:      logger,
		middlewares: make([]mux.MiddlewareFunc, 0),
	}

	// 应用选项
	for _, opt := range opts {
		opt(s)
	}

	// 应用中间件
	for _, middleware := range s.middlewares {
		s.router.Use(middleware)
	}

	return s
}

// Router 返回路由器
func (s *HTTPServer) Router() *mux.Router {
	return s.router
}

// registerDebugHandlers 注册调试处理器
func (s *HTTPServer) registerDebugHandlers() {
	// 创建一个新的 ServeMux 用于 pprof
	pprofMux := http.NewServeMux()
	pprofMux.HandleFunc("/debug/pprof/", pprof.Index)
	pprofMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	pprofMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	pprofMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	pprofMux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// 创建一个新的 ServeMux 作为主处理器
	mainHandler := http.NewServeMux()

	// 将 API 路由器挂载到 /api 路径下
	mainHandler.Handle("/api/", http.StripPrefix("/api", s.router))

	// 将 pprof 处理器直接挂载到主处理器
	mainHandler.Handle("/debug/pprof/", pprofMux)

	// 设置主处理器
	s.mainHandler = mainHandler

	s.logger.Info("registered debug handlers", zap.String("path", "/debug/pprof/*"))
}

// Start 启动 HTTP 服务器
func (s *HTTPServer) Start(ctx context.Context) error {
	// 注册调试处理器
	s.registerDebugHandlers()

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    s.addr,
		Handler: s.mainHandler,
	}

	// 创建错误通道
	errChan := make(chan error, 1)

	// 在一个新的 goroutine 中启动服务器
	go func() {
		s.logger.Info("starting HTTP server", zap.String("addr", s.addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// 等待上下文取消或错误
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		// 创建一个关闭超时上下文
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// 优雅关闭 HTTP 服务器
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("HTTP server shutdown error: %w", err)
		}
		return nil
	}
}
