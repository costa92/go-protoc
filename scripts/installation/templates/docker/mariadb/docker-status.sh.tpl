#!/usr/bin/env bash

# MariaDB Docker状态检查脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: MariaDB ${MARIADB_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly SERVICE_NAME="mariadb"
readonly CONTAINER_NAME="proj-mariadb"
readonly SERVICE_PORT="${PROJ_MARIADB_PORT:-3307}"

echo "=== MariaDB Docker 容器状态检查 ==="
echo "容器名称: ${CONTAINER_NAME}"
echo "服务端口: ${SERVICE_PORT}"
echo ""

# 检查容器是否存在
if ! docker ps -a --filter name="^${CONTAINER_NAME}$" --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "❌ 容器 ${CONTAINER_NAME} 不存在"
    echo ""
    echo "建议运行: make docker.mariadb.start"
    exit 1
fi

# 获取容器详细状态
CONTAINER_STATUS=$(docker ps -a --filter name="^${CONTAINER_NAME}$" --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}\t{{.Image}}')

echo "📋 容器信息:"
echo "${CONTAINER_STATUS}"
echo ""

# 检查容器是否在运行
if docker ps --filter name="^${CONTAINER_NAME}$" --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "✅ 容器正在运行"
    
    # 检查健康状态
    HEALTH_STATUS=$(docker inspect "${CONTAINER_NAME}" --format '{{.State.Health.Status}}' 2>/dev/null || echo "unknown")
    echo "🏥 健康状态: ${HEALTH_STATUS}"
    
    # 检查端口连通性
    echo ""
    echo "🔌 连接测试:"
    if docker exec "${CONTAINER_NAME}" mariadb-admin ping -h localhost -u root -p"${MARIADB_ROOT_PASSWORD:-proj(#)666}" --silent >/dev/null 2>&1; then
        echo "  ✅ MariaDB服务可连接"
        
        # 显示版本信息
        echo ""
        echo "📊 MariaDB版本信息:"
        docker exec "${CONTAINER_NAME}" mariadb -u root -p"${MARIADB_ROOT_PASSWORD:-proj(#)666}" -e "SELECT VERSION() AS Version;" 2>/dev/null || echo "  获取版本信息失败"
        
        # 显示数据库列表
        echo ""
        echo "💾 数据库列表:"
        docker exec "${CONTAINER_NAME}" mariadb -u root -p"${MARIADB_ROOT_PASSWORD:-proj(#)666}" -e "SHOW DATABASES;" 2>/dev/null || echo "  获取数据库列表失败"
        
    else
        echo "  ❌ MariaDB服务无法连接"
    fi
    
    # 显示资源使用情况
    echo ""
    echo "📈 资源使用情况:"
    docker stats "${CONTAINER_NAME}" --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}" 2>/dev/null || echo "  无法获取资源使用情况"
    
else
    echo "❌ 容器未运行"
    
    # 显示最近日志
    echo ""
    echo "📝 最近日志 (最后10行):"
    docker logs "${CONTAINER_NAME}" --tail 10 2>/dev/null || echo "  无法获取日志"
fi

# 显示网络信息
echo ""
echo "🌐 网络信息:"
NETWORK_INFO=$(docker inspect "${CONTAINER_NAME}" --format '{{range .NetworkSettings.Networks}}{{.NetworkMode}} ({{.IPAddress}}){{end}}' 2>/dev/null || echo "  获取网络信息失败")
echo "  网络: ${NETWORK_INFO}"

# 显示卷信息
echo ""
echo "💿 数据卷信息:"
VOLUME_INFO=$(docker inspect "${CONTAINER_NAME}" --format '{{range .Mounts}}{{.Type}}: {{.Source}} -> {{.Destination}}{{"\n"}}{{end}}' 2>/dev/null || echo "  获取卷信息失败")
echo "${VOLUME_INFO}"

echo ""
echo "=== 状态检查完成 ==="