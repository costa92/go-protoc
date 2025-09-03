#!/usr/bin/env bash

# =============================================================================
# Docker标准化工具库
# Docker Helper Library
#
# 提供统一的Docker容器操作接口，标准化容器启动、管理和清理流程
# =============================================================================

set -eEuo pipefail

# 避免重复加载
[[ "${DOCKER_HELPER_LIB_LOADED:-false}" == "true" ]] && return 0

# Docker配置常量
readonly DOCKER_RESTART_POLICY="${DOCKER_RESTART_POLICY:-unless-stopped}"
readonly DOCKER_NETWORK_DRIVER="${DOCKER_NETWORK_DRIVER:-bridge}"

# 统一Docker服务启动函数
proj::docker::run_service() {
    local service_name="$1"
    local image="$2"
    local -n ports_ref=$3
    local -n volumes_ref=$4  
    local -n env_vars_ref=$5
    local extra_args="${6:-}"
    
    # 验证参数
    if [[ -z "$service_name" ]] || [[ -z "$image" ]]; then
        proj::log::error "Service name and image are required"
        return 1
    fi
    
    # 确保Docker可用
    if ! proj::docker::check_docker_available; then
        return 1
    fi
    
    # 构建容器配置
    local container_name="${NETWORK_NAME:-proj}-${service_name}"
    local docker_args=(
        "--name" "$container_name"
        "--restart" "$DOCKER_RESTART_POLICY"
        "--detach"
    )
    
    # 添加网络配置
    proj::docker::ensure_network "${NETWORK_NAME:-proj}"
    docker_args+=("--network" "${NETWORK_NAME:-proj}")
    
    # 动态构建端口映射
    for port_mapping in "${ports_ref[@]}"; do
        # 支持格式: "8080:8080" 或 "127.0.0.1:8080:8080"
        if [[ "$port_mapping" =~ ^[0-9]+:[0-9]+$ ]]; then
            # 简单格式，绑定到localhost
            docker_args+=("-p" "127.0.0.1:$port_mapping")
        else
            # 完整格式或已包含IP
            docker_args+=("-p" "$port_mapping")
        fi
    done
    
    # 动态构建卷映射
    for volume_mapping in "${volumes_ref[@]}"; do
        docker_args+=("-v" "$volume_mapping")
    done
    
    # 动态构建环境变量
    for env_var in "${env_vars_ref[@]}"; do
        docker_args+=("-e" "$env_var")
    done
    
    # 添加额外参数
    if [[ -n "$extra_args" ]]; then
        # 将额外参数按空格分割并添加
        read -ra extra_array <<< "$extra_args"
        docker_args+=("${extra_array[@]}")
    fi
    
    # 清理可能存在的同名容器
    proj::docker::cleanup_container "$container_name" || true
    
    # 执行Docker run命令
    proj::log::info "Starting Docker container: $container_name"
    proj::log::debug "Docker command: docker run ${docker_args[*]} $image"
    
    if docker run "${docker_args[@]}" "$image"; then
        proj::log::success "Container $container_name started successfully"
        
        # 等待容器完全启动
        proj::docker::wait_for_container "$container_name" 30
        
        return 0
    else
        local exit_code=$?
        proj::log::error "Failed to start container $container_name"
        
        # 显示容器日志以便调试
        proj::log::info "Container logs:"
        docker logs "$container_name" --tail 50 2>/dev/null || true
        
        return $exit_code
    fi
}

# 容器清理函数
proj::docker::cleanup_container() {
    local container_name="$1"
    
    if [[ -z "$container_name" ]]; then
        proj::log::error "Container name is required for cleanup"
        return 1
    fi
    
    # 检查容器是否存在
    if docker ps -a --format "{{.Names}}" | grep -q "^${container_name}$"; then
        proj::log::info "Cleaning up existing container: $container_name"
        
        # 停止容器（如果正在运行）
        if docker ps --format "{{.Names}}" | grep -q "^${container_name}$"; then
            proj::log::debug "Stopping running container: $container_name"
            docker stop "$container_name" >/dev/null 2>&1 || true
        fi
        
        # 删除容器
        proj::log::debug "Removing container: $container_name"
        docker rm "$container_name" >/dev/null 2>&1 || true
        
        proj::log::success "Container $container_name cleaned up successfully"
    else
        proj::log::debug "Container $container_name does not exist, skipping cleanup"
    fi
    
    return 0
}

# 容器状态检查
proj::docker::check_container_status() {
    local container_name="$1"
    local expected_status="${2:-running}"  # running|exited|created|paused
    
    if [[ -z "$container_name" ]]; then
        proj::log::error "Container name is required for status check"
        return 1
    fi
    
    # 检查容器是否存在
    if ! docker ps -a --format "{{.Names}}" | grep -q "^${container_name}$"; then
        proj::log::warn "Container $container_name does not exist"
        return 1
    fi
    
    # 获取容器状态
    local actual_status
    actual_status=$(docker inspect --format="{{.State.Status}}" "$container_name" 2>/dev/null)
    
    if [[ "$actual_status" == "$expected_status" ]]; then
        proj::log::success "Container $container_name is $actual_status"
        return 0
    else
        proj::log::warn "Container $container_name status is '$actual_status', expected '$expected_status'"
        return 1
    fi
}

# 等待容器启动
proj::docker::wait_for_container() {
    local container_name="$1"
    local timeout="${2:-30}"
    local check_interval=2
    local elapsed=0
    
    proj::log::info "Waiting for container $container_name to be ready..."
    
    while [[ $elapsed -lt $timeout ]]; do
        if proj::docker::check_container_status "$container_name" "running"; then
            proj::log::success "Container $container_name is ready"
            return 0
        fi
        
        sleep $check_interval
        elapsed=$((elapsed + check_interval))
    done
    
    proj::log::error "Container $container_name failed to start within ${timeout}s"
    return 1
}

# 确保Docker网络存在
proj::docker::ensure_network() {
    local network_name="${1:-proj}"
    
    # 检查网络是否存在
    if ! docker network ls --format "{{.Name}}" | grep -q "^${network_name}$"; then
        proj::log::info "Creating Docker network: $network_name"
        
        if docker network create --driver "$DOCKER_NETWORK_DRIVER" "$network_name" >/dev/null 2>&1; then
            proj::log::success "Docker network $network_name created successfully"
        else
            proj::log::warn "Failed to create network $network_name, it may already exist"
        fi
    else
        proj::log::debug "Docker network $network_name already exists"
    fi
}

# 检查Docker是否可用
proj::docker::check_docker_available() {
    if ! command -v docker >/dev/null 2>&1; then
        proj::log::error "Docker command not found. Please install Docker first."
        return 1
    fi
    
    if ! docker info >/dev/null 2>&1; then
        proj::log::error "Cannot connect to Docker daemon. Please ensure Docker is running."
        return 1
    fi
    
    return 0
}

# 获取容器IP地址
proj::docker::get_container_ip() {
    local container_name="$1"
    local network_name="${2:-${NETWORK_NAME:-proj}}"
    
    if [[ -z "$container_name" ]]; then
        proj::log::error "Container name is required"
        return 1
    fi
    
    # 检查容器是否存在且运行中
    if ! proj::docker::check_container_status "$container_name" "running"; then
        return 1
    fi
    
    # 获取指定网络中的IP地址
    local ip_address
    ip_address=$(docker inspect --format="{{.NetworkSettings.Networks.${network_name}.IPAddress}}" "$container_name" 2>/dev/null)
    
    if [[ -n "$ip_address" ]] && [[ "$ip_address" != "<no value>" ]]; then
        echo "$ip_address"
        return 0
    else
        proj::log::warn "Could not get IP address for container $container_name in network $network_name"
        return 1
    fi
}

# 显示容器的网络信息
proj::docker::show_network_info() {
    local container_name="$1"
    
    if [[ -z "$container_name" ]]; then
        proj::log::error "Container name is required"
        return 1
    fi
    
    # 检查容器是否存在
    if ! docker ps -a --format '{{.Names}}' | grep -q "^${container_name}$"; then
        proj::log::warn "Container $container_name does not exist"
        return 1
    fi
    
    # 获取网络信息
    local networks
    networks=$(docker inspect "$container_name" --format '{{range $net, $conf := .NetworkSettings.Networks}}{{$net}} {{end}}' 2>/dev/null)
    
    if [[ -n "$networks" ]]; then
        proj::log::info "🌐 Network: ${networks}"
        
        # 获取每个网络的IP地址
        for network in $networks; do
            local ip
            ip=$(docker inspect "$container_name" --format "{{.NetworkSettings.Networks.${network}.IPAddress}}" 2>/dev/null)
            if [[ -n "$ip" ]] && [[ "$ip" != "<no value>" ]]; then
                proj::log::info "   • IP in $network: $ip"
            fi
        done
    else
        proj::log::warn "No network information available"
    fi
}

# 显示容器日志
proj::docker::show_logs() {
    local container_name="$1"
    local lines="${2:-50}"
    local follow="${3:-false}"
    
    if [[ -z "$container_name" ]]; then
        proj::log::error "Container name is required"
        return 1
    fi
    
    # 检查容器是否存在
    if ! docker ps -a --format "{{.Names}}" | grep -q "^${container_name}$"; then
        proj::log::error "Container $container_name does not exist"
        return 1
    fi
    
    proj::log::info "Showing logs for container $container_name (last $lines lines)"
    
    if [[ "$follow" == "true" ]]; then
        docker logs --tail "$lines" --follow "$container_name"
    else
        docker logs --tail "$lines" "$container_name"
    fi
}

# 批量清理容器
proj::docker::cleanup_service_containers() {
    local service_pattern="${1:-${NETWORK_NAME:-proj}-*}"
    local containers_found=false
    
    proj::log::info "Cleaning up containers matching pattern: $service_pattern"
    
    # 获取匹配的容器名称
    while IFS= read -r container_name; do
        if [[ -n "$container_name" ]]; then
            containers_found=true
            proj::docker::cleanup_container "$container_name"
        fi
    done < <(docker ps -a --format "{{.Names}}" | grep "^${service_pattern//\*/.*}$" 2>/dev/null || true)
    
    if [[ "$containers_found" == "false" ]]; then
        proj::log::info "No containers found matching pattern: $service_pattern"
    fi
}

# 容器健康检查
proj::docker::health_check() {
    local container_name="$1"
    local max_attempts="${2:-5}"
    local check_interval="${3:-2}"
    
    if [[ -z "$container_name" ]]; then
        proj::log::error "Container name is required for health check"
        return 1
    fi
    
    proj::log::info "Performing health check for container: $container_name"
    
    local attempt=1
    while [[ $attempt -le $max_attempts ]]; do
        if proj::docker::check_container_status "$container_name" "running"; then
            # 检查容器是否healthy（如果有健康检查配置）
            local health_status
            health_status=$(docker inspect --format="{{.State.Health.Status}}" "$container_name" 2>/dev/null || echo "unknown")
            
            case "$health_status" in
                "healthy"|"unknown")
                    proj::log::success "Container $container_name health check passed"
                    return 0
                    ;;
                "unhealthy")
                    proj::log::warn "Container $container_name is unhealthy (attempt $attempt/$max_attempts)"
                    ;;
                "starting")
                    proj::log::info "Container $container_name health check is starting (attempt $attempt/$max_attempts)"
                    ;;
                *)
                    proj::log::warn "Container $container_name health status: $health_status (attempt $attempt/$max_attempts)"
                    ;;
            esac
        else
            proj::log::warn "Container $container_name is not running (attempt $attempt/$max_attempts)"
        fi
        
        if [[ $attempt -lt $max_attempts ]]; then
            sleep $check_interval
        fi
        attempt=$((attempt + 1))
    done
    
    proj::log::error "Container $container_name health check failed after $max_attempts attempts"
    return 1
}

# 标记已加载
export DOCKER_HELPER_LIB_LOADED=true

# 如果直接执行此脚本，显示Docker状态
if [[ "${BASH_SOURCE[0]:-${0}}" == "${0}" ]]; then
    proj::docker::check_docker_available && echo "Docker is available" || echo "Docker is not available"
fi