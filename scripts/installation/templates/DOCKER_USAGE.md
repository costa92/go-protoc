# Docker脚本模式使用指南

本项目采用**纯Docker命令**模式部署服务，完全不依赖docker-compose。所有服务通过基于模板生成的Shell脚本进行管理。

## 核心优势

✅ **无docker-compose依赖**: 只需要docker命令  
✅ **脚本化管理**: 生成的脚本可直接执行  
✅ **模板化配置**: 支持环境变量定制  
✅ **多环境支持**: 支持 dev/test/prod 环境动态切换  
✅ **统一配置源**: 所有配置来自 `manifests/env` 统一管理  
✅ **统一网络管理**: 自动创建和管理项目网络  
✅ **健康检查集成**: 内置容器健康监控  
✅ **日志标准化**: 统一的日志配置和轮转  

## 目录结构

```
templates/docker/{service}/
├── docker-run.sh.tpl    # 容器启动脚本模板
├── docker-stop.sh.tpl   # 容器停止脚本模板
├── docker-status.sh.tpl # 容器状态检查模板
└── {service}.conf.tpl   # 服务配置文件模板
```

## 配置系统

本项目使用**统一配置源**，所有版本和环境配置来自 `manifests/env` 目录，**禁止引用** `scripts/installation/versions.sh`。

### 环境配置文件

```
manifests/env/
├── env.base    # 基础配置，包含所有版本信息
├── env.dev     # 开发环境配置
├── env.test    # 测试环境配置
└── env.prod    # 生产环境配置
```

### 环境变量优先级

1. **固定值配置**: 项目标识、版本信息等
2. **第三方组件版本**: Redis 7.2.4, MariaDB 11.2.2 等
3. **环境特定配置**: 端口前缀、网络名称等
4. **服务配置**: 按依赖顺序 DB → Cache → MQ → Discovery → Monitoring

## 使用方法

### 1. Makefile集成使用（推荐）

```bash
# 开发环境（默认，端口前缀: 1）
make docker.redis.start
make docker.redis.status
make docker.redis.stop

# 测试环境（端口前缀: 2）
PROJ_ENVIRONMENT=test make docker.redis.start
PROJ_ENVIRONMENT=test make docker.redis.status
PROJ_ENVIRONMENT=test make docker.redis.stop

# 生产环境（无端口前缀）
PROJ_ENVIRONMENT=prod make docker.redis.start
PROJ_ENVIRONMENT=prod make docker.redis.status
PROJ_ENVIRONMENT=prod make docker.redis.stop

# 查看版本信息
make docker.env.versions
PROJ_ENVIRONMENT=test make docker.env.versions
PROJ_ENVIRONMENT=prod make docker.env.versions

# 查看帮助
make docker.help.templates
```

### 2. 低级API直接使用

```bash
# 加载工具库
source scripts/installation/lib/common_lib.sh

# 启动Redis服务
proj::docker::run_service_from_template "redis"

# 检查Redis状态
proj::docker::check_service_status_from_template "redis"

# 停止Redis服务
proj::docker::stop_service_from_template "redis"
```

### 3. 批量服务管理

```bash
# Makefile批量操作（推荐）
make docker.test-services.start   # 启动所有测试服务
make docker.test-services.status  # 检查所有服务状态  
make docker.test-services.stop    # 停止所有服务

# 不同环境的批量操作
PROJ_ENVIRONMENT=test make docker.test-services.start
PROJ_ENVIRONMENT=prod make docker.test-services.start

# 低级API批量操作
proj::docker::manage_services "start" "redis" "mysql" "victorialogs"
proj::docker::manage_services "status" "redis" "mysql" "victorialogs"
proj::docker::manage_services "stop" "redis" "mysql" "victorialogs"
proj::docker::manage_services "restart" "redis"
```

### 4. 手动脚本生成

```bash
# 生成特定服务的Docker脚本
proj::docker::generate_service_scripts "redis" "/tmp/redis-scripts"

# 执行生成的脚本
bash /tmp/redis-scripts/docker-run.sh
bash /tmp/redis-scripts/docker-status.sh
bash /tmp/redis-scripts/docker-stop.sh

# Makefile脚本管理（推荐）
make docker.redis.cleanup     # 清理Redis脚本
make docker.scripts.list      # 列出所有生成的脚本
make docker.scripts.cleanup-all  # 清理所有脚本
```

### 5. 脚本管理

```bash
# 低级API脚本管理
proj::docker::list_generated_scripts
proj::docker::list_generated_scripts "redis"
proj::docker::cleanup_generated_scripts "redis"
proj::docker::cleanup_generated_scripts

# Makefile脚本管理（推荐）
make docker.scripts.list           # 列出所有生成的脚本
make docker.redis.cleanup          # 清理特定服务脚本
make docker.scripts.cleanup-all    # 清理所有脚本
```

## 脚本功能详解

### docker-run.sh 脚本功能

- **网络管理**: 自动创建项目网络 `proj-network`
- **卷管理**: 创建命名卷进行数据持久化
- **容器清理**: 自动清理同名旧容器
- **端口映射**: 标准化的端口映射配置
- **环境变量**: 注入项目和服务相关环境变量
- **健康检查**: 配置容器健康检查策略
- **日志配置**: 统一的日志驱动和轮转策略
- **启动验证**: 验证服务是否正确启动

### docker-stop.sh 脚本功能

- **优雅停止**: 先停止后删除容器
- **数据保护**: 可选的数据卷删除（需要确认）
- **状态反馈**: 提供详细的停止过程信息

### docker-status.sh 脚本功能

- **容器状态**: 检查容器运行状态
- **健康检查**: 显示容器健康状态
- **连接测试**: 验证服务连接可用性
- **资源监控**: 显示CPU和内存使用情况
- **日志摘要**: 显示最近的容器日志

## 服务配置

### 多环境配置示例

所有配置统一从 `manifests/env` 目录加载，支持环境特定的端口和网络配置：

| 环境 | 端口前缀 | 网络名称 | Redis端口 | MySQL端口 |
|-----|---------|----------|----------|-----------|
| Development (默认) | 1 | proj-dev-network | 16379 | 13306 |
| Test | 2 | proj-test-network | 26379 | 23306 |
| Production | 无前缀 | proj-prod-network | 6379 | 3306 |

### Redis 配置示例

```bash
# 开发环境（默认）
make docker.redis.start  # 端口: 16379, 网络: proj-dev-network

# 测试环境  
PROJ_ENVIRONMENT=test make docker.redis.start  # 端口: 26379, 网络: proj-test-network

# 生产环境
PROJ_ENVIRONMENT=prod make docker.redis.start  # 端口: 6379, 网络: proj-prod-network
```

生成的Docker命令示例（开发环境）：
```bash
docker run -d \
    --name "proj-redis" \
    --network "proj-dev-network" \
    --restart unless-stopped \
    -p "16379:6379" \
    -v proj-redis-data:/data \
    -v "${CONFIG_DIR}/redis.conf:/usr/local/etc/redis/redis.conf:ro" \
    -e REDIS_REPLICATION_MODE=master \
    -e PROJ_SERVICE_NAME=redis \
    -e PROJ_SERVICE_VERSION=7.2.4 \
    -e PROJ_ENVIRONMENT=development \
    --health-cmd "redis-cli ping" \
    redis:7.2.4 \
    redis-server /usr/local/etc/redis/redis.conf
```

### MySQL/MariaDB 配置示例

```bash
# 开发环境（端口: 13306）
make docker.mariadb.start

# 测试环境（端口: 23306）
PROJ_ENVIRONMENT=test make docker.mariadb.start

# 生产环境（端口: 3306）
PROJ_ENVIRONMENT=prod make docker.mariadb.start

# 连接到数据库
make docker.mariadb.connect
# 或者直接使用: docker exec -it proj-mariadb mariadb -u root -p
```

配置来源（自动从环境文件加载）：
- **版本**: MariaDB 11.2.2（来自 env.base）
- **数据库名**: protoc（开发环境）、test_protoc（测试环境）、protoc_prod（生产环境）
- **用户凭据**: root/proj(#)666（来自环境配置）

### Kafka 配置示例

```bash
# 环境变量配置
export KAFKA_VERSION="6.2.0"
export PROJ_KAFKA_PORT="9092"
export KAFKA_BOOTSTRAP_SERVERS="proj-kafka:9092"
export KAFKA_ZOOKEEPER_CONNECT="proj-zookeeper:2181"
export PROJ_KAFKA_CONFIG_DIR="/path/to/kafka/config"
export PROJ_KAFKA_DATA_DIR="/path/to/kafka/data"
export PROJ_KAFKA_LOG_DIR="/path/to/kafka/logs"

# 启动Kafka（会自动检查和启动依赖的Zookeeper）
proj::docker::run_service_from_template "kafka"
```

生成的Docker命令：
```bash
docker run -d \
    --name "proj-kafka" \
    --network "proj-network" \
    --restart unless-stopped \
    -p "9092:9092" \
    -p "29092:29092" \
    -p "9999:9999" \
    -v proj-kafka-data:/opt/kafka/logs \
    -v proj-kafka-logs:/var/log/kafka \
    -e KAFKA_BROKER_ID=1 \
    -e KAFKA_ZOOKEEPER_CONNECT="proj-zookeeper:2181" \
    -e KAFKA_ADVERTISED_LISTENERS="PLAINTEXT://localhost:9092,PLAINTEXT_INTERNAL://proj-kafka:29092" \
    -e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
    -e KAFKA_AUTO_CREATE_TOPICS_ENABLE=true \
    --health-cmd "/bin/kafka-topics --bootstrap-server localhost:9092 --list >/dev/null 2>&1" \
    --health-interval 30s \
    --health-timeout 15s \
    --health-retries 5 \
    --health-start-period 60s \
    confluentinc/cp-kafka:6.2.0
```

**特殊功能**：
- ✅ **自动依赖管理**: 启动时自动检查并启动Zookeeper依赖
- ✅ **健康检查修复**: 使用正确的命令路径 `/bin/kafka-topics`
- ✅ **多端口支持**: 外部访问(9092)、内部通信(29092)、JMX监控(9999)
- ✅ **单节点集群**: 适合开发环境的单节点配置

### OpenTelemetry Collector 配置示例

```bash
# 环境变量配置
export OTELCOL_VERSION="0.132.0"
export PROJ_OTELCOL_GRPC_PORT="4327"
export PROJ_OTELCOL_HTTP_PORT="4328"

# 启动OTEL Collector
proj::docker::run_service_from_template "otelcol"
```

## 网络架构

每个环境运行在独立的Docker网络中，避免环境间的冲突：

### 开发环境网络 (`proj-dev-network`)
```
proj-dev-network (bridge)
├── proj-redis (内部:6379, 外部:16379)
├── proj-mariadb (内部:3306, 外部:13306)
├── proj-kafka (内部:9092, 外部:19092)
├── proj-nacos (内部:8848, 外部:18848)
├── proj-victorialogs (内部:9428, 外部:19428)
└── proj-{service} (...)
```

### 测试环境网络 (`proj-test-network`)  
```
proj-test-network (bridge)
├── proj-redis (内部:6379, 外部:26379)
├── proj-mariadb (内部:3306, 外部:23306)
├── proj-kafka (内部:9092, 外部:29092)
├── proj-nacos (内部:8848, 外部:28848)
├── proj-victorialogs (内部:9428, 外部:29428)
└── proj-{service} (...)
```

### 生产环境网络 (`proj-prod-network`)
```
proj-prod-network (bridge)  
├── proj-redis (内部:6379, 外部:6379)
├── proj-mariadb (内部:3306, 外部:3306)
├── proj-kafka (内部:9092, 外部:9092)
├── proj-nacos (内部:8848, 外部:8848)
├── proj-victorialogs (内部:9428, 外部:9428)
└── proj-{service} (...)
```

### 服务间通信

在同一环境内，服务可以通过容器名直接通信（使用内部端口）：
- Redis: `proj-redis:6379`
- MariaDB: `proj-mariadb:3306`  
- Kafka: `proj-kafka:9092` (外部) / `proj-kafka:29092` (内部)
- Nacos: `proj-nacos:8848`
- VictoriaLogs: `proj-victorialogs:9428`

## 数据持久化

每个服务使用命名卷进行数据持久化：

| 服务 | Docker卷名称 | 容器挂载点 |
|-----|-------------|-----------|
| Redis | `proj-redis-data` | `/data` |
| MySQL | `proj-mysql-data` | `/var/lib/mysql` |
| Kafka | `proj-kafka-data`, `proj-kafka-logs` | `/opt/kafka/logs`, `/var/log/kafka` |
| Zookeeper | `proj-zookeeper-data`, `proj-zookeeper-logs` | `/var/lib/zookeeper/data`, `/var/lib/zookeeper/log` |
| OTEL Collector | `proj-otelcol-data` | `/data` |
| VictoriaLogs | `proj-victorialogs-data` | `/victoria-logs-data` |

## 故障排除

### 容器启动失败

```bash
# 检查容器日志
docker logs proj-{service}

# 检查容器状态
docker inspect proj-{service}

# 重新生成脚本
proj::docker::cleanup_generated_scripts "{service}"
proj::docker::run_service_from_template "{service}"
```

### 网络问题

```bash
# 检查项目网络
docker network ls | grep proj-network

# 重新创建网络
docker network rm proj-network
proj::docker::ensure_project_network
```

### 端口冲突

```bash
# 检查端口占用
lsof -i :6379

# 修改环境变量中的端口配置
export PROJ_REDIS_PORT="6380"
```

### 数据卷问题

```bash
# 检查数据卷
docker volume ls | grep proj-

# 删除损坏的数据卷（⚠️ 会丢失数据）
docker volume rm proj-redis-data
```

### Kafka专项问题

```bash
# Kafka健康检查失败
docker exec proj-kafka /bin/kafka-topics --bootstrap-server localhost:9092 --list

# 检查Zookeeper依赖
docker exec proj-zookeeper sh -c "echo srvr | nc localhost 2181"

# Kafka容器卡在启动状态
docker logs proj-kafka --tail 50
docker logs proj-zookeeper --tail 20

# 重新启动Kafka集群
docker stop proj-kafka proj-zookeeper
docker rm proj-kafka proj-zookeeper
# 然后重新启动
```

## 与Makefile集成

本项目已完全集成Docker模板系统到Makefile中，**推荐使用Make命令**进行日常操作：

### 可用的Make命令

```bash
# 个人服务管理（支持环境切换）
make docker.{service}.start      # 启动服务
make docker.{service}.stop       # 停止服务  
make docker.{service}.status     # 检查状态
make docker.{service}.restart    # 重启服务
make docker.{service}.cleanup    # 清理脚本

# 支持的服务: redis, mysql, mariadb, etcd, otelcol, victorialogs, prometheus, kafka, nacos 等

# 批量服务管理
make docker.test-services.start   # 启动所有测试服务
make docker.test-services.stop    # 停止所有测试服务
make docker.test-services.status  # 检查所有服务状态
make docker.test-services.restart # 重启所有服务

# 环境和配置管理
make docker.env.check             # 检查Docker环境
make docker.env.versions          # 显示版本信息
make docker.network.create        # 创建项目网络
make docker.network.info          # 显示网络信息

# 脚本和调试
make docker.scripts.list          # 列出生成的脚本
make docker.scripts.cleanup-all   # 清理所有脚本
make docker.debug.redis           # Redis调试信息

# 连接和日志
make docker.redis.connect         # 连接Redis CLI
make docker.mariadb.connect       # 连接MariaDB CLI  
make docker.redis.logs            # 显示Redis日志
make docker.mariadb.logs          # 显示MariaDB日志

# 帮助
make docker.help.templates        # 显示完整帮助信息
```

### 环境切换示例

```bash
# 开发环境（默认）
make docker.redis.start           # 端口: 16379

# 测试环境
PROJ_ENVIRONMENT=test make docker.redis.start    # 端口: 26379

# 生产环境  
PROJ_ENVIRONMENT=prod make docker.redis.start    # 端口: 6379

# 查看不同环境的版本
make docker.env.versions
PROJ_ENVIRONMENT=test make docker.env.versions
PROJ_ENVIRONMENT=prod make docker.env.versions
```

### 自定义集成

如果需要添加自定义的Makefile规则：

```makefile
# 自定义服务操作
.PHONY: dev-setup test-setup prod-deploy

dev-setup:
	@echo "设置开发环境..."
	@make docker.redis.start
	@make docker.mariadb.start
	@make docker.victorialogs.start

test-setup:
	@echo "设置测试环境..."
	@PROJ_ENVIRONMENT=test make docker.redis.start
	@PROJ_ENVIRONMENT=test make docker.mariadb.start

prod-deploy:
	@echo "部署生产环境..."
	@PROJ_ENVIRONMENT=prod make docker.redis.start
	@PROJ_ENVIRONMENT=prod make docker.mariadb.start
```

## 高级功能

### 配置管理最佳实践

#### 1. 环境配置定制

修改 `manifests/env/env.{dev,test,prod}` 文件来定制环境特定的配置：

```bash
# manifests/env/env.dev 示例
export PROJ_ENVIRONMENT=development
export PROJ_ACCESS_PORT_PREFIX="1"

# 开发环境特定配置
export REDIS_MAX_MEMORY=512m  # 开发环境使用较小内存
export NACOS_MAX_MEMORY=512m
```

#### 2. 版本管理

所有版本信息统一在 `manifests/env/env.base` 中管理：

```bash
# 第三方组件版本 (Third-party Component Versions)
export REDIS_VERSION=7.2.4
export MARIADB_VERSION=11.2.2
export MONGODB_VERSION=7.0.5
# ... 其他版本
```

**重要**: 禁止在Docker模板系统中引用 `scripts/installation/versions.sh`，确保版本信息来源唯一。

#### 3. 网络和端口定制

通过环境变量定制网络和端口配置：

```bash
# 自定义网络名称
export PROJ_NETWORK_NAME="my-custom-network"

# 自定义端口前缀（测试环境示例）
export PROJ_ACCESS_PORT_PREFIX="3"  # 所有服务端口前缀为3

# 启动服务
PROJ_ENVIRONMENT=test make docker.redis.start  # 端口: 36379
```

#### 4. 脚本钩子和扩展

可以在生成的脚本模板中添加钩子：

```bash
# 在 scripts/installation/templates/docker/redis/docker-run.sh.tpl 中添加
# Pre-start hook
if [[ -x "${PROJ_REDIS_CONFIG_DIR}/pre-start.sh" ]]; then
    echo "执行Redis启动前钩子..."
    bash "${PROJ_REDIS_CONFIG_DIR}/pre-start.sh"
fi

# ... container run command ...

# Post-start hook  
if [[ -x "${PROJ_REDIS_CONFIG_DIR}/post-start.sh" ]]; then
    echo "执行Redis启动后钩子..."
    bash "${PROJ_REDIS_CONFIG_DIR}/post-start.sh"
fi
```

#### 5. 调试和故障排除

```bash
# 调试特定服务的模板生成
make docker.debug.redis

# 查看生成的脚本内容
make docker.scripts.list
cat _generated/docker-scripts/redis/docker-run.sh

# 手动执行步骤进行调试
make docker.redis.cleanup
PROJ_ENVIRONMENT=test make docker.redis.start
```

### 系统集成

#### CI/CD 集成示例

```bash
# .github/workflows/test.yml 示例
- name: Setup Test Environment
  run: |
    PROJ_ENVIRONMENT=test make docker.test-services.start
    
- name: Run Tests
  run: |
    # 测试服务现在运行在测试端口
    # Redis: localhost:26379
    # MariaDB: localhost:23306
    make test
    
- name: Cleanup
  run: |
    PROJ_ENVIRONMENT=test make docker.test-services.stop
```

#### Docker Compose 迁移

如果从docker-compose迁移，这个系统提供了相同的功能：

| docker-compose | Docker模板系统 |
|----------------|---------------|
| `docker-compose up redis` | `make docker.redis.start` |
| `docker-compose ps` | `make docker.containers.info` |
| `docker-compose down` | `make docker.test-services.stop` |
| `docker-compose logs redis` | `make docker.redis.logs` |

这种设计提供了比docker-compose更精细的控制，同时保持了简洁性和可扩展性。