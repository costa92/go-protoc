#!/usr/bin/env bash
#
# Grafana Loki Installation Script
# Compatible with OpenTelemetry Collector for better log storage
#

set -o errexit
set -o nounset
set -o pipefail

# 加载通用配置和版本管理
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Loki配置
PROJ_LOKI_HOST=${PROJ_LOKI_HOST:-127.0.0.1}
PROJ_LOKI_HTTP_PORT=${PROJ_LOKI_HTTP_PORT:-3100}
LOKI_DOCKER_NAME=${NETWORK_NAME}-loki

# Docker安装Grafana Loki
proj::loki::docker::install() {
  proj::log::info "Installing Grafana Loki..."
  proj::common::network
  
  # 创建Loki数据目录
  local loki_data_dir="${PROJ_THIRDPARTY_INSTALL_DIR}/loki"
  local loki_config_dir="${PROJ_THIRDPARTY_INSTALL_DIR}/loki/config"
  mkdir -p ${loki_data_dir}
  mkdir -p ${loki_config_dir}
  
  # 创建Loki配置文件
  cat > ${loki_config_dir}/loki.yaml << 'EOF'
auth_enabled: false

server:
  http_listen_port: 3100
  grpc_listen_port: 9096

common:
  instance_addr: 127.0.0.1
  path_prefix: /loki
  storage:
    filesystem:
      chunks_directory: /loki/chunks
      rules_directory: /loki/rules
  replication_factor: 1
  ring:
    kvstore:
      store: inmemory

query_range:
  results_cache:
    cache:
      embedded_cache:
        enabled: true
        max_size_mb: 100

schema_config:
  configs:
    - from: 2020-10-24
      store: tsdb
      object_store: filesystem
      schema: v13
      index:
        prefix: index_
        period: 24h

ruler:
  alertmanager_url: http://localhost:9093

limits_config:
  reject_old_samples: true
  reject_old_samples_max_age: 168h
  ingestion_rate_mb: 16
  ingestion_burst_size_mb: 32
  allow_structured_metadata: true

analytics:
  reporting_enabled: false
EOF

  # 清理可能存在的同名容器
  proj::common::docker::cleanup_container "${LOKI_DOCKER_NAME}"

  # 启动Loki容器
  proj::log::info "Starting Grafana Loki container..."
  
  docker run -d --name ${LOKI_DOCKER_NAME} \
    --restart always \
    --network ${NETWORK_NAME} \
    -v ${loki_config_dir}/loki.yaml:/etc/loki/loki.yaml:ro \
    -v ${loki_data_dir}:/loki \
    -p 127.0.0.1:3100:3100 \
    -p 127.0.0.1:9096:9096 \
    grafana/loki:3.0.0 \
    -config.file=/etc/loki/loki.yaml
  
  # 等待Loki启动
  proj::log::info "Waiting for Loki to start..."
  sleep 8
  
  # 健康检查
  local max_retries=6
  local retry_count=0
  while [[ ${retry_count} -lt ${max_retries} ]]; do
    if curl -f -s -m 10 http://127.0.0.1:3100/ready >/dev/null 2>&1; then
      break
    fi
    retry_count=$((retry_count + 1))
    proj::log::info "Health check failed, retrying... (${retry_count}/${max_retries})"
    sleep 5
  done
  
  if [[ ${retry_count} -eq ${max_retries} ]]; then
    proj::log::error "Loki health check failed after ${max_retries} attempts"
    return 1
  fi
  
  proj::loki::info
  proj::log::info "install Grafana Loki successfully"
}

# 卸载Loki
proj::loki::docker::uninstall() {
  docker rm -f ${LOKI_DOCKER_NAME} &>/dev/null
  proj::util::sudo "rm -rf ${PROJ_THIRDPARTY_INSTALL_DIR}/loki"
  proj::log::info "uninstall Grafana Loki successfully"
}

# 显示Loki信息
proj::loki::info() {
  echo -e ${C_GREEN}Grafana Loki has been installed, here are some useful information:${C_NORMAL}
  cat << EOF | sed 's/^/  /'
Grafana Loki HTTP endpoint: http://${PROJ_LOKI_HOST}:${PROJ_LOKI_HTTP_PORT}
      Grafana Loki data dir: ${PROJ_THIRDPARTY_INSTALL_DIR}/loki
    Grafana Loki config dir: ${PROJ_THIRDPARTY_INSTALL_DIR}/loki/config
         Grafana Loki ready: http://${PROJ_LOKI_HOST}:${PROJ_LOKI_HTTP_PORT}/ready
         Grafana Loki query: http://${PROJ_LOKI_HOST}:${PROJ_LOKI_HTTP_PORT}/loki/api/v1/query
EOF
}

# 状态检查
proj::loki::status() {
  if curl -f -s -m 10 http://${PROJ_LOKI_HOST}:${PROJ_LOKI_HTTP_PORT}/ready >/dev/null 2>&1; then
    return 0
  else
    proj::log::error "Grafana Loki health check failed"
    return 1
  fi
}

# 允许直接调用函数
if [[ "$*" =~ proj::loki:: ]]; then
  eval $*
fi