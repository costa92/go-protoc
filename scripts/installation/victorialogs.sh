#!/usr/bin/env bash
#
# VictoriaLogs Installation Script
# This script provides functions to install, configure, and manage VictoriaLogs server
# both natively and via Docker containers.
#

# Exit on any error, undefined variables, or pipe failures
set -o errexit
set -o nounset
set -o pipefail

# Environment variables for VictoriaLogs configuration
# Can be overridden by setting these variables before running the script
PROJ_VICTORIALOGS_HOST=${PROJ_VICTORIALOGS_HOST:-127.0.0.1}                # VictoriaLogs server host
PROJ_VICTORIALOGS_PORT=${PROJ_VICTORIALOGS_PORT:-9428}                     # VictoriaLogs server port
PROJ_VICTORIALOGS_DATA_DIR=${PROJ_VICTORIALOGS_DATA_DIR:-/var/lib/victorialogs} # VictoriaLogs data directory
PROJ_VICTORIALOGS_CONFIG_DIR=${PROJ_VICTORIALOGS_CONFIG_DIR:-/etc/victorialogs} # VictoriaLogs config directory
# 加载通用配置和版本管理
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# 版本信息从统一配置文件加载：PROJ_VICTORIALOGS_VERSION 在 versions.sh 中定义
VICTORIALOGS_DOCKER_MNAME=${NETWORK_NAME}-victorialogs

# Function to install VictoriaLogs natively
proj::victorialogs::install() {
  proj::victorialogs::pre_install

  # 创建 victorialogs 相关目录
  proj::util::sudo "mkdir -p ${PROJ_VICTORIALOGS_DATA_DIR}"
  proj::util::sudo "mkdir -p ${PROJ_VICTORIALOGS_CONFIG_DIR}"
  proj::util::sudo "mkdir -p /var/log/victorialogs"
  
  # 创建 victorialogs 用户
  if ! id -u victorialogs >/dev/null 2>&1; then
    proj::util::sudo "useradd --system --shell /bin/false victorialogs"
  fi
  
  # 设置正确的目录所有权
  proj::util::sudo "chown -R victorialogs:victorialogs ${PROJ_VICTORIALOGS_DATA_DIR}"
  proj::util::sudo "chown -R victorialogs:victorialogs ${PROJ_VICTORIALOGS_CONFIG_DIR}"
  proj::util::sudo "chown -R victorialogs:victorialogs /var/log/victorialogs"

  # 下载 VictoriaLogs 二进制文件
  local victorialogs_download_dir="/tmp/victorialogs-download-test"
  mkdir -p ${victorialogs_download_dir}

  local victorialogs_arch="amd64"
  if [[ $(uname -m) == "aarch64" ]]; then
    victorialogs_arch="arm64"
  fi

  local victorialogs_url="https://github.com/VictoriaMetrics/VictoriaMetrics/releases/download/${PROJ_VICTORIALOGS_VERSION}/victoria-logs-linux-${victorialogs_arch}-${PROJ_VICTORIALOGS_VERSION}.tar.gz"

  # 检查文件是否已存在，避免重复下载
  if [[ -f "${victorialogs_download_dir}/victorialogs.tar.gz" ]]; then
    proj::log::info "VictoriaLogs ${PROJ_VICTORIALOGS_VERSION} tarball already exists, skipping download..."
  else
    proj::log::info "Downloading VictoriaLogs ${PROJ_VICTORIALOGS_VERSION} for linux-${victorialogs_arch}..."
    curl -L ${victorialogs_url} -o ${victorialogs_download_dir}/victorialogs.tar.gz
  fi

  # 解压并安装
  tar xzf ${victorialogs_download_dir}/victorialogs.tar.gz -C ${victorialogs_download_dir}
  proj::util::sudo "cp ${victorialogs_download_dir}/victoria-logs-prod /usr/local/bin/victoria-logs"
  proj::util::sudo "chmod +x /usr/local/bin/victoria-logs"

  # 创建 systemd 服务文件
  local victorialogs_service_file="/etc/systemd/system/victorialogs.service"
  local temp_service_file="/tmp/victorialogs.service.tmp"

  cat > ${temp_service_file} << EOF
[Unit]
Description=VictoriaLogs
Documentation=https://docs.victoriametrics.com/VictoriaLogs/
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
User=victorialogs
Group=victorialogs
ExecReload=/bin/kill -HUP \$MAINPID
ExecStart=/usr/local/bin/victoria-logs \\
  -storageDataPath=${PROJ_VICTORIALOGS_DATA_DIR} \\
  -httpListenAddr=0.0.0.0:${PROJ_VICTORIALOGS_PORT} \\
  -loggerLevel=INFO \\
  -loggerOutput=stderr

SyslogIdentifier=victorialogs
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

  # 复制服务文件到系统目录
  proj::util::sudo "cp ${temp_service_file} ${victorialogs_service_file}"
  rm -f ${temp_service_file}

  # 启动 VictoriaLogs 服务
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "systemctl enable victorialogs"
  proj::util::sudo "systemctl start victorialogs"

  # 清理下载目录
  # rm -rf ${victorialogs_download_dir}

  sleep 5
  proj::victorialogs::status || return 1
  proj::victorialogs::info
  proj::log::info "install victorialogs successfully"
}

# Uninstall VictoriaLogs step by step
proj::victorialogs::uninstall() {
  set +o errexit
  proj::util::sudo "systemctl stop victorialogs"
  proj::util::sudo "systemctl disable victorialogs"
  proj::util::sudo "rm -f /etc/systemd/system/victorialogs.service"
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "rm -f /usr/local/bin/victoria-logs"
  proj::util::sudo "rm -rf ${PROJ_VICTORIALOGS_CONFIG_DIR}"
  proj::util::sudo "rm -rf ${PROJ_VICTORIALOGS_DATA_DIR}"
  proj::util::sudo "rm -rf /var/log/victorialogs"
  proj::util::sudo "userdel victorialogs" 2>/dev/null || true
  set -o errexit
  proj::log::info "uninstall victorialogs successfully"
  return 0
}

# Pre-install VictoriaLogs by checking system requirements
proj::victorialogs::pre_install() {
  proj::log::info "Pre-installing VictoriaLogs..."

  # 判断是 mac 还是 linux
  if proj::util::is_mac; then
    proj::log::info "Mac OS detected, checking for VictoriaLogs installation..."
    if ! command -v victoria-logs >/dev/null 2>&1; then
      proj::log::info "Installing VictoriaLogs via brew..."
      # Note: VictoriaLogs may not be available via brew, manual installation needed
      proj::log::warning "VictoriaLogs brew package not available, please install manually"
    else
      proj::log::info "VictoriaLogs already installed, skipping..."
    fi
  else
    # 检查必要的依赖
    if ! command -v curl >/dev/null 2>&1; then
      proj::log::info "Installing curl..."
      proj::util::sudo "apt update && apt install -y curl"
    fi

    if ! command -v tar >/dev/null 2>&1; then
      proj::log::info "Installing tar..."
      proj::util::sudo "apt install -y tar"
    fi
  fi
}

# Install VictoriaLogs using a Docker container
proj::victorialogs::docker::install() {
  proj::log::info "Installing docker VictoriaLogs..."

  proj::victorialogs::pre_install
  proj::common::network

  # 创建数据目录
  local victorialogs_data_dir="${PROJ_THIRDPARTY_INSTALL_DIR}/victorialogs"
  mkdir -p ${victorialogs_data_dir}

  docker run -d --name ${VICTORIALOGS_DOCKER_MNAME} \
    --restart always \
    --network ${NETWORK_NAME} \
    -v ${victorialogs_data_dir}:/victoria-logs-data \
    -p ${PROJ_VICTORIALOGS_HOST}:${PROJ_VICTORIALOGS_PORT}:9428 \
    victoriametrics/victoria-logs:${PROJ_VICTORIALOGS_VERSION} \
    -storageDataPath=/victoria-logs-data \
    -httpListenAddr=0.0.0.0:9428 \
    -loggerLevel=INFO

  sleep 5
  if proj::util::is_linux; then
    proj::victorialogs::status || return 1
  fi
  proj::victorialogs::info
  proj::log::info "install victorialogs successfully"
}

# Print necessary information after docker or native installation
proj::victorialogs::info() {
  echo -e ${C_GREEN}VictoriaLogs has been installed, here are some useful information:${C_NORMAL}
  cat << EOF | sed 's/^/  /'
VictoriaLogs access endpoint is: http://${PROJ_VICTORIALOGS_HOST}:${PROJ_VICTORIALOGS_PORT}
       VictoriaLogs data dir: ${PROJ_VICTORIALOGS_DATA_DIR}
     VictoriaLogs config dir: ${PROJ_VICTORIALOGS_CONFIG_DIR}
        VictoriaLogs web UI: http://${PROJ_VICTORIALOGS_HOST}:${PROJ_VICTORIALOGS_PORT}
     VictoriaLogs metrics: http://${PROJ_VICTORIALOGS_HOST}:${PROJ_VICTORIALOGS_PORT}/metrics
        VictoriaLogs query: http://${PROJ_VICTORIALOGS_HOST}:${PROJ_VICTORIALOGS_PORT}/select/logsql/query
     VictoriaLogs insert: http://${PROJ_VICTORIALOGS_HOST}:${PROJ_VICTORIALOGS_PORT}/insert/jsonline
EOF
}

# Uninstall the docker container
proj::victorialogs::docker::uninstall() {
  docker rm -f ${VICTORIALOGS_DOCKER_MNAME} &>/dev/null
  proj::util::sudo "rm -rf ${PROJ_THIRDPARTY_INSTALL_DIR}/victorialogs"
  proj::log::info "uninstall victorialogs successfully"
}

# Status check after docker or native installation
proj::victorialogs::status() {
  proj::util::telnet ${PROJ_VICTORIALOGS_HOST} ${PROJ_VICTORIALOGS_PORT} || return 1

  # 检查 VictoriaLogs 健康状态
  local health_check_url="http://${PROJ_VICTORIALOGS_HOST}:${PROJ_VICTORIALOGS_PORT}/health"
  if command -v curl >/dev/null 2>&1; then
    curl -f -s ${health_check_url} >/dev/null || {
      proj::log::error "VictoriaLogs health check failed, VictoriaLogs maybe not initialized properly."
      return 1
    }
  else
    proj::log::warning "curl not found, skipping health check"
  fi
}

# This block allows calling functions in this script directly from the command line.
# It checks if the script's arguments contain a function name with the 'proj::victorialogs::' prefix
# and, if so, executes that function.
# For example: ./victorialogs.sh proj::victorialogs::install
if [[ "$*" =~ proj::victorialogs:: ]]; then
  eval $*
fi