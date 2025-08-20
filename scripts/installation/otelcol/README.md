# OpenTelemetry Collector 配置文档

## 概述

本目录包含 OpenTelemetry Collector 的配置文件和相关文档，用于在不同环境中部署和运行日志、指标、链路追踪数据收集服务。

## 配置文件说明

### 1. `config.yaml` - Linux 原生安装配置

**用途**：专为在 Linux 宿主机上原生运行的 OpenTelemetry Collector 设计

**特点**：

- 支持 systemd 服务管理
- 原生文件系统访问
- 完整的扩展功能（pprof、zpages）
- 适合生产环境部署

**部署方式**：

```bash
# 使用 make 命令安装
make deploy.install.otelcol

# 或直接使用脚本
./scripts/installation/otelcol.sh proj::otelcol::install
```

**配置特点**：

- 监控路径：`/var/log/app/**/*.log`
- 健康检查：`localhost:${PROJ_OTELCOL_HEALTH_PORT}`
- 性能分析：`localhost:1777` (pprof)
- 调试界面：`localhost:55679` (zpages)

### 2. `config-docker.yaml` - Docker 容器化配置

**用途**：专为在 Docker 容器中运行的 OpenTelemetry Collector 设计

**特点**：

- 容器网络优化
- 卷挂载支持
- 环境变量配置
- 适合容器化部署

**部署方式**：

```bash
# 使用 make 命令安装
make deploy.install.docker.otelcol

# 或直接使用脚本
./scripts/installation/otelcol.sh proj::otelcol::docker::install
```

**配置特点**：

- 监控路径：`/host/logs/**/*.log` (通过卷挂载)
- 健康检查：`0.0.0.0:13133`
- 容器网络：通过服务名称访问其他服务
- 资源限制：内存限制 512MB

## 环境变量配置

### Linux 原生环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `PROJ_OTELCOL_HOST` | 127.0.0.1 | Collector 主机地址 |
| `PROJ_OTELCOL_GRPC_PORT` | 4327 | OTLP gRPC 接收端口 |
| `PROJ_OTELCOL_HTTP_PORT` | 4328 | OTLP HTTP 接收端口 |
| `PROJ_OTELCOL_HEALTH_PORT` | 13133 | 健康检查端口 |
| `PROJ_VICTORIALOGS_HOST` | 127.0.0.1 | VictoriaLogs 主机地址 |
| `PROJ_VICTORIALOGS_PORT` | 9428 | VictoriaLogs 端口 |
| `PROJ_SERVICE_NAME` | otelcol | 服务名称 |
| `PROJ_SERVICE_VERSION` | unknown | 服务版本 |
| `PROJ_ENVIRONMENT` | development | 运行环境 |

### Docker 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `PROJ_SERVICE_NAME` | apiserver | 服务名称（通过容器启动注入） |
| `PROJ_SERVICE_VERSION` | v2.0.0 | 服务版本（通过容器启动注入） |
| `PROJ_ENVIRONMENT` | development | 运行环境（通过容器启动注入） |

## 核心组件配置说明

### 接收器 (Receivers)

#### 1. OTLP 接收器

- **用途**：接收应用程序发送的标准 OpenTelemetry 数据
- **协议**：支持 gRPC 和 HTTP 两种协议
- **端口**：gRPC 4327，HTTP 4328
- **优化**：支持 keepalive、最大消息大小等配置

#### 2. 文件日志接收器 (filelog)

- **用途**：实时监控文件系统中的日志文件变化
- **特点**：
  - 毫秒级响应（500ms 轮询间隔）
  - JSON 自动解析
  - 时间戳提取
  - 元数据增强

#### 3. Prometheus 接收器

- **用途**：抓取 Prometheus 格式的指标数据
- **配置**：支持多个作业配置
- **监控**：包括 Collector 自身指标监控

### 处理器 (Processors)

#### 1. 批处理器 (batch)

- **用途**：将多个数据点打包发送，提高传输效率
- **配置**：
  - 超时时间：1秒
  - 批大小：1024个数据点
  - 最大批大小：2048个数据点

#### 2. 内存限制器 (memory_limiter)

- **用途**：防止内存使用过高导致系统不稳定
- **配置**：
  - Linux：512MB 限制，128MB 峰值
  - Docker：512MB 限制，128MB 峰值

#### 3. 资源处理器 (resource)

- **用途**：为所有遥测数据添加资源属性
- **属性**：服务名称、版本、环境、主机名等

### 导出器 (Exporters)

#### 1. 控制台日志导出器 (logging)

- **用途**：调试和开发，输出到标准输出
- **配置**：支持详细级别和采样控制

#### 2. Prometheus 导出器

- **用途**：提供 Prometheus 兼容的指标端点
- **端口**：8889
- **特点**：支持命名空间、固定标签等

#### 3. Jaeger 导出器 (otlp/jaeger)

- **用途**：发送链路追踪数据到 Jaeger
- **配置**：
  - Linux：`http://127.0.0.1:14268/api/traces`
  - Docker：`http://jaeger:14268/api/traces`
- **优化**：重试机制、队列配置

#### 4. VictoriaLogs 导出器 (loki/victorialogs)

- **用途**：发送日志数据到 VictoriaLogs 存储
- **配置**：
  - Linux：`http://127.0.0.1:9428/insert/loki/api/v1/push`
  - Docker：`http://proj-victorialogs:9428/insert/loki/api/v1/push`
- **特点**：支持标签配置、重试机制

#### 5. 文件导出器 (file) - 仅 Linux

- **用途**：将日志写入本地文件作为备份
- **配置**：支持文件轮转、压缩等功能

### 扩展 (Extensions)

#### 1. 健康检查 (health_check)

- **端点**：`http://0.0.0.0:13133/`
- **功能**：提供服务健康状态检查
- **特点**：支持流水线健康检查

#### 2. 性能分析 (pprof) - 仅 Linux

- **端点**：`http://localhost:1777`
- **功能**：Go pprof 性能分析
- **用途**：性能调优和问题排查

#### 3. zPages (zpages) - 仅 Linux

- **端点**：`http://localhost:55679`
- **功能**：服务状态和调试信息 Web 界面
- **页面**：
  - `/debug/tracez` - 链路追踪信息
  - `/debug/pipelinez` - 流水线状态
  - `/debug/servicez` - 服务状态

## 数据流水线 (Pipelines)

### 链路追踪流水线 (traces)

- **接收器**：otlp
- **处理器**：memory_limiter → resource → batch
- **导出器**：logging + otlp/jaeger

### 指标流水线 (metrics)

- **接收器**：otlp + prometheus
- **处理器**：memory_limiter → resource → batch
- **导出器**：logging + prometheus

### 日志流水线 (logs)

- **接收器**：otlp + filelog
- **处理器**：memory_limiter → resource → batch
- **导出器**：
  - Linux：logging + loki/victorialogs + file/backup
  - Docker：logging + loki/victorialogs

## 部署最佳实践

### Linux 原生部署

1. **系统要求**：
   - Linux 操作系统
   - systemd 支持
   - 512MB 可用内存

2. **目录结构**：

   ```
   /usr/local/bin/otelcol-contrib     # 二进制文件
   /etc/otelcol/config.yaml           # 配置文件
   /var/log/otelcol/                  # 日志目录
   /var/lib/otelcol/                  # 数据目录
   ```

3. **服务管理**：

   ```bash
   sudo systemctl start otelcol
   sudo systemctl enable otelcol
   sudo systemctl status otelcol
   ```

### Docker 容器部署

1. **容器要求**：
   - Docker 网络：proj
   - 内存限制：建议 1GB
   - CPU 限制：建议 0.5 核心

2. **卷挂载**：

   ```bash
   # 配置文件挂载（只读）
   -v /path/to/config-docker.yaml:/etc/otelcol-contrib/config.yaml:ro

   # 宿主机日志目录挂载（只读）
   -v /host/logs:/host/logs:ro

   # 数据目录挂载
   -v otelcol_data:/var/log/otelcol
   ```

3. **端口映射**：

   ```bash
   -p 127.0.0.1:4327:4327    # OTLP gRPC
   -p 127.0.0.1:4328:4328    # OTLP HTTP
   -p 127.0.0.1:8888:8888    # Metrics
   -p 127.0.0.1:13133:13133  # Health check
   ```

## 监控和维护

### 健康检查

```bash
# 检查服务健康状态
curl http://localhost:13133/

# 预期响应
{
  "status": "Server available",
  "upSince": "2025-08-20T14:00:00Z",
  "uptime": "1h30m45s"
}
```

### 指标监控

```bash
# 查看 Collector 指标
curl http://localhost:8889/metrics

# 常用指标
otelcol_receiver_accepted_spans_total
otelcol_exporter_sent_spans_total
otelcol_processor_batch_batch_send_size_sum
```

### 日志监控

```bash
# Linux systemd 日志
sudo journalctl -u otelcol -f

# Docker 容器日志
docker logs -f proj-otelcol
```

### 性能调优

1. **内存优化**：
   - 根据数据量调整 `memory_limiter` 配置
   - 监控内存使用情况

2. **批处理优化**：
   - 调整 `batch` 处理器的大小和超时
   - 平衡延迟和吞吐量

3. **网络优化**：
   - 配置适当的超时时间
   - 启用重试机制

## 故障排查

### 常见问题

1. **服务无法启动**：
   - 检查配置文件语法
   - 验证端口是否被占用
   - 查看系统日志

2. **数据丢失**：
   - 检查导出器连通性
   - 验证重试配置
   - 监控队列状态

3. **性能问题**：
   - 使用 pprof 进行性能分析
   - 检查内存使用情况
   - 优化批处理配置

### 调试工具

1. **zPages 界面** (仅 Linux)：
   - 访问 `http://localhost:55679`
   - 查看实时状态信息

2. **pprof 分析** (仅 Linux)：
   - 访问 `http://localhost:1777/debug/pprof/`
   - 生成性能分析报告

3. **详细日志**：
   - 设置日志级别为 `debug`
   - 启用详细输出模式

## 相关文件

- `config.yaml` - Linux 原生安装配置文件
- `config-docker.yaml` - Docker 容器化配置文件
- `otelcol.service` - systemd 服务文件模板
- `../otelcol.sh` - 安装和管理脚本

## 更多资源

- [OpenTelemetry Collector 官方文档](https://opentelemetry.io/docs/collector/)
- [配置参考手册](https://opentelemetry.io/docs/collector/configuration/)
- [最佳实践指南](https://opentelemetry.io/docs/collector/deployment/)
