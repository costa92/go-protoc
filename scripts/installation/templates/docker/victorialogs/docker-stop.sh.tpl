#!/usr/bin/env bash

# VictoriaLogs Docker停止脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: VictoriaLogs ${VICTORIALOGS_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly CONTAINER_NAME="${PROJ_PREFIX}-victorialogs"
readonly VOLUME_NAME="${PROJ_PREFIX}-victorialogs-data"

echo "正在停止VictoriaLogs容器..."

# 停止容器
if docker stop "${PROJ_PREFIX}-victorialogs" 2>/dev/null; then
    echo "VictoriaLogs容器已停止: ${PROJ_PREFIX}-victorialogs"
else
    echo "VictoriaLogs容器未运行或已停止"
fi

# 删除容器
if docker rm "${PROJ_PREFIX}-victorialogs" 2>/dev/null; then
    echo "VictoriaLogs容器已删除: ${PROJ_PREFIX}-victorialogs"
else
    echo "VictoriaLogs容器不存在或已删除"
fi

echo "VictoriaLogs服务已完全停止"

# 可选：删除数据卷（谨慎使用）
if [[ "${1:-}" == "--remove-data" ]]; then
    echo "警告：将删除VictoriaLogs数据卷！"
    read -p "确认删除数据卷？(y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker volume rm "${PROJ_PREFIX}-victorialogs-data" 2>/dev/null || true
        echo "VictoriaLogs数据卷已删除: ${PROJ_PREFIX}-victorialogs-data"
    fi
fi