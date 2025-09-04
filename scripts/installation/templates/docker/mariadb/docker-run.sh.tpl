#!/usr/bin/env bash

# MariaDB Docker运行脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: MariaDB ${MARIADB_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly SERVICE_NAME="mariadb"
readonly CONTAINER_NAME="${CONTAINER_NAME_MARIADB}"
readonly IMAGE_NAME="mariadb:${MARIADB_VERSION}"
readonly SERVICE_PORT="${PROJ_MARIADB_PORT:-3307}"

# 目录配置
readonly CONFIG_DIR="${PROJ_MARIADB_CONFIG_DIR}"
readonly DATA_DIR="${PROJ_MARIADB_DATA_DIR}"
readonly LOG_DIR="${PROJ_MARIADB_LOG_DIR}"

# Docker网络
readonly NETWORK_NAME="${PROJ_NETWORK_NAME}"

# MariaDB配置
readonly MARIADB_ROOT_PASSWORD="${MARIADB_ROOT_PASSWORD:-proj(#)666}"
readonly MARIADB_DATABASE="${MARIADB_DATABASE:-onex}"
readonly MARIADB_USER="${MARIADB_USER:-onex}"
readonly MARIADB_PASSWORD="${MARIADB_PASSWORD:-proj(#)666}"

# 创建必要的目录
mkdir -p "${PROJ_MARIADB_CONFIG_DIR}" "${PROJ_MARIADB_DATA_DIR}" "${PROJ_MARIADB_LOG_DIR}"

# 创建Docker网络（如果不存在）
docker network create "${NETWORK_NAME}" 2>/dev/null || true

# 创建Docker卷（如果不存在）
docker volume create "${CONTAINER_NAME_MARIADB}-data" 2>/dev/null || true

# 停止并删除现有容器（如果存在）
docker stop "${CONTAINER_NAME}" 2>/dev/null || true
docker rm "${CONTAINER_NAME}" 2>/dev/null || true

# 运行MariaDB容器
docker run -d \
    --name "${CONTAINER_NAME}" \
    --network "${NETWORK_NAME}" \
    --restart unless-stopped \
    -p "${PROJ_MARIADB_PORT:-3307}:3306" \
    -v "${CONTAINER_NAME_MARIADB}-data:/var/lib/mysql" \
    -v "${PROJ_MARIADB_CONFIG_DIR}:/etc/mysql/conf.d:ro" \
    -v "${PROJ_MARIADB_LOG_DIR}:/var/log/mysql" \
    -e MARIADB_ROOT_PASSWORD="${MARIADB_ROOT_PASSWORD}" \
    -e MARIADB_DATABASE="${MARIADB_DATABASE}" \
    -e MARIADB_USER="${MARIADB_USER}" \
    -e MARIADB_PASSWORD="${MARIADB_PASSWORD}" \
    -e PROJ_SERVICE_NAME=mariadb \
    -e PROJ_SERVICE_VERSION=${MARIADB_VERSION} \
    -e PROJ_ENVIRONMENT=${PROJ_ENVIRONMENT:-development} \
    --log-driver json-file \
    --log-opt max-size=10m \
    --log-opt max-file=3 \
    --health-cmd "mariadb-admin ping -h localhost -u root -p${MARIADB_ROOT_PASSWORD}" \
    --health-interval 30s \
    --health-timeout 10s \
    --health-retries 3 \
    --health-start-period 60s \
    "mariadb:${MARIADB_VERSION}"

echo "MariaDB容器已启动: ${CONTAINER_NAME}"
echo "端口映射: localhost:${PROJ_MARIADB_PORT:-3307} -> container:3306"
echo "数据库: ${MARIADB_DATABASE}"
echo "用户: ${MARIADB_USER}"
echo "Root密码: ${MARIADB_ROOT_PASSWORD}"
echo "配置目录: ${PROJ_MARIADB_CONFIG_DIR}"
echo "数据目录: ${PROJ_MARIADB_DATA_DIR}"

# 显示容器状态
docker ps --filter name="${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# 等待MariaDB完全启动
echo "等待MariaDB启动完成..."
timeout=60
while [ $timeout -gt 0 ]; do
    if docker exec "${CONTAINER_NAME}" mariadb-admin ping -h localhost -u root -p"${MARIADB_ROOT_PASSWORD}" --silent >/dev/null 2>&1; then
        echo "MariaDB启动成功 ✅"
        break
    fi
    sleep 2
    ((timeout-=2))
done

if [ $timeout -le 0 ]; then
    echo "MariaDB启动超时，请检查日志 ❌"
    docker logs "${CONTAINER_NAME}" --tail 20
    exit 1
fi

# 显示MariaDB版本信息
docker exec "${CONTAINER_NAME}" mariadb -u root -p"${MARIADB_ROOT_PASSWORD}" -e "SELECT VERSION();" 2>/dev/null || true