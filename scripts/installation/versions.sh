#!/usr/bin/env bash

# =============================================================================
# 统一版本管理配置文件
# Central Version Management Configuration
#
# 此文件管理所有第三方组件的版本信息，确保版本统一和易于升级
# This file manages version information for all third-party components
# =============================================================================

set -eEuo pipefail

# =============================================================================
# 项目基础配置 (Project Base Configuration)
# =============================================================================

# 项目基础信息
export PROJ_NAME=${PROJ_NAME:-go-protoc}
export PROJ_PREFIX=${PROJ_PREFIX:-proj}
export PROJ_ENVIRONMENT=${PROJ_ENVIRONMENT:-development}
export PROJ_NAMESPACE=${PROJ_NAMESPACE:-default}

# Docker网络配置
export PROJ_NETWORK_NAME=${PROJ_NETWORK_NAME:-${PROJ_PREFIX}-network}

# =============================================================================
# 基础设施组件版本 (Infrastructure Component Versions)
# =============================================================================

# 数据库相关 (Database)
export REDIS_VERSION=${REDIS_VERSION:-7.2.4}
export MARIADB_VERSION=${MARIADB_VERSION:-11.2.2}
export MYSQL_VERSION=${MYSQL_VERSION:-8.0}
export MONGODB_VERSION=${MONGODB_VERSION:-7.0.5}

# 分布式系统 (Distributed Systems)
export ETCD_VERSION=${ETCD_VERSION:-v3.5.12}

# Kafka 版本配置 - 根据平台自动选择
# macOS/ARM64 使用 7.4.0，Linux/AMD64 使用 6.2.0
if [[ "$(uname)" == "Darwin" ]]; then
    export KAFKA_VERSION=${KAFKA_VERSION:-7.4.0}
else
    export KAFKA_VERSION=${KAFKA_VERSION:-6.2.0}
fi

export ZOOKEEPER_VERSION=${ZOOKEEPER_VERSION:-latest}
export NACOS_VERSION=${NACOS_VERSION:-v2.1.2}

# 可观测性栈 (Observability Stack)
export JAEGER_VERSION=${JAEGER_VERSION:-1.52.0}
export PROMETHEUS_VERSION=${PROMETHEUS_VERSION:-2.48.1}
export GRAFANA_VERSION=${GRAFANA_VERSION:-10.2.4}
export ALERTMANAGER_VERSION=${ALERTMANAGER_VERSION:-0.26.0}
export OTELCOL_VERSION=${OTELCOL_VERSION:-0.132.0}
export OTEL_VERSION=${OTEL_VERSION:-0.132.0}
export PYROSCOPE_VERSION=${PYROSCOPE_VERSION:-1.9.0}
export SENTRY_VERSION=${SENTRY_VERSION:-latest}
export LOKI_VERSION=${LOKI_VERSION:-3.0.0}

# 日志管理 (Log Management)
export VICTORIALOGS_VERSION=${VICTORIALOGS_VERSION:-1.28.0}

# VictoriaMetrics 监控栈 (VictoriaMetrics Stack)
export VICTORIAMETRICS_VERSION=${VICTORIAMETRICS_VERSION:-1.96.0}
export VMAGENT_VERSION=${VMAGENT_VERSION:-1.96.0}
export VMALERT_VERSION=${VMALERT_VERSION:-1.96.0}

# 工具版本 (Tools)
export DOCKER_COMPOSE_VERSION=${DOCKER_COMPOSE_VERSION:-v2.29.7}

# =============================================================================
# 配置目录管理 (Configuration Directory Management)
# =============================================================================

# 获取项目根目录
if [[ -n "${BASH_SOURCE[0]:-}" ]]; then
    PROJ_ROOT_DIR="${PROJ_ROOT_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
else
    # 当通过envsubst等方式调用时的回退路径，假设在项目根目录下
    PROJ_ROOT_DIR="${PROJ_ROOT_DIR:-/Users/costalong/code/go/src/github.com/costa92/go-protoc}"
fi

# 第三方服务安装目录配置
export PROJ_THIRDPARTY_INSTALL_DIR=${PROJ_THIRDPARTY_INSTALL_DIR:-"${PROJ_ROOT_DIR}/_thirdparty"}

# 数据库服务配置目录
export PROJ_REDIS_CONFIG_DIR=${PROJ_REDIS_CONFIG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/redis/config"}
export PROJ_REDIS_DATA_DIR=${PROJ_REDIS_DATA_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/redis/data"}
export PROJ_REDIS_LOG_DIR=${PROJ_REDIS_LOG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/redis/logs"}
export PROJ_MYSQL_CONFIG_DIR=${PROJ_MYSQL_CONFIG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/mysql/config"}
export PROJ_MYSQL_DATA_DIR=${PROJ_MYSQL_DATA_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/mysql/data"}
export PROJ_MYSQL_LOG_DIR=${PROJ_MYSQL_LOG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/mysql/logs"}
export PROJ_MARIADB_CONFIG_DIR=${PROJ_MARIADB_CONFIG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/mariadb/config"}
export PROJ_MARIADB_DATA_DIR=${PROJ_MARIADB_DATA_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/mariadb/data"}
export PROJ_MARIADB_LOG_DIR=${PROJ_MARIADB_LOG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/mariadb/logs"}
export PROJ_MONGODB_CONFIG_DIR=${PROJ_MONGODB_CONFIG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/mongodb/config"}
export PROJ_MONGODB_DATA_DIR=${PROJ_MONGODB_DATA_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/mongodb/data"}
export PROJ_MONGODB_LOG_DIR=${PROJ_MONGODB_LOG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/mongodb/logs"}

# etcd 服务配置目录
export PROJ_ETCD_CONFIG_DIR=${PROJ_ETCD_CONFIG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/etcd/config"}
export PROJ_ETCD_DATA_DIR=${PROJ_ETCD_DATA_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/etcd/data"}
export PROJ_ETCD_LOG_DIR=${PROJ_ETCD_LOG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/etcd/logs"}

# 分布式系统配置目录
export PROJ_NACOS_CONFIG_DIR=${PROJ_NACOS_CONFIG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/nacos/config"}
export PROJ_NACOS_DATA_DIR=${PROJ_NACOS_DATA_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/nacos/data"}
export PROJ_NACOS_LOG_DIR=${PROJ_NACOS_LOG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/nacos/logs"}
export PROJ_KAFKA_CONFIG_DIR=${PROJ_KAFKA_CONFIG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/kafka/config"}
export PROJ_KAFKA_DATA_DIR=${PROJ_KAFKA_DATA_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/kafka/data"}
export PROJ_KAFKA_LOG_DIR=${PROJ_KAFKA_LOG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/kafka/logs"}
export PROJ_ZOOKEEPER_CONFIG_DIR=${PROJ_ZOOKEEPER_CONFIG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/zookeeper/config"}
export PROJ_ZOOKEEPER_DATA_DIR=${PROJ_ZOOKEEPER_DATA_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/zookeeper/data"}
export PROJ_ZOOKEEPER_LOG_DIR=${PROJ_ZOOKEEPER_LOG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/zookeeper/logs"}

# 可观测性服务配置目录
export PROJ_OTEL_AGENT_CONFIG_DIR=${PROJ_OTEL_AGENT_CONFIG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/otel-agent/config"}
export PROJ_OTEL_AGENT_DATA_DIR=${PROJ_OTEL_AGENT_DATA_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/otel-agent/data"}
export PROJ_OTEL_AGENT_LOG_DIR=${PROJ_OTEL_AGENT_LOG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/otel-agent/logs"}
export PROJ_OTELCOL_CONFIG_DIR=${PROJ_OTELCOL_CONFIG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/otelcol/config"}
export PROJ_OTELCOL_DATA_DIR=${PROJ_OTELCOL_DATA_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/otelcol/data"}
export PROJ_OTELCOL_LOG_DIR=${PROJ_OTELCOL_LOG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/otelcol/logs"}
export PROJ_JAEGER_CONFIG_DIR=${PROJ_JAEGER_CONFIG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/jaeger/config"}
export PROJ_JAEGER_DATA_DIR=${PROJ_JAEGER_DATA_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/jaeger/data"}
export PROJ_JAEGER_LOG_DIR=${PROJ_JAEGER_LOG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/jaeger/logs"}
export PROJ_PROMETHEUS_CONFIG_DIR=${PROJ_PROMETHEUS_CONFIG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/prometheus/config"}
export PROJ_PROMETHEUS_DATA_DIR=${PROJ_PROMETHEUS_DATA_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/prometheus/data"}
export PROJ_PROMETHEUS_LOG_DIR=${PROJ_PROMETHEUS_LOG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/prometheus/logs"}
export PROJ_GRAFANA_CONFIG_DIR=${PROJ_GRAFANA_CONFIG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/grafana/config"}
export PROJ_GRAFANA_DATA_DIR=${PROJ_GRAFANA_DATA_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/grafana/data"}
export PROJ_GRAFANA_LOG_DIR=${PROJ_GRAFANA_LOG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/grafana/logs"}

# 日志管理服务配置目录
export PROJ_VICTORIALOGS_CONFIG_DIR=${PROJ_VICTORIALOGS_CONFIG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/victorialogs/config"}
export PROJ_VICTORIALOGS_DATA_DIR=${PROJ_VICTORIALOGS_DATA_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/victorialogs/data"}
export PROJ_VICTORIALOGS_LOG_DIR=${PROJ_VICTORIALOGS_LOG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/victorialogs/logs"}
export PROJ_LOKI_CONFIG_DIR=${PROJ_LOKI_CONFIG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/loki/config"}
export PROJ_LOKI_DATA_DIR=${PROJ_LOKI_DATA_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/loki/data"}
export PROJ_LOKI_LOG_DIR=${PROJ_LOKI_LOG_DIR:-"${PROJ_THIRDPARTY_INSTALL_DIR}/loki/logs"}


# 服务端口配置
export PROJ_REDIS_PORT=${PROJ_REDIS_PORT:-6379}
export PROJ_MYSQL_PORT=${PROJ_MYSQL_PORT:-3306}
export PROJ_MARIADB_PORT=${PROJ_MARIADB_PORT:-3307}
export PROJ_MONGODB_PORT=${PROJ_MONGODB_PORT:-27017}
export PROJ_ETCD_PORT=${PROJ_ETCD_PORT:-2379}
export PROJ_NACOS_PORT=${PROJ_NACOS_PORT:-8848}
export PROJ_NACOS_GRPC_PORT=${PROJ_NACOS_GRPC_PORT:-9848}
export PROJ_KAFKA_PORT=${PROJ_KAFKA_PORT:-9092}
export PROJ_ZOOKEEPER_PORT=${PROJ_ZOOKEEPER_PORT:-2181}
export PROJ_OTELCOL_GRPC_PORT=${PROJ_OTELCOL_GRPC_PORT:-4327}
export PROJ_OTELCOL_HTTP_PORT=${PROJ_OTELCOL_HTTP_PORT:-4328}
export PROJ_JAEGER_PORT=${PROJ_JAEGER_PORT:-14268}
export PROJ_JAEGER_UI_PORT=${PROJ_JAEGER_UI_PORT:-16686}
export PROJ_PROMETHEUS_PORT=${PROJ_PROMETHEUS_PORT:-9090}
export PROJ_GRAFANA_PORT=${PROJ_GRAFANA_PORT:-3000}
export PROJ_VICTORIALOGS_PORT=${PROJ_VICTORIALOGS_PORT:-9428}

# =============================================================================
# 平台安装偏好配置 (Platform Installation Preferences)
# =============================================================================

# 每个服务在不同平台上的推荐安装方式 (docker|native)
# Redis 偏好配置
export INSTALL_PREFERENCE_REDIS_UBUNTU=${INSTALL_PREFERENCE_REDIS_UBUNTU:-"docker"}
export INSTALL_PREFERENCE_REDIS_MACOS=${INSTALL_PREFERENCE_REDIS_MACOS:-"native"}

# OTEL Collector 偏好配置
export INSTALL_PREFERENCE_OTELCOL_UBUNTU=${INSTALL_PREFERENCE_OTELCOL_UBUNTU:-"native"}
export INSTALL_PREFERENCE_OTELCOL_MACOS=${INSTALL_PREFERENCE_OTELCOL_MACOS:-"docker"}

# Prometheus 偏好配置
export INSTALL_PREFERENCE_PROMETHEUS_UBUNTU=${INSTALL_PREFERENCE_PROMETHEUS_UBUNTU:-"docker"}
export INSTALL_PREFERENCE_PROMETHEUS_MACOS=${INSTALL_PREFERENCE_PROMETHEUS_MACOS:-"native"}

# Grafana 偏好配置
export INSTALL_PREFERENCE_GRAFANA_UBUNTU=${INSTALL_PREFERENCE_GRAFANA_UBUNTU:-"docker"}
export INSTALL_PREFERENCE_GRAFANA_MACOS=${INSTALL_PREFERENCE_GRAFANA_MACOS:-"native"}

# Jaeger 偏好配置
export INSTALL_PREFERENCE_JAEGER_UBUNTU=${INSTALL_PREFERENCE_JAEGER_UBUNTU:-"docker"}
export INSTALL_PREFERENCE_JAEGER_MACOS=${INSTALL_PREFERENCE_JAEGER_MACOS:-"docker"}

# VictoriaLogs 偏好配置
export INSTALL_PREFERENCE_VICTORIALOGS_UBUNTU=${INSTALL_PREFERENCE_VICTORIALOGS_UBUNTU:-"native"}
export INSTALL_PREFERENCE_VICTORIALOGS_MACOS=${INSTALL_PREFERENCE_VICTORIALOGS_MACOS:-"docker"}

# =============================================================================
# 平台特定包名和公式映射 (Platform-specific Package/Formula Mapping)
# =============================================================================

# 获取Ubuntu包名的函数
proj::versions::get_ubuntu_package() {
    local service="$1"
    case "$service" in
        "redis") echo "redis-server" ;;
        "mysql") echo "mysql-server" ;;
        "mariadb") echo "mariadb-server" ;;
        "mongodb") echo "mongodb" ;;
        "prometheus") echo "prometheus" ;;
        "grafana") echo "grafana" ;;
        "nginx") echo "nginx" ;;
        "postgresql") echo "postgresql" ;;
        *) echo "$service" ;;
    esac
}

# 获取macOS Homebrew公式的函数
proj::versions::get_macos_formula() {
    local service="$1"
    case "$service" in
        "redis") echo "redis" ;;
        "mysql") echo "mysql" ;;
        "mariadb") echo "mariadb" ;;
        "mongodb") echo "mongodb" ;;
        "prometheus") echo "prometheus" ;;
        "grafana") echo "grafana" ;;
        "nginx") echo "nginx" ;;
        "postgresql") echo "postgresql" ;;
        "node") echo "node" ;;
        "python") echo "python" ;;
        *) echo "$service" ;;
    esac
}

# =============================================================================
# 版本兼容性映射 (Version Compatibility Mapping)
# =============================================================================

# 为了向后兼容，保持原有的环境变量名称
export PROJ_ETCD_VERSION=${ETCD_VERSION}
export PROJ_JAEGER_VERSION=${JAEGER_VERSION}
export PROJ_GRAFANA_VERSION=${GRAFANA_VERSION}
export PROJ_PROMETHEUS_VERSION=${PROMETHEUS_VERSION}
export PROJ_ALERTMANAGER_VERSION=${ALERTMANAGER_VERSION}
export PROJ_OTELCOL_VERSION=${OTELCOL_VERSION}
export PROJ_OTEL_VERSION=${OTEL_VERSION}
export PROJ_PYROSCOPE_VERSION=${PYROSCOPE_VERSION}
export PROJ_VICTORIALOGS_VERSION=${VICTORIALOGS_VERSION}
export PROJ_LOKI_VERSION=${LOKI_VERSION}
export PROJ_KAFKA_VERSION=${KAFKA_VERSION}
export PROJ_ZOOKEEPER_VERSION=${ZOOKEEPER_VERSION}
export PROJ_REDIS_VERSION=${REDIS_VERSION}

# =============================================================================
# 容器名称管理函数 (Container Name Management Functions)
# =============================================================================

# 获取服务的容器名称
proj::versions::get_container_name() {
  local service_name="$1"
  echo "${PROJ_PREFIX}-${service_name}"
}

# 获取服务的数据卷名称
proj::versions::get_volume_name() {
  local service_name="$1"
  local volume_type="${2:-data}"  # data, logs, config
  echo "${PROJ_PREFIX}-${service_name}-${volume_type}"
}

# 获取服务的网络名称
proj::versions::get_network_name() {
  echo "${PROJ_NETWORK_NAME}"
}

# =============================================================================
# 版本检查函数 (Version Check Functions)
# =============================================================================

# 检查版本格式是否正确
proj::versions::validate_version() {
  local version=$1
  local component=$2

  if [[ -z "$version" ]]; then
    echo "错误: $component 版本不能为空" >&2
    return 1
  fi

  # 基本的版本格式检查（支持 v前缀、不带v的版本和latest标签）
  if [[ ! "$version" =~ ^(latest|v?[0-9]+\.[0-9]+(\.[0-9]+)?(-[a-zA-Z0-9\-]+)?)$ ]]; then
    echo "警告: $component 版本格式可能不正确: $version" >&2
  fi
}

# 显示所有版本信息
proj::versions::show_all() {
  echo "=== 第三方组件版本信息 ==="
  echo "数据库组件:"
  echo "  Redis:        $REDIS_VERSION"
  echo "  MariaDB:      $MARIADB_VERSION"
  echo "  MySQL:        $MYSQL_VERSION"
  echo "  MongoDB:      $MONGODB_VERSION"
  echo ""
  echo "分布式系统:"
  echo "  etcd:         $ETCD_VERSION"
  echo "  Kafka:        $KAFKA_VERSION"
  echo "  Zookeeper:    $ZOOKEEPER_VERSION"
  echo ""
  echo "可观测性栈:"
  echo "  Jaeger:       $JAEGER_VERSION"
  echo "  Prometheus:   $PROMETHEUS_VERSION"
  echo "  Grafana:      $GRAFANA_VERSION"
  echo "  AlertManager: $ALERTMANAGER_VERSION"
  echo "  OtelCol:      $OTELCOL_VERSION"
  echo "  OTEL:         $OTEL_VERSION"
  echo "  Pyroscope:    $PYROSCOPE_VERSION"
  echo ""
  echo "日志管理:"
  echo "  VictoriaLogs: $VICTORIALOGS_VERSION"
  echo "  Grafana Loki: $LOKI_VERSION"
  echo ""
  echo "VictoriaMetrics 监控栈:"
  echo "  VictoriaMetrics: $VICTORIAMETRICS_VERSION"
  echo "  vmagent:         $VMAGENT_VERSION"
  echo "  vmalert:         $VMALERT_VERSION"
  echo ""
  echo "工具:"
  echo "  Docker Compose: $DOCKER_COMPOSE_VERSION"
}

# 检查所有版本
proj::versions::validate_all() {
  echo "正在验证版本格式..."

  proj::versions::validate_version "$REDIS_VERSION" "Redis"
  proj::versions::validate_version "$MARIADB_VERSION" "MariaDB"
  proj::versions::validate_version "$MYSQL_VERSION" "MySQL"
  proj::versions::validate_version "$MONGODB_VERSION" "MongoDB"
  proj::versions::validate_version "$ETCD_VERSION" "etcd"
  proj::versions::validate_version "$KAFKA_VERSION" "Kafka"
  proj::versions::validate_version "$ZOOKEEPER_VERSION" "Zookeeper"
  proj::versions::validate_version "$JAEGER_VERSION" "Jaeger"
  proj::versions::validate_version "$PROMETHEUS_VERSION" "Prometheus"
  proj::versions::validate_version "$GRAFANA_VERSION" "Grafana"
  proj::versions::validate_version "$ALERTMANAGER_VERSION" "AlertManager"
  proj::versions::validate_version "$OTELCOL_VERSION" "OpenTelemetry Collector"
  proj::versions::validate_version "$OTEL_VERSION" "OTEL"
  proj::versions::validate_version "$PYROSCOPE_VERSION" "Pyroscope"
  proj::versions::validate_version "$VICTORIALOGS_VERSION" "VictoriaLogs"
  proj::versions::validate_version "$LOKI_VERSION" "Grafana Loki"
  proj::versions::validate_version "$VICTORIAMETRICS_VERSION" "VictoriaMetrics"
  proj::versions::validate_version "$VMAGENT_VERSION" "vmagent"
  proj::versions::validate_version "$VMALERT_VERSION" "vmalert"
  proj::versions::validate_version "$DOCKER_COMPOSE_VERSION" "Docker Compose"

  echo "版本验证完成"
}

# =============================================================================
# 使用示例 (Usage Examples)
# =============================================================================
#
# 在其他脚本中使用:
# source ./versions.sh
# echo "使用 Prometheus 版本: $PROMETHEUS_VERSION"
#
# 命令行使用:
# ./versions.sh show      # 显示所有版本
# ./versions.sh validate  # 验证所有版本格式
# =============================================================================

# 如果直接执行此脚本 - 兼容 Make 环境
if [[ "${BASH_SOURCE[0]:-${0##*/}}" == "${0##*/}" ]]; then
  case "${1:-}" in
    "show"|"list"|"")
      proj::versions::show_all
      ;;
    "validate"|"check")
      proj::versions::validate_all
      ;;
    *)
      echo "用法: $0 [show|validate]"
      echo "  show     - 显示所有版本信息"
      echo "  validate - 验证版本格式"
      exit 1
      ;;
  esac
fi

# 标记已加载
VERSIONS_LOADED=true