#!/usr/bin/env bash

# MongoDB Docker运行脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: MongoDB ${MONGODB_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly SERVICE_NAME="mongodb"
readonly CONTAINER_NAME="${PROJ_PREFIX}-mongodb"
readonly IMAGE_NAME="mongo:${MONGODB_VERSION}"
readonly MONGODB_PORT="${PROJ_MONGO_PORT:-27017}"

# 目录配置
readonly CONFIG_DIR="${PROJ_MONGODB_CONFIG_DIR}"
readonly DATA_DIR="${PROJ_MONGODB_DATA_DIR}"
readonly LOG_DIR="${PROJ_MONGODB_LOG_DIR}"

# Docker网络
readonly NETWORK_NAME="${PROJ_NETWORK_NAME}"

# MongoDB配置
readonly MONGODB_ROOT_USERNAME="${PROJ_MONGO_ADMIN_USERNAME:-root}"
readonly MONGODB_ROOT_PASSWORD="${PROJ_MONGO_ADMIN_PASSWORD}"
readonly MONGODB_DATABASE="${PROJ_MONGO_DATABASE:-proj}"
readonly MONGODB_USERNAME="${PROJ_MONGO_USERNAME:-proj}"
readonly MONGODB_PASSWORD="${PROJ_MONGO_PASSWORD}"

# 创建必要的目录
mkdir -p "$CONFIG_DIR" "$DATA_DIR" "$LOG_DIR"

# 创建Docker网络（如果不存在）
docker network create "$NETWORK_NAME" 2>/dev/null || true

# 创建Docker卷（如果不存在）
docker volume create "${PROJ_PREFIX}-mongodb-data" 2>/dev/null || true
docker volume create "${PROJ_PREFIX}-mongodb-logs" 2>/dev/null || true
docker volume create "${PROJ_PREFIX}-mongodb-config" 2>/dev/null || true

# 停止并删除现有容器（如果存在）
docker stop "$CONTAINER_NAME" 2>/dev/null || true
docker rm "$CONTAINER_NAME" 2>/dev/null || true

# 运行MongoDB容器
docker run -d \
    --name "$CONTAINER_NAME" \
    --network "$NETWORK_NAME" \
    --restart unless-stopped \
    -p "$MONGODB_PORT:27017" \
    -v "${PROJ_PREFIX}-mongodb-data:/data/db" \
    -v "${PROJ_PREFIX}-mongodb-config:/data/configdb" \
    -v "${PROJ_PREFIX}-mongodb-logs:/var/log/mongodb" \
    -v "$LOG_DIR:/host/logs" \
    -e MONGO_INITDB_ROOT_USERNAME="$MONGODB_ROOT_USERNAME" \
    -e MONGO_INITDB_ROOT_PASSWORD="$MONGODB_ROOT_PASSWORD" \
    -e MONGO_INITDB_DATABASE="$MONGODB_DATABASE" \
    -e PROJ_SERVICE_NAME=mongodb \
    -e PROJ_SERVICE_VERSION=${MONGODB_VERSION} \
    -e PROJ_ENVIRONMENT=${PROJ_ENVIRONMENT:-development} \
    --log-driver json-file \
    --log-opt max-size=10m \
    --log-opt max-file=3 \
    --health-cmd "mongosh --eval 'db.adminCommand(\"ping\")' --quiet || exit 1" \
    --health-interval 30s \
    --health-timeout 10s \
    --health-retries 5 \
    --health-start-period 40s \
    "$IMAGE_NAME" \
    mongod --auth --bind_ip_all

echo "MongoDB容器已启动: $CONTAINER_NAME"
echo "连接地址: mongodb://localhost:$MONGODB_PORT"
echo "Root用户: $MONGODB_ROOT_USERNAME"
echo "配置目录: $CONFIG_DIR"
echo "数据目录: $DATA_DIR"
echo "日志目录: $LOG_DIR"

# 显示容器状态
docker ps --filter name="$CONTAINER_NAME" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# 等待MongoDB完全启动
echo "等待MongoDB启动完成..."
timeout=60
while [ $timeout -gt 0 ]; do
    if docker exec "$CONTAINER_NAME" mongosh --eval "db.adminCommand('ping')" --quiet >/dev/null 2>&1; then
        echo "MongoDB启动成功 ✅"
        
        # 创建应用数据库和用户（如果需要）
        echo "创建应用数据库和用户..."
        docker exec "$CONTAINER_NAME" mongosh \
            -u "$MONGODB_ROOT_USERNAME" \
            -p "$MONGODB_ROOT_PASSWORD" \
            --authenticationDatabase admin \
            --eval "
                use $MONGODB_DATABASE;
                db.createUser({
                    user: '$MONGODB_USERNAME',
                    pwd: '$MONGODB_PASSWORD',
                    roles: [
                        { role: 'readWrite', db: '$MONGODB_DATABASE' },
                        { role: 'dbAdmin', db: '$MONGODB_DATABASE' }
                    ]
                });
            " 2>/dev/null || echo "用户可能已存在，跳过创建"
        
        break
    fi
    sleep 2
    ((timeout-=2))
done

if [ $timeout -le 0 ]; then
    echo "MongoDB启动超时，请检查日志 ❌"
    docker logs "$CONTAINER_NAME" --tail 20
    exit 1
fi

echo ""
echo "🎉 MongoDB服务已成功启动!"
echo "📊 连接字符串:"
echo "   - Root用户: mongodb://$MONGODB_ROOT_USERNAME:$MONGODB_ROOT_PASSWORD@localhost:$MONGODB_PORT/?authSource=admin"
echo "   - 应用用户: mongodb://$MONGODB_USERNAME:$MONGODB_PASSWORD@localhost:$MONGODB_PORT/$MONGODB_DATABASE?authSource=$MONGODB_DATABASE"
echo "📚 使用MongoDB Shell连接:"
echo "   docker exec -it $CONTAINER_NAME mongosh -u $MONGODB_ROOT_USERNAME -p $MONGODB_ROOT_PASSWORD --authenticationDatabase admin"