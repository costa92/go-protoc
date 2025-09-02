# Docker脚本模式使用指南

本项目采用**纯Docker命令**模式部署服务，完全不依赖docker-compose。所有服务通过基于模板生成的Shell脚本进行管理。

## 核心优势

✅ **无docker-compose依赖**: 只需要docker命令  
✅ **脚本化管理**: 生成的脚本可直接执行  
✅ **模板化配置**: 支持环境变量定制  
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

## 使用方法

### 1. 基础使用

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

### 2. 批量服务管理

```bash
# 启动多个服务
proj::docker::manage_services "start" "redis" "mysql" "victorialogs"

# 检查多个服务状态
proj::docker::manage_services "status" "redis" "mysql" "victorialogs"

# 停止多个服务
proj::docker::manage_services "stop" "redis" "mysql" "victorialogs"

# 重启服务
proj::docker::manage_services "restart" "redis"
```

### 3. 手动脚本生成

```bash
# 生成特定服务的Docker脚本
proj::docker::generate_service_scripts "redis" "/tmp/redis-scripts"

# 执行生成的脚本
bash /tmp/redis-scripts/docker-run.sh
bash /tmp/redis-scripts/docker-status.sh
bash /tmp/redis-scripts/docker-stop.sh
```

### 4. 脚本管理

```bash
# 列出所有生成的脚本
proj::docker::list_generated_scripts

# 列出特定服务的脚本
proj::docker::list_generated_scripts "redis"

# 清理生成的脚本
proj::docker::cleanup_generated_scripts "redis"  # 清理特定服务
proj::docker::cleanup_generated_scripts          # 清理所有脚本
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

### Redis 配置示例

```bash
# 环境变量配置
export REDIS_VERSION="7.2.4"
export PROJ_REDIS_PORT="6379"
export PROJ_REDIS_CONFIG_DIR="/Users/costalong/code/go/src/github.com/costa92/go-protoc/_data/redis/config"
export PROJ_REDIS_DATA_DIR="/Users/costalong/code/go/src/github.com/costa92/go-protoc/_data/redis/data"

# 启动Redis
proj::docker::run_service_from_template "redis"
```

生成的Docker命令：
```bash
docker run -d \
    --name "proj-redis" \
    --network "proj-network" \
    --restart unless-stopped \
    -p "6379:6379" \
    -v proj-redis-data:/data \
    -v "${CONFIG_DIR}/redis.conf:/usr/local/etc/redis/redis.conf:ro" \
    -e REDIS_REPLICATION_MODE=master \
    --health-cmd "redis-cli ping" \
    redis:7.2.4 \
    redis-server /usr/local/etc/redis/redis.conf
```

### MySQL 配置示例

```bash
# 环境变量配置
export MYSQL_VERSION="8.0"
export PROJ_MYSQL_PORT="3306"
export MYSQL_ROOT_PASSWORD="proj(#)666"
export MYSQL_DATABASE="onex"

# 启动MySQL
proj::docker::run_service_from_template "mysql"
```

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

所有服务运行在统一的Docker网络 `proj-network` 中：

```
proj-network (bridge)
├── proj-redis (6379)
├── proj-mysql (3306)  
├── proj-kafka (9092, 29092, 9999)
├── proj-zookeeper (2181, 2888, 3888)
├── proj-otelcol (4327, 4328, 8888, 13133)
├── proj-victorialogs (9428)
└── proj-{service} (...)
```

服务间可以通过容器名直接通信：
- Redis: `proj-redis:6379`
- MySQL: `proj-mysql:3306`  
- Kafka: `proj-kafka:9092` (外部) / `proj-kafka:29092` (内部)
- Zookeeper: `proj-zookeeper:2181`
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

推荐在Makefile中集成这些Docker脚本管理功能：

```makefile
# Docker服务管理
.PHONY: docker-start-redis docker-stop-redis docker-status-redis
.PHONY: docker-start-kafka docker-stop-kafka docker-status-kafka

docker-start-redis:
	@./scripts/installation/lib/docker_script_manager.sh start redis

docker-stop-redis:
	@./scripts/installation/lib/docker_script_manager.sh stop redis

docker-status-redis:
	@./scripts/installation/lib/docker_script_manager.sh status redis

docker-start-kafka:
	@./scripts/installation/lib/docker_script_manager.sh start kafka

docker-stop-kafka:
	@./scripts/installation/lib/docker_script_manager.sh stop kafka

docker-status-kafka:
	@./scripts/installation/lib/docker_script_manager.sh status kafka
```

## 高级功能

### 自定义环境变量文件

```bash
# 创建环境变量文件
cat > /tmp/redis.env << EOF
REDIS_MAX_MEMORY=512mb
REDIS_PASSWORD=my-secret
EOF

# 使用自定义环境变量启动
proj::docker::run_service_from_template "redis" "/tmp/redis.env"
```

### 脚本钩子

可以在生成的脚本中添加前置和后置钩子：

```bash
# 在docker-run.sh.tpl中添加
# Pre-start hook
if [[ -x "${CONFIG_DIR}/pre-start.sh" ]]; then
    bash "${CONFIG_DIR}/pre-start.sh"
fi

# ... container run command ...

# Post-start hook  
if [[ -x "${CONFIG_DIR}/post-start.sh" ]]; then
    bash "${CONFIG_DIR}/post-start.sh"
fi
```

这种设计提供了最大的灵活性和可控性，同时保持了简洁性和标准化。