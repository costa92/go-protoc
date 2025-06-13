# Telemetry 包

这个包提供了在 HTTP 和 gRPC 服务中集成 OpenTelemetry 跟踪的工具。

## 功能

- OpenTelemetry tracing 初始化
- HTTP 服务的 tracing 中间件
- gRPC 服务的 tracing 拦截器（服务端和客户端）

## 使用说明

### 初始化 Tracer

在应用程序启动时，需要初始化 tracer：

```go
import "github.com/costa92/go-protoc/pkg/telemetry"

func main() {
    // 服务名称和 OTLP 端点
    serviceName := "my-service"
    endpoint := "localhost:4317"

    // 初始化 tracer
    shutdown, err := telemetry.InitTracer(serviceName, endpoint)
    if err != nil {
        log.Fatalf("初始化 tracer 失败: %v", err)
    }

    // 在程序结束时关闭 tracer
    defer func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        if err := shutdown(ctx); err != nil {
            log.Fatalf("关闭 tracer 失败: %v", err)
        }
    }()

    // 应用程序的其余部分...
}
```

### 在 HTTP 服务中使用

在 HTTP 服务中集成 tracing 中间件：

```go
import (
    "github.com/costa92/go-protoc/pkg/telemetry"
    "github.com/gorilla/mux"
)

func setupRouter() http.Handler {
    r := mux.NewRouter()

    // 注册路由
    r.HandleFunc("/api/resource", handleResource).Methods("GET")

    // 应用 tracing 中间件
    return telemetry.TracingMiddleware(r)
}
```

### 在 gRPC 服务中使用

#### 服务端

在 gRPC 服务器中集成 tracing 拦截器：

```go
import (
    "github.com/costa92/go-protoc/pkg/telemetry"
    "google.golang.org/grpc"
)

func newGRPCServer() *grpc.Server {
    // 创建 gRPC 服务器，应用 tracing 拦截器
    server := grpc.NewServer(
        grpc.UnaryInterceptor(telemetry.UnaryServerInterceptor()),
    )

    // 注册 gRPC 服务
    // pb.RegisterYourServiceServer(server, &yourServiceImpl{})

    return server
}
```

#### 客户端

在 gRPC 客户端中集成 tracing 拦截器：

```go
import (
    "github.com/costa92/go-protoc/pkg/telemetry"
    "google.golang.org/grpc"
)

func newGRPCClient(target string) (*grpc.ClientConn, error) {
    // 创建 gRPC 客户端连接，应用 tracing 拦截器
    conn, err := grpc.Dial(
        target,
        grpc.WithInsecure(),
        grpc.WithUnaryInterceptor(telemetry.UnaryClientInterceptor()),
    )
    if err != nil {
        return nil, err
    }

    return conn, nil
}
```

## 配置 OTLP 导出器

默认情况下，跟踪数据将发送到环境变量 `OTLP_ENDPOINT` 指定的 OTLP 接收器。如果未设置此环境变量，则会使用默认值 `localhost:4317`。

您可以将 OTLP 接收器配置为将数据发送到各种后端，如 Jaeger、Zipkin 或 Prometheus。

## 示例

完整的示例代码位于 `example` 目录下：

- `example/http` - HTTP 服务器示例
- `example/grpc` - gRPC 服务器示例