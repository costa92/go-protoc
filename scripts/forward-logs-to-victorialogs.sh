#\!/bin/bash

# 日志转发脚本 - 将OTEL Collector处理的日志发送到VictoriaLogs
set -euo pipefail

# 配置
VICTORIALOGS_URL="http://127.0.0.1:9428/insert/loki/api/v1/push"
OTEL_LOG_FILE="/Users/costalong/code/go/src/github.com/costa92/go-protoc/logs/otelcol-processed.json"
SERVICE_NAME="apiserver"
ENVIRONMENT="development"

# 转换OTEL日志格式为Loki格式
convert_otel_to_loki() {
    local otel_log_file="$1"
    
    if [[ ! -f "$otel_log_file" ]]; then
        echo "OTEL日志文件不存在: $otel_log_file"
        return 1
    fi
    
    # 读取OTEL日志并转换为Loki格式
    jq -c --arg service "$SERVICE_NAME" --arg env "$ENVIRONMENT" '
    {
        "streams": [
            .resourceLogs[]?.scopeLogs[]?.logRecords[] | 
            {
                "stream": {
                    "service": $service,
                    "environment": $env,
                    "level": (.attributes[] | select(.key == "level") | .value.stringValue // "info"),
                    "caller": (.attributes[] | select(.key == "caller") | .value.stringValue // "unknown")
                },
                "values": [
                    [
                        ((.observedTimeUnixNano | tonumber) | tostring),
                        (.body.stringValue // (.attributes[] | select(.key == "msg") | .value.stringValue // "no message"))
                    ]
                ]
            }
        ]
    }' "$otel_log_file"
}

# 发送日志到VictoriaLogs
send_to_victorialogs() {
    local loki_payload="$1"
    
    echo "发送日志到VictoriaLogs..."
    
    local response
    if response=$(echo "$loki_payload" | curl -s -w "HTTP_CODE:%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d @- \
        "$VICTORIALOGS_URL" 2>/dev/null); then
        
        local http_code=$(echo "$response" | grep -o 'HTTP_CODE:[0-9]*' | cut -d: -f2)
        
        if [[ "$http_code" =~ ^2[0-9][0-9]$ ]]; then
            echo "日志发送成功 (HTTP $http_code)"
            return 0
        else
            echo "日志发送失败 (HTTP $http_code)"
            return 1
        fi
    else
        echo "无法连接到VictoriaLogs API"
        return 1
    fi
}

# 主函数
main() {
    echo "开始日志转发处理..."
    
    local loki_payload
    if loki_payload=$(convert_otel_to_loki "$OTEL_LOG_FILE"); then
        if [[ -n "$loki_payload" && "$loki_payload" != "null" ]]; then
            if send_to_victorialogs "$loki_payload"; then
                echo "日志转发完成"
                local log_count=$(echo "$loki_payload" | jq '.streams[].values | length' | awk '{sum+=$1} END {print sum+0}')
                echo "成功转发 $log_count 条日志记录"
            else
                exit 1
            fi
        else
            echo "没有找到有效的日志记录"
        fi
    else
        echo "日志格式转换失败"
        exit 1
    fi
}

# 运行主函数
main "$@"