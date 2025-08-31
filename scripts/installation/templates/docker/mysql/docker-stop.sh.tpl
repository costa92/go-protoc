#!/usr/bin/env bash

# MySQL Docker停止脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: MySQL ${MYSQL_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly CONTAINER_NAME="proj-mysql"
readonly VOLUME_NAME="proj-mysql-data"

echo "正在停止MySQL容器..."

# 停止容器
if docker stop "${CONTAINER_NAME}" 2>/dev/null; then
    echo "MySQL容器已停止: ${CONTAINER_NAME}"
else
    echo "MySQL容器未运行或已停止"
fi

# 删除容器
if docker rm "${CONTAINER_NAME}" 2>/dev/null; then
    echo "MySQL容器已删除: ${CONTAINER_NAME}"
else
    echo "MySQL容器不存在或已删除"
fi

echo "MySQL服务已完全停止"

# 可选：删除数据卷（谨慎使用）
if [[ "${1:-}" == "--remove-data" ]]; then
    echo "⚠️  警告：将删除MySQL数据卷！这将永久删除所有数据库数据！"
    read -p "确认删除数据卷？(y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker volume rm "${VOLUME_NAME}" 2>/dev/null || true
        echo "MySQL数据卷已删除: ${VOLUME_NAME}"
        echo "所有数据库数据已被永久删除"
    else
        echo "数据卷保留: ${VOLUME_NAME}"
    fi
fi