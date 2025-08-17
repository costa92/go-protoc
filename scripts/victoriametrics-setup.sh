#!/bin/bash

# VictoriaMetrics 完整设置脚本
# 包含 VictoriaMetrics, VictoriaLogs, vmagent, Fluent Bit

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# 检查Docker和Docker Compose
check_docker() {
    log_step "检查 Docker 环境..."

    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi

    if ! docker info &> /dev/null; then
        log_error "Docker 服务未运行，请启动 Docker"
        exit 1
    fi

    log_info "Docker 环境检查通过"
}

# 停止现有的 VictoriaLogs 容器（避免端口冲突）
stop_existing_services() {
    log_step "停止现有的相关服务..."

    # 停止可能存在的 VictoriaLogs 容器
    if docker ps -q -f name=proj-victorialogs | grep -q .; then
        log_warn "发现运行中的 VictoriaLogs 容器，正在停止..."
        docker stop proj-victorialogs || true
        docker rm proj-victorialogs || true
    fi

    # 停止其他可能冲突的容器
    containers=("proj-victoriametrics" "proj-vmagent" "proj-fluent-bit")
    for container in "${containers[@]}"; do
        if docker ps -q -f name=$container | grep -q .; then
            log_warn "停止容器: $container"
            docker stop $container || true
            docker rm $container || true
        fi
    done
}

# 启动 VictoriaMetrics 套件
start_victoria_suite() {
    local action=${1:-up}

    log_step "启动 VictoriaMetrics 完整套件..."

    cd deployments/victoriametrics

    case $action in
        "up")
            docker-compose up -d
            ;;
        "stop")
            docker-compose stop
            ;;
        "down")
            docker-compose down
            ;;
        "restart")
            docker-compose restart
            ;;
        "logs")
            docker-compose logs -f
            ;;
        "status")
            docker-compose ps
            ;;
        *)
            log_error "未知操作: $action"
            exit 1
            ;;
    esac

    cd - > /dev/null
}

# 检查服务健康状态
check_health() {
    log_step "检查服务健康状态..."

    services=(
        "VictoriaMetrics:http://localhost:8428/health"
        "VictoriaLogs:http://localhost:9428/health"
        "vmagent:http://localhost:8429/health"
    )

    for service_info in "${services[@]}"; do
        IFS=':' read -r service_name service_url <<< "$service_info"

        echo -n "检查 $service_name... "
        if curl -s "$service_url" > /dev/null 2>&1; then
            echo -e "${GREEN}✓${NC}"
        else
            echo -e "${RED}✗${NC}"
            log_warn "$service_name 可能未正常启动"
        fi
    done
}

# 显示访问信息
show_access_info() {
    log_step "服务访问信息"

    cat << EOF

🎯 VictoriaMetrics 套件已启动!

📊 访问地址:
┌─────────────────────────────────────────────────────────────────┐
│ VictoriaMetrics UI:  http://localhost:8428                     │
│ VictoriaLogs UI:     http://localhost:9428/select/vmui         │
│ vmagent UI:          http://localhost:8429                     │
│ Fluent Bit Status:   http://localhost:2020                     │
└─────────────────────────────────────────────────────────────────┘

🔧 管理命令:
- 查看状态: make victoria-status
- 查看日志: make victoria-logs
- 重启服务: make victoria-restart
- 停止服务: make victoria-stop

📈 数据采集:
- 应用指标: 自动从 localhost:8080/metrics 采集
- 应用日志: 通过修改后的 VictoriaLogs Writer 写入
- 系统指标: 通过 vmagent 采集

EOF
}

# 主函数
main() {
    local action=${1:-setup}

    case $action in
        "setup"|"start"|"up")
            log_info "开始设置 VictoriaMetrics 套件..."
            check_docker
            stop_existing_services
            start_victoria_suite "up"
            sleep 10  # 等待服务启动
            check_health
            show_access_info
            ;;
        "stop")
            start_victoria_suite "stop"
            log_info "VictoriaMetrics 套件已停止"
            ;;
        "down"|"clean")
            start_victoria_suite "down"
            log_info "VictoriaMetrics 套件已清理"
            ;;
        "restart")
            start_victoria_suite "restart"
            log_info "VictoriaMetrics 套件已重启"
            ;;
        "logs")
            start_victoria_suite "logs"
            ;;
        "status")
            start_victoria_suite "status"
            check_health
            ;;
        "health")
            check_health
            ;;
        *)
            echo "用法: $0 {setup|start|stop|restart|logs|status|health|clean}"
            echo ""
            echo "命令说明:"
            echo "  setup/start  - 启动完整的 VictoriaMetrics 套件"
            echo "  stop         - 停止服务但保留容器"
            echo "  restart      - 重启服务"
            echo "  logs         - 查看服务日志"
            echo "  status       - 查看服务状态"
            echo "  health       - 检查服务健康状态"
            echo "  clean/down   - 停止并删除容器"
            exit 1
            ;;
    esac
}

main "$@"