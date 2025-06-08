# 链路跟踪 (Distributed Tracing)

本文档描述了如何在项目中实现分布式链路跟踪，以及解决常见问题的方法。

## 概述

分布式链路跟踪是微服务架构中不可或缺的组件，它可以帮助我们了解请求如何在不同服务之间传播，识别性能瓶颈，并排查故障。我们使用 [OpenTelemetry](https://opentelemetry.io/) 作为链路跟踪的解决方案。

## 实现方式

在我们的项目中，我们使用以下方式实现链路跟踪：

1. **HTTP 请求**：使用 `otelhttp.NewHandler` 中间件包装 HTTP 处理器
2. **gRPC 服务**：使用 `otelgrpc.NewServerHandler` 和 `grpc.StatsHandler` 来追踪 gRPC 调用

### 具体实现

```go
// 初始化 OpenTelemetry Tracer
tp, err := tracing.InitTracer("go-protoc-service")
if err != nil {
    logger.Fatal("failed to initialize OpenTelemetry Tracer", zap.Error(err))
}

// 为 HTTP 请求创建链路跟踪中间件
otelHTTPMiddleware := func(next http.Handler) http.Handler {
    return otelhttp.NewHandler(next, "http-server")
}

// 为 gRPC 创建统计处理器
otelGrpcHandler := otelgrpc.NewServerHandler(otelgrpc.WithTracerProvider(tp))

// 在创建应用实例时应用这些中间件和处理器
apiServer := app.NewApp(":8090", ":8091", logger,
    // 添加 HTTP 中间件
    app.WithHTTPMiddlewares(
        otelHTTPMiddleware,
        // ... 其他中间件
    ),
    // 添加 gRPC 服务器选项
    app.WithGRPCOptions(
        grpc.StatsHandler(otelGrpcHandler),
    ),
)
```

## 常见问题与解决方案

### 问题：`otelgrpc.UnaryServerInterceptor` 未定义

**症状**：运行时出现错误 `undefined: otelgrpc.UnaryServerInterceptor`

**原因**：OpenTelemetry gRPC 集成的 API 在不同版本之间有变化。在较新版本中（如 v0.61.0），推荐使用 `NewServerHandler` 和 `StatsHandler` 而不是拦截器。

**解决方案**：

1. 移除对 `otelgrpc.UnaryServerInterceptor` 和 `otelgrpc.StreamServerInterceptor` 的调用
2. 使用新的 API 方式实现链路跟踪：

```go
// 旧方式（不再推荐）
app.WithGRPCUnaryInterceptors(
    otelgrpc.UnaryServerInterceptor(otelgrpc.WithTracerProvider(tp)),
    // ... 其他拦截器
)

// 新方式（推荐）
otelGrpcHandler := otelgrpc.NewServerHandler(otelgrpc.WithTracerProvider(tp))
app.WithGRPCOptions(
    grpc.StatsHandler(otelGrpcHandler),
)
```

### 与链路跟踪相关的包版本

确保在 `go.mod` 中使用兼容的版本：

```
go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.61.0
go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.49.0
go.opentelemetry.io/otel v1.24.0
go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.24.0
go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.24.0
go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.24.0
go.opentelemetry.io/otel/sdk v1.24.0
go.opentelemetry.io/otel/trace v1.24.0
```

## 输出与查看链路跟踪数据

默认情况下，我们的实现将跟踪数据输出到标准输出。在生产环境中，您可能需要将数据发送到链路跟踪后端系统，如 Jaeger、Zipkin 或 OpenTelemetry Collector。

要更改导出器，请修改 `pkg/tracing/tracing.go` 文件中的 `InitTracer` 函数。

## 参考链接

- [OpenTelemetry 官方文档](https://opentelemetry.io/docs/)
- [OpenTelemetry Go SDK](https://github.com/open-telemetry/opentelemetry-go)
- [OpenTelemetry gRPC 集成](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/google.golang.org/grpc/otelgrpc)
- [OpenTelemetry HTTP 集成](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/net/http/otelhttp)
- [gRPC StatsHandler 文档](https://pkg.go.dev/google.golang.org/grpc#StatsHandler)
- [OpenTelemetry Go Contrib 文档](https://pkg.go.dev/go.opentelemetry.io/contrib)
