#!/usr/bin/env bash

# =============================================================================
# macOS平台适配器
# macOS Platform Adapter
#
# 提供macOS平台特定的服务安装、管理和配置功能
# 支持Homebrew包管理、launchd服务管理和用户目录管理
# =============================================================================

set -eEuo pipefail

# 避免重复加载
[[ "${MACOS_ADAPTER_LIB_LOADED:-false}" == "true" ]] && return 0

# macOS系统路径常量
readonly MACOS_LAUNCHD_USER_PATH="$HOME/Library/LaunchAgents"
readonly MACOS_LAUNCHD_SYSTEM_PATH="/Library/LaunchDaemons"
readonly MACOS_BIN_PATH="/usr/local/bin"
readonly MACOS_CONFIG_PATH="$HOME/.config"
readonly MACOS_DATA_PATH="$HOME/.local/share"
readonly MACOS_LOG_PATH="$HOME/.local/var/log"
readonly MACOS_HOMEBREW_PREFIX="${HOMEBREW_PREFIX:-/opt/homebrew}"

# Homebrew包安装
proj::macos::install_package() {
    local package_name="$1"
    local version="${2:-latest}"
    local update_cache="${3:-false}"
    
    if [[ -z "$package_name" ]]; then
        proj::log::error "Package name is required"
        return 1
    fi
    
    # 确保Homebrew可用
    if ! proj::macos::ensure_homebrew; then
        return 1
    fi
    
    proj::log::info "Installing macOS package via Homebrew: $package_name"
    
    # 更新Homebrew（如果需要）
    if [[ "$update_cache" == "true" ]]; then
        proj::log::info "Updating Homebrew"
        brew update || {
            proj::log::warn "Failed to update Homebrew, continuing anyway"
        }
    fi
    
    # 安装包
    if [[ "$version" == "latest" ]]; then
        if brew install "$package_name"; then
            proj::log::success "Successfully installed package: $package_name"
            return 0
        else
            proj::log::error "Failed to install package: $package_name"
            return 1
        fi
    else
        # 尝试安装特定版本
        local versioned_formula="${package_name}@${version}"
        if brew install "$versioned_formula"; then
            proj::log::success "Successfully installed package: $versioned_formula"
            return 0
        else
            proj::log::warn "Failed to install specific version, trying latest"
            if brew install "$package_name"; then
                proj::log::success "Successfully installed package: $package_name (latest)"
                return 0
            else
                proj::log::error "Failed to install package: $package_name"
                return 1
            fi
        fi
    fi
}

# macOS服务安装
proj::macos::install_service() {
    local service_name="$1"
    local version="$2"
    local config_file="$3"
    local extra_args="${4:-}"
    
    if [[ -z "$service_name" ]] || [[ -z "$version" ]]; then
        proj::log::error "Service name and version are required"
        return 1
    fi
    
    proj::log::info "Installing service $service_name v$version on macOS"
    
    # 1. 创建用户目录
    proj::macos::setup_user "$service_name"
    
    # 2. 尝试通过Homebrew安装（如果有公式）
    local brew_formula
    brew_formula=$(proj::macos::get_brew_formula "$service_name")
    
    if [[ -n "$brew_formula" ]] && proj::macos::has_brew_formula "$brew_formula"; then
        proj::log::info "Installing via Homebrew formula: $brew_formula"
        if proj::macos::install_package "$brew_formula" "$version"; then
            proj::log::success "Installed $service_name via Homebrew"
        else
            proj::log::warn "Homebrew installation failed, falling back to binary installation"
            proj::macos::install_service_binary "$service_name" "$version"
        fi
    else
        proj::log::info "No Homebrew formula available, installing from binary"
        proj::macos::install_service_binary "$service_name" "$version"
    fi
    
    # 3. 生成配置文件（如果提供）
    if [[ -n "$config_file" ]]; then
        proj::macos::setup_service_config "$service_name" "$config_file"
    fi
    
    # 4. 创建launchd服务文件
    proj::macos::create_launchd_service "$service_name" "$config_file"
    
    # 5. 加载服务
    local plist_file="$MACOS_LAUNCHD_USER_PATH/com.proj.${service_name}.plist"
    if [[ -f "$plist_file" ]]; then
        launchctl load "$plist_file" || {
            proj::log::warn "Failed to load service immediately, will be loaded on next login"
        }
    fi
    
    proj::log::success "Service $service_name installed successfully on macOS"
}

# 通过二进制文件安装服务
proj::macos::install_service_binary() {
    local service_name="$1"
    local version="$2"
    
    # 获取下载URL
    local download_url
    download_url=$(proj::macos::get_binary_download_url "$service_name" "$version")
    
    if [[ -z "$download_url" ]]; then
        proj::log::error "No binary download URL available for $service_name"
        return 1
    fi
    
    proj::log::info "Installing $service_name from binary: $download_url"
    
    # 调整下载URL中的OS标识
    download_url="${download_url//linux/darwin}"
    
    proj::native::download_and_install_binary "$service_name" "$version" "$download_url" "$MACOS_BIN_PATH"
}

# 确保Homebrew可用
proj::macos::ensure_homebrew() {
    if command -v brew >/dev/null 2>&1; then
        return 0
    fi
    
    proj::log::warn "Homebrew not found, attempting to install"
    
    # 安装Homebrew
    if /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"; then
        # 更新PATH以包含Homebrew
        if [[ -f "$MACOS_HOMEBREW_PREFIX/bin/brew" ]]; then
            eval "$($MACOS_HOMEBREW_PREFIX/bin/brew shellenv)"
        elif [[ -f "/usr/local/bin/brew" ]]; then
            eval "$(/usr/local/bin/brew shellenv)"
        fi
        
        if command -v brew >/dev/null 2>&1; then
            proj::log::success "Homebrew installed successfully"
            return 0
        else
            proj::log::error "Homebrew installation completed but command not available"
            return 1
        fi
    else
        proj::log::error "Failed to install Homebrew"
        return 1
    fi
}

# 检查Homebrew公式是否存在
proj::macos::has_brew_formula() {
    local formula="$1"
    
    if [[ -z "$formula" ]]; then
        return 1
    fi
    
    # 搜索公式
    if brew search --formulae "^${formula}$" >/dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# 获取Homebrew公式名映射
proj::macos::get_brew_formula() {
    local service_name="$1"
    
    # 预定义的服务到Homebrew公式的映射
    case "$service_name" in
        "redis")
            echo "redis"
            ;;
        "prometheus")
            echo "prometheus"
            ;;
        "grafana")
            echo "grafana"
            ;;
        "nginx")
            echo "nginx"
            ;;
        "mysql")
            echo "mysql"
            ;;
        "postgresql")
            echo "postgresql"
            ;;
        "node")
            echo "node"
            ;;
        "python")
            echo "python"
            ;;
        *)
            # 对于没有预定义映射的服务，返回空（使用二进制安装）
            echo ""
            ;;
    esac
}

# 获取二进制下载URL
proj::macos::get_binary_download_url() {
    local service_name="$1"
    local version="$2"
    
    # 预定义的下载URL模板（macOS版本）
    case "$service_name" in
        "otelcol")
            echo "https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v${version}/otelcol-contrib_${version}_darwin_{ARCH}.tar.gz"
            ;;
        "victorialogs")
            echo "https://github.com/VictoriaMetrics/VictoriaMetrics/releases/download/v${version}/victoria-logs-darwin-{ARCH}-v${version}.tar.gz"
            ;;
        "victoriametrics")
            echo "https://github.com/VictoriaMetrics/VictoriaMetrics/releases/download/v${version}/victoria-metrics-darwin-{ARCH}-v${version}.tar.gz"
            ;;
        "jaeger")
            echo "https://github.com/jaegertracing/jaeger/releases/download/v${version}/jaeger-${version}-darwin-{ARCH}.tar.gz"
            ;;
        *)
            proj::log::warn "No binary download URL defined for $service_name"
            echo ""
            ;;
    esac
}

# launchd服务管理
proj::macos::start_service() {
    local service_name="$1"
    
    if [[ -z "$service_name" ]]; then
        proj::log::error "Service name is required"
        return 1
    fi
    
    local service_label="com.proj.${service_name}"
    
    proj::log::info "Starting macOS service: $service_name ($service_label)"
    
    if launchctl start "$service_label"; then
        proj::log::success "Service $service_name started successfully"
        return 0
    else
        proj::log::error "Failed to start service: $service_name"
        # 显示服务状态以便调试
        proj::macos::service_status "$service_name"
        return 1
    fi
}

proj::macos::stop_service() {
    local service_name="$1"
    
    if [[ -z "$service_name" ]]; then
        proj::log::error "Service name is required"
        return 1
    fi
    
    local service_label="com.proj.${service_name}"
    
    proj::log::info "Stopping macOS service: $service_name ($service_label)"
    
    if launchctl stop "$service_label"; then
        proj::log::success "Service $service_name stopped successfully"
        return 0
    else
        proj::log::error "Failed to stop service: $service_name"
        return 1
    fi
}

proj::macos::service_status() {
    local service_name="$1"
    
    if [[ -z "$service_name" ]]; then
        proj::log::error "Service name is required"
        return 1
    fi
    
    local service_label="com.proj.${service_name}"
    local plist_file="$MACOS_LAUNCHD_USER_PATH/com.proj.${service_name}.plist"
    
    # 检查plist文件是否存在
    if [[ ! -f "$plist_file" ]]; then
        proj::log::warn "Service $service_name is not installed (plist not found)"
        return 1
    fi
    
    # 检查服务是否加载
    if launchctl list | grep -q "$service_label"; then
        local pid status
        pid=$(launchctl list "$service_label" 2>/dev/null | grep PID | awk '{print $3}')
        status=$(launchctl list "$service_label" 2>/dev/null | grep LastExitStatus | awk '{print $3}')
        
        if [[ -n "$pid" ]] && [[ "$pid" != "-" ]]; then
            proj::log::success "Service $service_name is running (PID: $pid)"
            return 0
        else
            proj::log::warn "Service $service_name is loaded but not running (Exit status: $status)"
            return 1
        fi
    else
        proj::log::warn "Service $service_name is not loaded"
        return 1
    fi
}

# 用户目录管理（macOS适配）
proj::macos::setup_user() {
    local service_name="$1"
    local custom_dirs="${2:-}"
    
    if [[ -z "$service_name" ]]; then
        proj::log::error "Service name is required"
        return 1
    fi
    
    proj::log::info "Setting up directories for service: $service_name"
    
    # macOS使用当前用户的目录结构
    local service_dirs=(
        "$MACOS_CONFIG_PATH/$service_name"         # 配置目录
        "$MACOS_DATA_PATH/$service_name"           # 数据目录
        "$MACOS_LOG_PATH/$service_name"            # 日志目录
        "$MACOS_LAUNCHD_USER_PATH"                 # launchd plist目录
        "$HOME/.local/var/run/$service_name"       # 运行时目录
    )
    
    # 添加自定义目录
    if [[ -n "$custom_dirs" ]]; then
        IFS=',' read -ra custom_array <<< "$custom_dirs"
        service_dirs+=("${custom_array[@]}")
    fi
    
    # 创建目录并设置权限
    for dir in "${service_dirs[@]}"; do
        proj::log::debug "Creating directory: $dir"
        mkdir -p "$dir" || {
            proj::log::error "Failed to create directory: $dir"
            return 1
        }
        
        # macOS上设置适当的权限（用户拥有）
        chmod 755 "$dir" || {
            proj::log::warn "Failed to set permissions for directory: $dir"
        }
    done
    
    proj::log::success "Directories set up successfully for service: $service_name"
}

# 创建launchd服务文件
proj::macos::create_launchd_service() {
    local service_name="$1"
    local config_file="$2"
    
    if [[ -z "$service_name" ]]; then
        proj::log::error "Service name is required"
        return 1
    fi
    
    local plist_file="$MACOS_LAUNCHD_USER_PATH/com.proj.${service_name}.plist"
    
    proj::log::info "Creating launchd plist file: $plist_file"
    
    # 获取二进制文件路径
    local binary_path
    if command -v "$service_name" >/dev/null 2>&1; then
        binary_path=$(command -v "$service_name")
    else
        binary_path="$MACOS_BIN_PATH/$service_name"
    fi
    
    # 构建程序参数数组
    local program_args="<string>$binary_path</string>"
    if [[ -n "$config_file" ]]; then
        program_args="$program_args
        <string>--config=$config_file</string>"
    fi
    
    # 生成launchd plist文件
    cat > "$plist_file" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.proj.$service_name</string>
    
    <key>ProgramArguments</key>
    <array>
        $program_args
    </array>
    
    <key>WorkingDirectory</key>
    <string>$MACOS_DATA_PATH/$service_name</string>
    
    <key>StandardOutPath</key>
    <string>$MACOS_LOG_PATH/$service_name/$service_name.out.log</string>
    
    <key>StandardErrorPath</key>
    <string>$MACOS_LOG_PATH/$service_name/$service_name.err.log</string>
    
    <key>RunAtLoad</key>
    <true/>
    
    <key>KeepAlive</key>
    <dict>
        <key>SuccessfulExit</key>
        <false/>
    </dict>
    
    <key>ProcessType</key>
    <string>Background</string>
    
    <key>EnvironmentVariables</key>
    <dict>
        <key>HOME</key>
        <string>$HOME</string>
        <key>PATH</key>
        <string>/usr/local/bin:/usr/bin:/bin</string>
    </dict>
    
    <key>ThrottleInterval</key>
    <integer>5</integer>
    
    <key>ExitTimeOut</key>
    <integer>20</integer>
</dict>
</plist>
EOF
    
    if [[ $? -eq 0 ]]; then
        # 设置适当的权限
        chmod 644 "$plist_file"
        proj::log::success "Created launchd plist file: $plist_file"
        return 0
    else
        proj::log::error "Failed to create launchd plist file"
        return 1
    fi
}

# 设置服务配置
proj::macos::setup_service_config() {
    local service_name="$1"
    local config_file="$2"
    
    if [[ -z "$service_name" ]] || [[ -z "$config_file" ]]; then
        proj::log::error "Service name and config file are required"
        return 1
    fi
    
    local config_dir="$MACOS_CONFIG_PATH/$service_name"
    local target_config="$config_dir/config.yaml"
    
    proj::log::info "Setting up configuration for service: $service_name"
    
    # 确保配置目录存在
    mkdir -p "$config_dir"
    
    # 复制配置文件
    if [[ -f "$config_file" ]]; then
        cp "$config_file" "$target_config" || {
            proj::log::error "Failed to copy config file: $config_file"
            return 1
        }
        
        # 设置配置文件权限
        chmod 640 "$target_config"
        
        proj::log::success "Configuration file set up: $target_config"
    else
        proj::log::warn "Config file not found: $config_file"
    fi
}

# 服务卸载
proj::macos::uninstall_service() {
    local service_name="$1"
    local remove_data="${2:-false}"
    
    if [[ -z "$service_name" ]]; then
        proj::log::error "Service name is required"
        return 1
    fi
    
    proj::log::info "Uninstalling macOS service: $service_name"
    
    local service_label="com.proj.${service_name}"
    local plist_file="$MACOS_LAUNCHD_USER_PATH/com.proj.${service_name}.plist"
    
    # 停止并卸载服务
    launchctl stop "$service_label" 2>/dev/null || true
    launchctl unload "$plist_file" 2>/dev/null || true
    
    # 删除plist文件
    if [[ -f "$plist_file" ]]; then
        rm -f "$plist_file"
    fi
    
    # 删除二进制文件（如果是手动安装的）
    if [[ -f "$MACOS_BIN_PATH/$service_name" ]]; then
        rm -f "$MACOS_BIN_PATH/$service_name"
    fi
    
    # 删除配置文件
    if [[ -d "$MACOS_CONFIG_PATH/$service_name" ]]; then
        rm -rf "$MACOS_CONFIG_PATH/$service_name"
    fi
    
    # 删除数据和日志（如果请求）
    if [[ "$remove_data" == "true" ]]; then
        proj::log::info "Removing data and log directories"
        rm -rf "$MACOS_DATA_PATH/$service_name"
        rm -rf "$MACOS_LOG_PATH/$service_name"
        rm -rf "$HOME/.local/var/run/$service_name"
    fi
    
    proj::log::success "Service $service_name uninstalled successfully"
}

# 检查是否为macOS系统
proj::macos::validate_platform() {
    if [[ "$OSTYPE" != "darwin"* ]]; then
        proj::log::error "This adapter is designed for macOS systems, detected: $OSTYPE"
        return 1
    fi
    
    return 0
}

# 显示macOS特定信息
proj::macos::show_info() {
    local macos_version
    macos_version=$(sw_vers -productVersion 2>/dev/null || echo "Unknown")
    
    local homebrew_status="Not installed"
    if command -v brew >/dev/null 2>&1; then
        homebrew_status="Installed ($(brew --version | head -1))"
    fi
    
    cat << EOF
macOS Platform Information:
  macOS Version: $macos_version
  Homebrew Status: $homebrew_status
  Architecture: $(proj::platform::get_architecture)
  
Directory Structure:
  Config Path: $MACOS_CONFIG_PATH
  Data Path: $MACOS_DATA_PATH
  Log Path: $MACOS_LOG_PATH
  LaunchAgent Path: $MACOS_LAUNCHD_USER_PATH
  Binary Path: $MACOS_BIN_PATH
EOF
}

# 标记已加载
export MACOS_ADAPTER_LIB_LOADED=true

# 如果直接执行此脚本，显示macOS信息
if [[ "${BASH_SOURCE[0]:-${0}}" == "${0}" ]]; then
    proj::macos::validate_platform && proj::macos::show_info || echo "macOS adapter is not compatible with this system"
fi