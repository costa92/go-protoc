# Metrics Package

统一的指标收集和管理系统，为Go微服务提供完整的可观测性解决方案。

## 📋 目录

- [功能特性](#功能特性)
- [快速开始](#快速开始)
- [连接池监控](#连接池监控)
- [架构设计](#架构设计)
- [API参考](#api参考)
- [配置指南](#配置指南)
- [最佳实践](#最佳实践)
- [性能优化](#性能优化)
- [故障排除](#故障排除)

## 🎯 功能特性

### 指标类型支持
- **HTTP指标**: 请求数量、延迟、状态码、连接数等
- **gRPC指标**: 调用统计、消息大小、流状态等
- **数据库指标**: 连接池、查询性能、慢查询等
- **Redis指标**: 命令执行、连接状态、缓存命中率等
- **连接池监控**: 依赖注入式、可选启用、零侵入监控
- **自定义指标**: Counter、Gauge、Histogram、Summary

### 核心特性
- 🚀 **高性能**: 原子操作、对象池、批量处理
- 🔧 **易配置**: 链式配置、预设模板、选项模式
- 🔍 **可观测**: 健康检查、状态监控、分布式追踪
- 🔄 **兼容性**: 向后兼容、渐进迁移、标准接口
- 📊 **丰富指标**: Prometheus标准、自动标签、分层聚合

## 🚀 快速开始

### 基础使用

```go
package main

import (
    "context"
    "time"
    
    "github.com/costa92/go-protoc/v2/pkg/metrics"
)

func main() {
    // 1. 快速启动指标收集
    manager, err := metrics.StartMetrics("my-service",
        metrics.WithHTTP(true, 8080),
        metrics.WithGRPC(true),
        metrics.WithDatabase(true, metrics.DatabaseInfo{
            Name: "main", 
            Type: "mysql", 
            Endpoint: "localhost:3306",
        }),
    )
    if err != nil {
        panic(err)
    }
    defer manager.Stop()

    // 2. 记录HTTP请求
    metrics.RecordHTTPRequest("GET", "/api/users", 200, time.Millisecond*150)
    
    // 3. 记录数据库查询
    metrics.RecordDatabaseQuery("main", "SELECT", "users", time.Millisecond*50, true)
    
    // 4. 检查健康状态
    if metrics.IsHealthy() {
        println("系统运行正常")
    }
    
    // 5. 优雅关闭
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    metrics.Shutdown(ctx)
}
```

### Web服务集成

```go
func setupWebServer() {
    // 使用预设配置
    config := metrics.WebServiceConfig("web-api", 8080)
    manager := metrics.NewManager(config)
    manager.Start()
    
    // 获取中间件
    httpMiddleware, grpcInterceptor, err := metrics.SetupMiddleware()
    if err != nil {
        panic(err)
    }
    
    // 集成到HTTP框架
    router.Use(httpMiddleware.Wrap)
    
    // 集成到gRPC服务器
    server := grpc.NewServer(
        grpc.UnaryInterceptor(grpcInterceptor.UnaryServerInterceptor()),
        grpc.StreamInterceptor(grpcInterceptor.StreamServerInterceptor()),
    )
}
```

## 📊 连接池监控

### 设计理念

连接池监控采用**依赖注入**的设计模式，完美实现了关注点分离：
- `pkg/db` 专注数据库连接管理，保持单一职责
- `pkg/metrics` 专门负责监控功能
- 通过可选的 `PoolMonitor` 接口实现零侵入监控

### 快速集成

#### 1. Wire 依赖注入 (推荐)

```go
// internal/apiserver/wire.go

func ProvidePoolMonitor() db.PoolMonitor {
    config := metrics.NewConfig()
    config.Database.Enabled = true  // 启用数据库监控
    config.Redis.Enabled = true     // 启用Redis监控
    return metrics.NewPoolMonitor(config)
}

func ProvideGormDB(cfg *Config, monitor db.PoolMonitor) (*gorm.DB, error) {
    return cfg.MySQLOptions.NewDBWithMonitor(monitor)
}

// Wire 自动装配
wire.Build(
    ProvidePoolMonitor,
    ProvideGormDB,
    // ... 其他providers
)
```

#### 2. 手动集成

```go
import (
    "github.com/costa92/go-protoc/v2/pkg/db"
    "github.com/costa92/go-protoc/v2/pkg/metrics"
)

func setupDatabaseWithMonitoring() {
    // 创建监控器
    config := metrics.NewConfig()
    config.Database.Enabled = true
    config.CollectInterval = 30 * time.Second
    monitor := metrics.NewPoolMonitor(config)
    
    // 启动监控器
    if err := monitor.Start(); err != nil {
        panic(err)
    }
    
    // 创建带监控的数据库连接
    mysqlOpts := &db.MySQLOptions{
        Addr:     "localhost:3306",
        Username: "root",
        Password: "password", 
        Database: "myapp",
        Monitor:  monitor, // 注入监控器
    }
    
    gormDB, err := db.NewMySQL(mysqlOpts)
    if err != nil {
        panic(err)
    }
    
    // 监控器自动收集连接池指标
    // 指标会暴露在 /metrics 端点
}
```

#### 3. 不启用监控

```go
// 原有方式完全不变，零侵入
mysqlOpts := &db.MySQLOptions{
    Addr:     "localhost:3306",
    Username: "root",
    Password: "password",
    Database: "myapp",
    // 不设置 Monitor 字段即可
}

gormDB, err := db.NewMySQL(mysqlOpts)
// 没有任何监控开销
```

### 监控指标

#### MySQL/PostgreSQL 指标

```promql
# 连接池状态
database_pool_connections{database="myapp", type="mysql", state="idle"}
database_pool_connections{database="myapp", type="mysql", state="open"}
database_pool_connections{database="myapp", type="mysql", state="in_use"}

# 连接池操作
database_pool_operations_total{database="myapp", operation="acquire"}
database_pool_operation_duration_seconds{database="myapp", operation="acquire"}

# 查询统计
database_queries_total{database="myapp", operation="SELECT", status="success"}
database_query_duration_seconds{database="myapp", operation="SELECT"}
database_slow_queries_total{database="myapp", operation="SELECT"}
```

#### Redis 指标

```promql
# Redis连接池
redis_pool_connections{instance="localhost:6379", state="active"}
redis_pool_connections{instance="localhost:6379", state="idle"}

# Redis命令
redis_commands_total{instance="localhost:6379", command="GET", status="success"}
redis_command_duration_seconds{instance="localhost:6379", command="GET"}

# Redis连接池操作
redis_pool_operations_total{instance="localhost:6379", operation="hits"}
redis_pool_operations_total{instance="localhost:6379", operation="misses"}
```

### 监控配置

```go
// 完整配置示例
config := &metrics.Config{
    Namespace:       "myapp",
    CollectInterval: 30 * time.Second,
    
    Database: metrics.DatabaseConfig{
        Enabled:   true,
        SlowQuery: 1 * time.Second,  // 慢查询阈值
        Buckets:   []float64{0.001, 0.01, 0.1, 1, 10}, // 延迟桶
        Databases: []metrics.DatabaseInfo{
            {
                Name:     "main",
                Type:     "mysql", 
                Endpoint: "localhost:3306",
            },
        },
    },
    
    Redis: metrics.RedisConfig{
        Enabled: true,
        Buckets: []float64{0.001, 0.01, 0.1, 1, 10},
    },
}

monitor := metrics.NewPoolMonitor(config)
```

### Grafana 仪表板

```json
{
  "dashboard": {
    "title": "Database Connection Pool Monitoring",
    "panels": [
      {
        "title": "Active Connections", 
        "type": "graph",
        "targets": [
          {
            "expr": "database_pool_connections{state=\"open\"}"
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
        ]
      }
    ]
  }
}
```

### 自定义指标

```go
func useCustomMetrics() {
    // 获取自定义指标接口
    customMetrics, err := metrics.GetCustomMetrics()
    if err != nil {
        panic(err)
    }
    
    // 创建计数器
    counter, err := customMetrics.CreateCounter(
        "business_events_total",
        "Total number of business events",
        []string{"event_type", "status"},
    )
    if err != nil {
        panic(err)
    }
    
    // 使用计数器
    counter.WithLabelValues("user_login", "success").Inc()
    
    // 创建直方图
    histogram, err := customMetrics.CreateHistogram(
        "request_duration_seconds",
        "Request duration in seconds",
        []string{"method", "endpoint"},
        []float64{0.001, 0.01, 0.1, 1, 10},
    )
    if err != nil {
        panic(err)
    }
    
    // 记录持续时间
    start := time.Now()
    // ... 执行业务逻辑 ...
    histogram.WithLabelValues("POST", "/api/users").Observe(time.Since(start).Seconds())
}
```

## 🏗️ 架构设计

### 整体架构

```
┌─────────────────────────────────────┐
│              Manager                │  ← 统一管理器
├─────────────────────────────────────┤
│         Collector Layer             │  ← 收集器层
│  ┌─────┬─────┬─────┬─────┬─────────┐ │
│  │HTTP │gRPC │ DB  │Redis│ Custom  │ │
│  └─────┴─────┴─────┴─────┴─────────┘ │
├─────────────────────────────────────┤
│        Registry & Config            │  ← 注册与配置
├─────────────────────────────────────┤
│         Prometheus                  │  ← 指标存储
└─────────────────────────────────────┘
```

### 核心组件

#### Manager (管理器)
- 统一生命周期管理
- 收集器协调
- 配置分发
- 健康检查

#### Collectors (收集器)
- **HTTPCollector**: HTTP请求指标
- **GRPCCollector**: gRPC调用指标  
- **DatabaseCollector**: 数据库连接池指标
- **RedisCollector**: Redis操作指标
- **CustomCollector**: 自定义业务指标

#### Registry (注册表)
- Prometheus指标注册
- 收集器管理
- 指标聚合

#### Config (配置)
- 统一配置结构
- 选项模式构建
- 验证和默认值

## 📚 API参考

### 管理器API

```go
// 创建管理器
manager := metrics.NewManager(config, opts...)

// 生命周期管理
err := manager.Start()
err := manager.Stop()
running := manager.IsRunning()

// 获取收集器
collector, ok := manager.GetCollector(metrics.MetricTypeHTTP)

// 注册监控目标
err := manager.RegisterTarget(metrics.MetricTypeDatabase, "main", db)
err := manager.UnregisterTarget(metrics.MetricTypeDatabase, "main")

// 获取指标接口
httpMetrics, err := manager.GetHTTPMetrics()
grpcMetrics, err := manager.GetGRPCMetrics()
dbMetrics, err := manager.GetDatabaseMetrics()
redisMetrics, err := manager.GetRedisMetrics()
customMetrics, err := manager.GetCustomMetrics()

// 状态查询
status := manager.GetStatus()
```

### 指标接口

#### HTTP指标
```go
type HTTPMetrics interface {
    RecordRequest(method, path string, statusCode int, duration time.Duration)
    RecordRequestSize(method, path string, size int64)
    RecordResponseSize(method, path string, statusCode int, size int64)
    IncrementActiveConnections()
    DecrementActiveConnections()
    IncrementRequestsInFlight()
    DecrementRequestsInFlight()
}
```

#### 数据库指标
```go
type DatabaseMetrics interface {
    RecordQuery(database, operation, table string, duration time.Duration, success bool)
    RecordConnection(dbName string, state string, count int)
    RecordTransaction(database, operation string, duration time.Duration, success bool)
}
```

#### 自定义指标
```go
type CustomMetrics interface {
    CreateCounter(name, help string, labels []string) (*prometheus.CounterVec, error)
    CreateGauge(name, help string, labels []string) (*prometheus.GaugeVec, error)
    CreateHistogram(name, help string, labels []string, buckets []float64) (*prometheus.HistogramVec, error)
    CreateSummary(name, help string, labels []string, objectives map[float64]float64) (*prometheus.SummaryVec, error)
}
```

### 便捷函数

```go
// 快速启动
manager, err := metrics.StartMetrics(serviceName, options...)
err := metrics.StopMetrics(manager)

// 全局记录
metrics.RecordHTTPRequest(method, path, statusCode, duration)
metrics.RecordGRPCCall(service, method, code, duration)
metrics.RecordDatabaseQuery(database, operation, table, duration, success)
metrics.RecordRedisCommand(instance, command, duration, success)

// 全局注册
err := metrics.RegisterDatabase(name, db)
err := metrics.RegisterRedis(name, client)

// 全局状态
status := metrics.GetStatus()
healthy := metrics.IsHealthy()
err := metrics.Shutdown(ctx)
```

## ⚙️ 配置指南

### 配置结构

```go
type Config struct {
    // 全局配置
    Enabled         bool                  
    CollectInterval time.Duration         
    Registry        prometheus.Registerer 
    Namespace       string                
    ServiceName     string                

    // 各模块配置
    HTTP        HTTPConfig        
    GRPC        GRPCConfig        
    Database    DatabaseConfig    
    Redis       RedisConfig       
    Custom      CustomConfig      
    HealthCheck HealthCheckConfig 
}
```

### 配置选项

```go
// 基础选项
metrics.WithEnabled(true)
metrics.WithServiceName("my-service")
metrics.WithNamespace("app")
metrics.WithCollectInterval(30 * time.Second)
metrics.WithRegistry(prometheus.DefaultRegisterer)

// HTTP配置
metrics.WithHTTP(true, 8080)

// gRPC配置  
metrics.WithGRPC(true)

// 数据库配置
metrics.WithDatabase(true, metrics.DatabaseInfo{
    Name: "main",
    Type: "mysql", 
    Endpoint: "localhost:3306",
})

// Redis配置
metrics.WithRedis(true, metrics.RedisInfo{
    Name: "cache",
    Endpoint: "localhost:6379", 
})

// 自定义指标
metrics.WithCustom(true)

// 健康检查
metrics.WithHealthCheck(true, 30*time.Second, 5*time.Second)
```

### 预设配置

```go
// Web服务配置
config := metrics.WebServiceConfig("web-api", 8080)

// 数据库服务配置  
config := metrics.DatabaseServiceConfig("db-service", databases...)

// 微服务配置
config := metrics.MicroserviceConfig("micro-service", 8080, databases, redisInstances)

// 最小配置
config := metrics.MinimalConfig("simple-service")
```

### YAML配置文件

```yaml
metrics:
  enabled: true
  service_name: "my-service"
  namespace: "app"
  collect_interval: "30s"
  
  http:
    enabled: true
    port: 8080
    path: "/metrics"
    exclude_paths: ["/health", "/metrics"]
    buckets: [0.001, 0.01, 0.1, 1, 10]
    
  grpc:
    enabled: true
    buckets: [0.001, 0.01, 0.1, 1, 10]
    
  database:
    enabled: true
    slow_query: "1s"
    databases:
      - name: "main"
        type: "mysql"
        endpoint: "localhost:3306"
        
  redis:
    enabled: true
    instances:
      - name: "cache"
        endpoint: "localhost:6379"
        
  custom:
    enabled: true
    
  health_check:
    enabled: true
    interval: "30s"
    timeout: "5s"
```

## 📊 最佳实践

### 1. 指标命名规范

```go
// 好的命名
http_requests_total{method="GET", path="/api/users", status="200"}
database_query_duration_seconds{database="main", operation="SELECT", table="users"}
redis_commands_total{instance="cache", command="GET", status="success"}

// 避免的命名
requests  // 太泛化
db_time   // 不清晰
cache_ops // 缺乏单位
```

### 2. 标签使用

```go
// 合理的标签基数
labels := []string{"method", "status"}          // Good: 低基数
labels := []string{"method", "status", "user"}  // Bad: 高基数

// 标签值规范化
status := "2xx"  // Good: 归类状态码
status := "200"  // OK: 具体状态码
status := "success_user_login_from_mobile"  // Bad: 过于具体
```

### 3. 性能优化

```go
// 使用标签池减少内存分配
metrics.RecordHTTPRequest("GET", "/api/users", 200, duration)

// 批量记录
batch := []MetricRecord{...}
metrics.RecordBatch(batch)

// 条件记录
if metrics.IsEnabled() {
    metrics.RecordRequest(...)
}
```

### 4. 错误处理

```go
// 优雅降级
httpMetrics, err := manager.GetHTTPMetrics()
if err != nil {
    log.Warn("HTTP metrics not available", "error", err)
    return // 不影响主业务逻辑
}

// 健康检查
if !metrics.IsHealthy() {
    log.Warn("Metrics system unhealthy")
    // 考虑是否需要告警
}
```

### 5. 生命周期管理

```go
func main() {
    // 启动时初始化
    config := metrics.WebServiceConfig("my-service", 8080)
    manager := metrics.NewManager(config)
    
    if err := manager.Start(); err != nil {
        log.Fatal("Failed to start metrics", "error", err)
    }
    
    // 注册清理函数
    defer func() {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        
        if err := metrics.Shutdown(ctx); err != nil {
            log.Error("Failed to shutdown metrics", "error", err)
        }
    }()
    
    // 信号处理
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        manager.Stop()
        os.Exit(0)
    }()
}
```

## 🚀 性能优化

### 内存优化

```go
// 1. 对象池使用
type MetricPool struct {
    labelSets sync.Pool
}

func (p *MetricPool) GetLabelSet() *LabelSet {
    return p.labelSets.Get().(*LabelSet)
}

func (p *MetricPool) PutLabelSet(ls *LabelSet) {
    ls.Reset()
    p.labelSets.Put(ls)
}

// 2. 预分配容量
labels := make(prometheus.Labels, 4)  // 预分配常用容量

// 3. 避免字符串拼接
// Bad
metric := "http_request_" + method + "_total"

// Good  
metric := fmt.Sprintf("http_request_%s_total", method)

// Better
metric := strings.Builder{}
metric.WriteString("http_request_")
metric.WriteString(method)
metric.WriteString("_total")
```

### 并发优化

```go
// 1. 原子操作
type Collector struct {
    enabled atomic.Bool
    running atomic.Bool
}

func (c *Collector) IsEnabled() bool {
    return c.enabled.Load()  // 无锁读取
}

// 2. 分片减少锁竞争
type ShardedCounter struct {
    shards []shardedCounter
    mask   uint64
}

func (s *ShardedCounter) Inc(labels prometheus.Labels) {
    shard := s.getShard(labels)
    shard.Inc(labels)
}

// 3. 批量处理
func (c *Collector) ProcessBatch(metrics []Metric) {
    for _, metric := range metrics {
        c.processMetric(metric)
    }
}
```

### 采样优化

```go
// 1. 采样率控制
type SamplingCollector struct {
    sampleRate float64
    rand       *rand.Rand
}

func (s *SamplingCollector) ShouldSample() bool {
    return s.rand.Float64() < s.sampleRate
}

// 2. 条件收集
func (c *Collector) RecordIfEnabled(metric Metric) {
    if !c.enabled.Load() {
        return
    }
    c.record(metric)
}

// 3. 缓存频繁查询
type CachedCollector struct {
    cache map[string]*prometheus.CounterVec
    mu    sync.RWMutex
}
```

## 🔧 故障排除

### 常见问题

#### 1. 指标未显示

**症状**: Prometheus中看不到指标

**排查步骤**:
```go
// 检查管理器状态
status := manager.GetStatus()
fmt.Printf("Manager status: %+v\n", status)

// 检查收集器是否运行
collector, ok := manager.GetCollector(metrics.MetricTypeHTTP)
if !ok {
    fmt.Println("HTTP collector not found")
} else {
    fmt.Printf("HTTP collector running: %v\n", collector.IsRunning())
}

// 检查配置
config := manager.GetConfig()
fmt.Printf("HTTP enabled: %v\n", config.HTTP.Enabled)

// 检查注册表
gatherer := manager.GetRegistry().GetGatherer()
families, err := gatherer.Gather()
if err != nil {
    fmt.Printf("Gather error: %v\n", err)
}
fmt.Printf("Metrics families: %d\n", len(families))
```

#### 2. 性能问题

**症状**: 高CPU使用率或内存泄漏

**排查步骤**:
```go
// 检查收集间隔
if config.CollectInterval < time.Second {
    fmt.Println("Collection interval too short")
}

// 检查标签基数
for _, family := range families {
    if len(family.GetMetric()) > 1000 {
        fmt.Printf("High cardinality metric: %s\n", family.GetName())
    }
}

// 启用性能分析
import _ "net/http/pprof"
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

#### 3. 健康检查失败

**症状**: IsHealthy()返回false

**排查步骤**:
```go
// 获取详细健康状态
healthChecker := metrics.GetGlobalHealthChecker()
healthStatus := healthChecker.GetStatus()

for component, check := range healthStatus {
    if check.Status != metrics.HealthStatusHealthy {
        fmt.Printf("Component %s unhealthy: %s\n", component, check.Message)
    }
}

// 检查各收集器状态
collectors := []metrics.MetricType{
    metrics.MetricTypeHTTP,
    metrics.MetricTypeGRPC,
    metrics.MetricTypeDatabase,
    metrics.MetricTypeRedis,
}

for _, t := range collectors {
    if collector, ok := manager.GetCollector(t); ok {
        fmt.Printf("%s running: %v\n", t, collector.IsRunning())
    }
}
```

### 调试工具

```go
// 1. 启用详细日志
config.DebugMode = true

// 2. 指标导出
func exportMetrics() {
    gatherer := prometheus.DefaultGatherer
    families, _ := gatherer.Gather()
    
    for _, family := range families {
        fmt.Printf("Family: %s\n", family.GetName())
        for _, metric := range family.GetMetric() {
            fmt.Printf("  Metric: %+v\n", metric)
        }
    }
}

// 3. 状态端点
http.HandleFunc("/debug/metrics/status", func(w http.ResponseWriter, r *http.Request) {
    status := metrics.GetStatus()
    json.NewEncoder(w).Encode(status)
})
```

### 监控告警

```yaml
# Prometheus告警规则
groups:
- name: metrics_system
  rules:
  - alert: MetricsCollectorDown
    expr: up{job="metrics"} == 0
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "Metrics collector is down"
      
  - alert: HighMetricCardinality
    expr: prometheus_tsdb_symbol_table_size_bytes > 100000000
    for: 10m
    labels:
      severity: warning
    annotations:
      summary: "High metric cardinality detected"
```

## 📈 监控指标

### 系统指标

```
# HTTP指标
app_http_requests_total{method, path, status_code}
app_http_request_duration_seconds{method, path, status_code}
app_http_request_size_bytes{method, path}
app_http_response_size_bytes{method, path, status_code}
app_http_active_connections
app_http_requests_in_flight

# gRPC指标
app_grpc_calls_total{service, method, code, type}
app_grpc_call_duration_seconds{service, method, code, type}
app_grpc_messages_sent_total{service, method, type}
app_grpc_messages_received_total{service, method, type}
app_grpc_message_sent_size_bytes{service, method, type}
app_grpc_message_received_size_bytes{service, method, type}
app_grpc_active_streams{service, method, type}

# 数据库指标
app_database_pool_connections{database, type, state, endpoint}
app_database_pool_operations_total{database, type, operation, endpoint}
app_database_pool_operation_duration_seconds{database, type, operation, endpoint}
app_database_queries_total{database, operation, table, status}
app_database_query_duration_seconds{database, operation, table}
app_database_slow_queries_total{database, operation, table}
app_database_transactions_total{database, operation, status}
app_database_transaction_duration_seconds{database, operation}

# Redis指标
app_redis_pool_connections{instance, state, endpoint}
app_redis_pool_operations_total{instance, operation, endpoint}
app_redis_commands_total{instance, command, status}
app_redis_command_duration_seconds{instance, command}
app_redis_key_operations_total{instance, operation, key_type}
```

### 健康检查指标

```
# 健康状态
health_status{component}
health_checks_total{component, status}
health_check_duration_seconds{component}
health_last_check_timestamp{component}
```

---

## 🤝 贡献指南

1. **问题报告**: 请在GitHub Issues中报告bug
2. **功能请求**: 提交功能需求和改进建议
3. **代码贡献**: Fork项目并提交Pull Request
4. **文档改进**: 帮助完善文档和示例

## 📄 许可证

本项目采用MIT许可证，详见LICENSE文件。

---

**版本**: v2.0.0  
**最后更新**: 2025-01-15  
**维护者**: Go-Protoc Team