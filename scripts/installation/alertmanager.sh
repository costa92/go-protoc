#!/usr/bin/env bash
#
# AlertManager Installation Script
# This script provides functions to install, configure, and manage AlertManager server
# both natively and via Docker containers.
#

# Exit on any error, undefined variables, or pipe failures
set -o errexit
set -o nounset
set -o pipefail

# Environment variables for AlertManager configuration
# Can be overridden by setting these variables before running the script
PROJ_ALERTMANAGER_HOST=${PROJ_ALERTMANAGER_HOST:-127.0.0.1}                # AlertManager server host
PROJ_ALERTMANAGER_PORT=${PROJ_ALERTMANAGER_PORT:-9093}                     # AlertManager server port
PROJ_ALERTMANAGER_DATA_DIR=${PROJ_ALERTMANAGER_DATA_DIR:-/var/lib/alertmanager} # AlertManager data directory
PROJ_ALERTMANAGER_CONFIG_DIR=${PROJ_ALERTMANAGER_CONFIG_DIR:-/etc/alertmanager} # AlertManager config directory
# 加载通用配置和版本管理
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# 版本信息从统一配置文件加载：PROJ_ALERTMANAGER_VERSION 在 versions.sh 中定义
ALERTMANAGER_DOCKER_MNAME=${NETWORK_NAME}-alertmanager

# Function to install AlertManager natively
proj::alertmanager::install() {
  proj::alertmanager::pre_install

  # 创建 alertmanager 相关目录
  proj::util::sudo "mkdir -p ${PROJ_ALERTMANAGER_DATA_DIR}"
  proj::util::sudo "mkdir -p ${PROJ_ALERTMANAGER_CONFIG_DIR}"
  proj::util::sudo "mkdir -p /var/log/alertmanager"

  # 创建 alertmanager 用户
  if ! id -u alertmanager >/dev/null 2>&1; then
    proj::util::sudo "useradd --system --shell /bin/false alertmanager"
  fi

  # 设置正确的目录所有权
  proj::util::sudo "chown -R alertmanager:alertmanager ${PROJ_ALERTMANAGER_DATA_DIR}"
  proj::util::sudo "chown -R alertmanager:alertmanager ${PROJ_ALERTMANAGER_CONFIG_DIR}"
  proj::util::sudo "chown -R alertmanager:alertmanager /var/log/alertmanager"

  # 下载 AlertManager 二进制文件
  local alertmanager_download_dir="/tmp/alertmanager-download-test"
  mkdir -p ${alertmanager_download_dir}

  local alertmanager_arch="amd64"
  if [[ $(uname -m) == "aarch64" ]]; then
    alertmanager_arch="arm64"
  fi

  local alertmanager_url="https://github.com/prometheus/alertmanager/releases/download/v${PROJ_ALERTMANAGER_VERSION}/alertmanager-${PROJ_ALERTMANAGER_VERSION}.linux-${alertmanager_arch}.tar.gz"

  # 检查文件是否已存在，避免重复下载
  if [[ -f "${alertmanager_download_dir}/alertmanager.tar.gz" ]]; then
    proj::log::info "AlertManager ${PROJ_ALERTMANAGER_VERSION} tarball already exists, skipping download..."
  else
    proj::log::info "Downloading AlertManager ${PROJ_ALERTMANAGER_VERSION} for linux-${alertmanager_arch}..."
    curl -L ${alertmanager_url} -o ${alertmanager_download_dir}/alertmanager.tar.gz
  fi

  # 解压并安装
  tar xzf ${alertmanager_download_dir}/alertmanager.tar.gz -C ${alertmanager_download_dir} --strip-components=1
  proj::util::sudo "cp ${alertmanager_download_dir}/alertmanager /usr/local/bin/"
  proj::util::sudo "cp ${alertmanager_download_dir}/amtool /usr/local/bin/"
  proj::util::sudo "chmod +x /usr/local/bin/alertmanager"
  proj::util::sudo "chmod +x /usr/local/bin/amtool"

  # 创建 AlertManager 配置文件
  local alertmanager_conf_file="${PROJ_ALERTMANAGER_CONFIG_DIR}/alertmanager.yml"
  local temp_conf_file="/tmp/alertmanager.yml.tmp"

  cat > ${temp_conf_file} << EOF
global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'alertmanager@example.org'
  smtp_auth_username: 'alertmanager@example.org'
  smtp_auth_password: 'password'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'

receivers:
- name: 'web.hook'
  webhook_configs:
  - url: 'http://127.0.0.1:5001/'

inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'dev', 'instance']
EOF

  # 复制配置文件到系统目录
  proj::util::sudo "cp ${temp_conf_file} ${alertmanager_conf_file}"
  proj::util::sudo "chown alertmanager:alertmanager ${alertmanager_conf_file}"
  rm -f ${temp_conf_file}

  # 创建 systemd 服务文件
  local alertmanager_service_file="/etc/systemd/system/alertmanager.service"
  local temp_service_file="/tmp/alertmanager.service.tmp"

  cat > ${temp_service_file} << EOF
[Unit]
Description=AlertManager
Documentation=https://prometheus.io/docs/alerting/alertmanager/
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
User=alertmanager
Group=alertmanager
ExecReload=/bin/kill -HUP \$MAINPID
ExecStart=/usr/local/bin/alertmanager \\
  --config.file=${PROJ_ALERTMANAGER_CONFIG_DIR}/alertmanager.yml \\
  --storage.path=${PROJ_ALERTMANAGER_DATA_DIR} \\
  --web.listen-address=0.0.0.0:${PROJ_ALERTMANAGER_PORT} \\
  --web.external-url=

SyslogIdentifier=alertmanager
Restart=always

[Install]
WantedBy=multi-user.target
EOF

  # 复制服务文件到系统目录
  proj::util::sudo "cp ${temp_service_file} ${alertmanager_service_file}"
  rm -f ${temp_service_file}

  # 启动 AlertManager 服务
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "systemctl enable alertmanager"
  proj::util::sudo "systemctl start alertmanager"

  # 清理下载目录
  # rm -rf ${alertmanager_download_dir}

  sleep 5
  proj::alertmanager::status || return 1
  proj::alertmanager::info
  proj::log::info "install alertmanager successfully"
}

# Uninstall AlertManager step by step
proj::alertmanager::uninstall() {
  set +o errexit
  proj::util::sudo "systemctl stop alertmanager"
  proj::util::sudo "systemctl disable alertmanager"
  proj::util::sudo "rm -f /etc/systemd/system/alertmanager.service"
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "rm -f /usr/local/bin/alertmanager"
  proj::util::sudo "rm -f /usr/local/bin/amtool"
  proj::util::sudo "rm -rf ${PROJ_ALERTMANAGER_CONFIG_DIR}"
  proj::util::sudo "rm -rf ${PROJ_ALERTMANAGER_DATA_DIR}"
  proj::util::sudo "rm -rf /var/log/alertmanager"
  proj::util::sudo "userdel alertmanager" 2>/dev/null || true
  set -o errexit
  proj::log::info "uninstall alertmanager successfully"
  return 0
}

# Pre-install AlertManager by checking system requirements
proj::alertmanager::pre_install() {
  proj::log::info "Pre-installing AlertManager..."

  # 判断是 mac 还是 linux
  if proj::util::is_mac; then
    proj::log::info "Mac OS detected, checking for AlertManager installation..."
    if ! command -v alertmanager >/dev/null 2>&1; then
      proj::log::info "Installing AlertManager via brew..."
      brew install alertmanager
    else
      proj::log::info "AlertManager already installed, skipping..."
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

# Install AlertManager using a Docker container
proj::alertmanager::docker::install() {
  proj::log::info "Installing docker AlertManager..."

  proj::alertmanager::pre_install
  proj::common::network

  # 创建数据目录和配置目录
  local alertmanager_data_dir="${PROJ_THIRDPARTY_INSTALL_DIR}/alertmanager"
  local alertmanager_config_dir="${PROJ_THIRDPARTY_INSTALL_DIR}/alertmanager/config"
  mkdir -p ${alertmanager_data_dir}
  mkdir -p ${alertmanager_config_dir}

  # 创建配置文件
  cat > ${alertmanager_config_dir}/alertmanager.yml << EOF
global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'alertmanager@example.org'
  smtp_auth_username: 'alertmanager@example.org'
  smtp_auth_password: 'password'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'

receivers:
- name: 'web.hook'
  webhook_configs:
  - url: 'http://127.0.0.1:5001/'

inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'dev', 'instance']
EOF

  docker run -d --name ${ALERTMANAGER_DOCKER_MNAME} \
    --restart always \
    --network ${NETWORK_NAME} \
    -v ${alertmanager_data_dir}:/alertmanager \
    -v ${alertmanager_config_dir}/alertmanager.yml:/etc/alertmanager/alertmanager.yml \
    -p ${PROJ_ALERTMANAGER_HOST}:${PROJ_ALERTMANAGER_PORT}:9093 \
    prom/alertmanager:v${PROJ_ALERTMANAGER_VERSION} \
    --config.file=/etc/alertmanager/alertmanager.yml \
    --storage.path=/alertmanager \
    --web.listen-address=0.0.0.0:9093 \
    --web.external-url=

  sleep 5
  if proj::util::is_linux; then
    proj::alertmanager::status || return 1
  fi
  proj::alertmanager::info
  proj::log::info "install alertmanager successfully"
}

# Print necessary information after docker or native installation
proj::alertmanager::info() {
  echo -e ${C_GREEN}AlertManager has been installed, here are some useful information:${C_NORMAL}
  cat << EOF | sed 's/^/  /'
AlertManager access endpoint is: http://${PROJ_ALERTMANAGER_HOST}:${PROJ_ALERTMANAGER_PORT}
       AlertManager data dir: ${PROJ_ALERTMANAGER_DATA_DIR}
     AlertManager config dir: ${PROJ_ALERTMANAGER_CONFIG_DIR}
        AlertManager web UI: http://${PROJ_ALERTMANAGER_HOST}:${PROJ_ALERTMANAGER_PORT}
     AlertManager API status: http://${PROJ_ALERTMANAGER_HOST}:${PROJ_ALERTMANAGER_PORT}/api/v1/status
      AlertManager API alerts: http://${PROJ_ALERTMANAGER_HOST}:${PROJ_ALERTMANAGER_PORT}/api/v1/alerts
EOF
}

# Uninstall the docker container
proj::alertmanager::docker::uninstall() {
  docker rm -f ${ALERTMANAGER_DOCKER_MNAME} &>/dev/null
  proj::util::sudo "rm -rf ${PROJ_THIRDPARTY_INSTALL_DIR}/alertmanager"
  proj::log::info "uninstall alertmanager successfully"
}

# Status check after docker or native installation
proj::alertmanager::status() {
  proj::util::telnet ${PROJ_ALERTMANAGER_HOST} ${PROJ_ALERTMANAGER_PORT} || return 1

  # 检查 AlertManager 健康状态
  local health_check_url="http://${PROJ_ALERTMANAGER_HOST}:${PROJ_ALERTMANAGER_PORT}/-/healthy"
  if command -v curl >/dev/null 2>&1; then
    curl -f -s ${health_check_url} >/dev/null || {
      proj::log::error "AlertManager health check failed, AlertManager maybe not initialized properly."
      return 1
    }
  else
    proj::log::info "curl not found, skipping health check"
  fi
}

# This block allows calling functions in this script directly from the command line.
# It checks if the script's arguments contain a function name with the 'proj::alertmanager::' prefix
# and, if so, executes that function.
# For example: ./alertmanager.sh proj::alertmanager::install
if [[ "$*" =~ proj::alertmanager:: ]]; then
  eval $*
fi