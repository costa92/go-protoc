#!/usr/bin/env bash

# Exit on any error, undefined variables, or pipe failures
set -o errexit
set -o nounset
set -o pipefail

# 加载通用配置和版本管理
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Environment variables for MariaDB configuration
# Can be overridden by setting these variables before running the script
PROJ_MYSQL_HOST=${PROJ_MYSQL_HOST:-127.0.0.1}                          # MariaDB server host
PROJ_MYSQL_PORT=${PROJ_MYSQL_PORT:-3306}                               # MariaDB server port
PROJ_MYSQL_ADMIN_USERNAME=${PROJ_MYSQL_ADMIN_USERNAME:-root}           # Admin username
PROJ_MYSQL_ADMIN_PASSWORD=${PROJ_MYSQL_ADMIN_PASSWORD:-proj(#)666}     # Admin password
PROJ_PASSWORD=${PROJ_PASSWORD:-proj(#)666}                             # Legacy password variable
PROJ_MYSQL_DATA_DIR=${PROJ_MYSQL_DATA_DIR:-/var/lib/mysql}             # MariaDB data directory
PROJ_MYSQL_CONFIG_DIR=${PROJ_MYSQL_CONFIG_DIR:-/etc/mysql}             # MariaDB config directory
# 版本信息从统一配置文件加载：MARIADB_VERSION 在 versions.sh 中定义
MARIADB_DOCKER_MNAME=${NETWORK_NAME}-mariadb

proj::mariadb::pre_install()
{
  proj::log::info "Pre-installing MariaDB..."
  if proj::util::is_linux; then
      # 检查是否已安装 MariaDB 客户端，如果没有则安装
      if ! proj::util::cmd_exists "mariadb-client"; then
          proj::log::info "Installing mariadb-client for accessing MariaDB"
          proj::util::sudo "DEBIAN_FRONTEND=noninteractive apt install -y mariadb-client"
      else
          proj::log::info "mariadb-client already exists, skipping installation"
      fi

      # 检查 envsubst (gettext-base)
      if ! command -v envsubst >/dev/null 2>&1; then
          proj::log::info "Installing gettext-base for envsubst..."
          proj::util::sudo "apt install -y gettext-base"
      fi
  fi
}

# Install mariadb using containerization.
proj::mariadb::docker::install()
{
  proj::mariadb::pre_install
  proj::common::network
  
  # 清理可能存在的同名容器
  proj::common::docker::cleanup_container "${MARIADB_DOCKER_MNAME}"
  
  # 创建 MariaDB 数据和配置目录
  proj::util::sudo "mkdir -p ${PROJ_THIRDPARTY_INSTALL_DIR}/mariadb"
  proj::util::sudo "mkdir -p ${PROJ_THIRDPARTY_INSTALL_DIR}/mariadb/conf.d"
  
  # 在 macOS 上设置正确的权限，让 MariaDB 容器能够访问
  if [[ "$(uname -s)" == "Darwin" ]]; then
    proj::util::sudo "chmod 755 ${PROJ_THIRDPARTY_INSTALL_DIR}/mariadb"
    proj::util::sudo "chown -R $(id -u):$(id -g) ${PROJ_THIRDPARTY_INSTALL_DIR}/mariadb"
  fi

  # 创建 MariaDB Docker 配置文件
  local mariadb_docker_conf_file="${PROJ_THIRDPARTY_INSTALL_DIR}/mariadb/conf.d/mariadb-docker.cnf"
  local template_docker_conf_file="${SCRIPT_DIR}/mariadb/mariadb-docker.cnf"
  local temp_docker_conf_file="/tmp/mariadb-docker.cnf.tmp"

  # 使用 envsubst 替换模板中的环境变量
  envsubst < ${template_docker_conf_file} > ${temp_docker_conf_file}

  # 复制配置文件到数据目录
  proj::util::sudo "cp ${temp_docker_conf_file} ${mariadb_docker_conf_file}"
  rm -f ${temp_docker_conf_file}

  # 启动 MariaDB 容器，使用配置文件
  # 在 macOS 上使用 Docker named volume 避免权限问题
  if [[ "$(uname -s)" == "Darwin" ]]; then
    docker run -d --name ${MARIADB_DOCKER_MNAME} \
      --restart always \
      --network ${NETWORK_NAME} \
      -v ${MARIADB_DOCKER_MNAME}-data:/var/lib/mysql \
      -v ${PROJ_THIRDPARTY_INSTALL_DIR}/mariadb/conf.d:/etc/mysql/conf.d:ro \
      -p 0.0.0.0:${PROJ_MYSQL_PORT}:3306 \
      -e MYSQL_ROOT_PASSWORD=${PROJ_MYSQL_ADMIN_PASSWORD} \
      mariadb:${MARIADB_VERSION}
  else
    # Linux 使用 bind mount
    docker run -d --name ${MARIADB_DOCKER_MNAME} \
      --restart always \
      --network ${NETWORK_NAME} \
      -v ${PROJ_THIRDPARTY_INSTALL_DIR}/mariadb:/var/lib/mysql \
      -v ${PROJ_THIRDPARTY_INSTALL_DIR}/mariadb/conf.d:/etc/mysql/conf.d \
      -p 0.0.0.0:${PROJ_MYSQL_PORT}:3306 \
      -e MYSQL_ROOT_PASSWORD=${PROJ_MYSQL_ADMIN_PASSWORD} \
      mariadb:${MARIADB_VERSION}
  fi

  echo "Sleeping to wait for all mariadb container to complete startup ..."
  sleep 10
  proj::mariadb::status || return 1

  proj::mariadb::info
  proj::log::info "install mariadb successfully"
}

# Uninstall the docker container.
proj::mariadb::docker::uninstall()
{
  docker rm -f ${MARIADB_DOCKER_MNAME} &>/dev/null
  
  # 在 macOS 上清理 Docker volume
  if [[ "$(uname -s)" == "Darwin" ]]; then
    docker volume rm -f ${MARIADB_DOCKER_MNAME}-data &>/dev/null
  fi
  
  proj::util::sudo "rm -rf ${PROJ_THIRDPARTY_INSTALL_DIR}/mariadb" 2>/dev/null || true
  proj::log::info "uninstall mariadb successfully"
}

# Install the mariadb step by step.
# sbs is the abbreviation for "step by step".
proj::mariadb::install()
{
  # Ensure a clean state before installation
  proj::mariadb::uninstall

  # 本机 apt 安装后 MySQL 端口固定位 3306
  export PROJ_MYSQL_PORT=3306

  # Download MariaDB GPG key and store it in the keyring
  proj::util::sudo "mkdir -p /usr/share/keyrings"
  echo ${LINUX_PASSWORD} | sudo -S bash -c "curl -sL 'https://mariadb.org/mariadb_release_signing_key.asc' | gpg --dearmor > /usr/share/keyrings/mariadb-archive-keyring.gpg"


  # 配置 MariaDB 11.2.2 apt 源（docker install 和 sbs install 版本都要保持一致）
  # 根据系统类型设置不同的源
  if proj::util::is_ubuntu; then
    # 获取 Ubuntu 版本代号
    ubuntu_codename=$(lsb_release -cs)
    # 对于 Ubuntu 24.04 (noble) 及更新版本，使用 mantic (23.10) 的包，因为 noble 不在归档中
    if [[ "$ubuntu_codename" == "noble" ]] || [[ "$ubuntu_codename" > "mantic" ]]; then
      proj::log::info "Using MariaDB 11.2.2 archive repository with mantic packages for $ubuntu_codename"
      echo ${LINUX_PASSWORD} | sudo -S echo "deb [signed-by=/usr/share/keyrings/mariadb-archive-keyring.gpg arch=amd64,arm64] https://archive.mariadb.org/mariadb-11.2.2/repo/ubuntu/ mantic main" | sudo tee /etc/apt/sources.list.d/mariadb-11.2.2.list
    elif [[ "$ubuntu_codename" == "mantic" ]] || [[ "$ubuntu_codename" == "lunar" ]] || [[ "$ubuntu_codename" == "jammy" ]] || [[ "$ubuntu_codename" == "focal" ]] || [[ "$ubuntu_codename" == "bionic" ]]; then
      # 支持的 Ubuntu 版本使用对应的归档源
      proj::log::info "Using MariaDB 11.2.2 archive repository for $ubuntu_codename"
      echo ${LINUX_PASSWORD} | sudo -S echo "deb [signed-by=/usr/share/keyrings/mariadb-archive-keyring.gpg arch=amd64,arm64] https://archive.mariadb.org/mariadb-11.2.2/repo/ubuntu/ $ubuntu_codename main" | sudo tee /etc/apt/sources.list.d/mariadb-11.2.2.list
    else
      # 其他版本使用 jammy 作为回退
      proj::log::info "Using MariaDB 11.2.2 archive repository with jammy packages for $ubuntu_codename"
      echo ${LINUX_PASSWORD} | sudo -S echo "deb [signed-by=/usr/share/keyrings/mariadb-archive-keyring.gpg arch=amd64,arm64] https://archive.mariadb.org/mariadb-11.2.2/repo/ubuntu/ jammy main" | sudo tee /etc/apt/sources.list.d/mariadb-11.2.2.list
    fi
  elif proj::util::is_debian; then
    echo ${LINUX_PASSWORD} | sudo -S echo "deb [signed-by=/usr/share/keyrings/mariadb-archive-keyring.gpg arch=amd64,arm64] https://archive.mariadb.org/mariadb-11.2.2/repo/debian/ $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/mariadb-11.2.2.list
  else
    proj::log::error "Unsupported operating system. Only Ubuntu and Debian are supported."
    return 1
  fi

  # 注意：一定要执行 `apt update`，否则可能安装的还是旧的软件包
  proj::util::sudo "apt update"

  # 检查是否已安装 MariaDB 客户端，如果已存在则跳过安装
  proj::mariadb::pre_install

  # 需要先创建 /var/lib/mysql/ 目录，否则 `systemctl start mariadb` 时可能会报错
  proj::util::sudo "mkdir -p /var/lib/mysql"

    # 执行以下命令，防止uninstall后，出现：`update-alternatives: error: alternative path /etc/mysql/mariadb.cnf doesn't exist` 错误
  # 安装 MariaDB 客户端和 MariaDB 服务端
  proj::util::sudo "apt install -y -o Dpkg::Options::="--force-confmiss" --reinstall mariadb-client mariadb-server"

  # 启动 MariaDB，并设置开机启动
  proj::util::sudo "systemctl enable mariadb"

  # 创建 MariaDB 服务器配置文件
  local mariadb_server_conf_file="/etc/mysql/mariadb.conf.d/50-server.cnf"
  local template_server_conf_file="${SCRIPT_DIR}/mariadb/50-server.cnf"
  local temp_server_conf_file="/tmp/50-server.cnf.tmp"

  # 使用 envsubst 替换模板中的环境变量
  envsubst < ${template_server_conf_file} > ${temp_server_conf_file}

  # 复制配置文件到系统目录
  proj::util::sudo "cp ${temp_server_conf_file} ${mariadb_server_conf_file}"
  proj::util::sudo "chown root:root ${mariadb_server_conf_file}"
  rm -f ${temp_server_conf_file}

  # 创建客户端配置文件
  local mariadb_client_conf_file="/etc/mysql/conf.d/mariadb-client.cnf"
  local template_client_conf_file="${SCRIPT_DIR}/mariadb/mariadb-client.cnf"
  local temp_client_conf_file="/tmp/mariadb-client.cnf.tmp"

  # 使用 envsubst 替换模板中的环境变量
  envsubst < ${template_client_conf_file} > ${temp_client_conf_file}

  # 复制配置文件到系统目录
  proj::util::sudo "cp ${temp_client_conf_file} ${mariadb_client_conf_file}"
  proj::util::sudo "chown root:root ${mariadb_client_conf_file}"
  rm -f ${temp_client_conf_file}

  proj::util::sudo "systemctl restart mariadb"

  #  设置 root 初始密码
  proj::util::sudo "mysqladmin -u${PROJ_MYSQL_ADMIN_USERNAME} password ${PROJ_MYSQL_ADMIN_PASSWORD}"

  proj::mariadb::status || return 1
  proj::mariadb::info
  proj::log::info "install mariadb successfully"
}

# Uninstall the mariadb step by step.
proj::mariadb::uninstall()
{
  # `|| true` 实现幂等
  proj::util::sudo "systemctl stop mariadb" || true
  proj::util::sudo "systemctl disable mariadb" || true
  proj::util::sudo "apt remove -y mariadb-client mariadb-server" || true

  # 删除配置文件和数据目录，以及其他关联安装文件
  proj::util::sudo "rm -rvf /var/lib/mysql"
  proj::util::sudo "rm -rvf /etc/mysql"
  proj::util::sudo "rm -f /usr/share/keyrings/mariadb-archive-keyring.gpg"
  proj::util::sudo "rm -vf /etc/apt/sources.list.d/mariadb-11.2.2.list"
  proj::log::info "uninstall mariadb successfully"
}

# Print necessary information after docker or sbs installation.
proj::mariadb::info()
{
  echo -e ${C_GREEN}MariaDB has been installed, here are some useful information:${C_NORMAL}
  cat << EOF | sed 's/^/  /'
MariaDB access endpoint is: ${PROJ_MYSQL_HOST}:${PROJ_MYSQL_PORT}
       Admin username is: ${PROJ_MYSQL_ADMIN_USERNAME}
       Admin password is: ${PROJ_MYSQL_ADMIN_PASSWORD}
# \`mysql\` will be deprecated in the future, so here use \`mariadb\` instead.
Access command: mariadb -h ${PROJ_MYSQL_HOST} -P ${PROJ_MYSQL_PORT} -u ${PROJ_MYSQL_ADMIN_USERNAME} -p'${PROJ_MYSQL_ADMIN_PASSWORD}'
EOF
}

# Status check after docker or sbs installation.
proj::mariadb::status()
{
  sleep 20
  # 基础检查：检查端口，基础检查
  proj::util::telnet ${PROJ_MYSQL_HOST} ${PROJ_MYSQL_PORT} || return 1

  # 终态检查：检查 MariaDB 是否成功运行
  # 优先使用 docker exec 进行连接测试
  if docker ps --format "table {{.Names}}" | grep -q "${MARIADB_DOCKER_MNAME}"; then
    echo "Testing MariaDB connection via docker exec..."
    docker exec ${MARIADB_DOCKER_MNAME} mariadb -u${PROJ_MYSQL_ADMIN_USERNAME} -p${PROJ_MYSQL_ADMIN_PASSWORD} -e "quit" &>/dev/null || {
      proj::log::error "can not login with root via docker exec, mariadb maybe not initialized properly."
      return 1
    }
  else
    # 回退到本地客户端（如果可用）
    echo mariadb -h${PROJ_MYSQL_HOST} -P${PROJ_MYSQL_PORT} -u${PROJ_MYSQL_ADMIN_USERNAME} -p${PROJ_MYSQL_ADMIN_PASSWORD} -e quit
    if command -v mariadb >/dev/null 2>&1; then
      mariadb -h${PROJ_MYSQL_HOST} -P${PROJ_MYSQL_PORT} -u${PROJ_MYSQL_ADMIN_USERNAME} -p${PROJ_MYSQL_ADMIN_PASSWORD} -e quit &>/dev/null || {
        proj::log::error "can not login with root, mariadb maybe not initialized properly."
        return 1
      }
    else
      proj::log::warn "MariaDB client not installed locally, skipping connection test"
    fi
  fi
}

if [[ "$*" =~ proj::mariadb:: ]]; then
  eval $*
fi