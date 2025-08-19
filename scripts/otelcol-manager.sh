#!/bin/bash

# OTEL Collector 管理脚本
# 提供启动、停止、状态检查等功能，自动检测项目路径
set -euo pipefail

CONTAINER_NAME="otelcol-logs-collector"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DYNAMIC_START_SCRIPT="$SCRIPT_DIR/installation/start-otelcol-dynamic.sh"

# 显示帮助信息
show_help() {
    cat << EOF
OTEL Collector 管理工具

用法: $0 <命令> [选项]

命令:
  start     启动 OTEL Collector (自动检测项目路径)
  stop      停止 OTEL Collector
  restart   重启 OTEL Collector  
  status    查看 OTEL Collector 状态
  logs      查看 OTEL Collector 日志
  health    检查服务健康状态
  forward   转发日志到 VictoriaLogs

选项:
  -h, --help    显示此帮助信息

示例:
  $0 start                    # 启动服务
  $0 logs -f                  # 实时查看日志
  $0 status                   # 查看状态
  $0 forward                  # 转发日志到VictoriaLogs

EOF
}

# 检查容器是否运行
is_container_running() {
    docker ps --filter "name=$CONTAINER_NAME" --filter "status=running" --format '{{.Names}}' | grep -q "^$CONTAINER_NAME$"
}

# 检查容器是否存在
container_exists() {
    docker ps -a --filter "name=$CONTAINER_NAME" --format '{{.Names}}' | grep -q "^$CONTAINER_NAME$"
}

# 启动服务
start_service() {
    if is_container_running; then
        echo "ℹ️  OTEL Collector 已在运行中"
        show_status
        return 0
    fi
    
    if [[ ! -f "$DYNAMIC_START_SCRIPT" ]]; then
        echo "❌ 找不到启动脚本: $DYNAMIC_START_SCRIPT"
        exit 1
    fi
    
    echo "🚀 使用动态启动脚本..."
    "$DYNAMIC_START_SCRIPT"
}

# 停止服务
stop_service() {
    if ! container_exists; then
        echo "ℹ️  OTEL Collector 容器不存在"
        return 0
    fi
    
    echo "🛑 停止 OTEL Collector..."
    docker stop "$CONTAINER_NAME" >/dev/null 2>&1 || true
    docker rm "$CONTAINER_NAME" >/dev/null 2>&1 || true
    echo "✅ OTEL Collector 已停止"
}

# 重启服务
restart_service() {
    echo "🔄 重启 OTEL Collector..."
    stop_service
    sleep 2
    start_service
}

# 显示状态
show_status() {
    echo "📊 OTEL Collector 状态:"
    
    if is_container_running; then
        echo "✅ 容器状态: 运行中"
        docker ps --filter "name=$CONTAINER_NAME" --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}' | tail -n +2
        
        echo ""
        echo "🏥 健康检查:"
        if curl -s -f "http://localhost:13136/health" >/dev/null 2>&1; then
            echo "✅ 健康状态: 正常"
        else
            echo "❌ 健康状态: 异常"
        fi
        
        echo ""
        echo "🔗 服务端点:"
        echo "  - Health Check: http://localhost:13136/health"
        echo "  - Metrics: http://localhost:8891/metrics"
        
    elif container_exists; then
        echo "⏸️  容器状态: 已停止"
        docker ps -a --filter "name=$CONTAINER_NAME" --format 'table {{.Names}}\t{{.Status}}' | tail -n +2
    else
        echo "❌ 容器状态: 不存在"
    fi
}

# 查看日志
show_logs() {
    if ! container_exists; then
        echo "❌ OTEL Collector 容器不存在"
        exit 1
    fi
    
    local args="$*"
    if [[ -z "$args" ]]; then
        args="--tail 50"
    fi
    
    echo "📜 OTEL Collector 日志:"
    docker logs $args "$CONTAINER_NAME"
}

# 健康检查
health_check() {
    echo "🏥 执行健康检查..."
    
    if ! is_container_running; then
        echo "❌ 容器未运行"
        exit 1
    fi
    
    if curl -s -f "http://localhost:13136/health" >/dev/null 2>&1; then
        echo "✅ 健康检查通过"
        
        # 显示详细信息
        echo ""
        echo "📊 详细信息:"
        curl -s "http://localhost:13136/health" | jq . 2>/dev/null || echo "健康检查端点响应正常"
        
        # 检查指标端点
        if curl -s -f "http://localhost:8891/metrics" >/dev/null 2>&1; then
            echo "✅ 指标端点: 正常"
        else
            echo "⚠️  指标端点: 无响应"
        fi
    else
        echo "❌ 健康检查失败"
        exit 1
    fi
}

# 转发日志到VictoriaLogs
forward_logs() {
    local forward_script="$SCRIPT_DIR/forward-logs-to-victorialogs.sh"
    
    if [[ ! -f "$forward_script" ]]; then
        echo "❌ 找不到日志转发脚本: $forward_script"
        exit 1
    fi
    
    if ! is_container_running; then
        echo "❌ OTEL Collector 未运行，无法转发日志"
        exit 1
    fi
    
    echo "📤 转发日志到 VictoriaLogs..."
    "$forward_script"
}

# 主函数
main() {
    case "${1:-help}" in
        "start")
            start_service
            ;;
        "stop")
            stop_service
            ;;
        "restart")
            restart_service
            ;;
        "status")
            show_status
            ;;
        "logs")
            shift
            show_logs "$@"
            ;;
        "health")
            health_check
            ;;
        "forward")
            forward_logs
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            echo "❌ 未知命令: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"