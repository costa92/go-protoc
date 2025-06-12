# 可插拔中间件链 API 参考

本文档详细描述了可插拔中间件链系统的所有公共 API 接口。

## 核心接口

### Middleware 接口

```go
type Middleware interface {
    Type() MiddlewareType
    Name() string
    Priority() int
    Enabled() bool
}
```

**描述**: 定义了所有中间件必须实现的基础接口。

**方法**:

- `Type() MiddlewareType`: 返回中间件类型（HTTP、gRPC Unary、gRPC Stream）
- `Name() string`: 返回中间件的唯一名称
- `Priority() int`: 返回执行优先级，数字越小优先级越高
- `Enabled() bool`: 返回中间件是否启用

### MiddlewareFactory 接口

```go
type MiddlewareFactory interface {
    Name() string
    CreateHTTP(config map[string]interface{}) (*HTTPMiddlewareFunc, error)
    CreateGRPCUnary(config map[string]interface{}) (*GRPCUnaryMiddlewareFunc, error)
    CreateGRPCStream(config map[string]interface{}) (*GRPCStreamMiddlewareFunc, error)
}
```

**描述**: 中间件工厂接口，用于创建不同类型的中间件实例。

**方法**:

- `Name() string`: 返回工厂名称，必须与配置中的中间件名称匹配
- `CreateHTTP(config) (*HTTPMiddlewareFunc, error)`: 创建 HTTP 中间件
- `CreateGRPCUnary(config) (*GRPCUnaryMiddlewareFunc, error)`: 创建 gRPC 一元中间件
- `CreateGRPCStream(config) (*GRPCStreamMiddlewareFunc, error)`: 创建 gRPC 流式中间件

### ConfigurableMiddleware 接口

```go
type ConfigurableMiddleware interface {
    Middleware
    Configure(config map[string]interface{}) error
    GetConfig() map[string]interface{}
    Reconfigure(config map[string]interface{}) error
}
```

**描述**: 扩展了 Middleware 接口，支持运行时配置。

**方法**:

- `Configure(config) error`: 初始配置中间件
- `GetConfig() map[string]interface{}`: 获取当前配置
- `Reconfigure(config) error`: 重新配置中间件

## 核心结构体

### Manager

```go
type Manager struct {
    // 私有字段
}
```

**描述**: 中间件管理器，提供统一的中间件管理功能。

#### 构造函数

```go
func NewManager() *Manager
```

**返回**: 新的管理器实例

#### 方法

##### LoadFromConfig

```go
func (m *Manager) LoadFromConfig(config map[string]interface{}) error
```

**描述**: 从配置加载中间件
**参数**:

- `config`: 中间件配置映射
**返回**: 错误信息（如果有）

**示例**:

```go
config := map[string]interface{}{
    "logging": map[string]interface{}{
        "enabled": true,
        "priority": 100,
        "skip_paths": []string{"/health"},
    },
}
err := manager.LoadFromConfig(config)
```

##### ApplyToHTTPServer

```go
func (m *Manager) ApplyToHTTPServer(server interface{}) error
```

**描述**: 将中间件链应用到 HTTP 服务器
**参数**:

- `server`: 实现了 `AddMiddleware(mux.MiddlewareFunc)` 方法的服务器
**返回**: 错误信息（如果有）

##### ApplyToGRPCServer

```go
func (m *Manager) ApplyToGRPCServer(server interface{}) error
```

**描述**: 将中间件链应用到 gRPC 服务器
**参数**:

- `server`: 实现了添加拦截器方法的 gRPC 服务器
**返回**: 错误信息（如果有）

##### EnableMiddleware

```go
func (m *Manager) EnableMiddleware(name string) error
```

**描述**: 动态启用指定中间件
**参数**:

- `name`: 中间件名称
**返回**: 错误信息（如果有）

##### DisableMiddleware

```go
func (m *Manager) DisableMiddleware(name string) error
```

**描述**: 动态禁用指定中间件
**参数**:

- `name`: 中间件名称
**返回**: 错误信息（如果有）

##### ListMiddlewares

```go
func (m *Manager) ListMiddlewares(middlewareType string) []Middleware
```

**描述**: 列出指定类型的所有中间件
**参数**:

- `middlewareType`: 中间件类型（"http"、"grpc_unary"、"grpc_stream"）
**返回**: 中间件列表

##### GetBuilder

```go
func (m *Manager) GetBuilder() *ChainBuilder
```

**描述**: 获取链构建器
**返回**: ChainBuilder 实例

### MiddlewareChain

```go
type MiddlewareChain struct {
    // 私有字段
}
```

**描述**: 中间件链，管理一组中间件的执行顺序。

#### 构造函数

```go
func NewMiddlewareChain(chainType MiddlewareType) *MiddlewareChain
```

**参数**:

- `chainType`: 链类型
**返回**: 新的中间件链实例

#### 方法

##### Add

```go
func (c *MiddlewareChain) Add(middleware Middleware)
```

**描述**: 添加中间件到链中
**参数**:

- `middleware`: 要添加的中间件

##### Remove

```go
func (c *MiddlewareChain) Remove(name string) bool
```

**描述**: 从链中移除指定中间件
**参数**:

- `name`: 中间件名称
**返回**: 是否成功移除

##### List

```go
func (c *MiddlewareChain) List() []Middleware
```

**描述**: 获取链中的所有中间件
**返回**: 中间件列表（按优先级排序）

##### Clear

```go
func (c *MiddlewareChain) Clear()
```

**描述**: 清空中间件链

### ChainBuilder

```go
type ChainBuilder struct {
    // 私有字段
}
```

**描述**: 中间件链构建器，负责根据配置创建中间件链。

#### 构造函数

```go
func NewChainBuilder() *ChainBuilder
```

**返回**: 新的链构建器实例

#### 方法

##### RegisterFactory

```go
func (b *ChainBuilder) RegisterFactory(factory MiddlewareFactory) error
```

**描述**: 注册中间件工厂
**参数**:

- `factory`: 要注册的工厂
**返回**: 错误信息（如果有）

##### GetFactory

```go
func (b *ChainBuilder) GetFactory(name string) (MiddlewareFactory, bool)
```

**描述**: 获取指定名称的工厂
**参数**:

- `name`: 工厂名称
**返回**: 工厂实例和是否存在的标志

##### ListFactories

```go
func (b *ChainBuilder) ListFactories() []string
```

**描述**: 列出所有已注册的工厂名称
**返回**: 工厂名称列表

##### BuildChain

```go
func (b *ChainBuilder) BuildChain(
    chainType MiddlewareType,
    config map[string]interface{}
) (*MiddlewareChain, error)
```

**描述**: 根据配置构建中间件链
**参数**:

- `chainType`: 链类型
- `config`: 配置映射
**返回**: 构建的中间件链和错误信息

## 中间件类型

### HTTPMiddlewareFunc

```go
type HTTPMiddlewareFunc struct {
    name     string
    priority int
    enabled  bool
    handler  mux.MiddlewareFunc
}
```

**描述**: HTTP 中间件包装器

#### 构造函数

```go
func NewHTTPMiddleware(
    name string,
    priority int,
    enabled bool,
    handler mux.MiddlewareFunc
) *HTTPMiddlewareFunc
```

**参数**:

- `name`: 中间件名称
- `priority`: 优先级
- `enabled`: 是否启用
- `handler`: 实际的中间件处理函数

### GRPCUnaryMiddlewareFunc

```go
type GRPCUnaryMiddlewareFunc struct {
    name        string
    priority    int
    enabled     bool
    interceptor grpc.UnaryServerInterceptor
}
```

**描述**: gRPC 一元中间件包装器

#### 构造函数

```go
func NewGRPCUnaryMiddleware(
    name string,
    priority int,
    enabled bool,
    interceptor grpc.UnaryServerInterceptor
) *GRPCUnaryMiddlewareFunc
```

### GRPCStreamMiddlewareFunc

```go
type GRPCStreamMiddlewareFunc struct {
    name        string
    priority    int
    enabled     bool
    interceptor grpc.StreamServerInterceptor
}
```

**描述**: gRPC 流式中间件包装器

#### 构造函数

```go
func NewGRPCStreamMiddleware(
    name string,
    priority int,
    enabled bool,
    interceptor grpc.StreamServerInterceptor
) *GRPCStreamMiddlewareFunc
```

## 内置中间件工厂

### LoggingMiddlewareFactory

```go
type LoggingMiddlewareFactory struct{}
```

**配置参数**:

- `enabled` (bool): 是否启用，默认 true
- `priority` (int): 优先级，默认 100
- `skip_paths` ([]string): 跳过日志记录的路径
- `log_level` (string): 日志级别

**示例配置**:

```yaml
logging:
  enabled: true
  priority: 100
  skip_paths: ["/health", "/metrics"]
  log_level: "info"
```

### CORSMiddlewareFactory

```go
type CORSMiddlewareFactory struct{}
```

**配置参数**:

- `enabled` (bool): 是否启用，默认 true
- `priority` (int): 优先级，默认 50
- `allowed_origins` ([]string): 允许的源
- `allowed_methods` ([]string): 允许的方法
- `allowed_headers` ([]string): 允许的头部
- `max_age` (int): 预检请求缓存时间

**示例配置**:

```yaml
cors:
  enabled: true
  priority: 50
  allowed_origins: ["*"]
  allowed_methods: ["GET", "POST", "PUT", "DELETE"]
  allowed_headers: ["Content-Type", "Authorization"]
  max_age: 86400
```

### RecoveryMiddlewareFactory

```go
type RecoveryMiddlewareFactory struct{}
```

**配置参数**:

- `enabled` (bool): 是否启用，默认 true
- `priority` (int): 优先级，默认 10
- `enable_stack_trace` (bool): 是否启用堆栈跟踪
- `log_level` (string): 日志级别

**示例配置**:

```yaml
recovery:
  enabled: true
  priority: 10
  enable_stack_trace: true
  log_level: "error"
```

### RateLimitMiddlewareFactory

```go
type RateLimitMiddlewareFactory struct{}
```

**配置参数**:

- `enabled` (bool): 是否启用，默认 false
- `priority` (int): 优先级，默认 30
- `limit` (int): 限制数量
- `window` (string): 时间窗口
- `key_func` (string): 键函数类型

**示例配置**:

```yaml
rate_limit:
  enabled: true
  priority: 30
  limit: 1000
  window: "1m"
  key_func: "ip"
```

## 全局函数

### SetGlobalManager

```go
func SetGlobalManager(manager *Manager)
```

**描述**: 设置全局中间件管理器
**参数**:

- `manager`: 管理器实例

### GetGlobalManager

```go
func GetGlobalManager() *Manager
```

**描述**: 获取全局中间件管理器
**返回**: 管理器实例

### RegisterGlobalFactory

```go
func RegisterGlobalFactory(factory MiddlewareFactory) error
```

**描述**: 向全局管理器注册工厂
**参数**:

- `factory`: 要注册的工厂
**返回**: 错误信息（如果有）

## 常量和类型

### MiddlewareType

```go
type MiddlewareType string

const (
    HTTPMiddleware       MiddlewareType = "http"
    GRPCUnaryMiddleware  MiddlewareType = "grpc_unary"
    GRPCStreamMiddleware MiddlewareType = "grpc_stream"
)
```

**描述**: 中间件类型枚举

## 错误类型

### MiddlewareError

```go
type MiddlewareError struct {
    Name    string
    Type    string
    Message string
    Cause   error
}

func (e *MiddlewareError) Error() string
```

**描述**: 中间件相关错误

### 预定义错误

```go
var (
    ErrMiddlewareNotFound    = errors.New("middleware not found")
    ErrFactoryNotRegistered  = errors.New("factory not registered")
    ErrInvalidConfiguration  = errors.New("invalid configuration")
    ErrMiddlewareDisabled    = errors.New("middleware is disabled")
    ErrChainNotFound        = errors.New("middleware chain not found")
)
```

## 示例用法

### 基本使用

```go
package main

import (
    "github.com/costa92/go-protoc/pkg/middleware"
)

func main() {
    // 创建管理器
    manager := middleware.NewManager()

    // 注册工厂
    err := manager.GetBuilder().RegisterFactory(&middleware.LoggingMiddlewareFactory{})
    if err != nil {
        panic(err)
    }

    // 加载配置
    config := map[string]interface{}{
        "logging": map[string]interface{}{
            "enabled": true,
            "priority": 100,
        },
    }

    err = manager.LoadFromConfig(config)
    if err != nil {
        panic(err)
    }

    // 应用到服务器
    err = manager.ApplyToHTTPServer(httpServer)
    if err != nil {
        panic(err)
    }
}
```

### 自定义中间件

```go
// 实现自定义工厂
type CustomFactory struct{}

func (f *CustomFactory) Name() string {
    return "custom"
}

func (f *CustomFactory) CreateHTTP(config map[string]interface{}) (*middleware.HTTPMiddlewareFunc, error) {
    return middleware.NewHTTPMiddleware("custom", 50, true, func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 自定义逻辑
            next.ServeHTTP(w, r)
        })
    }), nil
}

func (f *CustomFactory) CreateGRPCUnary(config map[string]interface{}) (*middleware.GRPCUnaryMiddlewareFunc, error) {
    // 实现 gRPC 一元拦截器
    return nil, nil
}

func (f *CustomFactory) CreateGRPCStream(config map[string]interface{}) (*middleware.GRPCStreamMiddlewareFunc, error) {
    // 实现 gRPC 流式拦截器
    return nil, nil
}

// 注册和使用
manager.GetBuilder().RegisterFactory(&CustomFactory{})
```

## 注意事项

1. **线程安全**: 所有公共 API 都是线程安全的
2. **配置变更**: 中间件配置变更需要重新应用到服务器
3. **优先级**: 数字越小优先级越高，建议使用 10-999 范围
4. **gRPC 限制**: gRPC 服务器必须在构建前添加拦截器
5. **内存管理**: 管理器会持有中间件实例的引用，注意内存泄漏
