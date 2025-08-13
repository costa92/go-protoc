# go-protoc

基于 Kratos v2 框架的生产级 Go 微服务项目，采用 Protocol Buffers 作为单一数据源，支持完整的基础设施栈和可观测性。

## 快速开始

### 安装开发工具

```bash
# 安装所有开发工具
make install-tools A=1

# 安装特定工具
make tools.install.*
```

### 启动开发环境

```bash
# 完整开发环境设置（工具+服务+代码生成）
make dev-setup

# 启动 API 服务器
make run-api

# 构建项目
make build
```

## 基础设施服务

项目集成了完整的微服务基础设施栈，支持一键安装和管理：

### 数据存储层

- **Redis** - 高性能缓存服务
- **MariaDB** - MySQL 兼容的关系型数据库
- **MongoDB** - NoSQL 文档数据库
- **etcd** - 分布式键值存储，服务发现

### 消息传输层

- **Kafka** - 分布式消息队列，事件流处理

### 可观测性栈

- **Jaeger** - 分布式链路追踪
- **Prometheus** - 时序数据库，监控指标收集
- **Grafana** - 数据可视化，统一监控面板
- **Alertmanager** - 告警管理和路由
- **OpenTelemetry Collector** - 统一遥测数据收集
- **VictoriaLogs** - 高性能日志存储和查询

### 服务管理命令

```bash
# 安装服务（Docker 方式）
make deploy.install.docker.<service>

# 安装服务（原生方式）
make deploy.install.<service>

# 卸载服务
make deploy.uninstall.<service>
make deploy.uninstall.docker.<service>

# 可用服务: redis, mariadb, mongo, kafka, etcd, jaeger,
#          prometheus, grafana, alertmanager, otelcol, victorialogs
```

### 服务访问端点

| 服务 | 端口 | 访问地址 | 用途 |
|------|------|----------|------|
| Redis | 6379 | 127.0.0.1:6379 | 缓存访问 |
| MariaDB | 3306 | 127.0.0.1:3306 | 数据库访问 |
| MongoDB | 27017 | 127.0.0.1:27017 | 文档数据库 |
| etcd | 2379 | 127.0.0.1:2379 | 服务发现 |
| Kafka | 9092 | 127.0.0.1:9092 | 消息队列 |
| Jaeger | 16686 | <http://127.0.0.1:16686> | 链路追踪 UI |
| Prometheus | 9090 | <http://127.0.0.1:9090> | 监控查询 |
| Grafana | 3000 | <http://127.0.0.1:3000> | 监控面板 |
| Alertmanager | 9093 | <http://127.0.0.1:9093> | 告警管理 |
| VictoriaLogs | 9428 | <http://127.0.0.1:9428> | 日志查询 |

### 可观测性数据流

```m
应用程序 → OpenTelemetry Collector → Prometheus/Jaeger/VictoriaLogs
                                              ↓
                                          Grafana 可视化
                                              ↓
                                         Alertmanager 告警
```

## 重命名项目模块路径

如果您需要更改项目的 Go 模块路径 (例如，从 `github.com/old/project` 到 `github.com/new/project`)，可以使用 `rename-project` Make 目标。

**使用方法:**

```bash
make rename-project OLD_PATH=<current_module_path> NEW_PATH=<new_module_path>
```

**参数:**

- `OLD_PATH`: 当前项目的 Go 模块路径。
- `NEW_PATH`: 您希望使用的新 Go 模块路径。

**示例:**

假设您想将项目模块路径从 `github.com/costa92/go-protoc/v2` 更改为 `github.com/costa92/go-protoc/v3`，您可以运行：

```bash
make rename-project OLD_PATH=github.com/costa92/go-protoc/v2 NEW_PATH=github.com/costa92/go-protoc/v3
```

**注意:** 此命令会修改项目中的多个文件，包括 `go.mod`, `*.go` 文件, `*.sh` 文件等。执行后，请仔细检查更改并运行 `go mod tidy`。
