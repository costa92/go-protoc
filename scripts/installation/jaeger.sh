#!/usr/bin/env bash
#
# Jaeger Installation Script
# This script provides functions to install, configure, and manage Jaeger tracing server
# both natively and via Docker containers.
#

# Exit on any error, undefined variables, or pipe failures
set -o errexit
set -o nounset
set -o pipefail

# Environment variables for Jaeger configuration
# Can be overridden by setting these variables before running the script
PROJ_JAEGER_HOST=${PROJ_JAEGER_HOST:-127.0.0.1}                    # Jaeger server host
PROJ_JAEGER_UI_PORT=${PROJ_JAEGER_UI_PORT:-16686}                  # Jaeger UI port
PROJ_JAEGER_COLLECTOR_PORT=${PROJ_JAEGER_COLLECTOR_PORT:-14268}    # Jaeger collector port
PROJ_JAEGER_AGENT_PORT=${PROJ_JAEGER_AGENT_PORT:-6831}             # Jaeger agent port (UDP)
# 加载通用配置和版本管理
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# 版本信息从统一配置文件加载：PROJ_JAEGER_VERSION 在 versions.sh 中定义
PROJ_JAEGER_DATA_DIR=${PROJ_JAEGER_DATA_DIR:-/var/lib/jaeger}      # Jaeger data directory
JAEGER_DOCKER_MNAME=${NETWORK_NAME}-jaeger

# Function to install Jaeger natively
proj::jaeger::install() {
  proj::jaeger::pre_install

  # 创建 Jaeger 数据目录
  proj::util::sudo "mkdir -p ${PROJ_JAEGER_DATA_DIR}"
  proj::util::sudo "chmod 755 ${PROJ_JAEGER_DATA_DIR}"

  # 创建 jaeger 用户
  if ! id -u jaeger >/dev/null 2>&1; then
    proj::util::sudo "useradd --system --shell /bin/false jaeger"
  fi

  # 设置正确的目录所有权
  proj::util::sudo "chown -R jaeger:jaeger ${PROJ_JAEGER_DATA_DIR}"

  # 下载 Jaeger 二进制文件
  local jaeger_download_dir="/tmp/jaeger-download-test"
  mkdir -p ${jaeger_download_dir}

  local jaeger_arch="amd64"
  if [[ $(uname -m) == "aarch64" ]]; then
    jaeger_arch="arm64"
  fi

  local jaeger_url="https://github.com/jaegertracing/jaeger/releases/download/v${PROJ_JAEGER_VERSION}/jaeger-${PROJ_JAEGER_VERSION}-linux-${jaeger_arch}.tar.gz"

  # 检查文件是否已存在，避免重复下载
  if [[ -f "${jaeger_download_dir}/jaeger.tar.gz" ]]; then
    proj::log::info "Jaeger ${PROJ_JAEGER_VERSION} tarball already exists, skipping download..."
  else
    proj::log::info "Downloading Jaeger ${PROJ_JAEGER_VERSION} for linux-${jaeger_arch}..."
    curl -L ${jaeger_url} -o ${jaeger_download_dir}/jaeger.tar.gz
  fi

  # 解压并安装
  tar xzf ${jaeger_download_dir}/jaeger.tar.gz -C ${jaeger_download_dir} --strip-components=1
  proj::util::sudo "cp ${jaeger_download_dir}/jaeger-all-in-one /usr/local/bin/"
  proj::util::sudo "cp ${jaeger_download_dir}/jaeger-agent /usr/local/bin/"
  proj::util::sudo "cp ${jaeger_download_dir}/jaeger-collector /usr/local/bin/"
  proj::util::sudo "cp ${jaeger_download_dir}/jaeger-query /usr/local/bin/"
  proj::util::sudo "chmod +x /usr/local/bin/jaeger-*"

  # 创建 systemd 服务文件
  local jaeger_service_file="/etc/systemd/system/jaeger.service"

  # 创建临时文件
  local temp_service_file="/tmp/jaeger.service.tmp"
  cat > ${temp_service_file} << 'EOF'
[Unit]
Description=Jaeger Tracing Platform
Documentation=https://www.jaegertracing.io/
After=network.target

[Service]
Type=simple
User=jaeger
ExecStart=/usr/local/bin/jaeger-all-in-one \
  --collector.grpc-server.host-port=0.0.0.0:14250 \
  --collector.http-server.host-port=0.0.0.0:14268 \
  --query.http-server.host-port=0.0.0.0:16686 \
  --processor.jaeger-compact.server-host-port=0.0.0.0:6831 \
  --memory.max-traces=50000 \
  --log-level=info
Restart=always
RestartSec=10s
LimitNOFILE=40000

[Install]
WantedBy=multi-user.target
EOF

  # 复制到系统目录
  proj::util::sudo "cp ${temp_service_file} ${jaeger_service_file}"
  rm -f ${temp_service_file}

  # 启动 Jaeger 服务
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "systemctl enable jaeger"
  proj::util::sudo "systemctl start jaeger"

  # 清理下载目录
  # rm -rf ${jaeger_download_dir}

  sleep 3
  proj::jaeger::status || return 1
  proj::jaeger::info
  proj::log::info "install jaeger successfully"
}

# Uninstall Jaeger step by step
proj::jaeger::uninstall() {
  set +o errexit
  proj::util::sudo "systemctl stop jaeger"
  proj::util::sudo "systemctl disable jaeger"
  proj::util::sudo "rm -f /etc/systemd/system/jaeger.service"
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "rm -f /usr/local/bin/jaeger-*"
  proj::util::sudo "rm -rf ${PROJ_JAEGER_DATA_DIR}"
  proj::util::sudo "userdel jaeger" 2>/dev/null || true
  set -o errexit
  proj::log::info "uninstall jaeger successfully"
  return 0
}

# Pre-install Jaeger by checking system requirements
proj::jaeger::pre_install() {
  proj::log::info "Pre-installing Jaeger..."

  # 判断是 mac 还是 linux
  if proj::util::is_mac; then
    proj::log::info "Mac OS detected"
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

# Install Jaeger using a Docker container
proj::jaeger::docker::install() {
  proj::log::info "Installing docker Jaeger..."

  proj::jaeger::pre_install
  proj::common::network

  # 创建数据目录
  local jaeger_data_dir="${PROJ_THIRDPARTY_INSTALL_DIR}/jaeger"
  proj::util::sudo "mkdir -p ${jaeger_data_dir}"

  docker run -d --name ${JAEGER_DOCKER_MNAME} \
    --restart always \
    --network ${NETWORK_NAME} \
    -p ${PROJ_JAEGER_HOST}:${PROJ_JAEGER_UI_PORT}:16686 \
    -p ${PROJ_JAEGER_HOST}:${PROJ_JAEGER_COLLECTOR_PORT}:14268 \
    -p ${PROJ_JAEGER_HOST}:14250:14250 \
    -p ${PROJ_JAEGER_HOST}:${PROJ_JAEGER_AGENT_PORT}:6831/udp \
    -e COLLECTOR_OTLP_ENABLED=true \
    jaegertracing/all-in-one:${PROJ_JAEGER_VERSION}

  sleep 3
  if proj::util::is_linux; then
    proj::jaeger::status || return 1
  fi
  proj::jaeger::info
  proj::log::info "install jaeger successfully"
}

# Print necessary information after docker or native installation
proj::jaeger::info() {
  echo -e ${C_GREEN}Jaeger has been installed, here are some useful information:${C_NORMAL}
  cat << EOF | sed 's/^/  /'
Jaeger UI endpoint is: http://${PROJ_JAEGER_HOST}:${PROJ_JAEGER_UI_PORT}
    Jaeger collector: http://${PROJ_JAEGER_HOST}:${PROJ_JAEGER_COLLECTOR_PORT}
       Jaeger agent: ${PROJ_JAEGER_HOST}:${PROJ_JAEGER_AGENT_PORT}/udp
      Jaeger health: curl http://${PROJ_JAEGER_HOST}:${PROJ_JAEGER_COLLECTOR_PORT}/health
     Access Jaeger UI: Open http://${PROJ_JAEGER_HOST}:${PROJ_JAEGER_UI_PORT} in your browser
EOF
}

# Uninstall the docker container
proj::jaeger::docker::uninstall() {
  docker rm -f ${JAEGER_DOCKER_MNAME} &>/dev/null
  proj::util::sudo "rm -rf ${PROJ_THIRDPARTY_INSTALL_DIR}/jaeger"
  proj::log::info "uninstall jaeger successfully"
}

# Status check after docker or native installation
proj::jaeger::status() {
  proj::util::telnet ${PROJ_JAEGER_HOST} ${PROJ_JAEGER_UI_PORT} || return 1

  # 检查 Jaeger UI 可访问性
  if command -v curl >/dev/null 2>&1; then
    curl -f http://${PROJ_JAEGER_HOST}:${PROJ_JAEGER_UI_PORT} >/dev/null 2>&1 || {
      proj::log::error "Jaeger UI access check failed, Jaeger maybe not initialized properly."
      return 1
    }
  else
    proj::log::info "curl not found, skipping health check"
  fi
}

# This block allows calling functions in this script directly from the command line.
# It checks if the script's arguments contain a function name with the 'proj::jaeger::' prefix
# and, if so, executes that function.
# For example: ./jaeger.sh proj::jaeger::install
if [[ "$*" =~ proj::jaeger:: ]]; then
  eval $*
fi