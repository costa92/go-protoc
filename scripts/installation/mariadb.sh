#!/usr/bin/env bash


# The root of the build/dist directory.
PROJ_ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/../..
# If common.sh has already been sourced, it will not be sourced again here.
[[ -z ${COMMON_SOURCED} ]] && source ${PROJ_ROOT_DIR}/scripts/installation/common.sh
# Set some environment variables.
PROJ_MYSQL_HOST=${PROJ_MYSQL_HOST:-127.0.0.1}
PROJ_MYSQL_PORT=${PROJ_MYSQL_PORT:-3306}
PROJ_PASSWORD=${PROJ_PASSWORD:-onex(#)666}
MARIADB_DOCKER_MNAME=${NETWORK_NAME}-mariadb

proj::mariadb::pre_install()
{
  if proj::util::is_linux; then
      # 检查是否已安装 MariaDB 客户端，如果没有则安装
      if ! proj::util::cmd_exists "mariadb-client"; then
          proj::log::info "Installing mariadb-client for accessing MariaDB"
          proj::util::sudo "DEBIAN_FRONTEND=noninteractive apt install -y mariadb-client"
      else
          proj::log::info "mariadb-client already exists, skipping installation"
      fi
  fi
}

# Install mariadb using containerization.
proj::mariadb::docker::install()
{
  proj::mariadb::pre_install
  proj::common::network
  docker run -d --name ${MARIADB_DOCKER_MNAME} \
    --restart always \
    --network ${NETWORK_NAME} \
    -v ${PROJ_THIRDPARTY_INSTALL_DIR}/mariadb:/var/lib/mysql \
    -p 0.0.0.0:${PROJ_MYSQL_PORT}:3306 \
    -e MYSQL_ROOT_PASSWORD=${PROJ_PASSWORD} \
    mariadb:11.2.2

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
  proj::util::sudo "rm -rf ${PROJ_THIRDPARTY_INSTALL_DIR}/mariadb"
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

  # 为了方便你访问 MySQL，这里我们设置 MySQL 允许从所有机器网卡访问
  echo ${LINUX_PASSWORD} | sudo -S sed -i 's/^bind-address.*/bind-address = 0.0.0.0/g' /etc/mysql/mariadb.conf.d/50-server.cnf

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
  proj::color::green "mariadb has been installed, here are some useful information:"
  cat << EOF | sed 's/^/  /'
MySQL access endpoint is: ${PROJ_MYSQL_HOST}:${PROJ_MYSQL_PORT}
        root password is: ${PROJ_PASSWORD}
# `mysql` will be deprecated in the future, so here use `mariadb` instead.
Access command: mariadb -h ${PROJ_MYSQL_HOST} -P ${PROJ_MYSQL_PORT} -u root -p'${PROJ_PASSWORD}'
EOF
}

# Status check after docker or sbs installation.
proj::mariadb::status()
{
  sleep 20
  # 基础检查：检查端口，基础检查
  proj::util::telnet ${PROJ_MYSQL_HOST} ${PROJ_MYSQL_PORT} || return 1

  # 终态检查：检查 MySQL 是否成功运行
  echo mariadb -h${PROJ_MYSQL_HOST} -P${PROJ_MYSQL_PORT} -u${PROJ_MYSQL_ADMIN_USERNAME} -p${PROJ_MYSQL_ADMIN_PASSWORD} -e quit
  mariadb -h${PROJ_MYSQL_HOST} -P${PROJ_MYSQL_PORT} -u${PROJ_MYSQL_ADMIN_USERNAME} -p${PROJ_MYSQL_ADMIN_PASSWORD} -e quit &>/dev/null || {
    proj::log::error "can not login with root, mariadb maybe not initialized properly."
    return 1
  }
}

if [[ "$*" =~ proj::mariadb:: ]]; then
  eval $*
fi