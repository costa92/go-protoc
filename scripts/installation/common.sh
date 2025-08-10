#!/usr/bin/env bash

set -eEuo pipefail

# 获取项目根目录
PROJ_ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/../..

# 都会统一加载 scripts/common.sh 脚本
source "${PROJ_ROOT_DIR}/scripts/common.sh"


COMMON_SOURCED=true # Sourced flag