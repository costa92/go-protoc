#!/usr/bin/env bash

# Prometheus Docker停止脚本 - 简化版
# Project: ${PROJ_NAME:-go-protoc}
# Service: Prometheus ${PROMETHEUS_VERSION}

set -eEuo pipefail

# 基础配置
readonly CONTAINER_NAME="${PROJ_PREFIX}-prometheus"
readonly DATA_VOLUME_NAME="${PROJ_PREFIX}-prometheus-data"

# Exporter容器列表
readonly EXPORTER_CONTAINERS=(
    "${PROJ_PREFIX}-redis-exporter"
    "${PROJ_PREFIX}-mysql-exporter"
)

echo "🛑 停止Prometheus监控服务..."

# 停止函数
stop_container() {
    local container_name="$1"
    local description="${2:-容器}"
    
    if docker ps -a --filter name="^\${container_name}\$" --format "{{.Names}}" | grep -q "^\${container_name}\$"; then
        echo "停止\${description}: \${container_name}"
        docker stop "\${container_name}" 2>/dev/null || true
        docker rm "\${container_name}" 2>/dev/null || true
        echo "✅ \${description}已停止并删除"
    else
        echo "ℹ️ \${description}不存在，跳过"
    fi
}

# 停止所有exporter
echo ""
echo "🔌 停止监控组件..."
for exporter in "\${EXPORTER_CONTAINERS[@]}"; do
    exporter_type="\${exporter#${PROJ_PREFIX}-}"
    stop_container "\${exporter}" "\${exporter_type}"
done

# 停止Prometheus主服务
echo ""
echo "📊 停止Prometheus主服务..."
stop_container "\${CONTAINER_NAME}" "Prometheus主服务"

# 处理数据卷
echo ""
echo "💾 数据卷处理..."

# 检查命令行参数
remove_data=false
for arg in "$@"; do
    case $arg in
        --remove-data|--force)
            remove_data=true
            break
            ;;
    esac
done

if [ "$remove_data" = true ]; then
    echo "删除数据卷..."
    if docker volume rm "${DATA_VOLUME_NAME}" 2>/dev/null; then
        echo "✅ 数据卷已删除: ${DATA_VOLUME_NAME}"
    else
        echo "ℹ️ 数据卷不存在或删除失败"
    fi
else
    if docker volume inspect "${DATA_VOLUME_NAME}" >/dev/null 2>&1; then
        if [ -t 0 ]; then  # 交互式终端
            read -p "是否删除Prometheus数据卷？这将永久删除所有监控数据 (y/N): " -n 1 -r
            echo
            if [[ \$REPLY =~ ^[Yy]\$ ]]; then
                if docker volume rm "\${DATA_VOLUME_NAME}" 2>/dev/null; then
                    echo "✅ 数据卷已删除"
                else
                    echo "❌ 数据卷删除失败"
                fi
            else
                echo "ℹ️ 保留数据卷: ${DATA_VOLUME_NAME}"
            fi
        else
            echo "ℹ️ 保留数据卷: ${DATA_VOLUME_NAME}"
        fi
    else
        echo "ℹ️ 数据卷不存在"
    fi
fi

echo ""
echo "📋 停止操作完成！"
echo ""
echo "💡 常用命令:"
echo "  重新启动: make docker.prometheus.start"
echo "  查看状态: make docker.prometheus.status"
echo "  完全清理: \$0 --remove-data"