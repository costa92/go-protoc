#!/usr/bin/env bash

# 简化的Redis Docker测试脚本
set -eEuo pipefail

# 基本配置
export PROJ_ROOT_DIR="/Users/costalong/code/go/src/github.com/costa92/go-protoc"
export REDIS_VERSION="7.2.4"
export PROJ_REDIS_PORT="6379"
export PROJ_REDIS_CONFIG_DIR="/Users/costalong/code/go/src/github.com/costa92/go-protoc/_data/redis/config"
export PROJ_REDIS_DATA_DIR="/Users/costalong/code/go/src/github.com/costa92/go-protoc/_data/redis/data"
export PROJ_ENVIRONMENT="development"

# 基础日志函数
log_info() {
    echo "[INFO] $*"
}

log_error() {
    echo "[ERROR] $*" >&2
}

log_success() {
    echo "[SUCCESS] $*"
}

# 模拟模板替换
generate_redis_script() {
    local template_file="$PROJ_ROOT_DIR/scripts/installation/templates/docker/redis/docker-run.sh.tpl"
    local output_file="/tmp/redis-docker-run.sh"
    
    log_info "Generating Redis Docker script from template"
    
    # 设置更多环境变量供envsubst使用
    export PROJ_NAME="go-protoc"
    export PROJ_REDIS_LOG_DIR="$PROJ_REDIS_DATA_DIR/logs"
    
    # 使用envsubst进行变量替换
    if command -v envsubst >/dev/null 2>&1; then
        envsubst < "$template_file" > "$output_file"
    else
        log_error "envsubst not found, please install gettext"
        exit 1
    fi
    
    chmod +x "$output_file"
    log_success "Redis Docker script generated: $output_file"
}

# 执行生成的脚本
run_redis() {
    local script_file="/tmp/redis-docker-run.sh"
    
    if [[ -x "$script_file" ]]; then
        log_info "Executing Redis Docker script"
        bash "$script_file"
    else
        log_error "Script not found or not executable: $script_file"
        exit 1
    fi
}

# 主流程
main() {
    log_info "Starting Redis Docker test..."
    
    # 检查Docker
    if ! command -v docker >/dev/null 2>&1; then
        log_error "Docker not found"
        exit 1
    fi
    
    if ! docker info >/dev/null 2>&1; then
        log_error "Docker daemon not running"
        exit 1
    fi
    
    # 检查模板文件
    local template_file="$PROJ_ROOT_DIR/scripts/installation/templates/docker/redis/docker-run.sh.tpl"
    if [[ ! -f "$template_file" ]]; then
        log_error "Template file not found: $template_file"
        exit 1
    fi
    
    # 创建必要的目录
    mkdir -p "$PROJ_REDIS_CONFIG_DIR" "$PROJ_REDIS_DATA_DIR"
    
    # 生成并运行脚本
    generate_redis_script
    run_redis
    
    log_success "Redis Docker test completed"
}

main "$@"