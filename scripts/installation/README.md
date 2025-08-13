# 安装脚本服务功能与关联说明

本目录包含完整的服务依赖安装脚本，为 Go 微服务框架提供数据存储、消息队列、可观测性等基础设施支持。

## 服务架构分层

### 数据存储层 (Data Layer)

- **Redis** (`redis.sh`) - 缓存服务，高性能键值存储
- **MariaDB** (`mariadb.sh`) - 关系型数据库，MySQL兼容
- **MongoDB** (`mongo.sh`) - NoSQL文档数据库
- **etcd** (`etcd.sh`) - 分布式键值存储，服务发现

### 消息传输层 (Message Layer)

- **Kafka** (`kafka.sh`) - 分布式消息队列，事件流处理

### 可观测性栈 (Observability Stack)

- **Jaeger** (`jaeger.sh`) - 分布式链路追踪
- **Prometheus** (`prometheus.sh`) - 时序数据库，监控指标收集
- **Grafana** (`grafana.sh`) - 数据可视化，统一监控面板
- **Alertmanager** (`alertmanager.sh`) - 告警管理和路由
- **OpenTelemetry Collector** (`otelcol.sh`) - 统一遥测数据收集
- **VictoriaLogs** (`victorialogs.sh`) - 高性能日志存储和查询

## 服务关联关系

### 主数据流

```
应用程序 → [OpenTelemetry Collector] → [Jaeger/Prometheus/VictoriaLogs]
     ↓                                         ↓
   业务数据                                监控数据
     ↓                                         ↓
[Redis/MariaDB/MongoDB]                   [Grafana显示]
     ↓                                         ↓
   持久化存储                            [Alertmanager告警]
```

### 服务发现与配置

```
etcd ← 服务注册发现 → 应用服务
  ↓
配置管理
```

### 消息流转

```
应用服务 → Kafka → 消息消费者
```

### 可观测性数据流

```
应用指标/追踪/日志 → OpenTelemetry Collector → Prometheus/Jaeger/VictoriaLogs
                                                      ↓
                                                   Grafana
                                                      ↓
                                                 Alertmanager
```

## 安装使用方式

### 统一安装

```bash
# 安装所有服务
./install.sh

# 或通过 Makefile
make dev-setup
```

### 单独安装服务

```bash
# Docker 方式
./redis.sh proj::redis::docker::install
./mariadb.sh proj::mariadb::docker::install
./victorialogs.sh proj::victorialogs::docker::install

# 原生安装方式
./redis.sh proj::redis::install
./mariadb.sh proj::mariadb::install
./victorialogs.sh proj::victorialogs::install
```

### 服务管理

```bash
# 状态检查
./redis.sh proj::redis::status

# 卸载服务
./redis.sh proj::redis::uninstall
./redis.sh proj::redis::docker::uninstall
```

## 部署模式

每个服务支持两种安装方式：

### Docker容器化部署

- **优势**: 隔离性好，部署简单，环境一致
- **命名规范**: `${NETWORK_NAME}-<service>` (如 `proj-redis`)
- **网络**: 统一Docker网络 `proj`
- **数据持久化**: 挂载到 `${PROJ_THIRDPARTY_INSTALL_DIR}`

### 原生系统服务(SBS - Step By Step)

- **优势**: 性能更好，资源占用少
- **管理**: systemd服务管理
- **配置**: 标准系统路径
- **用户隔离**: 每个服务独立系统用户

## 端口分配

| 服务 | 默认端口 | 功能 | 脚本文件 |
|------|----------|------|----------|
| Redis | 6379 | 缓存访问 | `redis.sh` |
| MariaDB | 3306 | 数据库访问 | `mariadb.sh` |
| MongoDB | 27017 | 文档数据库 | `mongo.sh` |
| etcd | 2379/2380 | 客户端/集群通信 | `etcd.sh` |
| Kafka | 4317 | 消息队列 | `kafka.sh` |
| Jaeger | 16686/14268 | UI/数据收集 | `jaeger.sh` |
| Prometheus | 9090 | 监控查询 | `prometheus.sh` |
| Grafana | 3000 | 可视化面板 | `grafana.sh` |
| Alertmanager | 9093 | 告警管理 | `alertmanager.sh` |
| OTEL Collector | 4317/4318 | gRPC/HTTP接收 | `otelcol.sh` |
| VictoriaLogs | 9428 | 日志存储查询 | `victorialogs.sh` |

## 环境变量配置

每个服务都支持通过环境变量自定义配置：

### Redis 配置

```bash
PROJ_REDIS_HOST=127.0.0.1
PROJ_REDIS_PORT=6379
PROJ_REDIS_PASSWORD=proj(#)666
```

### MariaDB 配置

```bash
PROJ_MYSQL_HOST=127.0.0.1
PROJ_MYSQL_PORT=3306
PROJ_PASSWORD=proj(#)666
```

### MongoDB 配置

```bash
PROJ_MONGO_HOST=127.0.0.1
PROJ_MONGO_PORT=27017
PROJ_MONGO_ADMIN_USERNAME=root
PROJ_MONGO_ADMIN_PASSWORD=proj(#)666
```

### VictoriaLogs 配置

```bash
PROJ_VICTORIALOGS_HOST=127.0.0.1
PROJ_VICTORIALOGS_PORT=9428
PROJ_VICTORIALOGS_DATA_DIR=/var/lib/victorialogs
PROJ_VICTORIALOGS_VERSION=v0.5.2-victorialogs
```

## 统一管理特性

- **环境变量配置**: 所有服务通过环境变量统一配置
- **健康检查**: 内置端口探测和服务状态检查
- **幂等安装**: 支持重复执行，自动跳过已安装组件
- **完整卸载**: 清理所有相关文件和配置
- **日志管理**: 统一日志路径和格式
- **权限管理**: 服务用户隔离，最小权限原则
- **版本控制**: 支持指定服务版本

## 文件说明

- `install.sh` - 主安装脚本，统一调用各服务安装
- `common.sh` - 公共函数库，网络管理、工具函数
- `service.sh` - 服务管理相关函数
- `docker-compose.sh` - Docker Compose 相关功能
- `victorialogs.sh` - VictoriaLogs 日志存储服务安装脚本
- `各服务.sh` - 单独服务的安装、配置、管理脚本

## 依赖关系

```sh
install.sh
├── common.sh (公共函数)
├── redis.sh
├── mariadb.sh
├── mongo.sh
├── kafka.sh (依赖 zookeeper)
├── etcd.sh
├── jaeger.sh
├── prometheus.sh
├── grafana.sh (可视化 prometheus 数据)
├── alertmanager.sh (处理 prometheus 告警)
├── otelcol.sh (收集数据到 jaeger/prometheus/victorialogs)
└── victorialogs.sh (高性能日志存储和查询)
```

## 故障排查

### 检查服务状态

```bash
# 检查所有服务状态
for service in redis mariadb mongo kafka etcd jaeger prometheus grafana alertmanager otelcol victorialogs; do
    echo "=== $service ==="
    ./${service}.sh proj::${service}::status
done
```

### 查看日志

```bash
# Docker 服务日志
docker logs proj-redis
docker logs proj-victorialogs

# 系统服务日志
journalctl -u redis-server -f
journalctl -u victorialogs -f
```

### 网络连通性

```bash
# 检查端口监听
netstat -tlnp | grep -E "(6379|3306|27017|2379|9090|3000|9428)"

# 检查容器网络
docker network inspect proj

# VictoriaLogs 特定检查
curl -f http://127.0.0.1:9428/health
curl -s http://127.0.0.1:9428/metrics | head -10
```
