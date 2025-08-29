#!/usr/bin/env bash
#
# OpenTelemetry (OTEL) Installation Script
# This script provides functions to install, configure, and manage OpenTelemetry Collector
# both natively and via Docker containers.
#

# Exit on any error, undefined variables, or pipe failures
set -o errexit
set -o nounset
set -o pipefail

# 加载通用配置和版本管理
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJ_ROOT_DIR="${SCRIPT_DIR}/../.."
source "${SCRIPT_DIR}/common.sh"

# Environment variables for OTEL configuration
# Can be overridden by setting these variables before running the script
PROJ_OTEL_HOST=${PROJ_OTEL_HOST:-127.0.0.1}                    # OTEL Collector host
PROJ_OTEL_GRPC_PORT=${PROJ_OTEL_GRPC_PORT:-4317}              # OTEL OTLP gRPC port
PROJ_OTEL_HTTP_PORT=${PROJ_OTEL_HTTP_PORT:-4318}              # OTEL OTLP HTTP port
PROJ_OTEL_HEALTH_PORT=${PROJ_OTEL_HEALTH_PORT:-13133}         # OTEL Health check port
PROJ_OTEL_METRICS_PORT=${PROJ_OTEL_METRICS_PORT:-8888}        # OTEL Metrics port
PROJ_OTEL_CONFIG_DIR=${PROJ_OTEL_CONFIG_DIR:-/etc/otel}       # OTEL config directory
PROJ_OTEL_DATA_DIR=${PROJ_OTEL_DATA_DIR:-/var/lib/otel}       # OTEL data directory
PROJ_OTEL_LOG_DIR=${PROJ_OTEL_LOG_DIR:-/var/log/otel}         # OTEL log directory
# 版本信息从统一配置文件加载：OTEL_VERSION 在 versions.sh 中定义
OTEL_DOCKER_MNAME=${NETWORK_NAME}-otel-collector

# Function to perform pre-installation checks
proj::otel::pre_install() {
  proj::log::info "Performing pre-installation checks for OpenTelemetry Collector..."

  # Check if Docker is available for Docker installation
  if command -v docker > /dev/null 2>&1; then
    proj::log::info "Docker is available for container-based installation"
  else
    proj::log::warn "Docker not found. Only native installation will be available."
  fi

  # Check system architecture
  local arch=$(uname -m)
  case $arch in
    x86_64|amd64)
      proj::log::info "System architecture: x86_64 (supported)"
      ;;
    aarch64|arm64)
      proj::log::info "System architecture: arm64 (supported)"
      ;;
    *)
      proj::log::error "Unsupported architecture: $arch"
      return 1
      ;;
  esac

  proj::log::info "Pre-installation checks passed"
}

# Function to install OTEL Collector natively
proj::otel::install() {
  proj::otel::pre_install

  # 创建 OTEL 相关目录
  proj::util::sudo "mkdir -p ${PROJ_OTEL_CONFIG_DIR}"
  proj::util::sudo "mkdir -p ${PROJ_OTEL_DATA_DIR}"
  proj::util::sudo "mkdir -p ${PROJ_OTEL_LOG_DIR}"

  # 创建 otel 用户（如果不存在）
  if ! id -u otel >/dev/null 2>&1; then
    proj::util::sudo "useradd --system --shell /bin/false otel"
  fi

  # 设置正确的目录所有权
  proj::util::sudo "chown -R otel:otel ${PROJ_OTEL_CONFIG_DIR}"
  proj::util::sudo "chown -R otel:otel ${PROJ_OTEL_DATA_DIR}"
  proj::util::sudo "chown -R otel:otel ${PROJ_OTEL_LOG_DIR}"

  local otel_download_dir="/tmp/otel-install"
  local arch=$(uname -m)
  local os="linux"
  
  # Convert architecture names
  case $arch in
    x86_64) arch="amd64" ;;
    aarch64) arch="arm64" ;;
  esac

  # 下载 OTEL Collector
  mkdir -p ${otel_download_dir}
  local download_url="https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v${OTEL_VERSION}/otelcol-contrib_${OTEL_VERSION}_${os}_${arch}.tar.gz"
  
  proj::log::info "Downloading OTEL Collector from: ${download_url}"
  curl -fsSL "${download_url}" -o "${otel_download_dir}/otelcol-contrib.tar.gz"
  
  # 解压并安装
  cd ${otel_download_dir}
  tar -xzf otelcol-contrib.tar.gz
  proj::util::sudo "cp otelcol-contrib /usr/local/bin/"
  proj::util::sudo "chmod +x /usr/local/bin/otelcol-contrib"

  # 复制配置文件
  proj::util::sudo "cp ${SCRIPT_DIR}/otel-collector/config.yaml ${PROJ_OTEL_CONFIG_DIR}/config.yaml"
  
  # 创建 systemd 服务文件
  local service_file="/etc/systemd/system/otel-collector.service"
  proj::util::sudo "tee ${service_file} > /dev/null << EOF
[Unit]
Description=OpenTelemetry Collector
After=network.target

[Service]
Type=simple
User=otel
Group=otel
ExecStart=/usr/local/bin/otelcol-contrib --config=${PROJ_OTEL_CONFIG_DIR}/config.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF"

  # 启动 OTEL Collector 服务
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "systemctl enable otel-collector"
  proj::util::sudo "systemctl start otel-collector"

  # 清理下载目录
  rm -rf ${otel_download_dir}

  sleep 5
  proj::otel::status || return 1
  proj::otel::info
  proj::log::info "install OTEL successfully"
}

# Uninstall OTEL components step by step
proj::otel::uninstall() {
  set +o errexit
  proj::util::sudo "systemctl stop otel-collector" 2>/dev/null || true
  proj::util::sudo "systemctl disable otel-collector" 2>/dev/null || true
  proj::util::sudo "rm -f /etc/systemd/system/otel-collector.service"
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "rm -f /usr/local/bin/otelcol-contrib"
  proj::util::sudo "rm -rf ${PROJ_OTEL_CONFIG_DIR}"
  proj::util::sudo "rm -rf ${PROJ_OTEL_DATA_DIR}"
  proj::util::sudo "rm -rf ${PROJ_OTEL_LOG_DIR}"
  set -o errexit
  proj::log::info "uninstall OTEL successfully"
}

# Function to install OTEL Collector using Docker
proj::otel::docker::install() {
  proj::otel::pre_install
  proj::common::network

  local otel_config_dir="${PROJ_THIRDPARTY_INSTALL_DIR}/otel/config"
  local template_conf_file="${SCRIPT_DIR}/otel-collector/config-docker.yaml"
  
  # 检查配置文件是否存在
  if [[ ! -f "${template_conf_file}" ]]; then
    proj::log::error "OTEL Collector configuration template file not found: ${template_conf_file}"
    return 1
  fi
  
  # 使用 envsubst 替换模板中的环境变量
  export PROJ_SERVICE_NAME="${PROJ_SERVICE_NAME:-apiserver}"
  export PROJ_ENVIRONMENT="${PROJ_ENVIRONMENT:-development}"
  export PROJ_SERVICE_NAMESPACE="${PROJ_SERVICE_NAMESPACE:-default}"
  export PROJ_OTELCOL_VERSION="${OTEL_VERSION}"
  
  # 创建配置目录
  mkdir -p ${otel_config_dir}
  
  # 生成配置文件
  envsubst < ${template_conf_file} > ${otel_config_dir}/config.yaml
  
  # 清理可能存在的同名容器
  proj::common::docker::cleanup_container "${OTEL_DOCKER_MNAME}"
  
  # 检查 Jaeger 是否存在，决定网络配置
  local network_args=""
  local logs_mount_path="/host/logs"
  
  if docker ps --format '{{.Names}}' | grep -q "${NETWORK_NAME}-jaeger"; then
    proj::log::info "Jaeger container detected, using file monitoring with networked Docker configuration..."
    network_args="--network ${NETWORK_NAME}"
    logs_mount_path="/host/logs"
  else
    proj::log::info "Using standalone Docker configuration..."
  fi

  # 确保日志目录存在
  local logs_dir="${PROJ_ROOT_DIR}/logs"
  mkdir -p "${logs_dir}"
  proj::log::info "Starting OTEL Collector container with logs directory: ${logs_dir}"

  # 启动 OTEL Collector 容器
  docker run -d \
    --name ${OTEL_DOCKER_MNAME} \
    ${network_args} \
    -p ${PROJ_OTEL_GRPC_PORT}:4317 \
    -p ${PROJ_OTEL_HTTP_PORT}:4318 \
    -p ${PROJ_OTEL_HEALTH_PORT}:13133 \
    -p ${PROJ_OTEL_METRICS_PORT}:8888 \
    -v ${otel_config_dir}/config.yaml:/etc/otelcol-contrib/config.yaml \
    -v "${logs_dir}:${logs_mount_path}" \
    -e PROJ_ENVIRONMENT="${PROJ_ENVIRONMENT}" \
    -e PROJ_SERVICE_NAMESPACE="${PROJ_SERVICE_NAMESPACE}" \
    -e PROJ_OTELCOL_VERSION="${PROJ_OTELCOL_VERSION}" \
    --restart unless-stopped \
    otel/opentelemetry-collector-contrib:${OTEL_VERSION} \
    --config=/etc/otelcol-contrib/config.yaml

  # 等待容器启动
  proj::log::info "Waiting for OTEL Collector to start..."
  sleep 5
  
  # 检查容器健康状态
  if ! curl -s "http://${PROJ_OTEL_HOST}:${PROJ_OTEL_HEALTH_PORT}" > /dev/null; then
    proj::log::error "OTEL Collector container failed to start. Checking logs..."
    docker logs ${OTEL_DOCKER_MNAME}
    return 1
  fi

  proj::otel::info
  proj::log::info "install OTEL successfully"
}

# Function to uninstall Docker-based OTEL Collector
proj::otel::docker::uninstall() {
  proj::log::info "Uninstalling OTEL Collector Docker container..."
  
  proj::common::docker::cleanup_container "${OTEL_DOCKER_MNAME}"
  
  # 可选：清理配置目录
  local otel_config_dir="${PROJ_THIRDPARTY_INSTALL_DIR}/otel"
  if [[ -d "${otel_config_dir}" ]]; then
    proj::log::info "Cleaning up OTEL configuration directory: ${otel_config_dir}"
    rm -rf "${otel_config_dir}"
  fi
  
  proj::log::info "uninstall OTEL successfully"
}

# Function to check OTEL Collector status
proj::otel::status() {
  # 检查原生安装状态
  if systemctl is-active --quiet otel-collector 2>/dev/null; then
    proj::log::info "OTEL Collector (native) is running"
    return 0
  fi
  
  # 检查Docker容器状态
  if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^${OTEL_DOCKER_MNAME}$"; then
    proj::log::info "OTEL Collector (Docker) is running"
    return 0
  fi
  
  proj::log::warn "OTEL Collector is not running"
  return 1
}

# Function to display OTEL Collector information
proj::otel::info() {
  echo -e ${C_GREEN}OpenTelemetry Collector has been installed, here are some useful information:${C_NORMAL}
  
  # 检查原生安装
  if systemctl is-active --quiet otel-collector 2>/dev/null; then
    echo "  OpenTelemetry Collector OTLP HTTP endpoint: http://${PROJ_OTEL_HOST}:${PROJ_OTEL_HTTP_PORT}"
    echo "  OpenTelemetry Collector OTLP gRPC endpoint: ${PROJ_OTEL_HOST}:${PROJ_OTEL_GRPC_PORT}"
    echo "         OpenTelemetry Collector config dir: ${PROJ_OTEL_CONFIG_DIR}"
    echo "          OpenTelemetry Collector data dir: ${PROJ_OTEL_DATA_DIR}"
    echo "      OpenTelemetry Collector metrics port: ${PROJ_OTEL_METRICS_PORT}"
    echo "       OpenTelemetry Collector health check: http://${PROJ_OTEL_HOST}:${PROJ_OTEL_HEALTH_PORT}"
  fi
  
  # 检查Docker安装
  if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^${OTEL_DOCKER_MNAME}$"; then
    echo "  OpenTelemetry Collector OTLP HTTP endpoint: http://${PROJ_OTEL_HOST}:${PROJ_OTEL_HTTP_PORT}"
    echo "  OpenTelemetry Collector OTLP gRPC endpoint: ${PROJ_OTEL_HOST}:${PROJ_OTEL_GRPC_PORT}"
    echo "         OpenTelemetry Collector data dir: ${PROJ_THIRDPARTY_INSTALL_DIR}/otel"
    echo "       OpenTelemetry Collector config dir: ${PROJ_THIRDPARTY_INSTALL_DIR}/otel/config"
    echo "      OpenTelemetry Collector metrics port: ${PROJ_OTEL_METRICS_PORT}"
    echo "       OpenTelemetry Collector health check: http://${PROJ_OTEL_HOST}:${PROJ_OTEL_HEALTH_PORT}"
  fi
}

# Handle command line arguments only when script is executed directly (not sourced)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]] && [[ $# -gt 0 ]]; then
  case $1 in
    install)
      proj::otel::install
      ;;
    uninstall)
      proj::otel::uninstall
      ;;
    docker.install)
      proj::otel::docker::install
      ;;
    docker.uninstall)
      proj::otel::docker::uninstall
      ;;
    status)
      proj::otel::status
      ;;
    info)
      proj::otel::info
      ;;
    *)
      proj::log::error "Unknown command: $1"
      echo "Usage: $0 {install|uninstall|docker.install|docker.uninstall|status|info}"
      exit 1
      ;;
  esac
fi