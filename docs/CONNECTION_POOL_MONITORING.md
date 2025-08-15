# 连接池监控使用指南

本指南详细介绍了项目中的依赖注入式连接池监控功能的使用方法和最佳实践。

## 📋 目录

- [概述](#概述)
- [设计原理](#设计原理)
- [快速开始](#快速开始)
- [配置选项](#配置选项)
- [监控指标](#监控指标)
- [集成方式](#集成方式)
- [故障排除](#故障排除)
- [最佳实践](#最佳实践)

## 🎯 概述

### 核心特性

- **完全解耦**: `pkg/db` 保持单一职责，监控逻辑完全分离
- **可选启用**: 通过依赖注入实现可选监控，无性能损耗
- **零侵入**: 原有代码无需任何修改
- **丰富指标**: 提供完整的连接池状态和性能指标
- **实时监控**: 自动收集并暴露 Prometheus 指标

### 架构关系

```
                    ┌─────────────────────────────────┐
                    │         Wire 依赖注入          │
                    │    (可选装配监控器)            │
                    └─────────────┬───────────────────┘
                                  │
                    ┌─────────────▼───────────────────┐
                    │         pkg/db                  │
                    │  • MySQLOptions.Monitor         │
                    │  • RedisOptions.Monitor         │ 
                    │  • 单一职责：连接管理          │
                    └─────────────┬───────────────────┘
                                  │ 接口调用
                    ┌─────────────▼───────────────────┐
                    │       pkg/metrics               │
                    │  • PoolMonitorImpl              │
                    │  • DatabaseCollector            │
                    │  • RedisCollector               │
                    │  • Prometheus 指标暴露         │
                    └─────────────────────────────────┘
```

## 🚀 快速开始

### 1. 启用监控 (推荐使用)

监控功能已经通过 Wire 自动装配到项目中：

```bash
# 启动 API 服务器，监控功能自动启用
make run-api

# 启动日志会显示：
# "Connection pool metrics collection is enabled through dependency injection"
# "Metrics will be automatically collected by the PoolMonitor when database connections are established"
```

### 2. 查看监控指标

```bash
# 查看所有指标
curl http://localhost:8080/metrics

# 仅查看数据库相关指标
curl http://localhost:8080/metrics | grep -E "(database_|redis_)"

# 查看连接池状态
curl http://localhost:8080/metrics | grep "pool_connections"
```

### 3. 不启用监控

如果需要禁用监控，只需要修改 Wire 配置：

```go
// internal/apiserver/wire.go

// 注释掉监控器提供者
// ProvidePoolMonitor,

// 使用原有的数据库提供者
func ProvideGormDB(cfg *Config) (*gorm.DB, error) {
    return cfg.MySQLOptions.NewDB()  // 不注入监控器
}
```

## ⚙️ 配置选项

### 监控器配置

```go
// internal/apiserver/wire.go

func ProvidePoolMonitor() db.PoolMonitor {
    config := metrics.NewConfig()
    
    // 数据库监控配置
    config.Database.Enabled = true                    // 启用数据库监控
    config.Database.SlowQuery = 1 * time.Second      // 慢查询阈值
    config.Database.Buckets = []float64{             // 延迟分布桶
        0.001, 0.01, 0.1, 1, 10,
    }
    
    // Redis监控配置
    config.Redis.Enabled = true                      // 启用Redis监控
    config.Redis.Buckets = []float64{                // 延迟分布桶
        0.001, 0.01, 0.1, 1, 10,
    }
    
    // 全局配置
    config.Namespace = "apiserver"                    // 指标命名空间
    config.CollectInterval = 30 * time.Second        // 收集间隔
    
    return metrics.NewPoolMonitor(config)
}
```

### 环境变量配置

```bash
# 可通过环境变量控制监控行为
export METRICS_DATABASE_ENABLED=true
export METRICS_REDIS_ENABLED=true
export METRICS_COLLECT_INTERVAL=30s
export METRICS_SLOW_QUERY_THRESHOLD=1s
```

## 📊 监控指标

### MySQL/PostgreSQL 指标

| 指标名称 | 类型 | 描述 | 标签 |
|---------|------|------|------|
| `database_pool_connections` | Gauge | 连接池连接数 | `database`, `type`, `state`, `endpoint` |
| `database_pool_operations_total` | Counter | 连接池操作计数 | `database`, `type`, `operation`, `endpoint` |
| `database_pool_operation_duration_seconds` | Histogram | 连接池操作耗时 | `database`, `type`, `operation`, `endpoint` |
| `database_queries_total` | Counter | 查询总数 | `database`, `operation`, `table`, `status` |
| `database_query_duration_seconds` | Histogram | 查询耗时 | `database`, `operation`, `table` |
| `database_slow_queries_total` | Counter | 慢查询数量 | `database`, `operation`, `table` |

### Redis 指标

| 指标名称 | 类型 | 描述 | 标签 |
|---------|------|------|------|
| `redis_pool_connections` | Gauge | Redis连接池状态 | `instance`, `state`, `endpoint` |
| `redis_pool_operations_total` | Counter | Redis连接池操作 | `instance`, `operation`, `endpoint` |
| `redis_commands_total` | Counter | Redis命令执行 | `instance`, `command`, `status` |
| `redis_command_duration_seconds` | Histogram | Redis命令耗时 | `instance`, `command` |

### 示例查询

```promql
# 数据库连接池使用率
database_pool_connections{state="in_use"} / database_pool_connections{state="open"} * 100

# 平均查询延迟
rate(database_query_duration_seconds_sum[5m]) / rate(database_query_duration_seconds_count[5m])

# Redis命令错误率
rate(redis_commands_total{status="error"}[5m]) / rate(redis_commands_total[5m]) * 100

# 慢查询率
rate(database_slow_queries_total[5m]) / rate(database_queries_total[5m]) * 100
```

## 🔧 集成方式

### 1. Wire 自动装配 (推荐)

项目默认使用这种方式，已在 `internal/apiserver/wire.go` 中配置：

```go
func InitializeWebServer(...) (server.Server, error) {
    wire.Build(
        // 监控器提供者
        ProvidePoolMonitor,
        
        // 带监控的数据库提供者
        ProvideGormDB,
        
        // ... 其他提供者
    )
}
```

### 2. 手动集成

如需手动控制监控器的创建和配置：

```go
func setupDatabaseWithCustomMonitoring() {
    // 创建自定义配置
    config := &metrics.Config{
        Namespace:       "myapp",
        CollectInterval: 15 * time.Second,
        Database: metrics.DatabaseConfig{
            Enabled:   true,
            SlowQuery: 500 * time.Millisecond,
        },
    }
    
    // 创建监控器
    monitor := metrics.NewPoolMonitor(config)
    
    // 启动监控器
    if err := monitor.Start(); err != nil {
        panic(err)
    }
    defer monitor.Stop()
    
    // 创建带监控的数据库选项
    mysqlOpts := &db.MySQLOptions{
        Addr:     "localhost:3306",
        Username: "root",
        Password: "password",
        Database: "myapp",
        Monitor:  monitor,  // 注入监控器
    }
    
    // 创建数据库连接
    gormDB, err := db.NewMySQL(mysqlOpts)
    if err != nil {
        panic(err)
    }
    
    // 监控器会自动收集连接池指标
}
```

### 3. 测试环境集成

在测试中使用 Mock 监控器：

```go
func TestWithMockMonitoring(t *testing.T) {
    // 创建 Mock 监控器
    mockMonitor := &MockPoolMonitor{}
    
    // 创建测试数据库
    mysqlOpts := &db.MySQLOptions{
        Addr:     "localhost:3306",
        Username: "test",
        Password: "test",
        Database: "test_db",
        Monitor:  mockMonitor,
    }
    
    db, err := db.NewMySQL(mysqlOpts)
    require.NoError(t, err)
    
    // 验证监控器调用
    assert.True(t, mockMonitor.RegisterDatabaseCalled)
}

type MockPoolMonitor struct {
    RegisterDatabaseCalled bool
}

func (m *MockPoolMonitor) RecordConnection(poolName string, idle, open, inuse int) {}
func (m *MockPoolMonitor) RecordOperation(poolName string, operation string, success bool, duration float64) {}
func (m *MockPoolMonitor) RecordError(poolName string, operation string, err error) {}
func (m *MockPoolMonitor) RegisterDatabase(name string, target interface{}) error {
    m.RegisterDatabaseCalled = true
    return nil
}
```

## 🛠️ 故障排除

### 常见问题

#### 1. 指标没有数据

**症状**: `/metrics` 端点没有数据库相关指标

**解决方案**:
```bash
# 检查监控器是否启用
curl http://localhost:8080/metrics | grep -c "database_"

# 检查启动日志
make run-api | grep -i "metrics"

# 确认配置正确
grep -n "ProvidePoolMonitor" internal/apiserver/wire.go
```

#### 2. 监控影响性能

**症状**: 数据库操作变慢

**解决方案**:
```go
// 调整收集间隔
config.CollectInterval = 60 * time.Second  // 增加到60秒

// 或者临时禁用监控
config.Database.Enabled = false
config.Redis.Enabled = false
```

#### 3. 内存占用过高

**症状**: 监控器占用大量内存

**解决方案**:
```go
// 减少延迟桶数量
config.Database.Buckets = []float64{0.01, 0.1, 1, 10}  // 减少桶

// 或设置指标过期时间
config.MetricsTTL = 5 * time.Minute
```

### 调试模式

启用调试模式查看详细信息：

```bash
# 设置调试级别
export LOG_LEVEL=debug

# 启动服务
make run-api

# 查看监控器状态
curl http://localhost:8080/debug/metrics/status
```

## 🏆 最佳实践

### 1. 生产环境配置

```go
func ProvidePoolMonitor() db.PoolMonitor {
    config := metrics.NewConfig()
    
    // 生产环境推荐配置
    config.CollectInterval = 30 * time.Second    // 平衡性能和实时性
    config.Database.SlowQuery = 1 * time.Second  // 合理的慢查询阈值
    
    // 精简的延迟桶，减少内存占用
    config.Database.Buckets = []float64{0.01, 0.1, 1, 5, 30}
    config.Redis.Buckets = []float64{0.001, 0.01, 0.1, 1, 10}
    
    return metrics.NewPoolMonitor(config)
}
```

### 2. 开发环境配置

```go
func ProvidePoolMonitor() db.PoolMonitor {
    config := metrics.NewConfig()
    
    // 开发环境：更详细的监控
    config.CollectInterval = 10 * time.Second     // 更频繁的收集
    config.Database.SlowQuery = 100 * time.Millisecond  // 更严格的慢查询
    
    // 更多的延迟桶，便于调试
    config.Database.Buckets = prometheus.ExponentialBuckets(0.001, 2, 15)
    
    return metrics.NewPoolMonitor(config)
}
```

### 3. 监控告警配置

```yaml
# Prometheus 告警规则示例
groups:
- name: database_connection_pool
  rules:
  - alert: HighConnectionPoolUsage
    expr: database_pool_connections{state="in_use"} / database_pool_connections{state="open"} > 0.8
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Database connection pool usage is high"
      
  - alert: SlowQueryRateHigh
    expr: rate(database_slow_queries_total[5m]) / rate(database_queries_total[5m]) > 0.1
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "Slow query rate is above 10%"
```

### 4. Grafana 仪表板

```json
{
  "dashboard": {
    "title": "Connection Pool Monitoring",
    "panels": [
      {
        "title": "Database Connections",
        "type": "graph",
        "targets": [
          {
            "expr": "database_pool_connections",
            "legendFormat": "{{database}} - {{state}}"
          }
        ]
      },
      {
        "title": "Connection Pool Utilization",
        "type": "singlestat",
        "targets": [
          {
            "expr": "database_pool_connections{state=\"in_use\"} / database_pool_connections{state=\"open\"} * 100"
          }
        ],
        "thresholds": "70,85"
      },
      {
        "title": "Query Performance",
        "type": "graph", 
        "targets": [
          {
            "expr": "rate(database_query_duration_seconds_sum[5m]) / rate(database_query_duration_seconds_count[5m])",
            "legendFormat": "Avg Query Time"
          }
        ]
      }
    ]
  }
}
```

## 📚 进阶用法

### 自定义监控扩展

```go
// 扩展监控器以支持自定义指标
type CustomPoolMonitor struct {
    *metrics.PoolMonitorImpl
    customMetrics *prometheus.CounterVec
}

func NewCustomPoolMonitor(config *metrics.Config) db.PoolMonitor {
    base := metrics.NewPoolMonitor(config)
    
    custom := &CustomPoolMonitor{
        PoolMonitorImpl: base.(*metrics.PoolMonitorImpl),
        customMetrics: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "custom_database_events_total",
                Help: "Custom database events",
            },
            []string{"event_type", "database"},
        ),
    }
    
    prometheus.MustRegister(custom.customMetrics)
    return custom
}

func (c *CustomPoolMonitor) RecordCustomEvent(eventType, database string) {
    c.customMetrics.WithLabelValues(eventType, database).Inc()
}
```

### 动态配置

```go
// 支持运行时配置更新
type DynamicPoolMonitor struct {
    *metrics.PoolMonitorImpl
    configMutex sync.RWMutex
    config      *metrics.Config
}

func (d *DynamicPoolMonitor) UpdateConfig(newConfig *metrics.Config) {
    d.configMutex.Lock()
    defer d.configMutex.Unlock()
    
    d.config = newConfig
    // 重新初始化监控器
    d.reinitialize()
}
```

---

## 📞 支持与反馈

如果在使用连接池监控功能时遇到问题：

1. **检查文档**: 首先查看本指南和相关 README 文档
2. **查看日志**: 检查应用启动日志中的监控相关信息
3. **验证配置**: 确认 Wire 配置和监控器设置正确
4. **测试指标**: 通过 `/metrics` 端点验证指标是否正常暴露

**版本**: v2.0.0  
**最后更新**: 2025-08-15  
**维护者**: Go-Protoc Team