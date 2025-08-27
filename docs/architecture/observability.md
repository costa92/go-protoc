# 统一可观测性架构设计

基于 OpenTelemetry 构建统一的可观测性平台，实现 **Agent → Gateway → SaaS** 架构模式，为微服务应用提供完整的遥测数据收集、处理和分析能力。

## 🎯 架构愿景

### 核心价值

- **统一可观测性**: 集成 Traces、Metrics、Logs 三大支柱
- **标准化协议**: 基于 OpenTelemetry 行业标准
- **灵活部署**: 支持云原生和传统部署方式
- **多平台集成**: 兼容主流 SaaS 和开源解决方案
- **高性能**: 优化的数据处理管道和资源使用

## 🏗️ 系统架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                    Agent → Gateway → SaaS 架构                    │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     Agent       │    │    Gateway      │    │     SaaS        │
│   (应用端)       │────│ (OTEL Collector)│────│   (存储分析)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 详细数据流架构

```
                      ┌─── Agent Layer ───┐
                      │                   │
┌─────────────────┐   │ ┌───────────────┐ │   ┌─────────────────┐
│  Application    │   │ │ OTEL SDKs     │ │   │    Gateway      │
│                 │───┼─│ - Traces      │─┼───│ OTEL Collector  │
│  - API Server   │   │ │ - Metrics     │ │   │                 │
│  - Microservices│   │ │ - Logs        │ │   │ ┌─────────────┐ │
│  - Background   │   │ └───────────────┘ │   │ │ Receivers   │ │
│    Jobs         │   │                   │   │ │ - OTLP      │ │
└─────────────────┘   │ ┌───────────────┐ │   │ │ - Filelog   │ │
                      │ │ File Outputs  │─┼───│ │ - Prometheus│ │
                      │ │ - JSON Logs   │ │   │ └─────────────┘ │
                      │ │ - Structured  │ │   │                 │
                      │ └───────────────┘ │   │ ┌─────────────┐ │
                      └───────────────────┘   │ │ Processors  │ │
                                              │ │ - Batch     │ │
                      ┌─── Monitoring ───┐   │ │ - Resource  │ │
                      │                  │   │ │ - Sampling  │ │
                      │ ┌──────────────┐ │   │ └─────────────┘ │
                      │ │ Prometheus   │─┼───│                 │
                      │ │ Metrics      │ │   │ ┌─────────────┐ │
                      │ └──────────────┘ │   │ │ Exporters   │ │
                      └──────────────────┘   │ │ - Victoria  │ │
                                              │ │ - Jaeger    │ │
                                              │ │ - Prometheus│ │
                                              │ │ - SaaS APIs │ │
                                              │ └─────────────┘ │
                                              └─────────────────┘
                                                      │
                              ┌─────────────────────────────────┘
                              │
                      ┌─── SaaS Layer ───┐
                      │                  │
            ┌─────────┼──────────────────┼─────────┐
            │         │                  │         │
    ┌───────▼───┐ ┌───▼────┐ ┌──────▼───┐ ┌───▼────┐
    │VictoriaLogs│ │ Jaeger │ │Prometheus│ │ Cloud  │
    │            │ │        │ │          │ │ SaaS   │
    │ - Logs     │ │- Traces│ │ -Metrics │ │-Datadog│
    │ - Search   │ │- APM   │ │ -Alerts  │ │-NewRelic│
    └────────────┘ └────────┘ └──────────┘ └────────┘
```

## 🔧 组件设计

### Agent 层设计

#### 功能职责

- **数据生成**: 应用程序集成 OpenTelemetry SDKs 生成遥测数据
- **本地处理**: 基础的数据预处理和缓冲
- **协议发送**: 通过 OTLP 协议向 Gateway 发送数据
- **文件输出**: 关键日志写入本地文件作为备份

#### 技术实现

```go
// Agent 示例架构
type OTELAgent struct {
    tracerProvider *sdktrace.TracerProvider
    meterProvider  *sdkmetric.MeterProvider
    logger         *slog.Logger

    // 配置
    serviceName    string
    serviceVersion string
    environment    string
}
```

#### 部署策略

- **Sidecar 模式**: 每个应用实例配置独立的 Agent
- **SDK 集成**: 直接在应用代码中集成 OTEL SDKs
- **文件备份**: 重要日志同时写入本地文件

### Gateway 层设计

#### 功能职责

- **数据接收**: 统一接收来自多个 Agent 的遥测数据
- **数据处理**: 批处理、资源标记、采样、格式转换
- **路由分发**: 根据配置将数据路由到不同的后端存储
- **监控告警**: 自身健康监控和数据流质量监控

#### 核心组件

##### Receivers (接收器)

```yaml
receivers:
  otlp:                    # 接收 Agent 发送的 OTLP 数据
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

  filelog:                 # 监控文件日志变化
    include: ["/logs/**/*.log"]

  prometheus:              # 抓取 Prometheus 指标
    config:
      scrape_configs: [...]
```

##### Processors (处理器)

```yaml
processors:
  batch:                   # 批处理优化传输效率
    timeout: 1s
    send_batch_size: 1000

  resource:                # 添加资源标记
    attributes:
      - key: environment
        value: "${ENV}"

  memory_limiter:          # 内存保护
    limit_mib: 512

  probabilistic_sampler:   # Traces 采样
    sampling_percentage: 10
```

##### Exporters (导出器)

```yaml
exporters:
  otlphttp/victorialogs:   # 日志存储
    endpoint: "http://victorialogs:9428"

  otlphttp/jaeger:         # 链路追踪
    endpoint: "http://jaeger:4318"

  prometheus:              # 指标监控
    endpoint: "0.0.0.0:8889"

  otlphttp/datadog:        # SaaS 集成
    endpoint: "https://otlp.datadoghq.com"
```

#### 数据管道设计

```yaml
service:
  pipelines:
    logs:
      receivers: [filelog, otlp]
      processors: [memory_limiter, resource, batch]
      exporters: [otlphttp/victorialogs]

    metrics:
      receivers: [prometheus, otlp]
      processors: [memory_limiter, resource, batch]
      exporters: [prometheus]

    traces:
      receivers: [otlp]
      processors: [memory_limiter, resource, probabilistic_sampler, batch]
      exporters: [otlphttp/jaeger]
```

### SaaS 层设计

#### 开源解决方案集成

| 组件 | 用途 | 接口协议 | 端口 |
|------|------|----------|------|
| VictoriaLogs | 日志存储和搜索 | OTLP HTTP | 9428 |
| Jaeger | 分布式链路追踪 | OTLP HTTP | 4318 |
| Prometheus | 指标监控和告警 | Prometheus | 8889 |

#### 商业 SaaS 平台集成

| 平台 | 协议 | 认证方式 | 配置 |
|------|------|----------|------|
| Datadog | OTLP HTTPS | API Key | `DD_API_KEY` |
| New Relic | OTLP HTTPS | API Key | `NEW_RELIC_API_KEY` |
| Grafana Cloud | OTLP HTTPS | API Token | `GRAFANA_API_TOKEN` |

## 🚀 部署架构

### 部署模式对比

| 特性 | Docker 部署 | 原生部署 |
|------|-------------|----------|
| **部署复杂度** | 低 | 中等 |
| **资源隔离** | 好 | 一般 |
| **性能开销** | 略高 | 低 |
| **扩展性** | 优秀 | 良好 |
| **维护成本** | 低 | 中等 |
| **生产就绪** | ✅ 推荐 | ✅ 支持 |

### Docker 部署架构

```yaml
# Docker Compose 示例
version: '3.8'
services:
  otel-collector:
    image: otel/opentelemetry-collector-contrib:0.132.0
    container_name: proj-otel-collector
    networks:
      - proj
    ports:
      - "4317:4317"   # OTLP gRPC
      - "4318:4318"   # OTLP HTTP
      - "13133:13133" # Health
      - "8888:8888"   # Metrics
    volumes:
      - ./config:/etc/otelcol-contrib/
      - ./logs:/host/logs:ro
    environment:
      - PROJ_ENVIRONMENT=production
      - PROJ_SERVICE_NAMESPACE=default
```

### Kubernetes 部署架构

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: otel-collector-gateway
spec:
  replicas: 3
  selector:
    matchLabels:
      app: otel-collector-gateway
  template:
    spec:
      containers:
      - name: otel-collector
        image: otel/opentelemetry-collector-contrib:0.132.0
        ports:
        - containerPort: 4317
        - containerPort: 4318
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

## ⚡ 性能设计

### 性能目标

| 指标 | 目标值 | 监控方式 |
|------|--------|----------|
| **数据延迟** | < 1s (P95) | 端到端链路追踪 |
| **吞吐量** | > 10K events/s | Prometheus 指标 |
| **内存使用** | < 512MB | 资源监控 |
| **CPU 使用** | < 2 cores | 资源监控 |
| **数据丢失** | < 0.01% | 数据质量监控 |

### 性能优化策略

#### 批处理优化

```yaml
processors:
  batch:
    # 生产环境配置
    timeout: 1s                    # 适中延迟
    send_batch_size: 1000         # 较大批次
    send_batch_max_size: 1500     # 上限保护
```

#### 内存管理

```yaml
processors:
  memory_limiter:
    check_interval: 1s
    limit_mib: 512               # 根据环境调整
```

#### 采样策略

```yaml
processors:
  probabilistic_sampler:
    sampling_percentage: 1       # 生产环境低采样
```

## 🔒 安全设计

### 网络安全

#### 传输加密

- **TLS 支持**: OTLP HTTPS 传输加密
- **证书管理**: 自动证书轮换机制
- **网络隔离**: 容器网络和防火墙配置

#### 认证授权

```yaml
# mTLS 配置示例
extensions:
  tls:
    cert_file: /etc/certs/server.crt
    key_file: /etc/certs/server.key
    ca_file: /etc/certs/ca.crt
```

### 数据安全

#### 敏感数据处理

- **数据脱敏**: PII 数据自动脱敏
- **字段过滤**: 敏感字段过滤和替换
- **访问控制**: 基于角色的数据访问控制

#### 密钥管理

- **环境变量**: API Keys 通过环境变量注入
- **密钥轮换**: 支持动态密钥更新
- **加密存储**: 本地密钥加密存储

## 🛠️ 运维设计

### 部署自动化

#### 脚本化部署

```bash
# 统一部署接口
./scripts/installation/otel-agent.sh docker.install
./scripts/installation/otel-collector.sh docker.install
```

### 监控告警

#### 系统监控

| 监控项 | 指标 | 告警阈值 |
|--------|------|----------|
| 服务可用性 | HTTP 200 | < 99.9% |
| 数据处理延迟 | P95 延迟 | > 2s |
| 错误率 | 失败请求比例 | > 1% |
| 资源使用 | 内存/CPU | > 80% |

#### 业务监控

- **数据量监控**: 各数据类型的吞吐量
- **质量监控**: 数据完整性和正确性
- **SLA 监控**: 端到端服务质量

## 🔮 未来演进

### 短期目标 (3-6个月)

- **智能采样**: 基于业务规则的动态采样
- **数据压缩**: 高效的数据压缩算法
- **多租户**: 支持多租户数据隔离

### 中期目标 (6-12个月)

- **机器学习**: 异常检测和智能告警
- **边缘计算**: 边缘节点的 OTEL 支持
- **实时分析**: 流式数据处理和实时分析

### 长期目标 (12个月+)

- **AI 集成**: 基于 AI 的可观测性洞察
- **自动化运维**: 全自动化的运维体系
- **生态整合**: 与更多开源和商业平台集成

## 📚 相关文档

- [OTEL 部署指南](../guides/deployment/otel-installation.md)
- [示例代码](../../examples/otel/)
- [故障排除](../guides/troubleshooting/otel-fixes.md)
