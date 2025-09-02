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

# 自动停止Zookeeper依赖
echo ""
echo "🔄 正在停止Zookeeper依赖..."

ZOOKEEPER_CONTAINER_NAME="${PROJ_PREFIX}-zookeeper"
ZOOKEEPER_DATA_VOLUME_NAME="${PROJ_PREFIX}-zookeeper-data"
ZOOKEEPER_LOGS_VOLUME_NAME="${PROJ_PREFIX}-zookeeper-logs"

# 检查Zookeeper容器是否存在并停止
if docker ps --format "{{.Names}}" | grep -q "^${ZOOKEEPER_CONTAINER_NAME}$"; then
    echo "正在停止Zookeeper容器..."
    if docker stop "${ZOOKEEPER_CONTAINER_NAME}" 2>/dev/null; then
        echo "✅ Zookeeper容器已停止: ${ZOOKEEPER_CONTAINER_NAME}"
    else
        echo "❌ Zookeeper容器停止失败"
    fi
    
    # 删除Zookeeper容器
    if docker rm "${ZOOKEEPER_CONTAINER_NAME}" 2>/dev/null; then
        echo "✅ Zookeeper容器已删除: ${ZOOKEEPER_CONTAINER_NAME}"
    else
        echo "❌ Zookeeper容器删除失败"
    fi
else
    echo "📋 Zookeeper容器未运行或不存在"
fi

echo ""
echo "🎉 Kafka 和 Zookeeper 服务已完全停止！"