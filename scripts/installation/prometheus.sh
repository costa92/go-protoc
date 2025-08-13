#!/usr/bin/env bash
#
# Prometheus Installation Script
# This script provides functions to install, configure, and manage Prometheus server
# both natively and via Docker containers.
#

# Exit on any error, undefined variables, or pipe failures
set -o errexit
set -o nounset
set -o pipefail

# 加载通用配置和版本管理
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Environment variables for Prometheus configuration
# Can be overridden by setting these variables before running the script
PROJ_PROMETHEUS_HOST=${PROJ_PROMETHEUS_HOST:-127.0.0.1}                # Prometheus server host
PROJ_PROMETHEUS_PORT=${PROJ_PROMETHEUS_PORT:-9090}                     # Prometheus server port
PROJ_PROMETHEUS_DATA_DIR=${PROJ_PROMETHEUS_DATA_DIR:-/var/lib/prometheus} # Prometheus data directory
PROJ_PROMETHEUS_CONFIG_DIR=${PROJ_PROMETHEUS_CONFIG_DIR:-/etc/prometheus} # Prometheus config directory
# 版本信息从统一配置文件加载：PROJ_PROMETHEUS_VERSION 在 versions.sh 中定义
PROMETHEUS_DOCKER_MNAME=${NETWORK_NAME}-prometheus

# Function to install Prometheus natively
proj::prometheus::install() {
  proj::prometheus::pre_install

  # 创建 prometheus 相关目录
  proj::util::sudo "mkdir -p ${PROJ_PROMETHEUS_DATA_DIR}"
  proj::util::sudo "mkdir -p ${PROJ_PROMETHEUS_CONFIG_DIR}"
  proj::util::sudo "mkdir -p /var/log/prometheus"
  
  # 创建 prometheus 用户
  if ! id -u prometheus >/dev/null 2>&1; then
    proj::util::sudo "useradd --system --shell /bin/false prometheus"
  fi
  
  # 设置正确的目录所有权
  proj::util::sudo "chown -R prometheus:prometheus ${PROJ_PROMETHEUS_DATA_DIR}"
  proj::util::sudo "chown -R prometheus:prometheus ${PROJ_PROMETHEUS_CONFIG_DIR}"
  proj::util::sudo "chown -R prometheus:prometheus /var/log/prometheus"

  # 下载 Prometheus 二进制文件
  local prometheus_download_dir="/tmp/prometheus-download-test"
  mkdir -p ${prometheus_download_dir}

  local prometheus_arch="amd64"
  if [[ $(uname -m) == "aarch64" ]]; then
    prometheus_arch="arm64"
  fi

  local prometheus_url="https://github.com/prometheus/prometheus/releases/download/v${PROJ_PROMETHEUS_VERSION}/prometheus-${PROJ_PROMETHEUS_VERSION}.linux-${prometheus_arch}.tar.gz"

  # 检查文件是否已存在，避免重复下载
  if [[ -f "${prometheus_download_dir}/prometheus.tar.gz" ]]; then
    proj::log::info "Prometheus ${PROJ_PROMETHEUS_VERSION} tarball already exists, skipping download..."
  else
    proj::log::info "Downloading Prometheus ${PROJ_PROMETHEUS_VERSION} for linux-${prometheus_arch}..."
    curl -L ${prometheus_url} -o ${prometheus_download_dir}/prometheus.tar.gz
  fi

  # 解压并安装
  tar xzf ${prometheus_download_dir}/prometheus.tar.gz -C ${prometheus_download_dir} --strip-components=1
  proj::util::sudo "cp ${prometheus_download_dir}/prometheus /usr/local/bin/"
  proj::util::sudo "cp ${prometheus_download_dir}/promtool /usr/local/bin/"
  proj::util::sudo "chmod +x /usr/local/bin/prometheus"
  proj::util::sudo "chmod +x /usr/local/bin/promtool"

  # 复制配置文件
  proj::util::sudo "cp -r ${prometheus_download_dir}/consoles ${PROJ_PROMETHEUS_CONFIG_DIR}/"
  proj::util::sudo "cp -r ${prometheus_download_dir}/console_libraries ${PROJ_PROMETHEUS_CONFIG_DIR}/"

  # 创建 Prometheus 配置文件
  local prometheus_conf_file="${PROJ_PROMETHEUS_CONFIG_DIR}/prometheus.yml"
  local temp_conf_file="/tmp/prometheus.yml.tmp"

  cat > ${temp_conf_file} << EOF
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  # - "first_rules.yml"
  # - "second_rules.yml"

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'node'
    static_configs:
      - targets: ['localhost:9100']

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          # - alertmanager:9093
EOF

  # 复制配置文件到系统目录
  proj::util::sudo "cp ${temp_conf_file} ${prometheus_conf_file}"
  proj::util::sudo "chown prometheus:prometheus ${prometheus_conf_file}"
  rm -f ${temp_conf_file}

  # 创建 systemd 服务文件
  local prometheus_service_file="/etc/systemd/system/prometheus.service"
  local temp_service_file="/tmp/prometheus.service.tmp"

  cat > ${temp_service_file} << EOF
[Unit]
Description=Prometheus
Documentation=https://prometheus.io/docs/introduction/overview/
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
User=prometheus
Group=prometheus
ExecReload=/bin/kill -HUP \$MAINPID
ExecStart=/usr/local/bin/prometheus \\
  --config.file=${PROJ_PROMETHEUS_CONFIG_DIR}/prometheus.yml \\
  --storage.tsdb.path=${PROJ_PROMETHEUS_DATA_DIR} \\
  --web.console.templates=${PROJ_PROMETHEUS_CONFIG_DIR}/consoles \\
  --web.console.libraries=${PROJ_PROMETHEUS_CONFIG_DIR}/console_libraries \\
  --web.listen-address=0.0.0.0:${PROJ_PROMETHEUS_PORT} \\
  --web.external-url=

SyslogIdentifier=prometheus
Restart=always

[Install]
WantedBy=multi-user.target
EOF

  # 复制服务文件到系统目录
  proj::util::sudo "cp ${temp_service_file} ${prometheus_service_file}"
  rm -f ${temp_service_file}

  # 启动 Prometheus 服务
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "systemctl enable prometheus"
  proj::util::sudo "systemctl start prometheus"

  # 清理下载目录
  # rm -rf ${prometheus_download_dir}

  sleep 5
  proj::prometheus::status || return 1
  proj::prometheus::info
  proj::log::info "install prometheus successfully"
}

# Uninstall Prometheus step by step
proj::prometheus::uninstall() {
  set +o errexit
  proj::util::sudo "systemctl stop prometheus"
  proj::util::sudo "systemctl disable prometheus"
  proj::util::sudo "rm -f /etc/systemd/system/prometheus.service"
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "rm -f /usr/local/bin/prometheus"
  proj::util::sudo "rm -f /usr/local/bin/promtool"
  proj::util::sudo "rm -rf ${PROJ_PROMETHEUS_CONFIG_DIR}"
  proj::util::sudo "rm -rf ${PROJ_PROMETHEUS_DATA_DIR}"
  proj::util::sudo "rm -rf /var/log/prometheus"
  proj::util::sudo "userdel prometheus" 2>/dev/null || true
  set -o errexit
  proj::log::info "uninstall prometheus successfully"
  return 0
}

# Pre-install Prometheus by checking system requirements
proj::prometheus::pre_install() {
  proj::log::info "Pre-installing Prometheus..."

  # 判断是 mac 还是 linux
  if proj::util::is_mac; then
    proj::log::info "Mac OS detected, checking for Prometheus installation..."
    if ! command -v prometheus >/dev/null 2>&1; then
      proj::log::info "Installing Prometheus via brew..."
      brew install prometheus
    else
      proj::log::info "Prometheus already installed, skipping..."
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

# Install Prometheus using a Docker container
proj::prometheus::docker::install() {
  proj::log::info "Installing docker Prometheus..."

  proj::prometheus::pre_install
  proj::common::network

  # 创建数据目录和配置目录
  local prometheus_data_dir="${PROJ_THIRDPARTY_INSTALL_DIR}/prometheus"
  local prometheus_config_dir="${PROJ_THIRDPARTY_INSTALL_DIR}/prometheus/config"
  mkdir -p ${prometheus_data_dir}
  mkdir -p ${prometheus_config_dir}

  # 创建配置文件
  cat > ${prometheus_config_dir}/prometheus.yml << EOF
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  # - "first_rules.yml"
  # - "second_rules.yml"

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'node'
    static_configs:
      - targets: ['localhost:9100']

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          # - alertmanager:9093
EOF

  docker run -d --name ${PROMETHEUS_DOCKER_MNAME} \
    --restart always \
    --network ${NETWORK_NAME} \
    -v ${prometheus_data_dir}:/prometheus \
    -v ${prometheus_config_dir}/prometheus.yml:/etc/prometheus/prometheus.yml \
    -p ${PROJ_PROMETHEUS_HOST}:${PROJ_PROMETHEUS_PORT}:9090 \
    prom/prometheus:v${PROJ_PROMETHEUS_VERSION} \
    --config.file=/etc/prometheus/prometheus.yml \
    --storage.tsdb.path=/prometheus \
    --web.console.templates=/etc/prometheus/consoles \
    --web.console.libraries=/etc/prometheus/console_libraries \
    --web.listen-address=0.0.0.0:9090 \
    --web.external-url=

  sleep 5
  if proj::util::is_linux; then
    proj::prometheus::status || return 1
  fi
  proj::prometheus::info
  proj::log::info "install prometheus successfully"
}

# Print necessary information after docker or native installation
proj::prometheus::info() {
  echo -e ${C_GREEN}Prometheus has been installed, here are some useful information:${C_NORMAL}
  cat << EOF | sed 's/^/  /'
Prometheus access endpoint is: http://${PROJ_PROMETHEUS_HOST}:${PROJ_PROMETHEUS_PORT}
       Prometheus data dir: ${PROJ_PROMETHEUS_DATA_DIR}
     Prometheus config dir: ${PROJ_PROMETHEUS_CONFIG_DIR}
        Prometheus web UI: http://${PROJ_PROMETHEUS_HOST}:${PROJ_PROMETHEUS_PORT}
      Prometheus metrics: http://${PROJ_PROMETHEUS_HOST}:${PROJ_PROMETHEUS_PORT}/metrics
       Prometheus targets: http://${PROJ_PROMETHEUS_HOST}:${PROJ_PROMETHEUS_PORT}/targets
EOF
}

# Uninstall the docker container
proj::prometheus::docker::uninstall() {
  docker rm -f ${PROMETHEUS_DOCKER_MNAME} &>/dev/null
  proj::util::sudo "rm -rf ${PROJ_THIRDPARTY_INSTALL_DIR}/prometheus"
  proj::log::info "uninstall prometheus successfully"
}

# Status check after docker or native installation
proj::prometheus::status() {
  proj::util::telnet ${PROJ_PROMETHEUS_HOST} ${PROJ_PROMETHEUS_PORT} || return 1

  # 检查 Prometheus 健康状态
  local health_check_url="http://${PROJ_PROMETHEUS_HOST}:${PROJ_PROMETHEUS_PORT}/-/healthy"
  if command -v curl >/dev/null 2>&1; then
    curl -f -s ${health_check_url} >/dev/null || {
      proj::log::error "Prometheus health check failed, Prometheus maybe not initialized properly."
      return 1
    }
  else
    proj::log::warning "curl not found, skipping health check"
  fi
}

# This block allows calling functions in this script directly from the command line.
# It checks if the script's arguments contain a function name with the 'proj::prometheus::' prefix
# and, if so, executes that function.
# For example: ./prometheus.sh proj::prometheus::install
if [[ "$*" =~ proj::prometheus:: ]]; then
  eval $*
fi