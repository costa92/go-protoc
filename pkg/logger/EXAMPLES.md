# Logger Package 使用示例

本文档提供 Logger 包在各种实际场景中的详细使用示例，包括配置、代码实现和最佳实践。

## 目录

- [基础使用示例](#基础使用示例)
- [配置示例](#配置示例)
- [部署场景示例](#部署场景示例)
- [高级功能示例](#高级功能示例)
- [集成示例](#集成示例)
- [故障排除示例](#故障排除示例)
- [性能优化示例](#性能优化示例)

## 基础使用示例

### 1. 最简单的使用方式

```go
package main

import (
    "github.com/costa92/go-protoc/v2/pkg/logger"
)

func main() {
    // 使用全局默认日志器
    logger.Info("应用启动")
    logger.Infow("用户登录", "user_id", 12345, "ip", "192.168.1.1")
}
```

**输出**：
```json
{"level":"info","ts":"2025-08-28T10:30:15.123+0800","caller":"main.go:8","msg":"应用启动","type":"zap"}
{"level":"info","ts":"2025-08-28T10:30:15.124+0800","caller":"main.go:9","msg":"用户登录","type":"zap","user_id":12345,"ip":"192.168.1.1"}
```

### 2. 创建自定义日志器

```go
package main

import (
    "github.com/costa92/go-protoc/v2/pkg/logger"
)

func main() {
    // 简化配置方式
    config := &logger.QuickConfig{
        Preset: logger.PresetDevelopment,
        Type:   "zap",
        Level:  "debug",
    }
    
    log, err := logger.NewLoggerFromQuickConfig(config)
    if err != nil {
        panic(err)
    }
    
    // 使用自定义日志器
    log.Debug("调试信息")
    log.Info("应用启动")
    log.Warn("警告消息")
}
```

### 3. 带上下文的日志器

```go
package main

import (
    "context"
    "github.com/costa92/go-protoc/v2/pkg/logger"
)

func main() {
    // 创建带有公共字段的日志器
    serviceLogger := logger.With("service", "user-service", "version", "v1.0.0")
    
    // 所有日志都会包含service和version字段
    serviceLogger.Info("服务初始化")
    
    // 创建带请求上下文的日志器
    ctx := context.Background()
    requestLogger := serviceLogger.WithCtx(ctx, "request_id", "req-abc123", "user_id", 456)
    
    requestLogger.Info("处理用户请求")
    requestLogger.Warn("请求处理缓慢", "duration_ms", 1500)
}
```

## 配置示例

### 1. OTLP 日志收集配置

#### 基础OTLP配置（推荐）

```yaml
# config/app.yaml
log:
  preset: "observability"          # 可观测性预设
  type: "zap"                     # 日志器类型
  level: "info"                   # 日志级别
  
  # OTLP配置 - 仅需端点即可自动启用
  otlp-endpoint: "127.0.0.1:4327" # OTEL Agent端点
```

```go
package main

import (
    "github.com/spf13/viper"
    "github.com/costa92/go-protoc/v2/pkg/logger"
)

func main() {
    // 加载配置文件
    viper.SetConfigFile("config/app.yaml")
    viper.ReadInConfig()
    
    // 解析简化配置
    config := &logger.QuickConfig{
        Preset:       logger.LogPreset(viper.GetString("log.preset")),
        Type:         viper.GetString("log.type"),
        Level:        viper.GetString("log.level"),
        OTLPEndpoint: viper.GetString("log.otlp-endpoint"),
    }
    
    log, err := logger.NewLoggerFromQuickConfig(config)
    if err != nil {
        panic(err)
    }
    
    log.Info("OTLP日志器启动成功")
    // 日志将通过OTLP发送到OTEL Agent，然后转发到Collector
}
```

#### 详细OTLP配置

```yaml
log:
  type: "zap"
  level: "info"
  format: "json"
  output-paths: ["stdout"]
  
  # 详细OTLP配置
  otlp:
    enabled: true
    endpoint: "otelcol.logging.svc.cluster.local:4317"
    protocol: "grpc"
    timeout: "10s"
    batch_timeout: "5s"
    batch_size: 200
    insecure: true
    headers:
      authorization: "Bearer your-token"
      x-tenant-id: "tenant-123"
    service_name: "user-service"
    service_version: "v2.1.0"
    environment: "production"
    resource_attributes:
      deployment.environment: "k8s"
      service.namespace: "backend"
```

```go
func createDetailedOTLPLogger() logger.Logger {
    opts := &logger.LogsOptions{
        Type:         logger.LoggerTypeZap,
        Level:        "info",
        Format:       "json",
        OutputPaths:  []string{"stdout"},
        
        OTLP: &logger.OTLPConfig{
            Enabled:      true,
            Endpoint:     "otelcol.logging.svc.cluster.local:4317",
            Protocol:     "grpc",
            Timeout:      10 * time.Second,
            BatchTimeout: 5 * time.Second,
            BatchSize:    200,
            Insecure:     true,
            Headers: map[string]string{
                "authorization": "Bearer your-token",
                "x-tenant-id":   "tenant-123",
            },
            ServiceName:    "user-service",
            ServiceVersion: "v2.1.0",
            Environment:    "production",
        },
    }
    
    log, err := logger.NewLogger(opts)
    if err != nil {
        panic(err)
    }
    return log
}
```

### 2. 多环境配置

#### 开发环境配置

```yaml
# config/development.yaml
log:
  preset: "development"           # 开发环境预设
  type: "zap"                    # 日志器类型
  level: "debug"                 # 调试级别
  log-dir: "logs"                # 本地文件输出
```

```go
func developmentLogger() logger.Logger {
    config := &logger.QuickConfig{
        Preset: logger.PresetDevelopment,
        Type:   "zap",
        Level:  "debug",
        LogDir: "logs",  // 本地文件便于调试
    }
    
    log, _ := logger.NewLoggerFromQuickConfig(config)
    return log
}

func exampleDevelopmentUsage() {
    log := developmentLogger()
    
    // 开发环境的详细日志
    log.Debug("数据库连接初始化", "host", "localhost", "port", 3306)
    log.Info("HTTP服务器启动", "port", 8080)
    log.Warn("配置文件使用默认值", "config", "database.timeout")
    
    // 调试用的结构化信息
    log.Debugw("用户认证流程",
        "step", "validate_token",
        "token_type", "JWT",
        "expiry", time.Now().Add(24*time.Hour),
        "claims", map[string]interface{}{
            "user_id": 123,
            "role":    "admin",
        },
    )
}
```

#### 生产环境配置

```yaml
# config/production.yaml
log:
  preset: "production"            # 生产环境预设
  type: "zap"                    # 高性能Zap
  level: "warn"                  # 仅警告和错误
  log-dir: "/var/log/app"        # 标准日志目录
```

```go
func productionLogger() logger.Logger {
    config := &logger.QuickConfig{
        Preset: logger.PresetProduction,
        Type:   "zap",
        Level:  "warn",
        LogDir: "/var/log/app",
    }
    
    log, _ := logger.NewLoggerFromQuickConfig(config)
    return log
}

func exampleProductionUsage() {
    log := productionLogger()
    
    // 生产环境关注的日志
    log.Warn("数据库连接池接近上限", "current", 95, "max", 100)
    log.Error("支付处理失败", "order_id", "ORD-2025-001", "amount", 99.99)
    
    // 关键业务事件
    log.Warnw("用户异常行为检测",
        "user_id", 12345,
        "behavior", "rapid_requests",
        "count", 100,
        "window", "1m",
        "action", "rate_limit_applied",
    )
}
```

#### 测试环境配置

```yaml
# config/testing.yaml
log:
  preset: "testing"              # 测试环境预设
  type: "zap"                   # 快速启动
```

```go
func testLogger() logger.Logger {
    config := &logger.QuickConfig{
        Preset: logger.PresetTesting,
        Type:   "zap",
    }
    
    log, _ := logger.NewLoggerFromQuickConfig(config)
    return log
}

func TestUserService(t *testing.T) {
    log := testLogger()
    
    // 测试环境只记录错误，避免测试输出噪音
    service := NewUserService(log)
    
    user, err := service.CreateUser("john@example.com")
    if err != nil {
        log.Error("用户创建失败", "error", err, "email", "john@example.com")
        t.Fatal(err)
    }
    
    assert.NotNil(t, user)
}
```

### 3. 混合架构配置

同时输出到本地文件和OTLP：

```yaml
# config/hybrid.yaml
log:
  type: "zap"
  level: "info"
  format: "json"
  
  # 明确指定输出路径（包含文件）
  output-paths: ["stdout", "/var/log/app/app.log"]
  error-output-paths: ["stderr"]
  
  # 同时启用OTLP
  otlp:
    enabled: true
    endpoint: "127.0.0.1:4327"
    batch_size: 100
```

```go
func hybridLogger() logger.Logger {
    opts := &logger.LogsOptions{
        Type:             logger.LoggerTypeZap,
        Level:            "info",
        Format:           "json",
        OutputPaths:      []string{"stdout", "/var/log/app/app.log"},
        ErrorOutputPaths: []string{"stderr"},
        
        OTLP: &logger.OTLPConfig{
            Enabled:   true,
            Endpoint:  "127.0.0.1:4327",
            BatchSize: 100,
        },
    }
    
    log, _ := logger.NewLogger(opts)
    return log
}
```

## 部署场景示例

### 1. Docker 容器部署

#### Dockerfile

```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o server cmd/server/main.go

FROM alpine:3.18
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# 创建日志目录
RUN mkdir -p /var/log/app

COPY --from=builder /app/server .
COPY --from=builder /app/config/ ./config/

CMD ["./server", "-c", "config/docker.yaml"]
```

#### Docker配置

```yaml
# config/docker.yaml
log:
  preset: "observability"
  type: "zap"
  level: "info"
  
  # Docker环境连接OTEL Collector
  otlp-endpoint: "otelcol:4317"
```

#### docker-compose.yml

```yaml
version: '3.8'

services:
  app:
    build: .
    environment:
      - LOG_LEVEL=info
    depends_on:
      - otelcol
    volumes:
      - ./logs:/var/log/app
    networks:
      - logging

  otelcol:
    image: otel/opentelemetry-collector-contrib:0.91.0
    command: ["--config=/etc/otel-collector-config.yaml"]
    volumes:
      - ./otel-collector-config.yaml:/etc/otel-collector-config.yaml
    ports:
      - "4317:4317"   # OTLP gRPC receiver
      - "4318:4318"   # OTLP HTTP receiver
    networks:
      - logging

networks:
  logging:
    driver: bridge
```

### 2. Kubernetes 部署

#### ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: default
data:
  app.yaml: |
    log:
      preset: "observability"
      type: "zap"
      level: "info"
      
      # Kubernetes环境中的OTEL Collector
      otlp-endpoint: "otelcol.logging.svc.cluster.local:4317"
```

#### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: user-service
  template:
    metadata:
      labels:
        app: user-service
    spec:
      containers:
      - name: user-service
        image: myregistry/user-service:v1.0.0
        args: ["-c", "/etc/config/app.yaml"]
        volumeMounts:
        - name: config
          mountPath: /etc/config
        env:
        - name: SERVICE_NAME
          value: "user-service"
        - name: LOG_LEVEL
          value: "info"
      volumes:
      - name: config
        configMap:
          name: app-config
```

#### Service与Ingress

```yaml
apiVersion: v1
kind: Service
metadata:
  name: user-service
  namespace: default
spec:
  selector:
    app: user-service
  ports:
  - port: 80
    targetPort: 8080

---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: user-service-ingress
  namespace: default
spec:
  rules:
  - host: api.example.com
    http:
      paths:
      - path: /users
        pathType: Prefix
        backend:
          service:
            name: user-service
            port:
              number: 80
```

### 3. 云原生部署

#### AWS ECS 配置

```json
{
  "family": "user-service",
  "containerDefinitions": [
    {
      "name": "user-service",
      "image": "myregistry/user-service:v1.0.0",
      "memory": 512,
      "cpu": 256,
      "environment": [
        {
          "name": "LOG_LEVEL",
          "value": "info"
        },
        {
          "name": "OTLP_ENDPOINT", 
          "value": "otel-collector.internal:4317"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/aws/ecs/user-service",
          "awslogs-region": "us-west-2",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ]
}
```

## 高级功能示例

### 1. 动态日志级别调整

```go
package main

import (
    "net/http"
    "github.com/costa92/go-protoc/v2/pkg/logger"
)

var globalLogger logger.Logger

func main() {
    // 创建支持动态配置的日志器
    opts := &logger.LogsOptions{
        Type:    logger.LoggerTypeZap,
        Dynamic: true,  // 启用动态配置
        Level:   "info",
    }
    
    globalLogger, _ = logger.NewLogger(opts)
    
    // 启动HTTP服务器用于动态调整
    http.HandleFunc("/admin/log-level", handleLogLevel)
    http.ListenAndServe(":8080", nil)
}

func handleLogLevel(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPUT {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    level := r.URL.Query().Get("level")
    if level == "" {
        http.Error(w, "Missing level parameter", http.StatusBadRequest)
        return
    }
    
    // 动态调整日志级别
    if dynamicLogger, ok := globalLogger.(*logger.DynamicLogger); ok {
        err := dynamicLogger.UpdateLevel(level)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        
        globalLogger.Infow("日志级别已更新", "new_level", level)
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("日志级别已更新为: " + level))
    } else {
        http.Error(w, "动态配置未启用", http.StatusBadRequest)
    }
}

// 使用示例：
// curl -X PUT "http://localhost:8080/admin/log-level?level=debug"
```

### 2. 结构化日志与追踪集成

```go
package main

import (
    "context"
    "time"
    
    "github.com/costa92/go-protoc/v2/pkg/logger"
    "go.opentelemetry.io/otel/trace"
)

type UserService struct {
    log logger.Logger
}

func NewUserService(log logger.Logger) *UserService {
    return &UserService{
        log: log.With("component", "user-service"),
    }
}

func (s *UserService) CreateUser(ctx context.Context, email string) (*User, error) {
    // 从context中提取追踪信息
    span := trace.SpanFromContext(ctx)
    traceID := span.SpanContext().TraceID().String()
    spanID := span.SpanContext().SpanID().String()
    
    // 创建带追踪信息的日志器
    reqLog := s.log.WithCtx(ctx, 
        "trace_id", traceID,
        "span_id", spanID,
        "operation", "create_user",
    )
    
    reqLog.Infow("开始创建用户", "email", email)
    
    start := time.Now()
    
    // 模拟用户创建过程
    if email == "" {
        reqLog.Error("用户创建失败：邮箱为空")
        return nil, errors.New("email is required")
    }
    
    // 验证邮箱格式
    if !isValidEmail(email) {
        reqLog.Errorw("用户创建失败：邮箱格式无效", "email", email)
        return nil, errors.New("invalid email format")
    }
    
    // 检查用户是否存在
    exists, err := s.checkUserExists(ctx, email)
    if err != nil {
        reqLog.Errorw("用户存在性检查失败", "error", err)
        return nil, err
    }
    
    if exists {
        reqLog.Warnw("用户创建跳过：用户已存在", "email", email)
        return nil, errors.New("user already exists")
    }
    
    // 创建用户
    user := &User{
        ID:        generateUserID(),
        Email:     email,
        CreatedAt: time.Now(),
    }
    
    err = s.saveUser(ctx, user)
    if err != nil {
        reqLog.Errorw("用户保存失败", "error", err, "user_id", user.ID)
        return nil, err
    }
    
    duration := time.Since(start)
    
    // 记录成功创建的结构化日志
    reqLog.Infow("用户创建成功",
        "user_id", user.ID,
        "email", user.Email,
        "duration_ms", duration.Milliseconds(),
        "created_at", user.CreatedAt,
    )
    
    // 记录业务指标
    reqLog.Infow("业务指标记录",
        "metric_type", "user_creation",
        "metric_value", 1,
        "metric_labels", map[string]string{
            "status": "success",
            "method": "api",
        },
    )
    
    return user, nil
}

func (s *UserService) checkUserExists(ctx context.Context, email string) (bool, error) {
    reqLog := s.log.WithCtx(ctx, "operation", "check_user_exists")
    
    reqLog.Debugw("检查用户是否存在", "email", email)
    
    // 模拟数据库查询
    time.Sleep(10 * time.Millisecond)
    
    return false, nil
}

func (s *UserService) saveUser(ctx context.Context, user *User) error {
    reqLog := s.log.WithCtx(ctx, "operation", "save_user")
    
    reqLog.Debugw("保存用户到数据库", "user_id", user.ID)
    
    // 模拟数据库保存
    time.Sleep(50 * time.Millisecond)
    
    reqLog.Debugw("用户保存完成", "user_id", user.ID)
    return nil
}

type User struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

func generateUserID() string {
    return "usr_" + time.Now().Format("20060102150405")
}

func isValidEmail(email string) bool {
    return strings.Contains(email, "@")
}
```

### 3. 错误处理与恢复日志

```go
package main

import (
    "context"
    "errors"
    "runtime/debug"
    "time"
    
    "github.com/costa92/go-protoc/v2/pkg/logger"
)

type PaymentService struct {
    log logger.Logger
}

func NewPaymentService(log logger.Logger) *PaymentService {
    return &PaymentService{
        log: log.With("service", "payment-service"),
    }
}

func (s *PaymentService) ProcessPayment(ctx context.Context, orderID string, amount float64) error {
    reqLog := s.log.WithCtx(ctx, 
        "operation", "process_payment",
        "order_id", orderID,
        "amount", amount,
    )
    
    // 设置恢复处理
    defer func() {
        if r := recover(); r != nil {
            stack := string(debug.Stack())
            reqLog.Errorw("支付处理发生严重错误",
                "panic", r,
                "stack_trace", stack,
                "recovery_time", time.Now(),
            )
        }
    }()
    
    reqLog.Infow("开始处理支付")
    
    // 输入验证
    if orderID == "" {
        reqLog.Error("支付处理失败：订单ID为空")
        return errors.New("order ID is required")
    }
    
    if amount <= 0 {
        reqLog.Errorw("支付处理失败：金额无效", "amount", amount)
        return errors.New("invalid amount")
    }
    
    // 重试机制的支付处理
    maxRetries := 3
    for attempt := 1; attempt <= maxRetries; attempt++ {
        attemptLog := reqLog.With("attempt", attempt, "max_retries", maxRetries)
        
        err := s.attemptPayment(ctx, orderID, amount, attemptLog)
        if err == nil {
            reqLog.Infow("支付处理成功", "attempts_used", attempt)
            return nil
        }
        
        if attempt < maxRetries {
            backoff := time.Duration(attempt) * time.Second
            attemptLog.Warnw("支付尝试失败，将重试",
                "error", err,
                "next_attempt_in", backoff,
                "remaining_attempts", maxRetries-attempt,
            )
            time.Sleep(backoff)
        } else {
            attemptLog.Errorw("支付处理最终失败",
                "error", err,
                "total_attempts", maxRetries,
                "final_failure", true,
            )
        }
    }
    
    return errors.New("payment processing failed after all retries")
}

func (s *PaymentService) attemptPayment(ctx context.Context, orderID string, amount float64, log logger.Logger) error {
    log.Debugw("执行支付尝试")
    
    // 模拟支付网关调用
    start := time.Now()
    
    // 模拟随机失败
    if time.Now().UnixNano()%3 == 0 {
        duration := time.Since(start)
        log.Warnw("支付网关响应错误", 
            "error", "gateway_timeout",
            "duration_ms", duration.Milliseconds(),
        )
        return errors.New("payment gateway timeout")
    }
    
    duration := time.Since(start)
    log.Debugw("支付网关响应成功", "duration_ms", duration.Milliseconds())
    
    // 记录支付成功的业务事件
    log.Infow("支付事务完成",
        "transaction_id", "txn_"+orderID+"_"+time.Now().Format("150405"),
        "status", "completed",
        "gateway_response_time_ms", duration.Milliseconds(),
    )
    
    return nil
}

// 使用示例
func main() {
    config := &logger.QuickConfig{
        Preset:       logger.PresetObservability,
        Type:         "zap",
        Level:        "debug",
        OTLPEndpoint: "127.0.0.1:4327",
    }
    
    log, err := logger.NewLoggerFromQuickConfig(config)
    if err != nil {
        panic(err)
    }
    
    service := NewPaymentService(log)
    
    ctx := context.Background()
    err = service.ProcessPayment(ctx, "ORD-2025-001", 99.99)
    if err != nil {
        log.Errorw("支付处理失败", "error", err)
    }
}
```

## 集成示例

### 1. HTTP中间件集成

```go
package middleware

import (
    "net/http"
    "time"
    "strconv"
    
    "github.com/costa92/go-protoc/v2/pkg/logger"
    "github.com/gorilla/mux"
)

func LoggingMiddleware(log logger.Logger) mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            // 生成请求ID
            requestID := generateRequestID()
            
            // 创建带请求信息的日志器
            reqLog := log.WithCtx(r.Context(),
                "request_id", requestID,
                "method", r.Method,
                "path", r.URL.Path,
                "remote_addr", r.RemoteAddr,
                "user_agent", r.UserAgent(),
            )
            
            // 将日志器添加到请求上下文
            ctx := context.WithValue(r.Context(), "logger", reqLog)
            r = r.WithContext(ctx)
            
            // 包装ResponseWriter以捕获状态码
            wrapped := &responseWrapper{ResponseWriter: w}
            
            reqLog.Infow("HTTP请求开始")
            
            // 执行下一个处理器
            next.ServeHTTP(wrapped, r)
            
            duration := time.Since(start)
            
            // 记录请求完成日志
            reqLog.Infow("HTTP请求完成",
                "status_code", wrapped.statusCode,
                "duration_ms", duration.Milliseconds(),
                "response_size", wrapped.bytesWritten,
            )
            
            // 记录慢请求警告
            if duration > 1*time.Second {
                reqLog.Warnw("慢请求检测",
                    "duration_ms", duration.Milliseconds(),
                    "threshold_ms", 1000,
                )
            }
        })
    }
}

type responseWrapper struct {
    http.ResponseWriter
    statusCode   int
    bytesWritten int64
}

func (w *responseWrapper) WriteHeader(statusCode int) {
    w.statusCode = statusCode
    w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWrapper) Write(b []byte) (int, error) {
    if w.statusCode == 0 {
        w.statusCode = http.StatusOK
    }
    n, err := w.ResponseWriter.Write(b)
    w.bytesWritten += int64(n)
    return n, err
}

func generateRequestID() string {
    return "req_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

// 从上下文获取日志器的辅助函数
func GetLoggerFromContext(ctx context.Context) logger.Logger {
    if log, ok := ctx.Value("logger").(logger.Logger); ok {
        return log
    }
    return logger.GetDefaultLogger()
}
```

### 2. gRPC 拦截器集成

```go
package interceptor

import (
    "context"
    "time"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    
    "github.com/costa92/go-protoc/v2/pkg/logger"
)

func UnaryLoggingInterceptor(log logger.Logger) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        start := time.Now()
        
        // 创建带gRPC信息的日志器
        reqLog := log.WithCtx(ctx,
            "grpc_service", info.FullMethod,
            "grpc_type", "unary",
        )
        
        reqLog.Infow("gRPC请求开始", "request", req)
        
        // 执行处理器
        resp, err := handler(ctx, req)
        
        duration := time.Since(start)
        
        if err != nil {
            // 记录错误
            st := status.Convert(err)
            reqLog.Errorw("gRPC请求失败",
                "error", err,
                "grpc_code", st.Code(),
                "grpc_message", st.Message(),
                "duration_ms", duration.Milliseconds(),
            )
        } else {
            // 记录成功
            reqLog.Infow("gRPC请求完成",
                "duration_ms", duration.Milliseconds(),
                "response", resp,
            )
        }
        
        return resp, err
    }
}

func StreamLoggingInterceptor(log logger.Logger) grpc.StreamServerInterceptor {
    return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
        start := time.Now()
        
        reqLog := log.WithCtx(stream.Context(),
            "grpc_service", info.FullMethod,
            "grpc_type", "stream",
            "grpc_client_stream", info.IsClientStream,
            "grpc_server_stream", info.IsServerStream,
        )
        
        reqLog.Infow("gRPC流开始")
        
        err := handler(srv, stream)
        
        duration := time.Since(start)
        
        if err != nil {
            st := status.Convert(err)
            reqLog.Errorw("gRPC流失败",
                "error", err,
                "grpc_code", st.Code(),
                "duration_ms", duration.Milliseconds(),
            )
        } else {
            reqLog.Infow("gRPC流完成", "duration_ms", duration.Milliseconds())
        }
        
        return err
    }
}
```

### 3. 数据库查询日志集成

```go
package repository

import (
    "context"
    "time"
    
    "gorm.io/gorm"
    gormlogger "gorm.io/gorm/logger"
    
    "github.com/costa92/go-protoc/v2/pkg/logger"
)

type GormLoggerAdapter struct {
    log                   logger.Logger
    slowThreshold         time.Duration
    ignoreRecordNotFound  bool
}

func NewGormLogger(log logger.Logger, slowThreshold time.Duration) gormlogger.Interface {
    return &GormLoggerAdapter{
        log:                  log,
        slowThreshold:        slowThreshold,
        ignoreRecordNotFound: true,
    }
}

func (l *GormLoggerAdapter) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
    return l
}

func (l *GormLoggerAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
    l.log.WithCtx(ctx).Infof(msg, data...)
}

func (l *GormLoggerAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
    l.log.WithCtx(ctx).Warnf(msg, data...)
}

func (l *GormLoggerAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
    l.log.WithCtx(ctx).Errorf(msg, data...)
}

func (l *GormLoggerAdapter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
    elapsed := time.Since(begin)
    sql, rows := fc()
    
    reqLog := l.log.WithCtx(ctx,
        "duration_ms", elapsed.Milliseconds(),
        "rows_affected", rows,
        "sql", sql,
    )
    
    switch {
    case err != nil && (!l.ignoreRecordNotFound || !errors.Is(err, gorm.ErrRecordNotFound)):
        reqLog.Errorw("数据库查询错误", "error", err)
    case l.slowThreshold != 0 && elapsed > l.slowThreshold:
        reqLog.Warnw("慢查询检测", 
            "threshold_ms", l.slowThreshold.Milliseconds(),
        )
    default:
        reqLog.Debugw("数据库查询完成")
    }
}

// 使用示例
func setupDatabase(log logger.Logger) *gorm.DB {
    db, err := gorm.Open(mysql.Open("user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{
        Logger: NewGormLogger(log.With("component", "database"), 200*time.Millisecond),
    })
    
    if err != nil {
        log.Fatal("数据库连接失败", "error", err)
    }
    
    return db
}

type UserRepository struct {
    db  *gorm.DB
    log logger.Logger
}

func NewUserRepository(db *gorm.DB, log logger.Logger) *UserRepository {
    return &UserRepository{
        db:  db,
        log: log.With("component", "user-repository"),
    }
}

func (r *UserRepository) CreateUser(ctx context.Context, user *User) error {
    reqLog := r.log.WithCtx(ctx, "operation", "create_user")
    
    reqLog.Infow("创建用户", "email", user.Email)
    
    result := r.db.WithContext(ctx).Create(user)
    if result.Error != nil {
        reqLog.Errorw("用户创建失败", "error", result.Error)
        return result.Error
    }
    
    reqLog.Infow("用户创建成功", "user_id", user.ID)
    return nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uint) (*User, error) {
    reqLog := r.log.WithCtx(ctx, "operation", "get_user_by_id", "user_id", id)
    
    var user User
    result := r.db.WithContext(ctx).First(&user, id)
    
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        reqLog.Warnw("用户不存在")
        return nil, result.Error
    } else if result.Error != nil {
        reqLog.Errorw("查询用户失败", "error", result.Error)
        return nil, result.Error
    }
    
    reqLog.Debugw("查询用户成功")
    return &user, nil
}
```

## 故障排除示例

### 1. 配置验证与错误处理

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/costa92/go-protoc/v2/pkg/logger"
)

func validateAndCreateLogger() logger.Logger {
    config := &logger.QuickConfig{
        Preset:       logger.PresetObservability,
        Type:         "zap",
        Level:        "info",
        OTLPEndpoint: os.Getenv("OTLP_ENDPOINT"),
    }
    
    // 验证配置
    if err := validateConfig(config); err != nil {
        fmt.Printf("❌ 配置验证失败: %v\n", err)
        fmt.Println("🔧 使用默认配置...")
        config = logger.DefaultQuickConfig()
    }
    
    // 尝试创建日志器
    log, err := logger.NewLoggerFromQuickConfig(config)
    if err != nil {
        fmt.Printf("❌ 日志器创建失败: %v\n", err)
        fmt.Println("🔧 降级到标准日志器...")
        
        // 降级到最简单的配置
        fallbackConfig := &logger.QuickConfig{
            Preset: logger.PresetDevelopment,
            Type:   "zap",
        }
        
        log, err = logger.NewLoggerFromQuickConfig(fallbackConfig)
        if err != nil {
            panic("无法创建任何日志器: " + err.Error())
        }
        
        log.Warn("使用降级日志器配置")
    }
    
    // 测试日志器功能
    if err := testLogger(log); err != nil {
        fmt.Printf("⚠️ 日志器测试失败: %v\n", err)
    } else {
        fmt.Println("✅ 日志器测试通过")
    }
    
    return log
}

func validateConfig(config *logger.QuickConfig) error {
    if config == nil {
        return fmt.Errorf("config is nil")
    }
    
    // 验证预设
    validPresets := map[logger.LogPreset]bool{
        logger.PresetDevelopment:   true,
        logger.PresetProduction:    true,
        logger.PresetTesting:       true,
        logger.PresetObservability: true,
    }
    
    if !validPresets[config.Preset] {
        return fmt.Errorf("invalid preset: %s", config.Preset)
    }
    
    // 验证类型
    if config.Type != "" && config.Type != "zap" && config.Type != "slog" {
        return fmt.Errorf("invalid logger type: %s", config.Type)
    }
    
    // 验证级别
    validLevels := map[string]bool{
        "debug": true,
        "info":  true,
        "warn":  true,
        "error": true,
        "fatal": true,
    }
    
    if config.Level != "" && !validLevels[config.Level] {
        return fmt.Errorf("invalid log level: %s", config.Level)
    }
    
    // 验证OTLP端点格式
    if config.OTLPEndpoint != "" {
        if !isValidEndpoint(config.OTLPEndpoint) {
            return fmt.Errorf("invalid OTLP endpoint: %s", config.OTLPEndpoint)
        }
    }
    
    return nil
}

func testLogger(log logger.Logger) error {
    defer func() {
        if r := recover(); r != nil {
            return
        }
    }()
    
    // 测试基本日志功能
    log.Debug("测试debug日志")
    log.Info("测试info日志")
    log.Warn("测试warn日志")
    
    // 测试结构化日志
    log.Infow("测试结构化日志", "key1", "value1", "key2", 42)
    
    // 测试上下文日志
    contextLog := log.With("component", "test")
    contextLog.Info("测试上下文日志")
    
    return nil
}

func isValidEndpoint(endpoint string) bool {
    // 简单的端点格式验证
    return strings.Contains(endpoint, ":") && !strings.HasPrefix(endpoint, ":")
}
```

### 2. OTLP连接故障诊断

```go
package main

import (
    "context"
    "net"
    "time"
    
    "github.com/costa92/go-protoc/v2/pkg/logger"
)

func diagnoseOTLPConnection(endpoint string) {
    fmt.Printf("🔍 诊断OTLP连接: %s\n", endpoint)
    
    // 1. 网络连通性测试
    if err := testNetworkConnectivity(endpoint); err != nil {
        fmt.Printf("❌ 网络连通性测试失败: %v\n", err)
        suggestNetworkFixes(endpoint)
        return
    } else {
        fmt.Printf("✅ 网络连通性测试通过\n")
    }
    
    // 2. OTLP服务可用性测试
    if err := testOTLPService(endpoint); err != nil {
        fmt.Printf("❌ OTLP服务测试失败: %v\n", err)
        suggestOTLPFixes(endpoint)
        return
    } else {
        fmt.Printf("✅ OTLP服务测试通过\n")
    }
    
    // 3. 日志发送测试
    if err := testLogSending(endpoint); err != nil {
        fmt.Printf("❌ 日志发送测试失败: %v\n", err)
        suggestLogSendingFixes(endpoint)
        return
    } else {
        fmt.Printf("✅ 日志发送测试通过\n")
    }
    
    fmt.Println("🎉 OTLP连接诊断全部通过！")
}

func testNetworkConnectivity(endpoint string) error {
    conn, err := net.DialTimeout("tcp", endpoint, 5*time.Second)
    if err != nil {
        return err
    }
    defer conn.Close()
    return nil
}

func testOTLPService(endpoint string) error {
    // 这里可以使用OTLP客户端进行更详细的服务测试
    // 简化版本仅测试gRPC连接
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // 模拟gRPC连接测试
    time.Sleep(100 * time.Millisecond)
    
    return nil
}

func testLogSending(endpoint string) error {
    config := &logger.QuickConfig{
        Preset:       logger.PresetObservability,
        Type:         "zap",
        Level:        "info",
        OTLPEndpoint: endpoint,
    }
    
    log, err := logger.NewLoggerFromQuickConfig(config)
    if err != nil {
        return fmt.Errorf("failed to create logger: %w", err)
    }
    
    // 发送测试日志
    log.Info("OTLP连接诊断测试日志", "test", true, "timestamp", time.Now())
    
    // 等待一下确保日志发送
    time.Sleep(2 * time.Second)
    
    return nil
}

func suggestNetworkFixes(endpoint string) {
    fmt.Println("\n🔧 网络连接问题解决建议:")
    fmt.Printf("1. 检查端点地址是否正确: %s\n", endpoint)
    fmt.Println("2. 检查防火墙设置")
    fmt.Println("3. 检查网络路由")
    fmt.Println("4. 确认目标服务是否运行")
    
    // Docker环境检查
    if isDockerEnvironment() {
        fmt.Println("\n🐳 Docker环境特殊检查:")
        fmt.Println("- 检查容器间网络连通性")
        fmt.Println("- 确认服务名解析")
        fmt.Println("- 检查docker-compose网络配置")
    }
    
    // Kubernetes环境检查
    if isKubernetesEnvironment() {
        fmt.Println("\n☸️ Kubernetes环境特殊检查:")
        fmt.Println("- 检查Service是否存在")
        fmt.Println("- 确认NetworkPolicy配置")
        fmt.Println("- 检查DNS解析")
    }
}

func suggestOTLPFixes(endpoint string) {
    fmt.Println("\n🔧 OTLP服务问题解决建议:")
    fmt.Println("1. 检查OTEL Collector是否正确启动")
    fmt.Println("2. 验证Collector配置文件")
    fmt.Println("3. 检查接收器(receiver)配置")
    fmt.Println("4. 查看Collector日志")
    
    fmt.Printf("\n📋 建议的检查命令:\n")
    fmt.Printf("# 检查Collector状态\n")
    fmt.Printf("docker ps | grep otel\n")
    fmt.Printf("kubectl get pods -l app=otelcol\n")
    
    fmt.Printf("\n# 查看Collector日志\n")
    fmt.Printf("docker logs <otel-container-id>\n")
    fmt.Printf("kubectl logs -l app=otelcol\n")
}

func suggestLogSendingFixes(endpoint string) {
    fmt.Println("\n🔧 日志发送问题解决建议:")
    fmt.Println("1. 检查日志器配置")
    fmt.Println("2. 增加OTLP超时时间")
    fmt.Println("3. 调整批处理参数")
    fmt.Println("4. 检查认证设置")
    
    fmt.Printf("\n📝 建议的配置调整:\n")
    fmt.Printf(`
otlp:
  timeout: "30s"           # 增加超时时间
  batch_timeout: "5s"      # 增加批处理超时
  batch_size: 50           # 减小批处理大小
  insecure: true           # 检查TLS设置
`)
}

func isDockerEnvironment() bool {
    // 检查是否在Docker容器中运行
    _, err := os.Stat("/.dockerenv")
    return err == nil
}

func isKubernetesEnvironment() bool {
    // 检查Kubernetes环境变量
    return os.Getenv("KUBERNETES_SERVICE_HOST") != ""
}
```

## 性能优化示例

### 1. 高吞吐量场景优化

```go
package main

import (
    "context"
    "runtime"
    "sync"
    "time"
    
    "github.com/costa92/go-protoc/v2/pkg/logger"
)

// 高性能日志配置
func createHighThroughputLogger() logger.Logger {
    config := &logger.QuickConfig{
        Preset: logger.PresetProduction,  // 生产环境预设，性能优化
        Type:   "zap",                   // Zap性能更好
        Level:  "warn",                  // 只记录重要日志
    }
    
    // 转换为详细配置进行优化
    opts := config.ToFullOptions()
    
    // 进一步性能优化
    opts.DisableCaller = true        // 禁用调用者信息
    opts.DisableStacktrace = true    // 禁用堆栈跟踪
    opts.DisableFunctionAtInfo = true // 禁用函数名
    
    // OTLP批处理优化
    if opts.OTLP != nil {
        opts.OTLP.BatchSize = 1000           // 大批次
        opts.OTLP.BatchTimeout = 5 * time.Second // 长超时
    }
    
    log, err := logger.NewLogger(opts)
    if err != nil {
        panic(err)
    }
    
    return log
}

// 并发日志测试
func benchmarkConcurrentLogging() {
    log := createHighThroughputLogger()
    
    numWorkers := runtime.NumCPU()
    messagesPerWorker := 10000
    
    var wg sync.WaitGroup
    start := time.Now()
    
    // 启动多个协程并发写日志
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            
            workerLog := log.With("worker_id", workerID)
            
            for j := 0; j < messagesPerWorker; j++ {
                workerLog.Warnw("高频业务日志",
                    "message_id", j,
                    "data", "some_important_data",
                    "timestamp", time.Now().Unix(),
                )
            }
        }(i)
    }
    
    wg.Wait()
    duration := time.Since(start)
    
    totalMessages := numWorkers * messagesPerWorker
    messagesPerSecond := float64(totalMessages) / duration.Seconds()
    
    fmt.Printf("📊 并发日志性能测试结果:\n")
    fmt.Printf("   协程数: %d\n", numWorkers)
    fmt.Printf("   每个协程消息数: %d\n", messagesPerWorker)
    fmt.Printf("   总消息数: %d\n", totalMessages)
    fmt.Printf("   总耗时: %v\n", duration)
    fmt.Printf("   吞吐量: %.0f 消息/秒\n", messagesPerSecond)
}

// 内存优化的日志使用
func memoryOptimizedLogging() {
    log := createHighThroughputLogger()
    
    // 预分配日志字段避免频繁分配
    commonFields := []interface{}{
        "service", "payment-service",
        "version", "v1.0.0",
    }
    
    serviceLog := log.With(commonFields...)
    
    // 重用日志器实例
    var loggerPool = sync.Pool{
        New: func() interface{} {
            return serviceLog.With("request_id", "")
        },
    }
    
    // 处理请求时从池中获取日志器
    for i := 0; i < 1000; i++ {
        reqLog := loggerPool.Get().(logger.Logger)
        reqLog = reqLog.With("request_id", fmt.Sprintf("req-%d", i))
        
        reqLog.Warn("处理支付请求", "amount", 99.99)
        
        loggerPool.Put(reqLog)
    }
}
```

### 2. 异步日志处理

```go
package main

import (
    "context"
    "time"
    
    "github.com/costa92/go-protoc/v2/pkg/logger"
)

// 异步日志处理器
type AsyncLogProcessor struct {
    log     logger.Logger
    buffer  chan LogEntry
    done    chan struct{}
}

type LogEntry struct {
    Level   string
    Message string
    Fields  map[string]interface{}
}

func NewAsyncLogProcessor(log logger.Logger, bufferSize int) *AsyncLogProcessor {
    processor := &AsyncLogProcessor{
        log:    log,
        buffer: make(chan LogEntry, bufferSize),
        done:   make(chan struct{}),
    }
    
    go processor.process()
    return processor
}

func (p *AsyncLogProcessor) process() {
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()
    
    batch := make([]LogEntry, 0, 100)
    
    for {
        select {
        case entry := <-p.buffer:
            batch = append(batch, entry)
            
            // 当批次满了或超时时处理
            if len(batch) >= 100 {
                p.processBatch(batch)
                batch = batch[:0]
            }
            
        case <-ticker.C:
            if len(batch) > 0 {
                p.processBatch(batch)
                batch = batch[:0]
            }
            
        case <-p.done:
            // 处理剩余的日志
            if len(batch) > 0 {
                p.processBatch(batch)
            }
            return
        }
    }
}

func (p *AsyncLogProcessor) processBatch(batch []LogEntry) {
    for _, entry := range batch {
        switch entry.Level {
        case "debug":
            p.log.Debugw(entry.Message, mapToKeyValues(entry.Fields)...)
        case "info":
            p.log.Infow(entry.Message, mapToKeyValues(entry.Fields)...)
        case "warn":
            p.log.Warnw(entry.Message, mapToKeyValues(entry.Fields)...)
        case "error":
            p.log.Errorw(entry.Message, mapToKeyValues(entry.Fields)...)
        }
    }
}

func (p *AsyncLogProcessor) LogAsync(level, message string, fields map[string]interface{}) {
    select {
    case p.buffer <- LogEntry{Level: level, Message: message, Fields: fields}:
    default:
        // 缓冲区满时的策略：丢弃或阻塞
        p.log.Warn("异步日志缓冲区满，丢弃日志")
    }
}

func (p *AsyncLogProcessor) Close() {
    close(p.done)
}

func mapToKeyValues(m map[string]interface{}) []interface{} {
    result := make([]interface{}, 0, len(m)*2)
    for k, v := range m {
        result = append(result, k, v)
    }
    return result
}

// 使用示例
func main() {
    log := createHighThroughputLogger()
    asyncProcessor := NewAsyncLogProcessor(log, 1000)
    defer asyncProcessor.Close()
    
    // 异步记录日志，不阻塞主流程
    for i := 0; i < 10000; i++ {
        asyncProcessor.LogAsync("info", "异步处理用户请求", map[string]interface{}{
            "user_id":    i,
            "request_id": fmt.Sprintf("req-%d", i),
            "timestamp":  time.Now(),
        })
    }
    
    time.Sleep(2 * time.Second) // 等待处理完成
}
```

这些示例涵盖了Logger包在实际项目中的各种使用场景，从基础配置到复杂的集成和优化。通过这些示例，用户可以快速理解如何在自己的项目中有效使用Logger包，并根据具体需求进行相应的配置和优化。