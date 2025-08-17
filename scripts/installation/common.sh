#!/usr/bin/env bash

set -eEuo pipefail

# 获取项目根目录
INSTALLATION_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-${0}}")" && pwd)"
PROJ_ROOT_DIR="${INSTALLATION_DIR}/../.."

# 都会统一加载 scripts/common.sh 脚本
source "${PROJ_ROOT_DIR}/scripts/common.sh"

# 加载统一版本管理配置
source "${INSTALLATION_DIR}/versions.sh"

# 容器网络名称
NETWORK_NAME=${NETWORK_NAME:-proj}


COMMON_SOURCED=true # Sourced flag

# 设置 PROJ_ENV_FILE（重要）
PROJ_ENV_FILE=${PROJ_ENV_FILE:-${PROJ_ROOT_DIR}/manifests/env/env.dev}
# 加载本地安装环境变量（非常重要的一步，后面很多步骤都依赖于env.local中的变量设置）
source ${PROJ_ENV_FILE}

# 确保 proj 容器网络存在。
# 在 uninstall 时，可不删除 proj 容器网络，可以作为一个无害的无用数据
proj::common::network()
{
  if ! docker network ls | grep -q ${NETWORK_NAME}; then
    docker network create ${NETWORK_NAME} || {
      proj::log::info "Network ${NETWORK_NAME} already exists or failed to create, continuing..."
      return 0
    }
  fi
}
