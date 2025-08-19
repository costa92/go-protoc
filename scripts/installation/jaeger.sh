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
PROJ_JAEGER_QUERY_PORT=${PROJ_JAEGER_QUERY_PORT:-16686}            # Jaeger query/UI port
PROJ_JAEGER_COLLECTOR_HTTP_PORT=${PROJ_JAEGER_COLLECTOR_HTTP_PORT:-14268}  # Collector HTTP port
PROJ_JAEGER_COLLECTOR_GRPC_PORT=${PROJ_JAEGER_COLLECTOR_GRPC_PORT:-14250}  # Collector gRPC port
PROJ_JAEGER_AGENT_HTTP_PORT=${PROJ_JAEGER_AGENT_HTTP_PORT:-5778}    # Agent HTTP port
PROJ_JAEGER_AGENT_COMPACT_PORT=${PROJ_JAEGER_AGENT_COMPACT_PORT:-6831} # Agent compact thrift port
PROJ_JAEGER_AGENT_BINARY_PORT=${PROJ_JAEGER_AGENT_BINARY_PORT:-6832} # Agent binary thrift port
PROJ_JAEGER_CONFIG_DIR=${PROJ_JAEGER_CONFIG_DIR:-/etc/jaeger}       # Jaeger config directory
# 加载通用配置和版本管理
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# 版本信息从统一配置文件加载：PROJ_JAEGER_VERSION 在 versions.sh 中定义
PROJ_JAEGER_DATA_DIR=${PROJ_JAEGER_DATA_DIR:-/var/lib/jaeger}      # Jaeger data directory
JAEGER_DOCKER_MNAME=${NETWORK_NAME}-jaeger

# Function to install Jaeger natively
proj::jaeger::install() {
  proj::jaeger::pre_install

  # 创建 Jaeger 相关目录
  proj::util::sudo "mkdir -p ${PROJ_JAEGER_DATA_DIR}"
  proj::util::sudo "mkdir -p ${PROJ_JAEGER_CONFIG_DIR}"
  proj::util::sudo "mkdir -p /var/log/jaeger"
  proj::util::sudo "chmod 755 ${PROJ_JAEGER_DATA_DIR}"

  # 创建 jaeger 用户
  if ! id -u jaeger >/dev/null 2>&1; then
    proj::util::sudo "useradd --system --shell /bin/false jaeger"
  fi

  # 设置正确的目录所有权
  proj::util::sudo "chown -R jaeger:jaeger ${PROJ_JAEGER_DATA_DIR}"
  proj::util::sudo "chown -R jaeger:jaeger ${PROJ_JAEGER_CONFIG_DIR}"
  proj::util::sudo "chown -R jaeger:jaeger /var/log/jaeger"

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

  # 创建 Jaeger 配置文件
  local jaeger_conf_file="${PROJ_JAEGER_CONFIG_DIR}/jaeger-config.yaml"
  local template_conf_file="${SCRIPT_DIR}/jaeger/jaeger-config.yaml"
  local temp_conf_file="/tmp/jaeger-config.yaml.tmp"

  # 使用 envsubst 替换模板中的环境变量
  envsubst < ${template_conf_file} > ${temp_conf_file}

  # 复制配置文件到系统目录
  proj::util::sudo "cp ${temp_conf_file} ${jaeger_conf_file}"
  proj::util::sudo "chown jaeger:jaeger ${jaeger_conf_file}"
  rm -f ${temp_conf_file}

  # 创建 systemd 服务文件
  local jaeger_service_file="/etc/systemd/system/jaeger.service"
  local template_service_file="${SCRIPT_DIR}/jaeger/jaeger.service"
  local temp_service_file="/tmp/jaeger.service.tmp"

  # 使用 envsubst 替换模板中的环境变量
  envsubst < ${template_service_file} > ${temp_service_file}

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

    # 检查 envsubst (gettext-base)
    if ! command -v envsubst >/dev/null 2>&1; then
      proj::log::info "Installing gettext-base for envsubst..."
      proj::util::sudo "apt install -y gettext-base"
    fi
  fi
}

# Install Jaeger using a Docker container
proj::jaeger::docker::install() {
  proj::log::info "Installing docker Jaeger..."

  proj::jaeger::pre_install
  proj::common::network

  # 创建数据目录和配置目录
  local jaeger_data_dir="${PROJ_THIRDPARTY_INSTALL_DIR}/jaeger"
  local jaeger_config_dir="${PROJ_THIRDPARTY_INSTALL_DIR}/jaeger/config"
  proj::util::sudo "mkdir -p ${jaeger_data_dir}"
  proj::util::sudo "mkdir -p ${jaeger_config_dir}"

  # 创建 Docker 配置文件
  local template_conf_file="${SCRIPT_DIR}/jaeger/jaeger-config-docker.yaml"
  local temp_conf_file="/tmp/jaeger-config-docker.yaml.tmp"

  # 使用 envsubst 替换模板中的环境变量
  envsubst < ${template_conf_file} > ${temp_conf_file}
  proj::util::sudo "cp ${temp_conf_file} ${jaeger_config_dir}/jaeger-config.yaml"
  rm -f ${temp_conf_file}

  # 清理可能存在的同名容器
  proj::common::docker::cleanup_container "${JAEGER_DOCKER_MNAME}"

  docker run -d --name ${JAEGER_DOCKER_MNAME} \
    --restart always \
    --network ${NETWORK_NAME} \
    -p ${PROJ_JAEGER_HOST}:${PROJ_JAEGER_QUERY_PORT}:16686 \
    -p ${PROJ_JAEGER_HOST}:${PROJ_JAEGER_COLLECTOR_HTTP_PORT}:14268 \
    -p ${PROJ_JAEGER_HOST}:${PROJ_JAEGER_COLLECTOR_GRPC_PORT}:14250 \
    -p ${PROJ_JAEGER_HOST}:${PROJ_JAEGER_AGENT_HTTP_PORT}:5778 \
    -p ${PROJ_JAEGER_HOST}:${PROJ_JAEGER_AGENT_COMPACT_PORT}:6831/udp \
    -p ${PROJ_JAEGER_HOST}:${PROJ_JAEGER_AGENT_BINARY_PORT}:6832/udp \
    -v ${jaeger_config_dir}/jaeger-config.yaml:/etc/jaeger/jaeger-config.yaml \
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
Jaeger UI endpoint is: http://${PROJ_JAEGER_HOST}:${PROJ_JAEGER_QUERY_PORT}
    Jaeger collector HTTP: http://${PROJ_JAEGER_HOST}:${PROJ_JAEGER_COLLECTOR_HTTP_PORT}
    Jaeger collector gRPC: http://${PROJ_JAEGER_HOST}:${PROJ_JAEGER_COLLECTOR_GRPC_PORT}
       Jaeger agent HTTP: http://${PROJ_JAEGER_HOST}:${PROJ_JAEGER_AGENT_HTTP_PORT}
       Jaeger agent UDP: ${PROJ_JAEGER_HOST}:${PROJ_JAEGER_AGENT_COMPACT_PORT}/udp
      Jaeger health check: curl http://${PROJ_JAEGER_HOST}:${PROJ_JAEGER_COLLECTOR_HTTP_PORT}/health
     Access Jaeger UI: Open http://${PROJ_JAEGER_HOST}:${PROJ_JAEGER_QUERY_PORT} in your browser
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
  proj::util::telnet ${PROJ_JAEGER_HOST} ${PROJ_JAEGER_QUERY_PORT} || return 1

  # 检查 Jaeger UI 可访问性
  if command -v curl >/dev/null 2>&1; then
    curl -f http://${PROJ_JAEGER_HOST}:${PROJ_JAEGER_QUERY_PORT} >/dev/null 2>&1 || {
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