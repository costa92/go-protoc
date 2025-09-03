#!/usr/bin/env bash

set -eEuo pipefail

# 获取项目根目录 - 兼容 Make 环境
if [[ -n "${BASH_SOURCE[0]:-}" ]]; then
  INSTALLATION_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
else
  # 当在 Make 环境中运行时，BASH_SOURCE 可能未定义
  INSTALLATION_DIR="$(cd "$(dirname "${0}")" && pwd)"
fi
PROJ_ROOT_DIR="${INSTALLATION_DIR}/../.."

# 都会统一加载 scripts/common.sh 脚本
source "${PROJ_ROOT_DIR}/scripts/common.sh"

# 设置 PROJ_ENV_FILE（重要） - 支持动态环境选择
# 优先使用传入的 PROJ_ENVIRONMENT，否则默认为 dev
PROJ_ENVIRONMENT=${PROJ_ENVIRONMENT:-development}
case "${PROJ_ENVIRONMENT}" in
    "development"|"dev")
        PROJ_ENV_FILE=${PROJ_ENV_FILE:-${PROJ_ROOT_DIR}/manifests/env/env.dev}
        ;;
    "test"|"testing")
        PROJ_ENV_FILE=${PROJ_ENV_FILE:-${PROJ_ROOT_DIR}/manifests/env/env.test}
        ;;
    "production"|"prod")
        PROJ_ENV_FILE=${PROJ_ENV_FILE:-${PROJ_ROOT_DIR}/manifests/env/env.prod}
        ;;
    *)
        PROJ_ENV_FILE=${PROJ_ENV_FILE:-${PROJ_ROOT_DIR}/manifests/env/env.dev}
        ;;
esac

# 加载本地安装环境变量（非常重要的一步，后面很多步骤都依赖于env中的变量设置）
source ${PROJ_ENV_FILE}

# 注意：禁止在Docker模板系统中引用versions.sh，版本信息统一从manifests/env获取
# 为兼容性保留，但Docker模板系统不使用
# source "${INSTALLATION_DIR}/versions.sh"

# 容器网络名称
NETWORK_NAME=${NETWORK_NAME:-proj}

COMMON_SOURCED=true # Sourced flag

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

# 清理可能存在的同名 Docker 容器
# 用法: proj::common::docker::cleanup_container "容器名称"
proj::common::docker::cleanup_container()
{
  local container_name="${1}"
  
  if [[ -z "${container_name}" ]]; then
    echo "Error: Container name is required for cleanup" >&2
    return 1
  fi
  
  # 检查容器是否存在，如果存在则删除
  if docker ps -aq -f name="^${container_name}$" | grep -q .; then
    echo "Cleaning up existing container: ${container_name}"
    docker rm -f "${container_name}" 2>/dev/null || {
      echo "Warning: Failed to remove container ${container_name}, but continuing..."
      return 0
    }
  fi
  
  return 0
}
