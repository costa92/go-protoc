#!/usr/bin/env bash
#
# Pyroscope Installation Script
# This script provides functions to install, configure, and manage Pyroscope
# both natively and via Docker containers.
#

# Exit on any error, undefined variables, or pipe failures
set -o errexit
set -o nounset
set -o pipefail

# 加载通用配置和版本管理
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Environment variables for Pyroscope configuration
# Can be overridden by setting these variables before running the script
PROJ_PYROSCOPE_HOST=${PROJ_PYROSCOPE_HOST:-127.0.0.1}             # Pyroscope server host
PROJ_PYROSCOPE_PORT=${PROJ_PYROSCOPE_PORT:-4040}                  # Pyroscope server port
PROJ_PYROSCOPE_DATA_DIR=${PROJ_PYROSCOPE_DATA_DIR:-/var/lib/pyroscope}  # Pyroscope data directory
PROJ_PYROSCOPE_CONFIG_DIR=${PROJ_PYROSCOPE_CONFIG_DIR:-/etc/pyroscope}  # Pyroscope config directory
PROJ_PYROSCOPE_LOG_LEVEL=${PROJ_PYROSCOPE_LOG_LEVEL:-info}        # Log level
# 版本信息从统一配置文件加载：PYROSCOPE_VERSION 在 versions.sh 中定义
PYROSCOPE_DOCKER_MNAME=${NETWORK_NAME}-pyroscope


# Function to install Pyroscope natively
proj::pyroscope::install() {
  proj::pyroscope::pre_install

  # 创建 Pyroscope 相关目录
  proj::util::sudo "mkdir -p ${PROJ_PYROSCOPE_DATA_DIR}"
  proj::util::sudo "mkdir -p ${PROJ_PYROSCOPE_CONFIG_DIR}"
  proj::util::sudo "mkdir -p /var/log/pyroscope"

  # 创建 pyroscope 用户（如果不存在）
  if ! id -u pyroscope >/dev/null 2>&1; then
    proj::util::sudo "useradd --system --shell /bin/false pyroscope"
  fi

  # 设置正确的目录所有权
  proj::util::sudo "chown -R pyroscope:pyroscope ${PROJ_PYROSCOPE_DATA_DIR}"
  proj::util::sudo "chown -R pyroscope:pyroscope ${PROJ_PYROSCOPE_CONFIG_DIR}"
  proj::util::sudo "chown -R pyroscope:pyroscope /var/log/pyroscope"

  # 下载并安装 Pyroscope
  local pyroscope_binary="/usr/local/bin/pyroscope"
  local download_url="https://github.com/grafana/pyroscope/releases/download/v${PYROSCOPE_VERSION}/pyroscope_${PYROSCOPE_VERSION}_linux_amd64.tar.gz"
  local temp_dir="/tmp/pyroscope-install"

  mkdir -p ${temp_dir}
  
  if [[ ! -f "${pyroscope_binary}" ]] || ! ${pyroscope_binary} version | grep -q "${PYROSCOPE_VERSION}"; then
    proj::log::info "Downloading Pyroscope v${PYROSCOPE_VERSION}..."
    curl -L "${download_url}" -o "${temp_dir}/pyroscope.tar.gz"
    tar -xzf "${temp_dir}/pyroscope.tar.gz" -C "${temp_dir}"
    proj::util::sudo "cp ${temp_dir}/pyroscope ${pyroscope_binary}"
    proj::util::sudo "chmod +x ${pyroscope_binary}"
  fi

  # 创建 Pyroscope 配置文件
  local pyroscope_conf_file="${PROJ_PYROSCOPE_CONFIG_DIR}/server.yml"
  local template_conf_file="${SCRIPT_DIR}/pyroscope/server.yml"
  local temp_conf_file="/tmp/pyroscope-server.yml.tmp"

  # 创建默认配置文件模板（如果不存在）
  proj::util::sudo "mkdir -p ${SCRIPT_DIR}/pyroscope"
  if [[ ! -f "${template_conf_file}" ]]; then
    cat > "${temp_conf_file}" << EOF
# Pyroscope Server Configuration
server:
  http-listen-address: ${PROJ_PYROSCOPE_HOST}:${PROJ_PYROSCOPE_PORT}

storage:
  path: ${PROJ_PYROSCOPE_DATA_DIR}

log-level: ${PROJ_PYROSCOPE_LOG_LEVEL}

analytics:
  reporting-disabled: true
EOF
  else
    # 使用 envsubst 替换模板中的环境变量
    envsubst < ${template_conf_file} > ${temp_conf_file}
  fi

  # 复制配置文件到系统目录
  proj::util::sudo "cp ${temp_conf_file} ${pyroscope_conf_file}"
  proj::util::sudo "chown pyroscope:pyroscope ${pyroscope_conf_file}"
  rm -f ${temp_conf_file}

  # 创建 systemd 服务文件
  local pyroscope_service_file="/etc/systemd/system/pyroscope.service"
  local template_service_file="${SCRIPT_DIR}/pyroscope/pyroscope.service"
  local temp_service_file="/tmp/pyroscope.service.tmp"

  if [[ ! -f "${template_service_file}" ]]; then
    cat > "${temp_service_file}" << EOF
[Unit]
Description=Pyroscope Server
After=network.target
Wants=network.target

[Service]
User=pyroscope
Group=pyroscope
Type=simple
ExecStart=${pyroscope_binary} server -config-file ${pyroscope_conf_file}
WorkingDirectory=${PROJ_PYROSCOPE_DATA_DIR}
Restart=on-failure
RestartSec=5

# Output to journal
StandardOutput=journal
StandardError=journal
SyslogIdentifier=pyroscope

# Security settings
NoNewPrivileges=true
ProtectHome=true
ProtectSystem=strict
ReadWritePaths=${PROJ_PYROSCOPE_DATA_DIR}

[Install]
WantedBy=multi-user.target
EOF
  else
    # 使用 envsubst 替换模板中的环境变量
    envsubst < ${template_service_file} > ${temp_service_file}
  fi

  # 复制服务文件到系统目录
  proj::util::sudo "cp ${temp_service_file} ${pyroscope_service_file}"
  rm -f ${temp_service_file}

  # 启动 Pyroscope 服务
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "systemctl enable pyroscope"
  proj::util::sudo "systemctl start pyroscope"

  # 清理临时文件
  rm -rf ${temp_dir}

  sleep 2
  proj::pyroscope::status || return 1
  proj::pyroscope::info
  proj::log::info "install pyroscope successfully"
}


# Uninstall the pyroscope step by step.
proj::pyroscope::uninstall()
{
  set +o errexit
  proj::util::sudo "systemctl stop pyroscope"
  proj::util::sudo "systemctl disable pyroscope"
  proj::util::sudo "rm -f /etc/systemd/system/pyroscope.service"
  proj::util::sudo "rm -f /usr/local/bin/pyroscope"
  proj::util::sudo "rm -rf ${PROJ_PYROSCOPE_DATA_DIR}"
  proj::util::sudo "rm -rf ${PROJ_PYROSCOPE_CONFIG_DIR}"
  proj::util::sudo "rm -rf /var/log/pyroscope"
  proj::util::sudo "userdel pyroscope" 2>/dev/null || true
  proj::util::sudo "systemctl daemon-reload"
  set -o errexit
  proj::log::info "uninstall pyroscope successfully"
  return 0
}


# Pre-install Pyroscope by system init and package management.
proj::pyroscope::pre_install(){
    proj::log::info "Pre-installing Pyroscope..."
    # 判断是 mac 还是 linux
    if proj::util::is_mac; then
        proj::log::info "Mac OS detected, installing required tools..."
        if ! command -v curl >/dev/null 2>&1; then
            proj::log::info "Installing curl..."
            brew install curl
        fi
    else
        # 检查必要工具
        if ! command -v curl >/dev/null 2>&1; then
            proj::log::info "Installing curl..."
            proj::util::sudo "apt install -y curl"
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

# Install Pyroscope using a Docker container.
proj::pyroscope::docker::install(){
    proj::log::info "Installing docker Pyroscope..."

    proj::pyroscope::pre_install
    proj::common::network

    # 清理可能存在的同名容器
    proj::common::docker::cleanup_container "${PYROSCOPE_DOCKER_MNAME}"

    # 创建 Pyroscope 数据目录并设置权限
    proj::util::sudo "mkdir -p ${PROJ_THIRDPARTY_INSTALL_DIR}/pyroscope"
    proj::util::sudo "chown -R ${USER}:${USER} ${PROJ_THIRDPARTY_INSTALL_DIR}/pyroscope"

    # 创建 Pyroscope Docker 配置文件
    local pyroscope_docker_conf_file="${PROJ_THIRDPARTY_INSTALL_DIR}/pyroscope/server.yml"
    local temp_docker_conf_file="/tmp/pyroscope-docker-server.yml.tmp"

    # 创建 Docker 配置文件
    cat > ${temp_docker_conf_file} << EOF
# Pyroscope Server Configuration for Docker
server:
  http-listen-address: 0.0.0.0:4040

storage:
  path: /data

log-level: ${PROJ_PYROSCOPE_LOG_LEVEL}

analytics:
  reporting-disabled: true
EOF

    # 复制配置文件到数据目录
    cp ${temp_docker_conf_file} ${pyroscope_docker_conf_file}
    rm -f ${temp_docker_conf_file}

    # 启动 Pyroscope 容器
    docker run -d --name ${PYROSCOPE_DOCKER_MNAME} \
      --restart always \
      --network ${NETWORK_NAME} \
      -v ${PROJ_THIRDPARTY_INSTALL_DIR}/pyroscope:/data \
      -v ${pyroscope_docker_conf_file}:/etc/pyroscope/server.yml:ro \
      -p ${PROJ_PYROSCOPE_HOST}:${PROJ_PYROSCOPE_PORT}:4040 \
      grafana/pyroscope:${PYROSCOPE_VERSION} \
      server -config-file /etc/pyroscope/server.yml

    sleep 3
    if proj::util::is_linux; then
        proj::pyroscope::status || return 1
    fi
    proj::pyroscope::info
    proj::log::info "install pyroscope successfully"
}


# Print necessary information after docker or sbs installation.
proj::pyroscope::info()
{
  echo -e ${C_GREEN}pyroscope has been installed, here are some useful information:${C_NORMAL}
  cat << EOF | sed 's/^/  /'
Pyroscope Web UI is: http://${PROJ_PYROSCOPE_HOST}:${PROJ_PYROSCOPE_PORT}
      Pyroscope API: http://${PROJ_PYROSCOPE_HOST}:${PROJ_PYROSCOPE_PORT}/api/v1/
   Pyroscope Data Dir: ${PROJ_PYROSCOPE_DATA_DIR}
EOF
}

# Uninstall the docker container.
proj::pyroscope::docker::uninstall()
{
  docker rm -f ${PYROSCOPE_DOCKER_MNAME} &>/dev/null
  proj::util::sudo "rm -rf ${PROJ_THIRDPARTY_INSTALL_DIR}/pyroscope"
  proj::log::info "uninstall pyroscope successfully"
}

# Status check after docker or sbs installation.
proj::pyroscope::status()
{
  proj::util::telnet ${PROJ_PYROSCOPE_HOST} ${PROJ_PYROSCOPE_PORT} || return 1
  
  # 检查 Pyroscope API 是否可访问
  if command -v curl >/dev/null 2>&1; then
    if curl -s --max-time 5 "http://${PROJ_PYROSCOPE_HOST}:${PROJ_PYROSCOPE_PORT}/ready" >/dev/null; then
      proj::log::info "Pyroscope is healthy and ready"
      return 0
    else
      proj::log::error "Pyroscope API is not responding"
      return 1
    fi
  fi
  
  return 0
}

# This block allows calling functions in this script directly from the command line.
# It checks if the script's arguments contain a function name with the 'proj::pyroscope::' prefix
# and, if so, executes that function.
# For example: ./pyroscope.sh proj::pyroscope::install
if [[ "$*" =~ proj::pyroscope:: ]]; then
  eval $*
fi