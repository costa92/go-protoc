#!/usr/bin/env bash

# OTEL 代理启动脚本 - 按顺序启动 Collector 和 Agent
# Project: ${PROJ_NAME:-go-protoc}
# Service: OpenTelemetry Stack (Collector + Agent) ${OTELCOL_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

set -eEuo pipefail

# 加载工具函数和环境配置
if [[ -f "${PROJ_ROOT_DIR}/scripts/lib/util.sh" ]]; then
    source "${PROJ_ROOT_DIR}/scripts/lib/util.sh"
elif [[ -f "$(pwd)/scripts/lib/util.sh" ]]; then
    source "$(pwd)/scripts/lib/util.sh"
else
    proj::util::is_mac() { [[ "$(uname)" == "Darwin" ]]; }
fi

# 加载环境配置
if [[ -f "${PROJ_ROOT_DIR}/manifests/env/env.dev" ]]; then
    source "${PROJ_ROOT_DIR}/manifests/env/env.dev"
elif [[ -f "$(pwd)/manifests/env/env.dev" ]]; then
    source "$(pwd)/manifests/env/env.dev"
fi

# 获取模板脚本管理器
DOCKER_TEMPLATE_LIB_LOADER="source ${PROJ_ROOT_DIR}/manifests/env/env.dev && \
                           source ${PROJ_ROOT_DIR}/scripts/installation/common.sh && \
                           source ${PROJ_ROOT_DIR}/scripts/installation/lib/common_lib.sh"

echo "🚀 启动 OTEL Stack (Collector + Agent)..."
echo "数据流: 应用程序 → Agent(4327) → Collector(4317) → 后端存储"
echo ""

# 步骤1: 启动 OTEL Collector
echo "===========> 步骤 1/2: 启动 OTEL Collector"
eval "${DOCKER_TEMPLATE_LIB_LOADER}"
if ! proj::docker::run_service_from_template 'otel-collector'; then
    echo "❌ OTEL Collector 启动失败"
    exit 1
fi

echo ""
echo "✅ OTEL Collector 启动成功，等待服务就绪..."
sleep 3

# 步骤2: 启动 OTEL Agent
echo "===========> 步骤 2/2: 启动 OTEL Agent"
eval "${DOCKER_TEMPLATE_LIB_LOADER}"
if ! proj::docker::run_service_from_template 'otel-agent'; then
    echo "❌ OTEL Agent 启动失败"
    echo "⚠️  Collector 仍在运行，如需清理请执行: make docker.otel-collector.stop"
    exit 1
fi

echo ""
echo "🎉 OTEL Stack 启动完成！"
echo ""
echo "服务状态:"
docker ps --filter name="proj-otel-" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo ""
echo "应用程序连接配置:"
echo "  OTLP gRPC: localhost:4327"
echo "  OTLP HTTP: localhost:4328"
echo ""
echo "监控端点:"
echo "  Collector Health: http://localhost:13133"
echo "  Agent Health: http://localhost:13134"
echo "  Prometheus Metrics: http://localhost:8888"
echo ""
echo "数据流路径已建立: App → Agent → Collector → VictoriaLogs/Jaeger/Prometheus"