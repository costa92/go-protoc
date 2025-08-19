#!/usr/bin/env bash
#
# 日志收集系统管理脚本
# 使用 OpenTelemetry Collector 作为 Agent 收集日志并发送到 VictoriaLogs
#

set -o errexit
set -o nounset
set -o pipefail

# 加载通用配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# 配置文件路径
OTELCOL_CONFIG_LOGS="${SCRIPT_DIR}/otelcol/config-logs-victorialogs.yaml"
OTELCOL_CONFIG_DOCKER="${SCRIPT_DIR}/otelcol/config-docker.yaml"

# 服务名称
OTELCOL_SERVICE_NAME="otelcol-logs"
VICTORIALOGS_SERVICE_NAME="victorialogs"

# Function: 启动完整日志收集系统
logs_collection::start() {
    proj::log::info "🚀 启动日志收集系统..."
    
    # 1. 启动 VictoriaLogs
    proj::log::info "启动 VictoriaLogs..."
    if ! logs_collection::victorialogs::start; then
        proj::log::error "VictoriaLogs 启动失败"
        return 1
    fi
    
    # 2. 启动 OpenTelemetry Collector
    proj::log::info "启动 OpenTelemetry Collector (日志代理)..."
    if ! logs_collection::otelcol::start; then
        proj::log::error "OpenTelemetry Collector 启动失败"
        return 1
    fi
    
    # 3. 验证服务状态
    proj::log::info "验证服务状态..."
    sleep 5
    logs_collection::status
    
    proj::log::success "✅ 日志收集系统启动完成!"
    logs_collection::info
}

# Function: 停止日志收集系统
logs_collection::stop() {
    proj::log::info "🛑 停止日志收集系统..."
    
    # 停止 OpenTelemetry Collector
    logs_collection::otelcol::stop
    
    # 停止 VictoriaLogs  
    logs_collection::victorialogs::stop
    
    proj::log::success "✅ 日志收集系统已停止"
}

# Function: 重启日志收集系统
logs_collection::restart() {
    proj::log::info "🔄 重启日志收集系统..."
    logs_collection::stop
    sleep 3
    logs_collection::start
}

# Function: 检查系统状态
logs_collection::status() {
    proj::log::info "📊 日志收集系统状态:"
    proj::log::info "========================"
    
    # 检查 VictoriaLogs
    local vl_status="❌ 离线"
    if curl -s --max-time 5 "http://127.0.0.1:9428/health" >/dev/null 2>&1; then
        vl_status="✅ 在线"
    fi
    proj::log::info "VictoriaLogs:     ${vl_status}"
    
    # 检查 OpenTelemetry Collector
    local otlp_status="❌ 离线" 
    if curl -s --max-time 5 "http://127.0.0.1:13133" >/dev/null 2>&1; then
        otlp_status="✅ 在线"
    fi
    proj::log::info "OTel Collector:   ${otlp_status}"
    
    # 检查端口监听
    proj::log::info ""
    proj::log::info "🔌 端口监听状态:"
    proj::log::info "- OTLP gRPC:      4327 $(logs_collection::check_port 4327)"
    proj::log::info "- OTLP HTTP:      4328 $(logs_collection::check_port 4328)"
    proj::log::info "- VictoriaLogs:   9428 $(logs_collection::check_port 9428)"
    proj::log::info "- Health Check:   13133 $(logs_collection::check_port 13133)"
}

# Function: 检查端口是否监听
logs_collection::check_port() {
    local port=$1
    if netstat -ln 2>/dev/null | grep -q ":${port} "; then
        echo "✅"
    else
        echo "❌"
    fi
}

# Function: 启动 VictoriaLogs
logs_collection::victorialogs::start() {
    if docker ps -q -f name=go-protoc-victorialogs | grep -q .; then
        proj::log::info "VictoriaLogs 已在运行"
        return 0
    fi
    
    "${SCRIPT_DIR}/service.sh" start victorialogs
}

# Function: 停止 VictoriaLogs  
logs_collection::victorialogs::stop() {
    "${SCRIPT_DIR}/service.sh" stop victorialogs
}

# Function: 启动 OpenTelemetry Collector
logs_collection::otelcol::start() {
    if docker ps -q -f name="${OTELCOL_SERVICE_NAME}" | grep -q .; then
        proj::log::info "OpenTelemetry Collector 已在运行"
        return 0
    fi
    
    # 确保配置文件存在
    if [[ ! -f "${OTELCOL_CONFIG_LOGS}" ]]; then
        proj::log::error "配置文件不存在: ${OTELCOL_CONFIG_LOGS}"
        return 1
    fi
    
    # 确保日志目录存在
    proj::util::sudo "mkdir -p /var/log/otelcol"
    proj::util::sudo "mkdir -p /var/log/app"
    proj::util::sudo "mkdir -p /var/log/apiserver"
    proj::util::sudo "chmod 777 /var/log/otelcol /var/log/app /var/log/apiserver"
    
    # 启动 OpenTelemetry Collector 容器
    proj::log::info "启动 OpenTelemetry Collector..."
    docker run -d --name "${OTELCOL_SERVICE_NAME}" \
        --restart always \
        --network go-protoc-network \
        -p 4327:4327 \
        -p 4328:4328 \
        -p 13133:13133 \
        -p 8889:8889 \
        -p 1777:1777 \
        -p 55679:55679 \
        -v "${OTELCOL_CONFIG_LOGS}:/etc/otelcol/config.yaml:ro" \
        -v /var/log:/var/log \
        otel/opentelemetry-collector-contrib:latest \
        --config=/etc/otelcol/config.yaml
    
    # 等待服务启动
    proj::log::info "等待 OpenTelemetry Collector 启动..."
    sleep 3
}

# Function: 停止 OpenTelemetry Collector
logs_collection::otelcol::stop() {
    if docker ps -q -f name="${OTELCOL_SERVICE_NAME}" | grep -q .; then
        proj::log::info "停止 OpenTelemetry Collector..."
        docker rm -f "${OTELCOL_SERVICE_NAME}"
    fi
}

# Function: 查看日志
logs_collection::logs() {
    local service=${1:-"all"}
    
    case "${service}" in
        otelcol|collector)
            proj::log::info "📋 OpenTelemetry Collector 日志:"
            docker logs -f "${OTELCOL_SERVICE_NAME}" 2>/dev/null || \
                proj::log::error "OpenTelemetry Collector 不在运行"
            ;;
        victorialogs|vl)
            proj::log::info "📋 VictoriaLogs 日志:"
            docker logs -f go-protoc-victorialogs 2>/dev/null || \
                proj::log::error "VictoriaLogs 不在运行"
            ;;
        all|*)
            proj::log::info "📋 所有服务日志 (按 Ctrl+C 退出):"
            proj::log::info "================================="
            docker logs -f "${OTELCOL_SERVICE_NAME}" &
            docker logs -f go-protoc-victorialogs &
            wait
            ;;
    esac
}

# Function: 显示系统信息
logs_collection::info() {
    proj::log::info ""
    proj::log::info "🔗 访问链接:"
    proj::log::info "========================="
    proj::log::info "VictoriaLogs UI:       http://127.0.0.1:9428/select/vmui"
    proj::log::info "OTel Health Check:     http://127.0.0.1:13133"
    proj::log::info "OTel Metrics:          http://127.0.0.1:8889/metrics"
    proj::log::info "OTel pprof:            http://127.0.0.1:1777/debug/pprof"
    proj::log::info "OTel zpages:           http://127.0.0.1:55679/debug/tracez"
    proj::log::info ""
    proj::log::info "📊 测试命令:"
    proj::log::info "========================="
    proj::log::info "检查状态:              ./logs-collection.sh status"
    proj::log::info "查看日志:              ./logs-collection.sh logs [otelcol|victorialogs|all]"
    proj::log::info "查询日志:              curl 'http://127.0.0.1:9428/select/logsql/query' -d 'query=*'"
    proj::log::info "发送测试日志:          ./logs-collection.sh test-log"
}

# Function: 发送测试日志
logs_collection::test_log() {
    proj::log::info "📤 发送测试日志到 OTLP..."
    
    # 使用 curl 发送 JSON 格式的测试日志
    curl -X POST "http://127.0.0.1:4328/v1/logs" \
        -H "Content-Type: application/json" \
        -d '{
            "resourceLogs": [{
                "resource": {
                    "attributes": [{
                        "key": "service.name",
                        "value": {"stringValue": "test-app"}
                    }, {
                        "key": "service.version", 
                        "value": {"stringValue": "1.0.0"}
                    }]
                },
                "scopeLogs": [{
                    "logRecords": [{
                        "timeUnixNano": "'$(date +%s)000000000'",
                        "severityText": "INFO",
                        "body": {"stringValue": "Test log message from logs-collection script at '$(date)'"},
                        "attributes": [{
                            "key": "level",
                            "value": {"stringValue": "info"}
                        }, {
                            "key": "environment",
                            "value": {"stringValue": "development"}
                        }, {
                            "key": "test_id",
                            "value": {"stringValue": "'$(uuidgen 2>/dev/null || echo "test-$(date +%s)")'"}
                        }]
                    }]
                }]
            }]
        }'
    
    proj::log::success "✅ 测试日志已发送"
    proj::log::info "等待 5 秒后查询日志..."
    sleep 5
    
    # 查询刚发送的日志
    proj::log::info "📊 查询测试日志:"
    curl -s "http://127.0.0.1:9428/select/logsql/query" \
        -d 'query=_msg:*test* | limit 10' | jq . || \
        curl -s "http://127.0.0.1:9428/select/logsql/query" \
            -d 'query=* | limit 5'
}

# Main function
main() {
    local action=${1:-}
    
    case "${action}" in
        start)
            logs_collection::start
            ;;
        stop) 
            logs_collection::stop
            ;;
        restart)
            logs_collection::restart
            ;;
        status)
            logs_collection::status
            ;;
        logs)
            logs_collection::logs "${2:-all}"
            ;;
        info)
            logs_collection::info
            ;;
        test-log)
            logs_collection::test_log
            ;;
        *)
            proj::log::error "Usage: $0 {start|stop|restart|status|logs [service]|info|test-log}"
            proj::log::info ""
            proj::log::info "Available commands:"
            proj::log::info "  start      - 启动完整日志收集系统"
            proj::log::info "  stop       - 停止日志收集系统"  
            proj::log::info "  restart    - 重启日志收集系统"
            proj::log::info "  status     - 检查系统状态"
            proj::log::info "  logs       - 查看服务日志 [otelcol|victorialogs|all]"
            proj::log::info "  info       - 显示系统信息"
            proj::log::info "  test-log   - 发送测试日志"
            return 1
            ;;
    esac
}

# 如果脚本直接执行，则运行 main 函数
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi