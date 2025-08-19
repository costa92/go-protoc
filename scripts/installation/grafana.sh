#!/usr/bin/env bash
#
# Grafana Installation Script
# This script provides functions to install, configure, and manage Grafana server
# both natively and via Docker containers.
#

# Exit on any error, undefined variables, or pipe failures
set -o errexit
set -o nounset
set -o pipefail

# 加载通用配置和版本管理
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Environment variables for Grafana configuration
# Can be overridden by setting these variables before running the script
PROJ_GRAFANA_HOST=${PROJ_GRAFANA_HOST:-127.0.0.1}                # Grafana server host
PROJ_GRAFANA_PORT=${PROJ_GRAFANA_PORT:-3000}                     # Grafana server port
PROJ_GRAFANA_ADMIN_USER=${PROJ_GRAFANA_ADMIN_USER:-admin}        # Grafana admin username
PROJ_GRAFANA_ADMIN_PASSWORD=${PROJ_GRAFANA_ADMIN_PASSWORD:-proj(#)666}  # Grafana admin password
PROJ_GRAFANA_DATA_DIR=${PROJ_GRAFANA_DATA_DIR:-/var/lib/grafana} # Grafana data directory
# 版本信息从统一配置文件加载：PROJ_GRAFANA_VERSION 在 versions.sh 中定义
GRAFANA_DOCKER_MNAME=${NETWORK_NAME}-grafana

# Function to install Grafana natively
proj::grafana::install() {
  proj::grafana::pre_install

  # 创建 grafana 数据目录
  proj::util::sudo "mkdir -p ${PROJ_GRAFANA_DATA_DIR}"
  proj::util::sudo "mkdir -p /etc/grafana"
  proj::util::sudo "mkdir -p /var/log/grafana"

  # 创建 grafana 用户
  if ! id -u grafana >/dev/null 2>&1; then
    proj::util::sudo "useradd --system --shell /bin/false grafana"
  fi

  # 设置正确的目录所有权
  proj::util::sudo "chown -R grafana:grafana ${PROJ_GRAFANA_DATA_DIR}"
  proj::util::sudo "chown -R grafana:grafana /var/log/grafana"

  # 下载 Grafana 二进制文件
  local grafana_download_dir="/tmp/grafana-download-test"
  mkdir -p ${grafana_download_dir}

  local grafana_arch="amd64"
  if [[ $(uname -m) == "aarch64" ]]; then
    grafana_arch="arm64"
  fi

  local grafana_url="https://dl.grafana.com/oss/release/grafana-${PROJ_GRAFANA_VERSION}.linux-${grafana_arch}.tar.gz"

  # 检查文件是否已存在，避免重复下载
  if [[ -f "${grafana_download_dir}/grafana.tar.gz" ]]; then
    proj::log::info "Grafana ${PROJ_GRAFANA_VERSION} tarball already exists, skipping download..."
  else
    proj::log::info "Downloading Grafana ${PROJ_GRAFANA_VERSION} for linux-${grafana_arch}..."
    curl -L ${grafana_url} -o ${grafana_download_dir}/grafana.tar.gz
  fi

  # 解压并安装
  tar xzf ${grafana_download_dir}/grafana.tar.gz -C ${grafana_download_dir} --strip-components=1
  proj::util::sudo "cp -r ${grafana_download_dir}/bin /usr/share/grafana/"
  proj::util::sudo "cp -r ${grafana_download_dir}/conf /usr/share/grafana/"
  proj::util::sudo "cp -r ${grafana_download_dir}/public /usr/share/grafana/"
  proj::util::sudo "chmod +x /usr/share/grafana/bin/grafana"
  proj::util::sudo "ln -sf /usr/share/grafana/bin/grafana /usr/local/bin/grafana"

  # 创建 Grafana 配置文件
  local grafana_conf_file="/etc/grafana/grafana.ini"
  local template_conf_file="${SCRIPT_DIR}/grafana/grafana.ini"
  local temp_conf_file="/tmp/grafana.ini.tmp"

  # 使用 envsubst 替换模板中的环境变量
  envsubst < ${template_conf_file} > ${temp_conf_file}

  # 复制配置文件到系统目录
  proj::util::sudo "cp ${temp_conf_file} ${grafana_conf_file}"
  proj::util::sudo "chown grafana:grafana ${grafana_conf_file}"
  rm -f ${temp_conf_file}

  # 创建 systemd 服务文件
  local grafana_service_file="/etc/systemd/system/grafana.service"
  local template_service_file="${SCRIPT_DIR}/grafana/grafana.service"
  local temp_service_file="/tmp/grafana.service.tmp"

  # 使用 envsubst 替换模板中的环境变量
  envsubst < ${template_service_file} > ${temp_service_file}

  # 复制服务文件到系统目录
  proj::util::sudo "cp ${temp_service_file} ${grafana_service_file}"
  rm -f ${temp_service_file}

  # 创建环境文件
  local grafana_env_file="/etc/default/grafana-server"
  local template_env_file="${SCRIPT_DIR}/grafana/grafana-server-env"
  local temp_env_file="/tmp/grafana-server.tmp"

  # 使用 envsubst 替换模板中的环境变量
  envsubst < ${template_env_file} > ${temp_env_file}

  proj::util::sudo "cp ${temp_env_file} ${grafana_env_file}"
  rm -f ${temp_env_file}

  # 启动 Grafana 服务
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "systemctl enable grafana"
  proj::util::sudo "systemctl start grafana"

  # 清理下载目录
  # rm -rf ${grafana_download_dir}

  sleep 5
  proj::grafana::status || return 1
  proj::grafana::info
  proj::log::info "install grafana successfully"
}

# Uninstall Grafana step by step
proj::grafana::uninstall() {
  set +o errexit
  proj::util::sudo "systemctl stop grafana"
  proj::util::sudo "systemctl disable grafana"
  proj::util::sudo "rm -f /etc/systemd/system/grafana.service"
  proj::util::sudo "rm -f /etc/default/grafana-server"
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "rm -rf /usr/share/grafana"
  proj::util::sudo "rm -f /usr/local/bin/grafana"
  proj::util::sudo "rm -rf /etc/grafana"
  proj::util::sudo "rm -rf ${PROJ_GRAFANA_DATA_DIR}"
  proj::util::sudo "rm -rf /var/log/grafana"
  proj::util::sudo "userdel grafana" 2>/dev/null || true
  set -o errexit
  proj::log::info "uninstall grafana successfully"
  return 0
}

# Pre-install Grafana by checking system requirements
proj::grafana::pre_install() {
  proj::log::info "Pre-installing Grafana..."

  # 判断是 mac 还是 linux
  if proj::util::is_mac; then
    proj::log::info "Mac OS detected, checking for Grafana installation..."
    if ! command -v grafana >/dev/null 2>&1; then
      proj::log::info "Installing Grafana via brew..."
      brew install grafana
    else
      proj::log::info "Grafana already installed, skipping..."
    fi
  else
    # 检查 envsubst (gettext-base)
    if ! command -v envsubst >/dev/null 2>&1; then
        proj::log::info "Installing gettext-base for envsubst..."
        proj::util::sudo "apt install -y gettext-base"
    fi

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

# Install Grafana using a Docker container
proj::grafana::docker::install() {
  proj::log::info "Installing docker Grafana..."

  proj::grafana::pre_install
  proj::common::network

  # 创建数据和配置目录
  local grafana_data_dir="${PROJ_THIRDPARTY_INSTALL_DIR}/grafana"
  mkdir -p ${grafana_data_dir}
  mkdir -p ${grafana_data_dir}/conf

  # 创建 Grafana Docker 配置文件
  local grafana_docker_conf_file="${grafana_data_dir}/conf/grafana.ini"
  local template_docker_conf_file="${SCRIPT_DIR}/grafana/grafana-docker.ini"
  local temp_docker_conf_file="/tmp/grafana-docker.ini.tmp"

  # 使用 envsubst 替换模板中的环境变量
  envsubst < ${template_docker_conf_file} > ${temp_docker_conf_file}

  # 复制配置文件到数据目录
  proj::util::sudo "cp ${temp_docker_conf_file} ${grafana_docker_conf_file}"
  rm -f ${temp_docker_conf_file}

  # 清理可能存在的同名容器
  proj::common::docker::cleanup_container "${GRAFANA_DOCKER_MNAME}"

  docker run -d --name ${GRAFANA_DOCKER_MNAME} \
    --restart always \
    --network ${NETWORK_NAME} \
    -v ${grafana_data_dir}:/var/lib/grafana \
    -v ${grafana_data_dir}/conf/grafana.ini:/etc/grafana/grafana.ini \
    -p ${PROJ_GRAFANA_HOST}:${PROJ_GRAFANA_PORT}:3000 \
    -e "GF_SECURITY_ADMIN_USER=${PROJ_GRAFANA_ADMIN_USER}" \
    -e "GF_SECURITY_ADMIN_PASSWORD=${PROJ_GRAFANA_ADMIN_PASSWORD}" \
    -e "GF_USERS_ALLOW_SIGN_UP=false" \
    grafana/grafana:${PROJ_GRAFANA_VERSION}

  sleep 5
  if proj::util::is_linux; then
    proj::grafana::status || return 1
  fi
  proj::grafana::info
  proj::log::info "install grafana successfully"
}

# Print necessary information after docker or native installation
proj::grafana::info() {
  echo -e ${C_GREEN}Grafana has been installed, here are some useful information:${C_NORMAL}
  cat << EOF | sed 's/^/  /'
Grafana access endpoint is: http://${PROJ_GRAFANA_HOST}:${PROJ_GRAFANA_PORT}
       Grafana admin user: ${PROJ_GRAFANA_ADMIN_USER}
   Grafana admin password: ${PROJ_GRAFANA_ADMIN_PASSWORD}
       Grafana data dir: ${PROJ_GRAFANA_DATA_DIR}
        Grafana web UI: http://${PROJ_GRAFANA_HOST}:${PROJ_GRAFANA_PORT}
EOF
}

# Uninstall the docker container
proj::grafana::docker::uninstall() {
  docker rm -f ${GRAFANA_DOCKER_MNAME} &>/dev/null
  proj::util::sudo "rm -rf ${PROJ_THIRDPARTY_INSTALL_DIR}/grafana"
  proj::log::info "uninstall grafana successfully"
}

# Status check after docker or native installation
proj::grafana::status() {
  proj::util::telnet ${PROJ_GRAFANA_HOST} ${PROJ_GRAFANA_PORT} || return 1

  # 检查 Grafana 健康状态
  local health_check_url="http://${PROJ_GRAFANA_HOST}:${PROJ_GRAFANA_PORT}/api/health"
  if command -v curl >/dev/null 2>&1; then
    curl -f -s ${health_check_url} >/dev/null || {
      proj::log::error "Grafana health check failed, Grafana maybe not initialized properly."
      return 1
    }
  else
    proj::log::info "curl not found, skipping health check"
  fi
}

# This block allows calling functions in this script directly from the command line.
# It checks if the script's arguments contain a function name with the 'proj::grafana::' prefix
# and, if so, executes that function.
# For example: ./grafana.sh proj::grafana::install
if [[ "$*" =~ proj::grafana:: ]]; then
  eval $*
fi