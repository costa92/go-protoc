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
# 版本信息从统一配置文件加载：REDIS_VERSION 在 versions.sh 中定义
REDIS_DOCKER_MNAME=${NETWORK_NAME}-redis


# Function to install Redis using kubectl
proj::redis::install() {
  proj::redis::pre_install

  # 创建 `/var/lib/redis` 目录，否则 `redis-server` 命令启动时
  # 会报：`Can't chdir to '/var/lib/redis': No such file or directory` 错误
  proj::util::sudo "mkdir -p /var/lib/redis"

  # 安装 Redis
  proj::util::sudo "apt install -y -o Dpkg::Options::="--force-confmiss" --reinstall redis-server"

  # 配置 Redis
  # 修改 `/etc/redis/redis.conf` 文件，将 daemonize 由 no 改成 yes，表示允许 Redis 在后台启动
  redis_conf=/etc/redis/redis.conf
  # 注意：有的系统 redis 配置文件路径为 `/etc/redis.conf`
  [[ -f /etc/redis.conf ]] && redis_conf=/etc/redis.conf

  echo ${LINUX_PASSWORD} | sudo -S sed -i '/^daemonize/{s/no/yes/}' ${redis_conf}

  # 修改 Redis 端口为 ${PROJ_REDIS_PORT}
  echo ${LINUX_PASSWORD} | sudo -S sed -i "s/^port.*/port ${PROJ_REDIS_PORT}/g" ${redis_conf}

  # 在 `bind 127.0.0.1` 前面添加 `#` 将其注释掉，默认情况下只允许本地连接，注释掉后外网可以连接 Redis
  echo ${LINUX_PASSWORD} | sudo -S sed -i '/^bind .*127.0.0.1/s/^/# /' ${redis_conf}

  # 修改 requirepass 配置，设置 Redis 密码
  echo ${LINUX_PASSWORD} | sudo -S sed -i 's/^# requirepass.*$/requirepass '"${PROJ_REDIS_PASSWORD}"'/' ${redis_conf}

  # 因为我们上面配置了密码登录，需要将 protected-mode 设置为 no，关闭保护模式
  echo ${LINUX_PASSWORD} | sudo -S sed -i '/^protected-mode/{s/yes/no/}' ${redis_conf}

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
    fi
}

# Func
# Install Redis using a Docker container.
proj::redis::docker::install(){
    proj::log::info "Installing docker Redis..."

    proj::redis::pre_install
    proj::common::network

    docker run -d --name ${REDIS_DOCKER_MNAME} \
      --restart always \
      --network ${NETWORK_NAME} \
      -v ${PROJ_THIRDPARTY_INSTALL_DIR}/redis:/data \
      -p ${PROJ_REDIS_HOST}:${PROJ_REDIS_PORT}:6379 \
      redis:7.2.3 \
      redis-server \
      --appendonly yes \
      --save 60 1 \
      --protected-mode no \
      --requirepass ${PROJ_REDIS_PASSWORD} \
      --loglevel debug

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




