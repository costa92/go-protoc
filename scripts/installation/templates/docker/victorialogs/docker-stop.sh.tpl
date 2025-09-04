#!/usr/bin/env bash

# VictoriaLogs Docker停止脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: VictoriaLogs ${VICTORIALOGS_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly CONTAINER_NAME="${PROJ_PREFIX}-victorialogs"
readonly VOLUME_NAME="${CONTAINER_NAME_VICTORIALOGS}-data"

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

# === 统一数据卷删除处理 ===
SERVICE_NAME="VictoriaLogs"
DATA_VOLUMES="${CONTAINER_NAME_VICTORIALOGS}-data"
DATA_DESCRIPTION="  - 所有应用程序日志
  - 系统日志记录
  - 结构化日志数据
  - 索引和查询缓存
  - 日志流元数据"

# 支持多种数据删除触发方式
should_remove_data=false

# 方式1: 环境变量 (推荐用于Make命令)
if [[ "${REMOVE_DATA:-}" == "true" ]]; then
    should_remove_data=true
fi

# 方式2: 命令行参数
for arg in "$@"; do
    case $arg in
        --remove-data|--force)
            should_remove_data=true
            break
            ;;
    esac
done

if [ "$should_remove_data" = true ]; then
    echo ""
    echo "⚠️  警告：将删除${SERVICE_NAME}数据卷！这将永久删除所有日志数据！"
    echo "包括："
    echo "$DATA_DESCRIPTION"
    echo ""
    
    # 交互式确认
    if [ -t 0 ]; then  # 检查是否为交互式终端
        read -p "确认删除数据卷？(y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            IFS=',' read -ra VOLUME_ARRAY <<< "$DATA_VOLUMES"
            deleted_volumes=()
            for volume in "${VOLUME_ARRAY[@]}"; do
                volume=$(echo "$volume" | xargs)  # 去除空格
                if docker volume rm "$volume" 2>/dev/null; then
                    deleted_volumes+=("$volume")
                fi
            done
            if [ ${#deleted_volumes[@]} -gt 0 ]; then
                echo "${SERVICE_NAME}数据卷已删除: ${deleted_volumes[*]}"
                echo "所有${SERVICE_NAME}数据已被永久删除"
            else
                echo "未找到要删除的数据卷"
            fi
        else
            echo "数据卷保留: $DATA_VOLUMES"
        fi
    else
        # 非交互式环境，需要环境变量确认
        if [[ "${FORCE_DELETE:-}" == "true" ]]; then
            IFS=',' read -ra VOLUME_ARRAY <<< "$DATA_VOLUMES"
            for volume in "${VOLUME_ARRAY[@]}"; do
                volume=$(echo "$volume" | xargs)
                docker volume rm "$volume" 2>/dev/null || true
            done
            echo "${SERVICE_NAME}数据卷已强制删除: $DATA_VOLUMES"
        else
            echo "⚠️  非交互式环境，需要设置 FORCE_DELETE=true 进行强制删除"
            echo "数据卷保留: $DATA_VOLUMES"
        fi
    fi
else
    echo ""
    echo "💡 提示: ${SERVICE_NAME}数据卷已保留"
    IFS=',' read -ra VOLUME_ARRAY <<< "$DATA_VOLUMES"
    for volume in "${VOLUME_ARRAY[@]}"; do
        volume=$(echo "$volume" | xargs)
        echo "   数据卷: $volume"
    done
    echo ""
    echo "   删除数据的方式："
    echo "     环境变量: REMOVE_DATA=true make docker.victorialogs.stop"
    echo "     直接调用: bash <脚本路径> --remove-data"
    echo "     强制删除: REMOVE_DATA=true FORCE_DELETE=true make docker.victorialogs.stop"
fi