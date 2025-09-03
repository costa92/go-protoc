#!/usr/bin/env bash
# =============================================================================
# Jaeger Docker 运行脚本
# 由模板系统自动生成，请勿手动修改
# =============================================================================

set -euo pipefail

# 加载环境配置
# 将 development/test/production 映射为 dev/test/prod
case "${PROJ_ENVIRONMENT}" in
    "development")
        ENV_FILE="env.dev"
        ;;
    "test"|"testing")
        ENV_FILE="env.test"
        ;;
    "production")
        ENV_FILE="env.prod"
        ;;
    *)
        ENV_FILE="env.${PROJ_ENVIRONMENT:-dev}"
        ;;
esac

source "${PROJ_ROOT_DIR}/manifests/env/${ENV_FILE}"

# =============================================================================
# 服务配置
# =============================================================================

SERVICE_NAME="jaeger"
CONTAINER_NAME="${PROJ_PREFIX}-${SERVICE_NAME}"
IMAGE_NAME="jaegertracing/all-in-one:${JAEGER_VERSION}"
NETWORK_NAME="${PROJ_PREFIX}-${PROJ_ENV_SHORT}-network"

# 端口配置 - 支持多环境
COLLECTOR_PORT="${PROJ_ACCESS_PORT_PREFIX}14268"    # Collector HTTP endpoint
UI_PORT="${PROJ_ACCESS_PORT_PREFIX}16686"           # Web UI
AGENT_UDP_PORT="${PROJ_ACCESS_PORT_PREFIX}6831"     # Agent UDP compact
AGENT_HTTP_PORT="${PROJ_ACCESS_PORT_PREFIX}6832"    # Agent HTTP compact
QUERY_PORT="${PROJ_ACCESS_PORT_PREFIX}16687"        # Query service
ADMIN_PORT="${PROJ_ACCESS_PORT_PREFIX}14269"        # Admin port

# 数据目录
DATA_DIR="${PROJ_THIRDPARTY_DIR}/${SERVICE_NAME}/data"

# =============================================================================
# 辅助函数
# =============================================================================

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 检查容器是否存在
container_exists() {
    docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"
}

# 检查容器是否运行中
is_running() {
    docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"
}

# =============================================================================
# 主要操作函数
# =============================================================================

# 启动服务
start_service() {
    log_info "Starting ${SERVICE_NAME}..."
    
    # 确保网络存在
    if ! docker network ls | grep -q "${NETWORK_NAME}"; then
        log_info "Creating network ${NETWORK_NAME}..."
        docker network create "${NETWORK_NAME}"
    fi
    
    # 创建数据目录
    mkdir -p "${DATA_DIR}"
    
    # 如果容器已存在
    if container_exists; then
        if is_running; then
            log_warning "${SERVICE_NAME} is already running"
            return 0
        else
            log_info "Starting existing container..."
            docker start "${CONTAINER_NAME}"
            return $?
        fi
    fi
    
    # 运行新容器
    log_info "Creating and starting new ${SERVICE_NAME} container..."
    docker run -d \
        --name "${CONTAINER_NAME}" \
        --network "${NETWORK_NAME}" \
        --restart unless-stopped \
        -p "${COLLECTOR_PORT}:14268" \
        -p "${UI_PORT}:16686" \
        -p "${AGENT_UDP_PORT}:6831/udp" \
        -p "${AGENT_HTTP_PORT}:6832/udp" \
        -p "${QUERY_PORT}:16687" \
        -p "${ADMIN_PORT}:14269" \
        -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
        -e QUERY_BASE_PATH=/ \
        -e SPAN_STORAGE_TYPE=badger \
        -e BADGER_EPHEMERAL=false \
        -e BADGER_DIRECTORY_VALUE=/badger/data \
        -e BADGER_DIRECTORY_KEY=/badger/key \
        -v "${DATA_DIR}:/badger" \
        --label "project=${PROJ_NAME}" \
        --label "service=${SERVICE_NAME}" \
        --label "environment=${PROJ_ENVIRONMENT}" \
        "${IMAGE_NAME}"
    
    # 等待服务就绪
    log_info "Waiting for ${SERVICE_NAME} to be ready..."
    sleep 3
    
    # 检查服务状态
    if is_running; then
        log_info "✅ ${SERVICE_NAME} started successfully!"
        log_info "📊 Jaeger UI: http://${PROJ_ACCESS_HOST}:${UI_PORT}"
        log_info "📡 Collector endpoint: http://${PROJ_ACCESS_HOST}:${COLLECTOR_PORT}/api/traces"
        log_info "🔌 Agent UDP: ${PROJ_ACCESS_HOST}:${AGENT_UDP_PORT}"
        return 0
    else
        log_error "${SERVICE_NAME} failed to start"
        docker logs "${CONTAINER_NAME}" --tail 50
        return 1
    fi
}

# 停止服务
stop_service() {
    log_info "Stopping ${SERVICE_NAME}..."
    
    if ! container_exists; then
        log_warning "${SERVICE_NAME} container does not exist"
        return 0
    fi
    
    if is_running; then
        docker stop "${CONTAINER_NAME}"
        log_info "✅ ${SERVICE_NAME} stopped"
    else
        log_warning "${SERVICE_NAME} is not running"
    fi
}

# 重启服务
restart_service() {
    log_info "Restarting ${SERVICE_NAME}..."
    stop_service
    sleep 2
    start_service
}

# 查看状态
status_service() {
    if ! container_exists; then
        log_info "${SERVICE_NAME} container does not exist"
        return 1
    fi
    
    if is_running; then
        log_info "✅ ${SERVICE_NAME} is running"
        echo ""
        docker ps --filter "name=${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
        echo ""
        
        # 显示网络信息
        local network_info
        network_info=$(docker inspect ${CONTAINER_NAME} --format '{{range $net, $conf := .NetworkSettings.Networks}}{{$net}} {{end}}' 2>/dev/null || echo "Unknown")
        log_info "Network: ${network_info}"
        echo ""
        
        log_info "Service endpoints:"
        log_info "  • UI: http://${PROJ_ACCESS_HOST}:${UI_PORT}"
        log_info "  • Collector: http://${PROJ_ACCESS_HOST}:${COLLECTOR_PORT}"
        log_info "  • Agent UDP: ${PROJ_ACCESS_HOST}:${AGENT_UDP_PORT}"
        log_info "  • Query API: http://${PROJ_ACCESS_HOST}:${QUERY_PORT}"
        log_info "  • Admin: http://${PROJ_ACCESS_HOST}:${ADMIN_PORT}"
    else
        log_warning "⚠️  ${SERVICE_NAME} container exists but is not running"
        docker ps -a --filter "name=${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}"
    fi
}

# 查看日志
logs_service() {
    if ! container_exists; then
        log_error "${SERVICE_NAME} container does not exist"
        return 1
    fi
    
    docker logs "${CONTAINER_NAME}" "$@"
}

# 清理服务
cleanup_service() {
    log_info "Cleaning up ${SERVICE_NAME}..."
    
    if container_exists; then
        if is_running; then
            docker stop "${CONTAINER_NAME}"
        fi
        docker rm "${CONTAINER_NAME}"
        log_info "✅ ${SERVICE_NAME} container removed"
    fi
    
    # 可选：清理数据
    if [[ "${CLEANUP_DATA:-false}" == "true" ]]; then
        log_warning "Removing ${SERVICE_NAME} data directory: ${DATA_DIR}"
        rm -rf "${DATA_DIR}"
    fi
}

# =============================================================================
# 主程序
# =============================================================================

main() {
    local action="${1:-start}"
    
    case "$action" in
        start)
            start_service
            ;;
        stop)
            stop_service
            ;;
        restart)
            restart_service
            ;;
        status)
            status_service
            ;;
        logs)
            shift
            logs_service "$@"
            ;;
        cleanup)
            cleanup_service
            ;;
        *)
            log_error "Unknown action: $action"
            echo "Usage: $0 {start|stop|restart|status|logs|cleanup}"
            exit 1
            ;;
    esac
}

# 执行主程序
main "$@"