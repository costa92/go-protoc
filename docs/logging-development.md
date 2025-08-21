# 日志系统开发使用指南

## 🚀 快速开始

### 前置要求

- Go 1.21+
- Docker 20.10+
- Docker Compose 3.8+
- kubectl (Kubernetes 环境)

### 本地开发环境搭建

```bash
# 1. 克隆项目
git clone <repository-url>
cd go-protoc

# 2. 安装开发工具
make install-tools A=1

# 3. 启动基础服务
make dev-setup

# 4. 启动日志收集栈
./scripts/installation/service.sh start victorialogs
./scripts/installation/service.sh start otelcol

# 5. 验证服务状态
./scripts/installation/service.sh status all
```

## 📝 应用程序日志配置

### 基础配置

在 `configs/apiserver.yaml` 中配置日志参数：

```yaml
log:
  type: "zap"                     # 使用 Zap 日志框架
  level: "debug"                  # 开发环境使用 debug 级别
  format: "json"                  # JSON 格式便于解析
  disable-caller: false           # 显示调用位置
  disable-stacktrace: false       # 显示堆栈跟踪
  
  # 输出路径配置
  output-paths: 
    - "stdout"                    # 控制台输出
    - "logs/apiserver/app.log"    # 文件输出
  
  # OTLP 配置
  otlp:
    enabled: true                 # 启用 OTLP 发送
    endpoint: "127.0.0.1:4327"    # OTEL Collector 地址
    insecure: true                # 开发环境非安全连接
    timeout: 10s
    batch_size: 100               # 批量发送大小
    
  # 初始字段（所有日志包含）
  initial-fields:
    service: "apiserver"
    version: "v2.0.0"
    environment: "development"
```

### 环境特定配置

```bash
# 开发环境
export LOG_LEVEL=debug
export LOG_FORMAT=json
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318

# 测试环境
export LOG_LEVEL=info
export LOG_FORMAT=json
export OTEL_EXPORTER_OTLP_ENDPOINT=http://otelcol-test:4318

# 生产环境
export LOG_LEVEL=warn
export LOG_FORMAT=json
export OTEL_EXPORTER_OTLP_ENDPOINT=http://otelcol.logging.svc.cluster.local:4318
```

## 💻 代码中使用日志

### 基础用法

```go
package main

import (
    "context"
    "github.com/costa92/go-protoc/v2/pkg/logger"
)

func main() {
    // 初始化日志器
    log := logger.New()
    
    // 基础日志记录
    log.Info("服务启动成功")
    log.Error("数据库连接失败", "error", err)
    log.Debug("处理请求", "user_id", userID, "request_id", reqID)
    
    // 带上下文的日志
    ctx := context.Background()
    log.InfoContext(ctx, "处理用户请求", 
        "user_id", userID,
        "action", "create_order",
        "duration", time.Since(start),
    )
}
```

### 结构化日志最佳实践

```go
// ✅ 推荐：使用结构化字段
log.Info("用户登录成功",
    "user_id", user.ID,
    "username", user.Name,
    "ip_address", req.RemoteAddr,
    "user_agent", req.UserAgent(),
    "duration_ms", time.Since(start).Milliseconds(),
)

// ❌ 不推荐：字符串格式化
log.Infof("用户 %s (ID: %d) 从 %s 登录成功，耗时 %dms", 
    user.Name, user.ID, req.RemoteAddr, time.Since(start).Milliseconds())

// ✅ 推荐：错误处理
if err := db.Create(&user); err != nil {
    log.Error("创建用户失败",
        "error", err,
        "user_data", user,
        "operation", "db.create",
        "table", "users",
    )
    return err
}

// ✅ 推荐：性能监控
start := time.Now()
defer func() {
    log.Debug("API 调用完成",
        "endpoint", "/api/users",
        "method", "POST", 
        "duration_ms", time.Since(start).Milliseconds(),
        "status_code", statusCode,
    )
}()
```

### 上下文相关日志

```go
// 使用上下文传递追踪信息
func ProcessOrder(ctx context.Context, orderID string) error {
    log := logger.FromContext(ctx)
    
    log.Info("开始处理订单",
        "order_id", orderID,
        "trace_id", getTraceID(ctx),
    )
    
    // 业务逻辑...
    
    log.Info("订单处理完成",
        "order_id", orderID,
        "status", "completed",
        "duration", time.Since(start),
    )
    
    return nil
}

// HTTP 中间件示例
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        requestID := generateRequestID()
        
        // 添加请求 ID 到上下文
        ctx := context.WithValue(r.Context(), "request_id", requestID)
        r = r.WithContext(ctx)
        
        // 记录请求开始
        log.Info("HTTP 请求开始",
            "method", r.Method,
            "path", r.URL.Path,
            "remote_addr", r.RemoteAddr,
            "request_id", requestID,
        )
        
        next.ServeHTTP(w, r)
        
        // 记录请求完成
        log.Info("HTTP 请求完成",
            "method", r.Method,
            "path", r.URL.Path,
            "request_id", requestID,
            "duration_ms", time.Since(start).Milliseconds(),
        )
    })
}
```

## 🐳 Docker 环境使用

### 基础部署

```bash
# 使用项目提供的脚本
./scripts/installation/service.sh start all

# 或使用 Docker Compose
docker-compose -f deployments/docker-compose/logging-stack.yml up -d
```

### 开发环境 Docker Compose

创建 `docker-compose.dev.yml`：

```yaml
version: '3.8'

services:
  # 应用程序
  apiserver:
    build: .
    ports:
      - "8080:8080"
    environment:
      - LOG_LEVEL=debug
      - OTEL_EXPORTER_OTLP_ENDPOINT=http://otelcol:4318
      - OTEL_SERVICE_NAME=apiserver
      - OTEL_SERVICE_VERSION=dev
    volumes:
      - ./logs:/app/logs
      - ./configs:/app/configs
    depends_on:
      - otelcol
    networks:
      - logging

  # OpenTelemetry Collector
  otelcol:
    image: otel/opentelemetry-collector-contrib:0.91.0
    ports:
      - "4317:4317"
      - "4318:4318"
      - "13133:13133"
    volumes:
      - ./scripts/installation/otelcol/config-docker-files.yaml:/etc/otelcol-contrib/config.yaml
      - ./logs:/host/logs:ro
    environment:
      - PROJ_SERVICE_NAME=apiserver
      - PROJ_SERVICE_VERSION=dev
      - PROJ_ENVIRONMENT=development
    depends_on:
      - victorialogs
    networks:
      - logging

  # VictoriaLogs
  victorialogs:
    image: victoriametrics/victoria-logs:v1.28.0
    ports:
      - "9428:9428"
    volumes:
      - victorialogs_data:/victorialogs-data
    command:
      - -storageDataPath=/victorialogs-data
      - -httpListenAddr=:9428
    networks:
      - logging

networks:
  logging:
    driver: bridge

volumes:
  victorialogs_data:
```

### 启动和测试

```bash
# 启动开发环境
docker-compose -f docker-compose.dev.yml up -d

# 查看服务状态
docker-compose -f docker-compose.dev.yml ps

# 查看应用日志
docker-compose -f docker-compose.dev.yml logs -f apiserver

# 测试日志收集
curl -X POST http://localhost:8080/api/test

# 查询收集的日志
curl -s "http://localhost:9428/select/logsql/query" \
  -d 'query=service.name:apiserver'
```

## ☸️ Kubernetes 环境使用

### 命名空间和基础资源

```bash
# 创建命名空间
kubectl create namespace logging
kubectl create namespace app

# 部署 VictoriaLogs
kubectl apply -f deployments/kubernetes/victorialogs/

# 部署 OTEL Collector DaemonSet
kubectl apply -f deployments/kubernetes/otelcol/
```

### 应用程序部署配置

```yaml
# k8s-app.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: apiserver
  namespace: app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: apiserver
  template:
    metadata:
      labels:
        app: apiserver
        version: v2.0.0
      annotations:
        # 启用日志收集
        fluentbit.io/parser: json
    spec:
      containers:
      - name: apiserver
        image: your-registry/apiserver:v2.0.0
        ports:
        - containerPort: 8080
        env:
        # OTLP 配置
        - name: OTEL_EXPORTER_OTLP_ENDPOINT
          value: "http://otelcol.logging.svc.cluster.local:4318"
        - name: OTEL_SERVICE_NAME
          value: "apiserver"
        - name: OTEL_SERVICE_VERSION
          value: "v2.0.0"
        - name: OTEL_RESOURCE_ATTRIBUTES
          value: "k8s.namespace.name=$(NAMESPACE),k8s.pod.name=$(POD_NAME),k8s.container.name=$(CONTAINER_NAME)"
        
        # Kubernetes 环境变量
        - name: NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: CONTAINER_NAME
          value: "apiserver"
        
        # 应用配置
        - name: LOG_LEVEL
          value: "info"
        - name: LOG_FORMAT
          value: "json"
        
        volumeMounts:
        - name: config
          mountPath: /app/configs
          readOnly: true
        
        # 健康检查
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "1Gi"
            cpu: "500m"
      
      volumes:
      - name: config
        configMap:
          name: apiserver-config

---
apiVersion: v1
kind: Service
metadata:
  name: apiserver
  namespace: app
spec:
  selector:
    app: apiserver
  ports:
  - port: 8080
    targetPort: 8080
  type: ClusterIP

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: apiserver-config
  namespace: app
data:
  apiserver.yaml: |
    log:
      type: "zap"
      level: "info"
      format: "json"
      output-paths: ["stdout"]
      otlp:
        enabled: true
        endpoint: "otelcol.logging.svc.cluster.local:4317"
        insecure: true
        batch_size: 100
        resource_attributes:
          service.name: "apiserver"
          service.version: "v2.0.0"
          deployment.environment: "production"
```

### 部署和验证

```bash
# 部署应用
kubectl apply -f k8s-app.yaml

# 检查部署状态
kubectl get pods -n app
kubectl get pods -n logging

# 查看应用日志
kubectl logs -n app deployment/apiserver -f

# 端口转发访问 VictoriaLogs UI
kubectl port-forward -n logging svc/victorialogs 9428:9428

# 在另一个终端查询日志
curl -s "http://localhost:9428/select/logsql/query" \
  -d 'query=k8s.namespace.name:app'
```

## 🔍 日志查询和分析

### VictoriaLogs 查询语法

```bash
# 基础查询
curl -s "http://localhost:9428/select/logsql/query" -d 'query=*'

# 服务筛选
curl -s "http://localhost:9428/select/logsql/query" \
  -d 'query=service.name:apiserver'

# 日志级别筛选
curl -s "http://localhost:9428/select/logsql/query" \
  -d 'query=level:error'

# 时间范围查询
curl -s "http://localhost:9428/select/logsql/query" \
  -d 'query=_time:>now-1h'

# 复合条件查询
curl -s "http://localhost:9428/select/logsql/query" \
  -d 'query=service.name:apiserver AND level:error AND _time:>now-1h'

# 正则表达式查询
curl -s "http://localhost:9428/select/logsql/query" \
  -d 'query=msg:~"user.*login"'

# 字段存在性查询
curl -s "http://localhost:9428/select/logsql/query" \
  -d 'query=error:*'

# 统计查询
curl -s "http://localhost:9428/select/logsql/query" \
  -d 'query=* | stats by (level) count() total'
```

### 常用查询示例

```bash
# 1. 查看最近的错误日志
curl -s "http://localhost:9428/select/logsql/query" \
  -d 'query=level:error AND _time:>now-30m | sort by (_time) desc | limit 50'

# 2. 统计各服务的日志量
curl -s "http://localhost:9428/select/logsql/query" \
  -d 'query=* | stats by (service.name) count() logs_count'

# 3. 查找特定用户的操作日志
curl -s "http://localhost:9428/select/logsql/query" \
  -d 'query=user_id:"12345" AND _time:>now-1d'

# 4. 性能分析：查找慢请求
curl -s "http://localhost:9428/select/logsql/query" \
  -d 'query=duration_ms:>1000 AND _time:>now-1h'

# 5. 数据库错误分析
curl -s "http://localhost:9428/select/logsql/query" \
  -d 'query=msg:~"database.*error" OR msg:~"connection.*failed"'
```

### Web UI 使用

访问 `http://localhost:9428/select/vmui/` 使用图形界面：

1. **查询构建器**：可视化构建查询条件
2. **时间范围选择**：快速选择时间范围
3. **日志流**：实时查看日志流
4. **导出功能**：导出查询结果

## 📊 监控和告警

### Grafana 集成

```yaml
# grafana-datasource.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-datasources
data:
  datasources.yaml: |
    apiVersion: 1
    datasources:
    - name: VictoriaLogs
      type: victorialogs-datasource
      url: http://victorialogs:9428
      isDefault: true
      access: proxy
      jsonData:
        maxLines: 1000
```

### 常用仪表板

```json
{
  "dashboard": {
    "title": "应用日志监控",
    "panels": [
      {
        "title": "日志量趋势",
        "type": "graph",
        "targets": [
          {
            "query": "* | stats by (_time:5m) count() logs",
            "refId": "A"
          }
        ]
      },
      {
        "title": "错误日志统计",
        "type": "stat", 
        "targets": [
          {
            "query": "level:error | stats count() errors",
            "refId": "B"
          }
        ]
      },
      {
        "title": "服务日志分布",
        "type": "piechart",
        "targets": [
          {
            "query": "* | stats by (service.name) count() total",
            "refId": "C"
          }
        ]
      }
    ]
  }
}
```

## 🛠️ 开发工具和脚本

### 日志生成器（开发测试用）

```go
// tools/log-generator/main.go
package main

import (
    "math/rand"
    "time"
    "github.com/costa92/go-protoc/v2/pkg/logger"
)

func main() {
    log := logger.New()
    
    levels := []string{"debug", "info", "warn", "error"}
    users := []string{"user1", "user2", "user3", "user4"}
    actions := []string{"login", "logout", "create", "update", "delete"}
    
    for {
        level := levels[rand.Intn(len(levels))]
        user := users[rand.Intn(len(users))]
        action := actions[rand.Intn(len(actions))]
        
        switch level {
        case "debug":
            log.Debug("用户操作", 
                "user_id", user, 
                "action", action,
                "timestamp", time.Now().Unix(),
            )
        case "info":
            log.Info("用户操作成功", 
                "user_id", user, 
                "action", action,
                "duration_ms", rand.Intn(1000),
            )
        case "warn":
            log.Warn("用户操作警告", 
                "user_id", user, 
                "action", action,
                "reason", "rate_limit",
            )
        case "error":
            log.Error("用户操作失败", 
                "user_id", user, 
                "action", action,
                "error", "database_connection_failed",
            )
        }
        
        time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
    }
}
```

### 日志查询脚本

```bash
#!/bin/bash
# scripts/query-logs.sh

VICTORIA_LOGS_URL="http://localhost:9428"
QUERY="$1"
LIMIT="${2:-100}"

if [ -z "$QUERY" ]; then
    echo "Usage: $0 <query> [limit]"
    echo "Examples:"
    echo "  $0 'level:error' 50"
    echo "  $0 'service.name:apiserver AND _time:>now-1h'"
    exit 1
fi

curl -s "${VICTORIA_LOGS_URL}/select/logsql/query" \
    -d "query=${QUERY} | limit ${LIMIT}" \
    | jq -r '
        if type == "array" then
            .[] | [._time, .attributes.level // "unknown", .attributes.msg // .body] | @tsv
        else
            [._time, .attributes.level // "unknown", .attributes.msg // .body] | @tsv
        end
    ' \
    | column -t -s $'\t'
```

### 性能测试脚本

```bash
#!/bin/bash
# scripts/log-performance-test.sh

# 配置参数
CONCURRENT_USERS=10
DURATION=60
LOG_ENDPOINT="http://localhost:4318/v1/logs"

echo "启动日志性能测试..."
echo "并发用户: $CONCURRENT_USERS"
echo "测试时长: ${DURATION}秒"

# 启动多个并发进程
for i in $(seq 1 $CONCURRENT_USERS); do
    {
        end_time=$(($(date +%s) + $DURATION))
        count=0
        
        while [ $(date +%s) -lt $end_time ]; do
            # 发送测试日志
            curl -s -X POST "$LOG_ENDPOINT" \
                -H "Content-Type: application/json" \
                -d '{
                    "resource_logs": [{
                        "resource": {
                            "attributes": [{
                                "key": "service.name",
                                "value": {"string_value": "performance-test"}
                            }]
                        },
                        "scope_logs": [{
                            "log_records": [{
                                "time_unix_nano": "'$(date +%s%N)'",
                                "severity_text": "INFO",
                                "body": {
                                    "string_value": "Performance test log #'$count' from user '$i'"
                                },
                                "attributes": [{
                                    "key": "user_id",
                                    "value": {"string_value": "user_'$i'"}
                                }, {
                                    "key": "test_iteration", 
                                    "value": {"int_value": '$count'}
                                }]
                            }]
                        }]
                    }]
                }' > /dev/null
            
            count=$((count + 1))
            sleep 0.1
        done
        
        echo "用户 $i 发送了 $count 条日志"
    } &
done

# 等待所有进程完成
wait

echo "性能测试完成！"

# 查询测试结果
sleep 5
echo "查询测试日志统计..."
curl -s "http://localhost:9428/select/logsql/query" \
    -d 'query=service.name:performance-test | stats count() total' \
    | jq -r '.total // 0' \
    | xargs -I {} echo "总计收集日志: {} 条"
```

## 🔧 故障排除

### VictoriaLogs _msg 字段映射问题

**问题描述**: 在VictoriaLogs中查询日志时出现 `"_msg":"missing _msg field"` 错误，无法正确显示日志消息内容。

**根本原因**: VictoriaLogs要求日志消息必须存储在 `_msg` 字段中，但OpenTelemetry Collector的默认配置将消息映射到了 `body.msg` 字段。

**解决方案**:

1. **修改OpenTelemetry Collector配置文件**:
   ```bash
   # 找到实际使用的配置文件
   docker inspect proj-otelcol | grep config
   
   # 修改配置文件中的字段映射操作符
   vim /path/to/otelcol/config.yaml
   ```

2. **更新字段映射配置**:
   ```yaml
   receivers:
     filelog:
       operators:
         - type: json_parser
           id: parse_json
         # ✅ 正确：映射到 attributes._msg 
         - type: move
           from: attributes.msg
           to: attributes._msg
         # ❌ 错误：映射到 body.msg (VictoriaLogs无法识别)
         # - type: move
         #   from: attributes.msg  
         #   to: body.msg
   ```

3. **重启OpenTelemetry Collector**:
   ```bash
   docker restart proj-otelcol
   ```

4. **验证修复结果**:
   ```bash
   # 生成测试日志
   echo '{"level":"info","ts":"'$(date -Iseconds)'","msg":"VictoriaLogs _msg field test"}' >> logs/apiserver/app.log
   
   # 等待几秒后查询验证
   sleep 3
   curl -s "http://localhost:9428/select/logsql/query" -d 'query=_msg:"VictoriaLogs _msg field test"'
   ```

5. **预期结果对比**:
   ```bash
   # ❌ 修复前: 
   {"_msg":"missing _msg field; see https://docs.victoriametrics.com/victorialogs/keyconcepts/#message-field","attributes.msg":"actual message content"}
   
   # ✅ 修复后:
   {"_msg":"actual message content","level":"info","ts":"2025-08-21T23:15:00+08:00"}
   ```

**配置文件位置参考**:
- Docker部署: `/path/to/_thirdparty/otelcol/config/config.yaml`
- 模板文件: `scripts/installation/otelcol/config-docker.yaml`

**注意事项**:
- 确保修改的是容器实际挂载的配置文件
- 修改后必须重启容器才能生效
- 建议同时更新模板文件以保持一致性

### 常见问题诊断

```bash
# 1. 检查服务状态
./scripts/installation/service.sh status all

# 2. 检查容器日志
docker logs proj-otelcol --tail 50
docker logs proj-victorialogs --tail 50

# 3. 检查网络连通性
docker exec proj-otelcol wget -qO- http://proj-victorialogs:9428/health

# 4. 检查配置文件
docker exec proj-otelcol cat /etc/otelcol-contrib/config.yaml

# 5. 测试 OTLP 端点
curl -X POST http://localhost:4318/v1/logs \
    -H "Content-Type: application/json" \
    -d '{"resource_logs":[{"logs":[{"body":{"string_value":"test"}}]}]}'
```

### 性能调优

```yaml
# OTEL Collector 高性能配置
processors:
  batch:
    timeout: 100ms              # 降低延迟
    send_batch_size: 5000       # 增加批量大小
    send_batch_max_size: 8000   # 最大批量大小
  
  memory_limiter:
    limit_mib: 4096             # 增加内存限制
    spike_limit_mib: 1024       # 峰值内存

# 资源限制
resources:
  limits:
    memory: 8Gi                 # 增加内存
    cpu: 4000m                  # 增加 CPU
  requests:
    memory: 2Gi
    cpu: 1000m
```

### 调试技巧

```bash
# 1. 启用详细日志
export OTEL_LOG_LEVEL=debug

# 2. 使用本地文件验证
tail -f logs/apiserver/app.log

# 3. 直接查询 VictoriaLogs API
curl -v "http://localhost:9428/select/logsql/query" -d 'query=*'

# 4. 监控 OTEL Collector 指标
curl http://localhost:8888/metrics | grep otelcol_receiver

# 5. 使用 tcpdump 抓包分析
sudo tcpdump -i lo port 4318 -A
```

## 🚀 最佳实践

### 日志设计原则

1. **结构化优先**：始终使用 JSON 格式
2. **上下文丰富**：包含足够的上下文信息
3. **性能考虑**：避免在热路径中记录过多日志
4. **安全意识**：不记录敏感信息
5. **标准化字段**：使用统一的字段命名

### 开发流程建议

1. **本地开发**：使用完整的日志栈进行开发和测试
2. **代码审查**：审查日志记录是否合理
3. **性能测试**：验证日志对性能的影响
4. **集成测试**：测试日志在各环境中的表现
5. **监控告警**：为关键错误设置告警

---

这份开发指南涵盖了从本地开发到生产部署的完整流程。如有任何问题，请参考故障排除部分或联系技术支持团队。