#!/usr/bin/env bash
#
# VictoriaMetrics Time Series Database Installation Script
# This script provides functions to install, configure, and manage VictoriaMetrics
# both natively and via Docker containers.
#

# Exit on any error, undefined variables, or pipe failures
set -o errexit
set -o nounset
set -o pipefail

# 加载通用配置和版本管理
COMPONENT_SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${COMPONENT_SCRIPT_DIR}/../common.sh"

# Environment variables for VictoriaMetrics configuration
# Can be overridden by setting these variables before running the script
PROJ_VICTORIAMETRICS_HOST=${PROJ_VICTORIAMETRICS_HOST:-127.0.0.1}
PROJ_VICTORIAMETRICS_PORT=${PROJ_VICTORIAMETRICS_PORT:-8428}

# 版本信息从统一配置文件加载：VICTORIAMETRICS_VERSION 在 versions.sh 中定义
VICTORIAMETRICS_DOCKER_NAME=${NETWORK_NAME}-victoriametrics

# VictoriaMetrics 数据目录
VICTORIAMETRICS_DATA_DIR=${PROJ_ROOT_DIR}/data/victoriametrics

# Function to install VictoriaMetrics natively
proj::victoriametrics::install() {
  proj::victoriametrics::pre_install

  proj::log::info "Installing VictoriaMetrics ${VICTORIAMETRICS_VERSION}..."

  # 检测系统架构
  local arch=""
  case "$(uname -m)" in
    x86_64) arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *) proj::log::error "Unsupported architecture: $(uname -m)"; return 1 ;;
  esac

  # 检测操作系统
  local os=""
  case "$(uname -s)" in
    Linux) os="linux" ;;
    Darwin) os="darwin" ;;
    *) proj::log::error "Unsupported operating system: $(uname -s)"; return 1 ;;
  esac

  # 创建安装目录
  proj::util::sudo "mkdir -p /opt/victoriametrics"
  proj::util::sudo "mkdir -p ${VICTORIAMETRICS_DATA_DIR}"

  # 下载并安装 VictoriaMetrics
  local vm_url="https://github.com/VictoriaMetrics/VictoriaMetrics/releases/download/v${VICTORIAMETRICS_VERSION}/victoria-metrics-${os}-${arch}-v${VICTORIAMETRICS_VERSION}.tar.gz"
  proj::log::info "Downloading VictoriaMetrics from: ${vm_url}"
  
  if ! curl -L "${vm_url}" | proj::util::sudo "tar -xz -C /opt/victoriametrics"; then
    proj::log::error "Failed to download and extract VictoriaMetrics"
    return 1
  fi

  # 设置可执行权限
  proj::util::sudo "chmod +x /opt/victoriametrics/victoria-metrics-prod"

  # 创建符号链接
  proj::util::sudo "ln -sf /opt/victoriametrics/victoria-metrics-prod /usr/local/bin/victoria-metrics"

  # 创建系统服务文件
  proj::victoriametrics::create_systemd_service

  proj::log::info "VictoriaMetrics ${VICTORIAMETRICS_VERSION} installed successfully"
  proj::victoriametrics::info
}

# Function to uninstall VictoriaMetrics natively
proj::victoriametrics::uninstall() {
  proj::log::info "Uninstalling VictoriaMetrics..."

  # 停止服务
  proj::util::sudo "systemctl stop victoriametrics || true"

  # 禁用服务
  proj::util::sudo "systemctl disable victoriametrics || true"

  # 删除系统服务文件
  proj::util::sudo "rm -f /etc/systemd/system/victoriametrics.service"

  # 重新加载 systemd
  proj::util::sudo "systemctl daemon-reload"

  # 删除二进制文件和符号链接
  proj::util::sudo "rm -f /usr/local/bin/victoria-metrics"
  proj::util::sudo "rm -rf /opt/victoriametrics"

  proj::log::info "Data directory preserved: ${VICTORIAMETRICS_DATA_DIR}"
  proj::log::info "Run 'sudo rm -rf ${VICTORIAMETRICS_DATA_DIR}' to remove data"

  proj::log::info "VictoriaMetrics uninstalled successfully"
}

# Function to install VictoriaMetrics using Docker
proj::victoriametrics::docker::install() {
  proj::victoriametrics::pre_install
  proj::common::network

  proj::log::info "Installing VictoriaMetrics using Docker..."

  # 创建数据目录并设置权限
  proj::util::sudo "mkdir -p ${VICTORIAMETRICS_DATA_DIR}"
  proj::util::sudo "chmod -R 777 ${VICTORIAMETRICS_DATA_DIR}"

  # 启动 VictoriaMetrics 主服务
  proj::log::info "Starting VictoriaMetrics server..."
  docker run -d --name ${VICTORIAMETRICS_DOCKER_NAME} \
    --restart always \
    --network ${NETWORK_NAME} \
    -p ${PROJ_VICTORIAMETRICS_HOST}:${PROJ_VICTORIAMETRICS_PORT}:8428 \
    -v ${VICTORIAMETRICS_DATA_DIR}:/victoria-metrics-data \
    victoriametrics/victoria-metrics:v${VICTORIAMETRICS_VERSION} \
    -storageDataPath=/victoria-metrics-data \
    -httpListenAddr=:8428 \
    -retentionPeriod=30d \
    -maxConcurrentInserts=8 \
    -search.maxQueryDuration=30s

  # 等待服务启动
  proj::log::info "Waiting for VictoriaMetrics to start..."
  sleep 5

  # 验证服务状态
  proj::victoriametrics::docker::status

  proj::log::info "VictoriaMetrics Docker installation completed"
  proj::victoriametrics::info
}

# Function to uninstall VictoriaMetrics Docker containers
proj::victoriametrics::docker::uninstall() {
  proj::log::info "Uninstalling VictoriaMetrics Docker container..."

  # 停止并删除容器
  docker rm -f ${VICTORIAMETRICS_DOCKER_NAME} &>/dev/null || true

  # 保留数据目录（可选择性删除）
  proj::log::info "Data directory preserved: ${VICTORIAMETRICS_DATA_DIR}"
  proj::log::info "Run 'sudo rm -rf ${VICTORIAMETRICS_DATA_DIR}' to remove data"

  proj::log::info "VictoriaMetrics Docker container removed"
}

# Function to perform pre-installation checks
proj::victoriametrics::pre_install() {
  proj::log::info "Performing pre-installation checks for VictoriaMetrics..."

  # 检查是否已安装
  if command -v victoria-metrics >/dev/null 2>&1; then
    proj::log::info "VictoriaMetrics is already installed"
  fi

  # 检查必要的命令
  if ! command -v curl >/dev/null 2>&1; then
    proj::log::error "curl is required but not installed"
    return 1
  fi
  
  proj::log::info "Pre-installation checks passed"
}

# Function to check VictoriaMetrics status
proj::victoriametrics::status() {
  proj::log::info "Checking VictoriaMetrics status..."

  # 检查原生安装状态
  if command -v victoria-metrics >/dev/null 2>&1; then
    proj::log::info "Native installation:"
    proj::log::info "  VictoriaMetrics: $(victoria-metrics --version 2>&1 | head -n1 || echo 'Version check failed')"

    # 检查服务状态
    local vm_status=$(systemctl is-active victoriametrics 2>/dev/null || echo "inactive")
    proj::log::info "  Service status: ${vm_status}"
  fi

  # 检查 Docker 安装状态
  proj::victoriametrics::docker::status
}

# Function to check Docker container status
proj::victoriametrics::docker::status() {
  proj::log::info "Docker container:"
  
  if docker ps -q -f name="${VICTORIAMETRICS_DOCKER_NAME}" | grep -q .; then
    local status=$(docker inspect --format='{{.State.Status}}' "${VICTORIAMETRICS_DOCKER_NAME}" 2>/dev/null || echo "not found")
    proj::log::info "  ${VICTORIAMETRICS_DOCKER_NAME}: ${status}"
  else
    proj::log::info "  ${VICTORIAMETRICS_DOCKER_NAME}: not running"
  fi

  # 检查端口连通性
  proj::victoriametrics::check_connectivity
}

# Function to check service connectivity
proj::victoriametrics::check_connectivity() {
  proj::log::info "Checking VictoriaMetrics connectivity..."

  local url="http://${PROJ_VICTORIAMETRICS_HOST}:${PROJ_VICTORIAMETRICS_PORT}/health"
  
  if curl -s --max-time 5 "${url}" >/dev/null 2>&1; then
    proj::log::info "  VictoriaMetrics: ✓ (${url})"
  else
    proj::log::error "  VictoriaMetrics: ✗ (${url})"
  fi
}

# Function to display VictoriaMetrics information
proj::victoriametrics::info() {
  proj::log::info "VictoriaMetrics Information:"
  proj::log::info "=========================="
  proj::log::info ""
  proj::log::info "🎯 Access URLs:"
  proj::log::info "  VictoriaMetrics UI:  http://${PROJ_VICTORIAMETRICS_HOST}:${PROJ_VICTORIAMETRICS_PORT}"
  proj::log::info ""
  proj::log::info "📊 API Endpoints:"
  proj::log::info "  Metrics Query:       http://${PROJ_VICTORIAMETRICS_HOST}:${PROJ_VICTORIAMETRICS_PORT}/api/v1/query"
  proj::log::info "  Metrics Insert:      http://${PROJ_VICTORIAMETRICS_HOST}:${PROJ_VICTORIAMETRICS_PORT}/api/v1/write"
  proj::log::info ""
  proj::log::info "📁 Data Directory:"
  proj::log::info "  VictoriaMetrics:     ${VICTORIAMETRICS_DATA_DIR}"
  proj::log::info ""
  proj::log::info "🔧 Management Commands:"
  proj::log::info "  Status check:        make deploy.status.victoriametrics"
  proj::log::info "  Restart service:     sudo systemctl restart victoriametrics"
  proj::log::info "  View logs:           sudo journalctl -f -u victoriametrics"
  proj::log::info ""
}

# Function to create systemd service file
proj::victoriametrics::create_systemd_service() {
  proj::log::info "Creating VictoriaMetrics systemd service..."

  proj::util::sudo "tee /etc/systemd/system/victoriametrics.service > /dev/null" <<EOF
[Unit]
Description=VictoriaMetrics Time Series Database
Documentation=https://docs.victoriametrics.com/
After=network.target

[Service]
Type=simple
User=nobody
Group=nogroup
ExecStart=/opt/victoriametrics/victoria-metrics-prod \\
  -storageDataPath=${VICTORIAMETRICS_DATA_DIR} \\
  -httpListenAddr=:${PROJ_VICTORIAMETRICS_PORT} \\
  -retentionPeriod=30d \\
  -maxConcurrentInserts=8 \\
  -search.maxQueryDuration=30s
Restart=always
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

  # 重新加载 systemd 并启用服务
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "systemctl enable victoriametrics"
  proj::util::sudo "systemctl start victoriametrics"

  proj::log::info "VictoriaMetrics systemd service created and started"
}

# Main function to handle command line arguments
main() {
  local action=${1:-}
  
  case "${action}" in
    install)
      proj::victoriametrics::install
      ;;
    uninstall)
      proj::victoriametrics::uninstall
      ;;
    docker.install)
      proj::victoriametrics::docker::install
      ;;
    docker.uninstall)
      proj::victoriametrics::docker::uninstall
      ;;
    status)
      proj::victoriametrics::status
      ;;
    info)
      proj::victoriametrics::info
      ;;
    *)
      proj::log::error "Usage: $0 {install|uninstall|docker.install|docker.uninstall|status|info}"
      return 1
      ;;
  esac
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  main "$@"
fi