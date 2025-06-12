package middleware

import (
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

// MiddlewareType 定义中间件类型
type MiddlewareType string

const (
	// HTTPMiddleware HTTP 中间件类型
	HTTPMiddleware MiddlewareType = "http"
	// GRPCUnaryMiddleware gRPC 一元中间件类型
	GRPCUnaryMiddleware MiddlewareType = "grpc_unary"
	// GRPCStreamMiddleware gRPC 流式中间件类型
	GRPCStreamMiddleware MiddlewareType = "grpc_stream"
)

// Middleware 定义中间件接口
type Middleware interface {
	// Type 返回中间件类型
	Type() MiddlewareType
	// Name 返回中间件名称
	Name() string
	// Priority 返回中间件优先级，数字越小优先级越高
	Priority() int
	// Enabled 返回中间件是否启用
	Enabled() bool
}

// HTTPMiddlewareFunc 定义 HTTP 中间件函数
type HTTPMiddlewareFunc struct {
	name     string
	priority int
	enabled  bool
	handler  mux.MiddlewareFunc
}

func (m *HTTPMiddlewareFunc) Type() MiddlewareType        { return HTTPMiddleware }
func (m *HTTPMiddlewareFunc) Name() string                { return m.name }
func (m *HTTPMiddlewareFunc) Priority() int               { return m.priority }
func (m *HTTPMiddlewareFunc) Enabled() bool               { return m.enabled }
func (m *HTTPMiddlewareFunc) Handler() mux.MiddlewareFunc { return m.handler }

// GRPCUnaryMiddlewareFunc 定义 gRPC 一元中间件函数
type GRPCUnaryMiddlewareFunc struct {
	name     string
	priority int
	enabled  bool
	handler  grpc.UnaryServerInterceptor
}

func (m *GRPCUnaryMiddlewareFunc) Type() MiddlewareType                 { return GRPCUnaryMiddleware }
func (m *GRPCUnaryMiddlewareFunc) Name() string                         { return m.name }
func (m *GRPCUnaryMiddlewareFunc) Priority() int                        { return m.priority }
func (m *GRPCUnaryMiddlewareFunc) Enabled() bool                        { return m.enabled }
func (m *GRPCUnaryMiddlewareFunc) Handler() grpc.UnaryServerInterceptor { return m.handler }

// GRPCStreamMiddlewareFunc 定义 gRPC 流式中间件函数
type GRPCStreamMiddlewareFunc struct {
	name     string
	priority int
	enabled  bool
	handler  grpc.StreamServerInterceptor
}

func (m *GRPCStreamMiddlewareFunc) Type() MiddlewareType                  { return GRPCStreamMiddleware }
func (m *GRPCStreamMiddlewareFunc) Name() string                          { return m.name }
func (m *GRPCStreamMiddlewareFunc) Priority() int                         { return m.priority }
func (m *GRPCStreamMiddlewareFunc) Enabled() bool                         { return m.enabled }
func (m *GRPCStreamMiddlewareFunc) Handler() grpc.StreamServerInterceptor { return m.handler }

// Manager 定义中间件管理器
type Manager struct {
	middlewares []Middleware
}

// NewManager 创建新的中间件管理器
func NewManager() *Manager {
	return &Manager{
		middlewares: make([]Middleware, 0),
	}
}

// Add 添加中间件
func (m *Manager) Add(middleware Middleware) {
	m.middlewares = append(m.middlewares, middleware)
}

// GetHTTPMiddlewares 获取已启用的 HTTP 中间件，按优先级排序
func (m *Manager) GetHTTPMiddlewares() []mux.MiddlewareFunc {
	var result []mux.MiddlewareFunc
	for _, mw := range m.getSortedMiddlewares(HTTPMiddleware) {
		if httpMw, ok := mw.(*HTTPMiddlewareFunc); ok {
			result = append(result, httpMw.Handler())
		}
	}
	return result
}

// GetGRPCUnaryMiddlewares 获取已启用的 gRPC 一元中间件，按优先级排序
func (m *Manager) GetGRPCUnaryMiddlewares() []grpc.UnaryServerInterceptor {
	var result []grpc.UnaryServerInterceptor
	for _, mw := range m.getSortedMiddlewares(GRPCUnaryMiddleware) {
		if grpcMw, ok := mw.(*GRPCUnaryMiddlewareFunc); ok {
			result = append(result, grpcMw.Handler())
		}
	}
	return result
}

// GetGRPCStreamMiddlewares 获取已启用的 gRPC 流式中间件，按优先级排序
func (m *Manager) GetGRPCStreamMiddlewares() []grpc.StreamServerInterceptor {
	var result []grpc.StreamServerInterceptor
	for _, mw := range m.getSortedMiddlewares(GRPCStreamMiddleware) {
		if grpcMw, ok := mw.(*GRPCStreamMiddlewareFunc); ok {
			result = append(result, grpcMw.Handler())
		}
	}
	return result
}

// getSortedMiddlewares 获取指定类型的已启用中间件，并按优先级排序
func (m *Manager) getSortedMiddlewares(typ MiddlewareType) []Middleware {
	var result []Middleware
	for _, mw := range m.middlewares {
		if mw.Type() == typ && mw.Enabled() {
			result = append(result, mw)
		}
	}
	// 按优先级排序
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Priority() > result[j].Priority() {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result
}

// NewHTTPMiddleware 创建新的 HTTP 中间件
func NewHTTPMiddleware(name string, priority int, enabled bool, handler mux.MiddlewareFunc) *HTTPMiddlewareFunc {
	return &HTTPMiddlewareFunc{
		name:     name,
		priority: priority,
		enabled:  enabled,
		handler:  handler,
	}
}

// NewGRPCUnaryMiddleware 创建新的 gRPC 一元中间件
func NewGRPCUnaryMiddleware(name string, priority int, enabled bool, handler grpc.UnaryServerInterceptor) *GRPCUnaryMiddlewareFunc {
	return &GRPCUnaryMiddlewareFunc{
		name:     name,
		priority: priority,
		enabled:  enabled,
		handler:  handler,
	}
}

// NewGRPCStreamMiddleware 创建新的 gRPC 流式中间件
func NewGRPCStreamMiddleware(name string, priority int, enabled bool, handler grpc.StreamServerInterceptor) *GRPCStreamMiddlewareFunc {
	return &GRPCStreamMiddlewareFunc{
		name:     name,
		priority: priority,
		enabled:  enabled,
		handler:  handler,
	}
}
