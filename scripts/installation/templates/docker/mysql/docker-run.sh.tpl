#!/usr/bin/env bash

# MySQL Docker运行脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: MySQL ${MYSQL_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly SERVICE_NAME="mysql"
readonly CONTAINER_NAME="${CONTAINER_NAME_MYSQL}"
readonly IMAGE_NAME="mysql:${MYSQL_VERSION}"
readonly SERVICE_PORT="${PROJ_MYSQL_PORT:-3306}"

# 目录配置
readonly CONFIG_DIR="${PROJ_MYSQL_CONFIG_DIR}"
readonly DATA_DIR="${PROJ_MYSQL_DATA_DIR}"
readonly LOG_DIR="${DATA_DIR}/logs"

# Docker网络
readonly NETWORK_NAME="${PROJ_NETWORK_NAME}"

# MySQL配置
readonly MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-proj(#)666}"
readonly MYSQL_DATABASE="${MYSQL_DATABASE:-onex}"
readonly MYSQL_USER="${MYSQL_USER:-onex}"
readonly MYSQL_PASSWORD="${MYSQL_PASSWORD:-proj(#)666}"

# 创建必要的目录
mkdir -p "${CONFIG_DIR}" "${DATA_DIR}" "${LOG_DIR}"

# 创建Docker网络（如果不存在）
docker network create "${NETWORK_NAME}" 2>/dev/null || true

# 创建Docker卷（如果不存在）
docker volume create "${CONTAINER_NAME}-data" 2>/dev/null || true

# 停止并删除现有容器（如果存在）
docker stop "${CONTAINER_NAME}" 2>/dev/null || true
docker rm "${CONTAINER_NAME}" 2>/dev/null || true

# 运行MySQL容器
docker run -d \
    --name "${CONTAINER_NAME}" \
    --network "${NETWORK_NAME}" \
    --restart unless-stopped \
    -p "${SERVICE_PORT}:3306" \
    -v "${CONTAINER_NAME}-data:/var/lib/mysql" \
    -v "${CONFIG_DIR}:/etc/mysql/conf.d:ro" \
    -v "${LOG_DIR}:/var/log/mysql" \
    -e MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD}" \
    -e MYSQL_DATABASE="${MYSQL_DATABASE}" \
    -e MYSQL_USER="${MYSQL_USER}" \
    -e MYSQL_PASSWORD="${MYSQL_PASSWORD}" \
    -e PROJ_SERVICE_NAME=mysql \
    -e PROJ_SERVICE_VERSION=${MYSQL_VERSION} \
    -e PROJ_ENVIRONMENT=${PROJ_ENVIRONMENT:-development} \
    --log-driver json-file \
    --log-opt max-size=10m \
    --log-opt max-file=3 \
    --health-cmd "mysqladmin ping -h localhost -u root -p'${MYSQL_ROOT_PASSWORD}'" \
    --health-interval 30s \
    --health-timeout 10s \
    --health-retries 3 \
    --health-start-period 60s \
    "${IMAGE_NAME}"

echo "MySQL容器已启动: ${CONTAINER_NAME}"
echo "端口映射: localhost:${SERVICE_PORT} -> container:3306"
echo "数据库: ${MYSQL_DATABASE}"
echo "用户: ${MYSQL_USER}"
echo "Root密码: ${MYSQL_ROOT_PASSWORD}"
echo "配置目录: ${CONFIG_DIR}"
echo "数据目录: ${DATA_DIR}"

# 显示容器状态
docker ps --filter name="${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# 等待MySQL完全启动
echo "等待MySQL启动完成..."
timeout=60
while [ $timeout -gt 0 ]; do
    if docker exec "${CONTAINER_NAME}" mysqladmin ping -h localhost -u root -p"${MYSQL_ROOT_PASSWORD}" --silent >/dev/null 2>&1; then
        echo "MySQL启动成功 ✅"
        break
    fi
    sleep 2
    ((timeout-=2))
done

if [ $timeout -le 0 ]; then
    echo "MySQL启动超时，请检查日志 ❌"
    docker logs "${CONTAINER_NAME}" --tail 20
    exit 1
fi

# 显示MySQL版本信息
docker exec "${CONTAINER_NAME}" mysql -u root -p"${MYSQL_ROOT_PASSWORD}" -e "SELECT VERSION();" 2>/dev/null || true