#!/usr/bin/env bash
#
# CFS日志写入测试脚本
# 测试OTEL Collector向腾讯云CFS写入日志的功能
#

set -o errexit
set -o nounset
set -o pipefail

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJ_ROOT_DIR="${SCRIPT_DIR}/../../.."

# 颜色输出
C_GREEN='\033[0;32m'
C_YELLOW='\033[1;33m'
C_RED='\033[0;31m'
C_BLUE='\033[0;34m'
C_NORMAL='\033[0m'

# 日志函数
log_info() {
    echo -e "${C_GREEN}[INFO]${C_NORMAL} $1"
}

log_warn() {
    echo -e "${C_YELLOW}[WARN]${C_NORMAL} $1"
}

log_error() {
    echo -e "${C_RED}[ERROR]${C_NORMAL} $1"
}

log_step() {
    echo -e "${C_BLUE}[STEP]${C_NORMAL} $1"
}

# CFS配置
CFS_MOUNT_POINT=${CFS_MOUNT_POINT:-"/mnt/cfs-logs"}
SERVICE_NAME=${PROJ_SERVICE_NAME:-"apiserver"}
TEST_LOG_DIR="${PROJ_ROOT_DIR}/logs/${SERVICE_NAME}"
TEST_ID="cfs-test-$(date +%s)"

# 创建测试日志目录
setup_test_environment() {
    log_step "Setting up test environment..."
    
    mkdir -p "${TEST_LOG_DIR}"
    
    # 检查CFS挂载状态
    if ! mountpoint -q "${CFS_MOUNT_POINT}" 2>/dev/null; then
        log_error "CFS is not mounted at ${CFS_MOUNT_POINT}"
        log_info "Please run: ./scripts/installation/otel-cfs-setup.sh mount"
        exit 1
    fi
    
    log_info "CFS mounted successfully at ${CFS_MOUNT_POINT}"
}

# 生成测试日志
generate_test_logs() {
    log_step "Generating test logs..."
    
    local test_file="${TEST_LOG_DIR}/cfs-test.log"
    local timestamp=$(date -Iseconds)
    
    # 生成不同类型的测试日志
    log_info "Writing JSON structured logs..."
    cat >> "${test_file}" << EOF
{"timestamp":"${timestamp}","level":"info","msg":"CFS integration test - JSON log","service":"${SERVICE_NAME}","test_id":"${TEST_ID}","log_type":"json","sequence":1}
{"timestamp":"${timestamp}","level":"warn","msg":"CFS integration test - Warning message","service":"${SERVICE_NAME}","test_id":"${TEST_ID}","log_type":"json","sequence":2}
{"timestamp":"${timestamp}","level":"error","msg":"CFS integration test - Error message","service":"${SERVICE_NAME}","test_id":"${TEST_ID}","log_type":"json","sequence":3,"error_code":"TEST_ERROR"}
{"timestamp":"${timestamp}","level":"debug","msg":"CFS integration test - Debug message with metadata","service":"${SERVICE_NAME}","test_id":"${TEST_ID}","log_type":"json","sequence":4,"metadata":{"user_id":"test_user","operation":"cfs_test","duration_ms":150}}
EOF
    
    log_info "Writing plain text logs..."
    cat >> "${test_file}" << EOF
[${timestamp}] INFO CFS integration test - Plain text log (test_id: ${TEST_ID})
[${timestamp}] WARN CFS integration test - Warning in plain text (test_id: ${TEST_ID})
[${timestamp}] ERROR CFS integration test - Error in plain text (test_id: ${TEST_ID})
EOF
    
    log_info "Generated $(wc -l < "${test_file}") test log entries"
}

# 等待日志被收集
wait_for_log_collection() {
    log_step "Waiting for log collection and processing..."
    
    local max_wait=30
    local wait_time=0
    
    while [[ ${wait_time} -lt ${max_wait} ]]; do
        echo -n "."
        sleep 1
        ((wait_time++))
    done
    echo ""
    
    log_info "Waited ${max_wait} seconds for log processing"
}

# 验证CFS中的日志文件
verify_cfs_logs() {
    log_step "Verifying logs in CFS..."
    
    local date_partition=$(date +%Y/%m/%d)
    local cfs_log_dir="${CFS_MOUNT_POINT}/logs/${SERVICE_NAME}/${date_partition}"
    
    log_info "Checking CFS log directory: ${cfs_log_dir}"
    
    if [[ ! -d "${cfs_log_dir}" ]]; then
        log_warn "CFS log directory does not exist: ${cfs_log_dir}"
        log_info "Available directories in CFS:"
        find "${CFS_MOUNT_POINT}" -type d -name "*${SERVICE_NAME}*" 2>/dev/null || log_warn "No service directories found"
        return 1
    fi
    
    # 查找日志文件
    local log_files
    log_files=$(find "${cfs_log_dir}" -name "*.jsonl" -type f 2>/dev/null)
    
    if [[ -z "${log_files}" ]]; then
        log_warn "No .jsonl log files found in ${cfs_log_dir}"
        log_info "Available files:"
        ls -la "${cfs_log_dir}" 2>/dev/null || log_warn "Directory is empty or inaccessible"
        return 1
    fi
    
    log_info "Found log files in CFS:"
    echo "${log_files}" | while read -r file; do
        if [[ -n "${file}" ]]; then
            local file_size=$(stat -c%s "${file}" 2>/dev/null || echo "0")
            local file_lines=$(wc -l < "${file}" 2>/dev/null || echo "0")
            log_info "  📄 ${file} (${file_size} bytes, ${file_lines} lines)"
        fi
    done
    
    # 搜索测试ID
    log_info "Searching for test logs with ID: ${TEST_ID}"
    local found_logs=0
    
    echo "${log_files}" | while read -r file; do
        if [[ -n "${file}" && -f "${file}" ]]; then
            local matches
            matches=$(grep -c "${TEST_ID}" "${file}" 2>/dev/null || echo "0")
            if [[ ${matches} -gt 0 ]]; then
                log_info "  ✅ Found ${matches} test log entries in ${file}"
                ((found_logs++))
                
                # 显示找到的日志条目
                log_info "  Sample entries:"
                grep "${TEST_ID}" "${file}" | head -3 | while IFS= read -r line; do
                    echo "    ${line:0:100}..."
                done
            fi
        fi
    done
    
    return 0
}

# 验证OTEL Collector状态
verify_otel_collector() {
    log_step "Verifying OTEL Collector status..."
    
    # 检查容器状态
    local container_name="proj-otel-cfs-collector"
    if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^${container_name}$"; then
        log_info "✅ OTEL Collector container is running"
    else
        log_error "❌ OTEL Collector container is not running"
        return 1
    fi
    
    # 健康检查
    if curl -s "http://127.0.0.1:13133/health" > /dev/null 2>&1; then
        log_info "✅ OTEL Collector health check passed"
    else
        log_error "❌ OTEL Collector health check failed"
        return 1
    fi
    
    # 检查指标
    local metrics_response
    if metrics_response=$(curl -s "http://127.0.0.1:8888/metrics" 2>/dev/null); then
        log_info "✅ OTEL Collector metrics endpoint accessible"
        
        # 检查文件导出器指标
        local file_exported=$(echo "${metrics_response}" | grep -c "otelcol_exporter_sent_log_records" || echo "0")
        if [[ ${file_exported} -gt 0 ]]; then
            log_info "  📊 File exporter metrics available"
        fi
    else
        log_warn "⚠️ Could not access OTEL Collector metrics"
    fi
    
    return 0
}

# 清理测试文件
cleanup_test_files() {
    log_step "Cleaning up test files..."
    
    local test_file="${TEST_LOG_DIR}/cfs-test.log"
    if [[ -f "${test_file}" ]]; then
        rm -f "${test_file}"
        log_info "Removed test log file: ${test_file}"
    fi
}

# 显示测试报告
show_test_report() {
    log_step "Test Report Summary"
    
    echo -e "${C_BLUE}═══════════════════════════════════════════${C_NORMAL}"
    echo -e "${C_BLUE}          CFS Integration Test Report${C_NORMAL}"
    echo -e "${C_BLUE}═══════════════════════════════════════════${C_NORMAL}"
    echo ""
    echo "Test ID: ${TEST_ID}"
    echo "Service: ${SERVICE_NAME}"
    echo "CFS Mount: ${CFS_MOUNT_POINT}"
    echo "Test Time: $(date)"
    echo ""
    
    # CFS状态
    if mountpoint -q "${CFS_MOUNT_POINT}" 2>/dev/null; then
        echo -e "CFS Mount Status: ${C_GREEN}✅ Mounted${C_NORMAL}"
        local cfs_space=$(df -h "${CFS_MOUNT_POINT}" 2>/dev/null | tail -1 | awk '{print "Used: " $3 " / " $2 " (" $5 ")"}' || echo "Unknown")
        echo "CFS Space: ${cfs_space}"
    else
        echo -e "CFS Mount Status: ${C_RED}❌ Not Mounted${C_NORMAL}"
    fi
    
    echo ""
    echo -e "${C_BLUE}Next Steps:${C_NORMAL}"
    echo "1. Check OTEL Collector logs: docker logs proj-otel-cfs-collector"
    echo "2. Monitor CFS directory: ls -la ${CFS_MOUNT_POINT}/logs/${SERVICE_NAME}/"
    echo "3. View real-time logs: tail -f ${CFS_MOUNT_POINT}/logs/${SERVICE_NAME}/*/app.jsonl"
    echo "4. Check metrics: curl http://127.0.0.1:8888/metrics | grep file"
    echo ""
}

# 主函数
main() {
    log_info "Starting CFS Integration Test (Test ID: ${TEST_ID})"
    
    # 执行测试步骤
    setup_test_environment
    
    if ! verify_otel_collector; then
        log_error "OTEL Collector verification failed. Please check the service status."
        exit 1
    fi
    
    generate_test_logs
    wait_for_log_collection
    
    # 验证结果
    local test_success=true
    if ! verify_cfs_logs; then
        test_success=false
        log_error "CFS log verification failed"
    fi
    
    cleanup_test_files
    show_test_report
    
    if [[ "${test_success}" == "true" ]]; then
        log_info "🎉 CFS Integration Test PASSED"
        exit 0
    else
        log_error "❌ CFS Integration Test FAILED"
        log_info "Check OTEL Collector logs for more details:"
        echo "  docker logs proj-otel-cfs-collector --tail 50"
        exit 1
    fi
}

# 处理命令行参数
case "${1:-test}" in
    test)
        main
        ;;
    setup)
        setup_test_environment
        ;;
    verify)
        verify_cfs_logs
        ;;
    report)
        show_test_report
        ;;
    *)
        echo "Usage: $0 {test|setup|verify|report}"
        echo ""
        echo "Commands:"
        echo "  test   - Run complete CFS integration test (default)"
        echo "  setup  - Setup test environment only"
        echo "  verify - Verify CFS logs only"  
        echo "  report - Show test report"
        exit 1
        ;;
esac