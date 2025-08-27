# Database Examples

展示项目数据库集成和链路追踪的示例代码，包括 MySQL、Redis 等数据库操作的 OpenTelemetry 集成。

## 📁 文件结构

```
examples/database/
├── README.md          # 本文档  
└── tracing.go         # 数据库链路追踪示例
```

## 🏗️ 数据库集成特性

- **多数据库支持**: MySQL, PostgreSQL, Redis
- **GORM 集成**: 完整的 ORM 链路追踪
- **Redis 链路追踪**: go-redis 库集成
- **自动化追踪**: 数据库操作自动生成 spans
- **性能监控**: 查询性能和连接池监控

## 🚀 快速开始

### 1. 启动数据库服务

```bash
# 启动 MySQL
./scripts/installation/service.sh start mariadb

# 启动 Redis  
./scripts/installation/service.sh start redis

# 启动链路追踪
./scripts/installation/service.sh start jaeger
```

### 2. 运行示例

```bash
cd examples/database
go run tracing.go
```

## 📊 示例内容

### 1. MySQL 链路追踪

```go
// 创建带追踪的 MySQL 连接
mysqlOpts := &db.MySQLOptions{
    Addr:                  "127.0.0.1:3306",
    Username:              "root", 
    Password:              "proj(#)666",
    Database:              "onex",
    MaxIdleConnections:    10,
    MaxOpenConnections:    100,
    MaxConnectionLifeTime: time.Hour,
}

mysqlDB, err := db.NewMySQL(mysqlOpts)
if err != nil {
    log.Fatal(err)
}

// 使用上下文执行查询（自动生成 spans）
ctx, span := otel.Tracer("mysql-example").Start(ctx, "user-query")
defer span.End()

var users []User
result := mysqlDB.WithContext(ctx).Find(&users)
```

### 2. Redis 链路追踪

```go
// 创建带追踪的 Redis 连接
redisOpts := &db.RedisOptions{
    Addr:     "127.0.0.1:6379",
    Password: "",
    Database: 0,
}

redisClient, err := db.NewRedis(redisOpts)
if err != nil {
    log.Fatal(err)
}

// Redis 操作自动生成追踪
ctx, span := otel.Tracer("redis-example").Start(ctx, "cache-operations")
defer span.End()

// SET 操作
err = redisClient.Set(ctx, "user:123", userJSON, time.Hour).Err()

// GET 操作
val, err := redisClient.Get(ctx, "user:123").Result()
```

### 3. 事务链路追踪

```go
// 数据库事务自动追踪
err = mysqlDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    // 事务内的所有操作都会被追踪
    if err := tx.Create(&user).Error; err != nil {
        return err
    }
    
    if err := tx.Create(&userProfile).Error; err != nil {
        return err  
    }
    
    return nil
})
```

## 🔧 配置说明

### OTEL 追踪初始化

```go
func initTracing() error {
    ctx := context.Background()
    
    // 创建 OTLP gRPC 导出器
    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint("http://localhost:4317"),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return err
    }
    
    // 创建资源
    res, err := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceNameKey.String("database-example"),
            semconv.ServiceVersionKey.String("1.0.0"),
        ),
    )
    
    // 创建 TracerProvider
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(res),
        trace.WithSampler(trace.AlwaysSample()),
    )
    
    otel.SetTracerProvider(tp)
    return nil
}
```

### 数据库连接配置

```yaml
# configs/apiserver.yaml
database:
  mysql:
    addr: "127.0.0.1:3306"
    username: "root"
    password: "proj(#)666" 
    database: "onex"
    max_idle_connections: 10
    max_open_connections: 100
    max_connection_life_time: "1h"
    
  redis:
    addr: "127.0.0.1:6379"
    password: ""
    database: 0
    max_idle_connections: 10
    max_active_connections: 100
```

## 📈 链路追踪特性

### 自动追踪信息

| 操作类型 | Span 名称 | 属性 |
|----------|-----------|------|
| MySQL 查询 | `gorm:query` | `db.system`, `db.operation`, `db.table` |
| MySQL 事务 | `gorm:transaction` | `db.system`, `db.transaction` |
| Redis 命令 | `redis:command` | `db.system`, `db.operation`, `db.key` |
| 连接池 | `db:connection` | `db.pool.state`, `db.pool.connections` |

### Span 属性详情

```go
// MySQL Span 属性
span.SetAttributes(
    attribute.String("db.system", "mysql"),
    attribute.String("db.operation", "SELECT"),
    attribute.String("db.table", "users"), 
    attribute.String("db.statement", "SELECT * FROM users WHERE id = ?"),
    attribute.Int("db.rows_affected", 1),
)

// Redis Span 属性  
span.SetAttributes(
    attribute.String("db.system", "redis"),
    attribute.String("db.operation", "SET"),
    attribute.String("db.redis.key", "user:123"),
    attribute.String("db.redis.command", "SET user:123 {...}"),
)
```

## 📊 性能监控

### 连接池监控

```go
// 连接池状态监控 (通过 metrics)
mysqlStats := mysqlDB.DB().Stats()
fmt.Printf("Open connections: %d\n", mysqlStats.OpenConnections)
fmt.Printf("Idle connections: %d\n", mysqlStats.Idle) 
fmt.Printf("Max open connections: %d\n", mysqlStats.MaxOpenConnections)

// Redis 连接池统计
redisStats := redisClient.PoolStats()
fmt.Printf("Hits: %d\n", redisStats.Hits)
fmt.Printf("Misses: %d\n", redisStats.Misses)
fmt.Printf("Timeouts: %d\n", redisStats.Timeouts)
```

### 查询性能分析

```go
// 慢查询追踪
ctx, span := tracer.Start(ctx, "slow-query-analysis")
defer span.End()

start := time.Now()
result := db.WithContext(ctx).Raw("SELECT * FROM large_table").Scan(&results)
duration := time.Since(start)

// 添加性能属性
span.SetAttributes(
    attribute.Int64("db.query.duration_ms", duration.Milliseconds()),
    attribute.Int("db.query.rows", len(results)),
    attribute.Bool("db.query.slow", duration > 100*time.Millisecond),
)
```

## 🔍 链路追踪查看

### Jaeger 查看

```bash
# 打开 Jaeger UI
http://127.0.0.1:16686

# 搜索服务: database-example  
# 查看操作: user-query, cache-operations
# 分析性能: 查询延迟, 数据库连接时间
```

### 追踪信息示例

```
Trace: database-example
├── user-operations (200ms)
│   ├── gorm:query:users (50ms)
│   │   └── mysql:connection (5ms)
│   ├── redis:get:user:123 (2ms) 
│   └── redis:set:user:123 (1ms)
└── cleanup (10ms)
```

## 🛠️ 自定义追踪

### 添加自定义数据库操作追踪

```go
func customDatabaseOperation(ctx context.Context, db *gorm.DB) error {
    tracer := otel.Tracer("custom-db")
    ctx, span := tracer.Start(ctx, "custom-operation")
    defer span.End()
    
    // 添加自定义属性
    span.SetAttributes(
        attribute.String("operation.type", "batch_update"),
        attribute.String("table", "users"),
        attribute.Int("batch.size", 1000),
    )
    
    // 执行数据库操作
    result := db.WithContext(ctx).Model(&User{}).
        Where("status = ?", "inactive").
        Update("status", "active")
        
    if result.Error != nil {
        span.RecordError(result.Error)
        return result.Error
    }
    
    // 记录结果
    span.SetAttributes(
        attribute.Int64("rows.affected", result.RowsAffected),
    )
    
    return nil
}
```

### Redis 管道操作追踪

```go
func redisPipelineOperation(ctx context.Context, client *redis.Client) error {
    tracer := otel.Tracer("redis-pipeline")
    ctx, span := tracer.Start(ctx, "pipeline-operation")
    defer span.End()
    
    pipe := client.Pipeline()
    
    // 批量操作
    for i := 0; i < 100; i++ {
        key := fmt.Sprintf("batch:%d", i)
        pipe.Set(ctx, key, fmt.Sprintf("value-%d", i), time.Hour)
    }
    
    // 执行管道
    results, err := pipe.Exec(ctx)
    
    // 记录结果
    span.SetAttributes(
        attribute.Int("pipeline.commands", len(results)),
        attribute.Bool("pipeline.success", err == nil),
    )
    
    return err
}
```

## 🔧 故障排除

### 常见问题

1. **数据库连接失败**
   ```bash
   # 检查数据库服务状态
   ./scripts/installation/service.sh status mariadb
   ./scripts/installation/service.sh status redis
   ```

2. **追踪数据未显示**
   ```bash  
   # 检查 Jaeger 服务
   ./scripts/installation/service.sh status jaeger
   curl http://127.0.0.1:14269/health
   ```

3. **性能问题**
   ```bash
   # 检查连接池配置
   # 调整 max_open_connections 和 max_idle_connections
   ```

## 📚 相关文档

- [数据库包文档](../../pkg/db/README.md)
- [OTEL 追踪文档](../otel/README.md)  
- [数据库服务管理](../../scripts/installation/README.md)