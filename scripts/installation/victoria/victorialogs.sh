#!/usr/bin/env bash
#
# VictoriaLogs Log Database Installation Script
# This script provides functions to install, configure, and manage VictoriaLogs
# both natively and via Docker containers.
#

# Exit on any error, undefined variables, or pipe failures
set -o errexit
set -o nounset
set -o pipefail

# 加载通用配置和版本管理
COMPONENT_SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${COMPONENT_SCRIPT_DIR}/../common.sh"

# Environment variables for VictoriaLogs configuration
# Can be overridden by setting these variables before running the script
PROJ_VICTORIALOGS_HOST=${PROJ_VICTORIALOGS_HOST:-127.0.0.1}
PROJ_VICTORIALOGS_PORT=${PROJ_VICTORIALOGS_PORT:-9428}

# 版本信息从统一配置文件加载：VICTORIALOGS_VERSION 在 versions.sh 中定义
VICTORIALOGS_DOCKER_NAME=${NETWORK_NAME}-victorialogs

# VictoriaLogs 数据目录
VICTORIALOGS_DATA_DIR=${PROJ_ROOT_DIR}/data/victorialogs

# Function to install VictoriaLogs natively
proj::victorialogs::install() {
  proj::victorialogs::pre_install

  proj::log::info "Installing VictoriaLogs ${VICTORIALOGS_VERSION}..."

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
  proj::util::sudo "mkdir -p /opt/victorialogs"
  proj::util::sudo "mkdir -p ${VICTORIALOGS_DATA_DIR}"

  # 下载并安装 VictoriaLogs
  local vl_url="https://github.com/VictoriaMetrics/VictoriaMetrics/releases/download/v${VICTORIALOGS_VERSION}/victoria-logs-${os}-${arch}-v${VICTORIALOGS_VERSION}.tar.gz"
  proj::log::info "Downloading VictoriaLogs from: ${vl_url}"
  
  if ! curl -L "${vl_url}" | proj::util::sudo "tar -xz -C /opt/victorialogs"; then
    proj::log::error "Failed to download and extract VictoriaLogs"
    return 1
  fi

  # 设置可执行权限
  proj::util::sudo "chmod +x /opt/victorialogs/victoria-logs-prod"

  # 创建符号链接
  proj::util::sudo "ln -sf /opt/victorialogs/victoria-logs-prod /usr/local/bin/victoria-logs"

  # 创建系统服务文件
  proj::victorialogs::create_systemd_service

  proj::log::success "VictoriaLogs ${VICTORIALOGS_VERSION} installed successfully"
  proj::victorialogs::info
}

# Function to uninstall VictoriaLogs natively
proj::victorialogs::uninstall() {
  proj::log::info "Uninstalling VictoriaLogs..."

  # 停止服务
  proj::util::sudo "systemctl stop victorialogs || true"

  # 禁用服务
  proj::util::sudo "systemctl disable victorialogs || true"

  # 删除系统服务文件
  proj::util::sudo "rm -f /etc/systemd/system/victorialogs.service"

  # 重新加载 systemd
  proj::util::sudo "systemctl daemon-reload"

  # 删除二进制文件和符号链接
  proj::util::sudo "rm -f /usr/local/bin/victoria-logs"
  proj::util::sudo "rm -rf /opt/victorialogs"

  proj::log::info "Data directory preserved: ${VICTORIALOGS_DATA_DIR}"
  proj::log::info "Run 'sudo rm -rf ${VICTORIALOGS_DATA_DIR}' to remove data"

  proj::log::success "VictoriaLogs uninstalled successfully"
}

# Function to install VictoriaLogs using Docker
proj::victorialogs::docker::install() {
  proj::victorialogs::pre_install
  proj::common::network

  proj::log::info "Installing VictoriaLogs using Docker..."

  # 创建数据目录并设置权限
  proj::util::sudo "mkdir -p ${VICTORIALOGS_DATA_DIR}"
  proj::util::sudo "chmod -R 777 ${VICTORIALOGS_DATA_DIR}"

  # 启动 VictoriaLogs 服务
  proj::log::info "Starting VictoriaLogs server..."
  docker run -d --name ${VICTORIALOGS_DOCKER_NAME} \
    --restart always \
    --network ${NETWORK_NAME} \
    -p ${PROJ_VICTORIALOGS_HOST}:${PROJ_VICTORIALOGS_PORT}:9428 \
    -v ${VICTORIALOGS_DATA_DIR}:/victoria-logs-data \
    victoriametrics/victoria-logs:v${VICTORIALOGS_VERSION} \
    -storageDataPath=/victoria-logs-data \
    -httpListenAddr=:9428 \
    -retentionPeriod=30d

  # 等待服务启动
  proj::log::info "Waiting for VictoriaLogs to start..."
  sleep 5

  # 验证服务状态
  proj::victorialogs::docker::status

  proj::log::info "VictoriaLogs Docker installation completed"
  proj::victorialogs::info
}

# Function to uninstall VictoriaLogs Docker containers
proj::victorialogs::docker::uninstall() {
  proj::log::info "Uninstalling VictoriaLogs Docker container..."

  # 停止并删除容器
  docker rm -f ${VICTORIALOGS_DOCKER_NAME} &>/dev/null || true

  # 保留数据目录（可选择性删除）
  proj::log::info "Data directory preserved: ${VICTORIALOGS_DATA_DIR}"
  proj::log::info "Run 'sudo rm -rf ${VICTORIALOGS_DATA_DIR}' to remove data"

  proj::log::info "VictoriaLogs Docker container removed"
}

# Function to perform pre-installation checks
proj::victorialogs::pre_install() {
  proj::log::info "Performing pre-installation checks for VictoriaLogs..."

  # 检查是否已安装
  if command -v victoria-logs >/dev/null 2>&1; then
    proj::log::info "VictoriaLogs is already installed"
  fi

  # 检查必要的命令
  if ! command -v curl >/dev/null 2>&1; then
    proj::log::error "curl is required but not installed"
    return 1
  fi
  
  proj::log::info "Pre-installation checks passed"
}

# Function to check VictoriaLogs status
proj::victorialogs::status() {
  proj::log::info "Checking VictoriaLogs status..."

  # 检查原生安装状态
  if command -v victoria-logs >/dev/null 2>&1; then
    proj::log::info "Native installation:"
    proj::log::info "  VictoriaLogs: $(victoria-logs --version 2>&1 | head -n1 || echo 'Version check failed')"

    # 检查服务状态
    local vl_status=$(systemctl is-active victorialogs 2>/dev/null || echo "inactive")
    proj::log::info "  Service status: ${vl_status}"
  fi

  # 检查 Docker 安装状态
  proj::victorialogs::docker::status
}

# Function to check Docker container status
proj::victorialogs::docker::status() {
  proj::log::info "Docker container:"
  
  if docker ps -q -f name="${VICTORIALOGS_DOCKER_NAME}" | grep -q .; then
    local status=$(docker inspect --format='{{.State.Status}}' "${VICTORIALOGS_DOCKER_NAME}" 2>/dev/null || echo "not found")
    proj::log::info "  ${VICTORIALOGS_DOCKER_NAME}: ${status}"
  else
    proj::log::info "  ${VICTORIALOGS_DOCKER_NAME}: not running"
  fi

  # 检查端口连通性
  proj::victorialogs::check_connectivity
}

# Function to check service connectivity
proj::victorialogs::check_connectivity() {
  proj::log::info "Checking VictoriaLogs connectivity..."

  local url="http://${PROJ_VICTORIALOGS_HOST}:${PROJ_VICTORIALOGS_PORT}/health"
  
  if curl -s --max-time 5 "${url}" >/dev/null 2>&1; then
    proj::log::info "  VictoriaLogs: ✓ (${url})"
  else
    proj::log::error "  VictoriaLogs: ✗ (${url})"
  fi
}

# Function to display VictoriaLogs information
proj::victorialogs::info() {
  proj::log::info "VictoriaLogs Information:"
  proj::log::info "======================="
  proj::log::info ""
  proj::log::info "🎯 Access URLs:"
  proj::log::info "  VictoriaLogs UI:     http://${PROJ_VICTORIALOGS_HOST}:${PROJ_VICTORIALOGS_PORT}/select/vmui"
  proj::log::info ""
  proj::log::info "📊 API Endpoints:"
  proj::log::info "  Logs Query:          http://${PROJ_VICTORIALOGS_HOST}:${PROJ_VICTORIALOGS_PORT}/select/logsql/query"
  proj::log::info "  Logs Insert:         http://${PROJ_VICTORIALOGS_HOST}:${PROJ_VICTORIALOGS_PORT}/insert/jsonline"
  proj::log::info ""
  proj::log::info "📁 Data Directory:"
  proj::log::info "  VictoriaLogs:        ${VICTORIALOGS_DATA_DIR}"
  proj::log::info ""
  proj::log::info "🔧 Management Commands:"
  proj::log::info "  Status check:        make deploy.status.victorialogs"
  proj::log::info "  Restart service:     sudo systemctl restart victorialogs"
  proj::log::info "  View logs:           sudo journalctl -f -u victorialogs"
  proj::log::info ""
}

# Function to create systemd service file
proj::victorialogs::create_systemd_service() {
  proj::log::info "Creating VictoriaLogs systemd service..."

  proj::util::sudo "tee /etc/systemd/system/victorialogs.service > /dev/null" <<EOF
[Unit]
Description=VictoriaLogs Log Database
Documentation=https://docs.victoriametrics.com/victorialogs/
After=network.target

[Service]
Type=simple
User=nobody
Group=nogroup
ExecStart=/opt/victorialogs/victoria-logs-prod \\
  -storageDataPath=${VICTORIALOGS_DATA_DIR} \\
  -httpListenAddr=:${PROJ_VICTORIALOGS_PORT} \\
  -retentionPeriod=30d \\
  -maxConcurrentInserts=8
Restart=always
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

  # 重新加载 systemd 并启用服务
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "systemctl enable victorialogs"
  proj::util::sudo "systemctl start victorialogs"

  proj::log::success "VictoriaLogs systemd service created and started"
}

# Main function to handle command line arguments
main() {
  local action=${1:-}
  
  case "${action}" in
    install)
      proj::victorialogs::install
      ;;
    uninstall)
      proj::victorialogs::uninstall
      ;;
    docker.install)
      proj::victorialogs::docker::install
      ;;
    docker.uninstall)
      proj::victorialogs::docker::uninstall
      ;;
    status)
      proj::victorialogs::status
      ;;
    info)
      proj::victorialogs::info
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