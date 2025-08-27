# OpenTelemetry (OTEL) 安装使用指南

## 概览

OpenTelemetry (OTEL) 是一个开源的可观测性框架，提供统一的遥测数据收集、处理和导出能力。本项目采用**分离式双组件架构**，完全区分 Agent 和 Collector 的职责：

### 🔹 OTEL Agent (轻量级边车)

- **用途**: 应用程序边车容器，专注数据收集和快速转发
- **特点**: 低延迟、小内存占用、简单处理逻辑
- **部署**: 与应用程序一对一部署，就近收集遥测数据
- **端口**: gRPC 4327, HTTP 4328, 健康检查 13134

### 🔸 OTEL Collector (重量级网关)

- **用途**: 集中式数据处理网关，执行复杂的路由和导出
- **特点**: 高吞吐量、丰富处理能力、多后端支持
- **部署**: 独立部署，处理来自多个 Agent 的数据流
- **端口**: gRPC 4317, HTTP 4318, 健康检查 13133

## 功能特性

- ✅ **双部署模式**: 支持原生安装和 Docker 容器化部署
- ✅ **完整架构**: 实现 Agent → Gateway → SaaS 数据流模式
- ✅ **多数据源**: 支持文件日志、OTLP 协议、Prometheus 指标收集
- ✅ **多目标导出**: 支持 VictoriaLogs、Jaeger、Prometheus 等多种后端
- ✅ **SaaS 集成**: 预配置 Datadog、New Relic 等 SaaS 平台接口
- ✅ **健康监控**: 内置健康检查和性能监控端点
- ✅ **环境隔离**: 支持开发、测试、生产环境配置

## 快速开始

### 1. 检查系统要求

```bash
# 查看所有组件版本
./scripts/installation/versions.sh show

# 验证版本格式
./scripts/installation/versions.sh validate
```

### 2. Docker 部署 (推荐)

```bash
# 选项1: 安装完整 OTEL 栈 (Agent + Collector)
make deploy.install.docker.otel-stack

# 选项2: 仅安装 OTEL Collector (网关)
make deploy.install.docker.otel-collector

# 选项3: 仅安装 OTEL Agent (边车)
make deploy.install.docker.otel-agent

# 检查 Collector 状态
curl http://127.0.0.1:13133

# 检查 Agent 状态
curl http://127.0.0.1:13134
```

### 3. 原生部署

```bash
# 选项1: 安装完整 OTEL 栈
make deploy.install.otel-stack

# 选项2: 仅安装 OTEL Collector
make deploy.install.otel-collector

# 选项3: 仅安装 OTEL Agent
make deploy.install.otel-agent

# 检查服务状态
systemctl status otel-collector
systemctl status otel-agent
```

## 安装命令

### Make 命令

#### 完整栈部署

| 命令 | 描述 |
|------|------|
| `make deploy.install.otel-stack` | 原生安装完整 OTEL 栈 (Agent + Collector) |
| `make deploy.uninstall.otel-stack` | 卸载完整 OTEL 栈 |
| `make deploy.install.docker.otel-stack` | Docker 安装完整 OTEL 栈 |
| `make deploy.uninstall.docker.otel-stack` | Docker 卸载完整 OTEL 栈 |

#### OTEL Collector (网关)

| 命令 | 描述 |
|------|------|
| `make deploy.install.otel-collector` | 原生安装 OTEL Collector |
| `make deploy.uninstall.otel-collector` | 卸载 OTEL Collector |
| `make deploy.install.docker.otel-collector` | Docker 安装 OTEL Collector |
| `make deploy.uninstall.docker.otel-collector` | Docker 卸载 OTEL Collector |

#### OTEL Agent (边车)

| 命令 | 描述 |
|------|------|
| `make deploy.install.otel-agent` | 原生安装 OTEL Agent |
| `make deploy.uninstall.otel-agent` | 卸载 OTEL Agent |
| `make deploy.install.docker.otel-agent` | Docker 安装 OTEL Agent |
| `make deploy.uninstall.docker.otel-agent` | Docker 卸载 OTEL Agent |

### 脚本命令

#### OTEL Collector 脚本

```bash
./scripts/installation/otel-collector.sh install              # 原生安装
./scripts/installation/otel-collector.sh uninstall           # 原生卸载
./scripts/installation/otel-collector.sh docker.install      # Docker 安装
./scripts/installation/otel-collector.sh docker.uninstall    # Docker 卸载
./scripts/installation/otel-collector.sh status              # 检查状态
./scripts/installation/otel-collector.sh info                # 显示信息
```

#### OTEL Agent 脚本

```bash
./scripts/installation/otel-agent.sh install                 # 原生安装
./scripts/installation/otel-agent.sh uninstall              # 原生卸载
./scripts/installation/otel-agent.sh docker.install         # Docker 安装
./scripts/installation/otel-agent.sh docker.uninstall       # Docker 卸载
./scripts/installation/otel-agent.sh status                 # 检查状态
./scripts/installation/otel-agent.sh info                   # 显示信息
```

## 配置说明

### 环境变量配置

#### OTEL Collector 配置

```bash
# OTEL Collector 基础配置
export PROJ_OTEL_HOST=127.0.0.1                    # 服务主机
export PROJ_OTEL_GRPC_PORT=4317                     # OTLP gRPC 端口
export PROJ_OTEL_HTTP_PORT=4318                     # OTLP HTTP 端口
export PROJ_OTEL_HEALTH_PORT=13133                  # 健康检查端口
export PROJ_OTEL_METRICS_PORT=8888                  # 指标端口
export PROJ_OTEL_PROMETHEUS_PORT=8889               # Prometheus 导出端口
```

#### OTEL Agent 配置

```bash
# OTEL Agent 基础配置
export PROJ_OTEL_AGENT_HOST=127.0.0.1               # Agent 服务主机
export PROJ_OTEL_AGENT_GRPC_PORT=4327                # Agent OTLP gRPC 端口
export PROJ_OTEL_AGENT_HTTP_PORT=4328                # Agent OTLP HTTP 端口
export PROJ_OTEL_AGENT_HEALTH_PORT=13134             # Agent 健康检查端口
export OTEL_COLLECTOR_ENDPOINT=127.0.0.1:4317        # Collector 端点
```

#### 业务环境配置

```bash
export PROJ_ENVIRONMENT=development                 # 环境标识
export PROJ_SERVICE_NAMESPACE=default               # 服务命名空间
export PROJ_SERVICE_NAME=apiserver                  # 服务名称
```

### 配置文件

#### OTEL Collector 配置文件

| 配置文件 | 用途 | 描述 |
|----------|------|------|
| `scripts/installation/otel-collector/config.yaml` | 原生部署 | Collector 原生部署配置 |
| `scripts/installation/otel-collector/config-docker.yaml` | Docker 部署 | Collector 容器化部署配置 |

#### OTEL Agent 配置文件

| 配置文件 | 用途 | 描述 |
|----------|------|------|
| `scripts/installation/otel-agent/config.yaml` | 原生部署 | Agent 原生部署配置 |
| `scripts/installation/otel-agent/config-docker.yaml` | Docker 部署 | Agent 容器化部署配置 |

## 服务端点

### 核心服务端点

#### OTEL Collector 端点

| 端点 | 端口 | 协议 | 描述 |
|------|------|------|------|
| OTLP gRPC | 4317 | gRPC | 接收来自 Agent 的数据 |
| OTLP HTTP | 4318 | HTTP | 接收来自 Agent 的数据 |
| 健康检查 | 13133 | HTTP | Collector 健康状态监控 |
| 指标导出 | 8888 | HTTP | Collector 内部 Prometheus 指标 |
| Prometheus 导出 | 8889 | HTTP | 处理后的业务指标导出 |
| 性能分析 | 1777 | HTTP | pprof 性能分析 |

#### OTEL Agent 端点

| 端点 | 端口 | 协议 | 描述 |
|------|------|------|------|
| OTLP gRPC | 4327 | gRPC | 接收应用程序数据 |
| OTLP HTTP | 4328 | HTTP | 接收应用程序数据 |
| 健康检查 | 13134 | HTTP | Agent 健康状态监控 |

### 健康检查

```bash
# 检查 OTEL Collector 健康状态
curl http://127.0.0.1:13133

# 检查 OTEL Agent 健康状态
curl http://127.0.0.1:13134

# 查看 Collector Prometheus 指标
curl http://127.0.0.1:8888/metrics

# 查看处理后的业务指标
curl http://127.0.0.1:8889/metrics
```

## 数据流架构

### Agent → Gateway → SaaS 模式

```
Application → OTEL Agent → OTEL Collector → Backend/SaaS Platforms
     ↓            ↓              ↓                    ↓
  生成遥测数据   边车收集       网关处理           存储和分析
   - Traces    - 快速转发      - 批处理           - VictoriaLogs
   - Metrics   - 资源标记      - 复杂路由          - Jaeger
   - Logs      - 轻量处理      - 采样过滤          - Prometheus
               - 文件监控      - 格式转换          - Datadog/New Relic

端口映射:
App → Agent:4327/4328 → Collector:4317/4318 → Backends
```

### 数据接收器 (Receivers)

| 接收器 | 数据类型 | 描述 |
|--------|----------|------|
| `filelog` | Logs | 监控日志文件变化 |
| `otlp` | Traces/Metrics/Logs | 接收 OTLP 协议数据 |
| `prometheus` | Metrics | 抓取 Prometheus 指标 |

### 数据处理器 (Processors)

| 处理器 | 功能 | 描述 |
|--------|------|------|
| `batch` | 批处理 | 提高传输效率 |
| `resource` | 资源标记 | 添加环境和服务标识 |
| `memory_limiter` | 内存限制 | 防止内存溢出 |
| `probabilistic_sampler` | 采样 | 减少 traces 数据量 |

### 数据导出器 (Exporters)

| 导出器 | 目标 | 数据类型 | 描述 |
|--------|------|----------|------|
| `otlphttp/victorialogs` | VictoriaLogs | Logs | 日志存储和查询 |
| `otlphttp/jaeger` | Jaeger | Traces | 分布式链路追踪 |
| `prometheus` | Prometheus | Metrics | 指标监控 |
| `file` | 本地文件 | All | 本地备份和调试 |

## SaaS 平台集成

### 支持的 SaaS 平台

| 平台 | 环境变量 | 端点 |
|------|----------|------|
| Datadog | `DD_API_KEY` | `https://otlp.datadoghq.com` |
| New Relic | `NEW_RELIC_API_KEY` | `https://otlp.nr-data.net:4318` |

### 激活 SaaS 导出

1. 设置对应的 API Key 环境变量
2. 取消配置文件中相关导出器的注释
3. 重启 OTEL Collector 服务

```bash
# 设置 Datadog API Key
export DD_API_KEY="your-datadog-api-key"

# 设置 New Relic API Key
export NEW_RELIC_API_KEY="your-newrelic-api-key"
```

## 故障排查

### 常见问题

#### 1. 容器启动失败

```bash
# 检查 Collector 容器日志
docker logs proj-otel-collector

# 检查 Agent 容器日志
docker logs proj-otel-agent

# 检查端口冲突
netstat -tlnp | grep -E ':(4317|4318|4327|4328|13133|13134|8888)'

# 清理并重新启动 Collector
docker rm -f proj-otel-collector
make deploy.install.docker.otel-collector

# 清理并重新启动 Agent
docker rm -f proj-otel-agent
make deploy.install.docker.otel-agent

# 清理并重新启动完整栈
docker rm -f proj-otel-collector proj-otel-agent
make deploy.install.docker.otel-stack
```

#### 2. 健康检查失败

```bash
# 检查 Collector 健康端点
curl -v http://127.0.0.1:13133

# 检查 Agent 健康端点
curl -v http://127.0.0.1:13134

# 检查 Collector 配置文件语法
docker exec proj-otel-collector otelcol-contrib --config=/etc/otelcol-contrib/config.yaml --dry-run

# 检查 Agent 配置文件语法
docker exec proj-otel-agent otelcol-contrib --config=/etc/otelcol-contrib/config.yaml --dry-run
```

#### 3. Agent 与 Collector 连接问题

```bash
# 检查 Agent 到 Collector 的连接
docker exec proj-otel-agent nc -zv 127.0.0.1 4317

# 检查 Agent 配置中的 Collector 端点
docker exec proj-otel-agent grep -A 5 "otlp/collector" /etc/otelcol-contrib/config.yaml

# 验证 Agent 日志中的转发状态
docker logs proj-otel-agent | grep -i "export"
```

#### 4. 日志文件监控不工作

```bash
# 检查日志目录权限
ls -la /path/to/logs/

# 检查 Agent 文件路径配置 (Agent 负责文件监控)
grep -A 10 "filelog:" scripts/installation/otel-agent/config.yaml

# 手动测试日志写入
echo '{"level":"INFO","msg":"test"}' >> logs/complete-demo/complete.log
```

#### 4. 网络连接问题

```bash
# Docker 网络检查
docker network inspect proj

# 检查服务连通性
docker exec proj-otel-collector curl -s http://proj-victorialogs:9428/health
```

### 调试模式

```bash
# 启用 Collector 调试日志
# 在 scripts/installation/otel-collector/config-docker.yaml 中修改:
# telemetry:
#   logs:
#     level: debug

# 启用 Agent 调试日志
# 在 scripts/installation/otel-agent/config-docker.yaml 中修改:
# telemetry:
#   logs:
#     level: debug

# 查看 Collector 详细日志
docker logs -f proj-otel-collector

# 查看 Agent 详细日志
docker logs -f proj-otel-agent

# 同时查看两个组件的日志
docker logs -f proj-otel-collector & docker logs -f proj-otel-agent
```

## 性能调优

### 生产环境配置建议

```yaml
processors:
  batch:
    timeout: 1s                    # 生产环境适中的超时时间
    send_batch_size: 1000         # 较大的批处理大小
    send_batch_max_size: 1500

  memory_limiter:
    limit_mib: 512                # 根据可用内存调整

  probabilistic_sampler:
    sampling_percentage: 1        # 生产环境较低采样率
```

### 监控指标

| 指标 | 描述 | 告警阈值 |
|------|------|----------|
| `otelcol_receiver_refused_spans_total` | 被拒绝的 spans | > 0 |
| `otelcol_exporter_send_failed_spans_total` | 发送失败的 spans | > 100/min |
| `otelcol_processor_batch_timeout_trigger_send_total` | 批处理超时 | > 1000/min |

## 集成示例

### 应用程序集成

查看完整的示例代码：

- **Agent 示例**: `examples/otel/agent/main.go`
- **SaaS 集成示例**: `examples/otel/saas/main.go`
- **完整架构示例**: `examples/otel/complete/main.go`

### 运行示例

```bash
# 启动完整 OTEL 栈 (推荐)
make deploy.install.docker.otel-stack

# 或者分步启动
make deploy.install.docker.otel-collector  # 先启动 Collector
make deploy.install.docker.otel-agent      # 再启动 Agent

# 启动 VictoriaLogs (日志后端)
./scripts/installation/victoria.sh victorialogs.docker.install

# 运行完整示例 (向 Agent 发送数据)
cd examples/otel/complete
go run main.go

# 查询收集的日志
curl -s "http://127.0.0.1:9428/select/logsql/query" -d 'query=*'

# 验证数据流: App → Agent → Collector → VictoriaLogs
curl http://127.0.0.1:13134  # Agent 健康检查
curl http://127.0.0.1:13133  # Collector 健康检查
```

## 相关链接

- **OpenTelemetry 官方文档**: [https://opentelemetry.io/docs/](https://opentelemetry.io/docs/)
- **OTEL Collector 文档**: [https://opentelemetry.io/docs/collector/](https://opentelemetry.io/docs/collector/)
- **配置参考**: [https://github.com/open-telemetry/opentelemetry-collector/tree/main/receiver](https://github.com/open-telemetry/opentelemetry-collector/tree/main/receiver)
- **项目设计文档**: `docs/otel-architecture-design.md`
