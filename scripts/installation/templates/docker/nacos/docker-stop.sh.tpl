#!/usr/bin/env bash

# Nacos Docker停止脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: Nacos ${NACOS_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly CONTAINER_NAME="proj-nacos"
readonly DATA_VOLUME_NAME="proj-nacos-data"
readonly LOGS_VOLUME_NAME="proj-nacos-logs"

echo "正在停止Nacos容器..."

# 停止容器
if docker stop "${CONTAINER_NAME}" 2>/dev/null; then
    echo "Nacos容器已停止: ${CONTAINER_NAME}"
else
    echo "Nacos容器未运行或已停止"
fi

# 删除容器
if docker rm "${CONTAINER_NAME}" 2>/dev/null; then
    echo "Nacos容器已删除: ${CONTAINER_NAME}"
else
    echo "Nacos容器不存在或已删除"
fi

echo "Nacos服务已完全停止"

# 可选：删除数据卷（谨慎使用）
if [[ "${1:-}" == "--remove-data" ]]; then
    echo "⚠️  警告：将删除Nacos数据卷！这将永久删除所有配置和数据！"
    echo "包括："
    echo "  - 服务注册信息"
    echo "  - 配置管理数据"  
    echo "  - 命名空间配置"
    echo "  - 用户和权限数据"
    echo ""
    read -p "确认删除数据卷？(y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker volume rm "${DATA_VOLUME_NAME}" 2>/dev/null || true
        docker volume rm "${LOGS_VOLUME_NAME}" 2>/dev/null || true
        echo "Nacos数据卷已删除: ${DATA_VOLUME_NAME}, ${LOGS_VOLUME_NAME}"
        echo "所有Nacos数据已被永久删除"
    else
        echo "数据卷保留: ${DATA_VOLUME_NAME}, ${LOGS_VOLUME_NAME}"
    fi
fi