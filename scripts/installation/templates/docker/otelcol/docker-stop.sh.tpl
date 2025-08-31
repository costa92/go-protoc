#!/usr/bin/env bash

# OpenTelemetry Collector Docker停止脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: OpenTelemetry Collector ${OTELCOL_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly CONTAINER_NAME="proj-otelcol"
readonly VOLUME_NAME="proj-otelcol-data"

echo "正在停止OpenTelemetry Collector容器..."

# 停止容器
if docker stop "${CONTAINER_NAME}" 2>/dev/null; then
    echo "OpenTelemetry Collector容器已停止: ${CONTAINER_NAME}"
else
    echo "OpenTelemetry Collector容器未运行或已停止"
fi

# 删除容器
if docker rm "${CONTAINER_NAME}" 2>/dev/null; then
    echo "OpenTelemetry Collector容器已删除: ${CONTAINER_NAME}"
else
    echo "OpenTelemetry Collector容器不存在或已删除"
fi

echo "OpenTelemetry Collector服务已完全停止"

# 可选：删除数据卷（谨慎使用）
if [[ "${1:-}" == "--remove-data" ]]; then
    echo "警告：将删除OpenTelemetry Collector数据卷！"
    read -p "确认删除数据卷？(y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker volume rm "${VOLUME_NAME}" 2>/dev/null || true
        echo "OpenTelemetry Collector数据卷已删除: ${VOLUME_NAME}"
    fi
fi