#!/usr/bin/env bash

# OTEL Agent Docker运行脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: OpenTelemetry Agent ${OTELCOL_VERSION}
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
readonly SERVICE_NAME="otel-agent"
readonly CONTAINER_NAME="${CONTAINER_NAME_OTEL_AGENT}"
readonly IMAGE_NAME="otel/opentelemetry-collector-contrib:${OTELCOL_VERSION}"
readonly GRPC_PORT="${PROJ_OTEL_AGENT_GRPC_PORT:-4327}"
readonly HTTP_PORT="${PROJ_OTEL_AGENT_HTTP_PORT:-4328}"
readonly HEALTH_PORT="${PROJ_OTEL_AGENT_HEALTH_PORT:-13134}"

# 目录配置
readonly CONFIG_DIR="${PROJ_OTEL_AGENT_CONFIG_DIR}"
readonly DATA_DIR="${PROJ_OTEL_AGENT_DATA_DIR}"
readonly LOG_DIR="${DATA_DIR}/logs"

# Docker网络
readonly NETWORK_NAME="${PROJ_NETWORK_NAME}"

# Collector 依赖配置
readonly COLLECTOR_CONTAINER="${CONTAINER_NAME_OTEL_COLLECTOR}"
readonly COLLECTOR_ENDPOINT="${COLLECTOR_CONTAINER}:4317"

# 检查 Collector 是否运行
echo "检查OTEL Collector依赖..."
if ! docker ps --format '{{.Names}}' | grep -q "^${COLLECTOR_CONTAINER}$"; then
    echo "❌ OTEL Collector容器未运行: ${COLLECTOR_CONTAINER}"
    echo "请先启动 Collector: make docker.otel-collector.start"
    exit 1
fi

# 验证 Collector 健康状态
collector_health_port=$(docker port "${COLLECTOR_CONTAINER}" 13133/tcp 2>/dev/null | cut -d: -f2)
if [[ -n "${collector_health_port}" ]] && curl -s "http://127.0.0.1:${collector_health_port}" >/dev/null 2>&1; then
    echo "✅ OTEL Collector健康检查通过 (端口: ${collector_health_port})"
else
    echo "⚠️  OTEL Collector可能未就绪，但继续启动Agent"
fi

# 创建必要的目录
echo "创建OTEL Agent目录..."
mkdir -p "${CONFIG_DIR}" "${DATA_DIR}" "${LOG_DIR}"

# 创建Docker网络（如果不存在）
docker network create "${NETWORK_NAME}" 2>/dev/null || true

# 创建Docker卷（如果不存在）
docker volume create "${CONTAINER_NAME_OTEL_AGENT}-data" 2>/dev/null || true

# 停止并删除现有容器（如果存在）
docker stop "${CONTAINER_NAME}" 2>/dev/null || true
docker rm "${CONTAINER_NAME}" 2>/dev/null || true

# 配置文件已由template system生成
echo "使用已生成的OTEL Agent配置文件..."
# 配置文件在生成阶段已经完成变量替换
cp "$(dirname "${BASH_SOURCE[0]}")/config.yaml" "${CONFIG_DIR}/config.yaml"

# 运行OTEL Agent容器
echo "启动OTEL Agent容器..."
docker run -d \
    --name "${CONTAINER_NAME}" \
    --network "${NETWORK_NAME}" \
    --restart unless-stopped \
    -p "${GRPC_PORT}:4317" \
    -p "${HTTP_PORT}:4318" \
    -p "${HEALTH_PORT}:13133" \
    -v "${CONFIG_DIR}/config.yaml:/etc/otelcol-contrib/config.yaml:ro" \
    -v "${CONTAINER_NAME_OTEL_AGENT}-data:/data" \
    -v "${LOG_DIR}:/var/log/otelcol" \
    -e PROJ_SERVICE_NAME=${PROJ_SERVICE_NAME:-apiserver} \
    -e PROJ_SERVICE_VERSION=${PROJ_SERVICE_VERSION:-v2.0.0} \
    -e PROJ_ENVIRONMENT=${PROJ_ENVIRONMENT:-development} \
    -e OTEL_COLLECTOR_ENDPOINT=${COLLECTOR_ENDPOINT} \
    -e GOMEMLIMIT=256MiB \
    --log-driver json-file \
    --log-opt max-size=5m \
    --log-opt max-file=3 \
    "${IMAGE_NAME}" \
    --config=/etc/otelcol-contrib/config.yaml

echo "OTEL Agent容器已启动: ${CONTAINER_NAME}"
echo "应用连接端点:"
echo "  OTLP gRPC: localhost:${GRPC_PORT}"
echo "  OTLP HTTP: localhost:${HTTP_PORT}"
echo "  健康检查: localhost:${HEALTH_PORT}"
echo "转发目标: ${COLLECTOR_ENDPOINT}"
echo "配置文件: ${CONFIG_DIR}/config.yaml"
echo "数据目录: ${DATA_DIR}"

# 显示容器状态
docker ps --filter name="${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# 等待服务启动
echo "等待OTEL Agent启动..."
sleep 3

# 检查健康状态
if curl -s -f http://localhost:${HEALTH_PORT}/ >/dev/null; then
    echo "OTEL Agent健康检查通过 ✅"
    echo ""
    echo "🎉 OTEL Agent已就绪，可以接收应用程序的遥测数据"
    echo "数据流: 应用程序 → Agent(${GRPC_PORT}) → Collector(${COLLECTOR_CONTAINER}) → 后端存储"
else
    echo "OTEL Agent健康检查失败，请检查日志 ❌"
    docker logs "${CONTAINER_NAME}" --tail 20
fi