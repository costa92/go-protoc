# OTEL Gateway Example

演示 OpenTelemetry Gateway (Collector) 的管理和交互，展示集中式数据收集和处理功能。

## 🎯 功能特性

- **健康状态监控**: Gateway 服务健康检查
- **指标收集分析**: 处理统计和性能监控
- **配置管理**: 动态配置查询和管理
- **连接测试**: Gateway 连通性验证

## 🚀 运行示例

```bash
cd examples/otel/gateway
go run main.go
```

## 📊 示例内容

### 1. ExampleGatewayUsage()
基础 Gateway 交互演示：
- 创建 Gateway 客户端
- 健康状态检查
- 获取运行指标
- 配置信息查询

### 2. GatewayClient 功能
完整的 Gateway 管理功能：
- `CheckHealth()` - 健康检查
- `GetMetrics()` - 获取处理指标
- `GetConfig()` - 配置信息查询
- `TestConnection()` - 连接测试

## 🔧 Gateway 信息

### 端点配置
- **健康检查**: `http://127.0.0.1:13133`
- **指标端点**: `http://127.0.0.1:8888/metrics`
- **OTLP gRPC**: `127.0.0.1:4327`
- **OTLP HTTP**: `127.0.0.1:4328`

### 监控指标
```bash
# 查看 Gateway 处理指标
curl http://127.0.0.1:8888/metrics | grep otelcol

# 检查接收器状态
curl http://127.0.0.1:8888/metrics | grep receiver

# 查看导出器状态  
curl http://127.0.0.1:8888/metrics | grep exporter
```

## 📈 数据流监控

### 数据处理管道
```
接收器 (Receivers) → 处理器 (Processors) → 导出器 (Exporters)
    ↓                    ↓                     ↓
 OTLP, FileLog        Batch, Resource      Jaeger, VictoriaLogs
```

### 实时监控
```bash
# 实时查看处理统计
watch -n 5 'curl -s http://127.0.0.1:8888/metrics | grep -E "(processed|exported)"'

# 监控容器资源使用
docker stats proj-otel-collector
```

## 🔍 故障诊断

### 常见问题检查

1. **Gateway 无响应**
   ```bash
   docker restart proj-otel-collector
   curl http://127.0.0.1:13133
   ```

2. **数据处理异常**
   ```bash
   docker logs proj-otel-collector --tail 50
   ```

3. **端口连接问题**
   ```bash
   docker port proj-otel-collector
   netstat -tlnp | grep 432[78]
   ```

## 📊 配置管理

Gateway 配置位置：`/etc/otelcol/config.yaml`

主要配置段：
- `receivers` - 数据接收配置
- `processors` - 数据处理配置  
- `exporters` - 数据导出配置
- `service` - 管道和扩展配置

## 📚 相关文档

- [OTEL 主文档](../README.md)
- [Agent 示例](../agent/)
- [Complete 示例](../complete/)