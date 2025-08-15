#!/usr/bin/env bash
#
# MySQL Installation Script
# This script provides functions to install, configure, and manage MySQL server
# both natively and via Docker containers.
#

# Exit on any error, undefined variables, or pipe failures
set -o errexit
set -o nounset
set -o pipefail

# 加载通用配置和版本管理
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Environment variables for MySQL configuration
# Can be overridden by setting these variables before running the script
PROJ_MYSQL_HOST=${PROJ_MYSQL_HOST:-127.0.0.1}        # MySQL server host
PROJ_MYSQL_PORT=${PROJ_MYSQL_PORT:-3306}             # MySQL server port
PROJ_MYSQL_ROOT_PASSWORD=${PROJ_MYSQL_ROOT_PASSWORD:-proj\(\#\)666}  # MySQL root password
PROJ_MYSQL_USER=${PROJ_MYSQL_USER:-appuser}          # Application user
PROJ_MYSQL_PASSWORD=${PROJ_MYSQL_PASSWORD:-proj\(\#\)666}  # Application user password
# 版本信息从统一配置文件加载：MYSQL_VERSION 在 versions.sh 中定义
MYSQL_DOCKER_MNAME=${NETWORK_NAME}-mysql

# Function to install MySQL natively
proj::mysql::install() {
  proj::mysql::pre_install

  if proj::util::is_mac; then
    # macOS installation using Homebrew
    if ! command -v brew &> /dev/null; then
      proj::log::error "Homebrew is required for MySQL installation on macOS"
      return 1
    fi

    proj::log::info "Installing MySQL via Homebrew..."
    brew install mysql

    # Start MySQL service
    brew services start mysql

    # Secure installation (optional password setup)
    proj::log::info "Setting up MySQL root password..."
    mysql -u root -e "ALTER USER 'root'@'localhost' IDENTIFIED BY '${PROJ_MYSQL_ROOT_PASSWORD}';" 2>/dev/null || true

  else
    # Linux installation using apt
    proj::log::info "Installing MySQL on Linux..."

    # Update package index
    proj::util::sudo "apt update"

    # Install MySQL server
    export DEBIAN_FRONTEND=noninteractive
    proj::util::sudo "apt install -y mysql-server"

    # Start and enable MySQL service
    proj::util::sudo "systemctl start mysql"
    proj::util::sudo "systemctl enable mysql"

    # Configure MySQL root password
    proj::log::info "Configuring MySQL root password..."
    proj::util::sudo "mysql -e \"ALTER USER 'root'@'localhost' IDENTIFIED WITH mysql_native_password BY '${PROJ_MYSQL_ROOT_PASSWORD}';\""
    proj::util::sudo "mysql -e \"FLUSH PRIVILEGES;\""
  fi

  # Create application database and user
  proj::mysql::create_database

  proj::mysql::status || return 1
  proj::mysql::info
  proj::log::info "install MySQL successfully"
}

# Uninstall MySQL step by step
proj::mysql::uninstall() {
  set +o errexit

  if proj::util::is_mac; then
    # macOS uninstallation
    brew services stop mysql 2>/dev/null || true
    brew uninstall mysql 2>/dev/null || true
    proj::util::sudo "rm -rf /usr/local/var/mysql" 2>/dev/null || true
  else
    # Linux uninstallation
    proj::util::sudo "systemctl stop mysql" 2>/dev/null || true
    proj::util::sudo "systemctl disable mysql" 2>/dev/null || true
    proj::util::sudo "apt remove -y mysql-server mysql-client mysql-common" 2>/dev/null || true
    proj::util::sudo "apt autoremove -y" 2>/dev/null || true
    proj::util::sudo "rm -rf /var/lib/mysql" 2>/dev/null || true
    proj::util::sudo "rm -rf /etc/mysql" 2>/dev/null || true
  fi

  set -o errexit
  proj::log::info "uninstall MySQL successfully"
  return 0
}

# Pre-install MySQL by system init and package management
proj::mysql::pre_install() {
    proj::log::info "Pre-installing MySQL..."

    if proj::util::is_mac; then
        proj::log::info "Mac OS detected, checking for Homebrew..."
        if ! command -v brew &> /dev/null; then
            proj::log::error "Homebrew is required for MySQL installation on macOS"
            proj::log::info "Please install Homebrew first: /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
            return 1
        fi

        # Install MySQL client if not present
        if ! command -v mysql &> /dev/null; then
            proj::log::info "Installing MySQL client..."
            brew install mysql-client
        else
            proj::log::info "MySQL client already installed, skipping..."
        fi
    else
        # Linux pre-installation
        if ! proj::util::cmd_exists "mysql"; then
            proj::log::info "Installing MySQL client..."
            proj::util::sudo "apt update"
            proj::util::sudo "apt install -y mysql-client"
        else
            proj::log::info "MySQL client already installed, skipping..."
        fi
    fi
}

# Install MySQL using a Docker container
proj::mysql::docker::install() {
    proj::log::info "Installing Docker MySQL..."

    proj::mysql::pre_install
    proj::common::network

    # Stop any existing container
    docker rm -f ${MYSQL_DOCKER_MNAME} 2>/dev/null || true

    # Run MySQL container
    docker run -d --name ${MYSQL_DOCKER_MNAME} \
      --restart unless-stopped \
      --network ${NETWORK_NAME} \
      -v ${PROJ_THIRDPARTY_INSTALL_DIR}/mysql:/var/lib/mysql \
      -p ${PROJ_MYSQL_PORT}:3306 \
      -e MYSQL_ROOT_PASSWORD=${PROJ_MYSQL_ROOT_PASSWORD} \
      -e MYSQL_USER=${PROJ_MYSQL_USER} \
      -e MYSQL_PASSWORD=${PROJ_MYSQL_PASSWORD} \
      mysql:${MYSQL_VERSION} \
      --character-set-server=utf8mb4 \
      --collation-server=utf8mb4_unicode_ci \
      --default-authentication-plugin=mysql_native_password

    # Wait for MySQL to be ready
    proj::log::info "Waiting for MySQL to be ready..."
    sleep 30

    # Check if MySQL is ready
    local max_attempts=30
    local attempt=1
    while [ $attempt -le $max_attempts ]; do
        if mysql -h"${PROJ_MYSQL_HOST}" -P"${PROJ_MYSQL_PORT}" -uroot -p"${PROJ_MYSQL_ROOT_PASSWORD}" -e "SELECT 1" &>/dev/null; then
            proj::log::info "MySQL is ready!"
            break
        fi
        proj::log::info "Waiting for MySQL... (${attempt}/${max_attempts})"
        sleep 2
        ((attempt++))
    done

    if [ $attempt -gt $max_attempts ]; then
        proj::log::error "MySQL failed to start within expected time"
        return 1
    fi

    proj::mysql::status || return 1
    proj::mysql::info
    proj::log::info "install MySQL successfully"
}

# Print necessary information after docker or native installation
proj::mysql::info() {
  echo -e ${C_GREEN}MySQL has been installed, here are some useful information:${C_NORMAL}
  cat << EOF | sed 's/^/  /'
MySQL access endpoint is: ${PROJ_MYSQL_HOST}:${PROJ_MYSQL_PORT}
      MySQL root password: ${PROJ_MYSQL_ROOT_PASSWORD}
            Application user: ${PROJ_MYSQL_USER}
        Application password: ${PROJ_MYSQL_PASSWORD}
      MySQL Login Command: mysql -h ${PROJ_MYSQL_HOST} -P ${PROJ_MYSQL_PORT} -u root -p'${PROJ_MYSQL_ROOT_PASSWORD}'
       App Login Command: mysql -h ${PROJ_MYSQL_HOST} -P ${PROJ_MYSQL_PORT} -u ${PROJ_MYSQL_USER} -p'${PROJ_MYSQL_PASSWORD}'
EOF
}

# Uninstall the docker container
proj::mysql::docker::uninstall() {
  docker rm -f ${MYSQL_DOCKER_MNAME} &>/dev/null || true
  proj::util::sudo "rm -rf ${PROJ_THIRDPARTY_INSTALL_DIR}/mysql" 2>/dev/null || true
  proj::log::info "uninstall MySQL successfully"
}

# Status check after docker or native installation
proj::mysql::status() {
  proj::util::telnet ${PROJ_MYSQL_HOST} ${PROJ_MYSQL_PORT} || return 1
  mysql -h"${PROJ_MYSQL_HOST}" -P"${PROJ_MYSQL_PORT}" -uroot -p"${PROJ_MYSQL_ROOT_PASSWORD}" -e "SELECT VERSION();" || {
    proj::log::error "can not login with root user, MySQL maybe not initialized properly."
    return 1
  }
}

# This block allows calling functions in this script directly from the command line.
# It checks if the script's arguments contain a function name with the 'proj::mysql::' prefix
# and, if so, executes that function.
# For example: ./mysql.sh proj::mysql::install
if [[ "$*" =~ proj::mysql:: ]]; then
  eval $*
fi