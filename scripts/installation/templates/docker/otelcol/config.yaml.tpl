# OpenTelemetry Collector Configuration Template for Docker
# Project: ${PROJ_NAME:-go-protoc}
# Service: OpenTelemetry Collector ${OTELCOL_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

receivers:
  # OTLP receivers for telemetry data
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

  # File receiver for log files
  filelog:
    include:
      - /host/logs/**/*.log
      - /host/logs/**/*.json
    exclude:
      - /host/logs/**/*test*.log
    operators:
      - type: json_parser
        id: json_parser
        if: 'body matches "^\\{"'
      - type: move
        from: attributes.msg
        to: attributes._msg
      - type: add
        field: attributes.source
        value: file_receiver

processors:
  batch:
    timeout: 5s
    send_batch_size: 1024
    send_batch_max_size: 2048

  resource:
    attributes:
      - key: service.name
        value: ${PROJ_SERVICE_NAME:-apiserver}
        action: upsert
      - key: service.version
        value: ${PROJ_SERVICE_VERSION:-v2.0.0}
        action: upsert
      - key: deployment.environment
        value: ${PROJ_ENVIRONMENT:-development}
        action: upsert
      - key: collector.version
        value: ${OTELCOL_VERSION}
        action: upsert

  memory_limiter:
    limit_mib: 512

exporters:
  # VictoriaLogs for log data
  loki:
    endpoint: "http://victorialogs:9428/insert/loki/api/v1/push"
    headers:
      "Content-Type": "application/json"
    tenant_id: "proj"

  # Prometheus for metrics (if needed)
  prometheus:
    endpoint: "0.0.0.0:8888"
    namespace: "otelcol"
    const_labels:
      service: "${PROJ_SERVICE_NAME:-apiserver}"
      version: "${PROJ_SERVICE_VERSION:-v2.0.0}"

  # Jaeger for traces (if needed)
  jaeger:
    endpoint: "http://jaeger:14268/api/traces"

  # Debug exporter for troubleshooting
  debug:
    verbosity: basic

service:
  extensions: []
  pipelines:
    logs:
      receivers: [otlp, filelog]
      processors: [memory_limiter, resource, batch]
      exporters: [loki, debug]
    
    metrics:
      receivers: [otlp]
      processors: [memory_limiter, resource, batch]
      exporters: [prometheus]
    
    traces:
      receivers: [otlp]
      processors: [memory_limiter, resource, batch]
      exporters: [jaeger]

  telemetry:
    logs:
      level: info
    metrics:
      address: 0.0.0.0:8888