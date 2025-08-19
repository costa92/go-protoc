# OpenTelemetry Collector Configuration Templates

本目录包含 OpenTelemetry Collector 的不同安装环境配置模板。

## 配置文件说明

### 1. `config.yaml` - 宿主机安装配置

- **适用场景**: 直接在宿主机上安装 OpenTelemetry Collector
- **网络特点**: 使用 localhost 和主机网络
- **服务发现**: 通过主机端口连接 Jaeger 等服务
- **日志输出**: 文件输出到 `/var/log/otelcol/`

### 2. `config-docker.yaml` - Docker 网络化配置

- **适用场景**: Docker 容器安装，与其他服务容器在同一网络
- **网络特点**: 使用 Docker 容器网络和服务名
- **服务发现**: 通过容器名 (如 `jaeger`) 连接服务
- **日志输出**: 混合输出（日志到文件，追踪到 Jaeger）

### 3. `config-docker-standalone.yaml` - Docker 独立配置

- **适用场景**: Docker 容器安装，但没有其他依赖服务
- **网络特点**: 仅内部网络，不依赖外部服务
- **服务发现**: 无外部服务依赖
- **日志输出**: 全部输出到文件和日志

### 4. `otelcol.service` - SystemD 服务文件

- **适用场景**: 宿主机安装的 SystemD 服务配置
- **功能**: 自动启动、重启策略、权限管理

## 自动选择逻辑

脚本会根据安装方式和环境自动选择合适的配置：

- **宿主机安装** (`proj::otelcol::install`): 使用 `config.yaml`
- **Docker 安装** (`proj::otelcol::docker::install`):
  - 检测到 Jaeger 容器 → 使用 `config-docker.yaml`
  - 未检测到 Jaeger 容器 → 使用 `config-docker-standalone.yaml`

## 环境变量支持

所有配置文件都支持以下环境变量替换：

- `${PROJ_OTELCOL_HOST}` - OpenTelemetry Collector 主机地址
- `${PROJ_OTELCOL_HTTP_PORT}` - OTLP HTTP 接收器端口 (默认: 4318)
- `${PROJ_OTELCOL_GRPC_PORT}` - OTLP gRPC 接收器端口 (默认: 4317)
- `${PROJ_OTELCOL_HEALTH_PORT}` - 健康检查端口 (默认: 13133)
- `${PROJ_OTELCOL_METRICS_PORT}` - 指标端口 (默认: 8888)
- `${PROJ_OTELCOL_CONFIG_DIR}` - 配置目录 (宿主机安装)

## 配置特性对比

| 特性 | 宿主机版 | Docker网络版 | Docker独立版 |
|------|----------|--------------|--------------|
| Jaeger 追踪 | ✅ 主机端口 | ✅ 容器网络 | ❌ 仅文件输出 |
| Prometheus 指标 | ✅ | ✅ | ✅ |
| 文件日志 | ✅ | ✅ | ✅ |
| 控制台日志 | ✅ | ✅ | ✅ |
| 外部服务依赖 | 是 | 是 | 否 |
| 网络隔离 | 否 | 否 | 是 |

## 自定义配置

如需自定义配置，可以：

1. 复制对应的模板文件
2. 修改导出器、处理器等配置
3. 更新脚本中的模板文件路径

注意：修改后需要重新安装服务以使配置生效。
