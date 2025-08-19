#!/bin/bash

# 从 configs/apiserver.yaml 解析日志配置
# 提取日志输出路径并生成 OTEL Collector 配置
set -euo pipefail

# 解析日志输出路径
parse_log_paths() {
    local config_file="$1"
    local project_root="$2"
    
    if [[ ! -f "$config_file" ]]; then
        echo "❌ 配置文件不存在: $config_file" >&2
        return 1
    fi
    
    # 使用 yq 或者 grep+sed 解析 YAML 文件中的 output-paths
    local log_paths=()
    
    # 尝试使用 yq（如果可用）
    if command -v yq >/dev/null 2>&1; then
        # 使用 yq 解析 YAML
        local paths_json
        paths_json=$(yq eval '.log.output-paths' "$config_file" 2>/dev/null || echo "null")
        
        if [[ "$paths_json" != "null" ]]; then
            # 解析 JSON 数组
            while IFS= read -r path; do
                path=$(echo "$path" | sed 's/^[[:space:]]*"//' | sed 's/"[[:space:]]*$//')
                if [[ "$path" != "stdout" && "$path" != "stderr" ]]; then
                    # 转换为绝对路径
                    if [[ "$path" != /* ]]; then
                        path="$project_root/$path"
                    fi
                    log_paths+=("$path")
                fi
            done < <(echo "$paths_json" | yq eval '.[]' -)
        fi
    else
        # fallback: 使用 grep 和 sed 解析
        local output_paths_line
        output_paths_line=$(grep -A1 "output-paths:" "$config_file" | tail -n1 | sed 's/^[[:space:]]*//')
        
        if [[ "$output_paths_line" =~ ^\[.*\] ]]; then
            # 解析 [xxx, yyy] 格式
            local paths_content
            paths_content=$(echo "$output_paths_line" | sed 's/^\[//' | sed 's/\]$//' | tr ',' '\n')
            
            while IFS= read -r path; do
                path=$(echo "$path" | sed 's/^[[:space:]]*//' | sed 's/[[:space:]]*$//' | sed 's/^"//' | sed 's/"$//')
                if [[ "$path" != "stdout" && "$path" != "stderr" && -n "$path" ]]; then
                    # 转换为绝对路径
                    if [[ "$path" != /* ]]; then
                        path="$project_root/$path"
                    fi
                    log_paths+=("$path")
                fi
            done <<< "$paths_content"
        fi
    fi
    
    # 输出解析到的路径
    for path in "${log_paths[@]}"; do
        echo "$path"
    done
}

# 生成 OTEL Collector filelog 配置的 include 部分
generate_filelog_includes() {
    local project_root="$1"
    local config_file="$2"
    local include_paths=()
    
    # 解析配置文件中的日志路径
    while IFS= read -r log_path; do
        if [[ -n "$log_path" ]]; then
            # 转换为容器内路径 (/app 是挂载点)
            local container_path
            container_path=$(echo "$log_path" | sed "s|^$project_root|/app|")
            include_paths+=("          - \"$container_path\"")
            
            # 添加通配符版本以匹配轮转的日志文件
            local dir_path file_name
            dir_path=$(dirname "$container_path")
            file_name=$(basename "$container_path")
            
            if [[ "$file_name" == *.log ]]; then
                local base_name="${file_name%.log}"
                include_paths+=("          - \"$dir_path/${base_name}*.log\"")
            fi
        fi
    done < <(parse_log_paths "$config_file" "$project_root")
    
    # 添加其他已知的日志文件路径
    include_paths+=("          - \"/app/api_startup*.log\"")
    include_paths+=("          - \"/app/logs/*.log\"")
    include_paths+=("          - \"/app/logs/apiserver/*.log\"")
    
    # 输出去重后的路径
    printf '%s\n' "${include_paths[@]}" | sort -u
}

# 生成完整的 OTEL Collector 配置文件
generate_otel_config() {
    local project_root="$1"
    local config_file="$2"
    local output_file="$3"
    
    local include_paths
    include_paths=$(generate_filelog_includes "$project_root" "$config_file")
    
    cat > "$output_file" << EOF
# 动态生成的 OTEL Collector 配置
# 基于 $config_file 中的日志路径配置
# 生成时间: $(date '+%Y-%m-%d %H:%M:%S')

receivers:
  filelog/apiserver:
    include:
$include_paths
    start_at: end
    operators:
      # JSON 解析器 - 处理结构化日志
      - type: json_parser
        id: json_parser
        parse_from: attributes.log
        parse_to: attributes
        severity:
          parse_from: attributes.level
        timestamp:
          parse_from: attributes.ts
          layout_type: gotime
          layout: '2006-01-02T15:04:05.000-0700'
      
      # 日志级别解析
      - type: severity_parser
        id: severity_parser
        parse_from: attributes.level
        
      # 移动消息字段
      - type: move
        id: move_message
        from: attributes.msg
        to: body
        
      # 资源属性设置
      - type: add
        field: resource["service.name"]
        value: "apiserver"
        
      - type: add  
        field: resource["service.version"]
        value: "v2.0.0"
        
      - type: add
        field: resource["deployment.environment"] 
        value: "development"

processors:
  # 批处理器 - 提高性能
  batch:
    send_batch_size: 100
    send_batch_max_size: 1000
    timeout: 1s
    
  # 资源处理器 - 添加通用属性
  resource:
    attributes:
      - key: service.namespace
        value: "go-protoc"
        action: upsert
      - key: telemetry.sdk.name
        value: "otelcol"  
        action: upsert
      - key: host.name
        value: "\${env:HOSTNAME}"
        action: upsert
        
  # 转换处理器 - 清理和格式化
  transform:
    log_statements:
      - context: log
        statements:
          # 清理空属性
          - delete_key(attributes, "log") where attributes["log"] == nil
          - delete_key(attributes, "ts") where attributes["ts"] == nil
          # 确保时间戳格式正确
          - set(time_unix_nano, time_unix_nano) where time_unix_nano == nil
          # 设置默认severity
          - set(severity_text, "INFO") where severity_text == nil

exporters:
  # 调试输出
  debug:
    verbosity: detailed
    sampling_initial: 1
    sampling_thereafter: 1
    
  # 文件输出 - 本地备份
  file/processed:
    path: /app/logs/otelcol-processed.json
    rotation:
      max_megabytes: 100
      max_days: 7
      max_backups: 3
      
  # VictoriaLogs 输出
  loki/victorialogs:
    endpoint: "http://host.docker.internal:9428/insert/jsonline"
    headers:
      "Content-Type": "application/x-ndjson"
    default_labels_enabled:
      exporter: false
      job: false
    
  # 文件输出 - VictoriaLogs 格式备份
  file/victorialogs:
    path: /app/logs/otelcol-victorialogs.json
    rotation:
      max_megabytes: 100
      max_days: 7
      max_backups: 3

extensions:
  health_check:
    endpoint: 0.0.0.0:13136
    path: "/health"

service:
  telemetry:
    logs:
      level: info
      development: false
      sampling:
        enabled: true
    metrics:
      address: 0.0.0.0:8891
      level: basic
      
  extensions: [health_check]
  
  pipelines:
    logs:
      receivers: [filelog/apiserver]
      processors: [resource, transform, batch]
      exporters: [debug, file/processed, loki/victorialogs, file/victorialogs]
EOF

    echo "✅ 动态配置文件已生成: $output_file"
}

# 主函数
main() {
    local project_root="${1:-}"
    local config_file="${2:-}"
    local output_file="${3:-}"
    
    if [[ -z "$project_root" || -z "$config_file" || -z "$output_file" ]]; then
        cat << EOF
用法: $0 <project_root> <config_file> <output_file>

参数:
  project_root  项目根目录路径
  config_file   apiserver.yaml 配置文件路径  
  output_file   输出的 OTEL 配置文件路径

示例:
  $0 /app /app/configs/apiserver.yaml /tmp/otelcol-dynamic.yaml
EOF
        exit 1
    fi
    
    echo "🔍 解析日志配置..."
    echo "📁 项目根目录: $project_root"
    echo "📄 配置文件: $config_file"
    echo "📤 输出文件: $output_file"
    
    # 解析并显示找到的日志路径
    echo ""
    echo "📋 发现的日志路径:"
    local found_paths=0
    while IFS= read -r log_path; do
        if [[ -n "$log_path" ]]; then
            echo "  - $log_path"
            ((found_paths++))
        fi
    done < <(parse_log_paths "$config_file" "$project_root")
    
    if [[ $found_paths -eq 0 ]]; then
        echo "⚠️  未找到配置的日志路径，将使用默认路径"
    else
        echo "✅ 找到 $found_paths 个日志路径"
    fi
    
    # 生成配置文件
    echo ""
    echo "⚙️  生成 OTEL Collector 配置..."
    generate_otel_config "$project_root" "$config_file" "$output_file"
    
    echo ""
    echo "🎉 配置解析完成！"
}

# 如果脚本被直接执行
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi