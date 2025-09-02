#!/usr/bin/env bash

# =============================================================================
# 配置管理器
# Configuration Manager Library
# 
# 提供统一的配置文件管理、环境变量处理和配置验证功能
# =============================================================================

set -eEuo pipefail

# 避免重复加载
[[ "${CONFIG_MANAGER_LOADED:-false}" == "true" ]] && return 0

# 默认配置目录
CONFIG_MANAGER_CONFIG_DIR="${PROJ_ROOT_DIR:-$(pwd)}/configs"
CONFIG_MANAGER_DEFAULT_ENV="development"

# 配置文件后缀
CONFIG_MANAGER_SUPPORTED_FORMATS=("yaml" "yml" "json" "env")

# =============================================================================
# 核心配置管理函数
# =============================================================================

# 获取配置文件路径
proj::config::get_config_path() {
    local service_name="$1"
    local env="${2:-${PROJ_ENVIRONMENT:-${CONFIG_MANAGER_DEFAULT_ENV}}}"
    local format="${3:-yaml}"
    
    echo "${CONFIG_MANAGER_CONFIG_DIR}/${service_name}_${env}.${format}"
}

# 检查配置文件是否存在
proj::config::exists() {
    local config_path="$1"
    [[ -f "${config_path}" ]]
}

# 验证配置文件格式
proj::config::validate_format() {
    local config_path="$1"
    local format
    
    format=$(basename "${config_path}" | cut -d'.' -f2-)
    
    case "${format}" in
        yaml|yml)
            # 基本YAML语法检查
            if command -v yq >/dev/null 2>&1; then
                yq eval '.' "${config_path}" >/dev/null 2>&1
            else
                # 简单的语法检查
                python3 -c "import yaml; yaml.safe_load(open('${config_path}'))" 2>/dev/null
            fi
            ;;
        json)
            if command -v jq >/dev/null 2>&1; then
                jq empty "${config_path}" >/dev/null 2>&1
            else
                python3 -c "import json; json.load(open('${config_path}'))" 2>/dev/null
            fi
            ;;
        *)
            proj::log::warn "Unknown config format: ${format}"
            return 0
            ;;
    esac
}

# 加载环境变量配置
proj::config::load_env() {
    local env_file="$1"
    
    if [[ -f "${env_file}" ]]; then
        set -a
        source "${env_file}"
        set +a
        proj::log::debug "Loaded environment variables from: ${env_file}"
    else
        proj::log::warn "Environment file not found: ${env_file}"
        return 1
    fi
}

# 设置默认配置值
proj::config::set_defaults() {
    export PROJ_ENVIRONMENT="${PROJ_ENVIRONMENT:-development}"
    export PROJ_NAME="${PROJ_NAME:-go-protoc}"
    export PROJ_PREFIX="${PROJ_PREFIX:-proj}"
    export PROJ_ROOT_DIR="${PROJ_ROOT_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
    
    proj::log::debug "Set default configuration values"
}

# =============================================================================
# 环境变量处理函数
# =============================================================================

# 获取环境变量值（支持默认值）
proj::config::get_env() {
    local var_name="$1"
    local default_value="${2:-}"
    
    echo "${!var_name:-${default_value}}"
}

# 验证必需的环境变量
proj::config::require_env() {
    local var_name="$1"
    local error_msg="${2:-Environment variable ${var_name} is required}"
    
    if [[ -z "${!var_name:-}" ]]; then
        proj::log::error "${error_msg}"
        return 1
    fi
}

# 导出配置到环境变量
proj::config::export_vars() {
    local -a var_names=("$@")
    
    for var in "${var_names[@]}"; do
        if [[ -n "${!var:-}" ]]; then
            export "${var}"
            proj::log::debug "Exported: ${var}=${!var}"
        fi
    done
}

# =============================================================================
# 配置模板处理
# =============================================================================

# 处理配置模板（环境变量替换）
proj::config::process_template() {
    local template_file="$1"
    local output_file="$2"
    
    if [[ ! -f "${template_file}" ]]; then
        proj::log::error "Template file not found: ${template_file}"
        return 1
    fi
    
    # 确保输出目录存在
    local output_dir
    output_dir=$(dirname "${output_file}")
    mkdir -p "${output_dir}"
    
    # 使用envsubst处理模板
    if command -v envsubst >/dev/null 2>&1; then
        envsubst < "${template_file}" > "${output_file}"
    else
        # 简单的替换实现
        cp "${template_file}" "${output_file}"
        proj::log::warn "envsubst not available, template variables may not be processed"
    fi
    
    proj::log::debug "Processed template: ${template_file} -> ${output_file}"
}

# =============================================================================
# 服务特定配置
# =============================================================================

# 加载服务配置
proj::config::load_service_config() {
    local service_name="$1"
    local env="${2:-${PROJ_ENVIRONMENT:-development}}"
    
    local config_path
    config_path=$(proj::config::get_config_path "${service_name}" "${env}")
    
    if proj::config::exists "${config_path}"; then
        if proj::config::validate_format "${config_path}"; then
            proj::log::info "Loaded ${service_name} configuration: ${config_path}"
            return 0
        else
            proj::log::error "Invalid configuration format: ${config_path}"
            return 1
        fi
    else
        # 尝试默认配置
        local default_config
        default_config=$(proj::config::get_config_path "${service_name}" "default")
        
        if proj::config::exists "${default_config}"; then
            proj::log::info "Using default configuration: ${default_config}"
            return 0
        else
            proj::log::warn "No configuration found for service: ${service_name}"
            return 1
        fi
    fi
}

# 生成服务配置环境变量
proj::config::generate_service_vars() {
    local service_name="$1"
    local service_upper
    service_upper=$(echo "${service_name}" | tr '[:lower:]' '[:upper:]')
    
    # 基础服务变量
    export "PROJ_${service_upper}_SERVICE_NAME=${service_name}"
    export "PROJ_${service_upper}_CONFIG_DIR=${PROJ_ROOT_DIR}/configs"
    export "PROJ_${service_upper}_DATA_DIR=${PROJ_ROOT_DIR}/_thirdparty/${service_name}/data"
    export "PROJ_${service_upper}_LOG_DIR=${PROJ_ROOT_DIR}/_thirdparty/${service_name}/logs"
    
    proj::log::debug "Generated environment variables for service: ${service_name}"
}

# =============================================================================
# 初始化函数
# =============================================================================

# 初始化配置管理器
proj::config::init() {
    proj::config::set_defaults
    
    # 创建必要的目录
    mkdir -p "${CONFIG_MANAGER_CONFIG_DIR}"
    mkdir -p "${PROJ_ROOT_DIR}/_thirdparty"
    
    proj::log::debug "Configuration manager initialized"
}

# 自动初始化
if [[ "${BASH_SOURCE[0]}" == "${0}" ]] || [[ "${CONFIG_MANAGER_AUTO_INIT:-true}" == "true" ]]; then
    proj::config::init
fi

# 标记已加载
export CONFIG_MANAGER_LOADED=true