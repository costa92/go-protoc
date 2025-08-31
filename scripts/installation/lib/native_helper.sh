#!/usr/bin/env bash

# =============================================================================
# 宿主机安装工具库
# Native Installation Helper Library
#
# 提供统一的宿主机服务安装接口，支持跨平台的服务安装和管理
# =============================================================================

set -eEuo pipefail

# 避免重复加载
[[ "${NATIVE_HELPER_LIB_LOADED:-false}" == "true" ]] && return 0

# 统一宿主机服务安装函数
proj::native::install_service() {
    local service_name="$1"
    local version="$2"
    local config_file="$3"
    local extra_args="${4:-}"
    
    # 验证参数
    if [[ -z "$service_name" ]] || [[ -z "$version" ]]; then
        proj::log::error "Service name and version are required"
        return 1
    fi
    
    # 检测平台并路由到对应的适配器
    local os_type
    os_type=$(proj::platform::detect_os)
    
    proj::log::info "Installing $service_name v$version on $os_type using native installation"
    
    case "$os_type" in
        "ubuntu")
            proj::ubuntu::install_service "$service_name" "$version" "$config_file" "$extra_args"
            ;;
        "macos")
            proj::macos::install_service "$service_name" "$version" "$config_file" "$extra_args"
            ;;
        *)
            proj::log::error "Native installation not supported on platform: $os_type"
            proj::log::info "Consider using Docker installation instead"
            return 1
            ;;
    esac
}

# 统一服务启动
proj::native::start_service() {
    local service_name="$1"
    
    if [[ -z "$service_name" ]]; then
        proj::log::error "Service name is required"
        return 1
    fi
    
    local os_type
    os_type=$(proj::platform::detect_os)
    
    proj::log::info "Starting service: $service_name on $os_type"
    
    case "$os_type" in
        "ubuntu")
            proj::ubuntu::start_service "$service_name"
            ;;
        "macos")
            proj::macos::start_service "$service_name"
            ;;
        *)
            proj::log::error "Service management not supported on platform: $os_type"
            return 1
            ;;
    esac
}

# 统一服务停止
proj::native::stop_service() {
    local service_name="$1"
    
    if [[ -z "$service_name" ]]; then
        proj::log::error "Service name is required"
        return 1
    fi
    
    local os_type
    os_type=$(proj::platform::detect_os)
    
    proj::log::info "Stopping service: $service_name on $os_type"
    
    case "$os_type" in
        "ubuntu")
            proj::ubuntu::stop_service "$service_name"
            ;;
        "macos")
            proj::macos::stop_service "$service_name"
            ;;
        *)
            proj::log::error "Service management not supported on platform: $os_type"
            return 1
            ;;
    esac
}

# 统一服务重启
proj::native::restart_service() {
    local service_name="$1"
    
    if [[ -z "$service_name" ]]; then
        proj::log::error "Service name is required"
        return 1
    fi
    
    proj::log::info "Restarting service: $service_name"
    
    # 停止服务
    if proj::native::stop_service "$service_name"; then
        # 等待一下确保服务完全停止
        sleep 2
        
        # 启动服务
        proj::native::start_service "$service_name"
    else
        proj::log::warn "Failed to stop service $service_name, attempting to start anyway"
        proj::native::start_service "$service_name"
    fi
}

# 统一服务状态检查
proj::native::service_status() {
    local service_name="$1"
    
    if [[ -z "$service_name" ]]; then
        proj::log::error "Service name is required"
        return 1
    fi
    
    local os_type
    os_type=$(proj::platform::detect_os)
    
    case "$os_type" in
        "ubuntu")
            proj::ubuntu::service_status "$service_name"
            ;;
        "macos")
            proj::macos::service_status "$service_name"
            ;;
        *)
            proj::log::error "Service status check not supported on platform: $os_type"
            return 1
            ;;
    esac
}

# 统一服务卸载
proj::native::uninstall_service() {
    local service_name="$1"
    local remove_data="${2:-false}"  # 是否删除数据目录
    
    if [[ -z "$service_name" ]]; then
        proj::log::error "Service name is required"
        return 1
    fi
    
    local os_type
    os_type=$(proj::platform::detect_os)
    
    proj::log::info "Uninstalling service: $service_name on $os_type"
    
    # 先停止服务
    proj::native::stop_service "$service_name" || true
    
    case "$os_type" in
        "ubuntu")
            proj::ubuntu::uninstall_service "$service_name" "$remove_data"
            ;;
        "macos")
            proj::macos::uninstall_service "$service_name" "$remove_data"
            ;;
        *)
            proj::log::error "Service uninstallation not supported on platform: $os_type"
            return 1
            ;;
    esac
}

# 创建系统用户和目录
proj::native::setup_service_user() {
    local service_name="$1"
    local custom_dirs="${2:-}"  # 可选的自定义目录列表
    
    if [[ -z "$service_name" ]]; then
        proj::log::error "Service name is required"
        return 1
    fi
    
    local os_type
    os_type=$(proj::platform::detect_os)
    
    proj::log::info "Setting up user and directories for service: $service_name"
    
    case "$os_type" in
        "ubuntu")
            proj::ubuntu::setup_user "$service_name" "$custom_dirs"
            ;;
        "macos")
            proj::macos::setup_user "$service_name" "$custom_dirs"
            ;;
        *)
            proj::log::error "User setup not supported on platform: $os_type"
            return 1
            ;;
    esac
}

# 下载并安装二进制文件（通用函数）
proj::native::download_and_install_binary() {
    local service_name="$1"
    local version="$2"
    local download_url="$3"
    local install_path="${4:-/usr/local/bin}"
    local binary_name="${5:-$service_name}"
    
    # 验证参数
    if [[ -z "$service_name" ]] || [[ -z "$version" ]] || [[ -z "$download_url" ]]; then
        proj::log::error "Service name, version, and download URL are required"
        return 1
    fi
    
    proj::log::info "Downloading $service_name v$version from $download_url"
    
    # 创建临时下载目录
    local temp_dir
    temp_dir=$(mktemp -d)
    local temp_file="$temp_dir/${service_name}-${version}"
    
    # 确保清理临时文件
    trap "rm -rf '$temp_dir'" EXIT
    
    # 检测系统架构
    local arch
    arch=$(proj::platform::get_architecture)
    
    # 替换下载URL中的变量
    download_url="${download_url//\{VERSION\}/$version}"
    download_url="${download_url//\{ARCH\}/$arch}"
    download_url="${download_url//\{OS\}/linux}"  # 可以根据平台调整
    
    # 下载文件
    if ! curl -L -f -o "$temp_file" "$download_url"; then
        proj::log::error "Failed to download $service_name from $download_url"
        return 1
    fi
    
    # 检测文件类型并解压（如果需要）
    local file_type
    file_type=$(file -b --mime-type "$temp_file")
    
    case "$file_type" in
        "application/gzip"|"application/x-gzip")
            proj::log::info "Extracting gzipped archive"
            if ! tar -xzf "$temp_file" -C "$temp_dir"; then
                proj::log::error "Failed to extract archive"
                return 1
            fi
            # 查找解压后的二进制文件
            local extracted_binary
            extracted_binary=$(find "$temp_dir" -type f -executable -name "*${binary_name}*" | head -1)
            if [[ -z "$extracted_binary" ]]; then
                proj::log::error "Could not find binary $binary_name in extracted files"
                return 1
            fi
            temp_file="$extracted_binary"
            ;;
        "application/x-executable"|"application/octet-stream")
            # 已经是二进制文件，直接使用
            ;;
        *)
            proj::log::warn "Unknown file type: $file_type, attempting to install anyway"
            ;;
    esac
    
    # 安装二进制文件
    proj::log::info "Installing $binary_name to $install_path"
    
    # 确保安装目录存在
    sudo mkdir -p "$install_path"
    
    # 复制二进制文件并设置权限
    sudo cp "$temp_file" "$install_path/$binary_name"
    sudo chmod +x "$install_path/$binary_name"
    
    # 验证安装
    if "$install_path/$binary_name" --version >/dev/null 2>&1 || 
       "$install_path/$binary_name" -version >/dev/null 2>&1 ||
       "$install_path/$binary_name" version >/dev/null 2>&1; then
        proj::log::success "Successfully installed $binary_name v$version"
        return 0
    else
        proj::log::warn "Binary installed but version check failed (this may be normal)"
        return 0
    fi
}

# 检查服务是否已安装
proj::native::is_service_installed() {
    local service_name="$1"
    local check_binary="${2:-true}"  # 是否检查二进制文件
    
    if [[ -z "$service_name" ]]; then
        proj::log::error "Service name is required"
        return 1
    fi
    
    local os_type
    os_type=$(proj::platform::detect_os)
    
    local is_installed=false
    
    # 检查服务是否已注册
    case "$os_type" in
        "ubuntu")
            if systemctl list-unit-files --type=service | grep -q "^${service_name}\.service"; then
                is_installed=true
            fi
            ;;
        "macos")
            if launchctl list | grep -q "com\.proj\.${service_name}"; then
                is_installed=true
            fi
            ;;
    esac
    
    # 检查二进制文件（如果需要）
    if [[ "$check_binary" == "true" ]]; then
        if command -v "$service_name" >/dev/null 2>&1; then
            is_installed=true
        fi
    fi
    
    if [[ "$is_installed" == "true" ]]; then
        proj::log::info "Service $service_name is installed"
        return 0
    else
        proj::log::info "Service $service_name is not installed"
        return 1
    fi
}

# 批量服务操作
proj::native::batch_operation() {
    local operation="$1"
    shift
    local services=("$@")
    
    if [[ -z "$operation" ]]; then
        proj::log::error "Operation is required (start|stop|restart|status)"
        return 1
    fi
    
    if [[ ${#services[@]} -eq 0 ]]; then
        proj::log::error "At least one service name is required"
        return 1
    fi
    
    proj::log::info "Performing batch $operation on ${#services[@]} services"
    
    local failed_services=()
    local successful_services=()
    
    for service in "${services[@]}"; do
        proj::log::info "Running $operation on service: $service"
        
        case "$operation" in
            "start")
                if proj::native::start_service "$service"; then
                    successful_services+=("$service")
                else
                    failed_services+=("$service")
                fi
                ;;
            "stop")
                if proj::native::stop_service "$service"; then
                    successful_services+=("$service")
                else
                    failed_services+=("$service")
                fi
                ;;
            "restart")
                if proj::native::restart_service "$service"; then
                    successful_services+=("$service")
                else
                    failed_services+=("$service")
                fi
                ;;
            "status")
                if proj::native::service_status "$service"; then
                    successful_services+=("$service")
                else
                    failed_services+=("$service")
                fi
                ;;
            *)
                proj::log::error "Unknown operation: $operation"
                return 1
                ;;
        esac
    done
    
    # 报告结果
    if [[ ${#successful_services[@]} -gt 0 ]]; then
        proj::log::success "Successfully processed services: ${successful_services[*]}"
    fi
    
    if [[ ${#failed_services[@]} -gt 0 ]]; then
        proj::log::error "Failed to process services: ${failed_services[*]}"
        return 1
    fi
    
    return 0
}

# 显示所有服务状态
proj::native::show_all_services() {
    local os_type
    os_type=$(proj::platform::detect_os)
    
    proj::log::info "Listing all project services on $os_type"
    
    # 预定义的项目服务列表
    local services=("redis" "otelcol" "prometheus" "victorialogs" "grafana" "jaeger")
    
    echo "Service Status Overview:"
    printf "%-15s %-10s %-20s\n" "Service" "Status" "Details"
    printf "%-15s %-10s %-20s\n" "-------" "------" "-------"
    
    for service in "${services[@]}"; do
        local status="Unknown"
        local details=""
        
        if proj::native::is_service_installed "$service" false; then
            if proj::native::service_status "$service" >/dev/null 2>&1; then
                status="Running"
                details="Active"
            else
                status="Stopped"
                details="Inactive"
            fi
        else
            status="Not Installed"
            details="N/A"
        fi
        
        printf "%-15s %-10s %-20s\n" "$service" "$status" "$details"
    done
}

# 标记已加载
export NATIVE_HELPER_LIB_LOADED=true

# 如果直接执行此脚本，显示服务状态概览
if [[ "${BASH_SOURCE[0]:-${0}}" == "${0}" ]]; then
    proj::native::show_all_services
fi