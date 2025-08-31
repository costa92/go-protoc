#!/usr/bin/env bash

# =============================================================================
# 平台检测工具库
# Platform Detection Library
#
# 提供跨平台的操作系统检测、Docker检测和安装方式推荐功能
# =============================================================================

set -eEuo pipefail

# 避免重复加载
[[ "${PLATFORM_LIB_LOADED:-false}" == "true" ]] && return 0

# 检测操作系统类型
proj::platform::detect_os() {
    local os_type=""
    
    # 检测操作系统
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux系统，进一步检测发行版
        if [[ -f /etc/os-release ]]; then
            local os_id
            os_id=$(grep '^ID=' /etc/os-release | cut -d'=' -f2 | tr -d '"')
            case "$os_id" in
                "ubuntu"|"debian")
                    os_type="ubuntu"
                    ;;
                "centos"|"rhel"|"fedora"|"rocky"|"almalinux")
                    os_type="centos"  # 暂不支持，但可扩展
                    ;;
                *)
                    os_type="linux_other"
                    ;;
            esac
        else
            os_type="linux_unknown"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        os_type="macos"
    elif [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "msys" ]]; then
        os_type="windows"
    else
        os_type="unknown"
    fi
    
    echo "$os_type"
}

# 检测是否支持Docker
proj::platform::has_docker() {
    if command -v docker >/dev/null 2>&1; then
        # Docker命令存在，检查是否可以连接到Docker daemon
        if docker info >/dev/null 2>&1; then
            return 0  # Docker可用
        else
            proj::log::warn "Docker command found but daemon not accessible"
            return 1  # Docker命令存在但不可用
        fi
    else
        return 1  # Docker不存在
    fi
}

# 检测包管理器
proj::platform::get_package_manager() {
    local os_type
    os_type=$(proj::platform::detect_os)
    
    case "$os_type" in
        "ubuntu")
            echo "apt"
            ;;
        "macos")
            if command -v brew >/dev/null 2>&1; then
                echo "brew"
            else
                echo "none"  # macOS但没有Homebrew
            fi
            ;;
        "centos")
            echo "yum"
            ;;
        *)
            echo "unknown"
            ;;
    esac
}

# 获取推荐的安装方式
proj::platform::get_preferred_install_method() {
    local service_name="$1"
    local os_type
    local has_docker=false
    
    os_type=$(proj::platform::detect_os)
    proj::platform::has_docker && has_docker=true
    
    # 检查服务偏好配置（从versions.sh加载）
    local service_upper=$(echo "$service_name" | tr '[:lower:]' '[:upper:]')
    local os_upper=$(echo "$os_type" | tr '[:lower:]' '[:upper:]')
    local preference_var="INSTALL_PREFERENCE_${service_upper}_${os_upper}"
    local preference=""
    # 兼容旧版bash的变量存在性检查
    if [[ -n "${!preference_var:-}" ]]; then
        preference="${!preference_var}"
    fi
    
    # 如果有明确的偏好配置
    if [[ -n "$preference" ]]; then
        case "$preference" in
            "docker")
                if [[ "$has_docker" == "true" ]]; then
                    echo "docker"
                else
                    proj::log::warn "Docker preferred for $service_name on $os_type but Docker not available, falling back to native"
                    echo "native"
                fi
                ;;
            "native")
                echo "native"
                ;;
            *)
                proj::log::warn "Unknown preference '$preference' for $service_name on $os_type"
                proj::platform::get_default_install_method "$service_name" "$os_type" "$has_docker"
                ;;
        esac
    else
        # 使用默认策略
        proj::platform::get_default_install_method "$service_name" "$os_type" "$has_docker"
    fi
}

# 获取默认安装方式（内部函数）
proj::platform::get_default_install_method() {
    local service_name="$1"
    local os_type="$2"
    local has_docker="$3"
    
    case "$os_type" in
        "ubuntu")
            # Ubuntu默认优先使用Docker（如果可用）
            if [[ "$has_docker" == "true" ]]; then
                echo "docker"
            else
                echo "native"
            fi
            ;;
        "macos")
            # macOS根据服务类型决定
            case "$service_name" in
                "redis"|"prometheus"|"grafana")
                    # 这些服务在macOS上优先使用原生安装（Homebrew）
                    echo "native"
                    ;;
                *)
                    # 其他服务优先使用Docker
                    if [[ "$has_docker" == "true" ]]; then
                        echo "docker"
                    else
                        echo "native"
                    fi
                    ;;
            esac
            ;;
        *)
            # 其他平台默认使用Docker（如果可用）
            if [[ "$has_docker" == "true" ]]; then
                echo "docker"
            else
                proj::log::warn "Unsupported platform: $os_type, Docker installation recommended"
                echo "docker"
            fi
            ;;
    esac
}

# 检查平台是否支持指定服务
proj::platform::is_service_supported() {
    local service_name="$1"
    local install_method="${2:-auto}"  # docker|native|auto
    local os_type
    
    os_type=$(proj::platform::detect_os)
    
    # 检查Docker支持
    if [[ "$install_method" == "docker" ]] || [[ "$install_method" == "auto" ]]; then
        if ! proj::platform::has_docker; then
            if [[ "$install_method" == "docker" ]]; then
                proj::log::error "Docker installation requested but Docker not available on $os_type"
                return 1
            fi
        fi
    fi
    
    # 检查原生安装支持
    if [[ "$install_method" == "native" ]] || [[ "$install_method" == "auto" ]]; then
        case "$os_type" in
            "ubuntu"|"macos")
                return 0  # 支持原生安装
                ;;
            *)
                if [[ "$install_method" == "native" ]]; then
                    proj::log::error "Native installation requested but not supported on $os_type"
                    return 1
                fi
                ;;
        esac
    fi
    
    return 0
}

# 显示平台信息
proj::platform::show_info() {
    local os_type pkg_manager has_docker install_method
    
    os_type=$(proj::platform::detect_os)
    pkg_manager=$(proj::platform::get_package_manager)
    
    if proj::platform::has_docker; then
        has_docker="Yes"
    else
        has_docker="No"
    fi
    
    cat << EOF
Platform Information:
  Operating System: $os_type
  Package Manager: $pkg_manager
  Docker Available: $has_docker
  
Recommended Installation Methods:
EOF
    
    # 显示几个主要服务的推荐安装方式
    local services=("redis" "otelcol" "prometheus" "victorialogs")
    for service in "${services[@]}"; do
        install_method=$(proj::platform::get_preferred_install_method "$service")
        echo "  $service: $install_method"
    done
}

# 验证平台兼容性
proj::platform::validate_compatibility() {
    local os_type
    os_type=$(proj::platform::detect_os)
    
    case "$os_type" in
        "ubuntu"|"macos")
            proj::log::info "Platform $os_type is fully supported"
            return 0
            ;;
        "linux_other"|"linux_unknown")
            proj::log::warn "Linux distribution may be supported with Docker installation only"
            if proj::platform::has_docker; then
                proj::log::info "Docker available, most services should work"
                return 0
            else
                proj::log::error "Unsupported Linux distribution without Docker"
                return 1
            fi
            ;;
        "windows")
            proj::log::warn "Windows platform has limited support (Docker only)"
            if proj::platform::has_docker; then
                return 0
            else
                proj::log::error "Windows platform requires Docker for service installation"
                return 1
            fi
            ;;
        *)
            proj::log::error "Unsupported platform: $os_type"
            return 1
            ;;
    esac
}

# 获取系统架构
proj::platform::get_architecture() {
    local arch
    arch=$(uname -m)
    
    case "$arch" in
        "x86_64")
            echo "amd64"
            ;;
        "aarch64"|"arm64")
            echo "arm64"
            ;;
        "armv7l")
            echo "armv7"
            ;;
        *)
            echo "$arch"
            ;;
    esac
}

# 标记已加载
export PLATFORM_LIB_LOADED=true

# 如果直接执行此脚本，显示平台信息
if [[ "${BASH_SOURCE[0]:-${0}}" == "${0}" ]]; then
    proj::platform::show_info
fi