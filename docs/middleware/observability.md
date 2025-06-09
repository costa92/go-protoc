# 可观察性中间件配置指南

## 目录

- [概述](#概述)
- [白名单配置](#白名单配置)
- [中间件说明](#中间件说明)
- [使用示例](#使用示例)
- [自定义配置](#自定义配置)

## 概述

可观察性中间件用于处理 HTTP 请求的日志记录、监控指标收集和错误追踪。为了避免某些系统级路径产生不必要的日志和指标，我们实现了白名单机制。

## 白名单配置

### 默认白名单路径

以下路径默认不会被记录到可观察性系统中：

```go
SkipPaths: []string{
    "/metrics",           // Prometheus 指标
    "/debug/",           // Debug 端点
    "/swagger/",         // Swagger UI
    "/healthz",          // 健康检查
    "/favicon.ico",      // 浏览器图标请求
}
```

### 白名单效果

对于在白名单中的路径：

1. 不会生成访问日志
2. 不会记录 Prometheus 指标
3. 不会被超时中间件处理
4. 不会被恢复中间件处理

### 配置位置

白名单配置定义在 `pkg/middleware/config/config.go` 文件中：

```go
type ObservabilityConfig struct {
    // SkipPaths 定义不需要记录的路径前缀
    SkipPaths []string
}
```

## 中间件说明

### 1. 日志中间件 (LoggingMiddleware)

```go
// 使用默认配置
router.Use(middleware.LoggingMiddleware(logger))

// 使用自定义配置
router.Use(middleware.LoggingMiddlewareWithConfig(logger, customConfig))
```

功能：

- 记录请求方法、路径、状态码
- 记录请求处理时间
- 记录客户端信息
- 记录 TraceID（用于分布式追踪）

### 2. 恢复中间件 (RecoveryMiddleware)

```go
// 使用默认配置
router.Use(middleware.RecoveryMiddleware(logger))

// 使用自定义配置
router.Use(middleware.RecoveryMiddlewareWithConfig(logger, customConfig))
```

功能：

- 捕获并恢复 panic
- 记录错误信息
- 返回 500 Internal Server Error

### 3. 超时中间件 (TimeoutMiddleware)

```go
// 使用默认配置
router.Use(middleware.TimeoutMiddleware(5 * time.Second))

// 使用自定义配置
router.Use(middleware.TimeoutMiddlewareWithConfig(5 * time.Second, customConfig))
```

功能：

- 控制请求超时时间
- 自动取消超时的请求
- 返回 504 Gateway Timeout

## 使用示例

### 基本用法

```go
package main

import (
    "github.com/costa92/go-protoc/pkg/middleware/http"
    "github.com/gorilla/mux"
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewProduction()
    router := mux.NewRouter()

    // 使用默认配置的中间件
    router.Use(http.LoggingMiddleware(logger))
    router.Use(http.RecoveryMiddleware(logger))
    router.Use(http.TimeoutMiddleware(5 * time.Second))
}
```

### 自定义配置

```go
package main

import (
    "github.com/costa92/go-protoc/pkg/middleware/config"
    "github.com/costa92/go-protoc/pkg/middleware/http"
)

func main() {
    // 创建自定义配置
    customConfig := &config.ObservabilityConfig{
        SkipPaths: []string{
            "/metrics",
            "/debug/",
            "/swagger/",
            "/healthz",
            "/favicon.ico",
            "/custom/path",          // 自定义路径
            "/api/internal/",        // 内部 API
            "/monitoring/",          // 监控端点
        },
    }

    // 使用自定义配置创建中间件
    router.Use(http.LoggingMiddlewareWithConfig(logger, customConfig))
    router.Use(http.RecoveryMiddlewareWithConfig(logger, customConfig))
    router.Use(http.TimeoutMiddlewareWithConfig(5 * time.Second, customConfig))
}
```

## 最佳实践

1. 白名单路径建议：

   - 监控和指标端点
   - 健康检查端点
   - 调试和性能分析端点
   - API 文档端点
   - 内部管理端点

2. 路径匹配规则：

   - 使用前缀匹配（例如：`/api/internal/` 会匹配所有以此开头的路径）
   - 确保路径以 `/` 开头
   - 如果需要匹配目录，以 `/` 结尾

3. 性能考虑：

   - 白名单检查在中间件链的最前面执行
   - 使用前缀匹配可以快速跳过不需要处理的请求
   - 白名单路径越少，性能影响越小

4. 安全建议：
   - 将所有调试和管理端点添加到白名单
   - 确保这些端点有其他的安全保护措施
   - 定期审查白名单配置

## 配置更新

如需更新白名单配置：

1. 修改 `pkg/middleware/config/config.go` 中的 `DefaultObservabilityConfig` 函数
2. 或者在应用启动时提供自定义配置

```go
func UpdateObservabilityConfig(paths []string) *config.ObservabilityConfig {
    return &config.ObservabilityConfig{
        SkipPaths: paths,
    }
}
```
