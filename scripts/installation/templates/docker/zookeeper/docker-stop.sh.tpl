#!/usr/bin/env bash

# Zookeeper Docker停止脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: Zookeeper ${ZOOKEEPER_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly CONTAINER_NAME="${PROJ_PREFIX}-zookeeper"
readonly DATA_VOLUME_NAME="${PROJ_PREFIX}-zookeeper-data"
readonly LOGS_VOLUME_NAME="${PROJ_PREFIX}-zookeeper-logs"

echo "正在停止Zookeeper容器..."

# 停止容器
if docker stop "${CONTAINER_NAME}" 2>/dev/null; then
    echo "Zookeeper容器已停止: ${CONTAINER_NAME}"
else
    echo "Zookeeper容器未运行或已停止"
fi

# 删除容器
if docker rm "${CONTAINER_NAME}" 2>/dev/null; then
    echo "Zookeeper容器已删除: ${CONTAINER_NAME}"
else
    echo "Zookeeper容器不存在或已删除"
fi

echo "Zookeeper服务已完全停止"

# 可选：删除数据卷（谨慎使用）
if [[ "${1:-}" == "--remove-data" ]]; then
    echo "⚠️  警告：将删除Zookeeper数据卷！这将永久删除所有数据！"
    echo "包括："
    echo "  - Zookeeper集群数据"
    echo "  - 事务日志"
    echo "  - 快照数据"
    echo "  - 配置信息"
    echo ""
    read -p "确认删除数据卷？(y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker volume rm "${DATA_VOLUME_NAME}" 2>/dev/null || true
        docker volume rm "${LOGS_VOLUME_NAME}" 2>/dev/null || true
        echo "Zookeeper数据卷已删除: ${DATA_VOLUME_NAME}, ${LOGS_VOLUME_NAME}"
        echo "所有Zookeeper数据已被永久删除"
    else
        echo "数据卷保留: ${DATA_VOLUME_NAME}, ${LOGS_VOLUME_NAME}"
    fi
fi