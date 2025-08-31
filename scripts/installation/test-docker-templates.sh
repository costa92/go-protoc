#!/usr/bin/env bash

# =============================================================================
# Docker模板系统测试脚本
# Docker Template System Test Script
#
# 测试新的Docker模板系统是否正常工作
# =============================================================================

set -eEuo pipefail

# 脚本配置
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJ_ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

# 测试结果统计
TESTS_TOTAL=0
TESTS_PASSED=0
TESTS_FAILED=0

# 测试函数
run_test() {
    local test_name="$1"
    local test_command="$2"
    
    ((TESTS_TOTAL++))
    log_info "Running test: $test_name"
    
    if eval "$test_command" >/dev/null 2>&1; then
        log_info "✅ Test passed: $test_name"
        ((TESTS_PASSED++))
        return 0
    else
        log_error "❌ Test failed: $test_name"
        ((TESTS_FAILED++))
        return 1
    fi
}

# 检查先决条件
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # 检查Docker
    if ! command -v docker >/dev/null 2>&1; then
        log_error "Docker not found. Please install Docker first."
        exit 1
    fi
    
    if ! docker info >/dev/null 2>&1; then
        log_error "Docker daemon not running. Please start Docker first."
        exit 1
    fi
    
    log_info "Prerequisites check passed"
}

# 加载库函数
load_libraries() {
    log_info "Loading project libraries..."
    
    # 设置PROJ_ROOT_DIR环境变量
    export PROJ_ROOT_DIR
    
    # 加载必要的库
    source "${PROJ_ROOT_DIR}/scripts/installation/versions.sh"
    source "${PROJ_ROOT_DIR}/scripts/installation/common.sh" 2>/dev/null || {
        log_warn "Could not load common.sh, continuing anyway"
    }
    source "${PROJ_ROOT_DIR}/scripts/installation/lib/common_lib.sh"
    
    log_info "Libraries loaded successfully"
}

# 测试模板列表功能
test_template_listing() {
    log_info "Testing template listing functionality..."
    
    run_test "List Docker templates" \
        "proj::template::list_templates docker"
    
    run_test "List Redis templates" \
        "proj::template::list_templates docker redis"
}

# 测试脚本生成功能
test_script_generation() {
    log_info "Testing script generation functionality..."
    
    local test_dir="/tmp/docker-template-test"
    mkdir -p "$test_dir"
    
    run_test "Generate Redis Docker scripts" \
        "proj::docker::generate_service_scripts redis $test_dir"
    
    run_test "Check generated docker-run.sh exists" \
        "test -f $test_dir/docker-run.sh"
    
    run_test "Check generated docker-stop.sh exists" \
        "test -f $test_dir/docker-stop.sh"
    
    run_test "Check generated docker-status.sh exists" \
        "test -f $test_dir/docker-status.sh"
    
    run_test "Check scripts are executable" \
        "test -x $test_dir/docker-run.sh && test -x $test_dir/docker-stop.sh"
    
    # 清理测试目录
    rm -rf "$test_dir"
}

# 测试网络创建功能
test_network_management() {
    log_info "Testing Docker network management..."
    
    # 删除网络（如果存在）
    docker network rm proj-network 2>/dev/null || true
    
    run_test "Create project network" \
        "proj::docker::ensure_project_network"
    
    run_test "Verify network exists" \
        "docker network ls | grep -q proj-network"
    
    # 测试重复创建（应该不会失败）
    run_test "Create network again (should not fail)" \
        "proj::docker::ensure_project_network"
}

# 测试配置变量
test_configuration_variables() {
    log_info "Testing configuration variables..."
    
    run_test "Check REDIS_VERSION is set" \
        "test -n \"\$REDIS_VERSION\""
    
    run_test "Check PROJ_REDIS_PORT is set" \
        "test -n \"\$PROJ_REDIS_PORT\""
    
    run_test "Check template directory exists" \
        "test -d \"\$PROJ_ROOT_DIR/scripts/installation/templates/docker\""
}

# 测试模板验证功能
test_template_validation() {
    log_info "Testing template validation..."
    
    local redis_template="${PROJ_ROOT_DIR}/scripts/installation/templates/docker/redis/docker-run.sh.tpl"
    
    run_test "Validate Redis docker-run template" \
        "proj::template::validate_template \"$redis_template\""
}

# 主测试函数
main() {
    echo "=============================================="
    echo "Docker Template System Test Suite"
    echo "=============================================="
    echo ""
    
    # 检查先决条件
    check_prerequisites
    
    # 加载库函数
    load_libraries
    
    # 运行测试
    test_configuration_variables
    test_template_listing
    test_template_validation
    test_script_generation
    test_network_management
    
    # 显示测试结果
    echo ""
    echo "=============================================="
    echo "Test Results Summary"
    echo "=============================================="
    echo "Total tests: $TESTS_TOTAL"
    echo "Passed: $TESTS_PASSED"
    echo "Failed: $TESTS_FAILED"
    
    if [[ $TESTS_FAILED -eq 0 ]]; then
        log_info "🎉 All tests passed! Docker template system is working correctly."
        exit 0
    else
        log_error "💥 Some tests failed. Please check the output above."
        exit 1
    fi
}

# 运行主函数
main "$@"