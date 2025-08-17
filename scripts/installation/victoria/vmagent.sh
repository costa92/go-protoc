#!/usr/bin/env bash
#
# vmagent Metrics Collection Agent Installation Script
# This script provides functions to install, configure, and manage vmagent
# both natively and via Docker containers.
#

# Exit on any error, undefined variables, or pipe failures
set -o errexit
set -o nounset
set -o pipefail

# 加载通用配置和版本管理
COMPONENT_SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${COMPONENT_SCRIPT_DIR}/../common.sh"

# Environment variables for vmagent configuration
# Can be overridden by setting these variables before running the script
PROJ_VMAGENT_HOST=${PROJ_VMAGENT_HOST:-127.0.0.1}
PROJ_VMAGENT_PORT=${PROJ_VMAGENT_PORT:-8429}
PROJ_VICTORIAMETRICS_HOST=${PROJ_VICTORIAMETRICS_HOST:-127.0.0.1}
PROJ_VICTORIAMETRICS_PORT=${PROJ_VICTORIAMETRICS_PORT:-8428}

# 版本信息从统一配置文件加载：VMAGENT_VERSION 在 versions.sh 中定义
VMAGENT_DOCKER_NAME=${NETWORK_NAME}-vmagent
VICTORIAMETRICS_DOCKER_NAME=${NETWORK_NAME}-victoriametrics
VICTORIALOGS_DOCKER_NAME=${NETWORK_NAME}-victorialogs

# vmagent 数据目录
VMAGENT_DATA_DIR=${PROJ_THIRDPARTY_INSTALL_DIR}/vmagent

# Function to install vmagent natively
proj::vmagent::install() {
  proj::vmagent::pre_install

  proj::log::info "Installing vmagent ${VMAGENT_VERSION}..."

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
  proj::util::sudo "mkdir -p /opt/vmagent"
  proj::util::sudo "mkdir -p ${VMAGENT_DATA_DIR}"

  # 下载并安装 vmagent
  local vma_url="https://github.com/VictoriaMetrics/VictoriaMetrics/releases/download/v${VMAGENT_VERSION}/vmagent-${os}-${arch}-v${VMAGENT_VERSION}.tar.gz"
  proj::log::info "Downloading vmagent from: ${vma_url}"
  
  if ! curl -L "${vma_url}" | proj::util::sudo "tar -xz -C /opt/vmagent"; then
    proj::log::error "Failed to download and extract vmagent"
    return 1
  fi

  # 设置可执行权限
  proj::util::sudo "chmod +x /opt/vmagent/vmagent-prod"

  # 创建符号链接
  proj::util::sudo "ln -sf /opt/vmagent/vmagent-prod /usr/local/bin/vmagent"

  # 创建配置文件 (本地安装模式)
  proj::vmagent::create_native_config

  # 创建系统服务文件
  proj::vmagent::create_systemd_service

  proj::log::success "vmagent ${VMAGENT_VERSION} installed successfully"
  proj::vmagent::info
}

# Function to uninstall vmagent natively
proj::vmagent::uninstall() {
  proj::log::info "Uninstalling vmagent..."

  # 停止服务
  proj::util::sudo "systemctl stop vmagent || true"

  # 禁用服务
  proj::util::sudo "systemctl disable vmagent || true"

  # 删除系统服务文件
  proj::util::sudo "rm -f /etc/systemd/system/vmagent.service"

  # 重新加载 systemd
  proj::util::sudo "systemctl daemon-reload"

  # 删除二进制文件和符号链接
  proj::util::sudo "rm -f /usr/local/bin/vmagent"
  proj::util::sudo "rm -rf /opt/vmagent"

  proj::log::info "Data directory preserved: ${VMAGENT_DATA_DIR}"
  proj::log::info "Run 'sudo rm -rf ${VMAGENT_DATA_DIR}' to remove data"

  proj::log::success "vmagent uninstalled successfully"
}

# Function to install vmagent using Docker
proj::vmagent::docker::install() {
  proj::vmagent::pre_install
  proj::common::network

  proj::log::info "Installing vmagent using Docker..."

  # 创建数据目录并设置权限
  proj::util::sudo "mkdir -p ${VMAGENT_DATA_DIR}"
  proj::util::sudo "chmod -R 777 ${VMAGENT_DATA_DIR}"

  # 创建配置文件 (Docker 模式)
  proj::vmagent::create_docker_config

  # 启动 vmagent 服务
  proj::log::info "Starting vmagent server..."
  docker run -d --name ${VMAGENT_DOCKER_NAME} \
    --restart always \
    --network ${NETWORK_NAME} \
    -p ${PROJ_VMAGENT_HOST}:${PROJ_VMAGENT_PORT}:8429 \
    -v ${VMAGENT_DATA_DIR}:/vmagent-data \
    -v ${COMPONENT_SCRIPT_DIR}/vmagent-docker.yml:/etc/vmagent.yml \
    victoriametrics/vmagent:v${VMAGENT_VERSION} \
    -promscrape.config=/etc/vmagent.yml \
    -remoteWrite.url=http://${VICTORIAMETRICS_DOCKER_NAME}:8428/api/v1/write \
    -httpListenAddr=:8429 \
    -remoteWrite.tmpDataPath=/vmagent-data

  # 等待服务启动
  proj::log::info "Waiting for vmagent to start..."
  sleep 5

  # 验证服务状态
  proj::vmagent::docker::status

  proj::log::info "vmagent Docker installation completed"
  proj::vmagent::info
}

# Function to uninstall vmagent Docker containers
proj::vmagent::docker::uninstall() {
  proj::log::info "Uninstalling vmagent Docker container..."

  # 停止并删除容器
  docker rm -f ${VMAGENT_DOCKER_NAME} &>/dev/null || true

  # 保留数据目录（可选择性删除）
  proj::log::info "Data directory preserved: ${VMAGENT_DATA_DIR}"
  proj::log::info "Run 'sudo rm -rf ${VMAGENT_DATA_DIR}' to remove data"

  proj::log::info "vmagent Docker container removed"
}

# Function to perform pre-installation checks
proj::vmagent::pre_install() {
  proj::log::info "Performing pre-installation checks for vmagent..."

  # 检查是否已安装
  if command -v vmagent >/dev/null 2>&1; then
    proj::log::info "vmagent is already installed"
  fi

  # 检查必要的命令
  if ! command -v curl >/dev/null 2>&1; then
    proj::log::error "curl is required but not installed"
    return 1
  fi
  
  proj::log::info "Pre-installation checks passed"
}

# Function to check vmagent status
proj::vmagent::status() {
  proj::log::info "Checking vmagent status..."

  # 检查原生安装状态
  if command -v vmagent >/dev/null 2>&1; then
    proj::log::info "Native installation:"
    proj::log::info "  vmagent: $(vmagent --version 2>&1 | head -n1 || echo 'Version check failed')"

    # 检查服务状态
    local vma_status=$(systemctl is-active vmagent 2>/dev/null || echo "inactive")
    proj::log::info "  Service status: ${vma_status}"
  fi

  # 检查 Docker 安装状态
  proj::vmagent::docker::status
}

# Function to check Docker container status
proj::vmagent::docker::status() {
  proj::log::info "Docker container:"
  
  if docker ps -q -f name="${VMAGENT_DOCKER_NAME}" | grep -q .; then
    local status=$(docker inspect --format='{{.State.Status}}' "${VMAGENT_DOCKER_NAME}" 2>/dev/null || echo "not found")
    proj::log::info "  ${VMAGENT_DOCKER_NAME}: ${status}"
  else
    proj::log::info "  ${VMAGENT_DOCKER_NAME}: not running"
  fi

  # 检查端口连通性
  proj::vmagent::check_connectivity
}

# Function to check service connectivity
proj::vmagent::check_connectivity() {
  proj::log::info "Checking vmagent connectivity..."

  local url="http://${PROJ_VMAGENT_HOST}:${PROJ_VMAGENT_PORT}/health"
  
  if curl -s --max-time 5 "${url}" >/dev/null 2>&1; then
    proj::log::info "  vmagent: ✓ (${url})"
  else
    proj::log::error "  vmagent: ✗ (${url})"
  fi
}

# Function to display vmagent information
proj::vmagent::info() {
  proj::log::info "vmagent Information:"
  proj::log::info "=================="
  proj::log::info ""
  proj::log::info "🎯 Access URLs:"
  proj::log::info "  vmagent UI:          http://${PROJ_VMAGENT_HOST}:${PROJ_VMAGENT_PORT}"
  proj::log::info ""
  proj::log::info "📊 API Endpoints:"
  proj::log::info "  Metrics scraping:    Automatic based on configuration"
  proj::log::info "  Health check:        http://${PROJ_VMAGENT_HOST}:${PROJ_VMAGENT_PORT}/health"
  proj::log::info ""
  proj::log::info "📁 Data Directory:"
  proj::log::info "  vmagent:             ${VMAGENT_DATA_DIR}"
  proj::log::info ""
  proj::log::info "⚙️ Configuration:"
  proj::log::info "  Native config:       ${COMPONENT_SCRIPT_DIR}/vmagent.yml"
  proj::log::info "  Docker config:       ${COMPONENT_SCRIPT_DIR}/vmagent-docker.yml"
  proj::log::info ""
  proj::log::info "🔧 Management Commands:"
  proj::log::info "  Status check:        make deploy.status.vmagent"
  proj::log::info "  Restart service:     sudo systemctl restart vmagent"
  proj::log::info "  View logs:           sudo journalctl -f -u vmagent"
  proj::log::info ""
}

# Function to create Docker-specific vmagent configuration file
proj::vmagent::create_docker_config() {
  local config_file="${COMPONENT_SCRIPT_DIR}/vmagent-docker.yml"
  proj::log::info "Creating vmagent Docker configuration file: ${config_file}"

  # 确保目录存在
  mkdir -p "$(dirname "${config_file}")"

  # Docker 环境：使用容器名称和内部端口
  local vm_target="${VICTORIAMETRICS_DOCKER_NAME}:8428"
  local vl_target="${VICTORIALOGS_DOCKER_NAME}:9428"  
  local va_target="${VMAGENT_DOCKER_NAME}:8429"
  local api_target="host.docker.internal:8080"  # API 服务在宿主机上

  cat > "${config_file}" <<EOF
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'victoria-metrics'
    static_configs:
      - targets: ['${vm_target}']
    scrape_interval: 15s
    metrics_path: /metrics

  - job_name: 'victoria-logs'
    static_configs:
      - targets: ['${vl_target}']
    scrape_interval: 15s
    metrics_path: /metrics

  - job_name: 'vmagent'
    static_configs:
      - targets: ['${va_target}']
    scrape_interval: 15s
    metrics_path: /metrics

  - job_name: 'go-protoc-api'
    static_configs:
      - targets: ['${api_target}']
    scrape_interval: 15s
    metrics_path: /metrics
EOF

  proj::log::info "vmagent Docker configuration created successfully"
  proj::log::info "Docker target addresses:"
  proj::log::info "  VictoriaMetrics: ${vm_target}"
  proj::log::info "  VictoriaLogs: ${vl_target}"
  proj::log::info "  vmagent: ${va_target}"
  proj::log::info "  API Server: ${api_target}"
}

# Function to create native/local vmagent configuration file
proj::vmagent::create_native_config() {
  local config_file="${COMPONENT_SCRIPT_DIR}/vmagent.yml"
  proj::log::info "Creating vmagent native configuration file: ${config_file}"

  # 确保目录存在
  mkdir -p "$(dirname "${config_file}")"

  # 本地环境：使用 localhost 地址
  local vm_target="${PROJ_VICTORIAMETRICS_HOST}:${PROJ_VICTORIAMETRICS_PORT}"
  local vl_target="${PROJ_VICTORIALOGS_HOST:-127.0.0.1}:${PROJ_VICTORIALOGS_PORT:-9428}"
  local va_target="${PROJ_VMAGENT_HOST}:${PROJ_VMAGENT_PORT}"
  local api_target="${PROJ_VICTORIAMETRICS_HOST}:8080"

  cat > "${config_file}" <<EOF
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'victoria-metrics'
    static_configs:
      - targets: ['${vm_target}']
    scrape_interval: 15s
    metrics_path: /metrics

  - job_name: 'victoria-logs'
    static_configs:
      - targets: ['${vl_target}']
    scrape_interval: 15s
    metrics_path: /metrics

  - job_name: 'vmagent'
    static_configs:
      - targets: ['${va_target}']
    scrape_interval: 15s
    metrics_path: /metrics

  - job_name: 'go-protoc-api'
    static_configs:
      - targets: ['${api_target}']
    scrape_interval: 15s
    metrics_path: /metrics
EOF

  proj::log::info "vmagent native configuration created successfully"
  proj::log::info "Native target addresses:"
  proj::log::info "  VictoriaMetrics: ${vm_target}"
  proj::log::info "  VictoriaLogs: ${vl_target}"
  proj::log::info "  vmagent: ${va_target}"
  proj::log::info "  API Server: ${api_target}"
}

# Function to create systemd service file
proj::vmagent::create_systemd_service() {
  proj::log::info "Creating vmagent systemd service..."

  proj::util::sudo "tee /etc/systemd/system/vmagent.service > /dev/null" <<EOF
[Unit]
Description=VictoriaMetrics Agent
Documentation=https://docs.victoriametrics.com/vmagent.html
After=network.target

[Service]
Type=simple
User=nobody
Group=nogroup
ExecStart=/opt/vmagent/vmagent-prod \\
  -promscrape.config=${COMPONENT_SCRIPT_DIR}/vmagent.yml \\
  -remoteWrite.url=http://localhost:${PROJ_VICTORIAMETRICS_PORT}/api/v1/write \\
  -httpListenAddr=:${PROJ_VMAGENT_PORT} \\
  -remoteWrite.tmpDataPath=${VMAGENT_DATA_DIR}
Restart=always
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

  # 重新加载 systemd 并启用服务
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "systemctl enable vmagent"
  proj::util::sudo "systemctl start vmagent"

  proj::log::success "vmagent systemd service created and started"
}

# Main function to handle command line arguments
main() {
  local action=${1:-}
  
  case "${action}" in
    install)
      proj::vmagent::install
      ;;
    uninstall)
      proj::vmagent::uninstall
      ;;
    docker.install)
      proj::vmagent::docker::install
      ;;
    docker.uninstall)
      proj::vmagent::docker::uninstall
      ;;
    status)
      proj::vmagent::status
      ;;
    info)
      proj::vmagent::info
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