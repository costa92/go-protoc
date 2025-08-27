# OTEL Simple Demo

简化的 OpenTelemetry 演示，使用 HTTP 协议展示基本的追踪功能。

## 🎯 功能特性

- **简单易懂**: 最小化配置，专注核心功能
- **HTTP 协议**: 使用 OTLP HTTP 导出器 (端口 4328)
- **基础追踪**: 演示 Span 创建和属性设置
- **场景模拟**: 5个典型业务场景

## 🚀 运行示例

```bash
cd examples/otel/simple-demo  
go run main.go
```

## 📊 示例内容

### 生成的业务场景
1. **user-login** - 用户登录流程
2. **order-processing** - 订单处理流程
3. **inventory-check** - 库存检查流程
4. **payment-processing** - 支付处理流程
5. **notification-sending** - 通知发送流程

### 每个场景包含的操作
- **validate-input** (50ms) - 输入验证
- **database-query** (100ms) - 数据库查询
- **business-logic** (75ms) - 业务逻辑处理
- **response-formatting** (25ms) - 响应格式化

## 🔧 技术细节

### OTLP HTTP 配置
```go
exporter, err := otlptracehttp.New(ctx,
    otlptracehttp.WithEndpoint("127.0.0.1:4328"),
    otlptracehttp.WithInsecure(),
)
```

### 资源配置
```go
res, err := resource.New(ctx,
    resource.WithAttributes(
        semconv.ServiceNameKey.String("otel-simple-demo"),
        semconv.ServiceVersionKey.String("v1.0.0"),
        semconv.DeploymentEnvironmentKey.String("demo"),
    ),
)
```

### 追踪配置
```go
tp := sdktrace.NewTracerProvider(
    sdktrace.WithBatcher(exporter,
        sdktrace.WithBatchTimeout(5*time.Second),
        sdktrace.WithMaxExportBatchSize(100),
    ),
    sdktrace.WithResource(res),
    sdktrace.WithSampler(sdktrace.AlwaysSample()),
)
```

## 📈 验证数据

### 1. 检查 Gateway 处理状态
```bash
docker logs proj-otel-collector --tail 20
curl http://127.0.0.1:8888/metrics | grep traces
```

### 2. 查看 Jaeger 中的追踪数据
- 打开: http://127.0.0.1:16686
- 服务名: `otel-simple-demo`
- 查找操作: `user-login`, `order-processing` 等

### 3. 检查容器状态
```bash
docker ps | grep proj-
```

## 🔍 故障排除

### 常见错误及解决方案

1. **连接被拒绝**
   ```bash
   # 检查 OTEL Gateway 状态
   curl http://127.0.0.1:13133
   docker restart proj-otel-collector
   ```

2. **URL 解析错误**
   - 确认使用正确的端点格式：`127.0.0.1:4328`（不带协议前缀）
   - 使用项目配置的端口 4328 而非标准端口 4318

3. **数据未显示**
   - 等待 1-2 分钟让数据传输和处理
   - 刷新 Jaeger UI
   - 检查 Gateway 日志是否有处理错误

### 调试选项
```bash
# 启用详细日志
export OTEL_LOG_LEVEL=debug
go run main.go

# 检查 Gateway 连接
curl http://127.0.0.1:4328/v1/traces -X POST -H "Content-Type: application/json" -d '{}'
```

## 🔗 与其他示例对比

| 特性 | Simple Demo | gRPC Demo | Complete Demo |
|------|-------------|-----------|---------------|
| 协议 | HTTP (4328) | gRPC (4327) | HTTP + gRPC |
| 复杂度 | 简单 | 中等 | 复杂 |
| 场景数量 | 5个基础场景 | 5个场景 | 多业务并发 |
| 监控功能 | 基础追踪 | 基础追踪 | 全方位监控 |

## 📚 学习路径建议

1. **开始**: 运行此 Simple Demo 理解基础概念
2. **进阶**: 尝试 [gRPC Demo](../grpc-demo/) 学习不同协议
3. **深入**: 运行 [Complete Demo](../complete/) 了解完整架构
4. **定制**: 参考 [Agent](../agent/), [Gateway](../gateway/), [SaaS](../saas/) 示例

## 📚 相关文档

- [OTEL 主文档](../README.md)
- [gRPC Demo](../grpc-demo/)
- [Complete Demo](../complete/)