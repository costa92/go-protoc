#!/usr/bin/env bash

# Zookeeper Docker状态检查脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: Zookeeper ${ZOOKEEPER_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly SERVICE_NAME="zookeeper"
readonly CONTAINER_NAME="${PROJ_PREFIX}-zookeeper"
readonly SERVICE_PORT="${PROJ_ZOOKEEPER_PORT:-2181}"

# 检查容器状态
echo "=== Zookeeper容器状态检查 ==="
if docker ps --filter name="${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -q "${CONTAINER_NAME}"; then
    echo "✅ Zookeeper容器正在运行"
    docker ps --filter name="${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    
    # 检查服务健康状态
    echo ""
    echo "=== 服务健康检查 ==="
    if echo ruok | nc localhost "${SERVICE_PORT}" 2>/dev/null | grep -q imok; then
        echo "✅ Zookeeper健康检查通过 (ruok -> imok)"
    else
        echo "❌ Zookeeper健康检查失败"
    fi
    
    # 检查服务端口
    echo ""
    echo "=== 端口连通性 ==="
    if nc -z 127.0.0.1 "${SERVICE_PORT}" 2>/dev/null; then
        echo "✅ 客户端端口 ${SERVICE_PORT} 可访问"
    else
        echo "❌ 客户端端口 ${SERVICE_PORT} 不可访问"
    fi
    
    if nc -z 127.0.0.1 2888 2>/dev/null; then
        echo "✅ Follower端口 2888 可访问"
    else
        echo "❌ Follower端口 2888 不可访问"
    fi
    
    if nc -z 127.0.0.1 3888 2>/dev/null; then
        echo "✅ Leader选举端口 3888 可访问"
    else
        echo "❌ Leader选举端口 3888 不可访问"
    fi
    
    # 显示Zookeeper统计信息
    echo ""
    echo "=== Zookeeper统计信息 ==="
    echo "stat" | nc localhost "${SERVICE_PORT}" 2>/dev/null || echo "无法获取统计信息"
    
    # 显示服务信息
    echo ""
    echo "=== 服务信息 ==="
    echo "🔗 客户端连接: localhost:${SERVICE_PORT}"
    echo "⚡ Follower端口: localhost:2888"
    echo "🗳️  Leader选举端口: localhost:3888"
    echo "📊 状态检查命令: echo ruok | nc localhost ${SERVICE_PORT}"
    echo "📈 统计信息命令: echo stat | nc localhost ${SERVICE_PORT}"
    
    # 显示配置信息
    echo ""
    echo "=== 配置信息 ==="
    echo "conf" | nc localhost "${SERVICE_PORT}" 2>/dev/null | head -10 || echo "无法获取配置信息"
    
    # 显示最近日志
    echo ""
    echo "=== 最近日志 (最后10行) ==="
    docker logs "${CONTAINER_NAME}" --tail 10
    
else
    echo "❌ Zookeeper容器未运行"
    echo ""
    echo "可用操作:"
    echo "  启动: make docker.zookeeper.start"
    echo "  或: ./scripts/installation/templates/docker/zookeeper/docker-run.sh"
fi

# 检查数据卷
echo ""
echo "=== 数据卷状态 ==="
docker volume ls | grep -E "(${CONTAINER_NAME_ZOOKEEPER}-data|${CONTAINER_NAME_ZOOKEEPER}-logs)" || echo "未找到Zookeeper数据卷"