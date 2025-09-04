# OTEL Agent Configuration Template
# Role: 接收应用程序的OTLP数据，进行轻量级处理后转发给Collector
# Flow: Application → Agent → Collector
# Environment Variables: OTEL_COLLECTOR_ENDPOINT, PROJ_SERVICE_NAME, PROJ_SERVICE_VERSION, PROJ_ENVIRONMENT

receivers:
  # OTLP receivers - 接收应用程序发送的遥测数据
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317               # Agent gRPC端口（容器内部）
      http:
        endpoint: 0.0.0.0:4318               # Agent HTTP端口（容器内部）

processors:
  # 内存限制器 - 防止Agent内存溢出
  memory_limiter:
    check_interval: 1s
    limit_mib: 256                          # Agent内存限制256MB

  # 批处理器 - 快速转发数据
  batch:
    timeout: 1s                             # 快速批处理
    send_batch_size: 1024                   # 批处理大小
    send_batch_max_size: 2048

  # 资源处理器 - 添加Agent和服务标识
  resource:
    attributes:
      - key: otel.component
        value: "agent"
        action: upsert
      - key: agent.version
        value: "${OTELCOL_VERSION}"
        action: upsert
      - key: service.name
        value: "${PROJ_SERVICE_NAME:-apiserver}"
        action: upsert
      - key: service.version
        value: "${PROJ_SERVICE_VERSION:-v2.0.0}"
        action: upsert
      - key: deployment.environment
        value: "${PROJ_ENVIRONMENT:-development}"
        action: upsert

exporters:
  # 转发到OTEL Collector
  otlp/collector:
    endpoint: "${OTEL_COLLECTOR_ENDPOINT}"  # Collector gRPC端点（动态配置）
    tls:
      insecure: true
    compression: gzip
    timeout: 30s
    retry_on_failure:
      enabled: true
      initial_interval: 1s
      max_interval: 10s
      max_elapsed_time: 5m

  # 调试输出（开发环境）
  debug:
    verbosity: basic
    sampling_initial: 2
    sampling_thereafter: 500

extensions:
  # 健康检查
  health_check:
    endpoint: 0.0.0.0:13133

  # 性能分析
  pprof:
    endpoint: 0.0.0.0:1777

service:
  extensions: [health_check, pprof]

  pipelines:
    # 日志管道
    logs:
      receivers: [otlp]
      processors: [memory_limiter, resource, batch]
      exporters: [otlp/collector]

    # 指标管道
    metrics:
      receivers: [otlp]
      processors: [memory_limiter, resource, batch]
      exporters: [otlp/collector]

    # 追踪管道
    traces:
      receivers: [otlp]
      processors: [memory_limiter, resource, batch]
      exporters: [otlp/collector]

  # Agent遥测配置
  telemetry:
    logs:
      level: info                           # Agent日志级别
      development: false
      encoding: json
    metrics:
      level: basic