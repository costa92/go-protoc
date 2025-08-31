#!/usr/bin/env bash

# =============================================================================
# 标准化工具库加载器
# Common Library Loader for Installation Scripts
# 
# 此文件负责加载所有标准化工具库，提供统一的函数库管理机制
# =============================================================================

set -eEuo pipefail

# 避免重复加载
[[ "${COMMON_LIB_LOADED:-false}" == "true" ]] && return 0

# 获取工具库目录
LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALLATION_DIR="$(cd "${LIB_DIR}/.." && pwd)"
PROJ_ROOT_DIR="$(cd "${INSTALLATION_DIR}/../.." && pwd)"

# 确保基础common.sh已加载
source "${INSTALLATION_DIR}/common.sh"

# 工具库文件列表
TOOL_LIBRARIES=(
    "platform.sh"
    "docker_helper.sh"
    "docker_script_manager.sh"
    "native_helper.sh"
    "ubuntu_adapter.sh"
    "macos_adapter.sh"
    "template_manager.sh"
    "config_manager.sh"
    "health_check.sh"
    "retry.sh"
)

# 安全加载工具库函数
proj::lib::load_library() {
    local lib_file="$1"
    local lib_path="${LIB_DIR}/${lib_file}"
    
    if [[ -f "${lib_path}" ]]; then
        source "${lib_path}"
        proj::log::debug "Loaded library: ${lib_file}"
    else
        proj::log::warn "Library not found: ${lib_file} (${lib_path})"
    fi
}

# 加载所有可用的工具库
proj::lib::load_all() {
    proj::log::info "Loading standardized tool libraries..."
    
    for lib in "${TOOL_LIBRARIES[@]}"; do
        proj::lib::load_library "${lib}"
    done
    
    proj::log::info "Tool libraries loaded successfully"
}

# 检查库依赖
proj::lib::check_dependencies() {
    local missing_deps=()
    
    # 检查必需的系统命令
    local required_commands=("curl" "tar" "envsubst")
    for cmd in "${required_commands[@]}"; do
        if ! command -v "${cmd}" >/dev/null 2>&1; then
            missing_deps+=("${cmd}")
        fi
    done
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        proj::log::error "Missing required dependencies: ${missing_deps[*]}"
        proj::log::error "Please install missing dependencies before proceeding"
        return 1
    fi
    
    return 0
}

# 库版本信息
proj::lib::version() {
    echo "Standardized Tool Libraries v1.0.0"
    echo "Loaded from: ${LIB_DIR}"
    echo "Available libraries: ${TOOL_LIBRARIES[*]}"
}

# 自动加载（如果不是在测试环境中）
if [[ "${BASH_SOURCE[0]}" == "${0}" ]] || [[ "${LIB_AUTO_LOAD:-true}" == "true" ]]; then
    proj::lib::check_dependencies || exit 1
    proj::lib::load_all
fi

# 标记已加载
export COMMON_LIB_LOADED=true