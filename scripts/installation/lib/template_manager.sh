#!/usr/bin/env bash

# =============================================================================
# 模板管理器
# Template Manager
#
# 提供模板变量替换和配置文件生成功能
# 支持多平台、多服务的配置模板管理
# =============================================================================

set -eEuo pipefail

# 避免重复加载
[[ "${TEMPLATE_MANAGER_LIB_LOADED:-false}" == "true" ]] && return 0

# 模板目录常量
readonly TEMPLATES_ROOT_DIR="${PROJ_ROOT_DIR}/scripts/installation/templates"
readonly DOCKER_TEMPLATES_DIR="${TEMPLATES_ROOT_DIR}/docker"
readonly UBUNTU_TEMPLATES_DIR="${TEMPLATES_ROOT_DIR}/ubuntu"
readonly MACOS_TEMPLATES_DIR="${TEMPLATES_ROOT_DIR}/macos"

# 生成配置文件
proj::template::generate_config() {
    local service_name="$1"
    local platform="$2"       # docker|ubuntu|macos
    local template_name="$3"   # config.yaml.tpl|systemd.service.tpl|etc.
    local output_path="$4"
    local extra_vars="${5:-}"  # 额外的环境变量文件路径
    
    if [[ -z "$service_name" ]] || [[ -z "$platform" ]] || [[ -z "$template_name" ]] || [[ -z "$output_path" ]]; then
        proj::log::error "Service name, platform, template name, and output path are required"
        return 1
    fi
    
    # 确定模板路径
    local template_dir
    case "$platform" in
        "docker")
            template_dir="$DOCKER_TEMPLATES_DIR"
            ;;
        "ubuntu")
            template_dir="$UBUNTU_TEMPLATES_DIR"
            ;;
        "macos")
            template_dir="$MACOS_TEMPLATES_DIR"
            ;;
        *)
            proj::log::error "Unsupported platform: $platform"
            return 1
            ;;
    esac
    
    local template_path="$template_dir/$service_name/$template_name"
    
    if [[ ! -f "$template_path" ]]; then
        proj::log::error "Template not found: $template_path"
        return 1
    fi
    
    proj::log::info "Generating config from template: $template_path -> $output_path"
    
    # 设置服务特定的环境变量
    # 特殊处理：如果模板名包含zookeeper，使用zookeeper服务配置
    local actual_service_name="$service_name"
    if [[ "$template_name" == *"zookeeper"* ]]; then
        actual_service_name="zookeeper"
    fi
    proj::template::setup_service_variables "$actual_service_name" "$platform"
    
    # 加载额外的变量文件
    if [[ -n "$extra_vars" ]] && [[ -f "$extra_vars" ]]; then
        proj::log::debug "Loading extra variables from: $extra_vars"
        # shellcheck disable=SC1090
        source "$extra_vars"
    fi
    
    # 确保输出目录存在
    local output_dir
    output_dir=$(dirname "$output_path")
    mkdir -p "$output_dir"
    
    # 使用 envsubst 进行变量替换
    if command -v envsubst >/dev/null 2>&1; then
        proj::template::envsubst_generate "$template_path" "$output_path"
    else
        proj::template::bash_generate "$template_path" "$output_path"
    fi
    
    proj::log::success "Configuration file generated: $output_path"
}

# 使用 envsubst 进行变量替换（推荐）
# 设置服务特定的环境变量
proj::template::setup_service_variables() {
    local service_name="$1"
    local platform="$2"
    
    # 设置通用变量
    export CONTAINER_NAME="${PROJ_PREFIX}-${service_name}"
    export DATA_VOLUME_NAME="${PROJ_PREFIX}-${service_name}-data"
    export LOGS_VOLUME_NAME="${PROJ_PREFIX}-${service_name}-logs"
    export NETWORK_NAME="${PROJ_NETWORK_NAME:-proj-network}"
    export SERVICE_PORT=""
    
    # 设置服务特定变量
    case "$service_name" in
        "kafka")
            export SERVICE_PORT="${PROJ_KAFKA_PORT:-9092}"
            export ZOOKEEPER_PORT="${PROJ_ZOOKEEPER_PORT:-2181}"
            export CONFIG_DIR="${PROJ_KAFKA_CONFIG_DIR}"
            export DATA_DIR="${PROJ_KAFKA_DATA_DIR}"
            export LOG_DIR="${PROJ_KAFKA_LOG_DIR}"
            export IMAGE_NAME="confluentinc/cp-kafka:${KAFKA_VERSION:-6.2.0}"
            
            # Kafka 特定变量
            export KAFKA_BROKER_ID="1"
            export KAFKA_ZOOKEEPER_CONNECT="${PROJ_PREFIX}-zookeeper:2181"
            export KAFKA_ADVERTISED_LISTENERS="PLAINTEXT://localhost:${PROJ_KAFKA_PORT:-9092},PLAINTEXT_INTERNAL://${PROJ_PREFIX}-kafka:29092"
            export KAFKA_LISTENER_SECURITY_PROTOCOL_MAP="PLAINTEXT:PLAINTEXT,PLAINTEXT_INTERNAL:PLAINTEXT"
            export KAFKA_INTER_BROKER_LISTENER_NAME="PLAINTEXT_INTERNAL"
            export KAFKA_VERSION="${KAFKA_VERSION:-3.6.0}"
            
            # Zookeeper 变量
            export ZOOKEEPER_CONTAINER_NAME="${PROJ_PREFIX}-zookeeper"
            export ZOOKEEPER_DATA_VOLUME_NAME="${PROJ_PREFIX}-zookeeper-data"
            export ZOOKEEPER_LOGS_VOLUME_NAME="${PROJ_PREFIX}-zookeeper-logs"
            export ZOOKEEPER_VERSION="${ZOOKEEPER_VERSION:-latest}"
            ;;
        "redis")
            export SERVICE_PORT="${PROJ_REDIS_PORT:-6379}"
            export CONFIG_DIR="${PROJ_REDIS_CONFIG_DIR}"
            export DATA_DIR="${PROJ_REDIS_DATA_DIR}"
            export IMAGE_NAME="redis:${REDIS_VERSION:-7.2.4}"
            ;;
        "mysql")
            export SERVICE_PORT="${PROJ_MYSQL_PORT:-3306}"
            export CONFIG_DIR="${PROJ_MYSQL_CONFIG_DIR}"
            export DATA_DIR="${PROJ_MYSQL_DATA_DIR}"
            export IMAGE_NAME="mysql:${MYSQL_VERSION:-8.0}"
            ;;
        "victorialogs")
            export SERVICE_PORT="${PROJ_VICTORIALOGS_PORT:-9428}"
            export CONFIG_DIR="${PROJ_VICTORIALOGS_CONFIG_DIR}"
            export DATA_DIR="${PROJ_VICTORIALOGS_DATA_DIR}"
            export LOG_DIR="${PROJ_VICTORIALOGS_LOG_DIR}"
            export IMAGE_NAME="victoriametrics/victoria-logs:v${VICTORIALOGS_VERSION:-1.28.0}"
            export VICTORIALOGS_RETENTION="${VICTORIALOGS_RETENTION:-7d}"
            ;;
        "etcd")
            export SERVICE_PORT="${PROJ_ETCD_PORT:-2379}"
            export CONFIG_DIR="${PROJ_ETCD_CONFIG_DIR}"
            export DATA_DIR="${PROJ_ETCD_DATA_DIR}"
            export LOG_DIR="${PROJ_ETCD_LOG_DIR}"
            export IMAGE_NAME="quay.io/coreos/etcd:${ETCD_VERSION:-v3.5.12}"
            export PEER_PORT="${ETCD_PEER_PORT:-2380}"
            export ETCD_NAME="${ETCD_NAME:-etcd0}"
            export ETCD_DATA_DIR="/etcd-data"
            export ETCD_LISTEN_CLIENT_URLS="http://0.0.0.0:2379"
            export ETCD_ADVERTISE_CLIENT_URLS="http://${PROJ_PREFIX}-etcd:2379"
            export ETCD_LISTEN_PEER_URLS="http://0.0.0.0:2380"
            export ETCD_INITIAL_ADVERTISE_PEER_URLS="http://${PROJ_PREFIX}-etcd:2380"
            export ETCD_INITIAL_CLUSTER="${ETCD_NAME:-etcd0}=http://${PROJ_PREFIX}-etcd:2380"
            export ETCD_INITIAL_CLUSTER_STATE="new"
            ;;
        "zookeeper")
            export SERVICE_PORT="${PROJ_ZOOKEEPER_PORT:-2181}"
            export CONFIG_DIR="${PROJ_ZOOKEEPER_CONFIG_DIR}"
            export DATA_DIR="${PROJ_ZOOKEEPER_DATA_DIR}"
            export LOG_DIR="${PROJ_ZOOKEEPER_LOG_DIR}"
            export IMAGE_NAME="confluentinc/cp-zookeeper:${ZOOKEEPER_VERSION:-latest}"
            export CONTAINER_NAME="${PROJ_PREFIX}-zookeeper"
            export DATA_VOLUME_NAME="${PROJ_PREFIX}-zookeeper-data"
            export LOGS_VOLUME_NAME="${PROJ_PREFIX}-zookeeper-logs"
            ;;
        "prometheus")
            export SERVICE_PORT="${PROJ_PROMETHEUS_PORT:-9090}"
            export CONFIG_DIR="${PROJ_PROMETHEUS_CONFIG_DIR}"
            export DATA_DIR="${PROJ_PROMETHEUS_DATA_DIR}"
            export LOG_DIR="${PROJ_PROMETHEUS_LOG_DIR}"
            export IMAGE_NAME="prom/prometheus:v${PROMETHEUS_VERSION:-2.48.1}"
            ;;
        "redis")
            export SERVICE_PORT="${PROJ_REDIS_PORT:-6379}"
            export CONFIG_DIR="${PROJ_REDIS_CONFIG_DIR}"
            export DATA_DIR="${PROJ_REDIS_DATA_DIR}"
            export IMAGE_NAME="redis:${REDIS_VERSION:-7.2.4}"
            ;;
        "mysql")
            export SERVICE_PORT="${PROJ_MYSQL_PORT:-3306}"
            export CONFIG_DIR="${PROJ_MYSQL_CONFIG_DIR}"
            export DATA_DIR="${PROJ_MYSQL_DATA_DIR}"
            export IMAGE_NAME="mysql:${MYSQL_VERSION:-8.0}"
            ;;
        "victorialogs")
            export SERVICE_PORT="${PROJ_VICTORIALOGS_PORT:-9428}"
            export CONFIG_DIR="${PROJ_VICTORIALOGS_CONFIG_DIR}"
            export DATA_DIR="${PROJ_VICTORIALOGS_DATA_DIR}"
            export LOG_DIR="${PROJ_VICTORIALOGS_LOG_DIR}"
            export IMAGE_NAME="victoriametrics/victoria-logs:v${VICTORIALOGS_VERSION:-1.28.0}"
            export VICTORIALOGS_RETENTION="${VICTORIALOGS_RETENTION:-7d}"
            ;;
        "etcd")
            export SERVICE_PORT="${PROJ_ETCD_PORT:-2379}"
            export CONFIG_DIR="${PROJ_ETCD_CONFIG_DIR}"
            export DATA_DIR="${PROJ_ETCD_DATA_DIR}"
            export LOG_DIR="${PROJ_ETCD_LOG_DIR}"
            export IMAGE_NAME="quay.io/coreos/etcd:${ETCD_VERSION:-v3.5.12}"
            export PEER_PORT="${ETCD_PEER_PORT:-2380}"
            export ETCD_NAME="${ETCD_NAME:-etcd0}"
            export ETCD_DATA_DIR="/etcd-data"
            export ETCD_LISTEN_CLIENT_URLS="http://0.0.0.0:2379"
            export ETCD_ADVERTISE_CLIENT_URLS="http://${PROJ_PREFIX}-etcd:2379"
            export ETCD_LISTEN_PEER_URLS="http://0.0.0.0:2380"
            export ETCD_INITIAL_ADVERTISE_PEER_URLS="http://${PROJ_PREFIX}-etcd:2380"
            export ETCD_INITIAL_CLUSTER="${ETCD_NAME:-etcd0}=http://${PROJ_PREFIX}-etcd:2380"
            export ETCD_INITIAL_CLUSTER_STATE="new"
            ;;
        "nacos")
            export SERVICE_PORT="${PROJ_NACOS_PORT:-8848}"
            export HTTP_PORT="${PROJ_NACOS_PORT:-8848}"
            export GRPC_PORT="${PROJ_NACOS_GRPC_PORT:-9848}"
            export RAFT_PORT="${PROJ_NACOS_RAFT_PORT:-9849}"
            export CONFIG_DIR="${PROJ_NACOS_CONFIG_DIR}"
            export DATA_DIR="${PROJ_NACOS_DATA_DIR}"
            export LOG_DIR="${PROJ_NACOS_LOG_DIR}"
            export IMAGE_NAME="nacos/nacos-server:v${NACOS_VERSION:-2.1.2}"
            export CONTAINER_NAME="${PROJ_PREFIX}-nacos"
            export DATA_VOLUME_NAME="${PROJ_PREFIX}-nacos-data"
            export LOGS_VOLUME_NAME="${PROJ_PREFIX}-nacos-logs"
            ;;
        "otel-collector")
            export CONFIG_DIR="${PROJ_OTELCOL_CONFIG_DIR}"
            export DATA_DIR="${PROJ_OTELCOL_DATA_DIR}"
            export LOG_DIR="${PROJ_OTELCOL_DATA_DIR}/logs"
            export CONTAINER_NAME="${CONTAINER_NAME_OTEL_COLLECTOR}"
            export IMAGE_NAME="otel/opentelemetry-collector-contrib:${OTELCOL_VERSION:-0.132.0}"
            export GRPC_PORT="${PROJ_OTELCOL_GRPC_PORT:-4317}"
            export HTTP_PORT="${PROJ_OTELCOL_HTTP_PORT:-4318}"
            export METRICS_PORT="${PROJ_OTELCOL_METRICS_PORT:-8888}"
            export HEALTH_PORT="${PROJ_OTELCOL_HEALTH_PORT:-13133}"
            export NETWORK_NAME="${PROJ_NETWORK_NAME}"
            export PROJ_JAEGER_OTLP_HOST="${PROJ_JAEGER_OTLP_HOST:-${PROJ_ACCESS_HOST}}"
            export PROJ_JAEGER_OTLP_PORT="${PROJ_JAEGER_OTLP_PORT:-4317}"
            export VICTORIALOGS_ENDPOINT="${PROJ_VICTORIALOGS_HOST}:${PROJ_VICTORIALOGS_PORT}"
            export JAEGER_ENDPOINT="${PROJ_JAEGER_OTLP_HOST}:${PROJ_JAEGER_OTLP_PORT}"
            export PROMETHEUS_ENDPOINT="${PROJ_PROMETHEUS_HOST}:${PROJ_PROMETHEUS_PORT}"
            export PROJ_SERVICE_NAME="${PROJ_SERVICE_NAME:-apiserver}"
            export PROJ_SERVICE_VERSION="${PROJ_SERVICE_VERSION:-v2.0.0}"
            export PROJ_ENVIRONMENT="${PROJ_ENVIRONMENT:-development}"
            export OTELCOL_VERSION="${OTELCOL_VERSION}"
            ;;
        "otel-agent")
            export CONFIG_DIR="${PROJ_OTEL_AGENT_CONFIG_DIR}"
            export DATA_DIR="${PROJ_OTEL_AGENT_DATA_DIR}"
            export LOG_DIR="${PROJ_OTEL_AGENT_DATA_DIR}/logs"
            export CONTAINER_NAME="${CONTAINER_NAME_OTEL_AGENT}"
            export IMAGE_NAME="otel/opentelemetry-collector-contrib:${OTELCOL_VERSION:-0.132.0}"
            export GRPC_PORT="${PROJ_OTEL_AGENT_GRPC_PORT:-4327}"
            export HTTP_PORT="${PROJ_OTEL_AGENT_HTTP_PORT:-4328}"
            export HEALTH_PORT="${PROJ_OTEL_AGENT_HEALTH_PORT:-13134}"
            export NETWORK_NAME="${PROJ_NETWORK_NAME}"
            export COLLECTOR_CONTAINER="${CONTAINER_NAME_OTEL_COLLECTOR}"
            export COLLECTOR_ENDPOINT="${CONTAINER_NAME_OTEL_COLLECTOR}:4317"
            export OTEL_COLLECTOR_ENDPOINT="${CONTAINER_NAME_OTEL_COLLECTOR}:4317"
            export PROJ_SERVICE_NAME="${PROJ_SERVICE_NAME:-apiserver}"
            export PROJ_SERVICE_VERSION="${PROJ_SERVICE_VERSION:-v2.0.0}"
            export PROJ_ENVIRONMENT="${PROJ_ENVIRONMENT:-development}"
            export OTEL_AGENT_VERSION="${OTEL_AGENT_VERSION}"
            ;;
        *)
            proj::log::debug "No specific variables set for service: $service_name"
            ;;
    esac
}

proj::template::envsubst_generate() {
    local template_path="$1"
    local output_path="$2"
    
    # 创建临时文件
    local temp_file="/tmp/template_$$_$(date +%s).tmp"
    cp "$template_path" "$temp_file"
    
    # 方法：先处理带默认值的变量，将它们替换为简单变量
    # 例如：${PROJ_SERVICE_NAME:-apiserver} -> ${PROJ_SERVICE_NAME}
    # 然后为未设置的变量设置默认值
    
    # 查找所有带默认值的变量
    while IFS= read -r var_expr; do
        # 提取变量名和默认值
        var_name=$(echo "$var_expr" | sed 's/\${//' | sed 's/:[-][^}]*//' | sed 's/}//')
        default_val=$(echo "$var_expr" | grep -oE ':[-][^}]*' | sed 's/:-//')
        
        # 如果变量未设置，设置为默认值
        if [[ -z "${!var_name:-}" ]]; then
            export "$var_name=$default_val"
            proj::log::debug "Set default: $var_name=$default_val"
        fi
        
        # 将带默认值的语法替换为简单语法
        # 使用 | 作为sed分隔符，避免与路径中的/冲突
        sed -i '' "s|\${${var_name}:[-]${default_val}}|\${${var_name}}|g" "$temp_file"
    done < <(grep -oE '\$\{[A-Za-z_][A-Za-z0-9_]*:[-][^}]*\}' "$template_path")
    
    # 现在使用envsubst处理所有简单变量
    local simple_vars=$(grep -oE '\$\{[A-Za-z_][A-Za-z0-9_]*\}' "$temp_file" | sort -u | tr '\n' ' ')
    
    if [[ -n "$simple_vars" ]]; then
        proj::log::debug "Processing template variables: $simple_vars"
        envsubst "$simple_vars" < "$temp_file" > "$output_path"
    else
        cp "$temp_file" "$output_path"
    fi
    
    # 清理临时文件
    rm -f "$temp_file"
}

# 使用 bash 进行变量替换（备用方案）
proj::template::bash_generate() {
    local template_path="$1"
    local output_path="$2"
    
    proj::log::warn "envsubst not available, using bash variable substitution"
    
    # 读取模板内容并进行变量替换
    local content
    content=$(<"$template_path")
    
    # 使用 eval 进行变量替换（注意安全性）
    content=$(eval "cat << 'TEMPLATE_EOF'
$content
TEMPLATE_EOF")
    
    # 写入输出文件
    echo "$content" > "$output_path"
}

# 批量生成服务配置
proj::template::generate_service_configs() {
    local service_name="$1"
    local platform="$2"
    local config_dir="$3"
    local extra_vars="${4:-}"
    
    if [[ -z "$service_name" ]] || [[ -z "$platform" ]] || [[ -z "$config_dir" ]]; then
        proj::log::error "Service name, platform, and config directory are required"
        return 1
    fi
    
    proj::log::info "Generating all configuration files for $service_name on $platform"
    
    # 确定模板目录
    local template_dir
    case "$platform" in
        "docker")
            template_dir="$DOCKER_TEMPLATES_DIR/$service_name"
            ;;
        "ubuntu")
            template_dir="$UBUNTU_TEMPLATES_DIR/$service_name"
            ;;
        "macos")
            template_dir="$MACOS_TEMPLATES_DIR/$service_name"
            ;;
        *)
            proj::log::error "Unsupported platform: $platform"
            return 1
            ;;
    esac
    
    if [[ ! -d "$template_dir" ]]; then
        proj::log::error "Template directory not found: $template_dir"
        return 1
    fi
    
    # 生成所有模板文件
    local generated_count=0
    while IFS= read -r -d '' template_file; do
        local template_name
        template_name=$(basename "$template_file")
        
        # 移除 .tpl 后缀
        local output_name="${template_name%.tpl}"
        local output_path="$config_dir/$output_name"
        
        if proj::template::generate_config "$service_name" "$platform" "$template_name" "$output_path" "$extra_vars"; then
            ((generated_count++))
        fi
    done < <(find "$template_dir" -name "*.tpl" -type f -print0)
    
    proj::log::success "Generated $generated_count configuration files for $service_name"
}

# 验证模板文件
proj::template::validate_template() {
    local template_path="$1"
    
    if [[ ! -f "$template_path" ]]; then
        proj::log::error "Template file does not exist: $template_path"
        return 1
    fi
    
    proj::log::info "Validating template: $template_path"
    
    # 检查模板语法
    local template_vars
    template_vars=$(grep -oE '\$\{[A-Za-z_][A-Za-z0-9_]*[^}]*\}' "$template_path" || true)
    
    if [[ -n "$template_vars" ]]; then
        proj::log::debug "Found template variables:"
        echo "$template_vars" | while read -r var; do
            proj::log::debug "  $var"
        done
    fi
    
    # 检查未闭合的变量引用
    if grep -q '\${[^}]*$' "$template_path"; then
        proj::log::error "Found unclosed variable references in template"
        grep -n '\${[^}]*$' "$template_path" | while read -r line; do
            proj::log::error "  Line: $line"
        done
        return 1
    fi
    
    proj::log::success "Template validation passed: $template_path"
}

# 列出可用的模板
proj::template::list_templates() {
    local platform="${1:-}"
    local service="${2:-}"
    
    if [[ -z "$platform" ]]; then
        proj::log::info "Available platforms:"
        for platform_dir in "$DOCKER_TEMPLATES_DIR" "$UBUNTU_TEMPLATES_DIR" "$MACOS_TEMPLATES_DIR"; do
            if [[ -d "$platform_dir" ]]; then
                local platform_name
                platform_name=$(basename "$platform_dir")
                echo "  $platform_name"
            fi
        done
        return 0
    fi
    
    # 确定模板目录
    local template_base_dir
    case "$platform" in
        "docker")
            template_base_dir="$DOCKER_TEMPLATES_DIR"
            ;;
        "ubuntu")
            template_base_dir="$UBUNTU_TEMPLATES_DIR"
            ;;
        "macos")
            template_base_dir="$MACOS_TEMPLATES_DIR"
            ;;
        *)
            proj::log::error "Unsupported platform: $platform"
            return 1
            ;;
    esac
    
    if [[ ! -d "$template_base_dir" ]]; then
        proj::log::error "Template directory not found: $template_base_dir"
        return 1
    fi
    
    if [[ -z "$service" ]]; then
        proj::log::info "Available services for platform '$platform':"
        for service_dir in "$template_base_dir"/*; do
            if [[ -d "$service_dir" ]]; then
                local service_name
                service_name=$(basename "$service_dir")
                echo "  $service_name"
            fi
        done
        return 0
    fi
    
    # 列出特定服务的模板
    local service_template_dir="$template_base_dir/$service"
    if [[ ! -d "$service_template_dir" ]]; then
        proj::log::error "Service template directory not found: $service_template_dir"
        return 1
    fi
    
    proj::log::info "Available templates for service '$service' on platform '$platform':"
    find "$service_template_dir" -name "*.tpl" -type f | while read -r template; do
        local template_name
        template_name=$(basename "$template")
        echo "  $template_name"
    done
}

# 创建新模板
proj::template::create_template() {
    local service_name="$1"
    local platform="$2"
    local template_name="$3"
    local template_content="${4:-}"
    
    if [[ -z "$service_name" ]] || [[ -z "$platform" ]] || [[ -z "$template_name" ]]; then
        proj::log::error "Service name, platform, and template name are required"
        return 1
    fi
    
    # 确定模板目录
    local template_dir
    case "$platform" in
        "docker")
            template_dir="$DOCKER_TEMPLATES_DIR/$service_name"
            ;;
        "ubuntu")
            template_dir="$UBUNTU_TEMPLATES_DIR/$service_name"
            ;;
        "macos")
            template_dir="$MACOS_TEMPLATES_DIR/$service_name"
            ;;
        *)
            proj::log::error "Unsupported platform: $platform"
            return 1
            ;;
    esac
    
    # 确保模板目录存在
    mkdir -p "$template_dir"
    
    local template_path="$template_dir/$template_name"
    
    if [[ -f "$template_path" ]]; then
        proj::log::error "Template already exists: $template_path"
        return 1
    fi
    
    # 创建模板文件
    if [[ -n "$template_content" ]]; then
        echo "$template_content" > "$template_path"
    else
        # 创建基本模板结构
        cat > "$template_path" << EOF
# Configuration template for $service_name on $platform
# Project: \${PROJ_NAME:-go-protoc}
# Service: $service_name \${$(echo "$service_name" | tr '[:lower:]' '[:upper:]')_VERSION}
# Environment: \${PROJ_ENVIRONMENT:-development}

# Add your configuration here
EOF
    fi
    
    proj::log::success "Template created: $template_path"
}

# 标记已加载
export TEMPLATE_MANAGER_LIB_LOADED=true

# 如果直接执行此脚本，显示可用模板
if [[ "${BASH_SOURCE[0]:-${0##*/}}" == "${0##*/}" ]]; then
    proj::template::list_templates "$@"
fi