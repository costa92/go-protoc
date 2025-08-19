#!/usr/bin/env bash
#
# Sentry Installation Script  
# This script provides functions to install, configure, and manage Sentry
# both natively and via Docker containers.
#

# Exit on any error, undefined variables, or pipe failures
set -o errexit
set -o nounset
set -o pipefail

# 加载通用配置和版本管理
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Environment variables for Sentry configuration
# Can be overridden by setting these variables before running the script
PROJ_SENTRY_HOST=${PROJ_SENTRY_HOST:-127.0.0.1}                    # Sentry server host
PROJ_SENTRY_PORT=${PROJ_SENTRY_PORT:-9000}                         # Sentry server port  
PROJ_SENTRY_WEB_PORT=${PROJ_SENTRY_WEB_PORT:-9000}                 # Sentry web UI port
PROJ_SENTRY_POSTGRES_DB=${PROJ_SENTRY_POSTGRES_DB:-sentry}         # PostgreSQL database name
PROJ_SENTRY_POSTGRES_USER=${PROJ_SENTRY_POSTGRES_USER:-sentry}     # PostgreSQL user
PROJ_SENTRY_POSTGRES_PASSWORD=${PROJ_SENTRY_POSTGRES_PASSWORD:-sentry123} # PostgreSQL password
PROJ_SENTRY_REDIS_URL=${PROJ_SENTRY_REDIS_URL:-redis://redis:6379/0} # Redis URL for Sentry
PROJ_SENTRY_SECRET_KEY=${PROJ_SENTRY_SECRET_KEY:-$(openssl rand -hex 32)} # Sentry secret key

# 版本信息从统一配置文件加载：SENTRY_VERSION 在 versions.sh 中定义
SENTRY_DOCKER_MNAME=${NETWORK_NAME}-sentry
SENTRY_POSTGRES_DOCKER_MNAME=${NETWORK_NAME}-sentry-postgres
SENTRY_REDIS_DOCKER_MNAME=${NETWORK_NAME}-sentry-redis

# Function to install Sentry using Docker Compose (recommended approach)
proj::sentry::install() {
    proj::sentry::pre_install

    # 创建 Sentry 数据目录
    proj::util::sudo "mkdir -p ${PROJ_THIRDPARTY_INSTALL_DIR}/sentry/data"
    proj::util::sudo "mkdir -p ${PROJ_THIRDPARTY_INSTALL_DIR}/sentry/postgres"

    # 安装 Sentry 依赖 (Python, PostgreSQL client, Redis client)
    if proj::util::is_linux; then
        proj::log::info "Installing Sentry dependencies..."
        
        # 安装 PostgreSQL 客户端工具
        if ! proj::util::cmd_exists "psql"; then
            proj::log::info "Installing PostgreSQL client..."
            if ! proj::util::sudo "apt install -y postgresql-client" 2>/dev/null; then
                proj::log::warn "PostgreSQL client installation failed, but continuing..."
            fi
        fi
        
        # 安装 Redis 客户端工具 (如果尚未安装)
        if ! proj::util::cmd_exists "redis-cli"; then
            proj::log::info "Installing Redis client tools..."
            if ! proj::util::sudo "apt install -y redis-tools" 2>/dev/null; then
                proj::log::warn "Redis tools installation failed, but continuing..."
            fi
        fi
        
        # 安装 Python 依赖
        if ! proj::util::cmd_exists "python3"; then
            proj::log::info "Installing Python dependencies..."
            if ! proj::util::sudo "apt install -y python3 python3-pip python3-venv" 2>/dev/null; then
                proj::log::warn "Python installation failed, but continuing..."
            fi
        fi
    fi

    proj::log::info "Native installation of Sentry is complex and requires extensive setup."
    proj::log::info "For production use, please use Docker installation: make deploy.install.docker.sentry"
    proj::log::info "This native install creates a minimal setup for testing purposes only."
    
    # 创建 Sentry 配置文件
    local sentry_config="${PROJ_THIRDPARTY_INSTALL_DIR}/sentry/sentry.conf.py"
    local template_config_file="${SCRIPT_DIR}/sentry/sentry.conf.py"
    local temp_config_file="/tmp/sentry.conf.py.tmp"

    proj::util::sudo "mkdir -p $(dirname ${sentry_config})"

    # 使用 envsubst 替换模板中的环境变量
    envsubst < ${template_config_file} > ${temp_config_file}

    # 复制配置文件到系统目录
    proj::util::sudo "cp ${temp_config_file} ${sentry_config}"
    proj::util::sudo "chown root:root ${sentry_config}"
    rm -f ${temp_config_file}

    proj::sentry::status || return 1
    proj::sentry::info
    proj::log::info "install sentry configuration successfully"
}

# Uninstall Sentry native installation
proj::sentry::uninstall() {
    proj::log::info "Uninstalling native Sentry installation..."
    
    # 停止相关服务进程
    set +o errexit
    sentry_pids=$(pgrep -f "sentry" || true)
    if [[ -n "${sentry_pids}" ]]; then
        proj::util::sudo "kill -9 ${sentry_pids}"
    fi
    set -o errexit
    
    # 清理安装目录
    proj::util::sudo "rm -rf ${PROJ_THIRDPARTY_INSTALL_DIR}/sentry"
    
    proj::log::info "uninstall sentry successfully"
    return 0
}

# Pre-install Sentry by system init and package management
proj::sentry::pre_install() {
    proj::log::info "Pre-installing Sentry..."
    
    # 检查 envsubst (gettext-base)
    if ! command -v envsubst >/dev/null 2>&1; then
        proj::log::info "Installing gettext-base for envsubst..."
        if ! proj::util::sudo "apt install -y gettext-base" 2>/dev/null; then
            proj::log::warn "Failed to install gettext-base, but envsubst might be available from other sources"
            # 检查是否 envsubst 现在可用
            if ! command -v envsubst >/dev/null 2>&1; then
                proj::log::error "envsubst is required but not available. Please install gettext-base manually."
                return 1
            fi
        fi
    fi
    
    # 判断是 mac 还是 linux
    if proj::util::is_mac; then
        proj::log::info "Mac OS detected, installing dependencies via brew..."
        if ! proj::util::cmd_exists "brew"; then
            proj::log::error "Homebrew not found. Please install Homebrew first."
            return 1
        fi
        
        # 安装必要依赖
        brew install postgresql redis python3 || true
    else
        proj::log::info "Linux detected, installing dependencies via apt..."
        
        # 更新包列表，忽略失败的仓库
        proj::log::info "Updating package lists (ignoring repository errors)..."
        
        # 首先尝试标准更新
        if ! proj::util::sudo "apt update --allow-releaseinfo-change" 2>/dev/null; then
            proj::log::warn "Some repositories are unreachable, trying alternative update methods..."
            
            # 尝试忽略失败的仓库
            proj::util::sudo "apt update --allow-releaseinfo-change -o APT::Update::Error-Mode=any" 2>/dev/null || {
                proj::log::warn "Repository update partially failed, but continuing with available packages..."
                
                # 最后尝试：只更新可用的仓库
                proj::util::sudo "apt-get update --error-on=any" 2>/dev/null || {
                    proj::log::info "Using existing package cache..."
                }
            }
        fi
        
        # 检查并安装 Docker
        if ! proj::util::cmd_exists "docker"; then
            proj::log::info "Docker not found, attempting to install Docker..."
            if proj::util::sudo "apt install -y docker.io" 2>/dev/null; then
                proj::util::sudo "systemctl start docker" 2>/dev/null || true
                proj::util::sudo "systemctl enable docker" 2>/dev/null || true
                proj::log::info "Docker installation completed"
            else
                proj::log::warn "Docker installation failed, but continuing..."
            fi
        fi
        
        # 检查并安装 Docker Compose  
        if ! proj::util::cmd_exists "docker-compose"; then
            proj::log::info "Docker Compose not found, attempting to install..."
            if ! proj::util::sudo "apt install -y docker-compose" 2>/dev/null; then
                proj::log::warn "Docker Compose installation failed, but continuing..."
            fi
        fi
    fi
}

# Install Sentry using Docker containers (recommended approach)
proj::sentry::docker::install() {
    proj::log::info "Installing docker Sentry..."
    
    proj::sentry::pre_install
    proj::common::network

    # 创建持久化数据目录
    mkdir -p "${PROJ_THIRDPARTY_INSTALL_DIR}/sentry"
    mkdir -p "${PROJ_THIRDPARTY_INSTALL_DIR}/sentry/postgres"
    mkdir -p "${PROJ_THIRDPARTY_INSTALL_DIR}/sentry/data"
    
    # 安装 PostgreSQL 数据库容器 (Sentry 依赖)
    proj::log::info "Installing PostgreSQL for Sentry..."
    docker run -d --name ${SENTRY_POSTGRES_DOCKER_MNAME} \
        --restart always \
        --network ${NETWORK_NAME} \
        -v "${PROJ_THIRDPARTY_INSTALL_DIR}/sentry/postgres:/var/lib/postgresql/data" \
        -e POSTGRES_DB=${PROJ_SENTRY_POSTGRES_DB} \
        -e POSTGRES_USER=${PROJ_SENTRY_POSTGRES_USER} \
        -e POSTGRES_PASSWORD=${PROJ_SENTRY_POSTGRES_PASSWORD} \
        postgres:15-alpine

    # 等待 PostgreSQL 启动
    sleep 10
    
    # 安装 Redis 容器 (如果不存在)
    if ! docker ps --format "table {{.Names}}" | grep -q "${NETWORK_NAME}-redis"; then
        proj::log::info "Installing Redis for Sentry..."
        docker run -d --name ${SENTRY_REDIS_DOCKER_MNAME} \
            --restart always \
            --network ${NETWORK_NAME} \
            redis:7-alpine redis-server --appendonly yes
        sleep 5
    fi

    # 安装 Sentry 容器
    proj::log::info "Installing Sentry application..."
    
    # 清理可能存在的同名容器
    proj::common::docker::cleanup_container "${SENTRY_DOCKER_MNAME}"

    docker run -d --name ${SENTRY_DOCKER_MNAME} \
        --restart always \
        --network ${NETWORK_NAME} \
        -p ${PROJ_SENTRY_HOST}:${PROJ_SENTRY_WEB_PORT}:9000 \
        -v "${PROJ_THIRDPARTY_INSTALL_DIR}/sentry/data:/data" \
        -e SENTRY_SECRET_KEY="${PROJ_SENTRY_SECRET_KEY}" \
        -e SENTRY_POSTGRES_HOST="${SENTRY_POSTGRES_DOCKER_MNAME}" \
        -e SENTRY_POSTGRES_PORT=5432 \
        -e SENTRY_DB_NAME="${PROJ_SENTRY_POSTGRES_DB}" \
        -e SENTRY_DB_USER="${PROJ_SENTRY_POSTGRES_USER}" \
        -e SENTRY_DB_PASSWORD="${PROJ_SENTRY_POSTGRES_PASSWORD}" \
        -e SENTRY_REDIS_HOST="${SENTRY_REDIS_DOCKER_MNAME}" \
        -e SENTRY_REDIS_PORT=6379 \
        sentry:${SENTRY_VERSION} \
        run web

    # 等待 Sentry 启动
    sleep 15
    
    # 初始化 Sentry 数据库
    proj::log::info "Initializing Sentry database..."
    docker exec ${SENTRY_DOCKER_MNAME} sentry upgrade --noinput || true
    
    # 创建超级用户 (可选)
    proj::log::info "Creating Sentry superuser..."
    docker exec -i ${SENTRY_DOCKER_MNAME} sentry createuser --email admin@example.com --password admin123 --superuser --no-input || true

    sleep 5
    if proj::util::is_linux; then
        proj::sentry::status || return 1
    fi
    proj::sentry::info
    proj::log::info "install sentry successfully"
}

# Print necessary information after docker or native installation
proj::sentry::info() {
    echo -e ${C_GREEN}Sentry has been installed, here are some useful information:${C_NORMAL}
    cat << EOF | sed 's/^/  /'
Sentry Web UI endpoint: http://${PROJ_SENTRY_HOST}:${PROJ_SENTRY_WEB_PORT}
       Sentry Admin User: admin@example.com
   Sentry Admin Password: admin123
          Database Host: ${PROJ_SENTRY_HOST}
          Database Name: ${PROJ_SENTRY_POSTGRES_DB}
          Database User: ${PROJ_SENTRY_POSTGRES_USER}
      Database Password: ${PROJ_SENTRY_POSTGRES_PASSWORD}
            Secret Key: ${PROJ_SENTRY_SECRET_KEY}

Configuration Notes:
- Use the Web UI to create your first project and get DSN
- Configure your application to send errors to the DSN
- Access admin panel to manage projects and users
EOF
}

# Uninstall the docker containers
proj::sentry::docker::uninstall() {
    proj::log::info "Uninstalling Docker Sentry..."
    
    # 停止并删除 Sentry 容器
    docker rm -f ${SENTRY_DOCKER_MNAME} &>/dev/null || true
    
    # 停止并删除 PostgreSQL 容器
    docker rm -f ${SENTRY_POSTGRES_DOCKER_MNAME} &>/dev/null || true
    
    # 停止并删除专用 Redis 容器 (如果存在)
    docker rm -f ${SENTRY_REDIS_DOCKER_MNAME} &>/dev/null || true
    
    # 清理数据目录 (谨慎操作)
    proj::util::sudo "rm -rf ${PROJ_THIRDPARTY_INSTALL_DIR}/sentry"
    
    proj::log::info "uninstall sentry successfully"
}

# Status check after docker or native installation  
proj::sentry::status() {
    proj::log::info "Checking Sentry status..."
    
    # 检查端口是否开放
    proj::util::telnet ${PROJ_SENTRY_HOST} ${PROJ_SENTRY_WEB_PORT} || {
        proj::log::error "Sentry web interface is not accessible on ${PROJ_SENTRY_HOST}:${PROJ_SENTRY_WEB_PORT}"
        return 1
    }
    
    # 检查 HTTP 响应 (如果有 curl)
    if proj::util::cmd_exists "curl"; then
        local http_status
        http_status=$(curl -s -o /dev/null -w "%{http_code}" "http://${PROJ_SENTRY_HOST}:${PROJ_SENTRY_WEB_PORT}/" || echo "000")
        if [[ "${http_status}" =~ ^[23] ]]; then
            proj::log::info "Sentry web interface is responding correctly (HTTP ${http_status})"
        else
            proj::log::warn "Sentry web interface returned HTTP ${http_status}, may still be initializing"
        fi
    fi
    
    proj::log::info "Sentry status check completed successfully"
    return 0
}

# This block allows calling functions in this script directly from the command line.
# It checks if the script's arguments contain a function name with the 'proj::sentry::' prefix
# and, if so, executes that function.
# For example: ./sentry.sh proj::sentry::install
if [[ "$*" =~ proj::sentry:: ]]; then
    eval $*
fi