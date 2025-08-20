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
PROJ_OTELCOL_HTTP_PORT=${PROJ_OTELCOL_HTTP_PORT:-4328}           # OTLP HTTP receiver port (avoid conflict with Jaeger)
PROJ_OTELCOL_GRPC_PORT=${PROJ_OTELCOL_GRPC_PORT:-4327}           # OTLP gRPC receiver port (avoid conflict with Jaeger)
PROJ_OTELCOL_METRICS_PORT=${PROJ_OTELCOL_METRICS_PORT:-8888}     # Metrics port
PROJ_OTELCOL_HEALTH_PORT=${PROJ_OTELCOL_HEALTH_PORT:-13133}      # Health check port
PROJ_OTELCOL_DATA_DIR=${PROJ_OTELCOL_DATA_DIR:-/var/lib/otelcol} # OpenTelemetry Collector data directory
PROJ_OTELCOL_CONFIG_DIR=${PROJ_OTELCOL_CONFIG_DIR:-/etc/otelcol} # OpenTelemetry Collector config directory
# 加载通用配置和版本管理
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# 版本信息从统一配置文件加载：PROJ_OTELCOL_VERSION 在 versions.sh 中定义
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

  local otelcol_url="https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v${PROJ_OTELCOL_VERSION}/otelcol-contrib_${PROJ_OTELCOL_VERSION}_linux_${otelcol_arch}.tar.gz"

  # 检查文件是否已存在，避免重复下载
  if [[ -f "${otelcol_download_dir}/otelcol.tar.gz" ]]; then
    proj::log::info "OpenTelemetry Collector ${PROJ_OTELCOL_VERSION} tarball already exists, skipping download..."
  else
    proj::log::info "Downloading OpenTelemetry Collector ${PROJ_OTELCOL_VERSION} for linux-${otelcol_arch}..."
    curl -L ${otelcol_url} -o ${otelcol_download_dir}/otelcol.tar.gz
  fi

  # 解压并安装
  tar xzf ${otelcol_download_dir}/otelcol.tar.gz -C ${otelcol_download_dir}
  proj::util::sudo "cp ${otelcol_download_dir}/otelcol-contrib /usr/local/bin/"
  proj::util::sudo "chmod +x /usr/local/bin/otelcol-contrib"

  # 创建 OpenTelemetry Collector 配置文件（宿主机版本）
  local otelcol_conf_file="${PROJ_OTELCOL_CONFIG_DIR}/config.yaml"
  local template_conf_file="${SCRIPT_DIR}/otelcol/config.yaml"
  local temp_conf_file="/tmp/otelcol-config.yaml.tmp"

  # 使用 envsubst 替换模板中的环境变量
  envsubst < ${template_conf_file} > ${temp_conf_file}

  # 复制配置文件到系统目录
  proj::util::sudo "cp ${temp_conf_file} ${otelcol_conf_file}"
  proj::util::sudo "chown otelcol:otelcol ${otelcol_conf_file}"
  rm -f ${temp_conf_file}

  # 创建 systemd 服务文件
  local otelcol_service_file="/etc/systemd/system/otelcol.service"
  local template_service_file="${SCRIPT_DIR}/otelcol/otelcol.service"
  local temp_service_file="/tmp/otelcol.service.tmp"

  # 使用 envsubst 替换模板中的环境变量
  envsubst < ${template_service_file} > ${temp_service_file}

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
  proj::util::sudo "rm -f /usr/local/bin/otelcol-contrib"
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
    proj::log::info "Mac OS detected, Docker-only installation mode for OTEL Collector..."
    # macOS 环境下 Docker 安装不需要本地二进制文件，跳过 brew 安装
    # 只需要确保 Docker 可用和必要的工具存在
    if ! command -v docker >/dev/null 2>&1; then
      proj::log::error "Docker is required for macOS installation. Please install Docker first."
      return 1
    fi
    
    if ! command -v envsubst >/dev/null 2>&1; then
      proj::log::info "Installing gettext for envsubst..."
      if command -v brew >/dev/null 2>&1; then
        brew install gettext
        # 确保 envsubst 在 PATH 中
        if [[ ":$PATH:" != *":/opt/homebrew/opt/gettext/bin:"* ]]; then
          export PATH="/opt/homebrew/opt/gettext/bin:$PATH"
        fi
      else
        proj::log::error "Homebrew is required to install gettext. Please install Homebrew first."
        return 1
      fi
    fi
  else
    # Linux 环境检查必要的依赖
    if ! command -v curl >/dev/null 2>&1; then
      proj::log::info "Installing curl..."
      proj::util::sudo "apt update && apt install -y curl"
    fi

    if ! command -v tar >/dev/null 2>&1; then
      proj::log::info "Installing tar..."
      proj::util::sudo "apt install -y tar"
    fi

    if ! command -v envsubst >/dev/null 2>&1; then
      proj::log::info "Installing gettext-base for envsubst..."
      proj::util::sudo "apt install -y gettext-base"
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

  # 选择配置文件模板，使用标准化的 Docker 配置
  local template_conf_file="${SCRIPT_DIR}/otelcol/config-docker.yaml"
  
  # 检查配置文件是否存在
  if [[ ! -f "${template_conf_file}" ]]; then
    proj::log::error "Configuration template file not found: ${template_conf_file}"
    return 1
  fi
  
  # 检查是否有 Jaeger 容器运行，如果没有则使用独立配置
  if ! docker ps --format "table {{.Names}}" | grep -q "jaeger" 2>/dev/null; then
    proj::log::info "No Jaeger container found, using file monitoring Docker configuration..."
  else
    proj::log::info "Jaeger container detected, using file monitoring with networked Docker configuration..."
  fi
  
  # 使用 envsubst 替换模板中的环境变量
  export PROJ_SERVICE_NAME="${PROJ_SERVICE_NAME:-apiserver}"
  export PROJ_SERVICE_VERSION="${PROJ_SERVICE_VERSION:-v2.0.0}"
  export PROJ_ENVIRONMENT="${PROJ_ENVIRONMENT:-development}"
  
  # 检查 envsubst 是否可用
  if ! command -v envsubst >/dev/null 2>&1; then
    proj::log::error "envsubst command not found. Please install gettext package."
    return 1
  fi
  
  # 生成配置文件
  envsubst < ${template_conf_file} > ${otelcol_config_dir}/config.yaml
  
  # 验证配置文件生成成功
  if [[ ! -f "${otelcol_config_dir}/config.yaml" ]]; then
    proj::log::error "Failed to generate configuration file"
    return 1
  fi

  # 清理可能存在的同名容器
  proj::common::docker::cleanup_container "${OTELCOL_DOCKER_MNAME}"

  # 获取项目根目录以便映射日志文件
  local project_root="${SCRIPT_DIR}/../.."
  local logs_dir="$(cd "${project_root}/logs" 2>/dev/null && pwd)"
  
  # 确保日志目录存在
  if [[ -z "${logs_dir}" ]]; then
    logs_dir="${project_root}/logs"
    mkdir -p "${logs_dir}"
  fi
  
  # 启动容器
  proj::log::info "Starting OTEL Collector container with logs directory: ${logs_dir}"
  
  docker run -d --name ${OTELCOL_DOCKER_MNAME} \
    --restart always \
    --network ${NETWORK_NAME} \
    -v ${otelcol_config_dir}/config.yaml:/etc/otelcol-contrib/config.yaml:ro \
    -v ${logs_dir}:/host/logs:ro \
    -v ${otelcol_data_dir}:/var/log/otelcol \
    -p 127.0.0.1:4328:4328 \
    -p 127.0.0.1:4327:4327 \
    -p 127.0.0.1:8888:8888 \
    -p 127.0.0.1:13133:13133 \
    -e PROJ_SERVICE_NAME="${PROJ_SERVICE_NAME:-apiserver}" \
    -e PROJ_SERVICE_VERSION="${PROJ_SERVICE_VERSION:-v2.0.0}" \
    -e PROJ_ENVIRONMENT="${PROJ_ENVIRONMENT:-development}" \
    otel/opentelemetry-collector-contrib:${PROJ_OTELCOL_VERSION} \
    --config=/etc/otelcol-contrib/config.yaml

  # 等待容器启动
  proj::log::info "Waiting for OTEL Collector to start..."
  sleep 8
  
  # 检查容器状态
  if ! docker ps --filter "name=${OTELCOL_DOCKER_MNAME}" --filter "status=running" --format "{{.Names}}" | grep -q "^${OTELCOL_DOCKER_MNAME}$"; then
    proj::log::error "OTEL Collector container failed to start. Checking logs..."
    docker logs ${OTELCOL_DOCKER_MNAME} --tail 20 2>&1 || true
    return 1
  fi
  
  # 健康检查
  local max_retries=6
  local retry_count=0
  while [[ ${retry_count} -lt ${max_retries} ]]; do
    if proj::otelcol::status; then
      break
    fi
    retry_count=$((retry_count + 1))
    proj::log::info "Health check failed, retrying... (${retry_count}/${max_retries})"
    sleep 5
  done
  
  if [[ ${retry_count} -eq ${max_retries} ]]; then
    proj::log::error "OTEL Collector health check failed after ${max_retries} attempts"
    return 1
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
  # 检查 OpenTelemetry Collector 健康状态
  local health_check_url="http://${PROJ_OTELCOL_HOST}:${PROJ_OTELCOL_HEALTH_PORT}"
  
  if command -v curl >/dev/null 2>&1; then
    # 使用 curl 进行 HTTP 健康检查
    if curl -f -s -m 10 ${health_check_url} >/dev/null 2>&1; then
      return 0
    else
      proj::log::error "OpenTelemetry Collector health check failed at ${health_check_url}"
      return 1
    fi
  else
    proj::log::warn "curl not found, trying basic port check..."
    # 备用端口检查
    if proj::util::telnet ${PROJ_OTELCOL_HOST} ${PROJ_OTELCOL_HEALTH_PORT}; then
      return 0
    else
      proj::log::error "OpenTelemetry Collector port ${PROJ_OTELCOL_HEALTH_PORT} not accessible"
      return 1
    fi
  fi
}

# This block allows calling functions in this script directly from the command line.
# It checks if the script's arguments contain a function name with the 'proj::otelcol::' prefix
# and, if so, executes that function.
# For example: ./otelcol.sh proj::otelcol::install
if [[ "$*" =~ proj::otelcol:: ]]; then
  eval $*
fi