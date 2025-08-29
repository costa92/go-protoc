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
# 基础设施组件版本 (Infrastructure Component Versions)
# =============================================================================

# 数据库相关 (Database)
export REDIS_VERSION=${REDIS_VERSION:-7.2.4}
export MARIADB_VERSION=${MARIADB_VERSION:-11.2.2}
export MYSQL_VERSION=${MYSQL_VERSION:-8.0}
export MONGODB_VERSION=${MONGODB_VERSION:-7.0.5}

# 分布式系统 (Distributed Systems)
export ETCD_VERSION=${ETCD_VERSION:-v3.5.12}
export KAFKA_VERSION=${KAFKA_VERSION:-3.6}
export ZOOKEEPER_VERSION=${ZOOKEEPER_VERSION:-latest}

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
if [[ "${BASH_SOURCE[0]:-${0}}" == "${0}" ]]; then
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