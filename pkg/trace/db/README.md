# 数据库追踪包

这个包提供了统一的数据库追踪实现，包括 MySQL、Redis 和 MongoDB 的追踪支持。

## 🎯 **特性**

- ✅ **统一接口** - 所有数据库使用相同的追踪管理器
- ✅ **自动埋点** - 自动记录数据库操作的关键信息
- ✅ **事务支持** - 支持数据库事务的追踪
- ✅ **性能监控** - 记录执行时间、影响行数等指标
- ✅ **错误追踪** - 自动记录和报告错误

## 📚 **使用方法**

### 1. MySQL 追踪 (使用 GORM 官方插件)

```go
import (
    "github.com/costa92/go-protoc/v2/pkg/trace/db"
    "gorm.io/gorm"
)

// 为 GORM 数据库添加官方 OpenTelemetry 插件
func setupMySQLTracing(gormDB *gorm.DB) error {
    // 使用 GORM 官方推荐的方式
    plugin := db.NewMySQLPluginWithOptions()
    return gormDB.Use(plugin)
}

// 或者使用自定义配置
func setupMySQLTracingCustom(gormDB *gorm.DB) error {
    plugin := db.NewMySQLPlugin(
        // 可以传入自定义选项
    )
    return gormDB.Use(plugin)
}

// 使用示例 - 查询操作自动追踪
func userQuery(ctx context.Context, gormDB *gorm.DB) {
    var user User
    
    // GORM 官方插件会自动追踪所有 SQL 操作
    result := db.WithContext(ctx, gormDB).First(&user, 1)
    if result.Error != nil {
        log.Printf("查询失败: %v", result.Error)
    }
}

// 事务追踪 - 推荐使用 GORM 的 Transaction 方法
func userTransaction(ctx context.Context, gormDB *gorm.DB) error {
    return db.ExecuteInTransaction(ctx, gormDB, func(tx *gorm.DB) error {
        // 在事务中执行操作，自动追踪
        if err := tx.Create(&user).Error; err != nil {
            return err // 自动回滚
        }
        
        if err := tx.Create(&order).Error; err != nil {
            return err // 自动回滚
        }
        
        return nil // 自动提交
    })
}

// 手动事务控制（如果需要更细粒度控制）
func userTransactionManual(ctx context.Context, gormDB *gorm.DB) error {
    tx := db.StartTransaction(ctx, gormDB)
    
    defer func() {
        if r := recover(); r != nil {
            db.RollbackTransaction(tx)
            panic(r)
        }
    }()
    
    // 执行事务操作
    if err := tx.Create(&user).Error; err != nil {
        db.RollbackTransaction(tx)
        return err
    }
    
    return db.CommitTransaction(tx)
}
```

### 2. Redis 追踪

```go
import (
    "github.com/costa92/go-protoc/v2/pkg/trace/db"
    "github.com/redis/go-redis/v9"
)

// 为 Redis 客户端添加追踪 Hook
func setupRedisTracing(rdb *redis.Client, addr string) {
    hook := db.NewRedisHook(addr)
    rdb.AddHook(hook)
}

// 使用示例
func redisOperations(ctx context.Context, rdb *redis.Client) {
    // 自动追踪 Redis 操作
    result := rdb.Get(ctx, "user:123")
    if result.Err() != nil {
        log.Printf("Redis GET 失败: %v", result.Err())
    }
    
    // 使用追踪包装器
    tracedClient := db.NewTracedRedisClient(rdb)
    tracedCtx := tracedClient.WithTracing(ctx, "custom-operation")
    
    // Pipeline 操作也会自动追踪
    pipe := rdb.Pipeline()
    pipe.Set(ctx, "key1", "value1", 0)
    pipe.Get(ctx, "key2")
    _, err := pipe.Exec(ctx)
}
```

### 3. MongoDB 追踪

```go
import (
    "github.com/costa92/go-protoc/v2/pkg/trace/db"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

// 为 MongoDB 客户端添加追踪监控
func setupMongoDBTracing() *mongo.Client {
    monitor := db.NewMongoTraceMonitor()
    
    clientOpts := options.Client().ApplyURI("mongodb://localhost:27017")
    clientOpts.SetMonitor(&event.CommandMonitor{
        Started:   monitor.Started,
        Succeeded: monitor.Succeeded,
        Failed:    monitor.Failed,
    })
    
    client, err := mongo.Connect(context.Background(), clientOpts)
    return client
}

// 使用示例
func mongoOperations(ctx context.Context, client *mongo.Client) {
    collection := db.NewTracedMongoCollection(client, "mydb", "users")
    
    // 查询操作会自动追踪
    ctx, span := collection.WithTracing(ctx, "find-user")
    defer span.End()
    
    var result User
    err := collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&result)
}

// 事务追踪
func mongoTransaction(ctx context.Context, client *mongo.Client) error {
    transaction, err := db.StartTracedTransaction(ctx, client)
    if err != nil {
        return err
    }
    
    defer func() {
        if r := recover(); r != nil {
            transaction.Abort()
            panic(r)
        }
    }()
    
    // 使用事务执行操作
    session := transaction.Session()
    
    // ... 执行数据库操作
    
    return transaction.Commit()
}
```

## 🔧 **配置说明**

### 追踪器名称

每个数据库类型使用不同的追踪器名称：

```go
const (
    MySQLTracerName   = "gorm.mysql"    // MySQL/GORM 追踪器
    RedisTracerName   = "redis"         // Redis 追踪器
    MongoDBTracerName = "mongodb"       // MongoDB 追踪器
)
```

### 记录的属性

#### MySQL 属性
- `db.system`: "mysql"
- `db.operation`: 操作类型 (select, insert, update, delete)
- `db.sql.table`: 表名
- `db.statement`: SQL 语句
- `db.rows_affected`: 影响的行数
- `db.duration`: 执行时间（微秒）

#### Redis 属性
- `db.system`: "redis"
- `db.operation`: 命令类型 (get, set, hget, etc.)
- `net.peer.name`: Redis 服务器地址
- `net.peer.port`: Redis 服务器端口
- `db.redis.key`: 操作的键名
- `db.redis.args`: 命令参数
- `redis.duration`: 执行时间（微秒）

#### MongoDB 属性
- `db.system`: "mongodb"
- `db.operation`: 操作类型 (find, insert, update, delete)
- `db.name`: 数据库名
- `db.mongodb.collection`: 集合名
- `db.mongodb.filter`: 查询过滤器
- `db.statement`: 完整的命令
- `mongodb.duration`: 执行时间（微秒）

## 💡 **最佳实践**

### 1. 初始化顺序

```go
// 1. 先初始化全局追踪器
trace.InitializeGlobalTracer(tracerConfig)

// 2. 再设置数据库追踪
setupMySQLTracing(gormDB)
setupRedisTracing(rdb, "127.0.0.1:6379")
setupMongoDBTracing()
```

### 2. 上下文传递

```go
// 确保在所有数据库操作中传递上下文
func processUser(ctx context.Context, userID int) error {
    // 查询用户 - 自动追踪
    user, err := getUserFromDB(ctx, userID)
    if err != nil {
        return err
    }
    
    // 缓存用户信息 - 自动追踪
    return cacheUserToRedis(ctx, user)
}
```

### 3. 错误处理

```go
// 事务中的错误处理
func safeTransaction(ctx context.Context, db *gorm.DB) (err error) {
    tx, span := db.StartTransaction(ctx, db)
    
    defer func() {
        if err != nil {
            db.RollbackTransaction(tx, span)
        } else {
            err = db.CommitTransaction(tx, span)
        }
    }()
    
    // 执行业务逻辑
    err = doBusinessLogic(tx)
    return
}
```

### 4. 敏感数据保护

```go
// 避免在追踪中记录敏感信息
// MySQL: 使用参数化查询
db.Where("email = ?", email).First(&user)

// Redis: 避免在 key 中包含敏感信息
rdb.Set(ctx, "user:"+hashUserID(userID), userData, 0)
```

## 📊 **性能影响**

- **开销**: < 1% 额外 CPU 开销
- **内存**: 每个 span 约 1-2KB 内存
- **网络**: 批量发送到 Jaeger，影响最小
- **延迟**: 追踪记录是异步的，不影响响应时间

## 🚀 **从旧版本迁移**

### 旧代码
```go
// 旧的分散式追踪
db.Use(&trace.OpenTelemetryPlugin{}) // ❌ 旧方式
```

### 新代码  
```go
// 新的统一追踪
plugin := db.NewMySQLPlugin()        // ✅ 新方式
db.Use(plugin)
```

所有接口保持兼容，只需要更改导入路径：
```go
// 旧导入
import "github.com/costa92/go-protoc/v2/pkg/db"

// 新导入
import "github.com/costa92/go-protoc/v2/pkg/trace/db"
```
