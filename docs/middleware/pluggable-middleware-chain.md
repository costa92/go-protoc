# 可插拔中间件链（Pluggable Middleware Chaining）系统

## 概述

可插拔中间件链系统是一个灵活、可扩展的中间件管理架构，支持在运行时动态配置和组合中间件。该系统采用工厂模式、链式组合和配置驱动的设计理念，为 HTTP 和 gRPC 服务器提供统一的中间件管理能力。

## 系统架构

### 核心组件

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Configuration │    │  ChainBuilder   │    │ MiddlewareFactory│
│     Config      │───▶│   构建器        │───▶│    工厂模式     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Server   │◀───│     Manager     │───▶│ MiddlewareChain │
│     服务器      │    │   统一管理器    │    │   中间件链      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                        │
┌─────────────────┐            │                        ▼
│   gRPC Server   │◀───────────┘              ┌─────────────────┐
│     服务器      │                           │   Middleware    │
└─────────────────┘                           │   中间件实例    │
                                              └─────────────────┘
```

### 设计模式

1. **工厂模式 (Factory Pattern)**
   - `MiddlewareFactory` 接口定义中间件创建标准
   - 支持 HTTP、gRPC Unary、gRPC Stream 三种类型的中间件创建
   - 配置驱动的中间件实例化

2. **链式模式 (Chain Pattern)**
   - `MiddlewareChain` 管理中间件的执行顺序
   - 支持动态添加/移除中间件
   - 按优先级自动排序执行

3. **构建器模式 (Builder Pattern)**
   - `ChainBuilder` 根据配置构建完整的中间件链
   - 支持工厂注册和自动发现
   - 灵活的配置映射机制

## 核心接口定义

### Middleware 接口

```go
type Middleware interface {
    Type() MiddlewareType     // 返回中间件类型
    Name() string             // 返回中间件名称
    Priority() int            // 返回优先级（数字越小优先级越高）
    Enabled() bool            // 返回是否启用
}
```

### MiddlewareFactory 接口

```go
type MiddlewareFactory interface {
    Name() string
    CreateHTTP(config map[string]interface{}) (*HTTPMiddlewareFunc, error)
    CreateGRPCUnary(config map[string]interface{}) (*GRPCUnaryMiddlewareFunc, error)
    CreateGRPCStream(config map[string]interface{}) (*GRPCStreamMiddlewareFunc, error)
}
```

### ConfigurableMiddleware 接口

```go
type ConfigurableMiddleware interface {
    Middleware
    Configure(config map[string]interface{}) error
    GetConfig() map[string]interface{}
    Reconfigure(config map[string]interface{}) error
}
```

## 中间件链管理

### MiddlewareChain 结构

```go
type MiddlewareChain struct {
    middlewares []Middleware
    chainType   MiddlewareType
    mu          sync.RWMutex
}
```

**核心方法：**

- `Add(middleware Middleware)` - 添加中间件
- `Remove(name string) bool` - 移除指定中间件
- `List() []Middleware` - 获取所有中间件
- `Clear()` - 清空中间件链
- `Sort()` - 按优先级排序

### ChainBuilder 构建器

```go
type ChainBuilder struct {
    factories map[string]MiddlewareFactory
    mu        sync.RWMutex
}
```

**核心功能：**

- **工厂注册**: 注册各种中间件工厂
- **配置解析**: 解析配置文件中的中间件配置
- **链构建**: 根据配置自动构建中间件链
- **依赖管理**: 处理中间件之间的依赖关系

## 管理器 (Manager)

### Manager 结构

```go
type Manager struct {
    httpChain       *MiddlewareChain
    grpcUnaryChain  *MiddlewareChain
    grpcStreamChain *MiddlewareChain
    builder         *ChainBuilder
    config          map[string]interface{}
    mu              sync.RWMutex
}
```

### 核心功能

1. **统一管理**
   - 管理 HTTP 和 gRPC 中间件链
   - 提供统一的中间件操作接口
   - 支持中间件的生命周期管理

2. **自动应用**
   - `ApplyToHTTPServer(server interface{})` - 自动应用到 HTTP 服务器
   - `ApplyToGRPCServer(server interface{})` - 自动应用到 gRPC 服务器
   - 支持服务器接口的动态检测

3. **配置驱动**
   - `LoadFromConfig(config map[string]interface{})` - 从配置加载中间件
   - `EnableMiddleware(name string)` - 动态启用中间件
   - `DisableMiddleware(name string)` - 动态禁用中间件

## gRPC 拦截器链处理

### 问题解决

**原始问题**: gRPC 服务器不允许重复设置拦截器，导致 panic：

```
panic: The unary server interceptor was already set and may not be reset
```

**解决方案**: 实现拦截器链组合机制

### 拦截器链组合

```go
// 一元拦截器链组合
func chainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        return buildChain(interceptors, handler)(ctx, req, info)
    }
}

// 流式拦截器链组合
func chainStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
    return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
        return buildStreamChain(interceptors, handler)(srv, ss, info)
    }
}
```

### 延迟构建机制

```go
type GRPCServer struct {
    server            *grpc.Server
    listener          net.Listener
    name              string
    baseOpts          []grpc.ServerOption
    unaryInterceptors []grpc.UnaryServerInterceptor
    streamInterceptors []grpc.StreamServerInterceptor
    isBuilt           bool  // 延迟构建标记
}

func (s *GRPCServer) buildServer() {
    if s.isBuilt {
        return
    }

    // 组合所有拦截器为单一链
    if len(s.unaryInterceptors) > 0 {
        chainedUnaryInterceptor := chainUnaryInterceptors(s.unaryInterceptors...)
        opts = append(opts, grpc.UnaryInterceptor(chainedUnaryInterceptor))
    }

    s.server = grpc.NewServer(opts...)
    s.isBuilt = true
}
```

## 内置中间件工厂

### 1. LoggingMiddlewareFactory

**功能**: 请求日志记录
**配置参数**:

```yaml
logging:
  enabled: true
  priority: 100
  skip_paths: ["/health", "/metrics"]
  log_level: "info"
```

### 2. CORSMiddlewareFactory

**功能**: 跨域资源共享处理
**配置参数**:

```yaml
cors:
  enabled: true
  priority: 50
  allowed_origins: ["*"]
  allowed_methods: ["GET", "POST", "PUT", "DELETE"]
  allowed_headers: ["Content-Type", "Authorization"]
```

### 3. RecoveryMiddlewareFactory

**功能**: 错误恢复和处理
**配置参数**:

```yaml
recovery:
  enabled: true
  priority: 10
  enable_stack_trace: true
  log_level: "error"
```

### 4. RateLimitMiddlewareFactory

**功能**: 请求限流处理
**配置参数**:

```yaml
rate_limit:
  enabled: true
  priority: 30
  limit: 100
  window: "1m"
  key_func: "ip"
```

## 运行时集成

### APIServer 集成

```go
// internal/apiserver/runtime.go
func (s *APIServer) Run(opts *apiserver_options.Options) error {
    // 1. 初始化中间件管理器
    s.middlewareManager = middleware.NewManager()

    // 2. 注册内置中间件工厂
    s.registerBuiltinMiddlewares()

    // 3. 从配置加载中间件
    if err := s.middlewareManager.LoadFromConfig(middlewareConfig); err != nil {
        return err
    }

    // 4. 应用中间件链到服务器
    if err := s.middlewareManager.ApplyToHTTPServer(httpServer); err != nil {
        return err
    }

    if err := s.middlewareManager.ApplyToGRPCServer(grpcServer); err != nil {
        return err
    }

    // 5. 启动服务器
    return s.startServers(ctx, httpServer, grpcServer)
}
```

### 中间件注册

```go
func (s *APIServer) registerBuiltinMiddlewares() {
    builder := s.middlewareManager.GetBuilder()

    // 注册内置中间件工厂
    builder.RegisterFactory(&middleware.LoggingMiddlewareFactory{})
    builder.RegisterFactory(&middleware.CORSMiddlewareFactory{})
    builder.RegisterFactory(&middleware.RecoveryMiddlewareFactory{})
    builder.RegisterFactory(&middleware.RateLimitMiddlewareFactory{})

    logger.Info("已注册内置中间件工厂")
}
```

## 配置示例

### 完整配置文件

```yaml
# config.yaml
middleware:
  # 日志中间件
  logging:
    enabled: true
    priority: 100
    skip_paths: ["/health", "/metrics", "/debug"]
    log_level: "info"

  # CORS 中间件
  cors:
    enabled: true
    priority: 50
    allowed_origins: ["http://localhost:3000", "https://example.com"]
    allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allowed_headers: ["Content-Type", "Authorization", "X-Request-ID"]
    max_age: 86400

  # 恢复中间件
  recovery:
    enabled: true
    priority: 10
    enable_stack_trace: true
    log_level: "error"

  # 限流中间件
  rate_limit:
    enabled: false  # 可以选择性禁用
    priority: 30
    limit: 1000
    window: "1m"
    key_func: "ip"

  # 自定义中间件
  authentication:
    enabled: true
    priority: 20
    jwt_secret: "${JWT_SECRET}"
    skip_paths: ["/login", "/register"]
```

## 使用示例

### 1. 添加自定义中间件

```go
// 1. 实现中间件工厂
type AuthMiddlewareFactory struct{}

func (f *AuthMiddlewareFactory) Name() string {
    return "authentication"
}

func (f *AuthMiddlewareFactory) CreateHTTP(config map[string]interface{}) (*HTTPMiddlewareFunc, error) {
    secret := config["jwt_secret"].(string)
    skipPaths := config["skip_paths"].([]string)

    return NewHTTPMiddleware("authentication", 20, true,
        authMiddleware(secret, skipPaths)), nil
}

// 2. 注册工厂
manager.GetBuilder().RegisterFactory(&AuthMiddlewareFactory{})

// 3. 配置中启用
middlewareConfig := map[string]interface{}{
    "authentication": map[string]interface{}{
        "enabled": true,
        "priority": 20,
        "jwt_secret": "your-secret-key",
        "skip_paths": []string{"/login", "/register"},
    },
}

manager.LoadFromConfig(middlewareConfig)
```

### 2. 动态管理中间件

```go
// 动态启用中间件
manager.EnableMiddleware("rate_limit")

// 动态禁用中间件
manager.DisableMiddleware("cors")

// 重新配置中间件
newConfig := map[string]interface{}{
    "limit": 500,
    "window": "30s",
}
manager.ReconfigureMiddleware("rate_limit", newConfig)

// 获取中间件状态
middlewares := manager.ListMiddlewares("http")
for _, mw := range middlewares {
    fmt.Printf("中间件: %s, 启用: %v, 优先级: %d\n",
        mw.Name(), mw.Enabled(), mw.Priority())
}
```

### 3. 中间件链执行流程

```
请求 → Recovery → Auth → RateLimit → CORS → Logging → 业务逻辑
      ↓         ↓      ↓          ↓      ↓         ↓
    优先级:    10     20      30        50     100      ∞
```

## 性能优化

### 1. 中间件缓存

```go
type Manager struct {
    // ... 其他字段
    cachedHTTPMiddlewares []mux.MiddlewareFunc
    cacheValid          bool
    mu                  sync.RWMutex
}

func (m *Manager) GetHTTPMiddlewares() []mux.MiddlewareFunc {
    m.mu.RLock()
    if m.cacheValid {
        result := make([]mux.MiddlewareFunc, len(m.cachedHTTPMiddlewares))
        copy(result, m.cachedHTTPMiddlewares)
        m.mu.RUnlock()
        return result
    }
    m.mu.RUnlock()

    // 重新构建缓存
    m.rebuildCache()
    return m.cachedHTTPMiddlewares
}
```

### 2. 条件性执行

```go
// 支持条件性中间件执行
type ConditionalMiddleware struct {
    Middleware
    condition func(*http.Request) bool
}

func (m *ConditionalMiddleware) shouldExecute(req *http.Request) bool {
    return m.condition == nil || m.condition(req)
}
```

## 监控和观测

### 1. 中间件指标

```go
// 中间件执行指标
type MiddlewareMetrics struct {
    ExecutionCount   prometheus.Counter
    ExecutionTime    prometheus.Histogram
    ErrorCount       prometheus.Counter
    ActiveRequests   prometheus.Gauge
}

func (m *Middleware) recordMetrics(duration time.Duration, err error) {
    m.metrics.ExecutionCount.Inc()
    m.metrics.ExecutionTime.Observe(duration.Seconds())
    if err != nil {
        m.metrics.ErrorCount.Inc()
    }
}
```

### 2. 链路追踪

```go
// 支持分布式追踪
func (m *HTTPMiddlewareFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 开始 span
    span, ctx := opentracing.StartSpanFromContext(ctx,
        fmt.Sprintf("middleware.%s", m.name))
    defer span.Finish()

    // 执行中间件
    m.handler.ServeHTTP(w, r.WithContext(ctx))
}
```

## 最佳实践

### 1. 中间件优先级设计

```
优先级范围建议:
- 10-19: 基础设施中间件 (Recovery, Panic处理)
- 20-29: 安全中间件 (Authentication, Authorization)
- 30-39: 限流中间件 (RateLimit, Circuit Breaker)
- 40-49: 协议处理 (CORS, Headers)
- 50-99: 业务中间件 (Validation, Transform)
- 100+:  观测中间件 (Logging, Metrics, Tracing)
```

### 2. 配置管理

```go
// 支持环境变量替换
func expandEnvVars(config map[string]interface{}) {
    for key, value := range config {
        if str, ok := value.(string); ok {
            if strings.HasPrefix(str, "${") && strings.HasSuffix(str, "}") {
                envVar := str[2 : len(str)-1]
                config[key] = os.Getenv(envVar)
            }
        }
    }
}
```

### 3. 错误处理

```go
// 统一错误处理
type MiddlewareError struct {
    Name    string
    Type    string
    Message string
    Cause   error
}

func (e *MiddlewareError) Error() string {
    return fmt.Sprintf("middleware %s (%s): %s", e.Name, e.Type, e.Message)
}
```

## 扩展指南

### 1. 实现自定义中间件

```go
// 1. 定义中间件逻辑
func customMiddleware(config map[string]interface{}) mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 前置处理
            // ...

            next.ServeHTTP(w, r)

            // 后置处理
            // ...
        })
    }
}

// 2. 实现工厂
type CustomMiddlewareFactory struct{}

func (f *CustomMiddlewareFactory) Name() string {
    return "custom"
}

func (f *CustomMiddlewareFactory) CreateHTTP(config map[string]interface{}) (*HTTPMiddlewareFunc, error) {
    middleware := customMiddleware(config)
    return NewHTTPMiddleware("custom", 60, true, middleware), nil
}

// 3. 注册和使用
manager.GetBuilder().RegisterFactory(&CustomMiddlewareFactory{})
```

### 2. 中间件插件化

```go
// 支持插件化加载
type MiddlewarePlugin interface {
    Name() string
    Version() string
    Factories() []MiddlewareFactory
}

func (m *Manager) LoadPlugin(plugin MiddlewarePlugin) error {
    for _, factory := range plugin.Factories() {
        m.builder.RegisterFactory(factory)
    }
    return nil
}
```

## 总结

可插拔中间件链系统提供了：

1. **灵活性**: 支持动态配置和组合中间件
2. **可扩展性**: 工厂模式支持轻松添加新中间件
3. **统一性**: HTTP 和 gRPC 统一管理
4. **安全性**: 类型安全的设计和错误处理
5. **性能**: 优化的执行链和缓存机制
6. **可观测性**: 完整的监控和追踪支持

该系统完全解决了 gRPC 拦截器设置冲突的问题，并为项目提供了强大的中间件管理能力。
