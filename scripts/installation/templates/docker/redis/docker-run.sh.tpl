#!/usr/bin/env bash

# Redis Docker运行脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: Redis ${REDIS_VERSION}
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

# 加载环境配置（包含LINUX_PASSWORD）
if [[ -f "${PROJ_ROOT_DIR}/manifests/env/env.dev" ]]; then
    source "${PROJ_ROOT_DIR}/manifests/env/env.dev"
elif [[ -f "$(pwd)/manifests/env/env.dev" ]]; then
    source "$(pwd)/manifests/env/env.dev"
fi

# 服务配置
readonly SERVICE_NAME="redis"
readonly CONTAINER_NAME="${PROJ_PREFIX}-redis"
readonly IMAGE_NAME="redis:${REDIS_VERSION}"
readonly SERVICE_PORT="${PROJ_REDIS_PORT:-6379}"

# 目录配置
readonly CONFIG_DIR="${PROJ_REDIS_CONFIG_DIR}"
readonly DATA_DIR="${PROJ_REDIS_DATA_DIR}"
readonly LOG_DIR="${PROJ_REDIS_LOG_DIR:-${PROJ_REDIS_DATA_DIR}/logs}"

# Docker网络
readonly NETWORK_NAME="${PROJ_NETWORK_NAME}"

# 创建必要的目录（使用原始proj::util::sudo方法）
echo "创建Redis目录..."

# 使用与重构前一致的权限处理方式
proj::util::sudo "mkdir -p '${PROJ_REDIS_CONFIG_DIR}'" 
proj::util::sudo "mkdir -p '${PROJ_REDIS_DATA_DIR}'"
proj::util::sudo "mkdir -p '${PROJ_REDIS_LOG_DIR}'"

# 设置正确的目录权限
proj::util::sudo "chown -R $(whoami):$(id -gn) '${PROJ_REDIS_CONFIG_DIR}'"
proj::util::sudo "chown -R $(whoami):$(id -gn) '${PROJ_REDIS_DATA_DIR}'"
proj::util::sudo "chown -R $(whoami):$(id -gn) '${PROJ_REDIS_LOG_DIR}'"

echo "✅ Redis目录创建完成"

# 复制配置文件到配置目录
echo "复制Redis配置文件..."
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CONFIG_FILE="$SCRIPT_DIR/redis.conf"

if [[ -f "$CONFIG_FILE" ]]; then
    # 使用与重构前一致的方法复制配置文件
    proj::util::sudo "cp '$CONFIG_FILE' '${PROJ_REDIS_CONFIG_DIR}/redis.conf'"
    proj::util::sudo "chown $(whoami):$(id -gn) '${PROJ_REDIS_CONFIG_DIR}/redis.conf'"
    echo "✅ Redis配置文件复制完成"
else
    echo "⚠️  未找到Redis配置文件: $CONFIG_FILE"
fi

# 创建Docker网络（如果不存在）
docker network create "${PROJ_NETWORK_NAME}" 2>/dev/null || true

# 创建Docker卷（如果不存在）
docker volume create "${CONTAINER_NAME_REDIS}-data" 2>/dev/null || true

# 停止并删除现有容器（如果存在）
docker stop "${PROJ_PREFIX}-redis" 2>/dev/null || true
docker rm "${PROJ_PREFIX}-redis" 2>/dev/null || true

# 运行Redis容器
docker run -d \
    --name "${PROJ_PREFIX}-redis" \
    --network "${PROJ_NETWORK_NAME}" \
    --restart unless-stopped \
    -p "${PROJ_REDIS_PORT:-6379}:6379" \
    -v "${CONTAINER_NAME_REDIS}-data:/data" \
    -v "${PROJ_REDIS_CONFIG_DIR}/redis.conf:/usr/local/etc/redis/redis.conf:ro" \
    -v "${PROJ_REDIS_LOG_DIR}:/var/log/redis" \
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
    "redis:${REDIS_VERSION}" \
    redis-server /usr/local/etc/redis/redis.conf

echo "Redis容器已启动: ${PROJ_PREFIX}-redis"
echo "端口映射: localhost:${PROJ_REDIS_PORT:-6379} -> container:6379"
echo "配置文件: ${PROJ_REDIS_CONFIG_DIR}/redis.conf"
echo "数据目录: ${PROJ_REDIS_DATA_DIR}"

# 显示容器状态
docker ps --filter name="${PROJ_PREFIX}-redis" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"