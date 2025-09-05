#!/usr/bin/env bash

# OTEL Stack (Collector + Agent) 状态检查脚本
# Project: ${PROJ_NAME}
# Service: OpenTelemetry Stack ${OTELCOL_VERSION}
# Environment: ${PROJ_ENVIRONMENT}

set -eEuo pipefail

# 服务配置  
readonly COLLECTOR_CONTAINER="proj-otel-collector"
readonly AGENT_CONTAINER="proj-otel-agent"
readonly COLLECTOR_HEALTH_PORT="4317"
readonly AGENT_HEALTH_PORT="4327"

echo "=== OTEL Stack 状态检查 ==="
echo ""

# 检查 OTEL Collector 状态
echo "📊 OTEL Collector 状态:"
if docker ps --filter name="^proj-otel-collector$" --format "{{.Names}}" | grep -q "^proj-otel-collector$"; then
    echo "✅ 状态: 运行中"
    
    # 显示容器信息
    echo ""
    echo "📦 容器信息:"
    docker ps --filter name="^proj-otel-collector$" --format "table {{.Names}}\\t{{.Status}}\\t{{.Ports}}\\t{{.Image}}"
    
    # 健康检查
    echo ""
    echo "🏥 健康检查:"
    health_status=$(docker inspect --format='{{.State.Health.Status}}' "proj-otel-collector" 2>/dev/null || echo "no-healthcheck")
    case "${health_status}" in
        "healthy") echo "✅ 健康状态: 健康" ;;
        "unhealthy") echo "❌ 健康状态: 不健康" ;;
        "starting") echo "🔄 健康状态: 启动中" ;;
        *) echo "ℹ️ 健康状态: 无健康检查" ;;
    esac
    
    # 服务连接测试
    echo ""
    echo "🌐 连接测试:"
    if curl -s "http://localhost:13133/-/healthy" >/dev/null 2>&1; then
        echo "✅ Collector连接: 正常"
        echo "🌐 健康端点: http://localhost:13133/-/healthy"
        echo "📊 指标端点: http://localhost:8888/metrics"
    else
        echo "❌ Collector连接: 失败"
    fi
    
else
    echo "❌ 状态: 未运行"
    echo "可用操作:"
    echo "  启动: make docker.otel-collector.start"
fi

echo ""
echo "----------------------------------------"
echo ""

# 检查 OTEL Agent 状态
echo "🕵️ OTEL Agent 状态:"
if docker ps --filter name="^proj-otel-agent$" --format "{{.Names}}" | grep -q "^proj-otel-agent$"; then
    echo "✅ 状态: 运行中"
    
    # 显示容器信息
    echo ""
    echo "📦 容器信息:"
    docker ps --filter name="^proj-otel-agent$" --format "table {{.Names}}\\t{{.Status}}\\t{{.Ports}}\\t{{.Image}}"
    
    # 健康检查
    echo ""
    echo "🏥 健康检查:"
    health_status=$(docker inspect --format='{{.State.Health.Status}}' "proj-otel-agent" 2>/dev/null || echo "no-healthcheck")
    case "${health_status}" in
        "healthy") echo "✅ 健康状态: 健康" ;;
        "unhealthy") echo "❌ 健康状态: 不健康" ;;
        "starting") echo "🔄 健康状态: 启动中" ;;
        *) echo "ℹ️ 健康状态: 无健康检查" ;;
    esac
    
    # 服务连接测试
    echo ""
    echo "🌐 连接测试:"
    if curl -s "http://localhost:13134/-/healthy" >/dev/null 2>&1; then
        echo "✅ Agent连接: 正常"
        echo "🌐 应用连接端点: localhost:4327 (gRPC), localhost:4328 (HTTP)"
        echo "🏥 健康端点: http://localhost:13134/-/healthy"
    else
        echo "❌ Agent连接: 失败"
    fi
    
else
    echo "❌ 状态: 未运行"
    echo "可用操作:"
    echo "  启动: make docker.otel-agent.start"
fi

echo ""
echo "========================================"
echo ""

# 整体状态总结
collector_running=false
agent_running=false

if docker ps --filter name="^proj-otel-collector$" --format "{{.Names}}" | grep -q "^proj-otel-collector$"; then
    collector_running=true
fi

if docker ps --filter name="^proj-otel-agent$" --format "{{.Names}}" | grep -q "^proj-otel-agent$"; then
    agent_running=true
fi

echo "📋 OTEL Stack 整体状态:"
if [[ "$collector_running" == true && "$agent_running" == true ]]; then
    echo "✅ 状态: 完全运行"
    echo "🔄 数据流: 应用程序 → Agent(4327) → Collector(4317) → 后端存储"
    echo ""
    echo "💡 使用说明:"
    echo "  应用程序连接: localhost:4327 (gRPC), localhost:4328 (HTTP)"
    echo "  监控面板: http://localhost:8888/metrics (Prometheus指标)"
    echo "  健康检查: http://localhost:13133 (Collector), http://localhost:13134 (Agent)"
elif [[ "$collector_running" == true ]]; then
    echo "⚠️ 状态: 仅Collector运行"
    echo "建议: make docker.otel-agent.start"
elif [[ "$agent_running" == true ]]; then
    echo "⚠️ 状态: 仅Agent运行"
    echo "建议: make docker.otel-collector.start"
else
    echo "❌ 状态: 完全停止"
    echo "启动: make docker.otelcol.start"
fi

echo ""
echo "💡 常用命令:"
echo "  重新启动: make docker.otelcol.start"
echo "  停止服务: make docker.otelcol.stop"
echo "  查看日志: docker logs proj-otel-collector (或 proj-otel-agent)"