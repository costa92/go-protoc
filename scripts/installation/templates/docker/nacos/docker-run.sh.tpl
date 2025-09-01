#!/usr/bin/env bash

# Nacos Docker运行脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: Nacos ${NACOS_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly SERVICE_NAME="nacos"
readonly CONTAINER_NAME="${PROJ_PREFIX}-nacos"
readonly IMAGE_NAME="nacos/nacos-server:${NACOS_VERSION}"
readonly HTTP_PORT="${PROJ_NACOS_PORT:-8848}"
readonly GRPC_PORT="${PROJ_NACOS_GRPC_PORT:-9848}"

# 目录配置
readonly CONFIG_DIR="${PROJ_NACOS_CONFIG_DIR}"
readonly DATA_DIR="${PROJ_NACOS_DATA_DIR}"
readonly LOG_DIR="${PROJ_NACOS_LOG_DIR}"

# Docker网络
readonly NETWORK_NAME="${PROJ_NETWORK_NAME}"

# Nacos配置
readonly NACOS_SERVER_PORT="${HTTP_PORT}"
readonly NACOS_APPLICATION_PORT="${GRPC_PORT}"
readonly MODE="${NACOS_MODE:-standalone}"  # standalone | cluster
readonly NACOS_AUTH_ENABLE="${NACOS_AUTH_ENABLE:-false}"
readonly NACOS_AUTH_TOKEN="${NACOS_AUTH_TOKEN:-SecretKey012345678901234567890123456789012345678901234567890123456789}"
readonly NACOS_AUTH_IDENTITY_KEY="${NACOS_AUTH_IDENTITY_KEY:-serverIdentity}"
readonly NACOS_AUTH_IDENTITY_VALUE="${NACOS_AUTH_IDENTITY_VALUE:-security}"

# 创建必要的目录
mkdir -p "${CONFIG_DIR}" "${DATA_DIR}" "${LOG_DIR}"

# 创建Docker网络（如果不存在）
docker network create "${NETWORK_NAME}" 2>/dev/null || true

# 创建Docker卷（如果不存在）
docker volume create "${PROJ_PREFIX}-nacos-data" 2>/dev/null || true
docker volume create "${PROJ_PREFIX}-nacos-logs" 2>/dev/null || true

# 停止并删除现有容器（如果存在）
docker stop "${CONTAINER_NAME}" 2>/dev/null || true
docker rm "${CONTAINER_NAME}" 2>/dev/null || true

# 运行Nacos容器
docker run -d \
    --name "${CONTAINER_NAME}" \
    --network "${NETWORK_NAME}" \
    --restart unless-stopped \
    -p "${HTTP_PORT}:8848" \
    -p "${GRPC_PORT}:9848" \
    -p "9849:9849" \
    -v "${PROJ_PREFIX}-nacos-data:/home/nacos/data" \
    -v "${PROJ_PREFIX}-nacos-logs:/home/nacos/logs" \
    -v "${CONFIG_DIR}:/home/nacos/conf:ro" \
    -v "${LOG_DIR}:/var/log/nacos" \
    -e MODE="${MODE}" \
    -e NACOS_SERVER_PORT="${NACOS_SERVER_PORT}" \
    -e NACOS_APPLICATION_PORT="${NACOS_APPLICATION_PORT}" \
    -e SPRING_DATASOURCE_PLATFORM=embedded \
    -e NACOS_AUTH_ENABLE="${NACOS_AUTH_ENABLE}" \
    -e NACOS_AUTH_TOKEN="${NACOS_AUTH_TOKEN}" \
    -e NACOS_AUTH_IDENTITY_KEY="${NACOS_AUTH_IDENTITY_KEY}" \
    -e NACOS_AUTH_IDENTITY_VALUE="${NACOS_AUTH_IDENTITY_VALUE}" \
    -e NACOS_AUTH_CACHE_ENABLE=false \
    -e PROJ_SERVICE_NAME=nacos \
    -e PROJ_SERVICE_VERSION=${NACOS_VERSION} \
    -e PROJ_ENVIRONMENT=${PROJ_ENVIRONMENT:-development} \
    --log-driver json-file \
    --log-opt max-size=10m \
    --log-opt max-file=3 \
    --health-cmd "curl -f http://localhost:8848/nacos/actuator/health || exit 1" \
    --health-interval 30s \
    --health-timeout 10s \
    --health-retries 5 \
    --health-start-period 60s \
    "${IMAGE_NAME}"

echo "Nacos容器已启动: ${CONTAINER_NAME}"
echo "Console UI: http://localhost:${HTTP_PORT}/nacos/"
echo "HTTP端口: localhost:${HTTP_PORT} -> container:8848"
echo "gRPC端口: localhost:${GRPC_PORT} -> container:9848"
echo "运行模式: ${MODE}"
echo "认证状态: ${NACOS_AUTH_ENABLE}"
echo "配置目录: ${CONFIG_DIR}"
echo "数据目录: ${DATA_DIR}"
echo "日志目录: ${LOG_DIR}"

# 显示容器状态
docker ps --filter name="${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# 等待Nacos完全启动
echo "等待Nacos启动完成..."
timeout=120
while [ $timeout -gt 0 ]; do
    if curl -s -f "http://localhost:${HTTP_PORT}/nacos/actuator/health" >/dev/null 2>&1; then
        echo "Nacos启动成功 ✅"
        break
    fi
    sleep 3
    ((timeout-=3))
done

if [ $timeout -le 0 ]; then
    echo "Nacos启动超时，请检查日志 ❌"
    docker logs "${CONTAINER_NAME}" --tail 20
    exit 1
fi

echo ""
echo "🎉 Nacos服务已成功启动!"
echo "📱 Web Console: http://localhost:${HTTP_PORT}/nacos/"
echo "👤 默认用户名/密码: nacos/nacos"
echo "📚 API Base URL: http://localhost:${HTTP_PORT}/nacos/v1/"

# 显示服务基本信息
echo ""
echo "服务信息:"
curl -s "http://localhost:${HTTP_PORT}/nacos/actuator/health" | head -5 2>/dev/null || echo "无法获取服务状态"