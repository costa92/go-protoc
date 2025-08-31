#!/usr/bin/env bash

# Redis Docker运行脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: Redis ${REDIS_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly SERVICE_NAME="redis"
readonly CONTAINER_NAME="proj-redis"
readonly IMAGE_NAME="redis:${REDIS_VERSION}"
readonly SERVICE_PORT="${PROJ_REDIS_PORT:-6379}"

# 目录配置
readonly CONFIG_DIR="${PROJ_REDIS_CONFIG_DIR}"
readonly DATA_DIR="${PROJ_REDIS_DATA_DIR}"
readonly LOG_DIR="${PROJ_REDIS_LOG_DIR:-${DATA_DIR}/logs}"

# Docker网络
readonly NETWORK_NAME="proj-network"

# 创建必要的目录
mkdir -p "${CONFIG_DIR}" "${DATA_DIR}" "${LOG_DIR}"

# 创建Docker网络（如果不存在）
docker network create "${NETWORK_NAME}" 2>/dev/null || true

# 创建Docker卷（如果不存在）
docker volume create proj-redis-data 2>/dev/null || true

# 停止并删除现有容器（如果存在）
docker stop "${CONTAINER_NAME}" 2>/dev/null || true
docker rm "${CONTAINER_NAME}" 2>/dev/null || true

# 运行Redis容器
docker run -d \
    --name "${CONTAINER_NAME}" \
    --network "${NETWORK_NAME}" \
    --restart unless-stopped \
    -p "${SERVICE_PORT}:6379" \
    -v proj-redis-data:/data \
    -v "${CONFIG_DIR}/redis.conf:/usr/local/etc/redis/redis.conf:ro" \
    -v "${LOG_DIR}:/var/log/redis" \
    -e REDIS_REPLICATION_MODE=master \
    -e PROJ_SERVICE_NAME=redis \
    -e PROJ_SERVICE_VERSION=${REDIS_VERSION} \
    -e PROJ_ENVIRONMENT=${PROJ_ENVIRONMENT:-development} \
    --log-driver json-file \
    --log-opt max-size=10m \
    --log-opt max-file=3 \
    --health-cmd "redis-cli ping" \
    --health-interval 30s \
    --health-timeout 10s \
    --health-retries 3 \
    --health-start-period 30s \
    "${IMAGE_NAME}" \
    redis-server /usr/local/etc/redis/redis.conf

echo "Redis容器已启动: ${CONTAINER_NAME}"
echo "端口映射: localhost:${SERVICE_PORT} -> container:6379"
echo "配置文件: ${CONFIG_DIR}/redis.conf"
echo "数据目录: ${DATA_DIR}"

# 显示容器状态
docker ps --filter name="${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"