#!/usr/bin/env bash

# MongoDB Docker状态检查脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: MongoDB ${MONGODB_VERSION}

set -eEuo pipefail

# 服务配置
readonly CONTAINER_NAME="${PROJ_PREFIX}-mongodb"
readonly MONGODB_PORT="${PROJ_MONGO_PORT:-27017}"
readonly MONGODB_ROOT_USERNAME="${PROJ_MONGO_ADMIN_USERNAME:-root}"
readonly MONGODB_ROOT_PASSWORD="${PROJ_MONGO_ADMIN_PASSWORD}"

# 颜色定义
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly NC='\033[0m' # No Color

echo "========================================"
echo "MongoDB Service Status Check"
echo "========================================"

# 检查容器是否存在
if ! docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo -e "${RED}✗ MongoDB容器不存在${NC}"
    exit 1
fi

# 检查容器是否运行
if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo -e "${GREEN}✓ MongoDB容器正在运行${NC}"
    
    # 获取容器详细状态
    STATUS=$(docker inspect "$CONTAINER_NAME" --format='{{.State.Status}}')
    HEALTH=$(docker inspect "$CONTAINER_NAME" --format='{{.State.Health.Status}}' 2>/dev/null || echo "no health check")
    UPTIME=$(docker inspect "$CONTAINER_NAME" --format='{{.State.StartedAt}}')
    
    echo "  状态: $STATUS"
    echo "  健康状态: $HEALTH"
    echo "  启动时间: $UPTIME"
    
    # 显示端口映射
    echo -e "\n端口映射:"
    docker port "$CONTAINER_NAME"
    
    # 检查MongoDB连接
    echo -e "\n连接测试:"
    if docker exec "$CONTAINER_NAME" mongosh \
        -u "$MONGODB_ROOT_USERNAME" \
        -p "$MONGODB_ROOT_PASSWORD" \
        --authenticationDatabase admin \
        --eval "db.adminCommand('ping')" --quiet >/dev/null 2>&1; then
        echo -e "${GREEN}✓ MongoDB连接正常${NC}"
        
        # 显示数据库信息
        echo -e "\n数据库信息:"
        docker exec "$CONTAINER_NAME" mongosh \
            -u "$MONGODB_ROOT_USERNAME" \
            -p "$MONGODB_ROOT_PASSWORD" \
            --authenticationDatabase admin \
            --eval "db.adminCommand('listDatabases')" --quiet | head -20
        
        # 显示服务器状态
        echo -e "\n服务器状态:"
        docker exec "$CONTAINER_NAME" mongosh \
            -u "$MONGODB_ROOT_USERNAME" \
            -p "$MONGODB_ROOT_PASSWORD" \
            --authenticationDatabase admin \
            --eval "db.serverStatus().version" --quiet
    else
        echo -e "${RED}✗ MongoDB连接失败${NC}"
    fi
    
    # 显示容器资源使用
    echo -e "\n资源使用:"
    docker stats "$CONTAINER_NAME" --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}"
    
    # 显示最近日志
    echo -e "\n最近日志 (最后10行):"
    docker logs "$CONTAINER_NAME" --tail 10 2>&1
    
else
    echo -e "${YELLOW}⚠ MongoDB容器已停止${NC}"
    
    # 显示容器信息
    docker ps -a --filter name="$CONTAINER_NAME" --format "table {{.Names}}\t{{.Status}}\t{{.CreatedAt}}"
fi

# 检查数据卷
echo -e "\n数据卷状态:"
for volume in "${PROJ_PREFIX}-mongodb-data" "${PROJ_PREFIX}-mongodb-config" "${PROJ_PREFIX}-mongodb-logs"; do
    if docker volume ls --format '{{.Name}}' | grep -q "^${volume}$"; then
        SIZE=$(docker volume inspect "$volume" --format='{{.Mountpoint}}' | xargs du -sh 2>/dev/null | cut -f1 || echo "N/A")
        echo -e "${GREEN}✓${NC} $volume (大小: $SIZE)"
    else
        echo -e "${RED}✗${NC} $volume 不存在"
    fi
done

echo "========================================"