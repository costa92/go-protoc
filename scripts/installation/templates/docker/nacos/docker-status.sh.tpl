#!/usr/bin/env bash

# Nacos Docker状态检查脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: Nacos ${NACOS_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly SERVICE_NAME="nacos"
readonly CONTAINER_NAME="${PROJ_PREFIX}-nacos"
readonly HTTP_PORT="${PROJ_NACOS_PORT:-8848}"
readonly GRPC_PORT="${PROJ_NACOS_GRPC_PORT:-9848}"

# 检查容器状态
echo "=== Nacos容器状态检查 ==="
if docker ps --filter name="${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -q "${CONTAINER_NAME}"; then
    echo "✅ Nacos容器正在运行"
    docker ps --filter name="${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    
    # 检查健康状态
    echo ""
    echo "=== 健康检查 ==="
    if docker exec "${CONTAINER_NAME}" curl -f http://localhost:8848/nacos/actuator/health >/dev/null 2>&1; then
        echo "✅ Nacos健康检查通过"
    else
        echo "❌ Nacos健康检查失败"
    fi
    
    # 检查服务端口
    echo ""
    echo "=== 端口连通性 ==="
    if nc -z 127.0.0.1 "${HTTP_PORT}" 2>/dev/null; then
        echo "✅ HTTP端口 ${HTTP_PORT} 可访问"
    else
        echo "❌ HTTP端口 ${HTTP_PORT} 不可访问"
    fi
    
    if nc -z 127.0.0.1 "${GRPC_PORT}" 2>/dev/null; then
        echo "✅ gRPC端口 ${GRPC_PORT} 可访问"
    else
        echo "❌ gRPC端口 ${GRPC_PORT} 不可访问"
    fi
    
    # 显示服务信息
    echo ""
    echo "=== 服务信息 ==="
    echo "🌐 Console UI: http://localhost:${HTTP_PORT}/nacos/"
    echo "🔗 HTTP API:  http://localhost:${HTTP_PORT}/nacos/v1/"
    echo "⚡ gRPC端口:  localhost:${GRPC_PORT}"
    echo "👤 默认用户名/密码: nacos/nacos"
    
    # 显示内存信息
    echo ""
    echo "=== 内存配置 ==="
    echo "📊 容器内存限制: ${NACOS_MAX_MEMORY:-1g}"
    echo "☕ JVM堆内存: ${NACOS_JVM_XMS:-512m} - ${NACOS_JVM_XMX:-1g}"
    echo "🔥 新生代内存: ${NACOS_JVM_XMN:-256m}"
    
    # 显示最近日志
    echo ""
    echo "=== 最近日志 (最后10行) ==="
    docker logs "${CONTAINER_NAME}" --tail 10
    
else
    echo "❌ Nacos容器未运行"
    echo ""
    echo "可用操作:"
    echo "  启动: make docker.nacos.start"
    echo "  或: ./scripts/installation/templates/docker/nacos/docker-run.sh"
fi

# 检查数据卷
echo ""
echo "=== 数据卷状态 ==="
docker volume ls | grep -E "(${PROJ_PREFIX}-nacos-data|${PROJ_PREFIX}-nacos-logs)" || echo "未找到Nacos数据卷"