# OTEL Complete Example

端到端完整演示，展示 Agent → Gateway → SaaS 的完整架构工作流程。

## 🎯 功能特性

- **完整架构演示**: 三层架构的端到端数据流
- **多业务场景**: 电商、认证、数据处理等真实场景
- **并发工作流**: 多场景并行执行和监控
- **集成测试**: 自动化验证和故障诊断

## 🚀 运行示例

```bash
cd examples/otel/complete
go run main.go
```

## 📊 示例内容

### 1. CompleteOTELExample()
完整的端到端演示：
- 初始化 OTEL Agent
- 验证 Gateway 连通性
- 配置 SaaS 平台集成
- 运行综合工作流
- 监控数据流状态

### 2. 业务场景模拟

```go
scenarios := []WorkflowScenario{
    {Name: "E-commerce Order Processing", Duration: 15*time.Second, Operations: 10},
    {Name: "User Authentication Flow", Duration: 10*time.Second, Operations: 5},
    {Name: "API Gateway Traffic", Duration: 12*time.Second, Operations: 8},
}
```

每个场景包含：
- **数据库操作**: PostgreSQL 查询和事务
- **外部 API 调用**: 第三方服务集成
- **缓存操作**: Redis 读写操作
- **业务逻辑处理**: 复杂计算和验证

## 🔧 架构组件

### OTEL Agent 配置
```go
config := OTELAgentConfig{
    ServiceName:         "complete-example",
    ServiceVersion:      "v1.2.0",
    Environment:         "demo",
    GatewayHTTPEndpoint: "http://127.0.0.1:4328",
    GatewayGRPCEndpoint: "http://127.0.0.1:4327",
    EnableLogs:          true,
    EnableMetrics:       true,
    EnableTraces:        true,
}
```

### 数据流监控
```go
// 实时监控工作流进度
go monitorWorkflowProgress(ctx, gateway, saasManager)

// 检查 Gateway 处理状态
metrics, err := gateway.GetMetrics(ctx)

// 查询 SaaS 平台数据
result, err := client.Query(ctx, queryRequest)
```

## 📈 生成的遥测数据

### 追踪数据 (Traces)
- **根 Span**: 业务场景 (如 "E-commerce Order Processing")
- **子 Span**: 具体操作 (如 "database_operation", "external_api_call")
- **属性标签**: 场景名称、操作类型、性能指标

### 指标数据 (Metrics)
```go
// 操作计数器
operationCounter.Add(ctx, 1,
    attribute.String("scenario", scenarioName),
    attribute.String("status", "success"),
)

// 响应时间直方图
operationDuration.Record(ctx, duration,
    attribute.String("operation", "business_logic"),
)
```

### 日志数据 (Logs)
```go
logger.InfoContext(ctx, "Operation executed",
    "scenario", scenarioName,
    "operation_id", operationID,
    "status", "success",
    "processing_time_ms", processingTime,
)
```

## 📊 验证和监控

### 自动化集成测试
```go
func TestCompleteOTELIntegration() error {
    // 快速连通性测试
    // 生成测试数据
    // 验证数据处理
    // 检查各平台数据可用性
}
```

### 数据验证步骤

1. **Gateway 处理验证**
   ```bash
   docker logs proj-otel-collector
   curl http://127.0.0.1:8888/metrics | grep otelcol
   ```

2. **Jaeger 追踪验证**
   ```bash
   # Web UI
   http://127.0.0.1:16686
   # 搜索服务: complete-example
   ```

3. **VictoriaLogs 日志验证**
   ```bash
   # Web UI  
   http://127.0.0.1:9428/select/vmui/
   # 查询: service.name:complete-example
   ```

## 🔍 监控面板

### 实时数据流监控
- **Gateway 指标**: 处理的 traces/metrics/logs 数量
- **SaaS 平台状态**: 各平台数据接收状态
- **工作流进度**: 场景执行状态和完成度
- **错误统计**: 连接失败和处理错误统计

### 性能分析
```bash
# 查看处理延迟
curl http://127.0.0.1:8888/metrics | grep duration

# 检查内存使用
docker stats proj-otel-collector

# 监控数据吞吐量
watch -n 5 'curl -s http://127.0.0.1:8888/metrics | grep -E "(processed|exported)_total"'
```

## 🛠️ 自定义扩展

### 添加新的业务场景
```go
newScenario := WorkflowScenario{
    Name:       "Custom Business Process",
    Duration:   20 * time.Second,
    Operations: 15,
}

scenarios = append(scenarios, newScenario)
```

### 自定义遥测数据
```go
func executeCustomOperation(ctx context.Context, scenarioName string) {
    // 创建自定义追踪
    ctx, span := tracer.Start(ctx, "custom-operation")
    defer span.End()
    
    // 记录自定义指标
    customMetric.Record(ctx, value, attributes...)
    
    // 生成结构化日志
    logger.InfoContext(ctx, "Custom operation", fields...)
}
```

## 🔧 故障排除

### 常见问题解决

1. **Agent 初始化失败**
   ```bash
   # 检查 Gateway 状态
   curl http://127.0.0.1:13133
   docker restart proj-otel-collector
   ```

2. **数据未显示**
   ```bash
   # 等待数据处理 (1-2分钟)
   # 检查 SaaS 平台连接
   curl http://127.0.0.1:16686/api/services
   ```

3. **性能问题**
   ```bash
   # 调整批处理大小
   # 检查采样率配置
   # 监控资源使用情况
   ```

### 调试模式
```bash
export OTEL_LOG_LEVEL=debug
go run main.go
```

## 📚 相关文档

- [OTEL 主文档](../README.md)
- [Agent 示例](../agent/)
- [Gateway 示例](../gateway/)
- [SaaS 示例](../saas/)