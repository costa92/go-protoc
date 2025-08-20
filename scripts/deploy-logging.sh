#!/bin/bash
# 统一日志收集系统部署脚本
# 支持本地开发、Docker 和 Kubernetes 环境

set -euo pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
DEPLOYMENT_TYPE=""
NAMESPACE="logging"
APP_NAMESPACE="app"

# 帮助函数
usage() {
    cat << EOF
统一日志收集系统部署脚本

USAGE:
    $0 COMMAND [OPTIONS]

COMMANDS:
    local       部署本地开发环境
    docker      部署 Docker 环境
    k8s         部署 Kubernetes 环境
    clean       清理指定环境的部署
    status      检查部署状态
    test        测试日志收集功能

OPTIONS:
    -h, --help     显示帮助信息
    -v, --verbose  详细输出
    -n, --namespace NAMESPACE  指定 Kubernetes 命名空间 (默认: logging)

EXAMPLES:
    # 本地开发环境
    $0 local
    
    # Docker 环境
    $0 docker
    
    # Kubernetes 环境
    $0 k8s
    
    # 清理 Docker 环境
    $0 clean docker
    
    # 检查状态
    $0 status k8s

EOF
}

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

log_debug() {
    if [[ "${VERBOSE:-false}" == "true" ]]; then
        echo -e "${BLUE}[DEBUG]${NC} $1"
    fi
}

# 检查依赖
check_dependencies() {
    local env_type="$1"
    
    case "$env_type" in
        "local")
            command -v docker >/dev/null 2>&1 || { log_error "Docker 未安装"; exit 1; }
            ;;
        "docker")
            command -v docker >/dev/null 2>&1 || { log_error "Docker 未安装"; exit 1; }
            command -v docker-compose >/dev/null 2>&1 || { log_error "Docker Compose 未安装"; exit 1; }
            ;;
        "k8s")
            command -v kubectl >/dev/null 2>&1 || { log_error "kubectl 未安装"; exit 1; }
            kubectl cluster-info >/dev/null 2>&1 || { log_error "无法连接到 Kubernetes 集群"; exit 1; }
            ;;
    esac
}

# 本地开发环境部署
deploy_local() {
    log_info "部署本地开发环境..."
    
    # 启动基础服务
    cd "$PROJECT_ROOT"
    
    log_info "启动 VictoriaLogs..."
    ./scripts/installation/service.sh start victorialogs
    
    log_info "启动 OpenTelemetry Collector..."
    ./scripts/installation/service.sh start otelcol
    
    log_info "检查服务状态..."
    ./scripts/installation/service.sh status victorialogs
    ./scripts/installation/service.sh status otelcol
    
    log_info "本地开发环境部署完成！"
    cat << EOF

🎉 本地环境访问信息:
  VictoriaLogs UI:    http://localhost:9428/select/vmui/
  OTLP HTTP 端点:     http://localhost:4318
  OTLP gRPC 端点:     localhost:4317
  健康检查:           http://localhost:13133

📝 使用方法:
  1. 运行应用程序: make run-api
  2. 查询日志: curl "http://localhost:9428/select/logsql/query" -d 'query=*'

EOF
}

# Docker 环境部署
deploy_docker() {
    log_info "部署 Docker 环境..."
    
    cd "$PROJECT_ROOT"
    
    # 检查 Docker Compose 文件
    local compose_file="deployments/docker-compose/logging-stack.yml"
    if [[ ! -f "$compose_file" ]]; then
        log_error "Docker Compose 文件不存在: $compose_file"
        exit 1
    fi
    
    log_info "启动 Docker Compose 栈..."
    docker-compose -f "$compose_file" up -d
    
    log_info "等待服务启动..."
    sleep 10
    
    log_info "检查服务状态..."
    docker-compose -f "$compose_file" ps
    
    log_info "Docker 环境部署完成！"
    cat << EOF

🎉 Docker 环境访问信息:
  VictoriaLogs UI:    http://localhost:9428/select/vmui/
  Grafana:           http://localhost:3000 (admin/admin)
  Prometheus:        http://localhost:9090
  OTLP HTTP 端点:     http://localhost:4318
  应用程序:           http://localhost:8080

📝 使用方法:
  1. 查看日志: docker-compose -f $compose_file logs -f
  2. 查询日志: curl "http://localhost:9428/select/logsql/query" -d 'query=*'

EOF
}

# Kubernetes 环境部署
deploy_k8s() {
    log_info "部署 Kubernetes 环境..."
    
    cd "$PROJECT_ROOT"
    
    # 创建命名空间
    log_info "创建命名空间..."
    kubectl apply -f deployments/kubernetes/namespace/logging.yaml
    
    # 部署 VictoriaLogs
    log_info "部署 VictoriaLogs..."
    kubectl apply -f deployments/kubernetes/victorialogs/
    
    # 部署 OTEL Collector
    log_info "部署 OpenTelemetry Collector..."
    kubectl apply -f deployments/kubernetes/otelcol/
    
    # 等待服务就绪
    log_info "等待服务就绪..."
    kubectl wait --for=condition=ready pod -l app=victorialogs -n "$NAMESPACE" --timeout=300s
    kubectl wait --for=condition=ready pod -l app=otelcol -n "$NAMESPACE" --timeout=300s
    
    # 部署示例应用（可选）
    if [[ "${DEPLOY_APP:-false}" == "true" ]]; then
        log_info "部署示例应用..."
        kubectl apply -f deployments/kubernetes/app/
        kubectl wait --for=condition=ready pod -l app=apiserver -n "$APP_NAMESPACE" --timeout=300s
    fi
    
    log_info "Kubernetes 环境部署完成！"
    cat << EOF

🎉 Kubernetes 环境访问信息:
  查看资源: kubectl get all -n $NAMESPACE
  端口转发 VictoriaLogs: kubectl port-forward -n $NAMESPACE svc/victorialogs 9428:9428
  端口转发 OTEL Collector: kubectl port-forward -n $NAMESPACE svc/otelcol 4318:4318
  
📝 使用方法:
  1. 端口转发: kubectl port-forward -n $NAMESPACE svc/victorialogs 9428:9428
  2. 访问 UI: http://localhost:9428/select/vmui/
  3. 查询日志: curl "http://localhost:9428/select/logsql/query" -d 'query=*'

EOF
}

# 清理环境
clean_env() {
    local env_type="$1"
    
    case "$env_type" in
        "local")
            log_info "清理本地环境..."
            cd "$PROJECT_ROOT"
            ./scripts/installation/service.sh stop all
            ;;
        "docker")
            log_info "清理 Docker 环境..."
            cd "$PROJECT_ROOT"
            docker-compose -f deployments/docker-compose/logging-stack.yml down -v
            ;;
        "k8s")
            log_info "清理 Kubernetes 环境..."
            kubectl delete -f deployments/kubernetes/app/ --ignore-not-found=true
            kubectl delete -f deployments/kubernetes/otelcol/ --ignore-not-found=true
            kubectl delete -f deployments/kubernetes/victorialogs/ --ignore-not-found=true
            kubectl delete -f deployments/kubernetes/namespace/logging.yaml --ignore-not-found=true
            ;;
    esac
    
    log_info "环境清理完成！"
}

# 检查状态
check_status() {
    local env_type="$1"
    
    case "$env_type" in
        "local")
            log_info "检查本地环境状态..."
            cd "$PROJECT_ROOT"
            ./scripts/installation/service.sh status all
            ;;
        "docker")
            log_info "检查 Docker 环境状态..."
            cd "$PROJECT_ROOT"
            docker-compose -f deployments/docker-compose/logging-stack.yml ps
            ;;
        "k8s")
            log_info "检查 Kubernetes 环境状态..."
            echo "=== Namespaces ==="
            kubectl get ns "$NAMESPACE" "$APP_NAMESPACE" 2>/dev/null || true
            echo -e "\n=== Logging Namespace ==="
            kubectl get all -n "$NAMESPACE" 2>/dev/null || true
            echo -e "\n=== App Namespace ==="
            kubectl get all -n "$APP_NAMESPACE" 2>/dev/null || true
            ;;
    esac
}

# 测试日志收集
test_logging() {
    local env_type="$1"
    
    log_info "测试日志收集功能..."
    
    # 发送测试日志
    local endpoint="http://localhost:4318/v1/logs"
    local test_log='{
        "resource_logs": [{
            "resource": {
                "attributes": [{
                    "key": "service.name",
                    "value": {"string_value": "test-service"}
                }]
            },
            "scope_logs": [{
                "log_records": [{
                    "time_unix_nano": "'$(date +%s%N)'",
                    "severity_text": "INFO",
                    "body": {
                        "string_value": "这是一条测试日志消息 - '$(date)'"
                    },
                    "attributes": [{
                        "key": "test.source",
                        "value": {"string_value": "deploy-script"}
                    }]
                }]
            }]
        }]
    }'
    
    if curl -s -X POST "$endpoint" \
        -H "Content-Type: application/json" \
        -d "$test_log" > /dev/null; then
        log_info "测试日志发送成功"
    else
        log_error "测试日志发送失败"
        return 1
    fi
    
    # 等待日志处理
    sleep 5
    
    # 查询测试日志
    local query_endpoint="http://localhost:9428/select/logsql/query"
    local query="service.name:test-service"
    
    if result=$(curl -s "$query_endpoint" -d "query=$query"); then
        if [[ -n "$result" && "$result" != "null" ]]; then
            log_info "✅ 日志收集测试通过 - 找到测试日志"
            return 0
        else
            log_warn "⚠️  未找到测试日志，可能需要等待更长时间"
            return 1
        fi
    else
        log_error "❌ 日志查询失败"
        return 1
    fi
}

# 主函数
main() {
    local command=""
    
    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                usage
                exit 0
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -n|--namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            local|docker|k8s)
                command="$1"
                DEPLOYMENT_TYPE="$1"
                shift
                ;;
            clean)
                if [[ $# -gt 1 ]]; then
                    clean_env "$2"
                    exit 0
                else
                    log_error "clean 命令需要指定环境类型"
                    usage
                    exit 1
                fi
                ;;
            status)
                if [[ $# -gt 1 ]]; then
                    check_status "$2"
                    exit 0
                else
                    log_error "status 命令需要指定环境类型"
                    usage
                    exit 1
                fi
                ;;
            test)
                if [[ $# -gt 1 ]]; then
                    test_logging "$2"
                    exit 0
                else
                    log_error "test 命令需要指定环境类型"
                    usage
                    exit 1
                fi
                ;;
            *)
                log_error "未知参数: $1"
                usage
                exit 1
                ;;
        esac
    done
    
    # 检查命令
    if [[ -z "$command" ]]; then
        log_error "请指定部署类型"
        usage
        exit 1
    fi
    
    # 检查依赖
    check_dependencies "$command"
    
    # 执行部署
    case "$command" in
        "local")
            deploy_local
            ;;
        "docker")
            deploy_docker
            ;;
        "k8s")
            deploy_k8s
            ;;
        *)
            log_error "未知的部署类型: $command"
            usage
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"