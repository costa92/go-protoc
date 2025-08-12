#!/usr/bin/env bash



# The root of the build/dist directory.
PROJ_ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/../..
# If common.sh has already been sourced, it will not be sourced again here.
[[ -z ${COMMON_SOURCED} ]] && source ${PROJ_ROOT_DIR}/scripts/installation/common.sh
# Set some environment variables.
PROJ_MONGO_HOST=${PROJ_MONGO_HOST:-127.0.0.1}
PROJ_MONGO_PORT=${PROJ_MONGO_PORT:-27017}
PROJ_MONGO_URL=${PROJ_MONGO_HOST}:${PROJ_MONGO_PORT}
PROJ_MONGO_DATABASE=${PROJ_MONGO_DATABASE:-proj}
PROJ_MONGO_ADMIN_USERNAME=${PROJ_MONGO_ADMIN_USERNAME:-root}
PROJ_MONGO_ADMIN_PASSWORD=${PROJ_MONGO_ADMIN_PASSWORD:-'proj(#)666'}
#PROJ_MONGO_ADMIN_AUTH=${PROJ_MONGO_ADMIN_USERNAME}:"'${PROJ_MONGO_ADMIN_PASSWORD}'"
PROJ_MONGO_ADMIN_AUTH=${PROJ_MONGO_ADMIN_USERNAME}:${PROJ_MONGO_ADMIN_PASSWORD}

# Install mongo using containerization.
proj::mongo::docker::install()
{
  proj::mongo::pre_install

  proj::common::network
  docker run -d --name onex-mongo \
    --restart always \
    --network onex \
    -v ${PROJ_THIRDPARTY_INSTALL_DIR}/mongo:/data \
    -p ${PROJ_ACCESS_HOST}:${PROJ_MONGO_PORT}:27017 \
    -e MONGO_INITDB_ROOT_USERNAME=${PROJ_MONGO_ADMIN_USERNAME} \
    -e MONGO_INITDB_ROOT_PASSWORD=${PROJ_MONGO_ADMIN_PASSWORD} \
    mongodb/mongodb-community-server:7.0.3-ubuntu2204

  sleep 10
  proj::mongo::status || return 1
  proj::mongo::info
  proj::log::info "install mongo successfully"
}


proj::mongo::pre_install()
{
  # 获取 MongoDB 公钥
  echo ${LINUX_PASSWORD} | sudo -S wget -qO - https://www.mongodb.org/static/pgp/server-7.0.asc | sudo apt-key add -

  # 添加 MongoDB APT 源
  echo ${LINUX_PASSWORD} | sudo -S echo "deb [arch=amd64,arm64] https://repo.mongodb.org/apt/debian $(lsb_release -cs)/mongodb-org/7.0 main" | sudo tee /etc/apt/sources.list.d/mongodb-org-7.0.list

  # 安装libssl1.1，否则安装 mongo 时会报以下错误：
  # mongodb-org-mongos : Depends: libssl1.1 (>= 1.1.1) but it is not installable
  wget http://archive.ubuntu.com/ubuntu/pool/main/o/openssl/libssl1.1_1.1.1f-1ubuntu2_amd64.deb -P /tmp/
  echo ${LINUX_PASSWORD} | sudo -S -i dpkg -i /tmp/libssl1.1_1.1.1f-1ubuntu2_amd64.deb

  proj::util::sudo "apt update"

  # 安装 MongoDB 客户端
  proj::util::sudo "apt install -y mongodb-mongosh"
}

# Uninstall the docker container.
proj::mongo::docker::uninstall()
{
  docker rm -f onex-mongo &>/dev/null
  proj::util::sudo "rm -rf ${PROJ_THIRDPARTY_INSTALL_DIR}/mongo"
  proj::log::info "uninstall mongo successfully"
}

# Install the mongo step by step.
# sbs is the abbreviation for "step by step".
proj::mongo::sbs::install()
{
  proj::mongo::pre_install

  echo ${LINUX_PASSWORD} | sudo -S apt install -y gnupg

  # 安装 MongoDB 服务端
  # 以为我们uninstall时会删除配置文件，所以要使用--force-confmiss 重新安装配置文件
  proj::util::sudo "apt -y -o Dpkg::Options::="--force-confmiss" --reinstall install mongodb-org mongodb-org-server"

  # 开启外网访问权限和登录验证
  echo ${LINUX_PASSWORD} | sudo -S sed -i '/bindIp/{s/127.0.0.1/0.0.0.0/}' /etc/mongod.conf
  # 关闭认证以创建 root 用户
  echo ${LINUX_PASSWORD} | sudo -S sed -i '/^#security/a\security:\n  authorization: disabled' /etc/mongod.conf

  # 启动 MongoDB，并设置开机启动
  proj::util::sudo "systemctl enable mongod"
  proj::util::sudo "systemctl restart mongod"
  echo "Sleeping 5s to wait for mongo to complete startup ..."
  sleep 5

  # 创建管理员账号，设置管理员密码
  echo ${LINUX_PASSWORD} | sudo -S mongosh --quiet "mongodb://${PROJ_MONGO_URL}" <<EOF
use admin
db.createUser({user:"${PROJ_MONGO_ADMIN_USERNAME}",pwd:"${PROJ_MONGO_ADMIN_PASSWORD}",roles:["root"]})
db.auth("${PROJ_MONGO_ADMIN_USERNAME}", "${PROJ_MONGO_ADMIN_PASSWORD}")
quit
EOF
  # 开启认证
  echo ${LINUX_PASSWORD} | sudo -S sed -i '/authorization:/s/disabled/enabled/g' /etc/mongod.conf

  proj::util::sudo "systemctl restart mongod"

  echo "Sleeping 5s to wait for mongo to complete startup ..."
  sleep 5

  proj::mongo::status || return 1
  proj::mongo::info
  proj::log::info "install mongo successfully"
}

# Uninstall the mongo step by step.
proj::mongo::sbs::uninstall()
{
  set +o errexit
  proj::util::sudo "systemctl stop mongodb"
  proj::util::sudo "systemctl disable mongodb"
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