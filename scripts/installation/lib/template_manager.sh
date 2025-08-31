#!/usr/bin/env bash

# =============================================================================
# 模板管理器
# Template Manager
#
# 提供模板变量替换和配置文件生成功能
# 支持多平台、多服务的配置模板管理
# =============================================================================

set -eEuo pipefail

# 避免重复加载
[[ "${TEMPLATE_MANAGER_LIB_LOADED:-false}" == "true" ]] && return 0

# 模板目录常量
readonly TEMPLATES_ROOT_DIR="${PROJ_ROOT_DIR}/scripts/installation/templates"
readonly DOCKER_TEMPLATES_DIR="${TEMPLATES_ROOT_DIR}/docker"
readonly UBUNTU_TEMPLATES_DIR="${TEMPLATES_ROOT_DIR}/ubuntu"
readonly MACOS_TEMPLATES_DIR="${TEMPLATES_ROOT_DIR}/macos"

# 生成配置文件
proj::template::generate_config() {
    local service_name="$1"
    local platform="$2"       # docker|ubuntu|macos
    local template_name="$3"   # config.yaml.tpl|systemd.service.tpl|etc.
    local output_path="$4"
    local extra_vars="${5:-}"  # 额外的环境变量文件路径
    
    if [[ -z "$service_name" ]] || [[ -z "$platform" ]] || [[ -z "$template_name" ]] || [[ -z "$output_path" ]]; then
        proj::log::error "Service name, platform, template name, and output path are required"
        return 1
    fi
    
    # 确定模板路径
    local template_dir
    case "$platform" in
        "docker")
            template_dir="$DOCKER_TEMPLATES_DIR"
            ;;
        "ubuntu")
            template_dir="$UBUNTU_TEMPLATES_DIR"
            ;;
        "macos")
            template_dir="$MACOS_TEMPLATES_DIR"
            ;;
        *)
            proj::log::error "Unsupported platform: $platform"
            return 1
            ;;
    esac
    
    local template_path="$template_dir/$service_name/$template_name"
    
    if [[ ! -f "$template_path" ]]; then
        proj::log::error "Template not found: $template_path"
        return 1
    fi
    
    proj::log::info "Generating config from template: $template_path -> $output_path"
    
    # 加载额外的变量文件
    if [[ -n "$extra_vars" ]] && [[ -f "$extra_vars" ]]; then
        proj::log::debug "Loading extra variables from: $extra_vars"
        # shellcheck disable=SC1090
        source "$extra_vars"
    fi
    
    # 确保输出目录存在
    local output_dir
    output_dir=$(dirname "$output_path")
    mkdir -p "$output_dir"
    
    # 使用 envsubst 进行变量替换
    if command -v envsubst >/dev/null 2>&1; then
        proj::template::envsubst_generate "$template_path" "$output_path"
    else
        proj::template::bash_generate "$template_path" "$output_path"
    fi
    
    proj::log::success "Configuration file generated: $output_path"
}

# 使用 envsubst 进行变量替换（推荐）
proj::template::envsubst_generate() {
    local template_path="$1"
    local output_path="$2"
    
    # 获取模板中的所有变量
    local template_vars
    template_vars=$(grep -oE '\$\{[A-Za-z_][A-Za-z0-9_]*[^}]*\}' "$template_path" | sort -u | tr '\n' ' ')
    
    proj::log::debug "Template variables: $template_vars"
    
    # 使用 envsubst 替换变量
    envsubst "$template_vars" < "$template_path" > "$output_path"
}

# 使用 bash 进行变量替换（备用方案）
proj::template::bash_generate() {
    local template_path="$1"
    local output_path="$2"
    
    proj::log::warn "envsubst not available, using bash variable substitution"
    
    # 读取模板内容并进行变量替换
    local content
    content=$(<"$template_path")
    
    # 使用 eval 进行变量替换（注意安全性）
    content=$(eval "cat << 'TEMPLATE_EOF'
$content
TEMPLATE_EOF")
    
    # 写入输出文件
    echo "$content" > "$output_path"
}

# 批量生成服务配置
proj::template::generate_service_configs() {
    local service_name="$1"
    local platform="$2"
    local config_dir="$3"
    local extra_vars="${4:-}"
    
    if [[ -z "$service_name" ]] || [[ -z "$platform" ]] || [[ -z "$config_dir" ]]; then
        proj::log::error "Service name, platform, and config directory are required"
        return 1
    fi
    
    proj::log::info "Generating all configuration files for $service_name on $platform"
    
    # 确定模板目录
    local template_dir
    case "$platform" in
        "docker")
            template_dir="$DOCKER_TEMPLATES_DIR/$service_name"
            ;;
        "ubuntu")
            template_dir="$UBUNTU_TEMPLATES_DIR/$service_name"
            ;;
        "macos")
            template_dir="$MACOS_TEMPLATES_DIR/$service_name"
            ;;
        *)
            proj::log::error "Unsupported platform: $platform"
            return 1
            ;;
    esac
    
    if [[ ! -d "$template_dir" ]]; then
        proj::log::error "Template directory not found: $template_dir"
        return 1
    fi
    
    # 生成所有模板文件
    local generated_count=0
    while IFS= read -r -d '' template_file; do
        local template_name
        template_name=$(basename "$template_file")
        
        # 移除 .tpl 后缀
        local output_name="${template_name%.tpl}"
        local output_path="$config_dir/$output_name"
        
        if proj::template::generate_config "$service_name" "$platform" "$template_name" "$output_path" "$extra_vars"; then
            ((generated_count++))
        fi
    done < <(find "$template_dir" -name "*.tpl" -type f -print0)
    
    proj::log::success "Generated $generated_count configuration files for $service_name"
}

# 验证模板文件
proj::template::validate_template() {
    local template_path="$1"
    
    if [[ ! -f "$template_path" ]]; then
        proj::log::error "Template file does not exist: $template_path"
        return 1
    fi
    
    proj::log::info "Validating template: $template_path"
    
    # 检查模板语法
    local template_vars
    template_vars=$(grep -oE '\$\{[A-Za-z_][A-Za-z0-9_]*[^}]*\}' "$template_path" || true)
    
    if [[ -n "$template_vars" ]]; then
        proj::log::debug "Found template variables:"
        echo "$template_vars" | while read -r var; do
            proj::log::debug "  $var"
        done
    fi
    
    # 检查未闭合的变量引用
    if grep -q '\${[^}]*$' "$template_path"; then
        proj::log::error "Found unclosed variable references in template"
        grep -n '\${[^}]*$' "$template_path" | while read -r line; do
            proj::log::error "  Line: $line"
        done
        return 1
    fi
    
    proj::log::success "Template validation passed: $template_path"
}

# 列出可用的模板
proj::template::list_templates() {
    local platform="${1:-}"
    local service="${2:-}"
    
    if [[ -z "$platform" ]]; then
        proj::log::info "Available platforms:"
        for platform_dir in "$DOCKER_TEMPLATES_DIR" "$UBUNTU_TEMPLATES_DIR" "$MACOS_TEMPLATES_DIR"; do
            if [[ -d "$platform_dir" ]]; then
                local platform_name
                platform_name=$(basename "$platform_dir")
                echo "  $platform_name"
            fi
        done
        return 0
    fi
    
    # 确定模板目录
    local template_base_dir
    case "$platform" in
        "docker")
            template_base_dir="$DOCKER_TEMPLATES_DIR"
            ;;
        "ubuntu")
            template_base_dir="$UBUNTU_TEMPLATES_DIR"
            ;;
        "macos")
            template_base_dir="$MACOS_TEMPLATES_DIR"
            ;;
        *)
            proj::log::error "Unsupported platform: $platform"
            return 1
            ;;
    esac
    
    if [[ ! -d "$template_base_dir" ]]; then
        proj::log::error "Template directory not found: $template_base_dir"
        return 1
    fi
    
    if [[ -z "$service" ]]; then
        proj::log::info "Available services for platform '$platform':"
        for service_dir in "$template_base_dir"/*; do
            if [[ -d "$service_dir" ]]; then
                local service_name
                service_name=$(basename "$service_dir")
                echo "  $service_name"
            fi
        done
        return 0
    fi
    
    # 列出特定服务的模板
    local service_template_dir="$template_base_dir/$service"
    if [[ ! -d "$service_template_dir" ]]; then
        proj::log::error "Service template directory not found: $service_template_dir"
        return 1
    fi
    
    proj::log::info "Available templates for service '$service' on platform '$platform':"
    find "$service_template_dir" -name "*.tpl" -type f | while read -r template; do
        local template_name
        template_name=$(basename "$template")
        echo "  $template_name"
    done
}

# 创建新模板
proj::template::create_template() {
    local service_name="$1"
    local platform="$2"
    local template_name="$3"
    local template_content="${4:-}"
    
    if [[ -z "$service_name" ]] || [[ -z "$platform" ]] || [[ -z "$template_name" ]]; then
        proj::log::error "Service name, platform, and template name are required"
        return 1
    fi
    
    # 确定模板目录
    local template_dir
    case "$platform" in
        "docker")
            template_dir="$DOCKER_TEMPLATES_DIR/$service_name"
            ;;
        "ubuntu")
            template_dir="$UBUNTU_TEMPLATES_DIR/$service_name"
            ;;
        "macos")
            template_dir="$MACOS_TEMPLATES_DIR/$service_name"
            ;;
        *)
            proj::log::error "Unsupported platform: $platform"
            return 1
            ;;
    esac
    
    # 确保模板目录存在
    mkdir -p "$template_dir"
    
    local template_path="$template_dir/$template_name"
    
    if [[ -f "$template_path" ]]; then
        proj::log::error "Template already exists: $template_path"
        return 1
    fi
    
    # 创建模板文件
    if [[ -n "$template_content" ]]; then
        echo "$template_content" > "$template_path"
    else
        # 创建基本模板结构
        cat > "$template_path" << EOF
# Configuration template for $service_name on $platform
# Project: \${PROJ_NAME:-go-protoc}
# Service: $service_name \${$(echo "$service_name" | tr '[:lower:]' '[:upper:]')_VERSION}
# Environment: \${PROJ_ENVIRONMENT:-development}

# Add your configuration here
EOF
    fi
    
    proj::log::success "Template created: $template_path"
}

# 标记已加载
export TEMPLATE_MANAGER_LIB_LOADED=true

# 如果直接执行此脚本，显示可用模板
if [[ "${BASH_SOURCE[0]:-${0##*/}}" == "${0##*/}" ]]; then
    proj::template::list_templates "$@"
fi