#!/usr/bin/env bash

# Kafka Only Docker运行脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: Kafka (仅Kafka)
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly SERVICE_NAME="kafka"
readonly CONTAINER_NAME="${PROJ_PREFIX}-kafka"
readonly SERVICE_PORT="${PROJ_KAFKA_PORT:-9092}"

# 目录配置
readonly CONFIG_DIR="${PROJ_KAFKA_CONFIG_DIR}"
readonly DATA_DIR="${PROJ_KAFKA_DATA_DIR}"
readonly LOG_DIR="${PROJ_KAFKA_LOG_DIR}"

# Docker网络
readonly NETWORK_NAME="${PROJ_NETWORK_NAME}"

# 平台检测和镜像选择
if [[ "$(uname)" == "Darwin" ]]; then
    # macOS - 使用 Confluent 镜像
    readonly IMAGE_NAME="confluentinc/cp-kafka:${KAFKA_VERSION}"
    readonly KAFKA_TOOLS_PATH="/bin"
else
    # Linux - 使用 Bitnami 镜像
    readonly IMAGE_NAME="bitnami/kafka:${KAFKA_VERSION}"
    readonly KAFKA_TOOLS_PATH="/opt/bitnami/kafka/bin"
fi
readonly KAFKA_BROKER_ID=1
readonly KAFKA_ZOOKEEPER_CONNECT="${CONTAINER_NAME_ZOOKEEPER}:2181"
readonly KAFKA_ADVERTISED_LISTENERS="PLAINTEXT://localhost:${SERVICE_PORT},PLAINTEXT_INTERNAL://${CONTAINER_NAME_KAFKA}:${PROJ_KAFKA_INTERNAL_PORT}"
readonly KAFKA_LISTENER_SECURITY_PROTOCOL_MAP="PLAINTEXT:PLAINTEXT,PLAINTEXT_INTERNAL:PLAINTEXT"
readonly KAFKA_INTER_BROKER_LISTENER_NAME="PLAINTEXT_INTERNAL"

echo "检测到平台: $(uname)"
echo "Kafka 版本: ${KAFKA_VERSION}"
echo "使用镜像: ${IMAGE_NAME}"

# 检查Zookeeper依赖 (不启动，只检查)
echo "检查Zookeeper依赖..."
if ! docker ps --filter name="${PROJ_PREFIX}-zookeeper" --format "{{.Names}}" | grep -q "${PROJ_PREFIX}-zookeeper"; then
    echo "❌ Zookeeper容器未运行，请先启动Zookeeper!"
    echo "提示: 使用命令 make docker.zookeeper.start 启动Zookeeper"
    exit 1
fi
echo "✅ Zookeeper容器已在运行"

# 验证Zookeeper连通性
echo "验证Zookeeper连接..."
# 等待Zookeeper健康检查通过，使用srvr命令替代ruok
max_retries=30
retry_count=0
while [ $retry_count -lt $max_retries ]; do
    if docker exec ${PROJ_PREFIX}-zookeeper sh -c "echo srvr | nc localhost 2181" 2>/dev/null | grep -q "Zookeeper version"; then
        echo "✅ Zookeeper连接正常"
        break
    fi
    echo "等待Zookeeper就绪... ($((retry_count+1))/$max_retries)"
    sleep 2
    ((retry_count++))
done

if [ $retry_count -eq $max_retries ]; then
    echo "❌ 无法连接到Zookeeper (localhost:${PROJ_ZOOKEEPER_PORT:-2181})"
    docker logs ${PROJ_PREFIX}-zookeeper --tail 20
    exit 1
fi

# 创建必要的目录
mkdir -p "${CONFIG_DIR}" "${DATA_DIR}" "${LOG_DIR}"

# 创建Docker网络（如果不存在）
docker network create "${NETWORK_NAME}" 2>/dev/null || true

# 创建Docker卷（如果不存在）
docker volume create "${CONTAINER_NAME_KAFKA}-data" 2>/dev/null || true
docker volume create "${CONTAINER_NAME_KAFKA}-logs" 2>/dev/null || true

# 停止并删除现有容器（如果存在）
docker stop "${CONTAINER_NAME}" 2>/dev/null || true
docker rm "${CONTAINER_NAME}" 2>/dev/null || true

# 启动 Kafka 容器
docker run -d \
    --name "${CONTAINER_NAME}" \
    --network "${NETWORK_NAME}" \
    --restart unless-stopped \
    -p "${SERVICE_PORT}:9092" \
    -p "29092:29092" \
    -p "9999:9999" \
    -v "${CONTAINER_NAME_KAFKA}-data:/var/lib/kafka/data" \
    -v "${CONTAINER_NAME_KAFKA}-logs:/var/log/kafka" \
    -v "${LOG_DIR}:/opt/kafka/logs" \
    -e KAFKA_BROKER_ID=${KAFKA_BROKER_ID} \
    -e KAFKA_ZOOKEEPER_CONNECT="${KAFKA_ZOOKEEPER_CONNECT}" \
    -e KAFKA_ADVERTISED_LISTENERS="${KAFKA_ADVERTISED_LISTENERS}" \
    -e KAFKA_LISTENER_SECURITY_PROTOCOL_MAP="${KAFKA_LISTENER_SECURITY_PROTOCOL_MAP}" \
    -e KAFKA_INTER_BROKER_LISTENER_NAME="${KAFKA_INTER_BROKER_LISTENER_NAME}" \
    -e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
    -e KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR=1 \
    -e KAFKA_TRANSACTION_STATE_LOG_MIN_ISR=1 \
    -e KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS=0 \
    -e KAFKA_AUTO_CREATE_TOPICS_ENABLE=true \
    -e KAFKA_DELETE_TOPIC_ENABLE=true \
    -e KAFKA_LOG_RETENTION_HOURS=168 \
    -e KAFKA_LOG_SEGMENT_BYTES=1073741824 \
    -e KAFKA_LOG_CLEANUP_POLICY=delete \
    -e KAFKA_NUM_PARTITIONS=3 \
    -e KAFKA_DEFAULT_REPLICATION_FACTOR=1 \
    -e KAFKA_MIN_INSYNC_REPLICAS=1 \
    -e KAFKA_UNCLEAN_LEADER_ELECTION_ENABLE=false \
    -e KAFKA_JMX_PORT=9999 \
    -e KAFKA_JMX_HOSTNAME=localhost \
    -e PROJ_SERVICE_NAME=kafka \
    -e PROJ_SERVICE_VERSION=${KAFKA_VERSION} \
    -e PROJ_ENVIRONMENT=${PROJ_ENVIRONMENT:-development} \
    --log-driver json-file \
    --log-opt max-size=10m \
    --log-opt max-file=3 \
    --health-cmd "${KAFKA_TOOLS_PATH}/kafka-topics --bootstrap-server ${PROJ_KAFKA_INTERNAL_BROKER} --list >/dev/null 2>&1" \
    --health-interval 30s \
    --health-timeout 15s \
    --health-retries 5 \
    --health-start-period 60s \
    "${IMAGE_NAME}"

echo "Kafka容器已启动: ${CONTAINER_NAME}"
echo "端口映射: localhost:${SERVICE_PORT} -> container:9092"
echo "内部通信端口: 29092 (容器间通信)"
echo "JMX监控端口: localhost:9999"
echo "Zookeeper依赖: ${KAFKA_ZOOKEEPER_CONNECT}"
echo "配置目录: ${CONFIG_DIR}"
echo "数据目录: ${DATA_DIR}"
echo "日志目录: ${LOG_DIR}"
echo "使用镜像: ${IMAGE_NAME}"

# 显示容器状态
docker ps --filter name="${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# 等待Kafka启动
echo "等待Kafka启动完成..."
timeout=120
while [ $timeout -gt 0 ]; do
    if docker exec "${CONTAINER_NAME}" ${KAFKA_TOOLS_PATH}/kafka-topics --bootstrap-server ${PROJ_KAFKA_INTERNAL_BROKER} --list >/dev/null 2>&1; then
        echo "Kafka启动成功 ✅"
        break
    fi
    sleep 3
    ((timeout-=3))
done

if [ $timeout -le 0 ]; then
    echo "Kafka启动超时，请检查日志 ❌"
    docker logs "${CONTAINER_NAME}" --tail 20
    exit 1
fi

echo ""
echo "🎉 Kafka服务已成功启动!"
echo "🔗 Bootstrap服务器: localhost:${SERVICE_PORT}"
echo "📊 JMX监控: localhost:9999"
echo "🔌 Zookeeper连接: ${KAFKA_ZOOKEEPER_CONNECT}"
echo "📋 集群配置: 单节点集群 (broker.id=${KAFKA_BROKER_ID})"
echo ""
echo "🛠️  常用命令:"
echo "  创建Topic: docker exec ${CONTAINER_NAME} ${KAFKA_TOOLS_PATH}/kafka-topics --create --topic test --bootstrap-server ${PROJ_KAFKA_INTERNAL_BROKER}"
echo "  列出Topics: docker exec ${CONTAINER_NAME} ${KAFKA_TOOLS_PATH}/kafka-topics --list --bootstrap-server ${PROJ_KAFKA_INTERNAL_BROKER}"
echo "  生产消息: docker exec -it ${CONTAINER_NAME} ${KAFKA_TOOLS_PATH}/kafka-console-producer --topic test --bootstrap-server ${PROJ_KAFKA_INTERNAL_BROKER}"
echo "  消费消息: docker exec -it ${CONTAINER_NAME} ${KAFKA_TOOLS_PATH}/kafka-console-consumer --topic test --bootstrap-server ${PROJ_KAFKA_INTERNAL_BROKER} --from-beginning"