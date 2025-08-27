# OTEL Agent-Collector 联合使用演示

演示 Agent 和 Collector 联合使用的架构模式，对比不同数据路径的性能和特性。

## 🎯 功能特性

- **双路径对比**: 演示直接发送 vs Agent 转发的差异
- **架构演示**: 完整的 Agent → Collector → SaaS 数据流
- **性能对比**: 展示不同路径的延迟和处理特性
- **实时监控**: 提供完整的验证和监控方式

## 🏗️ 架构说明

### 数据流对比

```
路径 1: 直接发送
┌─────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Application │────│ OTEL Collector  │────│  SaaS Platforms │
│             │    │ (port 4318)     │    │ - Jaeger        │
└─────────────┘    └─────────────────┘    │ - VictoriaLogs  │
                                          │ - Prometheus    │
                                          └─────────────────┘

路径 2: Agent 转发
┌─────────────┐    ┌─────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Application │────│ OTEL Agent  │────│ OTEL Collector  │────│  SaaS Platforms │
│             │    │(port 4328)  │    │ (port 4317)     │    │ - Jaeger        │
└─────────────┘    └─────────────┘    └─────────────────┘    │ - VictoriaLogs  │
                                                             │ - Prometheus    │
                                                             └─────────────────┘
```

## 🚀 运行示例

### 前置条件

确保 OTEL 组件正在运行：

```bash
# 启动 Agent 和 Collector
make deploy.install.docker.otel-agent
make  deploy.install.docker.otel-collector

# 验证服务状态
curl http://127.0.0.1:13133  # Collector 健康检查
curl http://127.0.0.1:13134  # Agent 健康检查
```

### 运行演示

```bash
cd examples/otel/agent-collector-demo
go run main.go
```

## 📊 演示内容

### 1. 直接到 Collector 路径

- **端点**: `127.0.0.1:4318` (Collector OTLP HTTP)
- **数据流**: 应用 → Collector → SaaS
- **特点**: 延迟更低，配置简单
- **适用场景**: 开发环境，单体应用

### 2. 通过 Agent 转发路径

- **端点**: `127.0.0.1:4328` (Agent OTLP HTTP)
- **数据流**: 应用 → Agent → Collector → SaaS
- **特点**: 容错性强，可本地预处理
- **适用场景**: 生产环境，微服务架构

### 生成的追踪数据

每个路径都会生成追踪数据，包含：

- 服务名: `agent-collector-demo`
- 操作名: `demo-operation-{路径类型}`
- 属性标记: 路径类型、时间戳等

## 🔍 验证方式

### 1. 实时健康检查

```bash
# Agent 健康状态
curl http://127.0.0.1:13134

# Collector 健康状态
curl http://127.0.0.1:13133
```

### 2. 查看 Jaeger 追踪数据

- 打开: <http://127.0.0.1:16686>
- 服务名: `agent-collector-demo`
- 查找操作: `demo-operation-直接到Collector`, `demo-operation-通过Agent转发`

### 3. 监控处理指标

```bash
# Collector 处理指标
curl http://127.0.0.1:8888/metrics | grep -E "(receiver|exporter)"

# 查看处理日志
docker logs proj-otel-collector --tail 20
docker logs proj-otel-agent --tail 20
```

### 4. 容器状态检查

```bash
# 查看所有 OTEL 相关容器
docker ps | grep proj-otel

# 查看网络连接
docker network inspect proj
```

## 📈 性能对比分析

### 延迟对比

| 路径类型 | 网络跳数 | 预期延迟 | 处理开销 |
|---------|---------|----------|----------|
| **直接发送** | 1 跳 | 低 (~50ms) | 低 |
| **Agent 转发** | 2 跳 | 中 (~100ms) | 中 |

### 资源消耗对比

| 组件 | 内存使用 | CPU使用 | 网络带宽 |
|------|----------|---------|----------|
| **仅 Collector** | ~512MB | ~500m | 中等 |
| **Agent + Collector** | ~640MB | ~600m | 优化 |

### 可靠性对比

| 特性 | 直接发送 | Agent 转发 |
|------|----------|------------|
| **单点故障** | Collector 故障影响大 | Agent 本地缓存，影响小 |
| **网络容错** | 一般 | 优秀（本地重试） |
| **数据丢失** | 风险较高 | 风险较低 |
| **配置复杂度** | 简单 | 中等 |

## 🔧 配置差异说明

### Agent 配置特点

```yaml
# 轻量级转发配置
processors:
  batch:
    timeout: 200ms          # 快速转发
    send_batch_size: 100    # 小批次

  memory_limiter:
    limit_mib: 128         # 小内存限制

exporters:
  otlp/collector:          # 单一目标转发
    endpoint: "proj-otel-collector:4317"
```

### Collector 配置特点

```yaml
# 重型处理配置
receivers:
  otlp: [...]             # 接收多种数据源
  filelog: [...]
  prometheus: [...]

processors:
  batch:
    timeout: 1000ms        # 批量优化
    send_batch_size: 1000  # 大批次处理

  probabilistic_sampler:   # 高级处理
    sampling_percentage: 10

exporters:
  otlphttp/jaeger: [...]   # 多目标导出
  otlphttp/victorialogs: [...]
  prometheus: [...]
```

## 🎭 使用场景建议

### 选择直接发送的场景

- ✅ **开发/测试环境**: 快速验证功能
- ✅ **单体应用**: 应用实例数量少
- ✅ **简单部署**: 不需要复杂的数据处理
- ✅ **资源受限**: 内存和CPU资源紧张

### 选择 Agent 转发的场景

- ✅ **生产环境**: 高可靠性要求
- ✅ **微服务架构**: 多应用实例统一管理
- ✅ **网络复杂**: 需要本地缓存和容错
- ✅ **数据预处理**: 需要过滤、脱敏等处理
- ✅ **多云部署**: 跨区域数据收集

## 🔗 相关示例

- **Simple Demo**: [../simple-demo/](../simple-demo/) - 基础 HTTP 追踪示例
- **gRPC Demo**: [../grpc-demo/](../grpc-demo/) - gRPC 协议追踪示例
- **Complete Demo**: [../complete/](../complete/) - 完整架构演示
- **Agent 专项**: [../agent/](../agent/) - Agent 层深入示例
- **Gateway 专项**: [../gateway/](../gateway/) - Gateway 层管理示例

## 🔍 故障排除

### 常见问题

1. **连接被拒绝**

   ```bash
   # 检查组件状态
   docker ps | grep proj-otel
   ./scripts/installation/otel-agent.sh status
   ./scripts/installation/otel-collector.sh status
   ```

2. **数据未出现在 Jaeger**
   - 等待 1-2 分钟让数据传输完成
   - 检查 Collector 导出配置
   - 验证 Jaeger 容器运行状态

3. **Agent 转发失败**

   ```bash
   # 检查 Agent 到 Collector 的连接
   docker logs proj-otel-agent --tail 20
   # 验证网络连通性
   docker exec proj-otel-agent ping proj-otel-collector
   ```

### 调试选项

```bash
# 启用详细日志
export OTEL_LOG_LEVEL=debug
go run main.go

# 查看详细网络信息
docker network inspect proj | grep -A 20 -B 5 otel
```

## 📚 学习路径

1. **开始**: 运行本演示了解基础概念
2. **深入**: 分析日志和指标理解数据流
3. **对比**: 测试不同场景下的性能差异
4. **实践**: 在自己的应用中选择合适的架构
5. **优化**: 根据实际需求调整配置参数

## 📚 相关文档

- [OTEL 主文档](../README.md)
- [架构设计文档](../../../docs/otel-architecture-design.md)
- [安装指南](../../../scripts/installation/README.md)
