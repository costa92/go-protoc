#!/usr/bin/env bash

# OTEL 代理停止脚本 - 按顺序停止 Agent 和 Collector
# Project: ${PROJ_NAME:-go-protoc}
# Service: OpenTelemetry Stack (Collector + Agent) ${OTELCOL_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 加载工具函数和环境配置
if [[ -f "${PROJ_ROOT_DIR}/scripts/lib/util.sh" ]]; then
    source "${PROJ_ROOT_DIR}/scripts/lib/util.sh"
elif [[ -f "$(pwd)/scripts/lib/util.sh" ]]; then
    source "$(pwd)/scripts/lib/util.sh"
else
    proj::util::is_mac() { [[ "$(uname)" == "Darwin" ]]; }
fi

# 加载环境配置
if [[ -f "${PROJ_ROOT_DIR}/manifests/env/env.dev" ]]; then
    source "${PROJ_ROOT_DIR}/manifests/env/env.dev"
elif [[ -f "$(pwd)/manifests/env/env.dev" ]]; then
    source "$(pwd)/manifests/env/env.dev"
fi

echo "🛑 停止 OTEL Stack (Agent + Collector)..."
echo "按正确顺序: Agent 先停止，Collector 后停止"
echo ""

# 步骤1: 停止 OTEL Agent
echo "===========> 步骤 1/2: 停止 OTEL Agent"
cd "${PROJ_ROOT_DIR}" && make docker.otel-agent.stop || echo "⚠️  Agent 可能已经停止"

echo ""

# 步骤2: 停止 OTEL Collector  
echo "===========> 步骤 2/2: 停止 OTEL Collector"
cd "${PROJ_ROOT_DIR}" && make docker.otel-collector.stop || echo "⚠️  Collector 可能已经停止"

echo ""
echo "✅ OTEL Stack 已完全停止"

# 显示剩余容器状态
remaining_containers=$(docker ps --filter name="proj-otel-" --format "{{.Names}}" 2>/dev/null || true)
if [[ -n "$remaining_containers" ]]; then
    echo ""
    echo "⚠️  发现仍在运行的 OTEL 容器:"
    docker ps --filter name="proj-otel-" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
else
    echo "所有 OTEL 容器已清理完成"
fi

# === 统一数据卷删除处理 ===
SERVICE_NAME="OpenTelemetry Collector"
DATA_VOLUMES="${CONTAINER_NAME_OTELCOL}-data"
DATA_DESCRIPTION="  - 临时缓存数据
  - 处理器状态
  - 导出器队列
  - 配置文件备份
  - 统计和诊断数据"

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
    echo "⚠️  警告：将删除${SERVICE_NAME}数据卷！这将永久删除所有遥测数据！"
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
    echo "     环境变量: REMOVE_DATA=true make docker.otelcol.stop"
    echo "     直接调用: bash <脚本路径> --remove-data"
    echo "     强制删除: REMOVE_DATA=true FORCE_DELETE=true make docker.otelcol.stop"
fi