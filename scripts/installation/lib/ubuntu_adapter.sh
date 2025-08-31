#!/usr/bin/env bash

# =============================================================================
# Ubuntu平台适配器
# Ubuntu Platform Adapter
#
# 提供Ubuntu平台特定的服务安装、管理和配置功能
# 支持APT包管理、systemd服务管理和系统用户管理
# =============================================================================

set -eEuo pipefail

# 避免重复加载
[[ "${UBUNTU_ADAPTER_LIB_LOADED:-false}" == "true" ]] && return 0

# Ubuntu系统路径常量
readonly UBUNTU_SYSTEMD_PATH="/etc/systemd/system"
readonly UBUNTU_BIN_PATH="/usr/local/bin"
readonly UBUNTU_CONFIG_PATH="/etc"
readonly UBUNTU_DATA_PATH="/var/lib"
readonly UBUNTU_LOG_PATH="/var/log"

# Ubuntu包安装
proj::ubuntu::install_package() {
    local package_name="$1"
    local version="${2:-latest}"
    local update_cache="${3:-true}"
    
    if [[ -z "$package_name" ]]; then
        proj::log::error "Package name is required"
        return 1
    fi
    
    proj::log::info "Installing Ubuntu package: $package_name"
    
    # 更新包缓存（如果需要）
    if [[ "$update_cache" == "true" ]]; then
        proj::log::info "Updating APT package cache"
        sudo apt update || {
            proj::log::error "Failed to update APT cache"
            return 1
        }
    fi
    
    # 安装包
    if [[ "$version" == "latest" ]]; then
        sudo apt install -y "$package_name" || {
            proj::log::error "Failed to install package: $package_name"
            return 1
        }
    else
        # 尝试安装特定版本
        if ! sudo apt install -y "$package_name=$version"; then
            proj::log::warn "Failed to install specific version $version, trying without version"
            sudo apt install -y "$package_name" || {
                proj::log::error "Failed to install package: $package_name"
                return 1
            }
        fi
    fi
    
    proj::log::success "Successfully installed package: $package_name"
}

# Ubuntu服务安装
proj::ubuntu::install_service() {
    local service_name="$1"
    local version="$2"
    local config_file="$3"
    local extra_args="${4:-}"
    
    if [[ -z "$service_name" ]] || [[ -z "$version" ]]; then
        proj::log::error "Service name and version are required"
        return 1
    fi
    
    proj::log::info "Installing service $service_name v$version on Ubuntu"
    
    # 1. 创建服务用户和目录
    proj::ubuntu::setup_user "$service_name"
    
    # 2. 尝试通过APT安装（如果有包）
    local apt_package_name
    apt_package_name=$(proj::ubuntu::get_package_name "$service_name")
    
    if [[ -n "$apt_package_name" ]]; then
        proj::log::info "Installing via APT package: $apt_package_name"
        if proj::ubuntu::install_package "$apt_package_name" "$version"; then
            proj::log::success "Installed $service_name via APT"
        else
            proj::log::warn "APT installation failed, falling back to binary installation"
            proj::ubuntu::install_service_binary "$service_name" "$version"
        fi
    else
        proj::log::info "No APT package available, installing from binary"
        proj::ubuntu::install_service_binary "$service_name" "$version"
    fi
    
    # 3. 生成配置文件（如果提供）
    if [[ -n "$config_file" ]]; then
        proj::ubuntu::setup_service_config "$service_name" "$config_file"
    fi
    
    # 4. 创建systemd服务文件
    proj::ubuntu::create_systemd_service "$service_name" "$config_file"
    
    # 5. 启用并启动服务
    sudo systemctl daemon-reload
    sudo systemctl enable "$service_name"
    
    proj::log::success "Service $service_name installed successfully on Ubuntu"
}

# 通过二进制文件安装服务
proj::ubuntu::install_service_binary() {
    local service_name="$1"
    local version="$2"
    
    # 获取下载URL
    local download_url
    download_url=$(proj::ubuntu::get_binary_download_url "$service_name" "$version")
    
    if [[ -z "$download_url" ]]; then
        proj::log::error "No binary download URL available for $service_name"
        return 1
    fi
    
    proj::log::info "Installing $service_name from binary: $download_url"
    proj::native::download_and_install_binary "$service_name" "$version" "$download_url" "$UBUNTU_BIN_PATH"
}

# 获取包名映射
proj::ubuntu::get_package_name() {
    local service_name="$1"
    
    # 预定义的服务到Ubuntu包名的映射
    case "$service_name" in
        "redis")
            echo "redis-server"
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
            echo "mysql-server"
            ;;
        "postgresql")
            echo "postgresql"
            ;;
        *)
            # 对于没有预定义映射的服务，返回空（使用二进制安装）
            echo ""
            ;;
    esac
}

# 获取二进制下载URL
proj::ubuntu::get_binary_download_url() {
    local service_name="$1"
    local version="$2"
    
    # 预定义的下载URL模板
    case "$service_name" in
        "otelcol")
            echo "https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v${version}/otelcol-contrib_${version}_linux_{ARCH}.tar.gz"
            ;;
        "victorialogs")
            echo "https://github.com/VictoriaMetrics/VictoriaMetrics/releases/download/v${version}/victoria-logs-linux-{ARCH}-v${version}.tar.gz"
            ;;
        "victoriametrics")
            echo "https://github.com/VictoriaMetrics/VictoriaMetrics/releases/download/v${version}/victoria-metrics-linux-{ARCH}-v${version}.tar.gz"
            ;;
        "jaeger")
            echo "https://github.com/jaegertracing/jaeger/releases/download/v${version}/jaeger-${version}-linux-{ARCH}.tar.gz"
            ;;
        *)
            proj::log::warn "No binary download URL defined for $service_name"
            echo ""
            ;;
    esac
}

# systemd服务管理
proj::ubuntu::start_service() {
    local service_name="$1"
    
    if [[ -z "$service_name" ]]; then
        proj::log::error "Service name is required"
        return 1
    fi
    
    proj::log::info "Starting Ubuntu service: $service_name"
    
    if sudo systemctl start "$service_name"; then
        proj::log::success "Service $service_name started successfully"
        return 0
    else
        proj::log::error "Failed to start service: $service_name"
        # 显示服务状态以便调试
        sudo systemctl status "$service_name" --no-pager || true
        return 1
    fi
}

proj::ubuntu::stop_service() {
    local service_name="$1"
    
    if [[ -z "$service_name" ]]; then
        proj::log::error "Service name is required"
        return 1
    fi
    
    proj::log::info "Stopping Ubuntu service: $service_name"
    
    if sudo systemctl stop "$service_name"; then
        proj::log::success "Service $service_name stopped successfully"
        return 0
    else
        proj::log::error "Failed to stop service: $service_name"
        return 1
    fi
}

proj::ubuntu::service_status() {
    local service_name="$1"
    
    if [[ -z "$service_name" ]]; then
        proj::log::error "Service name is required"
        return 1
    fi
    
    # 检查服务是否存在
    if ! systemctl list-unit-files --type=service | grep -q "^${service_name}\.service"; then
        proj::log::warn "Service $service_name is not installed"
        return 1
    fi
    
    # 显示服务状态
    if systemctl is-active --quiet "$service_name"; then
        proj::log::success "Service $service_name is running"
        sudo systemctl status "$service_name" --no-pager --lines=5
        return 0
    else
        proj::log::warn "Service $service_name is not running"
        sudo systemctl status "$service_name" --no-pager --lines=10
        return 1
    fi
}

# 用户和目录管理
proj::ubuntu::setup_user() {
    local service_name="$1"
    local custom_dirs="${2:-}"
    
    if [[ -z "$service_name" ]]; then
        proj::log::error "Service name is required"
        return 1
    fi
    
    proj::log::info "Setting up user and directories for service: $service_name"
    
    # 创建系统用户（如果不存在）
    if ! id -u "$service_name" >/dev/null 2>&1; then
        proj::log::info "Creating system user: $service_name"
        sudo useradd --system --shell /bin/false --home-dir "/var/lib/$service_name" \
                     --create-home --comment "$service_name service user" "$service_name" || {
            proj::log::error "Failed to create user: $service_name"
            return 1
        }
    else
        proj::log::debug "User $service_name already exists"
    fi
    
    # 创建标准服务目录
    local service_dirs=(
        "$UBUNTU_DATA_PATH/$service_name"    # 数据目录
        "$UBUNTU_LOG_PATH/$service_name"     # 日志目录
        "$UBUNTU_CONFIG_PATH/$service_name"  # 配置目录
        "/run/$service_name"                 # 运行时目录（PID文件等）
    )
    
    # 添加自定义目录
    if [[ -n "$custom_dirs" ]]; then
        IFS=',' read -ra custom_array <<< "$custom_dirs"
        service_dirs+=("${custom_array[@]}")
    fi
    
    # 创建目录并设置权限
    for dir in "${service_dirs[@]}"; do
        proj::log::debug "Creating directory: $dir"
        sudo mkdir -p "$dir" || {
            proj::log::error "Failed to create directory: $dir"
            return 1
        }
        
        # 设置所有权
        sudo chown -R "$service_name:$service_name" "$dir" || {
            proj::log::warn "Failed to set ownership for directory: $dir"
        }
        
        # 设置适当的权限
        case "$dir" in
            *"/log/"*|*"/logs/"*)
                sudo chmod 755 "$dir"  # 日志目录
                ;;
            *"/config/"*)
                sudo chmod 750 "$dir"  # 配置目录（更严格）
                ;;
            *"/run/"*)
                sudo chmod 755 "$dir"  # 运行时目录
                ;;
            *)
                sudo chmod 755 "$dir"  # 默认权限
                ;;
        esac
    done
    
    proj::log::success "User and directories set up successfully for service: $service_name"
}

# 创建systemd服务文件
proj::ubuntu::create_systemd_service() {
    local service_name="$1"
    local config_file="$2"
    
    if [[ -z "$service_name" ]]; then
        proj::log::error "Service name is required"
        return 1
    fi
    
    local service_file="$UBUNTU_SYSTEMD_PATH/${service_name}.service"
    
    proj::log::info "Creating systemd service file: $service_file"
    
    # 获取二进制文件路径
    local binary_path
    if command -v "$service_name" >/dev/null 2>&1; then
        binary_path=$(command -v "$service_name")
    else
        binary_path="$UBUNTU_BIN_PATH/$service_name"
    fi
    
    # 构建启动命令
    local exec_start="$binary_path"
    if [[ -n "$config_file" ]]; then
        exec_start="$exec_start --config=$config_file"
    fi
    
    # 生成systemd服务文件
    sudo tee "$service_file" > /dev/null << EOF
[Unit]
Description=$service_name service
Documentation=https://github.com/costa92/go-protoc
After=network.target
Wants=network.target

[Service]
Type=simple
User=$service_name
Group=$service_name
ExecStart=$exec_start
ExecReload=/bin/kill -HUP \$MAINPID
KillMode=mixed
Restart=always
RestartSec=5
TimeoutStopSec=20

# 安全设置
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=/var/lib/$service_name /var/log/$service_name /run/$service_name

# 资源限制
LimitNOFILE=65536
LimitNPROC=32768

# 工作目录
WorkingDirectory=/var/lib/$service_name

# 环境变量
Environment=HOME=/var/lib/$service_name

[Install]
WantedBy=multi-user.target
EOF
    
    if [[ $? -eq 0 ]]; then
        proj::log::success "Created systemd service file: $service_file"
        return 0
    else
        proj::log::error "Failed to create systemd service file"
        return 1
    fi
}

# 设置服务配置
proj::ubuntu::setup_service_config() {
    local service_name="$1"
    local config_file="$2"
    
    if [[ -z "$service_name" ]] || [[ -z "$config_file" ]]; then
        proj::log::error "Service name and config file are required"
        return 1
    fi
    
    local config_dir="$UBUNTU_CONFIG_PATH/$service_name"
    local target_config="$config_dir/config.yaml"
    
    proj::log::info "Setting up configuration for service: $service_name"
    
    # 确保配置目录存在
    sudo mkdir -p "$config_dir"
    
    # 复制或链接配置文件
    if [[ -f "$config_file" ]]; then
        sudo cp "$config_file" "$target_config" || {
            proj::log::error "Failed to copy config file: $config_file"
            return 1
        }
        
        # 设置配置文件权限
        sudo chown "$service_name:$service_name" "$target_config"
        sudo chmod 640 "$target_config"
        
        proj::log::success "Configuration file set up: $target_config"
    else
        proj::log::warn "Config file not found: $config_file"
    fi
}

# 服务卸载
proj::ubuntu::uninstall_service() {
    local service_name="$1"
    local remove_data="${2:-false}"
    
    if [[ -z "$service_name" ]]; then
        proj::log::error "Service name is required"
        return 1
    fi
    
    proj::log::info "Uninstalling Ubuntu service: $service_name"
    
    # 停止并禁用服务
    sudo systemctl stop "$service_name" 2>/dev/null || true
    sudo systemctl disable "$service_name" 2>/dev/null || true
    
    # 删除systemd服务文件
    local service_file="$UBUNTU_SYSTEMD_PATH/${service_name}.service"
    if [[ -f "$service_file" ]]; then
        sudo rm -f "$service_file"
        sudo systemctl daemon-reload
    fi
    
    # 删除二进制文件
    if [[ -f "$UBUNTU_BIN_PATH/$service_name" ]]; then
        sudo rm -f "$UBUNTU_BIN_PATH/$service_name"
    fi
    
    # 删除配置文件
    if [[ -d "$UBUNTU_CONFIG_PATH/$service_name" ]]; then
        sudo rm -rf "$UBUNTU_CONFIG_PATH/$service_name"
    fi
    
    # 删除数据和日志（如果请求）
    if [[ "$remove_data" == "true" ]]; then
        proj::log::info "Removing data and log directories"
        sudo rm -rf "$UBUNTU_DATA_PATH/$service_name"
        sudo rm -rf "$UBUNTU_LOG_PATH/$service_name"
        sudo rm -rf "/run/$service_name"
    fi
    
    # 删除系统用户（可选）
    if id -u "$service_name" >/dev/null 2>&1; then
        if [[ "$remove_data" == "true" ]]; then
            sudo userdel "$service_name" 2>/dev/null || true
        else
            proj::log::info "Keeping user $service_name (use remove_data=true to delete)"
        fi
    fi
    
    proj::log::success "Service $service_name uninstalled successfully"
}

# 检查是否为Ubuntu系统
proj::ubuntu::validate_platform() {
    if [[ ! -f /etc/os-release ]]; then
        proj::log::error "Cannot detect OS release information"
        return 1
    fi
    
    local os_id
    os_id=$(grep '^ID=' /etc/os-release | cut -d'=' -f2 | tr -d '"')
    
    if [[ "$os_id" != "ubuntu" ]] && [[ "$os_id" != "debian" ]]; then
        proj::log::error "This adapter is designed for Ubuntu/Debian systems, detected: $os_id"
        return 1
    fi
    
    return 0
}

# 标记已加载
export UBUNTU_ADAPTER_LIB_LOADED=true

# 如果直接执行此脚本，验证平台兼容性
if [[ "${BASH_SOURCE[0]:-${0}}" == "${0}" ]]; then
    proj::ubuntu::validate_platform && echo "Ubuntu adapter is compatible with this system" || echo "Ubuntu adapter is not compatible with this system"
fi