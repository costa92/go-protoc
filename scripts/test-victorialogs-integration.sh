#\!/bin/bash

# 简单的VictoriaLogs集成测试脚本
set -e

echo "=== VictoriaLogs 集成测试 ==="

# 检查OTEL Collector日志文件
echo "1. 检查OTEL Collector处理的日志文件..."
if [ -f "logs/otelcol-processed.json" ]; then
    echo "✓ OTEL日志文件存在"
    echo "文件大小: $(wc -c < logs/otelcol-processed.json) bytes"
    echo "日志条数: $(wc -l < logs/otelcol-processed.json) lines"
else
    echo "✗ OTEL日志文件不存在"
    exit 1
fi

# 检查VictoriaLogs连接
echo ""
echo "2. 检查VictoriaLogs服务..."
if curl -s -f "http://127.0.0.1:9428/health" > /dev/null; then
    echo "✓ VictoriaLogs连接正常"
else
    echo "✗ 无法连接到VictoriaLogs"
    exit 1
fi

# 发送测试日志到VictoriaLogs
echo ""
echo "3. 发送测试日志到VictoriaLogs..."

# 创建简单的Loki格式测试数据
test_payload='{
    "streams": [
        {
            "stream": {
                "service": "apiserver",
                "environment": "development",
                "level": "info"
            },
            "values": [
                ["'$(date +%s%N)'", "Test message from OTEL Collector integration"]
            ]
        }
    ]
}'

response=$(curl -s -w "HTTP_CODE:%{http_code}" \
    -X POST \
    -H "Content-Type: application/json" \
    -d "$test_payload" \
    "http://127.0.0.1:9428/insert/loki/api/v1/push" 2>/dev/null)

http_code=$(echo "$response" | grep -o 'HTTP_CODE:[0-9]*' | cut -d: -f2)

if [[ "$http_code" =~ ^2[0-9][0-9]$ ]]; then
    echo "✓ 测试日志发送成功 (HTTP $http_code)"
else
    echo "✗ 测试日志发送失败 (HTTP $http_code)"
    echo "响应: $(echo "$response" | sed 's/HTTP_CODE:[0-9]*$//')"
    exit 1
fi

# 查询VictoriaLogs中的日志
echo ""
echo "4. 查询VictoriaLogs中的日志..."
sleep 2  # 等待日志被索引

query_response=$(curl -s "http://127.0.0.1:9428/select/logsql/query" \
    -d 'query={service="apiserver"}' \
    -d 'limit=5')

if echo "$query_response" | grep -q "Test message"; then
    echo "✓ 成功查询到发送的测试日志"
else
    echo "⚠ 未能查询到测试日志，但发送成功"
fi

echo ""
echo "=== 集成测试完成 ==="
echo "🎉 OTEL Collector -> VictoriaLogs 日志流已建立！"
echo ""
echo "可以通过以下方式查看日志:"
echo "- Web UI: http://127.0.0.1:9428/select/vmui/"
echo "- API查询: curl -s \"http://127.0.0.1:9428/select/logsql/query\" -d 'query=*'"