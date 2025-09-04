#!/usr/bin/env bash

# Redis Docker状态检查脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: Redis ${REDIS_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly CONTAINER_NAME="${PROJ_PREFIX}-redis"
readonly SERVICE_PORT="${PROJ_REDIS_PORT:-6379}"

echo "=== Redis Docker服务状态 ==="

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
    echo "网络信息:"
    NETWORKS=$(docker inspect "${CONTAINER_NAME}" --format '{{range $net, $conf := .NetworkSettings.Networks}}{{$net}} {{end}}' 2>/dev/null || echo "Unknown")
    echo "所在网络: ${NETWORKS}"
    for network in ${NETWORKS}; do
        IP=$(docker inspect "${CONTAINER_NAME}" --format "{{.NetworkSettings.Networks.${network}.IPAddress}}" 2>/dev/null)
        if [[ -n "$IP" ]] && [[ "$IP" != "<no value>" ]]; then
            echo "  • IP in ${network}: ${IP}"
        fi
    done
    
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
    
    # 测试Redis连接
    echo ""
    echo "连接测试:"
    if docker exec "${CONTAINER_NAME}" redis-cli ping >/dev/null 2>&1; then
        echo "Redis连接: 正常 ✅"
        
        # 显示Redis信息
        echo ""
        echo "Redis信息:"
        docker exec "${CONTAINER_NAME}" redis-cli info server | grep -E "redis_version|process_id|uptime_in_seconds"
    else
        echo "Redis连接: 失败 ❌"
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

# 显示数据卷信息
echo ""
echo "💿 数据卷信息:"
echo "数据卷列表:"
docker volume ls | grep -E "${PROJ_PREFIX}-redis" || echo "  无Redis数据卷"
echo ""
echo "数据卷挂载状态:"
if docker ps --filter name="${CONTAINER_NAME}" --format "{{.Names}}" | grep -q "${CONTAINER_NAME}"; then
    docker inspect ${CONTAINER_NAME} --format='{{range .Mounts}}{{if .Name}}  {{.Name}} -> {{.Destination}} ({{.Type}}){{"\n"}}{{end}}{{end}}' | grep -E "${PROJ_PREFIX}-redis" || echo "  无命名数据卷"
    
    echo ""
    echo "数据目录内容:"
    if docker exec ${CONTAINER_NAME} ls -la /data >/dev/null 2>&1; then
        redis_files=$(docker exec ${CONTAINER_NAME} sh -c "ls -1 /data 2>/dev/null | wc -l")
        echo "  数据文件数量: $redis_files 个"
        if [ $redis_files -gt 0 ]; then
            echo "  数据文件:"
            docker exec ${CONTAINER_NAME} sh -c "ls -la /data" | tail -n +2 | awk '{print "    " $9 " (" $5 " bytes)"}' | head -3
        fi
    else
        echo "  ❌ 无法访问数据目录"
    fi
else
    echo "  ❌ 容器未运行，无法检查挂载状态"
fi

# 显示日志摘要
echo ""
echo "最近日志 (最后10行):"
docker logs "${CONTAINER_NAME}" --tail 10 2>/dev/null || echo "无法获取日志"