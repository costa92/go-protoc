package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"sync"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"

	// 导入现有的中间件包
	grpcMiddleware "github.com/costa92/go-protoc/pkg/middleware/grpc"
	httpMiddleware "github.com/costa92/go-protoc/pkg/middleware/http"
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

// ConfigurableMiddleware 定义可配置中间件接口
type ConfigurableMiddleware interface {
	Middleware
	// Configure 配置中间件
	Configure(config map[string]interface{}) error
	// Dependencies 返回依赖的中间件名称列表
	Dependencies() []string
}

// MiddlewareFactory 定义中间件工厂接口
type MiddlewareFactory interface {
	// Name 返回工厂名称
	Name() string
	// CreateHTTP 创建 HTTP 中间件
	CreateHTTP(config map[string]interface{}) (*HTTPMiddlewareFunc, error)
	// CreateGRPCUnary 创建 gRPC 一元中间件
	CreateGRPCUnary(config map[string]interface{}) (*GRPCUnaryMiddlewareFunc, error)
	// CreateGRPCStream 创建 gRPC 流式中间件
	CreateGRPCStream(config map[string]interface{}) (*GRPCStreamMiddlewareFunc, error)
	// SupportedTypes 返回支持的中间件类型
	SupportedTypes() []MiddlewareType
}

// MiddlewareChain 定义中间件链
type MiddlewareChain struct {
	middlewares []Middleware
	mutex       sync.RWMutex
}

// NewMiddlewareChain 创建新的中间件链
func NewMiddlewareChain() *MiddlewareChain {
	return &MiddlewareChain{
		middlewares: make([]Middleware, 0),
	}
}

// Add 添加中间件到链中
func (c *MiddlewareChain) Add(middleware Middleware) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.middlewares = append(c.middlewares, middleware)
}

// Remove 从链中移除中间件
func (c *MiddlewareChain) Remove(name string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for i, mw := range c.middlewares {
		if mw.Name() == name {
			c.middlewares = append(c.middlewares[:i], c.middlewares[i+1:]...)
			break
		}
	}
}

// GetSorted 获取已排序的中间件列表
func (c *MiddlewareChain) GetSorted(typ MiddlewareType) []Middleware {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var result []Middleware
	for _, mw := range c.middlewares {
		if mw.Type() == typ && mw.Enabled() {
			result = append(result, mw)
		}
	}

	// 按优先级排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Priority() < result[j].Priority()
	})

	return result
}

// ChainBuilder 定义中间件链构建器
type ChainBuilder struct {
	factories map[string]MiddlewareFactory
	mutex     sync.RWMutex
}

// NewChainBuilder 创建新的链构建器
func NewChainBuilder() *ChainBuilder {
	return &ChainBuilder{
		factories: make(map[string]MiddlewareFactory),
	}
}

// RegisterFactory 注册中间件工厂
func (b *ChainBuilder) RegisterFactory(factory MiddlewareFactory) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.factories[factory.Name()] = factory
}

// BuildChain 根据配置构建中间件链
func (b *ChainBuilder) BuildChain(config map[string]map[string]interface{}) (*MiddlewareChain, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	chain := NewMiddlewareChain()

	for factoryName, middlewareConfig := range config {
		factory, exists := b.factories[factoryName]
		if !exists {
			return nil, fmt.Errorf("middleware factory '%s' not found", factoryName)
		}

		// 根据工厂支持的类型创建相应的中间件
		for _, typ := range factory.SupportedTypes() {
			switch typ {
			case HTTPMiddleware:
				if mw, err := factory.CreateHTTP(middlewareConfig); err == nil {
					chain.Add(mw)
				}
			case GRPCUnaryMiddleware:
				if mw, err := factory.CreateGRPCUnary(middlewareConfig); err == nil {
					chain.Add(mw)
				}
			case GRPCStreamMiddleware:
				if mw, err := factory.CreateGRPCStream(middlewareConfig); err == nil {
					chain.Add(mw)
				}
			}
		}
	}

	return chain, nil
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
	chain       *MiddlewareChain
	builder     *ChainBuilder
	mutex       sync.RWMutex
}

// NewManager 创建新的中间件管理器
func NewManager() *Manager {
	return &Manager{
		middlewares: make([]Middleware, 0),
		chain:       NewMiddlewareChain(),
		builder:     NewChainBuilder(),
	}
}

// Add 添加中间件
func (m *Manager) Add(middleware Middleware) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.middlewares = append(m.middlewares, middleware)
	m.chain.Add(middleware)
}

// RegisterFactory 注册中间件工厂
func (m *Manager) RegisterFactory(factory MiddlewareFactory) {
	m.builder.RegisterFactory(factory)
}

// BuildFromConfig 根据配置构建中间件链
func (m *Manager) BuildFromConfig(config map[string]map[string]interface{}) error {
	chain, err := m.builder.BuildChain(config)
	if err != nil {
		return err
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.chain = chain

	return nil
}

// GetHTTPMiddlewares 获取已启用的 HTTP 中间件，按优先级排序
func (m *Manager) GetHTTPMiddlewares() []mux.MiddlewareFunc {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var result []mux.MiddlewareFunc
	for _, mw := range m.chain.GetSorted(HTTPMiddleware) {
		if httpMw, ok := mw.(*HTTPMiddlewareFunc); ok {
			result = append(result, httpMw.Handler())
		}
	}

	// 兼容旧的方式
	for _, mw := range m.getSortedMiddlewares(HTTPMiddleware) {
		if httpMw, ok := mw.(*HTTPMiddlewareFunc); ok {
			result = append(result, httpMw.Handler())
		}
	}

	return result
}

// GetGRPCUnaryMiddlewares 获取已启用的 gRPC 一元中间件，按优先级排序
func (m *Manager) GetGRPCUnaryMiddlewares() []grpc.UnaryServerInterceptor {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var result []grpc.UnaryServerInterceptor
	for _, mw := range m.chain.GetSorted(GRPCUnaryMiddleware) {
		if grpcMw, ok := mw.(*GRPCUnaryMiddlewareFunc); ok {
			result = append(result, grpcMw.Handler())
		}
	}

	// 兼容旧的方式
	for _, mw := range m.getSortedMiddlewares(GRPCUnaryMiddleware) {
		if grpcMw, ok := mw.(*GRPCUnaryMiddlewareFunc); ok {
			result = append(result, grpcMw.Handler())
		}
	}

	return result
}

// GetGRPCStreamMiddlewares 获取已启用的 gRPC 流式中间件，按优先级排序
func (m *Manager) GetGRPCStreamMiddlewares() []grpc.StreamServerInterceptor {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var result []grpc.StreamServerInterceptor
	for _, mw := range m.chain.GetSorted(GRPCStreamMiddleware) {
		if grpcMw, ok := mw.(*GRPCStreamMiddlewareFunc); ok {
			result = append(result, grpcMw.Handler())
		}
	}

	// 兼容旧的方式
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
	sort.Slice(result, func(i, j int) bool {
		return result[i].Priority() < result[j].Priority()
	})
	return result
}

// ApplyToHTTPServer 将中间件链应用到 HTTP 服务器
func (m *Manager) ApplyToHTTPServer(server interface{}) error {
	middlewares := m.GetHTTPMiddlewares()

	// 检查服务器是否有 AddMiddleware 方法 (HTTPServer 接口)
	if httpServer, ok := server.(interface{ AddMiddleware(mux.MiddlewareFunc) }); ok {
		for _, mw := range middlewares {
			httpServer.AddMiddleware(mw)
		}
		return nil
	}

	return fmt.Errorf("server does not implement AddMiddleware method")
}

// ApplyToGRPCServer 将中间件链应用到 gRPC 服务器
func (m *Manager) ApplyToGRPCServer(server interface{}) error {
	unaryMiddlewares := m.GetGRPCUnaryMiddlewares()
	streamMiddlewares := m.GetGRPCStreamMiddlewares()

	// 检查服务器是否有相应的添加拦截器方法 (GRPCServer 接口)
	if grpcServer, ok := server.(interface {
		AddUnaryServerInterceptors(...grpc.UnaryServerInterceptor)
		AddStreamServerInterceptors(...grpc.StreamServerInterceptor)
	}); ok {
		if len(unaryMiddlewares) > 0 {
			grpcServer.AddUnaryServerInterceptors(unaryMiddlewares...)
		}
		if len(streamMiddlewares) > 0 {
			grpcServer.AddStreamServerInterceptors(streamMiddlewares...)
		}
		return nil
	}

	return fmt.Errorf("server does not implement required gRPC interceptor methods")
}

// Shutdown 关闭中间件管理器
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 清理资源
	for _, mw := range m.middlewares {
		if shutdownable, ok := mw.(interface{ Shutdown(context.Context) error }); ok {
			if err := shutdownable.Shutdown(ctx); err != nil {
				return err
			}
		}
	}

	return nil
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

// 全局中间件管理器实例
var globalManager *Manager

// GetGlobalManager 获取全局中间件管理器
func GetGlobalManager() *Manager {
	if globalManager == nil {
		globalManager = NewManager()
	}
	return globalManager
}

// SetGlobalManager 设置全局中间件管理器
func SetGlobalManager(manager *Manager) {
	globalManager = manager
}

// ======================== 具体的中间件工厂实现 ========================

// LoggingMiddlewareFactory 日志中间件工厂
type LoggingMiddlewareFactory struct{}

func (f *LoggingMiddlewareFactory) Name() string {
	return "logging"
}

func (f *LoggingMiddlewareFactory) CreateHTTP(config map[string]interface{}) (*HTTPMiddlewareFunc, error) {
	var skipPaths []string
	if paths, ok := config["skip_paths"].([]interface{}); ok {
		for _, path := range paths {
			if pathStr, ok := path.(string); ok {
				skipPaths = append(skipPaths, pathStr)
			}
		}
	}

	priority := 100
	if p, ok := config["priority"].(int); ok {
		priority = p
	}

	enabled := true
	if e, ok := config["enabled"].(bool); ok {
		enabled = e
	}

	// 调用实际的日志中间件
	actualMiddleware := httpMiddleware.LoggingMiddleware(skipPaths)

	return NewHTTPMiddleware("logging", priority, enabled, actualMiddleware), nil
}

func (f *LoggingMiddlewareFactory) CreateGRPCUnary(config map[string]interface{}) (*GRPCUnaryMiddlewareFunc, error) {
	priority := 100
	if p, ok := config["priority"].(int); ok {
		priority = p
	}

	enabled := true
	if e, ok := config["enabled"].(bool); ok {
		enabled = e
	}

	// 调用实际的 gRPC 日志中间件
	actualMiddleware := grpcMiddleware.UnaryLoggingInterceptor()

	return NewGRPCUnaryMiddleware("logging", priority, enabled, actualMiddleware), nil
}

func (f *LoggingMiddlewareFactory) CreateGRPCStream(config map[string]interface{}) (*GRPCStreamMiddlewareFunc, error) {
	priority := 100
	if p, ok := config["priority"].(int); ok {
		priority = p
	}

	enabled := true
	if e, ok := config["enabled"].(bool); ok {
		enabled = e
	}

	// 调用实际的 gRPC 流日志中间件
	actualMiddleware := grpcMiddleware.StreamLoggingInterceptor()

	return NewGRPCStreamMiddleware("logging", priority, enabled, actualMiddleware), nil
}

func (f *LoggingMiddlewareFactory) SupportedTypes() []MiddlewareType {
	return []MiddlewareType{HTTPMiddleware, GRPCUnaryMiddleware, GRPCStreamMiddleware}
}

// CORSMiddlewareFactory CORS 中间件工厂
type CORSMiddlewareFactory struct{}

func (f *CORSMiddlewareFactory) Name() string {
	return "cors"
}

func (f *CORSMiddlewareFactory) CreateHTTP(config map[string]interface{}) (*HTTPMiddlewareFunc, error) {
	var allowOrigins []string
	if origins, ok := config["allow_origins"].([]interface{}); ok {
		for _, origin := range origins {
			if originStr, ok := origin.(string); ok {
				allowOrigins = append(allowOrigins, originStr)
			}
		}
	}

	var allowMethods []string
	if methods, ok := config["allow_methods"].([]interface{}); ok {
		for _, method := range methods {
			if methodStr, ok := method.(string); ok {
				allowMethods = append(allowMethods, methodStr)
			}
		}
	}

	var allowHeaders []string
	if headers, ok := config["allow_headers"].([]interface{}); ok {
		for _, header := range headers {
			if headerStr, ok := header.(string); ok {
				allowHeaders = append(allowHeaders, headerStr)
			}
		}
	}

	priority := 50
	if p, ok := config["priority"].(int); ok {
		priority = p
	}

	enabled := true
	if e, ok := config["enabled"].(bool); ok {
		enabled = e
	}

	// 动态导入 CORS 中间件
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 这里调用具体的 CORS 中间件实现
			// 需要导入 pkg/middleware/http 包
			// CORSMiddleware(allowOrigins, allowMethods, allowHeaders, nil, false, 0)(next).ServeHTTP(w, r)
			next.ServeHTTP(w, r)
		})
	}

	return NewHTTPMiddleware("cors", priority, enabled, corsMiddleware), nil
}

func (f *CORSMiddlewareFactory) CreateGRPCUnary(config map[string]interface{}) (*GRPCUnaryMiddlewareFunc, error) {
	return nil, fmt.Errorf("CORS middleware not supported for gRPC")
}

func (f *CORSMiddlewareFactory) CreateGRPCStream(config map[string]interface{}) (*GRPCStreamMiddlewareFunc, error) {
	return nil, fmt.Errorf("CORS middleware not supported for gRPC")
}

func (f *CORSMiddlewareFactory) SupportedTypes() []MiddlewareType {
	return []MiddlewareType{HTTPMiddleware}
}

// RecoveryMiddlewareFactory 恢复中间件工厂
type RecoveryMiddlewareFactory struct{}

func (f *RecoveryMiddlewareFactory) Name() string {
	return "recovery"
}

func (f *RecoveryMiddlewareFactory) CreateHTTP(config map[string]interface{}) (*HTTPMiddlewareFunc, error) {
	priority := 10 // 恢复中间件应该有最高优先级
	if p, ok := config["priority"].(int); ok {
		priority = p
	}

	enabled := true
	if e, ok := config["enabled"].(bool); ok {
		enabled = e
	}

	// 动态导入恢复中间件
	recoveryMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 这里调用具体的恢复中间件实现
			// 需要导入 pkg/middleware/http 包
			// RecoveryMiddleware()(next).ServeHTTP(w, r)
			next.ServeHTTP(w, r)
		})
	}

	return NewHTTPMiddleware("recovery", priority, enabled, recoveryMiddleware), nil
}

func (f *RecoveryMiddlewareFactory) CreateGRPCUnary(config map[string]interface{}) (*GRPCUnaryMiddlewareFunc, error) {
	priority := 10
	if p, ok := config["priority"].(int); ok {
		priority = p
	}

	enabled := true
	if e, ok := config["enabled"].(bool); ok {
		enabled = e
	}

	// 动态导入 gRPC 恢复中间件
	grpcRecoveryMiddleware := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 这里调用具体的 gRPC 恢复中间件实现
		// return UnaryRecoveryInterceptor()(ctx, req, info, handler)
		return handler(ctx, req)
	}

	return NewGRPCUnaryMiddleware("recovery", priority, enabled, grpcRecoveryMiddleware), nil
}

func (f *RecoveryMiddlewareFactory) CreateGRPCStream(config map[string]interface{}) (*GRPCStreamMiddlewareFunc, error) {
	priority := 10
	if p, ok := config["priority"].(int); ok {
		priority = p
	}

	enabled := true
	if e, ok := config["enabled"].(bool); ok {
		enabled = e
	}

	// 动态导入 gRPC 流恢复中间件
	grpcStreamRecoveryMiddleware := func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// 这里调用具体的 gRPC 流恢复中间件实现
		// return StreamRecoveryInterceptor()(srv, ss, info, handler)
		return handler(srv, ss)
	}

	return NewGRPCStreamMiddleware("recovery", priority, enabled, grpcStreamRecoveryMiddleware), nil
}

func (f *RecoveryMiddlewareFactory) SupportedTypes() []MiddlewareType {
	return []MiddlewareType{HTTPMiddleware, GRPCUnaryMiddleware, GRPCStreamMiddleware}
}

// RegisterBuiltinFactories 注册所有内置的中间件工厂
func RegisterBuiltinFactories(manager *Manager) {
	manager.RegisterFactory(&LoggingMiddlewareFactory{})
	manager.RegisterFactory(&CORSMiddlewareFactory{})
	manager.RegisterFactory(&RecoveryMiddlewareFactory{})
}

// ======================== 中间件配置和初始化 ========================

// DefaultMiddlewareConfig 默认中间件配置
var DefaultMiddlewareConfig = map[string]map[string]interface{}{
	"recovery": {
		"enabled":  true,
		"priority": 10,
	},
	"logging": {
		"enabled":    true,
		"priority":   100,
		"skip_paths": []interface{}{"/health", "/metrics"},
	},
	"cors": {
		"enabled":       false,
		"priority":      50,
		"allow_origins": []interface{}{"*"},
		"allow_methods": []interface{}{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		"allow_headers": []interface{}{"Content-Type", "Authorization"},
	},
}

// InitializeMiddleware 初始化中间件系统
func InitializeMiddleware(config map[string]map[string]interface{}) (*Manager, error) {
	manager := NewManager()

	// 注册内置工厂
	RegisterBuiltinFactories(manager)

	// 如果没有提供配置，使用默认配置
	if config == nil {
		config = DefaultMiddlewareConfig
	}

	// 根据配置构建中间件链
	err := manager.BuildFromConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to build middleware chain: %w", err)
	}

	return manager, nil
}
