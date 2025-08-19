#!/bin/bash

# 动态启动 OpenTelemetry Collector Docker 容器
# 自动检测当前项目目录并挂载到Docker容器
set -euo pipefail

# 配置
CONTAINER_NAME="otelcol-logs-collector"
OTELCOL_IMAGE="otel/opentelemetry-collector-contrib:0.91.0"

# 自动检测当前项目根目录
detect_project_root() {
    local current_dir="$PWD"
    local project_root=""
    
    # 向上查找包含特征文件的目录（go.mod, CLAUDE.md等）
    while [[ "$current_dir" != "/" ]]; do
        if [[ -f "$current_dir/go.mod" ]] && [[ -f "$current_dir/CLAUDE.md" ]]; then
            project_root="$current_dir"
            break
        fi
        current_dir=$(dirname "$current_dir")
    done
    
    if [[ -z "$project_root" ]]; then
        # 如果没找到项目根目录，使用当前目录
        project_root="$PWD"
    fi
    
    echo "$project_root"
}

# 检查必要的配置文件
check_prerequisites() {
    local project_root="$1"
    local apiserver_config="$project_root/configs/apiserver.yaml"
    local parse_script="$project_root/scripts/installation/parse-log-config.sh"
    
    if [[ ! -f "$apiserver_config" ]]; then
        echo "❌ 找不到 apiserver 配置文件: $apiserver_config"
        exit 1
    fi
    
    if [[ ! -f "$parse_script" ]]; then
        echo "❌ 找不到配置解析脚本: $parse_script"
        exit 1
    fi
    
    echo "✅ apiserver 配置文件已找到: $apiserver_config"
    echo "✅ 配置解析脚本已找到: $parse_script"
}

# 创建必要的日志目录
create_log_directories() {
    local project_root="$1"
    
    mkdir -p "$project_root/logs"
    mkdir -p "$project_root/logs/apiserver"
    
    echo "✅ 日志目录已创建"
}

# 停止已存在的容器
stop_existing_container() {
    if docker ps -a --format 'table {{.Names}}' | grep -q "^$CONTAINER_NAME$"; then
        echo "🔄 停止现有容器: $CONTAINER_NAME"
        docker stop "$CONTAINER_NAME" >/dev/null 2>&1 || true
        docker rm "$CONTAINER_NAME" >/dev/null 2>&1 || true
    fi
}

# 生成动态配置文件
generate_dynamic_config() {
    local project_root="$1"
    local apiserver_config="$project_root/configs/apiserver.yaml"
    local parse_script="$project_root/scripts/installation/parse-log-config.sh"
    local dynamic_config="/tmp/otelcol-dynamic-$$.yaml"
    
    echo "⚙️  生成动态 OTEL Collector 配置..." >&2
    echo "📄 读取配置: $apiserver_config" >&2
    
    # 使用解析脚本生成动态配置（静默模式）
    if ! "$parse_script" "$project_root" "$apiserver_config" "$dynamic_config" >/dev/null 2>&1; then
        echo "❌ 动态配置生成失败"
        exit 1
    fi
    
    if [[ ! -f "$dynamic_config" ]]; then
        echo "❌ 动态配置文件不存在"
        exit 1
    fi
    
    echo "✅ 动态配置已生成: $dynamic_config" >&2
    echo "$dynamic_config"
}

# 启动OTEL Collector容器
start_otelcol_container() {
    local project_root="$1"
    local config_file
    local config_dir="/tmp/otelcol-config-$$"
    
    # 生成动态配置文件
    config_file=$(generate_dynamic_config "$project_root")
    
    echo "🚀 启动 OTEL Collector 容器..."
    echo "📁 项目根目录: $project_root"
    echo "📝 动态配置文件: $config_file"
    
    # 验证配置文件存在且为文件
    if [[ ! -f "$config_file" ]]; then
        echo "❌ 配置文件不存在或不是文件: $config_file"
        exit 1
    fi
    
    # 将配置文件放在项目目录下，避免macOS Docker挂载问题
    local project_config_dir="$project_root/.otelcol"
    mkdir -p "$project_config_dir"
    cp "$config_file" "$project_config_dir/config.yaml"
    
    echo "✅ 验证配置文件: $(ls -la "$project_config_dir/config.yaml")"
    echo "📂 配置目录: $project_config_dir"
    
    # 启动容器，使用项目目录下的配置
    docker run -d \
        --name "$CONTAINER_NAME" \
        --restart unless-stopped \
        -v "$project_root:/app:rw" \
        -p 4331:4317 \
        -p 4332:4318 \
        -p 8891:8891 \
        -p 13136:13136 \
        --add-host=host.docker.internal:host-gateway \
        "$OTELCOL_IMAGE" \
        --config=/app/.otelcol/config.yaml
        
    # 等待容器启动
    echo "⏳ 等待容器启动..."
    sleep 3
    
    # 检查容器状态
    if docker ps --format 'table {{.Names}}\t{{.Status}}' | grep -q "$CONTAINER_NAME.*Up"; then
        echo "✅ OTEL Collector 容器启动成功"
        echo "📊 容器状态:"
        docker ps --filter "name=$CONTAINER_NAME" --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'
        
        echo ""
        echo "🔗 服务端点:"
        echo "  - Health Check: http://localhost:13136/health"
        echo "  - Metrics: http://localhost:8891/metrics"
        echo "  - OTLP gRPC: localhost:4331"
        echo "  - OTLP HTTP: localhost:4332"
        
    else
        echo "❌ 容器启动失败"
        echo "📜 容器日志:"
        docker logs "$CONTAINER_NAME" || true
        exit 1
    fi
}

# 验证服务健康状态
check_health() {
    echo ""
    echo "🏥 检查服务健康状态..."
    
    # 等待服务完全启动
    local max_attempts=30
    local attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        if curl -s -f "http://localhost:13136/health" >/dev/null 2>&1; then
            echo "✅ OTEL Collector 健康检查通过"
            break
        else
            echo "⏳ 等待服务启动... ($attempt/$max_attempts)"
            sleep 2
            ((attempt++))
        fi
    done
    
    if [[ $attempt -gt $max_attempts ]]; then
        echo "❌ 健康检查超时，请检查容器日志"
        docker logs "$CONTAINER_NAME" --tail 20
        exit 1
    fi
}

# 显示日志监控状态
show_monitoring_status() {
    local project_root="$1"
    local apiserver_config="$project_root/configs/apiserver.yaml"
    local parse_script="$project_root/scripts/installation/parse-log-config.sh"
    
    echo ""
    echo "📂 文件监控状态:"
    echo "  基于配置文件: $apiserver_config"
    echo ""
    echo "  监控的日志路径:"
    
    # 显示从配置文件解析出的路径
    if [[ -f "$parse_script" ]]; then
        while IFS= read -r log_path; do
            if [[ -n "$log_path" ]]; then
                echo "    - $log_path (从配置文件解析)"
            fi
        done < <("$parse_script" "$project_root" "$apiserver_config" "/dev/null" 2>/dev/null | grep "  - " | sed 's/  - //' || true)
    fi
    
    # 显示其他监控路径
    echo "    - $project_root/api_startup*.log (runtime logs)"
    echo "    - $project_root/logs/*.log (general logs)"
    echo "    - $project_root/logs/apiserver/*.log (service logs)"
    echo ""
    echo "  输出文件:"
    echo "    - $project_root/logs/otelcol-processed.json"
    echo "    - $project_root/logs/otelcol-victorialogs.json"
    echo ""
    echo "🎯 使用以下命令查看实时日志:"
    echo "  docker logs -f $CONTAINER_NAME"
    echo ""
    echo "📊 使用以下脚本转发到VictoriaLogs:"
    echo "  ./scripts/forward-logs-to-victorialogs.sh"
    echo ""
    echo "🔧 重新生成配置:"
    echo "  $parse_script $project_root $apiserver_config /tmp/new-config.yaml"
}

# 主函数
main() {
    echo "=== 动态启动 OTEL Collector ==="
    echo ""
    
    # 自动检测项目根目录
    local project_root
    project_root=$(detect_project_root)
    echo "🔍 检测到项目根目录: $project_root"
    
    # 检查前置条件
    check_prerequisites "$project_root"
    
    # 创建必要目录
    create_log_directories "$project_root"
    
    # 停止现有容器
    stop_existing_container
    
    # 启动新容器
    start_otelcol_container "$project_root"
    
    # 健康检查
    check_health
    
    # 显示监控状态
    show_monitoring_status "$project_root"
    
    echo ""
    echo "🎉 OTEL Collector 动态启动完成！"
    echo "📍 项目目录已自动挂载到容器 /app 路径"
}

# 如果脚本被直接执行
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi