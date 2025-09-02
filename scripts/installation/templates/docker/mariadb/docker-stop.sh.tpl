#!/usr/bin/env bash

# MariaDB Docker停止脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: MariaDB ${MARIADB_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly SERVICE_NAME="mariadb"
readonly CONTAINER_NAME="proj-mariadb"

echo "正在停止MariaDB容器: ${CONTAINER_NAME}"

# 检查容器是否存在
if ! docker ps -a --filter name="^${CONTAINER_NAME}$" --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "容器 ${CONTAINER_NAME} 不存在"
    exit 0
fi

# 获取容器状态
CONTAINER_STATUS=$(docker ps --filter name="^${CONTAINER_NAME}$" --format '{{.Status}}' 2>/dev/null || echo "")

if [[ -n "$CONTAINER_STATUS" ]]; then
    echo "发现运行中的容器，状态: $CONTAINER_STATUS"
    
    # 优雅停止容器
    echo "正在优雅停止容器..."
    docker stop "${CONTAINER_NAME}" --time 30
    
    if [ $? -eq 0 ]; then
        echo "容器已优雅停止 ✅"
    else
        echo "优雅停止失败，强制停止容器..."
        docker kill "${CONTAINER_NAME}"
    fi
else
    echo "容器 ${CONTAINER_NAME} 未在运行"
fi

# 删除容器
echo "正在删除容器..."
docker rm "${CONTAINER_NAME}" 2>/dev/null || true

echo "MariaDB容器已停止并删除: ${CONTAINER_NAME}"

# 显示剩余的相关容器（如果有）
echo ""
echo "检查是否还有其他MariaDB相关容器："
docker ps -a --filter name="mariadb" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" || echo "未找到MariaDB相关容器"

# 可选：是否保留数据卷的提示
echo ""
echo "注意: 数据卷 'proj-mariadb-data' 已保留，如需完全清理请手动删除："
echo "  docker volume rm proj-mariadb-data"