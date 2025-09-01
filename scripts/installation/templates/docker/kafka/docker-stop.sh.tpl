#!/usr/bin/env bash

# Kafka Docker停止脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: Kafka ${KAFKA_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly CONTAINER_NAME="${PROJ_PREFIX}-kafka"
readonly DATA_VOLUME_NAME="${PROJ_PREFIX}-kafka-data"
readonly LOGS_VOLUME_NAME="${PROJ_PREFIX}-kafka-logs"

echo "正在停止Kafka容器..."

# 停止容器
if docker stop "${CONTAINER_NAME}" 2>/dev/null; then
    echo "Kafka容器已停止: ${CONTAINER_NAME}"
else
    echo "Kafka容器未运行或已停止"
fi

# 删除容器
if docker rm "${CONTAINER_NAME}" 2>/dev/null; then
    echo "Kafka容器已删除: ${CONTAINER_NAME}"
else
    echo "Kafka容器不存在或已删除"
fi

echo "Kafka服务已完全停止"

# 可选：删除数据卷（谨慎使用）
if [[ "${1:-}" == "--remove-data" ]]; then
    echo "⚠️  警告：将删除Kafka数据卷！这将永久删除所有数据！"
    echo "包括："
    echo "  - 所有Topic数据"
    echo "  - 消息日志"
    echo "  - 分区数据"
    echo "  - 索引文件"
    echo "  - 事务状态日志"
    echo ""
    read -p "确认删除数据卷？(y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker volume rm "${DATA_VOLUME_NAME}" 2>/dev/null || true
        docker volume rm "${LOGS_VOLUME_NAME}" 2>/dev/null || true
        echo "Kafka数据卷已删除: ${DATA_VOLUME_NAME}, ${LOGS_VOLUME_NAME}"
        echo "所有Kafka数据已被永久删除"
    else
        echo "数据卷保留: ${DATA_VOLUME_NAME}, ${LOGS_VOLUME_NAME}"
    fi
fi

# 提示Zookeeper状态
echo ""
echo "📋 注意: Zookeeper容器仍在运行，如需停止请运行:"
echo "  make docker.zookeeper.stop"
echo "  或"
echo "  ./scripts/installation/templates/docker/zookeeper/docker-stop.sh"