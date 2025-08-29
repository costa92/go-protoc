#!/usr/bin/env bash
#
# OpenTelemetry Agent Installation Script
# This script provides functions to install, configure, and manage lightweight OpenTelemetry Agent
# both natively and via Docker containers. Agent is designed for sidecar deployment.
#

# Exit on any error, undefined variables, or pipe failures
set -o errexit
set -o nounset
set -o pipefail

# 加载通用配置和版本管理
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJ_ROOT_DIR="${SCRIPT_DIR}/../.."
source "${SCRIPT_DIR}/common.sh"

# Environment variables for OTEL Agent configuration
# Can be overridden by setting these variables before running the script
PROJ_OTEL_AGENT_HOST=${PROJ_OTEL_AGENT_HOST:-127.0.0.1}            # OTEL Agent host
PROJ_OTEL_AGENT_GRPC_PORT=${PROJ_OTEL_AGENT_GRPC_PORT:-4327}       # OTEL Agent OTLP gRPC port (different from collector)
PROJ_OTEL_AGENT_HTTP_PORT=${PROJ_OTEL_AGENT_HTTP_PORT:-4328}       # OTEL Agent OTLP HTTP port (different from collector)
PROJ_OTEL_AGENT_HEALTH_PORT=${PROJ_OTEL_AGENT_HEALTH_PORT:-13134}  # OTEL Agent Health check port
PROJ_OTEL_AGENT_CONFIG_DIR=${PROJ_OTEL_AGENT_CONFIG_DIR:-/etc/otel-agent}       # OTEL Agent config directory
PROJ_OTEL_AGENT_DATA_DIR=${PROJ_OTEL_AGENT_DATA_DIR:-/var/lib/otel-agent}       # OTEL Agent data directory
PROJ_OTEL_AGENT_LOG_DIR=${PROJ_OTEL_AGENT_LOG_DIR:-/var/log/otel-agent}         # OTEL Agent log directory
# 版本信息从统一配置文件加载：OTEL_VERSION 在 versions.sh 中定义
OTEL_AGENT_DOCKER_MNAME=${NETWORK_NAME}-otel-agent

# Function to perform pre-installation checks
proj::otel_agent::pre_install() {
  proj::log::info "Performing pre-installation checks for OpenTelemetry Agent..."

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

# Function to install OTEL Agent natively
proj::otel_agent::install() {
  proj::otel_agent::pre_install

  # 创建 OTEL Agent 相关目录
  proj::util::sudo "mkdir -p ${PROJ_OTEL_AGENT_CONFIG_DIR}"
  proj::util::sudo "mkdir -p ${PROJ_OTEL_AGENT_DATA_DIR}"
  proj::util::sudo "mkdir -p ${PROJ_OTEL_AGENT_LOG_DIR}"

  # 创建 otel-agent 用户（如果不存在）
  if ! id -u otel-agent >/dev/null 2>&1; then
    proj::util::sudo "useradd --system --shell /bin/false otel-agent"
  fi

  # 设置正确的目录所有权
  proj::util::sudo "chown -R otel-agent:otel-agent ${PROJ_OTEL_AGENT_CONFIG_DIR}"
  proj::util::sudo "chown -R otel-agent:otel-agent ${PROJ_OTEL_AGENT_DATA_DIR}"
  proj::util::sudo "chown -R otel-agent:otel-agent ${PROJ_OTEL_AGENT_LOG_DIR}"

  local otel_download_dir="/tmp/otel-agent-install"
  local arch=$(uname -m)
  local os="linux"
  
  # Convert architecture names
  case $arch in
    x86_64) arch="amd64" ;;
    aarch64) arch="arm64" ;;
  esac

  # 下载 OTEL Collector (使用相同的二进制文件，但配置不同)
  mkdir -p ${otel_download_dir}
  local download_url="https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v${OTEL_VERSION}/otelcol-contrib_${OTEL_VERSION}_${os}_${arch}.tar.gz"
  
  proj::log::info "Downloading OTEL Agent from: ${download_url}"
  curl -fsSL "${download_url}" -o "${otel_download_dir}/otelcol-contrib.tar.gz"
  
  # 解压并安装
  cd ${otel_download_dir}
  tar -xzf otelcol-contrib.tar.gz
  proj::util::sudo "cp otelcol-contrib /usr/local/bin/otelcol-agent"
  proj::util::sudo "chmod +x /usr/local/bin/otelcol-agent"

  # 复制配置文件
  proj::util::sudo "cp ${SCRIPT_DIR}/otel-agent/config.yaml ${PROJ_OTEL_AGENT_CONFIG_DIR}/config.yaml"
  
  # 创建 systemd 服务文件
  local service_file="/etc/systemd/system/otel-agent.service"
  proj::util::sudo "tee ${service_file} > /dev/null << EOF
[Unit]
Description=OpenTelemetry Agent (Sidecar)
After=network.target

[Service]
Type=simple
User=otel-agent
Group=otel-agent
ExecStart=/usr/local/bin/otelcol-agent --config=${PROJ_OTEL_AGENT_CONFIG_DIR}/config.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF"

  # 启动 OTEL Agent 服务
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "systemctl enable otel-agent"
  proj::util::sudo "systemctl start otel-agent"

  # 清理下载目录
  rm -rf ${otel_download_dir}

  sleep 5
  proj::otel_agent::status || return 1
  proj::otel_agent::info
  proj::log::info "install OTEL Agent successfully"
}

# Uninstall OTEL Agent components step by step
proj::otel_agent::uninstall() {
  set +o errexit
  proj::util::sudo "systemctl stop otel-agent" 2>/dev/null || true
  proj::util::sudo "systemctl disable otel-agent" 2>/dev/null || true
  proj::util::sudo "rm -f /etc/systemd/system/otel-agent.service"
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "rm -f /usr/local/bin/otelcol-agent"
  proj::util::sudo "rm -rf ${PROJ_OTEL_AGENT_CONFIG_DIR}"
  proj::util::sudo "rm -rf ${PROJ_OTEL_AGENT_DATA_DIR}"
  proj::util::sudo "rm -rf ${PROJ_OTEL_AGENT_LOG_DIR}"
  set -o errexit
  proj::log::info "uninstall OTEL Agent successfully"
}

# Function to install OTEL Agent using Docker
proj::otel_agent::docker::install() {
  proj::otel_agent::pre_install
  proj::common::network

  local otel_config_dir="${PROJ_THIRDPARTY_INSTALL_DIR}/otel-agent/config"
  local template_conf_file="${SCRIPT_DIR}/otel-agent/config-docker.yaml"
  
  # 检查配置文件是否存在
  if [[ ! -f "${template_conf_file}" ]]; then
    proj::log::error "OTEL Agent configuration template file not found: ${template_conf_file}"
    return 1
  fi
  
  # 清理可能存在的同名容器
  proj::common::docker::cleanup_container "${OTEL_AGENT_DOCKER_MNAME}"
  
  # 检查 proj 网络是否存在，决定网络配置
  local network_args=""
  local collector_endpoint="127.0.0.1:4317"
  
  if docker network ls | grep -q ${NETWORK_NAME}; then
    proj::log::info "${NETWORK_NAME} network detected, using networked configuration..."
    network_args="--network ${NETWORK_NAME}"
    # 如果 OTEL Collector 存在，使用容器名作为端点
    if docker ps --format '{{.Names}}' | grep -q "${NETWORK_NAME}-otel-collector"; then
      collector_endpoint="${NETWORK_NAME}-otel-collector:4317"
    fi
  else
    proj::log::info "Using standalone configuration..."
  fi
  
  # 使用 envsubst 替换模板中的环境变量
  export PROJ_SERVICE_NAME="${PROJ_SERVICE_NAME:-apiserver}"
  export PROJ_ENVIRONMENT="${PROJ_ENVIRONMENT:-development}"
  export PROJ_SERVICE_NAMESPACE="${PROJ_SERVICE_NAMESPACE:-default}"
  export PROJ_OTEL_VERSION="${OTEL_VERSION}"
  export OTEL_COLLECTOR_ENDPOINT="${collector_endpoint}"
  
  # 创建配置目录
  mkdir -p ${otel_config_dir}
  
  # 生成配置文件
  envsubst < ${template_conf_file} > ${otel_config_dir}/config.yaml

  # 启动 OTEL Agent 容器
  docker run -d \
    --name ${OTEL_AGENT_DOCKER_MNAME} \
    ${network_args} \
    -p ${PROJ_OTEL_AGENT_GRPC_PORT}:4317 \
    -p ${PROJ_OTEL_AGENT_HTTP_PORT}:4318 \
    -p ${PROJ_OTEL_AGENT_HEALTH_PORT}:13133 \
    -v ${otel_config_dir}/config.yaml:/etc/otelcol-contrib/config.yaml \
    -e PROJ_ENVIRONMENT="${PROJ_ENVIRONMENT}" \
    -e PROJ_SERVICE_NAMESPACE="${PROJ_SERVICE_NAMESPACE}" \
    -e PROJ_OTEL_VERSION="${PROJ_OTEL_VERSION}" \
    -e OTEL_COLLECTOR_ENDPOINT="${collector_endpoint}" \
    --restart unless-stopped \
    otel/opentelemetry-collector-contrib:${OTEL_VERSION} \
    --config=/etc/otelcol-contrib/config.yaml

  # 等待容器启动
  proj::log::info "Waiting for OTEL Agent to start..."
  sleep 5
  
  # 检查容器健康状态
  if ! curl -s "http://${PROJ_OTEL_AGENT_HOST}:${PROJ_OTEL_AGENT_HEALTH_PORT}" > /dev/null; then
    proj::log::error "OTEL Agent container failed to start. Checking logs..."
    docker logs ${OTEL_AGENT_DOCKER_MNAME}
    return 1
  fi

  proj::otel_agent::info
  proj::log::info "install OTEL Agent successfully"
}

# Function to uninstall Docker-based OTEL Agent
proj::otel_agent::docker::uninstall() {
  proj::log::info "Uninstalling OTEL Agent Docker container..."
  
  proj::common::docker::cleanup_container "${OTEL_AGENT_DOCKER_MNAME}"
  
  # 可选：清理配置目录
  local otel_config_dir="${PROJ_THIRDPARTY_INSTALL_DIR}/otel-agent"
  if [[ -d "${otel_config_dir}" ]]; then
    proj::log::info "Cleaning up OTEL Agent configuration directory: ${otel_config_dir}"
    rm -rf "${otel_config_dir}"
  fi
  
  proj::log::info "uninstall OTEL Agent successfully"
}

# Function to check OTEL Agent status
proj::otel_agent::status() {
  # 检查原生安装状态
  if systemctl is-active --quiet otel-agent 2>/dev/null; then
    proj::log::info "OTEL Agent (native) is running"
    return 0
  fi
  
  # 检查Docker容器状态
  if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^${OTEL_AGENT_DOCKER_MNAME}$"; then
    proj::log::info "OTEL Agent (Docker) is running"
    return 0
  fi
  
  proj::log::warn "OTEL Agent is not running"
  return 1
}

# Function to display OTEL Agent information
proj::otel_agent::info() {
  echo -e ${C_GREEN}OpenTelemetry Agent has been installed, here are some useful information:${C_NORMAL}
  
  # 检查原生安装
  if systemctl is-active --quiet otel-agent 2>/dev/null; then
    echo "  OpenTelemetry Agent OTLP HTTP endpoint: http://${PROJ_OTEL_AGENT_HOST}:${PROJ_OTEL_AGENT_HTTP_PORT}"
    echo "  OpenTelemetry Agent OTLP gRPC endpoint: ${PROJ_OTEL_AGENT_HOST}:${PROJ_OTEL_AGENT_GRPC_PORT}"
    echo "         OpenTelemetry Agent config dir: ${PROJ_OTEL_AGENT_CONFIG_DIR}"
    echo "          OpenTelemetry Agent data dir: ${PROJ_OTEL_AGENT_DATA_DIR}"
    echo "       OpenTelemetry Agent health check: http://${PROJ_OTEL_AGENT_HOST}:${PROJ_OTEL_AGENT_HEALTH_PORT}"
  fi
  
  # 检查Docker安装
  if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^${OTEL_AGENT_DOCKER_MNAME}$"; then
    echo "  OpenTelemetry Agent OTLP HTTP endpoint: http://${PROJ_OTEL_AGENT_HOST}:${PROJ_OTEL_AGENT_HTTP_PORT}"
    echo "  OpenTelemetry Agent OTLP gRPC endpoint: ${PROJ_OTEL_AGENT_HOST}:${PROJ_OTEL_AGENT_GRPC_PORT}"
    echo "         OpenTelemetry Agent data dir: ${PROJ_THIRDPARTY_INSTALL_DIR}/otel-agent"
    echo "       OpenTelemetry Agent config dir: ${PROJ_THIRDPARTY_INSTALL_DIR}/otel-agent/config"
    echo "       OpenTelemetry Agent health check: http://${PROJ_OTEL_AGENT_HOST}:${PROJ_OTEL_AGENT_HEALTH_PORT}"
  fi
}

# Handle command line arguments only when script is executed directly (not sourced)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]] && [[ $# -gt 0 ]]; then
  case $1 in
    install)
      proj::otel_agent::install
      ;;
    uninstall)
      proj::otel_agent::uninstall
      ;;
    docker.install)
      proj::otel_agent::docker::install
      ;;
    docker.uninstall)
      proj::otel_agent::docker::uninstall
      ;;
    status)
      proj::otel_agent::status
      ;;
    info)
      proj::otel_agent::info
      ;;
    *)
      proj::log::error "Unknown command: $1"
      echo "Usage: $0 {install|uninstall|docker.install|docker.uninstall|status|info}"
      exit 1
      ;;
  esac
fi