#!/usr/bin/env bash
#
# MySQL Database Management Script
# This script provides functions to manage MySQL database operations
# including setup, migration, backup, and maintenance tasks.
#

# Exit on any error, undefined variables, or pipe failures
set -o errexit
set -o nounset
set -o pipefail


# 加载通用配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
source "${PROJECT_ROOT}/scripts/common.sh"
# Environment variables for MySQL configuration
# Can be overridden by setting these variables before running the script
PROJ_MYSQL_HOST=${PROJ_MYSQL_HOST:-127.0.0.1}
PROJ_MYSQL_PORT=${PROJ_MYSQL_PORT:-3306}
PROJ_MYSQL_DATABASE=${PROJ_MYSQL_DATABASE:-protoc}
PROJ_MYSQL_USERNAME=${PROJ_MYSQL_USERNAME:-root}
PROJ_MYSQL_PASSWORD=${PROJ_MYSQL_PASSWORD:-proj(#)666}
PROJ_MYSQL_ADMIN_USERNAME=${PROJ_MYSQL_ADMIN_USERNAME:-root}
PROJ_MYSQL_ADMIN_PASSWORD=${PROJ_MYSQL_ADMIN_PASSWORD:-proj(#)666}

# Docker compose file path
MYSQL_COMPOSE_DIR="${PROJECT_ROOT}/deployments/mysql"
MIGRATIONS_DIR="${MYSQL_COMPOSE_DIR}/migrations"

# Function to setup complete database (start + migrate)
proj::mysql::setup() {
  proj::log::info "Setting up database..."
  proj::mysql::start
  proj::log::info "Waiting for MySQL to be ready..."
  sleep 30
  proj::mysql::migrate
  proj::log::info "Database setup completed!"
}

# Function to start MySQL database using Docker Compose
proj::mysql::start() {
  proj::log::info "Starting MySQL database..."
  proj::log::info "Using database: ${PROJ_MYSQL_DATABASE}"
  proj::log::info "Host: ${PROJ_MYSQL_HOST}:${PROJ_MYSQL_PORT}"

  cd "${MYSQL_COMPOSE_DIR}" && \
    export PROJ_MYSQL_ADMIN_PASSWORD="${PROJ_MYSQL_ADMIN_PASSWORD}" && \
    export PROJ_MYSQL_DATABASE="${PROJ_MYSQL_DATABASE}" && \
    export PROJ_MYSQL_USERNAME="${PROJ_MYSQL_USERNAME}" && \
    export PROJ_MYSQL_PASSWORD="${PROJ_MYSQL_PASSWORD}" && \
    docker-compose up -d

  proj::log::info "MySQL is starting... Please wait for it to be ready."
}

# Function to stop MySQL database
proj::mysql::stop() {
  proj::log::info "Stopping MySQL database..."
  cd "${MYSQL_COMPOSE_DIR}" && docker-compose down
}

# Function to restart MySQL database
proj::mysql::restart() {
  proj::log::info "Restarting MySQL database..."
  proj::mysql::stop
  proj::mysql::start
}

# Function to show MySQL database logs
proj::mysql::logs() {
  cd "${MYSQL_COMPOSE_DIR}" && docker-compose logs -f mysql
}

# Function to check MySQL database status
proj::mysql::status() {
  proj::log::info "Checking MySQL status..."
  cd "${MYSQL_COMPOSE_DIR}" && docker-compose ps
}

# Function to run database migrations
proj::mysql::migrate() {
  proj::log::info "Running database migrations..."

  # Check if MySQL client is available
  if ! command -v mysql >/dev/null 2>&1; then
    proj::log::error "MySQL client not found. Please install MySQL client first."
    proj::log::info "On macOS: brew install mysql-client"
    proj::log::info "On Ubuntu: sudo apt-get install mysql-client"
    return 1
  fi

  proj::log::info "Applying migration: 001_create_users_table.sql"
  mysql -h "${PROJ_MYSQL_HOST}" -P "${PROJ_MYSQL_PORT}" \
        -u "${PROJ_MYSQL_ADMIN_USERNAME}" -p"${PROJ_MYSQL_ADMIN_PASSWORD}" \
        "${PROJ_MYSQL_DATABASE}" < "${MIGRATIONS_DIR}/001_create_users_table.sql"

  proj::log::info "Database migrations completed successfully!"
}

# Function to show what migrations would be applied (dry run)
proj::mysql::migrate_dry() {
  proj::log::info "=== Database Migration Plan ==="
  proj::log::info "The following SQL will be executed:"
  proj::log::info "File: ${MIGRATIONS_DIR}/001_create_users_table.sql"
  echo ""
  cat "${MIGRATIONS_DIR}/001_create_users_table.sql"
}

# Function to connect to MySQL database using CLI
proj::mysql::connect() {
  proj::log::info "Connecting to MySQL database..."
  mysql -h "${PROJ_MYSQL_HOST}" -P "${PROJ_MYSQL_PORT}" \
        -u "${PROJ_MYSQL_ADMIN_USERNAME}" -p"${PROJ_MYSQL_ADMIN_PASSWORD}"
}

# Function to open MySQL shell in container
proj::mysql::shell() {
  proj::log::info "Opening MySQL shell in container..."
  cd "${MYSQL_COMPOSE_DIR}" && docker-compose exec mysql mysql \
    -u "${PROJ_MYSQL_ADMIN_USERNAME}" -p"${PROJ_MYSQL_ADMIN_PASSWORD}" "${PROJ_MYSQL_DATABASE}"
}

# Function to create database backup
proj::mysql::backup() {
  proj::log::info "Creating database backup..."

  local backup_file="backup_${PROJ_MYSQL_DATABASE}_$(date +%Y%m%d_%H%M%S).sql"

  mysqldump -h "${PROJ_MYSQL_HOST}" -P "${PROJ_MYSQL_PORT}" \
            -u "${PROJ_MYSQL_ADMIN_USERNAME}" -p"${PROJ_MYSQL_ADMIN_PASSWORD}" \
            --single-transaction --routines --triggers "${PROJ_MYSQL_DATABASE}" > "${backup_file}"

  proj::log::info "Database backup created: ${backup_file}"
}

# Function to restore database from backup
proj::mysql::restore() {
  local backup_file="${1:-}"

  if [[ -z "${backup_file}" ]]; then
    proj::log::error "Please specify BACKUP_FILE. Usage: proj::mysql::restore backup.sql"
    return 1
  fi

  if [[ ! -f "${backup_file}" ]]; then
    proj::log::error "Backup file ${backup_file} not found!"
    return 1
  fi

  proj::log::info "Restoring database from ${backup_file}..."
  mysql -h "${PROJ_MYSQL_HOST}" -P "${PROJ_MYSQL_PORT}" \
        -u "${PROJ_MYSQL_ADMIN_USERNAME}" -p"${PROJ_MYSQL_ADMIN_PASSWORD}" \
        "${PROJ_MYSQL_DATABASE}" < "${backup_file}"

  proj::log::info "Database restore completed!"
}

# Function to reset database (WARNING: This will delete all data!)
proj::mysql::reset() {
  proj::log::warn "⚠️  WARNING: This will delete all data in the database!"
  echo "Are you sure you want to continue? [y/N]"
  read -r confirm

  if [[ "${confirm}" = "y" ]] || [[ "${confirm}" = "Y" ]]; then
    proj::log::info "Resetting database..."
    mysql -h "${PROJ_MYSQL_HOST}" -P "${PROJ_MYSQL_PORT}" \
          -u "${PROJ_MYSQL_ADMIN_USERNAME}" -p"${PROJ_MYSQL_ADMIN_PASSWORD}" \
          -e "DROP DATABASE IF EXISTS ${PROJ_MYSQL_DATABASE}; CREATE DATABASE ${PROJ_MYSQL_DATABASE} CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

    proj::mysql::migrate
    proj::log::info "Database reset completed!"
  else
    proj::log::info "Operation cancelled."
  fi
}

# Function to remove MySQL container and volumes (WARNING: All data will be lost!)
proj::mysql::clean() {
  proj::log::warn "⚠️  WARNING: This will remove all MySQL data permanently!"
  echo "Are you sure you want to continue? [y/N]"
  read -r confirm

  if [[ "${confirm}" = "y" ]] || [[ "${confirm}" = "Y" ]]; then
    proj::log::info "Removing MySQL container and volumes..."
    cd "${MYSQL_COMPOSE_DIR}" && docker-compose down -v
    docker volume rm go-protoc-mysql-data 2>/dev/null || true
    proj::log::info "MySQL cleanup completed!"
  else
    proj::log::info "Operation cancelled."
  fi
}

# Function to create a new migration file
proj::mysql::create_migration() {
  local migration_name="${1:-}"

  if [[ -z "${migration_name}" ]]; then
    proj::log::error "Please specify migration NAME. Usage: proj::mysql::create_migration add_user_table"
    return 1
  fi

  local migration_num
  migration_num=$(ls "${MIGRATIONS_DIR}/" | grep -E '^[0-9]+_' | wc -l | tr -d ' ')
  migration_num=$(printf "%03d" $((migration_num + 1)))

  local migration_file="${MIGRATIONS_DIR}/${migration_num}_${migration_name}.sql"

  cat > "${migration_file}" << EOF
-- Migration: ${migration_num}_${migration_name}.sql
-- Description: ${migration_name}
-- Created: $(date +%Y-%m-%d)

-- Add your SQL statements here

EOF

  proj::log::info "Migration file created: ${migration_file}"
}

# Function to show database commands help
proj::mysql::help() {
  echo "=== Database Management Commands ==="
  echo ""
  echo "Setup Commands:"
  echo "  proj::mysql::setup         - Complete database setup (start + migrate)"
  echo "  proj::mysql::start         - Start MySQL container"
  echo "  proj::mysql::stop          - Stop MySQL container"
  echo "  proj::mysql::restart       - Restart MySQL container"
  echo ""
  echo "Migration Commands:"
  echo "  proj::mysql::migrate       - Run all pending migrations"
  echo "  proj::mysql::migrate_dry   - Show migration plan (dry run)"
  echo "  proj::mysql::create_migration migration_name"
  echo ""
  echo "Database Operations:"
  echo "  proj::mysql::connect       - Connect via MySQL CLI"
  echo "  proj::mysql::shell         - Open MySQL shell in container"
  echo "  proj::mysql::backup        - Create database backup"
  echo "  proj::mysql::restore backup_file.sql"
  echo ""
  echo "Maintenance Commands:"
  echo "  proj::mysql::status        - Check container status"
  echo "  proj::mysql::logs          - Show container logs"
  echo "  proj::mysql::reset         - Reset database (delete all data)"
  echo "  proj::mysql::clean         - Remove container and volumes"
  echo ""
  echo "Configuration:"
  echo "  Database: ${PROJ_MYSQL_DATABASE}"
  echo "  Host: ${PROJ_MYSQL_HOST}:${PROJ_MYSQL_PORT}"
  echo "  Username: ${PROJ_MYSQL_ADMIN_USERNAME}"
  echo "  Password: ${PROJ_MYSQL_ADMIN_PASSWORD}"
}

# This block allows calling functions in this script directly from the command line.
# It checks if the script's arguments contain a function name with the 'proj::mysql::' prefix
# and, if so, executes that function.
# For example: ./mysql.sh proj::mysql::setup
if [[ "$*" =~ proj::mysql:: ]]; then
  eval "$*"
fi