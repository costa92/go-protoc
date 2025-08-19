#!/usr/bin/env bash

# Exit on any error, undefined variables, or pipe failures
set -o errexit
set -o nounset
set -o pipefail

# 加载通用配置和版本管理
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Environment variables for MongoDB configuration
# Can be overridden by setting these variables before running the script
PROJ_MONGO_HOST=${PROJ_MONGO_HOST:-127.0.0.1}                          # MongoDB server host
PROJ_MONGO_PORT=${PROJ_MONGO_PORT:-27017}                              # MongoDB server port
PROJ_MONGO_URL=${PROJ_MONGO_HOST}:${PROJ_MONGO_PORT}                   # MongoDB connection URL
PROJ_MONGO_DATABASE=${PROJ_MONGO_DATABASE:-proj}                       # Default database name
PROJ_MONGO_ADMIN_USERNAME=${PROJ_MONGO_ADMIN_USERNAME:-root}           # Admin username
PROJ_MONGO_ADMIN_PASSWORD=${PROJ_MONGO_ADMIN_PASSWORD:-proj(#)666}     # Admin password
PROJ_MONGO_ADMIN_AUTH=${PROJ_MONGO_ADMIN_USERNAME}:${PROJ_MONGO_ADMIN_PASSWORD}  # Auth string
PROJ_MONGO_DATA_DIR=${PROJ_MONGO_DATA_DIR:-/var/lib/mongodb}           # MongoDB data directory
PROJ_MONGO_CONFIG_DIR=${PROJ_MONGO_CONFIG_DIR:-/etc}                   # MongoDB config directory
# 版本信息从统一配置文件加载：MONGODB_VERSION 在 versions.sh 中定义
MONGO_DOCKER_MNAME=${NETWORK_NAME}-mongo

# Install mongo using containerization.
proj::mongo::docker::install()
{
  proj::mongo::pre_install

  proj::common::network
  
  # 清理可能存在的同名容器
  proj::common::docker::cleanup_container "${MONGO_DOCKER_MNAME}"

  # 创建 MongoDB 数据和配置目录
  proj::util::sudo "mkdir -p ${PROJ_THIRDPARTY_INSTALL_DIR}/mongo"
  proj::util::sudo "mkdir -p ${PROJ_THIRDPARTY_INSTALL_DIR}/mongo/config"

  # 创建 MongoDB Docker 配置文件
  local mongo_docker_conf_file="${PROJ_THIRDPARTY_INSTALL_DIR}/mongo/config/mongod.conf"
  local template_docker_conf_file="${SCRIPT_DIR}/mongo/mongod-docker.conf"
  local temp_docker_conf_file="/tmp/mongod-docker.conf.tmp"

  # 使用 envsubst 替换模板中的环境变量
  envsubst < ${template_docker_conf_file} > ${temp_docker_conf_file}

  # 复制配置文件到数据目录
  proj::util::sudo "cp ${temp_docker_conf_file} ${mongo_docker_conf_file}"
  rm -f ${temp_docker_conf_file}

  # 启动 MongoDB 容器，使用配置文件
  docker run -d --name ${MONGO_DOCKER_MNAME} \
    --restart always \
    --network ${NETWORK_NAME} \
    -v ${PROJ_THIRDPARTY_INSTALL_DIR}/mongo:/data \
    -v ${PROJ_THIRDPARTY_INSTALL_DIR}/mongo/config:/etc/mongo \
    -p ${PROJ_ACCESS_HOST}:${PROJ_MONGO_PORT}:27017 \
    -e MONGO_INITDB_ROOT_USERNAME=${PROJ_MONGO_ADMIN_USERNAME} \
    -e MONGO_INITDB_ROOT_PASSWORD=${PROJ_MONGO_ADMIN_PASSWORD} \
    mongodb/mongodb-community-server:${MONGODB_VERSION} \
    --config /etc/mongo/mongod.conf

  sleep 10
  proj::mongo::status || return 1
  proj::mongo::info
  proj::log::info "install mongo successfully"
}


proj::mongo::pre_install()
{
  proj::log::info "Pre-installing MongoDB..."
  
  # 检查 envsubst (gettext-base)
  if ! command -v envsubst >/dev/null 2>&1; then
      proj::log::info "Installing gettext-base for envsubst..."
      proj::util::sudo "apt install -y gettext-base"
  fi

  # 检查 MongoDB 密钥环文件是否已存在
  if [ -f /usr/share/keyrings/mongodb-server-7.0.gpg ]; then
    proj::log::info "MongoDB keyring file already exists, skipping download..."
  else
    # 获取 MongoDB 公钥并添加到现代密钥环
    # 使用 --homedir /tmp/gnupg 避免 GPG 家目录权限警告
    # 使用 --quiet 减少不必要的输出
    echo ${LINUX_PASSWORD} | sudo -S wget -qO - https://www.mongodb.org/static/pgp/server-7.0.asc | sudo gpg --dearmor --homedir /tmp/gnupg --quiet -o /usr/share/keyrings/mongodb-server-7.0.gpg
  fi

  if proj::util::is_ubuntu; then
    # 添加 MongoDB APT 源 - 对于较新的 Ubuntu 版本使用 jammy (22.04) 仓库
    UBUNTU_CODENAME=$(lsb_release -cs)
    # 如果是 noble (24.04) 或更新版本，使用 jammy 仓库
    if [[ "$UBUNTU_CODENAME" == "noble" ]] || [[ "$UBUNTU_CODENAME" > "jammy" ]]; then
      UBUNTU_CODENAME="jammy"
    fi
    echo ${LINUX_PASSWORD} | sudo -S echo "deb [arch=amd64,arm64 signed-by=/usr/share/keyrings/mongodb-server-7.0.gpg] https://repo.mongodb.org/apt/ubuntu ${UBUNTU_CODENAME}/mongodb-org/7.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-7.0.list
  elif proj::util::is_debian; then
    # 添加 MongoDB APT 源
    echo ${LINUX_PASSWORD} | sudo -S echo "deb [arch=amd64,arm64 signed-by=/usr/share/keyrings/mongodb-server-7.0.gpg] https://repo.mongodb.org/apt/debian $(lsb_release -cs)/mongodb-org/7.0 main" | sudo tee /etc/apt/sources.list.d/mongodb-org-7.0.list
  else
    proj::log::error "Unsupported operating system. Only Ubuntu and Debian are supported."
    return 1
  fi

  # 安装libssl1.1，否则安装 mongo 时会报以下错误：
  # mongodb-org-mongos : Depends: libssl1.1 (>= 1.1.1) but it is not installable
  wget http://archive.ubuntu.com/ubuntu/pool/main/o/openssl/libssl1.1_1.1.1f-1ubuntu2_amd64.deb -P /tmp/
  echo ${LINUX_PASSWORD} | sudo -S -i dpkg -i /tmp/libssl1.1_1.1.1f-1ubuntu2_amd64.deb

  proj::util::sudo "apt update"

  # 检查并安装 MongoDB 客户端
  if ! proj::util::cmd_exists "mongosh"; then
    proj::log::info "Installing mongodb-mongosh..."
    proj::util::sudo "apt install -y mongodb-mongosh"
  else
    proj::log::info "mongodb-mongosh already installed, skipping..."
  fi
}

# Uninstall the docker container.
proj::mongo::docker::uninstall()
{
  docker rm -f ${MONGO_DOCKER_MNAME} &>/dev/null
  proj::util::sudo "rm -rf ${PROJ_THIRDPARTY_INSTALL_DIR}/mongo"
  proj::log::info "uninstall mongo successfully"
}

# Install the mongo step by step.
# sbs is the abbreviation for "step by step".
proj::mongo::install()
{
  proj::mongo::pre_install

  echo ${LINUX_PASSWORD} | sudo -S apt install -y gnupg

  # 安装 MongoDB 服务端
  # 以为我们uninstall时会删除配置文件，所以要使用--force-confmiss 重新安装配置文件
  proj::util::sudo "apt -y -o Dpkg::Options::="--force-confmiss" --reinstall install mongodb-org mongodb-org-server"

  # 创建 MongoDB 配置目录
  proj::util::sudo "mkdir -p ${PROJ_MONGO_DATA_DIR}"
  proj::util::sudo "mkdir -p /var/log/mongodb"

  # 设置正确的目录所有权
  proj::util::sudo "chown -R mongodb:mongodb ${PROJ_MONGO_DATA_DIR}"
  proj::util::sudo "chown -R mongodb:mongodb /var/log/mongodb"

  # 创建临时的 MongoDB 配置文件（认证关闭，用于创建用户）
  local mongod_temp_conf_file="/etc/mongod-temp.conf"
  local template_conf_file="${SCRIPT_DIR}/mongo/mongod.conf"
  local temp_conf_file="/tmp/mongod.conf.tmp"

  # 使用 envsubst 替换模板中的环境变量
  envsubst < ${template_conf_file} > ${temp_conf_file}

  # 修改临时配置文件，关闭认证以便创建用户
  sed 's/authorization: enabled/authorization: disabled/' ${temp_conf_file} > /tmp/mongod-temp.conf.tmp

  # 复制临时配置文件到系统目录
  proj::util::sudo "cp /tmp/mongod-temp.conf.tmp ${mongod_temp_conf_file}"
  proj::util::sudo "chown root:root ${mongod_temp_conf_file}"
  rm -f /tmp/mongod-temp.conf.tmp

  # 创建 systemd 服务文件
  local mongod_service_file="/etc/systemd/system/mongod.service"
  local template_service_file="${SCRIPT_DIR}/mongo/mongod.service"
  local temp_service_file="/tmp/mongod.service.tmp"

  # 使用 envsubst 替换模板中的环境变量
  envsubst < ${template_service_file} > ${temp_service_file}

  # 复制服务文件到系统目录
  proj::util::sudo "cp ${temp_service_file} ${mongod_service_file}"
  rm -f ${temp_service_file}

  # 临时启动 MongoDB 用于创建用户（使用临时配置）
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "mongod --config ${mongod_temp_conf_file} --fork"
  echo "Sleeping 5s to wait for mongo to complete startup ..."
  sleep 5

  # 创建管理员账号，设置管理员密码
  mongosh --quiet "mongodb://${PROJ_MONGO_URL}" <<EOF
use admin
db.createUser({user:"${PROJ_MONGO_ADMIN_USERNAME}",pwd:"${PROJ_MONGO_ADMIN_PASSWORD}",roles:["root"]})
db.auth("${PROJ_MONGO_ADMIN_USERNAME}", "${PROJ_MONGO_ADMIN_PASSWORD}")
quit
EOF

  # 停止临时实例
  proj::util::sudo "pkill -f mongod" || true
  sleep 2

  # 创建最终的 MongoDB 配置文件（认证开启）
  local mongod_conf_file="/etc/mongod.conf"
  proj::util::sudo "cp ${temp_conf_file} ${mongod_conf_file}"
  proj::util::sudo "chown root:root ${mongod_conf_file}"
  rm -f ${temp_conf_file}

  # 清理临时配置文件
  proj::util::sudo "rm -f ${mongod_temp_conf_file}"

  # 启动 MongoDB，并设置开机启动
  proj::util::sudo "systemctl enable mongod"

  proj::util::sudo "systemctl restart mongod"

  echo "Sleeping 5s to wait for mongo to complete startup ..."
  sleep 5

  proj::mongo::status || return 1
  proj::mongo::info
  proj::log::info "install mongo successfully"
}

# Uninstall the mongo step by step.
proj::mongo::uninstall()
{
  set +o errexit
  proj::util::sudo "systemctl stop mongod"
  proj::util::sudo "systemctl disable mongod"
  proj::util::sudo "apt remove -y mongodb-org mongodb-org-server" # 这里我们客户端不卸载
  proj::util::sudo "rm -rvf /var/lib/mongodb"
  proj::util::sudo "rm -vf /etc/apt/sources.list.d/mongodb-org-7.0.list"
  proj::util::sudo "rm -vf /etc/mongod.conf"
  proj::util::sudo "rm -vf /lib/systemd/system/mongod.service"
  proj::util::sudo "rm -vf /tmp/mongodb-*.sock"
  set -o errexit

  proj::log::info "uninstall mongo successfully"
}

# Print necessary information after docker or sbs installation.
proj::mongo::info()
{
  echo -e ${C_GREEN}mongo has been installed, here are some useful information:${C_NORMAL}
  encoded=$(echo -n "${PROJ_MONGO_ADMIN_PASSWORD}"|jq -sRr @uri)
  cat << EOF | sed 's/^/  /'
Mongo access url is: mongodb://${PROJ_MONGO_URL}
  Mongo admin username is: ${PROJ_MONGO_ADMIN_USERNAME}
  Mongo admin password is: ${PROJ_MONGO_ADMIN_PASSWORD}
    MongoDB Login Command: mongosh mongodb://${PROJ_MONGO_ADMIN_USERNAME}:'${encoded}'@${PROJ_MONGO_URL}
EOF
}

# Status check after docker or sbs installation.
proj::mongo::status()
{
  proj::util::telnet ${PROJ_MONGO_HOST} ${PROJ_MONGO_PORT} || return 1
}

if [[ "$*" =~ proj::mongo:: ]]; then
  eval $*
fi