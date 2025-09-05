#!/usr/bin/env bash

# VictoriaLogs Docker运行脚本模板
# Project: ${PROJ_NAME}
# Service: VictoriaLogs ${VICTORIALOGS_VERSION}
# Environment: ${PROJ_ENVIRONMENT}

set -eEuo pipefail

# 服务配置
readonly SERVICE_NAME="victorialogs"
readonly CONTAINER_NAME="${PROJ_PREFIX}-victorialogs"
readonly IMAGE_NAME="victoriametrics/victoria-logs:v${VICTORIALOGS_VERSION}"
readonly SERVICE_PORT="${PROJ_VICTORIALOGS_PORT}"

# 目录配置
readonly CONFIG_DIR="${PROJ_VICTORIALOGS_CONFIG_DIR}"
readonly DATA_DIR="${PROJ_VICTORIALOGS_DATA_DIR}"
readonly LOG_DIR="${PROJ_VICTORIALOGS_LOG_DIR}"

# Docker网络
readonly NETWORK_NAME="${PROJ_NETWORK_NAME}"

# 创建必要的目录
mkdir -p "${CONFIG_DIR}" "${DATA_DIR}" "${LOG_DIR}"

# 创建Docker网络（如果不存在）
docker network create "${NETWORK_NAME}" 2>/dev/null || true

# 创建Docker卷（如果不存在）
docker volume create "${CONTAINER_NAME_VICTORIALOGS}-data" 2>/dev/null || true

# 停止并删除现有容器（如果存在）
docker stop "${CONTAINER_NAME}" 2>/dev/null || true
docker rm "${CONTAINER_NAME}" 2>/dev/null || true

# 运行VictoriaLogs容器
docker run -d \
    --name "${CONTAINER_NAME}" \
    --network "${NETWORK_NAME}" \
    --restart unless-stopped \
    -p "${SERVICE_PORT}:9428" \
    -v "${CONTAINER_NAME_VICTORIALOGS}-data":/victoria-logs-data \
    -v "${CONFIG_DIR}:/etc/victorialogs:ro" \
    -v "${LOG_DIR}:/var/log/victorialogs" \
    -e GOMEMLIMIT=1GiB \
    -e PROJ_SERVICE_NAME=victorialogs \
    -e PROJ_SERVICE_VERSION=${VICTORIALOGS_VERSION} \
    -e PROJ_ENVIRONMENT=${PROJ_ENVIRONMENT} \
    --log-driver json-file \
    --log-opt max-size=10m \
    --log-opt max-file=3 \
    --health-cmd "wget --no-verbose --tries=1 --spider http://127.0.0.1:9428/health || exit 1" \
    --health-interval 30s \
    --health-timeout 10s \
    --health-retries 3 \
    --health-start-period 30s \
    "${IMAGE_NAME}" \
    -httpListenAddr=0.0.0.0:9428 \
    -storageDataPath=/victoria-logs-data \
    -loggerLevel=INFO \
    -retentionPeriod=${VICTORIALOGS_RETENTION}

echo "VictoriaLogs容器已启动: ${CONTAINER_NAME}"
echo "Web UI: http://localhost:${SERVICE_PORT}/select/vmui/"
echo "API端点: http://localhost:${SERVICE_PORT}/"
echo "数据保留期: ${VICTORIALOGS_RETENTION}"
echo "配置目录: ${CONFIG_DIR}"
echo "数据目录: ${DATA_DIR}"

# 显示容器状态
docker ps --filter name="${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# 等待服务启动
echo "等待服务启动..."
sleep 10

# 检查健康状态
if curl -s -f "http://localhost:${SERVICE_PORT}/health" >/dev/null; then
    echo "VictoriaLogs健康检查通过 ✅"
    echo "访问Web界面: http://localhost:${SERVICE_PORT}/select/vmui/"
else
    echo "VictoriaLogs健康检查失败，请检查日志 ❌"
    docker logs "${CONTAINER_NAME}" --tail 20
fi