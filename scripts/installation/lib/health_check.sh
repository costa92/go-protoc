#!/usr/bin/env bash

# =============================================================================
# 健康检查库
# Health Check Library
# 
# 提供服务健康检查、依赖验证和系统监控功能
# =============================================================================

set -eEuo pipefail

# 避免重复加载
[[ "${HEALTH_CHECK_LOADED:-false}" == "true" ]] && return 0

# 默认健康检查配置
HEALTH_CHECK_TIMEOUT="${HEALTH_CHECK_TIMEOUT:-30}"
HEALTH_CHECK_INTERVAL="${HEALTH_CHECK_INTERVAL:-5}"
HEALTH_CHECK_MAX_RETRIES="${HEALTH_CHECK_MAX_RETRIES:-6}"

# =============================================================================
# HTTP 健康检查函数
# =============================================================================

# HTTP健康检查
proj::health::check_http() {
    local url="$1"
    local timeout="${2:-${HEALTH_CHECK_TIMEOUT}}"
    local expected_status="${3:-200}"
    
    local response_code
    response_code=$(curl -s -o /dev/null -w "%{http_code}" \
        --max-time "${timeout}" \
        --connect-timeout 5 \
        "${url}" 2>/dev/null || echo "000")
    
    if [[ "${response_code}" == "${expected_status}" ]]; then
        proj::log::debug "HTTP health check passed: ${url} (${response_code})"
        return 0
    else
        proj::log::debug "HTTP health check failed: ${url} (${response_code})"
        return 1
    fi
}

# 等待HTTP服务可用
proj::health::wait_for_http() {
    local url="$1"
    local timeout="${2:-${HEALTH_CHECK_TIMEOUT}}"
    local interval="${3:-${HEALTH_CHECK_INTERVAL}}"
    
    local end_time=$((SECONDS + timeout))
    
    proj::log::info "Waiting for HTTP service: ${url}"
    
    while [[ $SECONDS -lt $end_time ]]; do
        if proj::health::check_http "${url}" 5; then
            proj::log::info "HTTP service is ready: ${url}"
            return 0
        fi
        
        sleep "${interval}"
    done
    
    proj::log::error "HTTP service failed to become ready: ${url}"
    return 1
}

# =============================================================================
# 端口健康检查函数
# =============================================================================

# 检查端口是否开放
proj::health::check_port() {
    local host="$1"
    local port="$2"
    local timeout="${3:-5}"
    
    if command -v nc >/dev/null 2>&1; then
        # 使用netcat检查
        if nc -z -w"${timeout}" "${host}" "${port}" 2>/dev/null; then
            proj::log::debug "Port check passed: ${host}:${port}"
            return 0
        fi
    elif command -v telnet >/dev/null 2>&1; then
        # 使用telnet检查
        if timeout "${timeout}" telnet "${host}" "${port}" </dev/null >/dev/null 2>&1; then
            proj::log::debug "Port check passed: ${host}:${port}"
            return 0
        fi
    else
        # 使用bash内置的/dev/tcp检查
        if timeout "${timeout}" bash -c "exec 3<>/dev/tcp/${host}/${port}" 2>/dev/null; then
            exec 3<&-
            exec 3>&-
            proj::log::debug "Port check passed: ${host}:${port}"
            return 0
        fi
    fi
    
    proj::log::debug "Port check failed: ${host}:${port}"
    return 1
}

# 等待端口可用
proj::health::wait_for_port() {
    local host="$1"
    local port="$2"
    local timeout="${3:-${HEALTH_CHECK_TIMEOUT}}"
    local interval="${4:-${HEALTH_CHECK_INTERVAL}}"
    
    local end_time=$((SECONDS + timeout))
    
    proj::log::info "Waiting for port: ${host}:${port}"
    
    while [[ $SECONDS -lt $end_time ]]; do
        if proj::health::check_port "${host}" "${port}"; then
            proj::log::info "Port is ready: ${host}:${port}"
            return 0
        fi
        
        sleep "${interval}"
    done
    
    proj::log::error "Port failed to become ready: ${host}:${port}"
    return 1
}

# =============================================================================
# Docker 健康检查函数
# =============================================================================

# 检查Docker容器状态
proj::health::check_docker_container() {
    local container_name="$1"
    
    if ! command -v docker >/dev/null 2>&1; then
        proj::log::error "Docker is not installed"
        return 1
    fi
    
    local status
    status=$(docker inspect --format='{{.State.Status}}' "${container_name}" 2>/dev/null || echo "not_found")
    
    case "${status}" in
        "running")
            proj::log::debug "Docker container is running: ${container_name}"
            return 0
            ;;
        "not_found")
            proj::log::debug "Docker container not found: ${container_name}"
            return 1
            ;;
        *)
            proj::log::debug "Docker container status: ${container_name} (${status})"
            return 1
            ;;
    esac
}

# 检查Docker容器健康状态
proj::health::check_docker_health() {
    local container_name="$1"
    
    if ! proj::health::check_docker_container "${container_name}"; then
        return 1
    fi
    
    local health_status
    health_status=$(docker inspect --format='{{.State.Health.Status}}' "${container_name}" 2>/dev/null || echo "no_healthcheck")
    
    case "${health_status}" in
        "healthy")
            proj::log::debug "Docker container is healthy: ${container_name}"
            return 0
            ;;
        "no_healthcheck")
            proj::log::debug "Docker container has no health check: ${container_name}"
            return 0
            ;;
        "starting")
            proj::log::debug "Docker container health check is starting: ${container_name}"
            return 1
            ;;
        "unhealthy")
            proj::log::debug "Docker container is unhealthy: ${container_name}"
            return 1
            ;;
        *)
            proj::log::debug "Docker container health status unknown: ${container_name} (${health_status})"
            return 1
            ;;
    esac
}

# 等待Docker容器健康
proj::health::wait_for_docker() {
    local container_name="$1"
    local timeout="${2:-${HEALTH_CHECK_TIMEOUT}}"
    local interval="${3:-${HEALTH_CHECK_INTERVAL}}"
    
    local end_time=$((SECONDS + timeout))
    
    proj::log::info "Waiting for Docker container: ${container_name}"
    
    while [[ $SECONDS -lt $end_time ]]; do
        if proj::health::check_docker_health "${container_name}"; then
            proj::log::info "Docker container is ready: ${container_name}"
            return 0
        fi
        
        sleep "${interval}"
    done
    
    proj::log::error "Docker container failed to become ready: ${container_name}"
    return 1
}

# =============================================================================
# 系统健康检查函数
# =============================================================================

# 检查系统命令是否可用
proj::health::check_command() {
    local cmd="$1"
    
    if command -v "${cmd}" >/dev/null 2>&1; then
        proj::log::debug "Command available: ${cmd}"
        return 0
    else
        proj::log::debug "Command not available: ${cmd}"
        return 1
    fi
}

# 检查文件是否存在且可读
proj::health::check_file() {
    local file_path="$1"
    
    if [[ -r "${file_path}" ]]; then
        proj::log::debug "File accessible: ${file_path}"
        return 0
    else
        proj::log::debug "File not accessible: ${file_path}"
        return 1
    fi
}

# 检查目录是否存在且可写
proj::health::check_directory() {
    local dir_path="$1"
    local create_if_missing="${2:-false}"
    
    if [[ -d "${dir_path}" ]]; then
        if [[ -w "${dir_path}" ]]; then
            proj::log::debug "Directory writable: ${dir_path}"
            return 0
        else
            proj::log::debug "Directory not writable: ${dir_path}"
            return 1
        fi
    elif [[ "${create_if_missing}" == "true" ]]; then
        if mkdir -p "${dir_path}" 2>/dev/null; then
            proj::log::debug "Directory created: ${dir_path}"
            return 0
        else
            proj::log::debug "Directory creation failed: ${dir_path}"
            return 1
        fi
    else
        proj::log::debug "Directory not found: ${dir_path}"
        return 1
    fi
}

# =============================================================================
# 复合健康检查函数
# =============================================================================

# 综合健康检查
proj::health::check_service() {
    local service_name="$1"
    local check_type="${2:-auto}"
    
    case "${check_type}" in
        "http")
            local url="${3:-http://localhost:8080/health}"
            proj::health::check_http "${url}"
            ;;
        "port")
            local host="${3:-localhost}"
            local port="${4:-8080}"
            proj::health::check_port "${host}" "${port}"
            ;;
        "docker")
            local container_name="${3:-${PROJ_PREFIX:-proj}-${service_name}}"
            proj::health::check_docker_health "${container_name}"
            ;;
        "auto")
            # 自动检测检查方式
            local container_name="${PROJ_PREFIX:-proj}-${service_name}"
            if proj::health::check_docker_container "${container_name}"; then
                proj::health::check_docker_health "${container_name}"
            else
                proj::log::warn "Auto-detection failed for service: ${service_name}"
                return 1
            fi
            ;;
        *)
            proj::log::error "Unknown health check type: ${check_type}"
            return 1
            ;;
    esac
}

# 等待服务就绪
proj::health::wait_for_service() {
    local service_name="$1"
    local check_type="${2:-auto}"
    local timeout="${3:-${HEALTH_CHECK_TIMEOUT}}"
    local interval="${4:-${HEALTH_CHECK_INTERVAL}}"
    
    local end_time=$((SECONDS + timeout))
    
    proj::log::info "Waiting for service: ${service_name}"
    
    while [[ $SECONDS -lt $end_time ]]; do
        if proj::health::check_service "${service_name}" "${check_type}" "${@:5}"; then
            proj::log::info "Service is ready: ${service_name}"
            return 0
        fi
        
        sleep "${interval}"
    done
    
    proj::log::error "Service failed to become ready: ${service_name}"
    return 1
}

# =============================================================================
# 批量健康检查函数
# =============================================================================

# 检查多个服务
proj::health::check_services() {
    local -a services=("$@")
    local failed_services=()
    
    proj::log::info "Checking health of ${#services[@]} services..."
    
    for service in "${services[@]}"; do
        if proj::health::check_service "${service}"; then
            proj::log::info "✓ ${service}: healthy"
        else
            proj::log::error "✗ ${service}: unhealthy"
            failed_services+=("${service}")
        fi
    done
    
    if [[ ${#failed_services[@]} -gt 0 ]]; then
        proj::log::error "Failed services: ${failed_services[*]}"
        return 1
    else
        proj::log::info "All services are healthy"
        return 0
    fi
}

# =============================================================================
# 初始化函数
# =============================================================================

# 健康检查库初始化
proj::health::init() {
    # 检查基础工具
    local missing_tools=()
    
    if ! command -v curl >/dev/null 2>&1; then
        missing_tools+=("curl")
    fi
    
    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        proj::log::warn "Some health check tools are missing: ${missing_tools[*]}"
        proj::log::warn "This may affect health check functionality"
    fi
    
    proj::log::debug "Health check library initialized"
}

# 自动初始化
if [[ "${BASH_SOURCE[0]}" == "${0}" ]] || [[ "${HEALTH_CHECK_AUTO_INIT:-true}" == "true" ]]; then
    proj::health::init
fi

# 标记已加载
export HEALTH_CHECK_LOADED=true