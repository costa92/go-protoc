#!/usr/bin/env bash

# MongoDB Docker停止脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: MongoDB ${MONGODB_VERSION}

set -eEuo pipefail

# 服务配置
readonly CONTAINER_NAME="${PROJ_PREFIX}-mongodb"
readonly DATA_VOLUME_NAME="${PROJ_PREFIX}-mongodb-data"
readonly CONFIG_VOLUME_NAME="${PROJ_PREFIX}-mongodb-config"
readonly LOGS_VOLUME_NAME="${PROJ_PREFIX}-mongodb-logs"

echo "正在停止MongoDB容器..."

# 停止容器
if docker ps -q -f name="$CONTAINER_NAME" | grep -q .; then
    docker stop "$CONTAINER_NAME"
    echo "MongoDB容器已停止: $CONTAINER_NAME"
else
    echo "MongoDB容器未运行: $CONTAINER_NAME"
fi

# 删除容器
if docker ps -aq -f name="$CONTAINER_NAME" | grep -q .; then
    docker rm "$CONTAINER_NAME"
    echo "MongoDB容器已删除: $CONTAINER_NAME"
else
    echo "MongoDB容器不存在: $CONTAINER_NAME"
fi

# 询问是否删除数据卷
read -p "是否删除MongoDB数据卷? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    docker volume rm "$DATA_VOLUME_NAME" 2>/dev/null && echo "数据卷已删除: $DATA_VOLUME_NAME" || echo "数据卷不存在: $DATA_VOLUME_NAME"
    docker volume rm "$CONFIG_VOLUME_NAME" 2>/dev/null && echo "配置卷已删除: $CONFIG_VOLUME_NAME" || echo "配置卷不存在: $CONFIG_VOLUME_NAME"
    docker volume rm "$LOGS_VOLUME_NAME" 2>/dev/null && echo "日志卷已删除: $LOGS_VOLUME_NAME" || echo "日志卷不存在: $LOGS_VOLUME_NAME"
else
    echo "保留数据卷"
fi

echo "MongoDB服务已完全停止"