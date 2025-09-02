#!/usr/bin/env bash

# Prometheus Docker运行脚本 - 简化版
# Project: ${PROJ_NAME:-go-protoc}
# Service: Prometheus ${PROMETHEUS_VERSION}

set -eEuo pipefail

# 基础配置
readonly CONTAINER_NAME="${PROJ_PREFIX}-prometheus"
readonly IMAGE_NAME="prom/prometheus:v${PROMETHEUS_VERSION}"
readonly SERVICE_PORT="${PROJ_PROMETHEUS_PORT:-9090}"
readonly NETWORK_NAME="${PROJ_NETWORK_NAME}"
readonly CONFIG_DIR="${PROJ_PROMETHEUS_CONFIG_DIR}"
readonly DATA_DIR="${PROJ_PROMETHEUS_DATA_DIR}"

echo "🚀 启动Prometheus服务..."

# 创建目录
echo "创建必要目录..."
sudo mkdir -p "${CONFIG_DIR}" "${DATA_DIR}" "${CONFIG_DIR}/rules"
sudo chown -R $(whoami):$(id -gn) "${CONFIG_DIR}" "${DATA_DIR}"

# 生成基础配置文件
echo "生成Prometheus配置..."
cat > "${CONFIG_DIR}/prometheus.yml" << EOF
# Prometheus基础配置
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: 'go-protoc'
    environment: 'development'

rule_files:
  - "/etc/prometheus/rules/*.yml"

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
        labels:
          service: 'prometheus'
          version: '${PROMETHEUS_VERSION}'

  - job_name: 'go-protoc-api'
    static_configs:
      - targets: ['${PROJ_PREFIX}-apiserver:8080']
        labels:
          service: 'apiserver'
          environment: 'development'
    metrics_path: '/metrics'
    scrape_interval: 30s

# alerting:
#   alertmanagers:
#     - static_configs:
#         - targets:
#           - alertmanager:9093
EOF

# 自动发现和配置exporters
echo "检测并配置监控目标..."

# 检测Redis
if docker ps --filter name="${PROJ_PREFIX}-redis" --format "{{.Names}}" | grep -q "${PROJ_PREFIX}-redis"; then
    echo "✅ 发现Redis服务，添加监控配置"
    
    # 启动Redis exporter（如果未运行）
    if ! docker ps --filter name="${PROJ_PREFIX}-redis-exporter" --format "{{.Names}}" | grep -q "${PROJ_PREFIX}-redis-exporter"; then
        echo "启动Redis exporter..."
        docker run -d \
            --name "${PROJ_PREFIX}-redis-exporter" \
            --network "${NETWORK_NAME}" \
            --restart unless-stopped \
            -p 9121:9121 \
            -e REDIS_ADDR="redis://${PROJ_PREFIX}-redis:6379" \
            oliver006/redis_exporter:latest
    fi
    
    # 添加Redis监控配置到scrape_configs部分
    sed -i '/^# alerting:/i \
  - job_name: '\''redis'\''\
    static_configs:\
      - targets: ['\''${PROJ_PREFIX}-redis-exporter:9121'\'']\
        labels:\
          service: '\''redis'\''\
          environment: '\''development'\''\
    scrape_interval: 30s\
' "${CONFIG_DIR}/prometheus.yml"
fi

# 检测MySQL/MariaDB
mysql_service=""
if docker ps --filter name="${PROJ_PREFIX}-mysql" --format "{{.Names}}" | grep -q "${PROJ_PREFIX}-mysql"; then
    mysql_service="mysql"
elif docker ps --filter name="${PROJ_PREFIX}-mariadb" --format "{{.Names}}" | grep -q "${PROJ_PREFIX}-mariadb"; then
    mysql_service="mariadb"
fi

if [[ -n "${mysql_service}" ]]; then
    echo "✅ 发现${mysql_service}服务，添加监控配置"
    
    # 启动MySQL exporter（如果未运行）
    if ! docker ps --filter name="${PROJ_PREFIX}-mysql-exporter" --format "{{.Names}}" | grep -q "${PROJ_PREFIX}-mysql-exporter"; then
        echo "启动MySQL exporter..."
        docker run -d \
            --name "${PROJ_PREFIX}-mysql-exporter" \
            --network "${NETWORK_NAME}" \
            --restart unless-stopped \
            -p 9104:9104 \
            -e DATA_SOURCE_NAME="root:proj(#)666@tcp(${PROJ_PREFIX}-${mysql_service}:3306)/" \
            prom/mysqld-exporter:latest
    fi
    
    # 添加MySQL监控配置到scrape_configs部分
    sed -i '/^# alerting:/i \
  - job_name: '\''mysql'\''\
    static_configs:\
      - targets: ['\''${PROJ_PREFIX}-mysql-exporter:9104'\'']\
        labels:\
          service: '\''mysql'\''\
          database: '\''${mysql_service}'\''\
          environment: '\''development'\''\
    scrape_interval: 30s\
' "${CONFIG_DIR}/prometheus.yml"
fi

# 创建Docker网络和卷
echo "准备Docker资源..."
docker network create "${NETWORK_NAME}" 2>/dev/null || true
docker volume create "${PROJ_PREFIX}-prometheus-data" 2>/dev/null || true

# 停止现有容器
docker stop "${CONTAINER_NAME}" 2>/dev/null || true
docker rm "${CONTAINER_NAME}" 2>/dev/null || true

# 启动Prometheus容器
echo "启动Prometheus容器..."
docker run -d \
    --name "${CONTAINER_NAME}" \
    --network "${NETWORK_NAME}" \
    --restart unless-stopped \
    -p "${SERVICE_PORT}:9090" \
    -v "${PROJ_PREFIX}-prometheus-data:/prometheus" \
    -v "${CONFIG_DIR}/prometheus.yml:/etc/prometheus/prometheus.yml:ro" \
    -v "${CONFIG_DIR}/rules:/etc/prometheus/rules:ro" \
    -e PROJ_SERVICE_NAME=prometheus \
    -e PROJ_SERVICE_VERSION=${PROMETHEUS_VERSION} \
    -e PROJ_ENVIRONMENT=${PROJ_ENVIRONMENT:-development} \
    --log-driver json-file \
    --log-opt max-size=10m \
    --log-opt max-file=3 \
    --health-cmd "wget --no-verbose --tries=1 --spider http://localhost:9090/-/healthy || exit 1" \
    --health-interval 30s \
    --health-timeout 10s \
    --health-retries 3 \
    --health-start-period 30s \
    "${IMAGE_NAME}" \
    --config.file=/etc/prometheus/prometheus.yml \
    --storage.tsdb.path=/prometheus \
    --storage.tsdb.retention.time=15d \
    --storage.tsdb.retention.size=10GB \
    --web.console.libraries=/etc/prometheus/console_libraries \
    --web.console.templates=/etc/prometheus/consoles \
    --web.enable-lifecycle \
    --web.external-url=http://localhost:${SERVICE_PORT}

echo ""
echo "✅ Prometheus启动完成！"
echo "🌐 Web界面: http://localhost:${SERVICE_PORT}"
echo "📊 指标端点: http://localhost:${SERVICE_PORT}/metrics"
echo ""

# 显示容器状态
docker ps --filter name="${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"