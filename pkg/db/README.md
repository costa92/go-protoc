# Database Package

数据库连接和管理包，提供统一的数据库连接抽象层，支持MySQL、PostgreSQL、Redis等多种数据库。

## 📋 目录

- [功能特性](#功能特性)
- [快速开始](#快速开始)
- [支持的数据库](#支持的数据库)
- [配置选项](#配置选项)
- [API参考](#api参考)
- [最佳实践](#最佳实践)

## 🎯 功能特性

### 核心特性
- 🗄️ **多数据库支持**: MySQL、PostgreSQL、Redis
- 🔧 **统一配置**: 标准化的配置接口和选项
- 🏊 **连接池管理**: 智能连接池配置和优化
- 🔄 **生命周期管理**: 完整的连接生命周期控制
- ⚡ **性能优化**: 连接复用、批量操作优化
- 📊 **可选监控**: 通过依赖注入支持连接池监控

### 接口设计
- **DatabaseConnection**: 数据库连接接口
- **ConfigValidator**: 配置验证器接口
- **LifecycleManager**: 生命周期管理器接口
- **PoolMonitor**: 连接池监控接口（可选）

## 🚀 快速开始

### MySQL连接

```go
package main

import (
	"time"
	"github.com/costa92/go-protoc/v2/pkg/db"
	"gorm.io/gorm"
)

func main() {
	// 1. 创建MySQL配置
	opts := &db.MySQLOptions{
		Addr:                  "localhost:3306",
		Username:              "root",
		Password:              "password",
		Database:              "myapp",
		MaxIdleConnections:    10,
		MaxOpenConnections:    100,
		MaxConnectionLifeTime: time.Hour,
	}

	// 2. 创建数据库连接
	gormDB, err := db.NewMySQL(opts)
	if err != nil {
		panic(err)
	}

	// 3. 使用GORM进行数据库操作
	var users []User
	result := gormDB.Find(&users)
	if result.Error != nil {
		log.Error("查询失败", result.Error)
	}
}
```

### Redis连接

```go
func setupRedis() {
	opts := &db.RedisOptions{
		Addr:         "localhost:6379",
		Password:     "",
		Database:     0,
		PoolSize:     100,
		MinIdleConns: 10,
	}

	// 创建Redis客户端
	client, err := db.NewRedis(opts)
	if err != nil {
		panic(err)
	}

	// 使用Redis
	ctx := context.Background()
	err = client.Set(ctx, "key", "value", time.Hour).Err()
	if err != nil {
		log.Error("Redis操作失败", err)
	}
}
```

## 🗄️ 支持的数据库

### MySQL
- **驱动**: GORM + MySQL驱动
- **特性**: 连接池、事务、迁移、hooks

```go
type MySQLOptions struct {
	Addr                  string
	Username              string  
	Password              string
	Database              string
	MaxIdleConnections    int
	MaxOpenConnections    int
	MaxConnectionLifeTime time.Duration
	Logger                logger.Interface
	// +optional 可选的连接池监控器
	Monitor               PoolMonitor
}
```

### PostgreSQL
- **驱动**: GORM + PostgreSQL驱动
- **特性**: 完整PostgreSQL特性支持
- **配置**: 类似MySQL，支持SSL连接

```go
type PostgreSQLOptions struct {
	Addr                  string
	Username              string
	Password              string
	Database              string
	SSLMode               string
	MaxIdleConnections    int
	MaxOpenConnections    int
	MaxConnectionLifeTime time.Duration
}
```

### Redis
- **驱动**: go-redis/v9
- **特性**: 集群支持、哨兵模式、管道操作

```go
type RedisOptions struct {
	Addr         string
	Username     string
	Password     string
	Database     int
	MaxRetries   int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolTimeout  time.Duration
	PoolSize     int
	// +optional 可选的连接池监控器
	Monitor      PoolMonitor
}
```

## ⚙️ 配置选项

### 连接池配置

```go
// MySQL连接池配置
mysqlOpts := &db.MySQLOptions{
	MaxIdleConnections:    10,   // 最大空闲连接数
	MaxOpenConnections:    100,  // 最大打开连接数
	MaxConnectionLifeTime: time.Hour, // 连接最大生存时间
}

// Redis连接池配置
redisOpts := &db.RedisOptions{
	PoolSize:     100,              // 连接池大小
	MinIdleConns: 10,               // 最小空闲连接数
	PoolTimeout:  30 * time.Second, // 连接池等待超时
}
```

## 📚 API参考

### 核心接口

```go
// DatabaseConnection 数据库连接接口
type DatabaseConnection interface {
	Close() error
	Ping(ctx context.Context) error
}

// ConfigValidator 配置验证器接口
type ConfigValidator interface {
	Validate() error
	GetDefaults() interface{}
	ApplyDefaults() error
}

// LifecycleManager 生命周期管理器接口
type LifecycleManager interface {
	Initialize(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Shutdown(ctx context.Context) error
	GetState() string
}

// PoolMonitor 连接池监控接口 - 可选的监控抽象
type PoolMonitor interface {
	// RecordConnection 记录连接数变化
	RecordConnection(poolName string, idle, open, inuse int)
	// RecordOperation 记录操作
	RecordOperation(poolName string, operation string, success bool, duration float64)
	// RecordError 记录错误
	RecordError(poolName string, operation string, err error)
}
```

### 数据库操作

```go
// MySQL操作
func NewMySQL(opts *MySQLOptions) (*gorm.DB, error)

// PostgreSQL操作
func NewPostgreSQL(opts *PostgreSQLOptions) (*gorm.DB, error)

// Redis操作
func NewRedis(opts *RedisOptions) (*redis.Client, error)
```

## 🏆 最佳实践

### 1. 连接池监控 (推荐)

```go
// 依赖注入方式启用监控 (推荐)
// 在 wire.go 中配置：
func ProvidePoolMonitor() db.PoolMonitor {
    config := metrics.NewConfig()
    config.Database.Enabled = true
    config.Redis.Enabled = true
    return metrics.NewPoolMonitor(config)
}

// 通过 Wire 自动装配
func ProvideGormDB(cfg *Config, monitor db.PoolMonitor) (*gorm.DB, error) {
    return cfg.MySQLOptions.NewDBWithMonitor(monitor)
}

// 手动注入监控器
func setupDatabaseWithMonitoring() {
    monitor := metrics.NewPoolMonitor(metrics.NewConfig())
    
    opts := &db.MySQLOptions{
        Addr:     "localhost:3306",
        Username: "root",
        Password: "password",
        Database: "myapp",
        Monitor:  monitor, // 注入监控器
    }
    
    gormDB, err := db.NewMySQL(opts)
    // 监控器会自动收集连接池指标
}

// 不启用监控 (原有方式)
func setupDatabaseWithoutMonitoring() {
    opts := &db.MySQLOptions{
        Addr:     "localhost:3306",
        Username: "root", 
        Password: "password",
        Database: "myapp",
        // 不设置 Monitor 字段
    }
    
    gormDB, err := db.NewMySQL(opts)
    // 没有监控开销，数据库包保持单一职责
}
```

### 2. 连接池配置

```go
// 生产环境推荐配置
opts := &db.MySQLOptions{
	MaxIdleConnections:    10,           // 根据并发量调整
	MaxOpenConnections:    100,          // 不要超过数据库最大连接数
	MaxConnectionLifeTime: time.Hour,    // 避免长时间持有连接
}

// 开发环境配置
devOpts := &db.MySQLOptions{
	MaxIdleConnections:    2,
	MaxOpenConnections:    10,
	MaxConnectionLifeTime: time.Minute * 30,
}
```

### 3. 错误处理

```go
// 统一错误处理
func handleDBError(err error, operation string) {
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("记录未找到", "operation", operation)
			return
		}
		
		log.Error("数据库操作失败", 
			"operation", operation,
			"error", err)
	}
}
```

### 4. 事务管理

```go
// 使用事务
func transferMoney(db *gorm.DB, from, to int, amount float64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 扣除转出账户
		if err := tx.Model(&Account{}).Where("id = ?", from).
			Update("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
			return err
		}
		
		// 增加转入账户
		if err := tx.Model(&Account{}).Where("id = ?", to).
			Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
			return err
		}
		
		return nil
	})
}
```

### 5. 性能优化

```go
// 批量操作
func batchInsert(db *gorm.DB, users []User) error {
	// 使用批量插入而不是循环插入
	return db.CreateInBatches(users, 100).Error
}

// 预编译语句
func setupPreparedStatements(db *gorm.DB) {
	// GORM自动处理预编译语句
	db.Config.PrepareStmt = true
}
```

---

## 🤝 贡献指南

1. **Bug报告**: 请在GitHub Issues中详细描述问题
2. **功能请求**: 提交功能需求和使用场景
3. **代码贡献**: Fork项目并提交Pull Request
4. **文档改进**: 帮助完善使用文档和示例

## 📄 许可证

本项目采用MIT许可证，详见LICENSE文件。

---

**版本**: v2.0.0  
**最后更新**: 2025-01-15  
**维护者**: Go-Protoc Team