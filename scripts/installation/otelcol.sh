#!/usr/bin/env bash
#
# OpenTelemetry Collector Installation Script
# This script provides functions to install, configure, and manage OpenTelemetry Collector
# both natively and via Docker containers.
#

# Exit on any error, undefined variables, or pipe failures
set -o errexit
set -o nounset
set -o pipefail

# Environment variables for OpenTelemetry Collector configuration
# Can be overridden by setting these variables before running the script
PROJ_OTELCOL_HOST=${PROJ_OTELCOL_HOST:-127.0.0.1}                # OpenTelemetry Collector server host
PROJ_OTELCOL_HTTP_PORT=${PROJ_OTELCOL_HTTP_PORT:-4318}           # OTLP HTTP receiver port
PROJ_OTELCOL_GRPC_PORT=${PROJ_OTELCOL_GRPC_PORT:-4317}           # OTLP gRPC receiver port
PROJ_OTELCOL_METRICS_PORT=${PROJ_OTELCOL_METRICS_PORT:-8888}     # Metrics port
PROJ_OTELCOL_HEALTH_PORT=${PROJ_OTELCOL_HEALTH_PORT:-13133}      # Health check port
PROJ_OTELCOL_DATA_DIR=${PROJ_OTELCOL_DATA_DIR:-/var/lib/otelcol} # OpenTelemetry Collector data directory
PROJ_OTELCOL_CONFIG_DIR=${PROJ_OTELCOL_CONFIG_DIR:-/etc/otelcol} # OpenTelemetry Collector config directory
PROJ_OTELCOL_VERSION=${PROJ_OTELCOL_VERSION:-0.91.0}             # OpenTelemetry Collector version
OTELCOL_DOCKER_MNAME=${NETWORK_NAME}-otelcol

# Function to install OpenTelemetry Collector natively
proj::otelcol::install() {
  proj::otelcol::pre_install

  # 创建 otelcol 相关目录
  proj::util::sudo "mkdir -p ${PROJ_OTELCOL_DATA_DIR}"
  proj::util::sudo "mkdir -p ${PROJ_OTELCOL_CONFIG_DIR}"
  proj::util::sudo "mkdir -p /var/log/otelcol"
  
  # 创建 otelcol 用户
  if ! id -u otelcol >/dev/null 2>&1; then
    proj::util::sudo "useradd --system --shell /bin/false otelcol"
  fi
  
  # 设置正确的目录所有权
  proj::util::sudo "chown -R otelcol:otelcol ${PROJ_OTELCOL_DATA_DIR}"
  proj::util::sudo "chown -R otelcol:otelcol ${PROJ_OTELCOL_CONFIG_DIR}"
  proj::util::sudo "chown -R otelcol:otelcol /var/log/otelcol"

  # 下载 OpenTelemetry Collector 二进制文件
  local otelcol_download_dir="/tmp/otelcol-download-test"
  mkdir -p ${otelcol_download_dir}

  local otelcol_arch="amd64"
  if [[ $(uname -m) == "aarch64" ]]; then
    otelcol_arch="arm64"
  fi

  local otelcol_url="https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v${PROJ_OTELCOL_VERSION}/otelcol_${PROJ_OTELCOL_VERSION}_linux_${otelcol_arch}.tar.gz"

  # 检查文件是否已存在，避免重复下载
  if [[ -f "${otelcol_download_dir}/otelcol.tar.gz" ]]; then
    proj::log::info "OpenTelemetry Collector ${PROJ_OTELCOL_VERSION} tarball already exists, skipping download..."
  else
    proj::log::info "Downloading OpenTelemetry Collector ${PROJ_OTELCOL_VERSION} for linux-${otelcol_arch}..."
    curl -L ${otelcol_url} -o ${otelcol_download_dir}/otelcol.tar.gz
  fi

  # 解压并安装
  tar xzf ${otelcol_download_dir}/otelcol.tar.gz -C ${otelcol_download_dir}
  proj::util::sudo "cp ${otelcol_download_dir}/otelcol /usr/local/bin/"
  proj::util::sudo "chmod +x /usr/local/bin/otelcol"

  # 创建 OpenTelemetry Collector 配置文件
  local otelcol_conf_file="${PROJ_OTELCOL_CONFIG_DIR}/config.yaml"
  local temp_conf_file="/tmp/otelcol-config.yaml.tmp"

  cat > ${temp_conf_file} << EOF
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:${PROJ_OTELCOL_GRPC_PORT}
      http:
        endpoint: 0.0.0.0:${PROJ_OTELCOL_HTTP_PORT}
  
  prometheus:
    config:
      scrape_configs:
        - job_name: 'otelcol'
          scrape_interval: 10s
          static_configs:
            - targets: ['0.0.0.0:8888']

processors:
  batch:
    timeout: 1s
    send_batch_size: 1024
  
  memory_limiter:
    limit_mib: 512

exporters:
  logging:
    loglevel: debug
  
  prometheus:
    endpoint: "0.0.0.0:8889"
  
  jaeger:
    endpoint: jaeger:14250
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [logging, jaeger]
    
    metrics:
      receivers: [otlp, prometheus]
      processors: [memory_limiter, batch]
      exporters: [logging, prometheus]
    
    logs:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [logging]

  extensions: [health_check]
  
extensions:
  health_check:
    endpoint: 0.0.0.0:${PROJ_OTELCOL_HEALTH_PORT}
EOF

  # 复制配置文件到系统目录
  proj::util::sudo "cp ${temp_conf_file} ${otelcol_conf_file}"
  proj::util::sudo "chown otelcol:otelcol ${otelcol_conf_file}"
  rm -f ${temp_conf_file}

  # 创建 systemd 服务文件
  local otelcol_service_file="/etc/systemd/system/otelcol.service"
  local temp_service_file="/tmp/otelcol.service.tmp"

  cat > ${temp_service_file} << EOF
[Unit]
Description=OpenTelemetry Collector
Documentation=https://opentelemetry.io/docs/collector/
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
User=otelcol
Group=otelcol
ExecReload=/bin/kill -HUP \$MAINPID
ExecStart=/usr/local/bin/otelcol --config=${PROJ_OTELCOL_CONFIG_DIR}/config.yaml

SyslogIdentifier=otelcol
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

  # 复制服务文件到系统目录
  proj::util::sudo "cp ${temp_service_file} ${otelcol_service_file}"
  rm -f ${temp_service_file}

  # 启动 OpenTelemetry Collector 服务
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "systemctl enable otelcol"
  proj::util::sudo "systemctl start otelcol"

  # 清理下载目录
  # rm -rf ${otelcol_download_dir}

  sleep 5
  proj::otelcol::status || return 1
  proj::otelcol::info
  proj::log::info "install otelcol successfully"
}

# Uninstall OpenTelemetry Collector step by step
proj::otelcol::uninstall() {
  set +o errexit
  proj::util::sudo "systemctl stop otelcol"
  proj::util::sudo "systemctl disable otelcol"
  proj::util::sudo "rm -f /etc/systemd/system/otelcol.service"
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "rm -f /usr/local/bin/otelcol"
  proj::util::sudo "rm -rf ${PROJ_OTELCOL_CONFIG_DIR}"
  proj::util::sudo "rm -rf ${PROJ_OTELCOL_DATA_DIR}"
  proj::util::sudo "rm -rf /var/log/otelcol"
  proj::util::sudo "userdel otelcol" 2>/dev/null || true
  set -o errexit
  proj::log::info "uninstall otelcol successfully"
  return 0
}

# Pre-install OpenTelemetry Collector by checking system requirements
proj::otelcol::pre_install() {
  proj::log::info "Pre-installing OpenTelemetry Collector..."

  # 判断是 mac 还是 linux
  if proj::util::is_mac; then
    proj::log::info "Mac OS detected, checking for OpenTelemetry Collector installation..."
    if ! command -v otelcol >/dev/null 2>&1; then
      proj::log::info "Installing OpenTelemetry Collector via brew..."
      brew install opentelemetry-collector
    else
      proj::log::info "OpenTelemetry Collector already installed, skipping..."
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

# Install OpenTelemetry Collector using a Docker container
proj::otelcol::docker::install() {
  proj::log::info "Installing docker OpenTelemetry Collector..."

  proj::otelcol::pre_install
  proj::common::network

  # 创建数据目录和配置目录
  local otelcol_data_dir="${PROJ_THIRDPARTY_INSTALL_DIR}/otelcol"
  local otelcol_config_dir="${PROJ_THIRDPARTY_INSTALL_DIR}/otelcol/config"
  mkdir -p ${otelcol_data_dir}
  mkdir -p ${otelcol_config_dir}

  # 创建配置文件
  cat > ${otelcol_config_dir}/config.yaml << EOF
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:${PROJ_OTELCOL_GRPC_PORT}
      http:
        endpoint: 0.0.0.0:${PROJ_OTELCOL_HTTP_PORT}
  
  prometheus:
    config:
      scrape_configs:
        - job_name: 'otelcol'
          scrape_interval: 10s
          static_configs:
            - targets: ['0.0.0.0:8888']

processors:
  batch:
    timeout: 1s
    send_batch_size: 1024
  
  memory_limiter:
    limit_mib: 512

exporters:
  logging:
    loglevel: debug
  
  prometheus:
    endpoint: "0.0.0.0:8889"
  
  jaeger:
    endpoint: jaeger:14250
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [logging, jaeger]
    
    metrics:
      receivers: [otlp, prometheus]
      processors: [memory_limiter, batch]
      exporters: [logging, prometheus]
    
    logs:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [logging]

  extensions: [health_check]
  
extensions:
  health_check:
    endpoint: 0.0.0.0:${PROJ_OTELCOL_HEALTH_PORT}
EOF

  docker run -d --name ${OTELCOL_DOCKER_MNAME} \
    --restart always \
    --network ${NETWORK_NAME} \
    -v ${otelcol_config_dir}/config.yaml:/etc/otelcol-contrib/config.yaml \
    -p ${PROJ_OTELCOL_HOST}:${PROJ_OTELCOL_HTTP_PORT}:4318 \
    -p ${PROJ_OTELCOL_HOST}:${PROJ_OTELCOL_GRPC_PORT}:4317 \
    -p ${PROJ_OTELCOL_HOST}:${PROJ_OTELCOL_METRICS_PORT}:8888 \
    -p ${PROJ_OTELCOL_HOST}:${PROJ_OTELCOL_HEALTH_PORT}:13133 \
    otel/opentelemetry-collector-contrib:${PROJ_OTELCOL_VERSION} \
    --config=/etc/otelcol-contrib/config.yaml

  sleep 5
  if proj::util::is_linux; then
    proj::otelcol::status || return 1
  fi
  proj::otelcol::info
  proj::log::info "install otelcol successfully"
}

# Print necessary information after docker or native installation
proj::otelcol::info() {
  echo -e ${C_GREEN}OpenTelemetry Collector has been installed, here are some useful information:${C_NORMAL}
  cat << EOF | sed 's/^/  /'
OpenTelemetry Collector OTLP HTTP endpoint: http://${PROJ_OTELCOL_HOST}:${PROJ_OTELCOL_HTTP_PORT}
OpenTelemetry Collector OTLP gRPC endpoint: ${PROJ_OTELCOL_HOST}:${PROJ_OTELCOL_GRPC_PORT}
       OpenTelemetry Collector data dir: ${PROJ_OTELCOL_DATA_DIR}
     OpenTelemetry Collector config dir: ${PROJ_OTELCOL_CONFIG_DIR}
    OpenTelemetry Collector metrics port: ${PROJ_OTELCOL_METRICS_PORT}
     OpenTelemetry Collector health check: http://${PROJ_OTELCOL_HOST}:${PROJ_OTELCOL_HEALTH_PORT}
EOF
}

# Uninstall the docker container
proj::otelcol::docker::uninstall() {
  docker rm -f ${OTELCOL_DOCKER_MNAME} &>/dev/null
  proj::util::sudo "rm -rf ${PROJ_THIRDPARTY_INSTALL_DIR}/otelcol"
  proj::log::info "uninstall otelcol successfully"
}

# Status check after docker or native installation
proj::otelcol::status() {
  proj::util::telnet ${PROJ_OTELCOL_HOST} ${PROJ_OTELCOL_HEALTH_PORT} || return 1

  # 检查 OpenTelemetry Collector 健康状态
  local health_check_url="http://${PROJ_OTELCOL_HOST}:${PROJ_OTELCOL_HEALTH_PORT}"
  if command -v curl >/dev/null 2>&1; then
    curl -f -s ${health_check_url} >/dev/null || {
      proj::log::error "OpenTelemetry Collector health check failed, OpenTelemetry Collector maybe not initialized properly."
      return 1
    }
  else
    proj::log::warning "curl not found, skipping health check"
  fi
}

# This block allows calling functions in this script directly from the command line.
# It checks if the script's arguments contain a function name with the 'proj::otelcol::' prefix
# and, if so, executes that function.
# For example: ./otelcol.sh proj::otelcol::install
if [[ "$*" =~ proj::otelcol:: ]]; then
  eval $*
fi