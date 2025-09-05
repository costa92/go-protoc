#!/usr/bin/env bash

# Jaeger Docker运行脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: Jaeger ${JAEGER_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 加载工具函数和环境配置
# 先尝试直接找到项目根目录，然后加载util.sh
if [[ -f "${PROJ_ROOT_DIR}/scripts/lib/util.sh" ]]; then
    source "${PROJ_ROOT_DIR}/scripts/lib/util.sh"
elif [[ -f "$(pwd)/scripts/lib/util.sh" ]]; then
    source "$(pwd)/scripts/lib/util.sh"
else
    # 回退方案：定义需要的函数
    proj::util::is_mac() { [[ "$(uname)" == "Darwin" ]]; }
fi

# 加载环境配置
if [[ -f "${PROJ_ROOT_DIR}/manifests/env/env.dev" ]]; then
    source "${PROJ_ROOT_DIR}/manifests/env/env.dev"
elif [[ -f "$(pwd)/manifests/env/env.dev" ]]; then
    source "$(pwd)/manifests/env/env.dev"
fi

# 服务配置
readonly SERVICE_NAME="jaeger"
readonly CONTAINER_NAME="${PROJ_PREFIX}-jaeger"
readonly IMAGE_NAME="jaegertracing/all-in-one:${JAEGER_VERSION}"

# 端口配置
readonly JAEGER_UI_PORT="${PROJ_JAEGER_UI_PORT:-16686}"
readonly JAEGER_COLLECTOR_PORT="${PROJ_JAEGER_COLLECTOR_PORT:-14268}"
readonly JAEGER_AGENT_PORT="${PROJ_JAEGER_AGENT_PORT:-6831}"
readonly JAEGER_GRPC_PORT="${PROJ_JAEGER_GRPC_PORT:-14250}"

# 目录配置
readonly DATA_DIR="${PROJ_JAEGER_DATA_DIR}"

# Docker网络
readonly NETWORK_NAME="${PROJ_NETWORK_NAME}"

# 创建必要的目录
echo "创建Jaeger目录..."

# 使用与重构前一致的权限处理方式
proj::util::sudo "mkdir -p '${PROJ_JAEGER_DATA_DIR}'"

# 设置正确的目录权限
proj::util::sudo "chown -R $(whoami):$(id -gn) '${PROJ_JAEGER_DATA_DIR}'"

echo "✅ Jaeger目录创建完成"

# 创建Docker网络（如果不存在）
docker network create "${PROJ_NETWORK_NAME}" 2>/dev/null || true

# 创建Docker卷（如果不存在）
docker volume create "${CONTAINER_NAME}-data" 2>/dev/null || true

# 停止并删除现有容器（如果存在）
docker stop "${PROJ_PREFIX}-jaeger" 2>/dev/null || true
docker rm "${PROJ_PREFIX}-jaeger" 2>/dev/null || true

# 运行Jaeger容器
docker run -d \
    --name "${PROJ_PREFIX}-jaeger" \
    --network "${PROJ_NETWORK_NAME}" \
    --restart unless-stopped \
    -p "${PROJ_JAEGER_UI_PORT}:16686" \
    -p "${PROJ_JAEGER_COLLECTOR_PORT}:14268" \
    -p "${PROJ_JAEGER_AGENT_PORT}:6831/udp" \
    -p "${PROJ_JAEGER_GRPC_PORT}:14250" \
    -v "${CONTAINER_NAME}-data:/tmp" \
    -e COLLECTOR_ZIPKIN_HTTP_PORT=9411 \
    -e COLLECTOR_OTLP_ENABLED=true \
    -e PROJ_SERVICE_NAME=jaeger \
    -e PROJ_SERVICE_VERSION=${JAEGER_VERSION} \
    -e PROJ_ENVIRONMENT=${PROJ_ENVIRONMENT:-development} \
    --log-driver json-file \
    --log-opt max-size=10m \
    --log-opt max-file=3 \
    --health-cmd "wget -q --spider http://127.0.0.1:16686/ || exit 1" \
    --health-interval 30s \
    --health-timeout 10s \
    --health-retries 3 \
    --health-start-period 30s \
    "jaegertracing/all-in-one:${JAEGER_VERSION}"

echo "Jaeger容器已启动: ${PROJ_PREFIX}-jaeger"
echo "Web UI: http://localhost:${PROJ_JAEGER_UI_PORT}"
echo "Collector HTTP: http://localhost:${PROJ_JAEGER_COLLECTOR_PORT}"
echo "Agent UDP: localhost:${PROJ_JAEGER_AGENT_PORT}"
echo "gRPC: localhost:${PROJ_JAEGER_GRPC_PORT}"
echo "数据目录: ${PROJ_JAEGER_DATA_DIR}"

# 显示容器状态
docker ps --filter name="${PROJ_PREFIX}-jaeger" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"