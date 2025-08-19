#!/usr/bin/env bash
#
# Redis Installation Script
# This script provides functions to install, configure, and manage Redis server
# both natively and via Docker containers.
#

# Exit on any error, undefined variables, or pipe failures
set -o errexit
set -o nounset
set -o pipefail

# 加载通用配置和版本管理
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Environment variables for Redis configuration
# Can be overridden by setting these variables before running the script
PROJ_REDIS_HOST=${PROJ_REDIS_HOST:-127.0.0.1}        # Redis server host
PROJ_REDIS_PORT=${PROJ_REDIS_PORT:-6379}             # Redis server port
PROJ_REDIS_PASSWORD=${PROJ_REDIS_PASSWORD:-proj(#)666}  # Redis authentication password
PROJ_REDIS_DATA_DIR=${PROJ_REDIS_DATA_DIR:-/var/lib/redis}  # Redis data directory
PROJ_REDIS_CONFIG_DIR=${PROJ_REDIS_CONFIG_DIR:-/etc/redis}  # Redis config directory
# 版本信息从统一配置文件加载：REDIS_VERSION 在 versions.sh 中定义
REDIS_DOCKER_MNAME=${NETWORK_NAME}-redis


# Function to install Redis using kubectl
proj::redis::install() {
  proj::redis::pre_install

  # 创建 Redis 相关目录
  proj::util::sudo "mkdir -p ${PROJ_REDIS_DATA_DIR}"
  proj::util::sudo "mkdir -p ${PROJ_REDIS_CONFIG_DIR}"
  proj::util::sudo "mkdir -p /var/log/redis"
  proj::util::sudo "mkdir -p /run/redis"

  # 创建 redis 用户（如果不存在）
  if ! id -u redis >/dev/null 2>&1; then
    proj::util::sudo "useradd --system --shell /bin/false redis"
  fi

  # 设置正确的目录所有权
  proj::util::sudo "chown -R redis:redis ${PROJ_REDIS_DATA_DIR}"
  proj::util::sudo "chown -R redis:redis ${PROJ_REDIS_CONFIG_DIR}"
  proj::util::sudo "chown -R redis:redis /var/log/redis"
  proj::util::sudo "chown -R redis:redis /run/redis"

  # 安装 Redis
  proj::util::sudo "apt install -y -o Dpkg::Options::="--force-confmiss" --reinstall redis-server"

  # 创建 Redis 配置文件
  local redis_conf_file="${PROJ_REDIS_CONFIG_DIR}/redis.conf"
  local template_conf_file="${SCRIPT_DIR}/redis/redis.conf"
  local temp_conf_file="/tmp/redis.conf.tmp"

  # 使用 envsubst 替换模板中的环境变量
  envsubst < ${template_conf_file} > ${temp_conf_file}

  # 复制配置文件到系统目录
  proj::util::sudo "cp ${temp_conf_file} ${redis_conf_file}"
  proj::util::sudo "chown redis:redis ${redis_conf_file}"
  rm -f ${temp_conf_file}

  # 创建 systemd 服务文件
  local redis_service_file="/etc/systemd/system/redis.service"
  local template_service_file="${SCRIPT_DIR}/redis/redis.service"
  local temp_service_file="/tmp/redis.service.tmp"

  # 使用 envsubst 替换模板中的环境变量
  envsubst < ${template_service_file} > ${temp_service_file}

  # 复制服务文件到系统目录
  proj::util::sudo "cp ${temp_service_file} ${redis_service_file}"
  rm -f ${temp_service_file}

  # 启动 Redis 服务
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "systemctl enable redis"

  # 为了能够远程连上 Redis，需要执行以下命令关闭防火墙，并禁止防火墙开机启动（如果不需要远程连接，可忽略此步骤）
  set +o errexit
  #proj::util::sudo "systemctl stop firewalld.service"
  #proj::util::sudo "systemctl disable firewalld.service"
  set -o errexit

  # 重启 Redis
  #proj::util::sudo "redis-server ${redis_conf}"
  proj::util::sudo "systemctl restart redis-server"

  proj::redis::status || return 1
  proj::redis::info
  proj::log::info "install redis successfully"
}


# Uninstall the redis step by step.
proj::redis::uninstall()
{
  # 先删除 redis-server 进程，否则 `systemctl stop redis-server` 可能会卡主
  set +o errexit
  redis_pid=$(pgrep -f redis-server)
  [[ ${redis_pid} != "" ]] && sudo kill -9 ${redis_pid}
  proj::util::sudo "systemctl stop redis-server"
  proj::util::sudo "systemctl disable redis-server"
  proj::util::sudo "apt remove -y redis-server"
  proj::util::sudo "rm -rf /var/lib/redis"
  set -o errexit
  proj::log::info "uninstall redis successfully"
  return 0
}


# Pre-install Redis by system init and package management.
proj::redis::pre_install(){
    proj::log::info "Pre-installing Redis..."
    # 判断是 mac 还是 linux
    if proj::util::is_mac; then
        proj::log::info "Mac OS detected, skipping Redis installation..."
        # proj::util::exec "brew install redis-cli"
    else
        if ! proj::util::cmd_exists "redis-tools"; then
            proj::log::info "Installing redis-tools..."
            proj::util::sudo "apt install -y redis-tools"
        else
            proj::log::info "redis-tools already installed, skipping..."
        fi

        # 检查 envsubst (gettext-base)
        if ! command -v envsubst >/dev/null 2>&1; then
            proj::log::info "Installing gettext-base for envsubst..."
            proj::util::sudo "apt install -y gettext-base"
        fi
    fi
}

# Install Redis using a Docker container.
proj::redis::docker::install(){
    proj::log::info "Installing docker Redis..."

    proj::redis::pre_install
    proj::common::network

    # 清理可能存在的同名容器
    proj::common::docker::cleanup_container "${REDIS_DOCKER_MNAME}"

    # 创建 Redis 数据目录
    proj::util::sudo "mkdir -p ${PROJ_THIRDPARTY_INSTALL_DIR}/redis"

    # 创建 Redis Docker 配置文件
    local redis_docker_conf_file="${PROJ_THIRDPARTY_INSTALL_DIR}/redis/redis.conf"
    local template_docker_conf_file="${SCRIPT_DIR}/redis/redis-docker.conf"
    local temp_docker_conf_file="/tmp/redis-docker.conf.tmp"

    # 使用 envsubst 替换模板中的环境变量
    envsubst < ${template_docker_conf_file} > ${temp_docker_conf_file}

    # 复制配置文件到数据目录
    proj::util::sudo "cp ${temp_docker_conf_file} ${redis_docker_conf_file}"
    rm -f ${temp_docker_conf_file}

    # 启动 Redis 容器，使用配置文件
    docker run -d --name ${REDIS_DOCKER_MNAME} \
      --restart always \
      --network ${NETWORK_NAME} \
      -v ${PROJ_THIRDPARTY_INSTALL_DIR}/redis:/data \
      -p ${PROJ_REDIS_HOST}:${PROJ_REDIS_PORT}:6379 \
      redis:${REDIS_VERSION} \
      redis-server /data/redis.conf

    sleep 2
    if proj::util::is_linux; then
        proj::redis::status || return 1
    fi
    proj::redis::info
    proj::log::info "install redis successfully"
}


# Print necessary information after docker or sbs installation.
proj::redis::info()
{
  echo -e ${C_GREEN}redis has been installed, here are some useful information:${C_NORMAL}
  cat << EOF | sed 's/^/  /'
Redis access endpoint is: ${PROJ_REDIS_HOST}:${PROJ_REDIS_PORT}
       Redis password is: ${PROJ_REDIS_PASSWORD}
     Redis Login Command: redis-cli --no-auth-warning -h ${PROJ_REDIS_HOST} -p ${PROJ_REDIS_PORT} -a '${PROJ_REDIS_PASSWORD}'
EOF
}

# Uninstall the docker container.
proj::redis::docker::uninstall()
{
  docker rm -f ${REDIS_DOCKER_MNAME} &>/dev/null
  proj::util::sudo "rm -rf ${PROJ_THIRDPARTY_INSTALL_DIR}/redis"
  proj::log::info "uninstall redis successfully"
}

# Status check after docker or sbs installation.
proj::redis::status()
{
  proj::util::telnet ${PROJ_REDIS_HOST} ${PROJ_REDIS_PORT} || return 1
  redis-cli --no-auth-warning -h ${PROJ_REDIS_HOST} -p ${PROJ_REDIS_PORT} -a "${PROJ_REDIS_PASSWORD}" --hotkeys || {
    proj::log::error "can not login with ${PROJ_REDIS_USERNAME}, redis maybe not initialized properly."
    return 1
  }
}

# This block allows calling functions in this script directly from the command line.
# It checks if the script's arguments contain a function name with the 'proj::redis::' prefix
# and, if so, executes that function.
# For example: ./redis.sh proj::redis::install
if [[ "$*" =~ proj::redis:: ]]; then
  eval $*
fi




