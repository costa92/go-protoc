package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/costa92/go-protoc/pkg/log"
	"github.com/gorilla/mux"
)

// HTTPServer 是对 http.Server 的包装，实现了 Server 接口
type HTTPServer struct {
	*http.Server
	router *mux.Router
	name   string
}

// NewHTTPServer 创建一个新的 HTTPServer 实例
func NewHTTPServer(name, addr string, middlewares ...mux.MiddlewareFunc) *HTTPServer {
	router := mux.NewRouter()
	for _, mw := range middlewares {
		router.Use(mw)
	}

	return &HTTPServer{
		Server: &http.Server{
			Addr:              addr,
			Handler:           router,
			ReadHeaderTimeout: 60 * time.Second,
		},
		router: router,
		name:   name,
	}
}

// Router 返回 mux.Router 实例
func (s *HTTPServer) Router() *mux.Router {
	return s.router
}

// Start 实现 Server 接口的 Start 方法
func (s *HTTPServer) Start(ctx context.Context) error {
	log.Infof("HTTP 服务器 %s 正在监听 %s", s.name, s.Addr)

	// 在后台启动 HTTP 服务器
	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("HTTP 服务器 %s 失败: %v", s.name, err)
		}
	}()

	// 等待上下文取消
	<-ctx.Done()
	return ctx.Err()
}

// Stop 实现 Server 接口的 Stop 方法
func (s *HTTPServer) Stop(ctx context.Context) error {
	log.Infof("正在关闭 HTTP 服务器 %s", s.name)
	if err := s.Shutdown(ctx); err != nil {
		return fmt.Errorf("HTTP 服务器 %s 关闭失败: %v", s.name, err)
	}
	log.Infof("HTTP 服务器 %s 已成功关闭", s.name)
	return nil
}
