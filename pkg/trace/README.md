# 统一追踪管理器

这个包提供了一个统一的追踪管理器，用于管理 OpenTelemetry 追踪的初始化和使用。

## 🎯 **设计原则**

1. **统一初始化** - 在应用启动时只初始化一次全局追踪器
2. **共享配置** - 所有组件共享相同的追踪配置（采样率、导出器等）
3. **简化使用** - 提供简单的 API 获取追踪器实例
4. **类型安全** - 使用统一的错误处理和日志记录

## 📚 **使用方法**

### 1. 初始化全局追踪器

在应用启动时（通常在 main 函数或服务初始化时）调用一次：

```go
import "github.com/costa92/go-protoc/v2/pkg/trace"

// 在应用启动时初始化
cfg := trace.TracerConfig{
    ServiceName: "apiserver",
    Endpoint:    "http://127.0.0.1:14268/api/traces",
    Environment: "development",
}

err := trace.InitializeGlobalTracer(cfg)
if err != nil {
    log.Fatalf("初始化追踪器失败: %v", err)
}
```

### 2. 在组件中获取追踪器

各个组件（MySQL、Redis、HTTP 等）通过 `GetTracer` 获取追踪器：

```go
// 获取通用追踪器
tracer := trace.GetTracer("mysql")

// 或者获取全局追踪器
tracer := trace.GetTracer("")

// 创建 span
ctx, span := tracer.Start(ctx, "mysql.query", trace.WithSpanKind(ottrace.SpanKindClient))
defer span.End()
```

### 3. 使用便利函数

包提供了一些便利函数简化使用：

```go
// 直接启动 span
ctx, span := trace.StartSpan(ctx, "my-operation",
    trace.WithAttributes(
        attribute.String("db.system", "mysql"),
        attribute.String("operation", "select"),
    ),
    trace.WithSpanKind(ottrace.SpanKindClient),
)
defer span.End()
```

### 4. 检查初始化状态

```go
if !trace.IsInitialized() {
    log.Warn("追踪器未初始化，将使用空操作追踪器")
}
```

### 5. 优雅关闭

在应用关闭时调用：

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := trace.Shutdown(ctx); err != nil {
    log.Errorf("关闭追踪器失败: %v", err)
}
```

## 🔧 **配置参数**

### TracerConfig

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| ServiceName | string | ✅ | 服务名称，在 Jaeger UI 中显示 |
| Endpoint | string | ✅ | Jaeger 收集器端点 URL |
| Environment | string | ✅ | 环境标识（dev/staging/prod） |

## 💡 **最佳实践**

### 1. 组件命名规范

为不同组件使用有意义的追踪器名称：

```go
const (
    MySQLTracerName = "gorm.mysql"
    RedisTracerName = "redis"
    HTTPTracerName  = "http.client"
)
```

### 2. Span 命名规范

使用清晰的 span 名称：

```go
// ✅ 好的命名
"mysql.query"
"mysql.transaction"
"redis.get"
"redis.pipeline"
"http.request"

// ❌ 避免的命名
"db"
"operation"
"request"
```

### 3. 属性设置

设置有用的 span 属性：

```go
span.SetAttributes(
    attribute.String("db.system", "mysql"),
    attribute.String("db.operation", "SELECT"),
    attribute.String("db.sql.table", "users"),
    attribute.Int64("db.rows_affected", 1),
)
```

### 4. 错误处理

正确处理错误：

```go
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
    return err
}
span.SetStatus(codes.Ok, "")
```

## 🚀 **迁移指南**

### 从旧版本迁移

#### 旧版本 (每个组件独立初始化)
```go
// MySQL
var tracer = otel.Tracer("gorm.mysql")

// Redis  
var redisTracer = otel.Tracer("redis")
```

#### 新版本 (统一管理)
```go
// 在应用启动时初始化一次
trace.InitializeGlobalTracer(cfg)

// 在组件中获取
tracer := trace.GetTracer("gorm.mysql")
redisTracer := trace.GetTracer("redis")
```

## 🐛 **故障排查**

### 常见问题

1. **追踪器未初始化**
   ```
   [Trace] 警告: 全局追踪器未初始化，返回空追踪器
   ```
   **解决方案**: 确保在应用启动时调用了 `InitializeGlobalTracer`

2. **重复初始化**
   ```
   error: global tracer already initialized
   ```
   **解决方案**: 只在应用启动时初始化一次

3. **连接 Jaeger 失败**
   ```
   创建 Jaeger 导出器失败: connection refused
   ```
   **解决方案**: 检查 Jaeger 服务是否运行，端点 URL 是否正确

### 调试日志

初始化时会输出详细日志：
```
[Trace] 正在初始化全局追踪器...
[Trace] 服务名称: apiserver
[Trace] 端点 URL: http://127.0.0.1:14268/api/traces
[Trace] 环境: development
[Trace] 全局追踪器初始化成功! 采样率: 100%
[Trace] 请访问 Jaeger UI: http://localhost:16686
```

## 📈 **性能说明**

- **初始化开销**: 只在启动时发生一次
- **运行时开销**: 最小化，使用全局 TracerProvider
- **内存使用**: 共享追踪器实例，减少内存占用
- **采样率**: 默认 100%，生产环境建议调整
