#!/usr/bin/env bash

# Zookeeper Docker运行脚本模板 (for Kafka)
# Project: ${PROJ_NAME:-go-protoc}
# Service: Zookeeper ${ZOOKEEPER_VERSION} (Kafka依赖)
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly SERVICE_NAME="zookeeper"
readonly CONTAINER_NAME="${PROJ_PREFIX}-zookeeper"
readonly SERVICE_PORT="${PROJ_ZOOKEEPER_PORT:-2181}"

# 目录配置
readonly CONFIG_DIR="${PROJ_ZOOKEEPER_CONFIG_DIR}"
readonly DATA_DIR="${PROJ_ZOOKEEPER_DATA_DIR}"
readonly LOG_DIR="${PROJ_ZOOKEEPER_LOG_DIR}"

# Docker网络
readonly NETWORK_NAME="${PROJ_NETWORK_NAME}"

# 平台检测和镜像选择
if [[ "$(uname)" == "Darwin" ]]; then
    # macOS - 使用 Confluent 镜像，支持 ARM64
    readonly IMAGE_NAME="confluentinc/cp-zookeeper:${ZOOKEEPER_VERSION:-7.4.0}"
else
    # Linux - 使用 Confluent 镜像
    readonly IMAGE_NAME="confluentinc/cp-zookeeper:${ZOOKEEPER_VERSION:-latest}"
fi

echo "检测到平台: $(uname)"
echo "使用镜像: ${IMAGE_NAME}"

# 创建必要的目录
mkdir -p "${CONFIG_DIR}" "${DATA_DIR}" "${LOG_DIR}"

# 创建Docker网络（如果不存在）
docker network create "${NETWORK_NAME}" 2>/dev/null || true

# 创建Docker卷（如果不存在）
docker volume create "${PROJ_PREFIX}-zookeeper-data" 2>/dev/null || true
docker volume create "${PROJ_PREFIX}-zookeeper-logs" 2>/dev/null || true

# 停止并删除现有容器（如果存在）
docker stop "${CONTAINER_NAME}" 2>/dev/null || true
docker rm "${CONTAINER_NAME}" 2>/dev/null || true

# 启动 Zookeeper 容器
echo "正在启动Zookeeper容器..."
if [[ "$(uname)" == "Darwin" ]]; then
    # macOS - 使用 Confluent 镜像
    docker run -d \
        --name "${CONTAINER_NAME}" \
        --network "${NETWORK_NAME}" \
        --restart unless-stopped \
        -p "${SERVICE_PORT}:2181" \
        -v "${PROJ_PREFIX}-zookeeper-data:/var/lib/zookeeper/data" \
        -v "${PROJ_PREFIX}-zookeeper-logs:/var/lib/zookeeper/log" \
        -e ZOOKEEPER_CLIENT_PORT=2181 \
        -e ZOOKEEPER_TICK_TIME=2000 \
        -e ZOOKEEPER_INIT_LIMIT=10 \
        -e ZOOKEEPER_SYNC_LIMIT=5 \
        -e ZOOKEEPER_MAX_CLIENT_CNXNS=60 \
        -e ZOOKEEPER_SNAP_RETAIN_COUNT=3 \
        -e ZOOKEEPER_PURGE_INTERVAL=12 \
        -e PROJ_SERVICE_NAME=zookeeper \
        -e PROJ_SERVICE_VERSION=${ZOOKEEPER_VERSION:-7.4.0} \
        -e PROJ_ENVIRONMENT=${PROJ_ENVIRONMENT:-development} \
        --log-driver json-file \
        --log-opt max-size=10m \
        --log-opt max-file=3 \
        --health-cmd "echo ruok | nc localhost 2181 | grep imok" \
        --health-interval 30s \
        --health-timeout 10s \
        --health-retries 3 \
        --health-start-period 30s \
        "${IMAGE_NAME}"
else
    # Linux - 使用 Confluent 镜像
    docker run -d \
        --name "${CONTAINER_NAME}" \
        --network "${NETWORK_NAME}" \
        --restart unless-stopped \
        -p "${SERVICE_PORT}:2181" \
        -p "2888:2888" \
        -p "3888:3888" \
        -v "${PROJ_PREFIX}-zookeeper-data:/var/lib/zookeeper/data" \
        -v "${PROJ_PREFIX}-zookeeper-logs:/var/lib/zookeeper/log" \
        -e ZOOKEEPER_CLIENT_PORT=2181 \
        -e ZOOKEEPER_TICK_TIME=2000 \
        -e ZOOKEEPER_INIT_LIMIT=10 \
        -e ZOOKEEPER_SYNC_LIMIT=5 \
        -e ZOOKEEPER_MAX_CLIENT_CNXNS=60 \
        -e ZOOKEEPER_SNAP_RETAIN_COUNT=3 \
        -e ZOOKEEPER_PURGE_INTERVAL=12 \
        -e PROJ_SERVICE_NAME=zookeeper \
        -e PROJ_SERVICE_VERSION=${ZOOKEEPER_VERSION:-3.8} \
        -e PROJ_ENVIRONMENT=${PROJ_ENVIRONMENT:-development} \
        --log-driver json-file \
        --log-opt max-size=10m \
        --log-opt max-file=3 \
        --health-cmd "echo ruok | nc localhost 2181 | grep imok" \
        --health-interval 30s \
        --health-timeout 10s \
        --health-retries 3 \
        --health-start-period 30s \
        "${IMAGE_NAME}"
fi

echo "Zookeeper容器已启动: ${CONTAINER_NAME}"
echo "端口映射: localhost:${SERVICE_PORT} -> container:2181"
if [[ "$(uname)" == "Linux" ]]; then
    echo "集群通信端口: 2888, 3888"
fi
echo "配置目录: ${CONFIG_DIR}"
echo "数据目录: ${DATA_DIR}"
echo "日志目录: ${LOG_DIR}"
echo "使用镜像: ${IMAGE_NAME}"

# 显示容器状态
docker ps --filter name="${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# 等待Zookeeper启动
echo "等待Zookeeper启动完成..."
timeout=60
while [ $timeout -gt 0 ]; do
    if docker exec "${CONTAINER_NAME}" sh -c "echo srvr | nc localhost 2181" 2>/dev/null | grep -q "Zookeeper version"; then
        echo "Zookeeper启动成功 ✅"
        break
    fi
    sleep 2
    ((timeout-=2))
done

if [ $timeout -le 0 ]; then
    echo "Zookeeper启动超时，请检查日志 ❌"
    docker logs "${CONTAINER_NAME}" --tail 20
    exit 1
fi

echo ""
echo "🎉 Zookeeper服务已成功启动!"
echo "🔗 客户端端点: localhost:${SERVICE_PORT}"
echo "📋 服务器信息:"
docker exec "${CONTAINER_NAME}" sh -c "echo srvr | nc localhost 2181" 2>/dev/null | head -5
echo ""
echo "🛠️  常用命令:"
echo "  服务器状态: docker exec ${CONTAINER_NAME} sh -c \"echo srvr | nc localhost 2181\""
echo "  客户端连接: docker exec -it ${CONTAINER_NAME} zookeeper-shell localhost:2181"
echo "  健康检查: docker exec ${CONTAINER_NAME} sh -c \"echo ruok | nc localhost 2181\""
echo "  配置信息: docker exec ${CONTAINER_NAME} sh -c \"echo conf | nc localhost 2181\""