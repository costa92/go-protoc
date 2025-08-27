#!/usr/bin/env bash
#
# OpenTelemetry Collector with Tencent Cloud CFS Integration Setup
# This script configures OTEL Collector to write logs to Tencent Cloud CFS
#

set -o errexit
set -o nounset  
set -o pipefail

# 加载通用配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJ_ROOT_DIR="${SCRIPT_DIR}/../.."
source "${SCRIPT_DIR}/common.sh"

# Tencent Cloud CFS Configuration
# These should be set as environment variables or in a config file
CFS_FILESYSTEM_ID=${CFS_FILESYSTEM_ID:-""}
CFS_IP=${CFS_IP:-""}                           # CFS Mount Target IP
CFS_MOUNT_POINT=${CFS_MOUNT_POINT:-"/mnt/cfs-logs"}
TENCENT_REGION=${TENCENT_REGION:-"ap-beijing"}
CFS_INSTANCE_ID=${CFS_INSTANCE_ID:-"otel-$(hostname)-$(date +%s)"}

# OTEL Configuration
OTEL_CFS_DOCKER_NAME=${NETWORK_NAME}-otel-cfs-collector
OTEL_CFS_CONFIG_DIR="${PROJ_THIRDPARTY_INSTALL_DIR}/otel-cfs/config"

# Function to validate CFS configuration
proj::otel::cfs::validate_config() {
    proj::log::info "Validating Tencent Cloud CFS configuration..."
    
    if [[ -z "${CFS_FILESYSTEM_ID}" ]]; then
        proj::log::error "CFS_FILESYSTEM_ID is required. Please set it as an environment variable."
        proj::log::info "Example: export CFS_FILESYSTEM_ID='cfs-xxxxxxxx'"
        return 1
    fi
    
    if [[ -z "${CFS_IP}" ]]; then
        proj::log::error "CFS_IP (Mount Target IP) is required. Please set it as an environment variable."
        proj::log::info "Example: export CFS_IP='10.0.0.100'"
        return 1
    fi
    
    proj::log::info "✅ CFS configuration validated"
    proj::log::info "   Filesystem ID: ${CFS_FILESYSTEM_ID}"
    proj::log::info "   Mount Target IP: ${CFS_IP}"
    proj::log::info "   Mount Point: ${CFS_MOUNT_POINT}"
    proj::log::info "   Region: ${TENCENT_REGION}"
}

# Function to setup CFS mount point
proj::otel::cfs::setup_mount() {
    proj::log::info "Setting up CFS mount point..."
    
    # 检查是否已经挂载
    if mountpoint -q "${CFS_MOUNT_POINT}" 2>/dev/null; then
        proj::log::info "CFS is already mounted at ${CFS_MOUNT_POINT}"
        return 0
    fi
    
    # 创建挂载点目录
    proj::util::sudo "mkdir -p ${CFS_MOUNT_POINT}"
    
    # 安装NFS工具（如果需要）
    if ! command -v mount.nfs4 > /dev/null 2>&1; then
        proj::log::info "Installing NFS utilities..."
        if command -v apt-get > /dev/null 2>&1; then
            proj::util::sudo "apt-get update && apt-get install -y nfs-common"
        elif command -v yum > /dev/null 2>&1; then
            proj::util::sudo "yum install -y nfs-utils"
        else
            proj::log::error "Cannot install NFS utilities. Please install manually."
            return 1
        fi
    fi
    
    # 挂载CFS
    proj::log::info "Mounting CFS filesystem..."
    proj::util::sudo "mount -t nfs -o vers=4.0,proto=tcp,fsc ${CFS_IP}:/ ${CFS_MOUNT_POINT}"
    
    # 验证挂载
    if mountpoint -q "${CFS_MOUNT_POINT}"; then
        proj::log::info "✅ CFS mounted successfully at ${CFS_MOUNT_POINT}"
        
        # 创建日志目录结构
        proj::util::sudo "mkdir -p ${CFS_MOUNT_POINT}/logs/${PROJ_SERVICE_NAME:-apiserver}"
        proj::util::sudo "mkdir -p ${CFS_MOUNT_POINT}/otel-logs"
        
        # 设置权限
        proj::util::sudo "chmod 755 ${CFS_MOUNT_POINT}/logs"
        proj::util::sudo "chmod 755 ${CFS_MOUNT_POINT}/otel-logs"
        
        proj::log::info "✅ CFS directory structure created"
    else
        proj::log::error "❌ Failed to mount CFS filesystem"
        return 1
    fi
}

# Function to create CFS fstab entry for persistence
proj::otel::cfs::setup_fstab() {
    proj::log::info "Setting up CFS fstab entry for automatic mounting..."
    
    local fstab_entry="${CFS_IP}:/ ${CFS_MOUNT_POINT} nfs vers=4.0,proto=tcp,fsc 0 0"
    
    # 检查是否已经存在条目
    if grep -q "${CFS_MOUNT_POINT}" /etc/fstab 2>/dev/null; then
        proj::log::info "CFS fstab entry already exists"
    else
        # 备份fstab
        proj::util::sudo "cp /etc/fstab /etc/fstab.backup.$(date +%Y%m%d-%H%M%S)"
        
        # 添加CFS条目
        echo "${fstab_entry}" | proj::util::sudo "tee -a /etc/fstab"
        proj::log::info "✅ CFS fstab entry added for automatic mounting"
    fi
}

# Function to generate CFS-specific OTEL configuration
proj::otel::cfs::generate_config() {
    proj::log::info "Generating CFS-specific OTEL Collector configuration..."
    
    # 创建配置目录
    mkdir -p "${OTEL_CFS_CONFIG_DIR}"
    
    # 设置环境变量用于模板替换
    export TENCENT_REGION="${TENCENT_REGION}"
    export CFS_MOUNT_POINT="${CFS_MOUNT_POINT}"
    export CFS_FILESYSTEM_ID="${CFS_FILESYSTEM_ID}"
    export CFS_DATE_PARTITION="$(date +%Y/%m/%d)"
    export CFS_INSTANCE_ID="${CFS_INSTANCE_ID}"
    export PROJ_ENVIRONMENT="${PROJ_ENVIRONMENT:-development}"
    export PROJ_SERVICE_NAMESPACE="${PROJ_SERVICE_NAMESPACE:-default}"
    export PROJ_SERVICE_NAME="${PROJ_SERVICE_NAME:-apiserver}"
    export PROJ_OTELCOL_VERSION="${OTEL_VERSION}"
    
    # 使用envsubst生成配置文件
    envsubst < "${SCRIPT_DIR}/otel-collector/config-tencent-cfs.yaml" > "${OTEL_CFS_CONFIG_DIR}/config.yaml"
    
    proj::log::info "✅ CFS OTEL configuration generated at: ${OTEL_CFS_CONFIG_DIR}/config.yaml"
}

# Function to install OTEL Collector with CFS support
proj::otel::cfs::install() {
    proj::log::info "Installing OTEL Collector with Tencent Cloud CFS support..."
    
    # 验证配置
    proj::otel::cfs::validate_config
    
    # 设置CFS挂载
    proj::otel::cfs::setup_mount
    
    # 设置fstab条目
    proj::otel::cfs::setup_fstab
    
    # 生成配置
    proj::otel::cfs::generate_config
    
    # 清理现有容器
    proj::common::docker::cleanup_container "${OTEL_CFS_DOCKER_NAME}"
    
    # 启动OTEL Collector容器，支持CFS挂载
    proj::log::info "Starting OTEL Collector with CFS integration..."
    
    docker run -d \
        --name "${OTEL_CFS_DOCKER_NAME}" \
        --network "${NETWORK_NAME}" \
        -p 4317:4317 \
        -p 4318:4318 \
        -p 13133:13133 \
        -p 8888:8888 \
        -p 8889:8889 \
        -v "${OTEL_CFS_CONFIG_DIR}/config.yaml:/etc/otelcol-contrib/config.yaml" \
        -v "${CFS_MOUNT_POINT}:${CFS_MOUNT_POINT}" \
        -v "${PROJ_ROOT_DIR}/logs:/opt/logs:ro" \
        -v "/tmp/otel-buffer:/tmp/otel-buffer" \
        -e TENCENT_REGION="${TENCENT_REGION}" \
        -e CFS_MOUNT_POINT="${CFS_MOUNT_POINT}" \
        -e CFS_FILESYSTEM_ID="${CFS_FILESYSTEM_ID}" \
        -e CFS_DATE_PARTITION="$(date +%Y/%m/%d)" \
        -e CFS_INSTANCE_ID="${CFS_INSTANCE_ID}" \
        -e PROJ_ENVIRONMENT="${PROJ_ENVIRONMENT:-development}" \
        -e PROJ_SERVICE_NAMESPACE="${PROJ_SERVICE_NAMESPACE:-default}" \
        -e PROJ_SERVICE_NAME="${PROJ_SERVICE_NAME:-apiserver}" \
        -e PROJ_OTELCOL_VERSION="${OTEL_VERSION}" \
        --restart unless-stopped \
        otel/opentelemetry-collector-contrib:${OTEL_VERSION} \
        --config=/etc/otelcol-contrib/config.yaml
    
    # 等待容器启动
    proj::log::info "Waiting for OTEL Collector to start..."
    sleep 8
    
    # 健康检查
    if ! curl -s "http://127.0.0.1:13133/health" > /dev/null; then
        proj::log::error "OTEL Collector failed to start. Checking logs..."
        docker logs "${OTEL_CFS_DOCKER_NAME}"
        return 1
    fi
    
    proj::otel::cfs::info
    proj::log::info "✅ OTEL Collector with CFS support installed successfully"
}

# Function to test CFS log writing
proj::otel::cfs::test() {
    proj::log::info "Testing log writing to CFS..."
    
    # 生成测试日志
    local test_log_file="${PROJ_ROOT_DIR}/logs/apiserver/cfs-test.log"
    mkdir -p "$(dirname "${test_log_file}")"
    
    # 写入测试日志条目
    echo "{\"timestamp\":\"$(date -Iseconds)\",\"level\":\"info\",\"msg\":\"CFS test log entry\",\"service\":\"cfs-test\",\"test_id\":\"$(uuidgen 2>/dev/null || date +%s)\"}" >> "${test_log_file}"
    
    # 等待日志被收集
    sleep 5
    
    # 检查CFS中是否有日志文件
    local cfs_logs_dir="${CFS_MOUNT_POINT}/logs/${PROJ_SERVICE_NAME:-apiserver}/$(date +%Y/%m/%d)"
    
    if [[ -d "${cfs_logs_dir}" ]] && [[ $(find "${cfs_logs_dir}" -name "*.jsonl" -type f | wc -l) -gt 0 ]]; then
        proj::log::info "✅ Test logs successfully written to CFS"
        proj::log::info "   CFS log directory: ${cfs_logs_dir}"
        proj::log::info "   Log files:"
        find "${cfs_logs_dir}" -name "*.jsonl" -type f -exec ls -la {} \;
    else
        proj::log::warn "❌ No log files found in CFS. Check OTEL Collector logs:"
        docker logs --tail 50 "${OTEL_CFS_DOCKER_NAME}"
    fi
    
    # 清理测试日志文件
    rm -f "${test_log_file}"
}

# Function to display CFS integration information
proj::otel::cfs::info() {
    echo -e "${C_GREEN}OTEL Collector with Tencent Cloud CFS Integration:${C_NORMAL}"
    echo "  CFS Mount Point: ${CFS_MOUNT_POINT}"
    echo "  CFS Filesystem ID: ${CFS_FILESYSTEM_ID}"
    echo "  CFS Region: ${TENCENT_REGION}"
    echo "  Instance ID: ${CFS_INSTANCE_ID}"
    echo "  OTLP HTTP Endpoint: http://127.0.0.1:4318"
    echo "  OTLP gRPC Endpoint: 127.0.0.1:4317"
    echo "  Health Check: http://127.0.0.1:13133/health"
    echo "  Metrics Endpoint: http://127.0.0.1:8888/metrics"
    echo "  CFS Logs Directory: ${CFS_MOUNT_POINT}/logs/${PROJ_SERVICE_NAME:-apiserver}"
    echo ""
    echo "  Container Name: ${OTEL_CFS_DOCKER_NAME}"
    echo "  Config Directory: ${OTEL_CFS_CONFIG_DIR}"
    echo ""
    echo "${C_GREEN}Usage:${C_NORMAL}"
    echo "  Test CFS logging: $0 test"
    echo "  Check status: $0 status"
    echo "  View logs: docker logs ${OTEL_CFS_DOCKER_NAME}"
    echo "  Check CFS mount: mountpoint ${CFS_MOUNT_POINT}"
}

# Function to check status
proj::otel::cfs::status() {
    echo -e "${C_GREEN}CFS Integration Status:${C_NORMAL}"
    
    # 检查CFS挂载状态
    if mountpoint -q "${CFS_MOUNT_POINT}" 2>/dev/null; then
        echo "  ✅ CFS mounted at ${CFS_MOUNT_POINT}"
        echo "     Available space: $(df -h "${CFS_MOUNT_POINT}" | tail -1 | awk '{print $4}')"
    else
        echo "  ❌ CFS not mounted at ${CFS_MOUNT_POINT}"
    fi
    
    # 检查容器状态
    if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^${OTEL_CFS_DOCKER_NAME}$"; then
        echo "  ✅ OTEL Collector container running"
        echo "     Container ID: $(docker ps -f name=${OTEL_CFS_DOCKER_NAME} --format '{{.ID}}')"
    else
        echo "  ❌ OTEL Collector container not running"
    fi
    
    # 健康检查
    if curl -s "http://127.0.0.1:13133/health" > /dev/null 2>&1; then
        echo "  ✅ OTEL Collector health check passed"
    else
        echo "  ❌ OTEL Collector health check failed"
    fi
}

# Function to uninstall CFS integration
proj::otel::cfs::uninstall() {
    proj::log::info "Uninstalling OTEL Collector CFS integration..."
    
    # 停止容器
    proj::common::docker::cleanup_container "${OTEL_CFS_DOCKER_NAME}"
    
    # 卸载CFS
    if mountpoint -q "${CFS_MOUNT_POINT}" 2>/dev/null; then
        proj::util::sudo "umount ${CFS_MOUNT_POINT}"
        proj::log::info "CFS unmounted from ${CFS_MOUNT_POINT}"
    fi
    
    # 清理配置
    if [[ -d "${OTEL_CFS_CONFIG_DIR}" ]]; then
        rm -rf "${OTEL_CFS_CONFIG_DIR}"
        proj::log::info "Configuration directory cleaned up"
    fi
    
    proj::log::info "✅ CFS integration uninstalled successfully"
}

# 处理命令行参数
if [[ $# -gt 0 ]]; then
    case $1 in
        install)
            proj::otel::cfs::install
            ;;
        test)
            proj::otel::cfs::test
            ;;
        status)
            proj::otel::cfs::status
            ;;
        info)
            proj::otel::cfs::info
            ;;
        uninstall)
            proj::otel::cfs::uninstall
            ;;
        mount)
            proj::otel::cfs::setup_mount
            ;;
        *)
            proj::log::error "Unknown command: $1"
            echo "Usage: $0 {install|test|status|info|uninstall|mount}"
            echo ""
            echo "Environment Variables Required:"
            echo "  CFS_FILESYSTEM_ID  - Tencent Cloud CFS filesystem ID"
            echo "  CFS_IP            - CFS mount target IP address"
            echo "  CFS_MOUNT_POINT   - Local mount point (default: /mnt/cfs-logs)"
            echo "  TENCENT_REGION    - Tencent Cloud region (default: ap-beijing)"
            exit 1
            ;;
    esac
else
    proj::otel::cfs::info
fi