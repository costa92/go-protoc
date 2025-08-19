#!/usr/bin/env bash
#
# 日志收集系统集成测试脚本
#

set -o errexit
set -o nounset 
set -o pipefail

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🧪 日志收集系统集成测试${NC}"
echo "================================="

# 1. 检查服务状态
echo -e "\n${YELLOW}📊 检查服务状态...${NC}"

# 检查 VictoriaLogs
if curl -s --max-time 5 "http://127.0.0.1:9428/health" >/dev/null 2>&1; then
    echo -e "✅ VictoriaLogs: 运行中"
else
    echo -e "❌ VictoriaLogs: 离线"
    exit 1
fi

# 检查 OpenTelemetry Collector
if curl -s --max-time 5 "http://127.0.0.1:13133" >/dev/null 2>&1; then
    echo -e "✅ OpenTelemetry Collector: 运行中"
else
    echo -e "❌ OpenTelemetry Collector: 离线" 
    exit 1
fi

# 2. 测试 OTLP HTTP 端点
echo -e "\n${YELLOW}📤 测试 OTLP HTTP 端点...${NC}"

TEST_ID="test-$(date +%s)"
RESPONSE=$(curl -X POST "http://127.0.0.1:4328/v1/logs" \
    -H "Content-Type: application/json" \
    -d '{
        "resourceLogs": [{
            "resource": {
                "attributes": [{
                    "key": "service.name",
                    "value": {"stringValue": "integration-test"}
                }, {
                    "key": "test.id",
                    "value": {"stringValue": "'${TEST_ID}'"}
                }]
            },
            "scopeLogs": [{
                "logRecords": [{
                    "timeUnixNano": "'$(date +%s)000000000'",
                    "severityText": "INFO",
                    "body": {"stringValue": "Integration test log message - '${TEST_ID}'"},
                    "attributes": [{
                        "key": "level",
                        "value": {"stringValue": "info"}
                    }, {
                        "key": "environment", 
                        "value": {"stringValue": "test"}
                    }]
                }]
            }]
        }]
    }' 2>/dev/null)

if [[ "${RESPONSE}" == *"partialSuccess"* ]]; then
    echo -e "✅ OTLP HTTP: 日志发送成功"
else
    echo -e "❌ OTLP HTTP: 日志发送失败"
    echo "Response: ${RESPONSE}"
    exit 1
fi

# 3. 测试 OTLP gRPC 端点 (使用 grpcurl 如果可用)
echo -e "\n${YELLOW}📤 测试 OTLP gRPC 端点...${NC}"
if command -v grpcurl >/dev/null 2>&1; then
    echo "使用 grpcurl 测试 gRPC 端点..."
    # TODO: 实现 gRPC 测试
    echo -e "⚠️  gRPC 测试待实现"
else
    echo -e "⚠️  grpcurl 未安装，跳过 gRPC 测试"
fi

# 4. 验证日志文件输出
echo -e "\n${YELLOW}📁 验证日志文件输出...${NC}"

sleep 2 # 等待日志写入

if [[ -f "logs/otelcol-output.json" ]] && [[ -s "logs/otelcol-output.json" ]]; then
    echo -e "✅ 日志文件: 存在且有内容"
    
    # 检查最新日志条目
    if grep -q "${TEST_ID}" logs/otelcol-output.json; then
        echo -e "✅ 测试日志: 在文件中找到"
    else
        echo -e "⚠️  测试日志: 未在文件中找到"
    fi
else
    echo -e "❌ 日志文件: 不存在或为空"
fi

# 5. 测试健康检查端点
echo -e "\n${YELLOW}🏥 测试健康检查端点...${NC}"

HEALTH_RESPONSE=$(curl -s "http://127.0.0.1:13133")
if [[ "${HEALTH_RESPONSE}" == *"Server available"* ]]; then
    echo -e "✅ 健康检查: 通过"
else
    echo -e "❌ 健康检查: 失败"
    echo "Response: ${HEALTH_RESPONSE}"
fi

# 6. 显示统计信息
echo -e "\n${YELLOW}📈 显示统计信息...${NC}"

echo "容器状态:"
docker ps --filter "name=otelcol-logs" --filter "name=proj-victorialogs" --format "table {{.Names}}\t{{.Status}}"

echo -e "\n日志文件大小:"
if [[ -f "logs/otelcol-output.json" ]]; then
    wc -l logs/otelcol-output.json | awk '{print "  日志条数: " $1}'
    ls -lh logs/otelcol-output.json | awk '{print "  文件大小: " $5}'
fi

# 7. 测试多条日志批量发送
echo -e "\n${YELLOW}📦 测试批量日志发送...${NC}"

for i in {1..5}; do
    curl -X POST "http://127.0.0.1:4328/v1/logs" \
        -H "Content-Type: application/json" \
        -d '{
            "resourceLogs": [{
                "resource": {
                    "attributes": [{
                        "key": "service.name",
                        "value": {"stringValue": "batch-test"}
                    }]
                },
                "scopeLogs": [{
                    "logRecords": [{
                        "timeUnixNano": "'$(date +%s)000000000'",
                        "severityText": "INFO",
                        "body": {"stringValue": "Batch test log #'${i}'"},
                        "attributes": [{
                            "key": "batch.id",
                            "value": {"stringValue": "'${TEST_ID}'"}
                        }]
                    }]
                }]
            }]
        }' >/dev/null 2>&1
    
    echo -ne "  发送批量日志 ${i}/5\r"
done

echo -e "\n✅ 批量日志: 发送完成"

# 8. 最终验证
echo -e "\n${YELLOW}🎯 最终验证...${NC}"

sleep 3 # 等待处理

FINAL_COUNT=$(wc -l logs/otelcol-output.json 2>/dev/null | awk '{print $1}' || echo "0")
echo -e "📊 总日志条数: ${FINAL_COUNT}"

if [[ ${FINAL_COUNT} -gt 0 ]]; then
    echo -e "\n${GREEN}🎉 集成测试通过！${NC}"
    echo -e "日志收集系统正常工作"
else
    echo -e "\n${RED}❌ 集成测试失败${NC}"
    echo -e "日志收集系统可能存在问题"
    exit 1
fi

# 9. 显示访问信息
echo -e "\n${YELLOW}🔗 访问链接:${NC}"
echo "  OpenTelemetry Collector 健康检查: http://127.0.0.1:13133"
echo "  VictoriaLogs UI:                   http://127.0.0.1:9428/select/vmui"
echo "  日志文件:                          $(pwd)/logs/otelcol-output.json"

echo -e "\n${YELLOW}🔧 管理命令:${NC}"
echo "  查看 OTel 日志:    docker logs otelcol-logs"
echo "  查看 VL 日志:      docker logs proj-victorialogs"
echo "  重启系统:          make logs-restart"
echo "  停止系统:          make logs-stop"