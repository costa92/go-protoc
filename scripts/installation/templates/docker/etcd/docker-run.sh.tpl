#!/usr/bin/env bash

# etcd Docker运行脚本模板
# Project: ${PROJ_NAME:-go-protoc}
# Service: etcd ${ETCD_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 服务配置
readonly SERVICE_NAME="etcd"
readonly CONTAINER_NAME="${PROJ_PREFIX}-etcd"
readonly IMAGE_NAME="quay.io/coreos/etcd:${ETCD_VERSION}"
readonly SERVICE_PORT="${PROJ_ETCD_PORT:-2379}"
readonly PEER_PORT="${ETCD_PEER_PORT:-2380}"

# 目录配置
readonly CONFIG_DIR="${PROJ_ETCD_CONFIG_DIR}"
readonly DATA_DIR="${PROJ_ETCD_DATA_DIR}"
readonly LOG_DIR="${PROJ_ETCD_LOG_DIR}"

# Docker网络
readonly NETWORK_NAME="${PROJ_NETWORK_NAME}"

# etcd 集群配置
readonly ETCD_NAME="${ETCD_NAME:-etcd0}"
readonly ETCD_DATA_DIR="/etcd-data"
readonly ETCD_LISTEN_CLIENT_URLS="http://0.0.0.0:2379"
readonly ETCD_ADVERTISE_CLIENT_URLS="http://${CONTAINER_NAME}:2379"
readonly ETCD_LISTEN_PEER_URLS="http://0.0.0.0:2380"
readonly ETCD_INITIAL_ADVERTISE_PEER_URLS="http://${CONTAINER_NAME}:2380"
readonly ETCD_INITIAL_CLUSTER="${ETCD_NAME}=http://${CONTAINER_NAME}:2380"
readonly ETCD_INITIAL_CLUSTER_STATE="new"

# 创建必要的目录
mkdir -p "${CONFIG_DIR}" "${DATA_DIR}" "${LOG_DIR}"

# 创建Docker网络（如果不存在）
docker network create "${NETWORK_NAME}" 2>/dev/null || true

# 创建Docker卷（如果不存在）
docker volume create "${PROJ_PREFIX}-etcd-data" 2>/dev/null || true

# 停止并删除现有容器（如果存在）
docker stop "${CONTAINER_NAME}" 2>/dev/null || true
docker rm "${CONTAINER_NAME}" 2>/dev/null || true

# 运行etcd容器
docker run -d \
    --name "${CONTAINER_NAME}" \
    --network "${NETWORK_NAME}" \
    --restart unless-stopped \
    -p "${SERVICE_PORT}:2379" \
    -p "${PEER_PORT}:2380" \
    -v "${PROJ_PREFIX}-etcd-data":${ETCD_DATA_DIR} \
    -v "${CONFIG_DIR}:/etc/etcd:ro" \
    -v "${LOG_DIR}:/var/log/etcd" \
    -e ETCD_NAME="${ETCD_NAME}" \
    -e ETCD_DATA_DIR="${ETCD_DATA_DIR}" \
    -e ETCD_LISTEN_CLIENT_URLS="${ETCD_LISTEN_CLIENT_URLS}" \
    -e ETCD_ADVERTISE_CLIENT_URLS="${ETCD_ADVERTISE_CLIENT_URLS}" \
    -e ETCD_LISTEN_PEER_URLS="${ETCD_LISTEN_PEER_URLS}" \
    -e ETCD_INITIAL_ADVERTISE_PEER_URLS="${ETCD_INITIAL_ADVERTISE_PEER_URLS}" \
    -e ETCD_INITIAL_CLUSTER="${ETCD_INITIAL_CLUSTER}" \
    -e ETCD_INITIAL_CLUSTER_STATE="${ETCD_INITIAL_CLUSTER_STATE}" \
    -e ETCD_INITIAL_CLUSTER_TOKEN="etcd-cluster-1" \
    -e ETCD_AUTO_COMPACTION_RETENTION="1" \
    -e ETCD_QUOTA_BACKEND_BYTES="4294967296" \
    -e ETCD_HEARTBEAT_INTERVAL="250" \
    -e ETCD_ELECTION_TIMEOUT="1250" \
    -e PROJ_SERVICE_NAME=etcd \
    -e PROJ_SERVICE_VERSION=${ETCD_VERSION} \
    -e PROJ_ENVIRONMENT=${PROJ_ENVIRONMENT:-development} \
    --log-driver json-file \
    --log-opt max-size=10m \
    --log-opt max-file=3 \
    --health-cmd "etcdctl endpoint health" \
    --health-interval 30s \
    --health-timeout 10s \
    --health-retries 3 \
    --health-start-period 30s \
    "${IMAGE_NAME}" \
    etcd

echo "etcd容器已启动: ${CONTAINER_NAME}"
echo "客户端端口: localhost:${SERVICE_PORT}"
echo "节点间通信端口: localhost:${PEER_PORT}"
echo "配置目录: ${CONFIG_DIR}"
echo "数据目录: ${DATA_DIR}"
echo "使用镜像: ${IMAGE_NAME}"

# 显示容器状态
docker ps --filter name="${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# 等待etcd启动
echo "等待etcd启动完成..."
timeout=60
while [ $timeout -gt 0 ]; do
    if docker exec "${CONTAINER_NAME}" etcdctl endpoint health >/dev/null 2>&1; then
        echo "etcd启动成功 ✅"
        break
    fi
    sleep 2
    ((timeout-=2))
done

if [ $timeout -le 0 ]; then
    echo "etcd启动超时，请检查日志 ❌"
    docker logs "${CONTAINER_NAME}" --tail 20
    exit 1
fi

echo ""
echo "🎉 etcd服务已成功启动!"
echo "🔗 客户端端点: localhost:${SERVICE_PORT}"
echo "🔌 节点间通信端点: localhost:${PEER_PORT}"
echo "📋 集群信息: 单节点集群 (${ETCD_NAME})"
echo ""
echo "🛠️  常用命令:"
echo "  设置键值: docker exec ${CONTAINER_NAME} etcdctl put mykey myvalue"
echo "  获取键值: docker exec ${CONTAINER_NAME} etcdctl get mykey"
echo "  列出所有键: docker exec ${CONTAINER_NAME} etcdctl get '' --from-key"
echo "  健康检查: docker exec ${CONTAINER_NAME} etcdctl endpoint health"
echo "  集群状态: docker exec ${CONTAINER_NAME} etcdctl endpoint status --write-out=table"