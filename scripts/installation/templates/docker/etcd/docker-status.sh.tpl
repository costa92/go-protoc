#!/usr/bin/env bash

# etcd Docker状态检查脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: etcd ${ETCD_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly CONTAINER_NAME="${PROJ_PREFIX}-etcd"
readonly SERVICE_PORT="${PROJ_ETCD_PORT:-2379}"

echo "=== etcd Docker服务状态 ==="

# 检查容器是否存在
if ! docker ps -a --filter name="^${CONTAINER_NAME}$" --format "table {{.Names}}" | grep -q "${CONTAINER_NAME}"; then
    echo "状态: 容器不存在"
    exit 1
fi

# 检查容器运行状态
if docker ps --filter name="^${CONTAINER_NAME}$" --format "table {{.Names}}" | grep -q "${CONTAINER_NAME}"; then
    echo "状态: 运行中 ✅"
    
    # 显示详细信息
    echo ""
    echo "容器信息:"
    docker ps --filter name="^${CONTAINER_NAME}$" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}\t{{.Image}}"
    
    echo ""
    echo "健康检查:"
    HEALTH_STATUS=$(docker inspect --format='{{.State.Health.Status}}' "${CONTAINER_NAME}" 2>/dev/null || echo "no-healthcheck")
    if [[ "${HEALTH_STATUS}" == "healthy" ]]; then
        echo "健康状态: 健康 ✅"
    elif [[ "${HEALTH_STATUS}" == "unhealthy" ]]; then
        echo "健康状态: 不健康 ❌"
    elif [[ "${HEALTH_STATUS}" == "starting" ]]; then
        echo "健康状态: 启动中 🔄"
    else
        echo "健康状态: 无健康检查"
    fi
    
    # 测试etcd连接
    echo ""
    echo "连接测试:"
    if docker exec "${CONTAINER_NAME}" etcdctl endpoint health >/dev/null 2>&1; then
        echo "etcd连接: 正常 ✅"
        echo "客户端端点: localhost:${SERVICE_PORT}"
        
        # 显示etcd信息
        echo ""
        echo "etcd集群信息:"
        docker exec "${CONTAINER_NAME}" etcdctl endpoint status --write-out=table 2>/dev/null || echo "无法获取集群状态"
        
        echo ""
        echo "版本信息:"
        docker exec "${CONTAINER_NAME}" etcdctl version 2>/dev/null | head -3 || echo "无法获取版本信息"
        
        echo ""
        echo "数据库大小:"
        docker exec "${CONTAINER_NAME}" etcdctl endpoint status --write-out=json 2>/dev/null | \
            grep -o '"dbSize":[0-9]*' | cut -d':' -f2 | \
            awk '{printf "数据库大小: %.2f MB\n", $1/1024/1024}' 2>/dev/null || echo "无法获取数据库大小"
    else
        echo "etcd连接: 失败 ❌"
    fi
    
    # 显示资源使用情况
    echo ""
    echo "资源使用:"
    docker stats "${CONTAINER_NAME}" --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}"
    
else
    echo "状态: 已停止 ⏹️"
    
    # 显示停止的容器信息
    echo ""
    echo "容器信息:"
    docker ps -a --filter name="^${CONTAINER_NAME}$" --format "table {{.Names}}\t{{.Status}}\t{{.Image}}"
fi

# 显示日志摘要
echo ""
echo "最近日志 (最后10行):"
docker logs "${CONTAINER_NAME}" --tail 10 2>/dev/null || echo "无法获取日志"