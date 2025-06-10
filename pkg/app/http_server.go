package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"sync"
	"time"

	"github.com/costa92/go-protoc/pkg/log"
	"github.com/costa92/go-protoc/pkg/response"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// HTTPServer 是对 http.Server 的包装，实现了 Server 接口
type HTTPServer struct {
	*http.Server
	router       *mux.Router
	gatewayMux   *runtime.ServeMux
	name         string
	mu           sync.Mutex // 保护路由注册的并发安全
	gatewayAdded bool       // 标记是否已添加 gRPC-Gateway 作为默认处理器
}

// NewHTTPServer 创建一个新的 HTTPServer 实例
func NewHTTPServer(name, addr string, middlewares ...mux.MiddlewareFunc) *HTTPServer {
	router := mux.NewRouter()

	// 应用中间件
	for _, mw := range middlewares {
		router.Use(mw)
	}

	// 创建 gRPC-Gateway mux
	gwmux := runtime.NewServeMux()
	response.Setup(gwmux)

	httpServer := &HTTPServer{
		Server: &http.Server{
			Addr:              addr,
			Handler:           router,
			ReadHeaderTimeout: 60 * time.Second,
		},
		router:       router,
		gatewayMux:   gwmux,
		name:         name,
		gatewayAdded: false,
	}

	// 注册健康检查和调试路由
	httpServer.registerDebugHandlers()

	return httpServer
}

// Router 返回 mux.Router 实例
func (s *HTTPServer) Router() *mux.Router {
	return s.router
}

// GatewayMux 返回 gRPC-Gateway ServeMux 实例
func (s *HTTPServer) GatewayMux() *runtime.ServeMux {
	return s.gatewayMux
}

// AddRoute 添加一个新的 HTTP 路由
// 此方法确保路由在 gRPC-Gateway 的 catch-all 路由之前添加
// path: 路由路径
// handler: HTTP 处理函数
// methods: HTTP 方法 (GET, POST, PUT, DELETE 等)
func (s *HTTPServer) AddRoute(path string, handler http.HandlerFunc, methods ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果已经添加了 gRPC-Gateway 作为默认处理器，警告用户
	if s.gatewayAdded {
		log.Warnw("尝试在 gRPC-Gateway 默认处理器之后添加路由，这可能导致路由无法访问", "path", path)
		// 我们需要移除之前的 catch-all 路由并在添加新路由后重新添加它
		// 但这在 gorilla/mux 中并不容易实现，因此只是发出警告
	}

	// 添加路由
	route := s.router.HandleFunc(path, handler)
	if len(methods) > 0 {
		route.Methods(methods...)
	}

	log.Infow("已添加自定义路由", "path", path, "methods", methods)
}

// FinalizeRoutes 在所有自定义路由添加完毕后调用，注册 gRPC-Gateway 作为默认处理器
func (s *HTTPServer) FinalizeRoutes() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.gatewayAdded {
		// 注册 gRPC-Gateway 路由作为默认处理器（始终放在最后）
		s.router.PathPrefix("/").Handler(s.gatewayMux)
		s.gatewayAdded = true
		log.Infow("已注册 gRPC-Gateway 作为默认处理器")
	} else {
		log.Warnw("尝试重复注册 gRPC-Gateway 作为默认处理器，操作被忽略")
	}
}

// Start 实现 Server 接口的 Start 方法
func (s *HTTPServer) Start(ctx context.Context) error {
	// 确保在启动前 gRPC-Gateway 已注册为默认处理器
	if !s.gatewayAdded {
		s.FinalizeRoutes()
	}

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

// registerDebugHandlers 注册调试处理器
func (s *HTTPServer) registerDebugHandlers() {
	// 注册健康检查路由
	s.AddRoute("/healthz", s.handleHealthCheck, "GET")

	// 注册 pprof 路由
	// 注意：使用 gorilla/mux 注册 pprof 路由需要单独为每个处理器注册路由
	s.AddRoute("/debug/pprof/", pprof.Index)
	s.AddRoute("/debug/pprof/cmdline", pprof.Cmdline)
	s.AddRoute("/debug/pprof/profile", pprof.Profile)
	s.AddRoute("/debug/pprof/symbol", pprof.Symbol)
	s.AddRoute("/debug/pprof/trace", pprof.Trace)

	// 添加堆、goroutine、线程创建、块分析等分析器
	s.router.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	s.router.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	s.router.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	s.router.Handle("/debug/pprof/block", pprof.Handler("block"))
	s.router.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))

	// 添加 allocs 分析器的直接支持
	// allocs 实际上是 heap 分析器的一种视图
	s.AddRoute("/debug/pprof/allocs", func(w http.ResponseWriter, r *http.Request) {
		// 复制请求并添加参数
		r2 := new(http.Request)
		*r2 = *r
		q := r2.URL.Query()
		q.Set("gc", "1") // 触发 GC
		r2.URL.RawQuery = q.Encode()
		pprof.Handler("allocs").ServeHTTP(w, r2)
	})

	log.Infow("已注册 pprof 调试路由", "path", "/debug/pprof/")
}

// handleHealthCheck 处理健康检查请求
func (s *HTTPServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	log.Infow("health check")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
