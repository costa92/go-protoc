# Prometheus基础配置模板 - 简化版
# Project: ${PROJ_NAME:-go-protoc}
# Service: Prometheus ${PROMETHEUS_VERSION}

global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: 'go-protoc'
    environment: '${PROJ_ENVIRONMENT:-development}'

rule_files:
  - "/etc/prometheus/rules/*.yml"

scrape_configs:
  # Prometheus self-monitoring
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
        labels:
          service: 'prometheus'
          version: '${PROMETHEUS_VERSION}'

  # Application API server
  - job_name: 'go-protoc-api'
    static_configs:
      - targets: ['${PROJ_PREFIX}-apiserver:8080']
        labels:
          service: 'apiserver'
          environment: '${PROJ_ENVIRONMENT:-development}'
    metrics_path: '/metrics'
    scrape_interval: 30s

# Alertmanager configuration  
alerting:
  alertmanagers:
    - static_configs:
        - targets:
          # - alertmanager:9093

# Note: Additional scrape configs for Redis/MySQL exporters 
# are dynamically added by the docker-run.sh script