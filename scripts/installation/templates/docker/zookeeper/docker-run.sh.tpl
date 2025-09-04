#!/usr/bin/env bash

# Zookeeper Docker运行脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: Zookeeper ${ZOOKEEPER_VERSION}
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
    # macOS 使用 confluentinc 镜像，支持 ARM64
    readonly IMAGE_NAME="confluentinc/cp-zookeeper:${ZOOKEEPER_VERSION:-7.4.0}"
    readonly ZOO_MY_ID=1
    readonly ZOO_SERVERS="server.1=0.0.0.0:2888:3888;2181"
else
    # Linux 使用官方镜像
    readonly IMAGE_NAME="zookeeper:${ZOOKEEPER_VERSION:-3.8}"
    readonly ZOO_MY_ID=1
    readonly ZOO_SERVERS="server.1=0.0.0.0:2888:3888;2181"
fi

echo "检测到平台: $(uname)"
echo "使用镜像: ${IMAGE_NAME}"

# 创建必要的目录
mkdir -p "${CONFIG_DIR}" "${DATA_DIR}" "${LOG_DIR}"

# 创建Docker网络（如果不存在）
docker network create "${NETWORK_NAME}" 2>/dev/null || true

# 创建Docker卷（如果不存在）
docker volume create "${CONTAINER_NAME_ZOOKEEPER}-data" 2>/dev/null || true
docker volume create "${CONTAINER_NAME_ZOOKEEPER}-logs" 2>/dev/null || true

# 停止并删除现有容器（如果存在）
docker stop "${CONTAINER_NAME}" 2>/dev/null || true
docker rm "${CONTAINER_NAME}" 2>/dev/null || true

# 根据平台选择不同的运行配置
if [[ "$(uname)" == "Darwin" ]]; then
    # macOS - 使用 Confluent 镜像
    docker run -d \
        --name "${CONTAINER_NAME}" \
        --network "${NETWORK_NAME}" \
        --restart unless-stopped \
        -p "${SERVICE_PORT}:2181" \
        -p "2888:2888" \
        -p "3888:3888" \
        -v "${CONTAINER_NAME_ZOOKEEPER}-data:/var/lib/zookeeper/data" \
        -v "${CONTAINER_NAME_ZOOKEEPER}-logs:/var/lib/zookeeper/log" \
        -v "${LOG_DIR}:/var/log/zookeeper" \
        -e ZOOKEEPER_CLIENT_PORT=2181 \
        -e ZOOKEEPER_TICK_TIME=2000 \
        -e ZOOKEEPER_INIT_LIMIT=5 \
        -e ZOOKEEPER_SYNC_LIMIT=2 \
        -e ZOOKEEPER_MAX_CLIENT_CNXNS=60 \
        -e ZOOKEEPER_AUTOPURGE_SNAP_RETAIN_COUNT=3 \
        -e ZOOKEEPER_AUTOPURGE_PURGE_INTERVAL=24 \
        -e ZOOKEEPER_4LW_COMMANDS_WHITELIST="srvr,ruok,conf,isro" \
        -e ZOOKEEPER_SERVER_ID=${ZOO_MY_ID} \
        -e PROJ_SERVICE_NAME=zookeeper \
        -e PROJ_SERVICE_VERSION=${ZOOKEEPER_VERSION:-7.4.0} \
        -e PROJ_ENVIRONMENT=${PROJ_ENVIRONMENT:-development} \
        --log-driver json-file \
        --log-opt max-size=10m \
        --log-opt max-file=3 \
        --health-cmd "echo ruok | nc localhost 2181" \
        --health-interval 30s \
        --health-timeout 10s \
        --health-retries 3 \
        --health-start-period 30s \
        "${IMAGE_NAME}"
else
    # Linux - 使用官方镜像
    docker run -d \
        --name "${CONTAINER_NAME}" \
        --network "${NETWORK_NAME}" \
        --restart unless-stopped \
        -p "${SERVICE_PORT}:2181" \
        -p "2888:2888" \
        -p "3888:3888" \
        -v "${CONTAINER_NAME_ZOOKEEPER}-data:/data" \
        -v "${CONTAINER_NAME_ZOOKEEPER}-logs:/datalog" \
        -v "${LOG_DIR}:/logs" \
        -e ZOO_MY_ID=${ZOO_MY_ID} \
        -e ZOO_SERVERS="${ZOO_SERVERS}" \
        -e ZOO_TICK_TIME=2000 \
        -e ZOO_INIT_LIMIT=5 \
        -e ZOO_SYNC_LIMIT=2 \
        -e ZOO_MAX_CLIENT_CNXNS=60 \
        -e ZOO_AUTOPURGE_SNAPRETAINCOUNT=3 \
        -e ZOO_AUTOPURGE_PURGEINTERVAL=24 \
        -e PROJ_SERVICE_NAME=zookeeper \
        -e PROJ_SERVICE_VERSION=${ZOOKEEPER_VERSION:-3.8} \
        -e PROJ_ENVIRONMENT=${PROJ_ENVIRONMENT:-development} \
        --log-driver json-file \
        --log-opt max-size=10m \
        --log-opt max-file=3 \
        --health-cmd "echo ruok | nc localhost 2181" \
        --health-interval 30s \
        --health-timeout 10s \
        --health-retries 3 \
        --health-start-period 30s \
        "${IMAGE_NAME}"
fi

echo "Zookeeper容器已启动: ${CONTAINER_NAME}"
echo "端口映射: localhost:${SERVICE_PORT} -> container:2181"
echo "管理端口: 2888 (follower连接leader), 3888 (leader选举)"
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
    if echo ruok | nc localhost "${SERVICE_PORT}" 2>/dev/null | grep -q imok; then
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
echo "🔗 连接地址: localhost:${SERVICE_PORT}"
echo "📊 状态检查: echo ruok | nc localhost ${SERVICE_PORT}"
echo "📋 集群模式: 单机模式 (server.1)"