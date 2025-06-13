# 可插拔中间件链快速入门指南

本指南将帮助您快速上手可插拔中间件链系统，从基本概念到实际应用。

## 🚀 5分钟快速开始

### 1. 理解核心概念

```
中间件工厂 → 创建中间件 → 添加到链中 → 应用到服务器
    ↓             ↓           ↓           ↓
Factory    →  Middleware  →  Chain   →  Server
```

### 2. 基本使用流程

```go
// 1. 创建管理器
manager := middleware.NewManager()

// 2. 注册内置工厂
manager.GetBuilder().RegisterFactory(&middleware.LoggingMiddlewareFactory{})

// 3. 加载配置
config := map[string]interface{}{
    "logging": map[string]interface{}{
        "enabled": true,
        "priority": 100,
    },
}
manager.LoadFromConfig(config)

// 4. 应用到服务器
manager.ApplyToHTTPServer(httpServer)
manager.ApplyToGRPCServer(grpcServer)
```

## 📋 支持的内置中间件

| 中间件名称 | 功能描述 | 默认优先级 | 协议支持 |
|-----------|----------|------------|----------|
| `recovery` | 错误恢复和 panic 处理 | 10 | HTTP + gRPC |
| `logging` | 请求日志记录 | 100 | HTTP + gRPC |
| `cors` | 跨域资源共享 | 50 | HTTP |
| `rate_limit` | 请求限流 | 30 | HTTP + gRPC |

## 🛠️ 快速配置示例

### 基础配置

```yaml
# config.yaml
middleware:
  # 错误恢复（必须）
  recovery:
    enabled: true
    priority: 10

  # 请求日志
  logging:
    enabled: true
    priority: 100
    skip_paths: ["/health", "/metrics"]
```

### 完整配置

```yaml
middleware:
  # 错误恢复
  recovery:
    enabled: true
    priority: 10
    enable_stack_trace: true
    log_level: "error"

  # 日志记录
  logging:
    enabled: true
    priority: 100
    skip_paths: ["/health", "/metrics", "/debug"]
    log_level: "info"

  # 跨域处理
  cors:
    enabled: true
    priority: 50
    allowed_origins: ["http://localhost:3000"]
    allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allowed_headers: ["Content-Type", "Authorization"]

  # 请求限流
  rate_limit:
    enabled: false  # 生产环境可启用
    priority: 30
    limit: 1000
    window: "1m"
```

## 🎯 实战示例

### 示例1: 最简单的 HTTP 服务器

```go
package main

import (
    "net/http"
    "github.com/gorilla/mux"
    "github.com/costa92/go-protoc/pkg/middleware"
)

func main() {
    // 1. 创建中间件管理器
    manager := middleware.NewManager()

    // 2. 注册必要的工厂
    manager.GetBuilder().RegisterFactory(&middleware.LoggingMiddlewareFactory{})
    manager.GetBuilder().RegisterFactory(&middleware.RecoveryMiddlewareFactory{})

    // 3. 配置中间件
    config := map[string]interface{}{
        "recovery": map[string]interface{}{
            "enabled": true,
            "priority": 10,
        },
        "logging": map[string]interface{}{
            "enabled": true,
            "priority": 100,
            "skip_paths": []string{"/health"},
        },
    }

    manager.LoadFromConfig(config)

    // 4. 创建 HTTP 服务器
    router := mux.NewRouter()
    router.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    // 5. 应用中间件（假设你有一个支持 AddMiddleware 的服务器包装器）
    // manager.ApplyToHTTPServer(httpServer)

    // 6. 启动服务器
    http.ListenAndServe(":8080", router)
}
```

### 示例2: 完整的 Web API 服务器

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "time"

    "github.com/gorilla/mux"
    "github.com/costa92/go-protoc/pkg/middleware"
)

func main() {
    // 初始化中间件系统
    manager := setupMiddleware()

    // 创建路由器
    router := mux.NewRouter()

    // 注册路由
    router.HandleFunc("/api/users", getUsersHandler).Methods("GET")
    router.HandleFunc("/api/users", createUserHandler).Methods("POST")
    router.HandleFunc("/health", healthHandler).Methods("GET")

    // 创建服务器
    server := &http.Server{
        Addr:    ":8080",
        Handler: router,
    }

    // 应用中间件（这里演示手动应用）
    applyMiddlewareManually(router, manager)

    // 优雅启动和关闭
    gracefulStartAndStop(server)
}

func setupMiddleware() *middleware.Manager {
    manager := middleware.NewManager()

    // 注册所有内置工厂
    factories := []middleware.MiddlewareFactory{
        &middleware.RecoveryMiddlewareFactory{},
        &middleware.LoggingMiddlewareFactory{},
        &middleware.CORSMiddlewareFactory{},
        &middleware.RateLimitMiddlewareFactory{},
    }

    for _, factory := range factories {
        if err := manager.GetBuilder().RegisterFactory(factory); err != nil {
            log.Fatalf("注册中间件工厂失败: %v", err)
        }
    }

    // 从环境变量或配置文件加载配置
    config := getMiddlewareConfig()
    if err := manager.LoadFromConfig(config); err != nil {
        log.Fatalf("加载中间件配置失败: %v", err)
    }

    return manager
}

func getMiddlewareConfig() map[string]interface{} {
    return map[string]interface{}{
        "recovery": map[string]interface{}{
            "enabled": true,
            "priority": 10,
            "enable_stack_trace": true,
        },
        "logging": map[string]interface{}{
            "enabled": true,
            "priority": 100,
            "skip_paths": []string{"/health", "/metrics"},
        },
        "cors": map[string]interface{}{
            "enabled": true,
            "priority": 50,
            "allowed_origins": []string{"http://localhost:3000"},
            "allowed_methods": []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        },
        "rate_limit": map[string]interface{}{
            "enabled": false, // 开发环境禁用
            "priority": 30,
        },
    }
}

func applyMiddlewareManually(router *mux.Router, manager *middleware.Manager) {
    // 获取 HTTP 中间件并手动应用
    middlewares := manager.GetHTTPMiddlewares()
    for _, mw := range middlewares {
        router.Use(mw)
    }
}

func getUsersHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`[{"id": 1, "name": "John"}]`))
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    w.Write([]byte(`{"id": 2, "name": "Jane"}`))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("OK"))
}

func gracefulStartAndStop(server *http.Server) {
    // 启动服务器
    go func() {
        log.Printf("服务器启动在 %s", server.Addr)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("服务器启动失败: %v", err)
        }
    }()

    // 等待中断信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt)
    <-quit

    log.Println("正在关闭服务器...")

    // 优雅关闭
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("服务器关闭失败: %v", err)
    }

    log.Println("服务器已关闭")
}
```

### 示例3: 添加自定义中间件

```go
package main

import (
    "fmt"
    "net/http"
    "strings"
    "time"

    "github.com/costa92/go-protoc/pkg/middleware"
)

// 自定义认证中间件工厂
type AuthMiddlewareFactory struct{}

func (f *AuthMiddlewareFactory) Name() string {
    return "authentication"
}

func (f *AuthMiddlewareFactory) CreateHTTP(config map[string]interface{}) (*middleware.HTTPMiddlewareFunc, error) {
    // 从配置中获取参数
    secretKey := "default-secret"
    if key, ok := config["secret_key"].(string); ok {
        secretKey = key
    }

    skipPaths := []string{"/login", "/register"}
    if paths, ok := config["skip_paths"].([]interface{}); ok {
        skipPaths = make([]string, len(paths))
        for i, path := range paths {
            if pathStr, ok := path.(string); ok {
                skipPaths[i] = pathStr
            }
        }
    }

    priority := 20
    if p, ok := config["priority"].(int); ok {
        priority = p
    }

    enabled := true
    if e, ok := config["enabled"].(bool); ok {
        enabled = e
    }

    // 创建认证中间件
    authMiddleware := createAuthMiddleware(secretKey, skipPaths)

    return middleware.NewHTTPMiddleware("authentication", priority, enabled, authMiddleware), nil
}

func (f *AuthMiddlewareFactory) CreateGRPCUnary(config map[string]interface{}) (*middleware.GRPCUnaryMiddlewareFunc, error) {
    // gRPC 认证拦截器实现
    return nil, fmt.Errorf("gRPC 认证中间件暂未实现")
}

func (f *AuthMiddlewareFactory) CreateGRPCStream(config map[string]interface{}) (*middleware.GRPCStreamMiddlewareFunc, error) {
    // gRPC 流式认证拦截器实现
    return nil, fmt.Errorf("gRPC 流式认证中间件暂未实现")
}

func createAuthMiddleware(secretKey string, skipPaths []string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 检查是否需要跳过认证
            for _, path := range skipPaths {
                if strings.HasPrefix(r.URL.Path, path) {
                    next.ServeHTTP(w, r)
                    return
                }
            }

            // 检查 Authorization 头
            auth := r.Header.Get("Authorization")
            if auth == "" {
                http.Error(w, "缺少认证信息", http.StatusUnauthorized)
                return
            }

            // 简单的 Bearer Token 验证（实际项目中应该验证 JWT）
            if !strings.HasPrefix(auth, "Bearer ") {
                http.Error(w, "无效的认证格式", http.StatusUnauthorized)
                return
            }

            token := strings.TrimPrefix(auth, "Bearer ")
            if !validateToken(token, secretKey) {
                http.Error(w, "无效的令牌", http.StatusUnauthorized)
                return
            }

            // 认证通过，继续处理
            next.ServeHTTP(w, r)
        })
    }
}

func validateToken(token, secretKey string) bool {
    // 这里应该实现真正的 JWT 验证
    // 为了演示，我们只是简单检查
    return token == "valid-token-"+secretKey
}

func main() {
    // 创建管理器
    manager := middleware.NewManager()

    // 注册内置工厂
    manager.GetBuilder().RegisterFactory(&middleware.RecoveryMiddlewareFactory{})
    manager.GetBuilder().RegisterFactory(&middleware.LoggingMiddlewareFactory{})

    // 注册自定义工厂
    manager.GetBuilder().RegisterFactory(&AuthMiddlewareFactory{})

    // 配置中间件
    config := map[string]interface{}{
        "recovery": map[string]interface{}{
            "enabled": true,
            "priority": 10,
        },
        "authentication": map[string]interface{}{
            "enabled": true,
            "priority": 20,
            "secret_key": "my-secret-key",
            "skip_paths": []string{"/login", "/register", "/health"},
        },
        "logging": map[string]interface{}{
            "enabled": true,
            "priority": 100,
        },
    }

    if err := manager.LoadFromConfig(config); err != nil {
        panic(err)
    }

    // 使用中间件...
    fmt.Println("自定义认证中间件已注册并配置完成！")
}
```

## 🔧 常见配置模式

### 开发环境配置

```yaml
# dev-config.yaml
middleware:
  recovery:
    enabled: true
    priority: 10
    enable_stack_trace: true  # 开发环境显示详细错误

  logging:
    enabled: true
    priority: 100
    log_level: "debug"        # 开发环境详细日志
    skip_paths: ["/health"]

  cors:
    enabled: true
    priority: 50
    allowed_origins: ["*"]    # 开发环境允许所有源

  rate_limit:
    enabled: false            # 开发环境禁用限流
```

### 生产环境配置

```yaml
# prod-config.yaml
middleware:
  recovery:
    enabled: true
    priority: 10
    enable_stack_trace: false # 生产环境隐藏详细错误

  logging:
    enabled: true
    priority: 100
    log_level: "info"         # 生产环境适中日志
    skip_paths: ["/health", "/metrics"]

  cors:
    enabled: true
    priority: 50
    allowed_origins: ["https://myapp.com"]  # 生产环境限制源

  rate_limit:
    enabled: true             # 生产环境启用限流
    priority: 30
    limit: 1000
    window: "1m"

  authentication:             # 生产环境添加认证
    enabled: true
    priority: 20
    jwt_secret: "${JWT_SECRET}"
```

## 🚨 故障排除

### 常见问题

#### 1. gRPC 拦截器设置冲突

**错误信息**: `panic: The unary server interceptor was already set and may not be reset`

**解决方案**:

- 确保使用我们的 gRPC 服务器包装器
- 在服务器构建前添加所有拦截器
- 使用延迟构建机制

#### 2. 中间件不生效

**检查清单**:

- [ ] 中间件是否已启用 (`enabled: true`)
- [ ] 工厂是否已注册
- [ ] 配置是否正确加载
- [ ] 是否应用到了服务器

**调试代码**:

```go
// 检查已注册的工厂
factories := manager.GetBuilder().ListFactories()
fmt.Printf("已注册工厂: %v\n", factories)

// 检查中间件状态
middlewares := manager.ListMiddlewares("http")
for _, mw := range middlewares {
    fmt.Printf("中间件: %s, 启用: %v, 优先级: %d\n",
        mw.Name(), mw.Enabled(), mw.Priority())
}
```

#### 3. 配置参数无效

**检查要点**:

- 参数类型是否正确（string vs int vs bool）
- 数组格式是否正确
- 必需参数是否提供

### 性能优化

#### 1. 减少中间件数量

```go
// 仅启用必要的中间件
config := map[string]interface{}{
    "recovery": map[string]interface{}{"enabled": true},
    // "logging": map[string]interface{}{"enabled": false}, // 生产环境可选择性禁用
}
```

#### 2. 优化优先级设置

```go
// 将最常用的中间件放在前面
priorities := map[string]int{
    "recovery":      10,  // 错误处理优先
    "authentication": 20,  // 认证次之
    "rate_limit":    30,  // 限流
    "cors":          50,  // CORS
    "logging":       100, // 日志最后
}
```

## 📚 下一步

1. **阅读完整文档**: [架构设计文档](./pluggable-middleware-chain.md)
2. **查看 API 参考**: [API 参考文档](./api-reference.md)
3. **学习高级特性**: 自定义中间件开发、插件化扩展
4. **性能调优**: 中间件缓存、条件执行等优化技巧

## 💡 最佳实践总结

1. **优先级设计**: 10-19 基础设施，20-29 安全，30-39 限流，50+ 业务
2. **错误处理**: 始终启用 recovery 中间件
3. **环境区分**: 开发和生产环境使用不同配置
4. **监控观测**: 启用 logging 和 metrics 中间件
5. **安全优先**: 生产环境必须启用认证和 HTTPS
