#!/usr/bin/env bash

# =============================================================================
# Docker脚本管理器
# Docker Script Manager
#
# 专门管理基于模板生成的Docker运行脚本，替代docker-compose
# =============================================================================

set -eEuo pipefail

# 避免重复加载
[[ "${DOCKER_SCRIPT_MANAGER_LIB_LOADED:-false}" == "true" ]] && return 0

# Docker脚本管理常量
readonly DOCKER_SCRIPTS_DIR="${PROJ_ROOT_DIR}/_generated/docker-scripts"
readonly DOCKER_NETWORK_NAME="proj-network"

# 生成并运行Docker脚本
proj::docker::run_service_from_template() {
    local service_name="$1"
    local extra_vars="${2:-}"
    
    if [[ -z "$service_name" ]]; then
        proj::log::error "Service name is required"
        return 1
    fi
    
    proj::log::info "Running Docker service: $service_name"
    
    # 确保Docker可用
    if ! proj::docker::check_docker_available; then
        return 1
    fi
    
    # 生成Docker脚本
    local script_dir="${DOCKER_SCRIPTS_DIR}/${service_name}"
    if ! proj::docker::generate_service_scripts "$service_name" "$script_dir" "$extra_vars"; then
        proj::log::error "Failed to generate Docker scripts for $service_name"
        return 1
    fi
    
    # 执行运行脚本
    local run_script="${script_dir}/docker-run.sh"
    if [[ -x "$run_script" ]]; then
        proj::log::info "Executing Docker run script: $run_script"
        bash "$run_script"
    else
        proj::log::error "Docker run script not found or not executable: $run_script"
        return 1
    fi
}

# 停止Docker服务
proj::docker::stop_service_from_template() {
    local service_name="$1"
    
    if [[ -z "$service_name" ]]; then
        proj::log::error "Service name is required"
        return 1
    fi
    
    proj::log::info "Stopping Docker service: $service_name"
    
    # 查找停止脚本
    local script_dir="${DOCKER_SCRIPTS_DIR}/${service_name}"
    local stop_script="${script_dir}/docker-stop.sh"
    
    if [[ -x "$stop_script" ]]; then
        proj::log::info "Executing Docker stop script: $stop_script"
        bash "$stop_script" "${2:-}"  # 传递额外参数（如 --remove-data）
    else
        proj::log::warn "Docker stop script not found, using fallback method"
        proj::docker::stop_container_by_name "proj-${service_name}"
    fi
}

# 检查Docker服务状态
proj::docker::check_service_status_from_template() {
    local service_name="$1"
    
    if [[ -z "$service_name" ]]; then
        proj::log::error "Service name is required"
        return 1
    fi
    
    # 查找状态检查脚本
    local script_dir="${DOCKER_SCRIPTS_DIR}/${service_name}"
    local status_script="${script_dir}/docker-status.sh"
    
    if [[ -x "$status_script" ]]; then
        proj::log::debug "Executing Docker status script: $status_script"
        bash "$status_script"
    else
        proj::log::warn "Docker status script not found, using fallback method"
        proj::docker::show_container_status "proj-${service_name}"
    fi
}

# 生成服务的Docker脚本
proj::docker::generate_service_scripts() {
    local service_name="$1"
    local output_dir="$2"
    local extra_vars="${3:-}"
    
    if [[ -z "$service_name" ]] || [[ -z "$output_dir" ]]; then
        proj::log::error "Service name and output directory are required"
        return 1
    fi
    
    proj::log::info "Generating Docker scripts for service: $service_name"
    
    # 确保输出目录存在
    mkdir -p "$output_dir"
    
    # 使用模板管理器生成脚本
    if ! proj::template::generate_service_configs "$service_name" "docker" "$output_dir" "$extra_vars"; then
        proj::log::error "Failed to generate Docker scripts from templates"
        return 1
    fi
    
    # 设置脚本执行权限
    find "$output_dir" -name "*.sh" -exec chmod +x {} \;
    
    proj::log::success "Docker scripts generated for $service_name in $output_dir"
}

# 停止容器（备用方法）
proj::docker::stop_container_by_name() {
    local container_name="$1"
    
    if [[ -z "$container_name" ]]; then
        proj::log::error "Container name is required"
        return 1
    fi
    
    proj::log::info "Stopping container: $container_name"
    
    # 停止容器
    if docker ps --format "{{.Names}}" | grep -q "^${container_name}$"; then
        docker stop "$container_name" || {
            proj::log::error "Failed to stop container: $container_name"
            return 1
        }
        proj::log::success "Container stopped: $container_name"
    else
        proj::log::info "Container not running: $container_name"
    fi
    
    # 删除容器
    if docker ps -a --format "{{.Names}}" | grep -q "^${container_name}$"; then
        docker rm "$container_name" || {
            proj::log::error "Failed to remove container: $container_name"
            return 1
        }
        proj::log::success "Container removed: $container_name"
    fi
}

# 显示容器状态（备用方法）
proj::docker::show_container_status() {
    local container_name="$1"
    
    if [[ -z "$container_name" ]]; then
        proj::log::error "Container name is required"
        return 1
    fi
    
    echo "=== Docker容器状态: $container_name ==="
    
    # 检查容器是否存在
    if ! docker ps -a --format "{{.Names}}" | grep -q "^${container_name}$"; then
        echo "状态: 容器不存在"
        return 1
    fi
    
    # 显示容器信息
    echo "容器信息:"
    docker ps -a --filter name="^${container_name}$" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}\t{{.Image}}"
    
    # 如果容器正在运行，显示更多信息
    if docker ps --format "{{.Names}}" | grep -q "^${container_name}$"; then
        echo ""
        echo "资源使用:"
        docker stats "$container_name" --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}"
        
        echo ""
        echo "最近日志 (最后10行):"
        docker logs "$container_name" --tail 10 2>/dev/null || echo "无法获取日志"
    fi
}

# 批量管理Docker服务
proj::docker::manage_services() {
    local action="$1"  # start|stop|status|restart
    shift
    local services=("$@")
    
    if [[ -z "$action" ]]; then
        proj::log::error "Action is required (start|stop|status|restart)"
        return 1
    fi
    
    if [[ ${#services[@]} -eq 0 ]]; then
        proj::log::error "At least one service name is required"
        return 1
    fi
    
    proj::log::info "Performing batch $action on ${#services[@]} Docker services"
    
    local failed_services=()
    local successful_services=()
    
    for service in "${services[@]}"; do
        proj::log::info "Running $action on Docker service: $service"
        
        case "$action" in
            "start")
                if proj::docker::run_service_from_template "$service"; then
                    successful_services+=("$service")
                else
                    failed_services+=("$service")
                fi
                ;;
            "stop")
                if proj::docker::stop_service_from_template "$service"; then
                    successful_services+=("$service")
                else
                    failed_services+=("$service")
                fi
                ;;
            "status")
                if proj::docker::check_service_status_from_template "$service"; then
                    successful_services+=("$service")
                else
                    failed_services+=("$service")
                fi
                ;;
            "restart")
                if proj::docker::stop_service_from_template "$service" && \
                   sleep 2 && \
                   proj::docker::run_service_from_template "$service"; then
                    successful_services+=("$service")
                else
                    failed_services+=("$service")
                fi
                ;;
            *)
                proj::log::error "Unknown action: $action"
                return 1
                ;;
        esac
    done
    
    # 报告结果
    if [[ ${#successful_services[@]} -gt 0 ]]; then
        proj::log::success "Successfully processed Docker services: ${successful_services[*]}"
    fi
    
    if [[ ${#failed_services[@]} -gt 0 ]]; then
        proj::log::error "Failed to process Docker services: ${failed_services[*]}"
        return 1
    fi
    
    return 0
}

# 确保Docker网络存在
proj::docker::ensure_project_network() {
    local network_name="${1:-$DOCKER_NETWORK_NAME}"
    
    if ! docker network ls --format "{{.Name}}" | grep -q "^${network_name}$"; then
        proj::log::info "Creating Docker network: $network_name"
        docker network create "$network_name" || {
            proj::log::error "Failed to create Docker network: $network_name"
            return 1
        }
        proj::log::success "Docker network created: $network_name"
    else
        proj::log::debug "Docker network already exists: $network_name"
    fi
}

# 清理所有生成的脚本
proj::docker::cleanup_generated_scripts() {
    local service_name="${1:-}"
    
    if [[ -n "$service_name" ]]; then
        # 清理特定服务的脚本
        local service_script_dir="${DOCKER_SCRIPTS_DIR}/${service_name}"
        if [[ -d "$service_script_dir" ]]; then
            proj::log::info "Cleaning up Docker scripts for service: $service_name"
            rm -rf "$service_script_dir"
            proj::log::success "Docker scripts cleaned up for: $service_name"
        fi
    else
        # 清理所有生成的脚本
        if [[ -d "$DOCKER_SCRIPTS_DIR" ]]; then
            proj::log::info "Cleaning up all generated Docker scripts"
            rm -rf "$DOCKER_SCRIPTS_DIR"
            proj::log::success "All Docker scripts cleaned up"
        fi
    fi
}

# 列出生成的脚本
proj::docker::list_generated_scripts() {
    local service_name="${1:-}"
    
    if [[ ! -d "$DOCKER_SCRIPTS_DIR" ]]; then
        proj::log::info "No generated Docker scripts found"
        return 0
    fi
    
    if [[ -n "$service_name" ]]; then
        # 列出特定服务的脚本
        local service_script_dir="${DOCKER_SCRIPTS_DIR}/${service_name}"
        if [[ -d "$service_script_dir" ]]; then
            proj::log::info "Generated Docker scripts for service '$service_name':"
            find "$service_script_dir" -name "*.sh" -type f | sed 's|.*/||' | sort
        else
            proj::log::info "No Docker scripts found for service: $service_name"
        fi
    else
        # 列出所有服务的脚本
        proj::log::info "All generated Docker scripts:"
        find "$DOCKER_SCRIPTS_DIR" -name "*.sh" -type f | sed "s|^${DOCKER_SCRIPTS_DIR}/||" | sort
    fi
}

# 标记已加载
export DOCKER_SCRIPT_MANAGER_LIB_LOADED=true

# 如果直接执行此脚本，显示生成的脚本列表
if [[ "${BASH_SOURCE[0]:-${0##*/}}" == "${0##*/}" ]]; then
    proj::docker::list_generated_scripts "$@"
fi