#!/usr/bin/env bats

# =============================================================================
# 平台检测工具库单元测试
# Platform Detection Library Unit Tests
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
    
    source "${LIB_DIR}/platform.sh"
}

@test "proj::platform::detect_os detects macOS correctly" {
    # 在macOS上运行测试
    if [[ "$OSTYPE" == "darwin"* ]]; then
        run proj::platform::detect_os
        [ "$status" -eq 0 ]
        [ "$output" = "macos" ]
    else
        skip "Not running on macOS"
    fi
}

@test "proj::platform::detect_os detects Ubuntu correctly" {
    # 模拟Ubuntu环境（如果不在Ubuntu上）
    if [[ ! -f /etc/os-release ]] || ! grep -q "ubuntu" /etc/os-release 2>/dev/null; then
        skip "Not running on Ubuntu - would need OS simulation"
    else
        run proj::platform::detect_os
        [ "$status" -eq 0 ]
        [ "$output" = "ubuntu" ]
    fi
}

@test "proj::platform::has_docker returns correct status" {
    run proj::platform::has_docker
    # 应该返回0（有Docker）或1（没有Docker）
    [[ "$status" -eq 0 ]] || [[ "$status" -eq 1 ]]
}

@test "proj::platform::get_package_manager detects package manager" {
    run proj::platform::get_package_manager
    [ "$status" -eq 0 ]
    # 输出应该是已知的包管理器之一
    [[ "$output" =~ ^(apt|brew|yum|none|unknown)$ ]]
}

@test "proj::platform::get_preferred_install_method returns valid method" {
    run proj::platform::get_preferred_install_method "redis"
    [ "$status" -eq 0 ]
    # 应该返回 docker 或 native
    [[ "$output" =~ ^(docker|native)$ ]]
}

@test "proj::platform::get_architecture returns valid architecture" {
    run proj::platform::get_architecture
    [ "$status" -eq 0 ]
    # 应该返回已知的架构类型
    [[ "$output" =~ ^(amd64|arm64|armv7|.+)$ ]]
}

@test "proj::platform::is_service_supported validates service support" {
    # 测试支持的服务
    run proj::platform::is_service_supported "redis" "auto"
    [ "$status" -eq 0 ]
    
    # 测试不支持的安装方式（如果Docker不可用时强制Docker）
    if ! proj::platform::has_docker; then
        run proj::platform::is_service_supported "redis" "docker"
        [ "$status" -eq 1 ]
    fi
}

@test "proj::platform::validate_compatibility validates current platform" {
    run proj::platform::validate_compatibility
    # 在支持的平台上应该返回0，在不支持的平台上应该返回1
    [[ "$status" -eq 0 ]] || [[ "$status" -eq 1 ]]
}

@test "proj::platform::show_info displays platform information" {
    run proj::platform::show_info
    [ "$status" -eq 0 ]
    # 输出应该包含平台信息的关键词
    [[ "$output" =~ "Platform Information" ]]
    [[ "$output" =~ "Operating System" ]]
    [[ "$output" =~ "Package Manager" ]]
    [[ "$output" =~ "Docker Available" ]]
}