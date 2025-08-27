# 错误处理故障排查指南

## 概述

本文档提供了错误处理相关问题的诊断和解决方案，帮助开发者快速定位和修复常见问题。

## 常见问题诊断

### 1. 错误码问题

#### 问题：错误码总是返回 50000（默认错误码）

**症状**:
```json
{
  "code": 50000,
  "reason": "UserAlreadyExists",
  "message": "用户已存在"
}
```

**原因分析**:
1. 错误码映射器未正确注册
2. 错误原因不在映射表中
3. 应用启动时未执行初始化代码

**诊断步骤**:

1. **检查错误码映射器注册**:
```bash
# 搜索错误码映射器的注册代码
grep -r "RegisterErrorCodeMapper" .
grep -r "HTTPStatusCodeMapper" .
```

2. **检查包导入**:
```bash
# 确认 error_mapper.go 所在的包被导入
grep -r "internal/apiserver" cmd/
```

3. **验证错误原因映射**:
```go
// 在代码中添加调试信息
func debugErrorMapping() {
    if code, ok := v1.ErrorReason_value["UserAlreadyExists"]; ok {
        log.Printf("UserAlreadyExists mapped to: %d", code)
    } else {
        log.Printf("UserAlreadyExists not found in mapping")
    }
}
```

**解决方案**:

1. **确保错误码映射器正确注册**:
```go
// internal/apiserver/error_mapper.go
package apiserver

import (
    "github.com/costa92/go-protoc/v2/pkg/api/apiserver/v1"
    "github.com/costa92/go-protoc/v2/pkg/server"
)

func init() {
    mapper := server.NewHTTPStatusCodeMapper(v1.ErrorReason_value)
    server.RegisterErrorCodeMapper(mapper)
}
```

2. **确保包被正确导入**:
```go
// cmd/apiserver/app/server.go
import (
    _ "github.com/costa92/go-protoc/v2/internal/apiserver" // 确保 init 函数执行
)
```

3. **验证修复**:
```bash
# 重新构建并测试
go build ./cmd/apiserver/
# 启动服务并测试错误响应
```

#### 问题：自定义错误码不生效

**症状**:
定义了新的错误码，但总是返回通用错误码。

**解决方案**:

1. **重新生成 protobuf 代码**:
```bash
./generate.sh
```

2. **检查错误定义**:
```protobuf
// pkg/api/apiserver/v1/errors.proto
enum ErrorReason {
  option (errors.default_code) = 500;
  
  USER_LOGIN_FAILED = 0 [(errors.code) = 401];
  USER_ALREADY_EXISTS = 1 [(errors.code) = 409];
  // 确保新错误有正确的错误码
  MY_CUSTOM_ERROR = 2 [(errors.code) = 422];
}
```

3. **重新注册映射器**:
```bash
# 重启应用以重新加载映射
```

### 2. 国际化问题

#### 问题：错误消息不支持国际化

**症状**:
无论客户端语言设置如何，错误消息总是显示为默认语言。

**诊断步骤**:

1. **检查国际化中间件**:
```bash
grep -r "i18n" internal/pkg/middleware/
```

2. **检查语言配置文件**:
```bash
ls -la configs/locales/
```

3. **验证请求头**:
```bash
# 使用 curl 测试
curl -H "Accept-Language: zh-CN" http://localhost:8080/api/v1/users
```

**解决方案**:

1. **确保国际化中间件正确配置**:
```go
// 在 HTTP 服务器中添加国际化中间件
func NewHTTPServer(cfg *config.Config) *http.Server {
    srv := http.NewServer(
        http.Middleware(
            middleware.I18n(), // 确保添加了国际化中间件
            // 其他中间件...
        ),
    )
    return srv
}
```

2. **检查国际化配置文件**:
```yaml
# configs/locales/zh-CN.yaml
user:
  already:
    exists: "用户已存在"
  not:
    found: "用户未找到"

# configs/locales/en-US.yaml
user:
  already:
    exists: "User already exists"
  not:
    found: "User not found"
```

3. **在代码中正确使用国际化**:
```go
// 正确的用法
message := i18n.FromContext(ctx).T(locales.UserAlreadyExists)
return v1.ErrorUserAlreadyExists(message)

// 错误的用法
return v1.ErrorUserAlreadyExists("用户已存在") // 硬编码消息
```

#### 问题：国际化键值未找到

**症状**:
错误消息显示为键值本身，如 "user.already.exists"。

**解决方案**:

1. **检查键值定义**:
```go
// internal/apiserver/pkg/locales/locales.go
const (
    UserAlreadyExists = "user.already.exists" // 确保键值正确
)
```

2. **检查配置文件路径**:
```yaml
# 确保配置文件结构正确
user:
  already:
    exists: "用户已存在"
```

3. **验证配置加载**:
```go
// 添加调试代码
func debugI18n(ctx context.Context) {
    i18nInstance := i18n.FromContext(ctx)
    message := i18nInstance.T("user.already.exists")
    log.Printf("Translated message: %s", message)
}
```

### 3. 验证问题

#### 问题：验证错误不正确

**症状**:
验证失败但返回了错误的错误码或消息。

**诊断步骤**:

1. **检查验证器注册**:
```bash
grep -r "RequestValidator" internal/
```

2. **检查验证逻辑**:
```go
// 在验证方法中添加日志
func (v *UserValidator) ValidateCreateUserRequest(ctx context.Context, rq any) error {
    log.Printf("Validating request: %+v", rq)
    
    req, ok := rq.(*v1.CreateUserRequest)
    if !ok {
        log.Printf("Invalid request type: %T", rq)
        return errno.ErrorInvalidParameter("invalid request type")
    }
    
    // 验证逻辑...
}
```

**解决方案**:

1. **确保验证器正确注册**:
```go
// internal/apiserver/httpserver.go
func NewHTTPServer() *http.Server {
    validator := validation.NewValidator()
    validator.RegisterValidator("CreateUserRequest", &UserValidator{})
    
    srv := http.NewServer(
        http.Middleware(
            validate.Validator(validator),
        ),
    )
    return srv
}
```

2. **检查验证方法签名**:
```go
// 确保实现了正确的接口
type RequestValidator interface {
    Validate(ctx context.Context, rq any) error
}
```

### 4. 日志问题

#### 问题：错误日志信息不足

**症状**:
错误发生时日志信息不够详细，难以定位问题。

**解决方案**:

1. **增强错误日志**:
```go
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    logger := s.logger.WithFields(logrus.Fields{
        "operation": "create_user",
        "username":  req.Username,
        "trace_id":  contextx.TraceID(ctx),
        "user_id":   contextx.UserID(ctx),
    })
    
    logger.Info("Starting user creation")
    
    if err := s.validateUser(req); err != nil {
        logger.WithError(err).Warn("User validation failed")
        return nil, err
    }
    
    user, err := s.repo.CreateUser(ctx, req)
    if err != nil {
        logger.WithError(err).Error("Failed to create user in repository")
        return nil, err
    }
    
    logger.WithField("user_id", user.ID).Info("User created successfully")
    return user, nil
}
```

2. **配置结构化日志**:
```go
// 使用结构化日志格式
logger := logrus.New()
logger.SetFormatter(&logrus.JSONFormatter{})
logger.SetLevel(logrus.InfoLevel)
```

## 调试工具和技巧

### 1. 错误追踪

**添加请求追踪**:
```go
func (h *Handler) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
    traceID := contextx.TraceID(ctx)
    
    h.logger.WithField("trace_id", traceID).Info("Request started")
    defer h.logger.WithField("trace_id", traceID).Info("Request completed")
    
    // 业务逻辑...
}
```

### 2. 错误码验证脚本

**创建验证脚本**:
```bash
#!/bin/bash
# scripts/verify-error-codes.sh

echo "验证错误码映射..."

# 检查所有错误定义
echo "检查 protobuf 错误定义:"
grep -r "errors.code" pkg/api/

# 检查映射器注册
echo "检查映射器注册:"
grep -r "RegisterErrorCodeMapper" .

# 测试错误响应
echo "测试错误响应:"
curl -s -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"existing_user"}' | jq .
```

### 3. 错误监控

**设置错误监控**:
```go
// 错误计数器
var errorCounter = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "api_errors_total",
        Help: "Total API errors",
    },
    []string{"method", "reason", "code"},
)

// 在错误处理中记录指标
func recordError(method, reason string, code int) {
    errorCounter.WithLabelValues(method, reason, strconv.Itoa(code)).Inc()
}
```

## 性能问题排查

### 1. 错误处理性能

**检查错误处理开销**:
```go
// 使用 benchmark 测试
func BenchmarkErrorCreation(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _ = v1.ErrorUserAlreadyExists("test message")
    }
}

func BenchmarkErrorWithI18n(b *testing.B) {
    ctx := context.Background()
    for i := 0; i < b.N; i++ {
        message := i18n.FromContext(ctx).T(locales.UserAlreadyExists)
        _ = v1.ErrorUserAlreadyExists(message)
    }
}
```

### 2. 内存泄露检查

**检查错误对象内存使用**:
```bash
# 使用 pprof 分析内存使用
go tool pprof http://localhost:8080/debug/pprof/heap

# 在 pprof 中查看错误相关的内存分配
(pprof) top -cum
(pprof) list ErrorUserAlreadyExists
```

## 预防措施

### 1. 代码审查检查清单

- [ ] 错误码定义正确
- [ ] 错误消息支持国际化
- [ ] 错误日志信息充分
- [ ] 错误处理测试覆盖
- [ ] 敏感信息不泄露

### 2. 自动化测试

**错误处理集成测试**:
```go
func TestErrorHandling(t *testing.T) {
    // 启动测试服务器
    server := setupTestServer()
    defer server.Close()
    
    tests := []struct {
        name         string
        request      interface{}
        expectedCode int
        expectedReason string
    }{
        {
            name: "user_already_exists",
            request: &v1.CreateUserRequest{Name: "existing_user"},
            expectedCode: 409,
            expectedReason: "UserAlreadyExists",
        },
        // 更多测试用例...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 发送请求并验证错误响应
            resp := sendRequest(server.URL, tt.request)
            assert.Equal(t, tt.expectedCode, resp.Code)
            assert.Equal(t, tt.expectedReason, resp.Reason)
        })
    }
}
```

### 3. 监控和告警

**设置关键指标监控**:
```yaml
# prometheus 告警规则
groups:
- name: error_handling
  rules:
  - alert: HighErrorRate
    expr: rate(api_errors_total[5m]) > 10
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "High API error rate detected"
      
  - alert: DefaultErrorCodeUsage
    expr: rate(api_errors_total{code="50000"}[5m]) > 1
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "Default error code being used"
```

## 总结

通过系统性的故障排查方法，可以快速定位和解决错误处理相关的问题：

1. **问题分类** - 按错误码、国际化、验证等分类处理
2. **诊断工具** - 使用日志、监控、调试脚本等工具
3. **预防措施** - 通过代码审查、测试、监控预防问题
4. **持续改进** - 根据问题反馈不断优化错误处理机制

记住，良好的错误处理不仅要解决当前问题，更要预防未来问题的发生。