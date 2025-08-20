# 日志系统快速开始指南

## 🚀 一分钟快速启动

### 本地开发环境

```bash
# 1. 启动日志收集栈
./scripts/deploy-logging.sh local

# 2. 运行应用程序
make run-api

# 3. 查看日志
curl "http://localhost:9428/select/logsql/query" -d 'query=*'

# 4. 访问 Web UI
open http://localhost:9428/select/vmui/
```

### Docker 环境

```bash
# 1. 启动完整栈
./scripts/deploy-logging.sh docker

# 2. 查看服务状态
docker-compose -f deployments/docker-compose/logging-stack.yml ps

# 3. 访问服务
open http://localhost:9428/select/vmui/  # VictoriaLogs
open http://localhost:3000               # Grafana (admin/admin)
```

### Kubernetes 环境

```bash
# 1. 部署到 K8s
./scripts/deploy-logging.sh k8s

# 2. 端口转发
kubectl port-forward -n logging svc/victorialogs 9428:9428

# 3. 访问 UI
open http://localhost:9428/select/vmui/
```

## 📊 常用操作

### 日志查询

```bash
# 查询所有日志
curl "http://localhost:9428/select/logsql/query" -d 'query=*'

# 查询错误日志
curl "http://localhost:9428/select/logsql/query" -d 'query=level:error'

# 查询特定服务
curl "http://localhost:9428/select/logsql/query" -d 'query=service.name:apiserver'

# 时间范围查询
curl "http://localhost:9428/select/logsql/query" -d 'query=_time:>now-1h'
```

### 测试日志收集

```bash
# 测试日志收集功能
./scripts/deploy-logging.sh test local    # 本地环境
./scripts/deploy-logging.sh test docker   # Docker 环境
./scripts/deploy-logging.sh test k8s      # K8s 环境
```

### 状态检查

```bash
# 检查服务状态
./scripts/deploy-logging.sh status local
./scripts/deploy-logging.sh status docker  
./scripts/deploy-logging.sh status k8s
```

## 🔧 故障排除

### 服务无法启动

```bash
# 检查 Docker 状态
docker ps
docker logs proj-otelcol
docker logs proj-victorialogs

# 检查端口占用
netstat -tlnp | grep -E "(4317|4318|9428|13133)"

# 重启服务
./scripts/deploy-logging.sh clean local
./scripts/deploy-logging.sh local
```

### 日志无法收集

```bash
# 检查 OTEL Collector 健康状态
curl http://localhost:13133

# 检查配置文件
docker exec proj-otelcol cat /etc/otelcol-contrib/config.yaml

# 测试 OTLP 端点
curl -X POST http://localhost:4318/v1/logs \
  -H "Content-Type: application/json" \
  -d '{"resource_logs":[{"logs":[{"body":{"string_value":"test"}}]}]}'
```

### Kubernetes 问题

```bash
# 检查 Pod 状态
kubectl get pods -n logging
kubectl describe pod -l app=otelcol -n logging
kubectl logs -l app=otelcol -n logging

# 检查服务连通性
kubectl exec -n logging deployment/otelcol -- wget -qO- http://victorialogs:9428/health
```

## 📋 访问地址

| 服务 | 本地 | Docker | Kubernetes |
|------|------|--------|------------|
| **VictoriaLogs UI** | http://localhost:9428/select/vmui/ | http://localhost:9428/select/vmui/ | kubectl port-forward |
| **OTLP HTTP** | http://localhost:4318 | http://localhost:4318 | kubectl port-forward |
| **OTLP gRPC** | localhost:4317 | localhost:4317 | kubectl port-forward |
| **健康检查** | http://localhost:13133 | http://localhost:13133 | kubectl port-forward |
| **Grafana** | - | http://localhost:3000 | kubectl port-forward |

## 💡 配置要点

### 应用程序配置

```yaml
log:
  type: "zap"
  level: "info"
  format: "json"
  output-paths: ["stdout", "logs/apiserver/app.log"]
  
  otlp:
    enabled: true
    endpoint: "127.0.0.1:4327"  # 本地/Docker
    # endpoint: "otelcol.logging.svc.cluster.local:4317"  # K8s
    insecure: true
    batch_size: 100
```

### 环境变量

```bash
# 开发环境
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
export OTEL_SERVICE_NAME=apiserver
export OTEL_SERVICE_VERSION=v2.0.0

# Kubernetes 环境
export OTEL_EXPORTER_OTLP_ENDPOINT=http://otelcol.logging.svc.cluster.local:4318
export OTEL_RESOURCE_ATTRIBUTES="k8s.namespace.name=app,k8s.pod.name=$POD_NAME"
```

## 🎯 性能优化

### 高吞吐量配置

```yaml
# OTEL Collector 批处理优化
processors:
  batch:
    timeout: 100ms
    send_batch_size: 5000
    send_batch_max_size: 8000
  
  memory_limiter:
    limit_mib: 2048
```

### 资源限制

```yaml
# Kubernetes 资源配置
resources:
  requests:
    memory: "1Gi"
    cpu: "500m"
  limits:
    memory: "4Gi"
    cpu: "2000m"
```

## 📚 更多文档

- [完整架构设计](logging-architecture.md)
- [详细开发指南](logging-development.md)
- [运维操作手册](logging-operations.md)

---

**需要帮助？** 请查看故障排除部分或联系技术支持团队。