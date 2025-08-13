#!/usr/bin/env bash
#
# etcd Installation Script
# This script provides functions to install, configure, and manage etcd server
# both natively and via Docker containers.
#

# Exit on any error, undefined variables, or pipe failures
set -o errexit
set -o nounset
set -o pipefail

# Environment variables for etcd configuration
# Can be overridden by setting these variables before running the script
PROJ_ETCD_HOST=${PROJ_ETCD_HOST:-127.0.0.1}                # etcd server host
PROJ_ETCD_PORT=${PROJ_ETCD_PORT:-2379}                     # etcd client port
PROJ_ETCD_PEER_PORT=${PROJ_ETCD_PEER_PORT:-2380}           # etcd peer port
PROJ_ETCD_DATA_DIR=${PROJ_ETCD_DATA_DIR:-/var/lib/etcd}    # etcd data directory
# 加载通用配置和版本管理
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# 版本信息从统一配置文件加载：PROJ_ETCD_VERSION 在 versions.sh 中定义
ETCD_DOCKER_MNAME=${NETWORK_NAME}-etcd

# Function to install etcd natively
proj::etcd::install() {
  proj::etcd::pre_install

  # 创建 etcd 数据目录
  proj::util::sudo "mkdir -p ${PROJ_ETCD_DATA_DIR}"
  proj::util::sudo "chmod 755 ${PROJ_ETCD_DATA_DIR}"

  # 创建 etcd 用户
  if ! id -u etcd >/dev/null 2>&1; then
    proj::util::sudo "useradd --system --shell /bin/false etcd"
  fi

  # 设置正确的目录所有权
  proj::util::sudo "chown -R etcd:etcd ${PROJ_ETCD_DATA_DIR}"

  # 下载 etcd 二进制文件
  local etcd_download_dir="/tmp/etcd-download-test"
  # proj::util::sudo "rm -rf ${etcd_download_dir}"
  mkdir -p ${etcd_download_dir}

  local etcd_arch="amd64"
  if [[ $(uname -m) == "aarch64" ]]; then
    etcd_arch="arm64"
  fi

  local etcd_url="https://github.com/etcd-io/etcd/releases/download/${PROJ_ETCD_VERSION}/etcd-${PROJ_ETCD_VERSION}-linux-${etcd_arch}.tar.gz"

  # 检查文件是否已存在，避免重复下载
  if [[ -f "${etcd_download_dir}/etcd.tar.gz" ]]; then
    proj::log::info "etcd ${PROJ_ETCD_VERSION} tarball already exists, skipping download..."
  else
    proj::log::info "Downloading etcd ${PROJ_ETCD_VERSION} for linux-${etcd_arch}..."
    curl -L ${etcd_url} -o ${etcd_download_dir}/etcd.tar.gz
  fi

  # 解压并安装
  tar xzf ${etcd_download_dir}/etcd.tar.gz -C ${etcd_download_dir} --strip-components=1
  proj::util::sudo "cp ${etcd_download_dir}/etcd /usr/local/bin/"
  proj::util::sudo "cp ${etcd_download_dir}/etcdctl /usr/local/bin/"
  proj::util::sudo "chmod +x /usr/local/bin/etcd"
  proj::util::sudo "chmod +x /usr/local/bin/etcdctl"

  # 创建 systemd 服务文件
  local etcd_service_file="/etc/systemd/system/etcd.service"

  # 创建临时文件
  local temp_service_file="/tmp/etcd.service.tmp"
  cat > ${temp_service_file} << 'EOF'
[Unit]
Description=etcd key-value store
Documentation=https://github.com/etcd-io/etcd
After=network.target

[Service]
Type=notify
User=etcd
ExecStart=/usr/local/bin/etcd \
  --name=etcd-node1 \
  --data-dir=/var/lib/etcd \
  --listen-client-urls=http://0.0.0.0:2379 \
  --advertise-client-urls=http://127.0.0.1:2379 \
  --listen-peer-urls=http://0.0.0.0:2380 \
  --initial-advertise-peer-urls=http://127.0.0.1:2380 \
  --initial-cluster=etcd-node1=http://127.0.0.1:2380 \
  --initial-cluster-token=etcd-cluster-1 \
  --initial-cluster-state=new \
  --log-level=info \
  --logger=zap \
  --log-outputs=stderr
Restart=always
RestartSec=10s
LimitNOFILE=40000

[Install]
WantedBy=multi-user.target
EOF

  # 复制到系统目录
  proj::util::sudo "cp ${temp_service_file} ${etcd_service_file}"
  rm -f ${temp_service_file}

  # 启动 etcd 服务
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "systemctl enable etcd"
  proj::util::sudo "systemctl start etcd"

  # 清理下载目录
  # rm -rf ${etcd_download_dir}

  sleep 3
  proj::etcd::status || return 1
  proj::etcd::info
  proj::log::info "install etcd successfully"
}

# Uninstall etcd step by step
proj::etcd::uninstall() {
  set +o errexit
  proj::util::sudo "systemctl stop etcd"
  proj::util::sudo "systemctl disable etcd"
  proj::util::sudo "rm -f /etc/systemd/system/etcd.service"
  proj::util::sudo "systemctl daemon-reload"
  proj::util::sudo "rm -f /usr/local/bin/etcd"
  proj::util::sudo "rm -f /usr/local/bin/etcdctl"
  proj::util::sudo "rm -rf ${PROJ_ETCD_DATA_DIR}"
  proj::util::sudo "userdel etcd" 2>/dev/null || true
  set -o errexit
  proj::log::info "uninstall etcd successfully"
  return 0
}

# Pre-install etcd by checking system requirements
proj::etcd::pre_install() {
  proj::log::info "Pre-installing etcd..."

  # 判断是 mac 还是 linux
  if proj::util::is_mac; then
    proj::log::info "Mac OS detected, checking for etcd installation..."
    if ! command -v etcd >/dev/null 2>&1; then
      proj::log::info "Installing etcd via brew..."
      brew install etcd
    else
      proj::log::info "etcd already installed, skipping..."
    fi
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
  fi
}

# Install etcd using a Docker container
proj::etcd::docker::install() {
  proj::log::info "Installing docker etcd..."

  proj::etcd::pre_install
  proj::common::network

  # 创建数据目录
  local etcd_data_dir="${PROJ_THIRDPARTY_INSTALL_DIR}/etcd"
  mkdir -p ${etcd_data_dir}

  docker run -d --name ${ETCD_DOCKER_MNAME} \
    --restart always \
    --network ${NETWORK_NAME} \
    -v ${etcd_data_dir}:/etcd-data \
    -p ${PROJ_ETCD_HOST}:${PROJ_ETCD_PORT}:2379 \
    -p ${PROJ_ETCD_HOST}:${PROJ_ETCD_PEER_PORT}:2380 \
    gcr.io/etcd-development/etcd:${PROJ_ETCD_VERSION} \
    /usr/local/bin/etcd \
    --name=etcd-docker \
    --data-dir=/etcd-data \
    --listen-client-urls=http://0.0.0.0:2379 \
    --advertise-client-urls=http://${PROJ_ETCD_HOST}:${PROJ_ETCD_PORT} \
    --listen-peer-urls=http://0.0.0.0:2380 \
    --initial-advertise-peer-urls=http://${PROJ_ETCD_HOST}:${PROJ_ETCD_PEER_PORT} \
    --initial-cluster=etcd-docker=http://${PROJ_ETCD_HOST}:${PROJ_ETCD_PEER_PORT} \
    --initial-cluster-token=etcd-cluster-1 \
    --initial-cluster-state=new \
    --log-level=info \
    --logger=zap \
    --log-outputs=stderr

  sleep 3
  if proj::util::is_linux; then
    proj::etcd::status || return 1
  fi
  proj::etcd::info
  proj::log::info "install etcd successfully"
}

# Print necessary information after docker or native installation
proj::etcd::info() {
  echo -e ${C_GREEN}etcd has been installed, here are some useful information:${C_NORMAL}
  cat << EOF | sed 's/^/  /'
etcd access endpoint is: ${PROJ_ETCD_HOST}:${PROJ_ETCD_PORT}
     etcd peer endpoint: ${PROJ_ETCD_HOST}:${PROJ_ETCD_PEER_PORT}
         etcd data dir: ${PROJ_ETCD_DATA_DIR}
     etcd version info: etcdctl version
   etcd cluster health: etcdctl --endpoints=http://${PROJ_ETCD_HOST}:${PROJ_ETCD_PORT} endpoint health
        etcd put test: etcdctl --endpoints=http://${PROJ_ETCD_HOST}:${PROJ_ETCD_PORT} put /test "hello world"
        etcd get test: etcdctl --endpoints=http://${PROJ_ETCD_HOST}:${PROJ_ETCD_PORT} get /test
EOF
}

# Uninstall the docker container
proj::etcd::docker::uninstall() {
  docker rm -f ${ETCD_DOCKER_MNAME} &>/dev/null
  proj::util::sudo "rm -rf ${PROJ_THIRDPARTY_INSTALL_DIR}/etcd"
  proj::log::info "uninstall etcd successfully"
}

# Status check after docker or native installation
proj::etcd::status() {
  proj::util::telnet ${PROJ_ETCD_HOST} ${PROJ_ETCD_PORT} || return 1

  # 检查 etcd 健康状态
  if command -v etcdctl >/dev/null 2>&1; then
    etcdctl --endpoints=http://${PROJ_ETCD_HOST}:${PROJ_ETCD_PORT} endpoint health || {
      proj::log::error "etcd health check failed, etcd maybe not initialized properly."
      return 1
    }
  else
    proj::log::info "etcdctl not found, skipping health check"
  fi
}

# This block allows calling functions in this script directly from the command line.
# It checks if the script's arguments contain a function name with the 'proj::etcd::' prefix
# and, if so, executes that function.
# For example: ./etcd.sh proj::etcd::install
if [[ "$*" =~ proj::etcd:: ]]; then
  eval $*
fi