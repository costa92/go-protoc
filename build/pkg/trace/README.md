# Trace Package

分布式链路追踪包，基于OpenTelemetry实现，提供统一的链路追踪解决方案，支持数据库、Redis、HTTP等组件的自动追踪。

## 📋 目录

- [功能特性](#功能特性)
- [快速开始](#快速开始)
- [架构设计](#架构设计)
- [数据库追踪](#数据库追踪)
- [配置选项](#配置选项)
- [API参考](#api参考)
- [最佳实践](#最佳实践)
- [故障排除](#故障排除)

## 🎯 功能特性

### 核心特性
- 🔍 **分布式追踪**: 基于OpenTelemetry的标准化追踪
- 🗄️ **数据库追踪**: MySQL、PostgreSQL、MongoDB自动追踪
- 📡 **多样化导出**: 支持Jaeger、Zipkin等追踪系统
- 🚀 **零侵入集成**: 通过插件和中间件自动集成
- ⚡ **高性能**: 批量处理、采样控制、异步导出
- 🔧 **配置灵活**: 支持动态配置和环境变量

### 支持的组件
- **GORM**: MySQL、PostgreSQL数据库追踪
- **MongoDB**: MongoDB操作追踪
- **Redis**: Redis命令追踪
- **HTTP**: HTTP请求/响应追踪
- **gRPC**: gRPC调用追踪

## 🚀 快速开始

### 全局追踪器初始化

```go
package main

import (
    "context"
    "log"
    
    "github.com/costa92/go-protoc/v2/pkg/trace"
)

func main() {
    // 1. 配置追踪器
    config := trace.TracerConfig{
        ServiceName: "my-service",
        Endpoint:    "http://localhost:14268/api/traces",
        Environment: "development",
    }

    // 2. 初始化全局追踪器
    err := trace.InitializeGlobalTracer(config)
    if err != nil {
        log.Fatal("初始化追踪器失败:", err)
    }

    // 3. 确保在程序退出时关闭
    defer func() {
        ctx := context.Background()
        trace.Shutdown(ctx)
    }()

    // 4. 开始使用追踪
    ctx := context.Background()
    ctx, span := trace.StartSpan(ctx, "main-operation")
    defer span.End()

    // 业务逻辑...
    doSomething(ctx)
}

func doSomething(ctx context.Context) {
    // 创建子span
    ctx, span := trace.StartSpan(ctx, "do-something",
        trace.WithAttributes(
            attribute.String("operation", "business-logic"),
            attribute.Int("user_id", 123),
        ),
    )
    defer span.End()

    // 模拟业务操作
    time.Sleep(100 * time.Millisecond)
}
```

### HTTP服务追踪

```go
func setupHTTPTracing() {
    // 使用中间件
    http.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
        ctx, span := trace.StartSpan(r.Context(), "get-users",
            trace.WithSpanKind(trace.SpanKindServer),
            trace.WithAttributes(
                attribute.String("http.method", r.Method),
                attribute.String("http.url", r.URL.String()),
            ),
        )
        defer span.End()

        // 处理请求
        users := getUsersFromDB(ctx)
        
        // 设置响应属性
        span.SetAttributes(
            attribute.Int("user.count", len(users)),
            attribute.Int("http.status_code", 200),
        )

        json.NewEncoder(w).Encode(users)
    })
}
```

## 🗄️ 数据库追踪

### MySQL/PostgreSQL (GORM)

```go
package main

import (
    "github.com/costa92/go-protoc/v2/pkg/trace/db"
    "gorm.io/gorm"
)

func setupMySQLTracing() {
    // 1. 创建追踪插件
    plugin := db.NewMySQLPlugin()
    
    // 2. 配置GORM
    gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        panic(err)
    }
    
    // 3. 安装插件
    err = gormDB.Use(plugin)
    if err != nil {
        panic(err)
    }

    // 4. 使用带追踪的数据库操作
    ctx := context.Background()
    
    // 自动追踪数据库操作
    var users []User
    err = gormDB.WithContext(ctx).Find(&users).Error
    if err != nil {
        log.Error("查询失败", err)
    }
}

// 事务追踪
func tracedTransaction(ctx context.Context, db *gorm.DB) error {
    return db.ExecuteInTransaction(ctx, db, func(tx *gorm.DB) error {
        // 创建用户
        user := &User{Name: "John", Email: "john@example.com"}
        if err := tx.Create(user).Error; err != nil {
            return err
        }
        
        // 创建用户配置
        profile := &UserProfile{UserID: user.ID, Avatar: "avatar.jpg"}
        if err := tx.Create(profile).Error; err != nil {
            return err
        }
        
        return nil
    })
}
```

### MongoDB追踪

```go
func setupMongoTracing() {
    // 1. 创建MongoDB追踪器
    tracer := db.NewMongoTracer("myapp-mongo")
    
    // 2. 创建带追踪的客户端
    client, err := mongo.Connect(context.Background(), options.Client().
        ApplyURI("mongodb://localhost:27017").
        SetMonitor(tracer.CommandMonitor()))
    if err != nil {
        panic(err)
    }

    // 3. 使用带追踪的操作
    collection := client.Database("myapp").Collection("users")
    
    ctx := context.Background()
    
    // 插入文档（自动追踪）
    _, err = collection.InsertOne(ctx, bson.M{
        "name":  "John",
        "email": "john@example.com",
    })
    
    // 查询文档（自动追踪）
    var user bson.M
    err = collection.FindOne(ctx, bson.M{"name": "John"}).Decode(&user)
}
```

### Redis追踪

```go
func setupRedisTracing() {
    // 1. 创建Redis追踪器
    tracer := db.NewRedisTracer("myapp-redis")
    
    // 2. 创建带追踪的Redis客户端
    client := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    // 3. 安装追踪钩子
    client.AddHook(tracer)

    // 4. 使用带追踪的Redis操作
    ctx := context.Background()
    
    // SET操作（自动追踪）
    err := client.Set(ctx, "user:123", "john", time.Hour).Err()
    if err != nil {
        log.Error("Redis SET失败", err)
    }
    
    // GET操作（自动追踪）
    val, err := client.Get(ctx, "user:123").Result()
    if err != nil {
        log.Error("Redis GET失败", err)
    }
}
```

## ⚙️ 配置选项

### 基础配置

```go
// 追踪器配置
type TracerConfig struct {
    ServiceName string // 服务名称
    Endpoint    string // 导出端点
    Environment string // 环境标识
}

// 详细配置
type AdvancedTracerConfig struct {
    ServiceName  string
    ServiceVersion string
    Environment  string
    
    // Jaeger配置
    JaegerEndpoint string
    JaegerUser     string
    JaegerPassword string
    
    // 采样配置
    SamplingRate float64
    
    // 批处理配置
    BatchTimeout   time.Duration
    BatchSize      int
    ExportTimeout  time.Duration
    
    // 资源属性
    ResourceAttributes map[string]string
}
```

### 环境变量配置

```bash
# 基础配置
export OTEL_SERVICE_NAME="my-service"
export OTEL_SERVICE_VERSION="1.0.0"
export OTEL_ENVIRONMENT="production"

# Jaeger配置
export JAEGER_ENDPOINT="http://localhost:14268/api/traces"
export JAEGER_USER="admin"
export JAEGER_PASSWORD="password"

# 采样配置
export OTEL_TRACES_SAMPLER="traceidratio"
export OTEL_TRACES_SAMPLER_ARG="1.0"

# 导出配置
export OTEL_TRACES_EXPORTER="jaeger"
export OTEL_EXPORTER_JAEGER_TIMEOUT="10s"
```

## 📚 API参考

### 全局追踪器

```go
// 初始化和管理
func InitializeGlobalTracer(cfg TracerConfig) error
func GetTracer(name string) trace.Tracer
func IsInitialized() bool
func Shutdown(ctx context.Context) error

// Span操作
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
func WithAttributes(attrs ...attribute.KeyValue) trace.SpanStartOption
func WithSpanKind(kind trace.SpanKind) trace.SpanStartOption
```

### 数据库追踪

```go
// MySQL/PostgreSQL (GORM)
func NewMySQLPlugin(opts ...gormOtel.Option) gorm.Plugin
func NewMySQLPluginWithOptions() gorm.Plugin
func WithContext(ctx context.Context, db *gorm.DB) *gorm.DB
func StartTransaction(ctx context.Context, db *gorm.DB) *gorm.DB
func ExecuteInTransaction(ctx context.Context, db *gorm.DB, fn func(*gorm.DB) error) error
func TracedQuery(ctx context.Context, db *gorm.DB, operation string, fn func(*gorm.DB) error) error

// MongoDB
func NewMongoTracer(serviceName string) *MongoTracer
func (t *MongoTracer) CommandMonitor() *event.CommandMonitor

// Redis  
func NewRedisTracer(serviceName string) *RedisTracer
func (t *RedisTracer) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error)
func (t *RedisTracer) AfterProcess(ctx context.Context, cmd redis.Cmder) error
```

## 🏆 最佳实践

### 1. Span命名约定

```go
// 好的命名
"GET /api/users"           // HTTP请求
"mysql.users.select"       // 数据库操作
"redis.get"               // Redis操作
"auth.validate_token"     // 业务操作

// 避免的命名  
"operation"               // 太泛化
"func1"                  // 无意义
"very_long_operation_name_that_describes_everything" // 过长
```

### 2. 属性设置

```go
// 推荐的属性
span.SetAttributes(
    // 业务属性
    attribute.String("user.id", userID),
    attribute.String("operation", "create_order"),
    
    // 技术属性
    attribute.String("db.system", "mysql"),
    attribute.String("db.table", "orders"),
    
    // 性能属性
    attribute.Int("result.count", len(results)),
    attribute.Float64("processing.duration", processingTime),
)

// 避免高基数属性
span.SetAttributes(
    // Bad: 会产生太多唯一值
    attribute.String("user.email", email),
    attribute.String("request.full_url", fullURL),
)
```

### 3. 错误处理

```go
func tracedOperation(ctx context.Context) error {
    ctx, span := trace.StartSpan(ctx, "traced-operation")
    defer span.End()

    result, err := doSomething(ctx)
    if err != nil {
        // 记录错误信息
        span.SetStatus(codes.Error, err.Error())
        span.SetAttributes(
            attribute.String("error.type", fmt.Sprintf("%T", err)),
            attribute.String("error.message", err.Error()),
        )
        return err
    }

    // 记录成功信息
    span.SetStatus(codes.Ok, "operation completed successfully")
    span.SetAttributes(
        attribute.Int("result.size", len(result)),
    )
    
    return nil
}
```

## 🔧 故障排除

### 常见问题

#### 1. 追踪数据未显示

**症状**: Jaeger UI中看不到追踪数据

**排查步骤**:
```go
// 检查初始化状态
if !trace.IsInitialized() {
    log.Error("全局追踪器未初始化")
}

// 检查配置
config := trace.GetCurrentConfig()
log.Info("追踪配置", "config", config)

// 测试连接
err := trace.TestConnection()
if err != nil {
    log.Error("追踪导出器连接失败", err)
}

// 强制导出
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
err = trace.ForceFlush(ctx)
if err != nil {
    log.Error("强制导出失败", err)
}
```

#### 2. 性能影响过大

**症状**: 应用性能显著下降

**解决方案**:
```go
// 降低采样率
config.SamplingRate = 0.01 // 1%采样

// 优化批处理
config.BatchSize = 1024
config.BatchTimeout = 10 * time.Second

// 异步导出
config.AsyncExport = true

// 排除低价值路径
config.ExcludePatterns = []string{
    "/health",
    "/metrics", 
    "/debug/*",
}
```

---

## 🤝 贡献指南

1. **问题报告**: 请在GitHub Issues中详细描述追踪问题
2. **功能请求**: 提交新的追踪需求和使用场景
3. **代码贡献**: Fork项目并提交Pull Request
4. **文档改进**: 帮助完善追踪文档和示例

## 📄 许可证

本项目采用MIT许可证，详见LICENSE文件。

---

**版本**: v2.0.0  
**最后更新**: 2025-01-15  
**维护者**: Go-Protoc Team
