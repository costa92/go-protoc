#!/usr/bin/env bash

# OTEL Collector Docker运行脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: OpenTelemetry Collector ${OTELCOL_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 加载工具函数和环境配置
if [[ -f "${PROJ_ROOT_DIR}/scripts/lib/util.sh" ]]; then
    source "${PROJ_ROOT_DIR}/scripts/lib/util.sh"
elif [[ -f "$(pwd)/scripts/lib/util.sh" ]]; then
    source "$(pwd)/scripts/lib/util.sh"
else
    proj::util::is_mac() { [[ "$(uname)" == "Darwin" ]]; }
fi

# 加载环境配置
if [[ -f "${PROJ_ROOT_DIR}/manifests/env/env.dev" ]]; then
    PROJ_ENVIRONMENT=${PROJ_ENVIRONMENT:-development}
    source "${PROJ_ROOT_DIR}/manifests/env/env.base"
elif [[ -f "$(pwd)/manifests/env/env.dev" ]]; then
    PROJ_ENVIRONMENT=${PROJ_ENVIRONMENT:-development}
    source "$(pwd)/manifests/env/env.base"
fi

# 服务配置
readonly SERVICE_NAME="otel-collector"
readonly CONTAINER_NAME="${CONTAINER_NAME_OTEL_COLLECTOR}"
readonly IMAGE_NAME="otel/opentelemetry-collector-contrib:${OTELCOL_VERSION}"
readonly GRPC_PORT="${PROJ_OTELCOL_GRPC_PORT:-4317}"
readonly HTTP_PORT="${PROJ_OTELCOL_HTTP_PORT:-4318}"
readonly METRICS_PORT="${PROJ_OTELCOL_METRICS_PORT:-8888}"
readonly HEALTH_PORT="${PROJ_OTELCOL_HEALTH_PORT:-13133}"

# 目录配置
readonly CONFIG_DIR="${PROJ_OTELCOL_CONFIG_DIR}"
readonly DATA_DIR="${PROJ_OTELCOL_DATA_DIR}"
readonly LOG_DIR="${DATA_DIR}/logs"

# Docker网络
readonly NETWORK_NAME="${PROJ_NETWORK_NAME}"

# 创建必要的目录
echo "创建OTEL Collector目录..."
mkdir -p "${CONFIG_DIR}" "${DATA_DIR}" "${LOG_DIR}"

# 创建Docker网络（如果不存在）
docker network create "${NETWORK_NAME}" 2>/dev/null || true

# 创建Docker卷（如果不存在）
docker volume create "${CONTAINER_NAME_OTEL_COLLECTOR}-data" 2>/dev/null || true

# 停止并删除现有容器（如果存在）
docker stop "${CONTAINER_NAME}" 2>/dev/null || true
docker rm "${CONTAINER_NAME}" 2>/dev/null || true

# 配置文件已由template system生成
echo "使用已生成的OTEL Collector配置文件..."
# 配置文件在生成阶段已经完成变量替换
cp "$(dirname "${BASH_SOURCE[0]}")/config.yaml" "${CONFIG_DIR}/config.yaml"

# 运行OTEL Collector容器
echo "启动OTEL Collector容器..."
docker run -d \
    --name "${CONTAINER_NAME}" \
    --network "${NETWORK_NAME}" \
    --restart unless-stopped \
    -p "${GRPC_PORT}:4317" \
    -p "${HTTP_PORT}:4318" \
    -p "${METRICS_PORT}:8888" \
    -p "${HEALTH_PORT}:13133" \
    -v "${CONFIG_DIR}/config.yaml:/etc/otelcol-contrib/config.yaml:ro" \
    -v "${CONTAINER_NAME_OTEL_COLLECTOR}-data:/data" \
    -v "${PROJ_ROOT_DIR}/logs:/host/logs:ro" \
    -v "${LOG_DIR}:/var/log/otelcol" \
    -e PROJ_SERVICE_NAME=${PROJ_SERVICE_NAME:-apiserver} \
    -e PROJ_SERVICE_VERSION=${PROJ_SERVICE_VERSION:-v2.0.0} \
    -e PROJ_ENVIRONMENT=${PROJ_ENVIRONMENT:-development} \
    -e VICTORIALOGS_ENDPOINT=${VICTORIALOGS_ENDPOINT} \
    -e JAEGER_ENDPOINT=${JAEGER_ENDPOINT} \
    -e PROMETHEUS_ENDPOINT=${PROMETHEUS_ENDPOINT} \
    -e GOMEMLIMIT=512MiB \
    --log-driver json-file \
    --log-opt max-size=10m \
    --log-opt max-file=3 \
    "${IMAGE_NAME}" \
    --config=/etc/otelcol-contrib/config.yaml

echo "OTEL Collector容器已启动: ${CONTAINER_NAME}"
echo "OTLP gRPC端口: localhost:${GRPC_PORT} -> container:4317"
echo "OTLP HTTP端口: localhost:${HTTP_PORT} -> container:4318"
echo "Prometheus指标: localhost:${METRICS_PORT} -> container:8888"
echo "健康检查: localhost:${HEALTH_PORT} -> container:13133"
echo "配置文件: ${CONFIG_DIR}/config.yaml"
echo "数据目录: ${DATA_DIR}"

# 显示容器状态
docker ps --filter name="${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# 等待服务启动
echo "等待OTEL Collector启动..."
sleep 5

# 检查健康状态
if curl -s -f http://localhost:${HEALTH_PORT}/ >/dev/null; then
    echo "OTEL Collector健康检查通过 ✅"
else
    echo "OTEL Collector健康检查失败，请检查日志 ❌"
    docker logs "${CONTAINER_NAME}" --tail 20
fi