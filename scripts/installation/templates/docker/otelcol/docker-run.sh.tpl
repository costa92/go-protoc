#!/usr/bin/env bash

# OpenTelemetry Collector Docker运行脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: OpenTelemetry Collector ${OTELCOL_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly SERVICE_NAME="otelcol"
readonly CONTAINER_NAME="proj-otelcol"
readonly IMAGE_NAME="otel/opentelemetry-collector-contrib:${OTELCOL_VERSION}"
readonly GRPC_PORT="${PROJ_OTELCOL_GRPC_PORT:-4327}"
readonly HTTP_PORT="${PROJ_OTELCOL_HTTP_PORT:-4328}"

# 目录配置
readonly CONFIG_DIR="${PROJ_OTELCOL_CONFIG_DIR}"
readonly DATA_DIR="${PROJ_OTELCOL_DATA_DIR}"
readonly LOG_DIR="${DATA_DIR}/logs"

# Docker网络
readonly NETWORK_NAME="proj-network"

# 创建必要的目录
mkdir -p "${CONFIG_DIR}" "${DATA_DIR}" "${LOG_DIR}"

# 创建Docker网络（如果不存在）
docker network create "${NETWORK_NAME}" 2>/dev/null || true

# 创建Docker卷（如果不存在）
docker volume create proj-otelcol-data 2>/dev/null || true

# 停止并删除现有容器（如果存在）
docker stop "${CONTAINER_NAME}" 2>/dev/null || true
docker rm "${CONTAINER_NAME}" 2>/dev/null || true

# 运行OpenTelemetry Collector容器
docker run -d \
    --name "${CONTAINER_NAME}" \
    --network "${NETWORK_NAME}" \
    --restart unless-stopped \
    -p "${GRPC_PORT}:4317" \
    -p "${HTTP_PORT}:4318" \
    -p "8888:8888" \
    -p "13133:13133" \
    -v "${CONFIG_DIR}/config.yaml:/etc/otelcol-contrib/config.yaml:ro" \
    -v proj-otelcol-data:/data \
    -v "${PROJ_ROOT_DIR}/logs:/host/logs:ro" \
    -v "${LOG_DIR}:/var/log/otelcol" \
    -e PROJ_SERVICE_NAME=${PROJ_SERVICE_NAME:-apiserver} \
    -e PROJ_SERVICE_VERSION=${PROJ_SERVICE_VERSION:-v2.0.0} \
    -e PROJ_ENVIRONMENT=${PROJ_ENVIRONMENT:-development} \
    -e GOMEMLIMIT=512m \
    --log-driver json-file \
    --log-opt max-size=10m \
    --log-opt max-file=3 \
    --health-cmd "curl -f http://localhost:13133/ || exit 1" \
    --health-interval 30s \
    --health-timeout 10s \
    --health-retries 3 \
    --health-start-period 30s \
    "${IMAGE_NAME}" \
    --config=/etc/otelcol-contrib/config.yaml

echo "OpenTelemetry Collector容器已启动: ${CONTAINER_NAME}"
echo "OTLP gRPC端口: localhost:${GRPC_PORT} -> container:4317"
echo "OTLP HTTP端口: localhost:${HTTP_PORT} -> container:4318"
echo "Prometheus指标: localhost:8888 -> container:8888"
echo "健康检查: localhost:13133 -> container:13133"
echo "配置文件: ${CONFIG_DIR}/config.yaml"
echo "数据目录: ${DATA_DIR}"

# 显示容器状态
docker ps --filter name="${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# 等待服务启动
echo "等待服务启动..."
sleep 5

# 检查健康状态
if curl -s -f http://localhost:13133/ >/dev/null; then
    echo "OpenTelemetry Collector健康检查通过 ✅"
else
    echo "OpenTelemetry Collector健康检查失败，请检查日志 ❌"
    docker logs "${CONTAINER_NAME}" --tail 20
fi