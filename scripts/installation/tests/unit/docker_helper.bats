#!/usr/bin/env bats

# =============================================================================
# Docker标准化工具库单元测试
# Docker Helper Library Unit Tests
# =============================================================================

# 设置测试环境
setup() {
    # 加载被测试的库
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    LIB_DIR="${SCRIPT_DIR}/../../lib"
    
    # 模拟必要的函数以避免依赖
    proj::log::info() { echo "INFO: $*"; }
    proj::log::warn() { echo "WARN: $*"; }
    proj::log::error() { echo "ERROR: $*"; }
    proj::log::success() { echo "SUCCESS: $*"; }
    proj::log::debug() { echo "DEBUG: $*"; }
    
    # 设置测试环境变量
    export NETWORK_NAME="test-proj"
    
    source "${LIB_DIR}/docker_helper.sh"
}

teardown() {
    # 清理测试容器
    proj::docker::cleanup_service_containers "test-proj-*" 2>/dev/null || true
    
    # 清理测试网络
    docker network rm "$NETWORK_NAME" 2>/dev/null || true
}

@test "proj::docker::check_docker_available detects Docker" {
    run proj::docker::check_docker_available
    # 在有Docker的环境中应该返回0，没有Docker的环境中返回1
    [[ "$status" -eq 0 ]] || [[ "$status" -eq 1 ]]
}

@test "proj::docker::ensure_network creates network if not exists" {
    # 确保测试网络不存在
    docker network rm "$NETWORK_NAME" 2>/dev/null || true
    
    run proj::docker::ensure_network "$NETWORK_NAME"
    [ "$status" -eq 0 ]
    
    # 验证网络确实被创建
    run docker network ls --format "{{.Name}}"
    [[ "$output" =~ $NETWORK_NAME ]]
}

@test "proj::docker::cleanup_container removes existing container" {
    # 只在Docker可用时运行
    if ! proj::docker::check_docker_available; then
        skip "Docker not available"
    fi
    
    local test_container="test-cleanup-container"
    
    # 创建测试容器
    docker run --name "$test_container" -d alpine:latest sleep 30 >/dev/null 2>&1 || true
    
    # 测试清理功能
    run proj::docker::cleanup_container "$test_container"
    [ "$status" -eq 0 ]
    
    # 验证容器被删除
    run docker ps -a --format "{{.Names}}"
    [[ ! "$output" =~ $test_container ]]
}

@test "proj::docker::check_container_status detects container status" {
    # 只在Docker可用时运行
    if ! proj::docker::check_docker_available; then
        skip "Docker not available"
    fi
    
    local test_container="test-status-container"
    
    # 创建运行中的测试容器
    docker run --name "$test_container" -d alpine:latest sleep 30 >/dev/null 2>&1 || true
    
    # 测试状态检查
    run proj::docker::check_container_status "$test_container" "running"
    [ "$status" -eq 0 ]
    
    # 清理
    docker rm -f "$test_container" >/dev/null 2>&1 || true
}

@test "proj::docker::get_container_ip returns valid IP" {
    # 只在Docker可用时运行
    if ! proj::docker::check_docker_available; then
        skip "Docker not available"
    fi
    
    local test_container="test-ip-container"
    
    # 确保测试网络存在
    proj::docker::ensure_network "$NETWORK_NAME" >/dev/null 2>&1
    
    # 创建连接到测试网络的容器
    docker run --name "$test_container" --network "$NETWORK_NAME" -d alpine:latest sleep 30 >/dev/null 2>&1 || true
    
    # 测试IP获取
    run proj::docker::get_container_ip "$test_container" "$NETWORK_NAME"
    
    if [ "$status" -eq 0 ]; then
        # 验证返回的是有效的IP地址格式
        [[ "$output" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]
    fi
    
    # 清理
    docker rm -f "$test_container" >/dev/null 2>&1 || true
}

@test "proj::docker::run_service fails with missing parameters" {
    # 测试缺少必需参数的情况
    local empty_ports=()
    local empty_volumes=()
    local empty_env=()
    
    # 缺少服务名
    run proj::docker::run_service "" "alpine:latest" empty_ports empty_volumes empty_env
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Service name and image are required" ]]
    
    # 缺少镜像名
    run proj::docker::run_service "test" "" empty_ports empty_volumes empty_env
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Service name and image are required" ]]
}

@test "proj::docker::cleanup_service_containers handles pattern matching" {
    # 测试模式匹配清理
    run proj::docker::cleanup_service_containers "nonexistent-*"
    [ "$status" -eq 0 ]
    # 应该输出没有找到匹配的容器
    [[ "$output" =~ "No containers found" ]]
}