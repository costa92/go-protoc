#!/usr/bin/env bash

# MariaDB Docker停止脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: MariaDB ${MARIADB_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly SERVICE_NAME="mariadb"
readonly CONTAINER_NAME="${CONTAINER_NAME_MARIADB}"

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

# === 统一数据卷删除处理 ===
SERVICE_NAME="MariaDB"
DATA_VOLUMES="${CONTAINER_NAME_MARIADB}-data"
DATA_DESCRIPTION="  - 所有数据库和表
  - 用户账户和权限
  - 索引和触发器
  - 存储过程和函数
  - 二进制日志"

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
    echo "⚠️  警告：将删除${SERVICE_NAME}数据卷！这将永久删除所有数据库数据！"
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
    echo "     环境变量: REMOVE_DATA=true make docker.mariadb.stop"
    echo "     直接调用: bash <脚本路径> --remove-data"
    echo "     强制删除: REMOVE_DATA=true FORCE_DELETE=true make docker.mariadb.stop"
fi