#!/usr/bin/env bash

# Prometheus Docker状态检查脚本 - 简化版
# Project: ${PROJ_NAME:-go-protoc}
# Service: Prometheus ${PROMETHEUS_VERSION}

set -eEuo pipefail

# 基础配置
readonly CONTAINER_NAME="${PROJ_PREFIX}-prometheus"
readonly SERVICE_PORT="${PROJ_PROMETHEUS_PORT:-9090}"

echo "=== Prometheus 服务状态检查 ==="

# 检查容器状态
if ! docker ps -a --filter name="^${CONTAINER_NAME}$" --format "{{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
    echo "❌ 状态: 容器不存在"
    exit 1
fi

if docker ps --filter name="^${CONTAINER_NAME}$" --format "{{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
    echo "✅ 状态: 运行中"
    
    # 显示容器信息
    echo ""
    echo "📦 容器信息:"
    docker ps --filter name="^${CONTAINER_NAME}$" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}\t{{.Image}}"
    
    # 健康检查
    echo ""
    echo "🏥 健康检查:"
    health_status=$(docker inspect --format='{{.State.Health.Status}}' "${CONTAINER_NAME}" 2>/dev/null || echo "no-healthcheck")
    case "${health_status}" in
        "healthy") echo "✅ 健康状态: 健康" ;;
        "unhealthy") echo "❌ 健康状态: 不健康" ;;
        "starting") echo "🔄 健康状态: 启动中" ;;
        *) echo "ℹ️ 健康状态: 无健康检查" ;;
    esac
    
    # 服务连接测试
    echo ""
    echo "🌐 连接测试:"
    if curl -s "http://localhost:${SERVICE_PORT}/-/healthy" >/dev/null 2>&1; then
        echo "✅ Prometheus连接: 正常"
        echo "🌐 Web界面: http://localhost:${SERVICE_PORT}"
        echo "📊 指标端点: http://localhost:${SERVICE_PORT}/metrics"
        
        # 获取版本信息
        echo ""
        echo "📋 服务信息:"
        version=$(curl -s "http://localhost:${SERVICE_PORT}/api/v1/status/buildinfo" 2>/dev/null | sed -n 's/.*"version":"\([^"]*\)".*/\1/p' || echo "Unknown")
        start_time=$(curl -s "http://localhost:${SERVICE_PORT}/api/v1/status/runtimeinfo" 2>/dev/null | sed -n 's/.*"startTime":"\([^"]*\)".*/\1/p' || echo "Unknown")
        echo "📌 版本: ${version}"
        echo "🕐 启动时间: ${start_time}"
    else
        echo "❌ Prometheus连接: 失败"
    fi
    
    # 检查exporters
    echo ""
    echo "📊 监控组件状态:"
    
    # Redis exporter
    if docker ps --filter name="${PROJ_PREFIX}-redis-exporter" --format "{{.Names}}" | grep -q "${PROJ_PREFIX}-redis-exporter"; then
        if curl -s "http://localhost:9121/metrics" >/dev/null 2>&1; then
            echo "✅ Redis Exporter: 运行正常 (http://localhost:9121/metrics)"
        else
            echo "⚠️ Redis Exporter: 容器运行但服务异常"
        fi
    else
        echo "ℹ️ Redis Exporter: 未运行"
    fi
    
    # MySQL exporter
    if docker ps --filter name="${PROJ_PREFIX}-mysql-exporter" --format "{{.Names}}" | grep -q "${PROJ_PREFIX}-mysql-exporter"; then
        if curl -s "http://localhost:9104/metrics" >/dev/null 2>&1; then
            echo "✅ MySQL Exporter: 运行正常 (http://localhost:9104/metrics)"
        else
            echo "⚠️ MySQL Exporter: 容器运行但服务异常"
        fi
    else
        echo "ℹ️ MySQL Exporter: 未运行"
    fi
    
    # 资源使用
    echo ""
    echo "💻 资源使用:"
    docker stats "${CONTAINER_NAME}" --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}"
    
else
    echo "⏹️ 状态: 已停止"
    
    # 显示停止的容器信息
    echo ""
    echo "📦 容器信息:"
    docker ps -a --filter name="^${CONTAINER_NAME}$" --format "table {{.Names}}\t{{.Status}}\t{{.Image}}"
fi

# 显示最近日志
echo ""
echo "📄 最近日志 (最后10行):"
docker logs "${CONTAINER_NAME}" --tail 10 2>/dev/null || echo "无法获取日志"

echo ""
echo "💡 常用命令:"
echo "  重新启动: make docker.prometheus.start"
echo "  停止服务: make docker.prometheus.stop"
echo "  查看所有容器: docker ps --filter name=${PROJ_PREFIX}-"