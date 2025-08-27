# OTEL SaaS Integration Example

演示与多个 SaaS 平台的集成，展示如何从不同平台查询和分析遥测数据。

## 🎯 功能特性

- **多平台支持**: VictoriaLogs, Jaeger, Prometheus
- **统一查询接口**: 跨平台数据查询和聚合
- **数据关联分析**: 追踪、指标、日志数据关联
- **自定义仪表板**: 创建综合分析视图

## 🚀 运行示例

```bash
cd examples/otel/saas
go run main.go
```

## 📊 示例内容

### 1. ExampleSaaSIntegration()

基础 SaaS 平台集成演示：

- 配置多个 SaaS 平台连接
- 统一查询接口使用
- 数据聚合和关联分析

### 2. SaaSManager 功能

完整的多平台管理：

- `AddPlatform()` - 添加平台配置
- `Query()` - 统一查询接口
- `GetAnalytics()` - 分析和聚合
- `CreateDashboard()` - 自定义仪表板

## 🔧 支持的平台

### VictoriaLogs (日志分析)

```go
config := SaaSConfig{
    Platform: VictoriaLogs,
    Config: map[string]interface{}{
        "endpoint": "http://127.0.0.1:9428",
        "timeout":  "10s",
    },
}
```

**查询示例**:

```bash
# Web UI
http://127.0.0.1:9428/select/vmui/

# API 查询
curl -s "http://127.0.0.1:9428/select/logsql/query" \
  -d 'query=service.name:"your-service" AND level:error'
```

### Jaeger (链路追踪)

```go
config := SaaSConfig{
    Platform: Jaeger,
    Config: map[string]interface{}{
        "endpoint": "http://127.0.0.1:16686",
        "timeout":  "10s",
    },
}
```

**查询示例**:

```bash
# Web UI
http://127.0.0.1:16686

# API 查询
curl "http://127.0.0.1:16686/api/traces?service=your-service"
```

### Prometheus (指标监控)

```go
config := SaaSConfig{
    Platform: Prometheus,
    Config: map[string]interface{}{
        "endpoint": "http://127.0.0.1:9090",
        "timeout":  "10s",
    },
}
```

**查询示例**:

```bash
# Web UI
http://127.0.0.1:9090

# API 查询
curl "http://127.0.0.1:9090/api/v1/query?query=up{job=\"your-service\"}"
```

## 📈 数据关联分析

### 跨平台查询示例

```go
manager := NewSaaSManager()

// 配置多个平台
manager.AddPlatform(VictoriaLogs, logsConfig)
manager.AddPlatform(Jaeger, tracesConfig)
manager.AddPlatform(Prometheus, metricsConfig)

// 关联查询
analytics, err := manager.GetAnalytics(AnalyticsRequest{
    Services:  []string{"user-service", "order-service"},
    TimeRange: TimeRange{Start: startTime, End: endTime},
    Metrics:   []string{"error_rate", "response_time"},
})
```

### 综合仪表板

```go
dashboard := CustomDashboard{
    Name: "Service Overview",
    Panels: []Panel{
        {Type: "error-logs", Platform: VictoriaLogs},
        {Type: "trace-analysis", Platform: Jaeger},
        {Type: "performance-metrics", Platform: Prometheus},
    },
}

err := manager.CreateDashboard(dashboard)
```

## 🔍 查询语法示例

### VictoriaLogs 查询

```bash
# 错误日志查询
query='level:error AND service.name:"user-service" AND _time:>now-1h'

# 性能分析查询
query='operation:"database_query" AND duration:>100ms'

# 用户行为分析
query='user_id:"12345" AND _time:>now-24h | stats count() by operation'
```

### Jaeger 追踪查询

```bash
# 服务调用链查询
service=user-service&operation=user_login

# 错误追踪查询
service=user-service&tags={"error":"true"}

# 性能分析查询
service=user-service&minDuration=100ms
```

### Prometheus 指标查询

```bash
# 错误率查询
rate(http_requests_total{status=~"5.."}[5m])

# 响应时间分位数
histogram_quantile(0.95, http_request_duration_seconds_bucket)

# 服务可用性
up{job="user-service"}
```

## 📊 自定义分析

### 添加新的 SaaS 平台

```go
// 实现 SaaSClient 接口
type CustomPlatformClient struct {
    endpoint string
    apiKey   string
}

func (c *CustomPlatformClient) Query(ctx context.Context, req QueryRequest) (*QueryResult, error) {
    // 实现查询逻辑
}

func (c *CustomPlatformClient) TestConnection(ctx context.Context) error {
    // 实现连接测试
}

// 添加到管理器
manager.AddPlatform("custom-platform", customConfig)
```

## 🔧 故障排除

### 连接问题

```bash
# 检查平台服务状态
curl http://127.0.0.1:9428/health    # VictoriaLogs
curl http://127.0.0.1:16686/health   # Jaeger
curl http://127.0.0.1:9090/-/healthy # Prometheus
```

### 数据查询问题

```bash
# 检查数据是否到达平台
curl http://127.0.0.1:9428/select/logsql/query -d 'query=*' | head

# 检查时间范围设置
curl "http://127.0.0.1:16686/api/services"

# 验证查询语法
curl "http://127.0.0.1:9090/api/v1/label/__name__/values"
```

## 📚 相关文档

- [OTEL 主文档](../README.md)
- [Gateway 示例](../gateway/)
- [Complete 示例](../complete/)
