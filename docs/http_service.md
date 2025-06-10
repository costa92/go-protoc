# HTTP 服务设计与使用指南

## 1. 服务概述

Go-Protoc 项目中的 HTTP 服务采用了先进的分层架构设计，将 gRPC 和 RESTful HTTP API 无缝整合，提供统一的 API 接口。该服务器架构支持直接的 HTTP 路由和通过 gRPC-Gateway 自动生成的 REST API，确保路由注册顺序正确，优化请求处理流程。

### 设计目标

- **统一服务入口**：同时支持 HTTP 和 gRPC 协议，减少代码重复
- **优先级路由处理**：确保显式路由优先于自动生成的 gRPC-Gateway 路由
- **安全并发**：线程安全的路由注册和请求处理
- **可扩展性**：轻松添加新的 API 组和路由
- **标准化错误处理**：一致的错误响应格式
- **内置调试工具**：集成 pprof 等性能分析工具

### 架构图

```
                   ┌─────────────────┐
                   │    客户端请求    │
                   └────────┬────────┘
                            │
                            ▼
                   ┌─────────────────┐
                   │   HTTP 中间件   │ ◄── 日志、认证、CORS 等
                   └────────┬────────┘
                            │
                            ▼
                   ┌─────────────────┐
          ┌───────►  HTTP 路由器     │
          │        └────────┬────────┘
          │                 │
┌─────────┴─────────┐      │      ┌─────────────────┐
│ 直接 HTTP 处理器  │◄─────┴─────►│ gRPC-Gateway   │
└───────────────────┘             └────────┬────────┘
                                           │
                                           ▼
                                  ┌─────────────────┐
                                  │   gRPC 服务     │
                                  └─────────────────┘
```

## 2. 核心组件

### HTTPServer 结构

`HTTPServer` 是 HTTP 服务的核心组件，它封装了标准库的 `http.Server`，并提供了额外的功能：

```go
type HTTPServer struct {
    *http.Server         // 内嵌标准库 HTTP 服务器
    router     *mux.Router        // HTTP 路由器
    gatewayMux *runtime.ServeMux  // gRPC-Gateway 多路复用器
    name       string             // 服务器名称
    mu         sync.Mutex         // 保护路由注册的并发安全
    gatewayAdded bool             // 标记 Gateway 是否已添加
}
```

### 关键方法

- **NewHTTPServer**：创建和初始化 HTTP 服务器实例
- **AddRoute**：安全地添加自定义 HTTP 路由
- **FinalizeRoutes**：完成路由注册，添加 gRPC-Gateway 默认处理器
- **Router/GatewayMux**：获取路由器和 Gateway 多路复用器
- **Start/Stop**：控制服务器生命周期

## 3. 路由注册流程

HTTP 服务的路由注册遵循特定顺序，确保所有路由正确处理：

```
┌─────────────────┐
│  创建 HTTPServer │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  应用全局中间件  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 注册健康检查路由 │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  注册调试路由    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 注册自定义 HTTP  │
│     路由        │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 注册 API 组     │ ◄── 向 gRPC-Gateway 注册路由
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  完成路由注册    │ ◄── 添加 gRPC-Gateway 作为默认处理器
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   启动服务器     │
└─────────────────┘
```

## 4. 请求处理流程

当 HTTP 服务器接收到请求时，请求的处理流程如下：

```
┌─────────────────┐
│   HTTP 请求     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  中间件链处理    │ ◄── 日志、超时、认证等
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 路由匹配决策    │
└────────┬────────┘
         │
    ┌────┴─────┐
    ▼          ▼
┌────────┐ ┌────────┐
│直接路由│ │Gateway │ ◄── 如果没有匹配的直接路由
└───┬────┘ └───┬────┘
    │          │
    ▼          ▼
┌────────┐ ┌────────┐
│HTTP处理│ │ gRPC   │
│  器    │ │ 服务   │
└───┬────┘ └───┬────┘
    │          │
    └────┬─────┘
         │
         ▼
┌─────────────────┐
│   HTTP 响应     │
└─────────────────┘
```

## 5. 使用案例

### 案例一：创建 HTTP 服务器并添加自定义路由

```go
package main

import (
    "context"
    "encoding/json"
    "net/http"
    "time"

    "github.com/costa92/go-protoc/pkg/app"
    "github.com/costa92/go-protoc/pkg/log"
)

func main() {
    // 创建 HTTP 服务器
    httpServer := app.NewHTTPServer("api-http", ":8080")

    // 添加自定义 API 路由
    httpServer.AddRoute("/api/v1/users", handleUsers, "GET", "POST")
    httpServer.AddRoute("/api/v1/users/{id}", handleUserById, "GET", "PUT", "DELETE")

    // 完成路由注册
    httpServer.FinalizeRoutes()

    // 创建上下文和取消函数
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 启动服务器
    if err := httpServer.Start(ctx); err != nil {
        log.Fatalf("服务器启动失败: %v", err)
    }
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
    // 处理用户集合
    response := map[string]interface{}{
        "users": []map[string]interface{}{
            {"id": 1, "name": "用户1"},
            {"id": 2, "name": "用户2"},
        },
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func handleUserById(w http.ResponseWriter, r *http.Request) {
    // 处理单个用户
    // ...
}
```

### 案例二：注册 API 组

```go
package myapi

import (
    "context"

    "github.com/costa92/go-protoc/internal/apiserver"
    "github.com/costa92/go-protoc/pkg/app"
    myapiv1 "github.com/costa92/go-protoc/pkg/api/myapi/v1"
)

// Installer 实现 APIGroupInstaller 接口
type Installer struct{}

// NewInstaller 创建一个新的安装器
func NewInstaller() *Installer {
    return &Installer{}
}

// Install 安装 API 组
func (i *Installer) Install(grpcServer *app.GRPCServer, httpServer *app.HTTPServer) error {
    // 创建服务实现
    service := NewMyServiceServer()

    // 注册 gRPC 服务
    myapiv1.RegisterMyServiceServer(grpcServer.Server(), service)

    // 注册 gRPC-Gateway 处理器
    if err := myapiv1.RegisterMyServiceHandlerServer(
        context.Background(),
        httpServer.GatewayMux(),
        service,
    ); err != nil {
        return err
    }

    // 添加自定义 HTTP 路由（如果需要）
    httpServer.AddRoute("/api/myservice/custom", handleCustomRequest, "GET")

    return nil
}

// 在初始化时自动注册
func init() {
    apiserver.RegisterAPIGroup(NewInstaller())
}
```

### 案例三：集成性能分析工具

HTTPServer 已内置 pprof 支持，可以轻松获取性能分析数据：

```go
// 服务器已经内置注册了 pprof 路由

// 使用 pprof 工具采集 CPU 分析数据
// go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30

// 查看内存分配情况
// go tool pprof http://localhost:8080/debug/pprof/heap

// 查看 goroutine 阻塞情况
// go tool pprof http://localhost:8080/debug/pprof/block
```

## 6. 最佳实践

### 路由注册顺序

1. **先注册中间件**：确保所有请求都经过必要的处理
2. **再注册显式路由**：清晰定义的 HTTP 路由（如 REST API）
3. **最后注册 gRPC-Gateway**：作为默认处理器

### 处理器实现建议

1. **保持简洁**：每个处理器专注于单一职责
2. **使用依赖注入**：避免全局状态和硬编码依赖
3. **统一错误处理**：使用统一的错误响应格式
4. **适当的日志记录**：记录关键操作和错误情况

### 路由命名约定

1. **REST 风格**：`/api/v1/resources/{id}`
2. **版本化**：在路径中包含 API 版本
3. **一致性**：保持路由格式的一致性

## 7. 错误处理与调试

### 标准错误响应格式

```json
{
  "status": "error",
  "code": 400,
  "message": "请求参数无效",
  "details": {
    "field": "username",
    "error": "长度必须在3-20之间"
  }
}
```

### 常见问题与解决方案

1. **路由无法访问**：确保路由注册在 `FinalizeRoutes()` 调用之前
2. **路由顺序错误**：检查路由注册顺序，确保特定路由在通用路由之前
3. **中间件不生效**：确保中间件在创建服务器时注册

## 8. 性能优化

1. **连接池**：使用适当的连接池配置
2. **超时控制**：为所有 HTTP 操作设置合理的超时
3. **压缩**：启用 HTTP 压缩以减少传输数据量
4. **缓存**：适当使用缓存减少重复计算

## 9. 监控与可观测性

### 关键指标

1. **请求计数**：每个路径的请求数
2. **响应时间**：请求处理的延迟分布
3. **错误率**：每个路径的错误百分比
4. **资源使用**：CPU、内存、网络和磁盘使用情况

### 集成 pprof

HTTP 服务器已集成 `net/http/pprof` 包，提供以下调试端点：

- `/debug/pprof/` - pprof 首页
- `/debug/pprof/heap` - 堆内存分析
- `/debug/pprof/goroutine` - goroutine 分析
- `/debug/pprof/profile` - CPU 分析
- `/debug/pprof/allocs` - 内存分配分析

## 10. 与其他组件的集成

### 与 gRPC 服务的集成

HTTP 服务器与 gRPC 服务紧密集成，通过 gRPC-Gateway 自动将 gRPC 服务暴露为 RESTful API。

### 与认证系统的集成

可以轻松集成各种认证中间件，如 JWT、OAuth 等。

### 与监控系统的集成

支持与 Prometheus、Jaeger 等监控工具集成，实现全面的可观测性。

## 11. 总结

我们的 HTTP 服务设计提供了一个强大而灵活的框架，同时支持传统 HTTP API 和 gRPC 服务。它的主要优势包括：

1. **统一接口**：同时支持 HTTP 和 gRPC 调用
2. **灵活路由**：支持动态路由注册和优先级管理
3. **并发安全**：线程安全的路由注册机制
4. **可扩展性**：易于添加新的 API 组和功能
5. **内置调试**：集成了 pprof 等调试工具

通过遵循本文档中的设计原则和最佳实践，您可以构建高性能、可维护和可扩展的 HTTP 服务。
