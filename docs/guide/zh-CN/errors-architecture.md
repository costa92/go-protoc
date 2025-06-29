# 错误处理系统架构设计

## 概述

本文档详细描述了 go-protoc 项目中统一错误处理系统的架构设计、核心组件和使用方法。该系统旨在提供一致、可维护、国际化的错误处理机制。

## 设计原则

### 1. 统一性
- 所有错误都使用统一的 `ErrorX` 结构
- 一致的错误响应格式
- 标准化的错误码和原因码

### 2. 可维护性
- 清晰的模块职责分离
- 集中的错误定义管理
- 完善的测试覆盖

### 3. 国际化支持
- 深度集成 i18n 系统
- 自动错误消息本地化
- 支持参数化消息模板

### 4. 可扩展性
- 灵活的错误注册机制
- 支持自定义错误类型
- 可插拔的错误处理器

### 5. 性能优化
- 错误对象复用
- 延迟国际化
- 最小化内存分配

## 核心架构

### 系统层次结构

```
┌─────────────────────────────────────────────────────────────┐
│                    应用层 (Handlers/Services)                │
├─────────────────────────────────────────────────────────────┤
│                    业务错误层 (pkg/errors)                   │
├─────────────────────────────────────────────────────────────┤
│                   核心错误层 (pkg/errorsx)                   │
├─────────────────────────────────────────────────────────────┤
│                   国际化层 (pkg/i18n)                       │
├─────────────────────────────────────────────────────────────┤
│                   中间件层 (HTTP/gRPC)                      │
└─────────────────────────────────────────────────────────────┘
```

### 模块组成

#### 1. 核心错误模块 (`pkg/errorsx`)

**职责**: 提供统一的错误处理基础设施

- `errorsx.go`: 核心错误结构和接口
- `builder.go`: 错误构建器模式
- `code.go`: 标准 HTTP 状态码
- `i18n.go`: 国际化集成
- `registry.go`: 错误模板注册器
- `middleware.go`: HTTP/gRPC 中间件
- `wrap.go`: 错误包装工具

#### 2. 业务错误模块 (`pkg/errors`)

**职责**: 定义具体的业务错误类型

- `user.go`: 用户相关错误
- `auth.go`: 认证授权错误
- `common.go`: 通用业务错误
- `errors.go`: 兼容性错误定义

#### 3. 国际化模块 (`pkg/i18n`)

**职责**: 提供多语言支持

- 错误消息本地化
- 参数化消息模板
- 语言环境检测

## 核心组件详解

### 1. ErrorX 结构

```go
type ErrorX struct {
    Code     int32             `json:"code"`     // HTTP 状态码
    Reason   string            `json:"reason"`   // 错误原因码
    Message  string            `json:"message"`  // 错误消息
    Metadata map[string]any    `json:"metadata,omitempty"` // 元数据
    
    // 内部字段
    i18nKey  string            // 国际化键
    cause    error             // 原始错误
}
```

**特性**:
- 实现标准 `error` 接口
- 支持错误链 (`Unwrap`, `Is`, `As`)
- 丰富的元数据支持
- 自动国际化

### 2. Builder 构建器

```go
type Builder struct {
    code     int32
    reason   string
    message  string
    i18nKey  string
    metadata map[string]any
    cause    error
}
```

**使用模式**:
```go
err := errorsx.BadRequest().
    WithReason("INVALID_PARAMETER").
    WithMessage("Invalid email format").
    WithI18nKey("errors.validation.invalid_email").
    AddMetadata("field", "email").
    AddMetadata("value", "invalid-email").
    Build()
```

### 3. Registry 注册器

```go
type Registry struct {
    templates map[string]*ErrorTemplate
    mutex     sync.RWMutex
}

type ErrorTemplate struct {
    Code    int32
    Reason  string
    I18nKey string
}
```

**功能**:
- 错误模板注册和管理
- 线程安全的访问
- 全局注册器支持

### 4. 中间件系统

#### HTTP 中间件
```go
func GinErrorMiddleware(handler ErrorHandler) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            response := handler.HandleError(c.Request.Context(), err)
            c.JSON(int(response.Code), response)
        }
    }
}
```

#### gRPC 中间件
```go
func GRPCErrorInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        resp, err := handler(ctx, req)
        if err != nil {
            if errorX, ok := err.(*ErrorX); ok {
                return resp, errorX.GRPCStatus().Err()
            }
        }
        return resp, err
    }
}
```

## 错误分类体系

### 1. 按 HTTP 状态码分类

| 状态码 | 类型 | 说明 | 示例 |
|--------|------|------|------|
| 400 | 客户端错误 | 请求参数错误 | 参数验证失败 |
| 401 | 认证错误 | 身份验证失败 | Token 无效 |
| 403 | 授权错误 | 权限不足 | 访问被拒绝 |
| 404 | 资源错误 | 资源不存在 | 用户未找到 |
| 409 | 冲突错误 | 资源冲突 | 用户已存在 |
| 429 | 限流错误 | 请求过于频繁 | 速率限制 |
| 500 | 服务器错误 | 内部错误 | 数据库连接失败 |
| 502 | 网关错误 | 外部服务错误 | 第三方 API 失败 |
| 503 | 服务不可用 | 服务暂停 | 维护模式 |

### 2. 按业务领域分类

#### 用户相关错误
- `USER_NOT_FOUND`: 用户不存在
- `USER_ALREADY_EXISTS`: 用户已存在
- `USER_DISABLED`: 用户被禁用
- `USER_PASSWORD_WEAK`: 密码强度不足

#### 认证相关错误
- `TOKEN_INVALID`: Token 无效
- `TOKEN_EXPIRED`: Token 过期
- `TOKEN_MISSING`: Token 缺失
- `LOGIN_FAILED`: 登录失败
- `INSUFFICIENT_PERMISSIONS`: 权限不足

#### 通用业务错误
- `RESOURCE_NOT_FOUND`: 资源不存在
- `INVALID_REQUEST`: 请求无效
- `RATE_LIMIT_EXCEEDED`: 速率限制
- `DATABASE_CONNECTION_ERROR`: 数据库连接错误

## 国际化集成

### 1. 消息模板

```yaml
# locales/en.yaml
errors:
  user:
    not_found: "User not found"
    already_exists: "User already exists"
    password_weak: "Password is too weak"
  auth:
    token_invalid: "Invalid token"
    token_expired: "Token has expired"
    login_failed: "Login failed"
```

```yaml
# locales/zh-CN.yaml
errors:
  user:
    not_found: "用户不存在"
    already_exists: "用户已存在"
    password_weak: "密码强度不足"
  auth:
    token_invalid: "无效的令牌"
    token_expired: "令牌已过期"
    login_failed: "登录失败"
```

### 2. 参数化消息

```yaml
errors:
  validation:
    min_length: "Field {{.field}} must be at least {{.min_length}} characters"
    max_length: "Field {{.field}} cannot exceed {{.max_length}} characters"
    invalid_format: "Field {{.field}} has invalid format: {{.format}}"
```

### 3. 使用示例

```go
// 自动国际化
err := errors.NewUserNotFoundError("123")
localizedErr := err.Localize(ctx) // 根据 ctx 中的语言环境自动本地化

// 参数化消息
err := errorsx.BadRequest().
    WithI18nKey("errors.validation.min_length").
    AddMetadata("field", "username").
    AddMetadata("min_length", 3).
    BuildWithContext(ctx)
```

## 性能优化策略

### 1. 错误对象复用

```go
// 预定义错误模板
var (
    ErrUserNotFound = errorsx.NotFound().
        WithReason("USER_NOT_FOUND").
        WithI18nKey("errors.user.not_found").
        Build()
)

// 复用模板创建具体错误
func NewUserNotFoundError(userID string) *errorsx.ErrorX {
    return ErrUserNotFound.WithMetadata(map[string]any{
        "user_id": userID,
    })
}
```

### 2. 延迟国际化

```go
// 只在需要时进行国际化
func (e *ErrorX) Localize(ctx context.Context) *ErrorX {
    if e.i18nKey == "" {
        return e
    }
    
    // 延迟执行国际化
    localizedMessage := globalI18n.Localize(ctx, e.i18nKey, e.Metadata)
    return e.WithMessage(localizedMessage)
}
```

### 3. 内存池优化

```go
var builderPool = sync.Pool{
    New: func() interface{} {
        return &Builder{}
    },
}

func NewBuilder() *Builder {
    b := builderPool.Get().(*Builder)
    b.Reset()
    return b
}

func (b *Builder) Release() {
    builderPool.Put(b)
}
```

## 最佳实践

### 1. 错误定义

```go
// ✅ 好的做法
var ErrUserNotFound = errorsx.NotFound().
    WithReason("USER_NOT_FOUND").
    WithI18nKey("errors.user.not_found").
    Build()

// ❌ 避免的做法
var ErrUserNotFound = errors.New("user not found")
```

### 2. 错误创建

```go
// ✅ 好的做法 - 使用构建器
func GetUser(id string) (*User, error) {
    user, err := repo.GetByID(id)
    if err != nil {
        if isNotFoundError(err) {
            return nil, errors.NewUserNotFoundError(id)
        }
        return nil, errorsx.InternalError().
            WithReason("DATABASE_ERROR").
            WithCause(err).
            Build()
    }
    return user, nil
}

// ❌ 避免的做法 - 直接返回原始错误
func GetUser(id string) (*User, error) {
    return repo.GetByID(id) // 直接返回数据库错误
}
```

### 3. 错误处理

```go
// ✅ 好的做法 - 类型安全的错误检查
if errorsx.IsCode(err, 404) {
    // 处理资源不存在
}

if errorsx.IsReason(err, "USER_NOT_FOUND") {
    // 处理用户不存在
}

// ❌ 避免的做法 - 字符串匹配
if strings.Contains(err.Error(), "not found") {
    // 不可靠的错误检查
}
```

### 4. 元数据使用

```go
// ✅ 好的做法 - 结构化元数据
err := errors.NewValidationError("email", "invalid format").
    AddMetadata("field", "email").
    AddMetadata("value", email).
    AddMetadata("expected_format", "user@domain.com").
    AddMetadata("validation_rule", "email_format")

// ❌ 避免的做法 - 在消息中包含所有信息
err := fmt.Errorf("validation failed: email '%s' is invalid, expected format: user@domain.com", email)
```

### 5. 国际化集成

```go
// ✅ 好的做法 - 使用 i18n 键
err := errorsx.BadRequest().
    WithI18nKey("errors.validation.required_field").
    AddMetadata("field", "username")

// ❌ 避免的做法 - 硬编码消息
err := errorsx.BadRequest().
    WithMessage("Username is required")
```

## 监控和可观测性

### 1. 错误指标

```go
// 错误计数器
var errorCounter = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "api_errors_total",
        Help: "Total number of API errors",
    },
    []string{"code", "reason", "endpoint"},
)

// 在中间件中记录指标
func (h *DefaultErrorHandler) HandleError(ctx context.Context, err error) *ErrorResponse {
    errorX := h.convertToErrorX(err)
    
    // 记录指标
    errorCounter.WithLabelValues(
        strconv.Itoa(int(errorX.Code)),
        errorX.Reason,
        getEndpoint(ctx),
    ).Inc()
    
    return h.buildResponse(ctx, errorX)
}
```

### 2. 结构化日志

```go
func (h *DefaultErrorHandler) HandleError(ctx context.Context, err error) *ErrorResponse {
    errorX := h.convertToErrorX(err)
    
    // 结构化日志
    logger.WithContext(ctx).WithFields(map[string]interface{}{
        "error_code":   errorX.Code,
        "error_reason": errorX.Reason,
        "error_metadata": errorX.Metadata,
        "request_id":  getRequestID(ctx),
        "user_id":     getUserID(ctx),
    }).Error("API error occurred")
    
    return h.buildResponse(ctx, errorX)
}
```

### 3. 错误追踪

```go
// 错误链追踪
func traceErrorChain(err error) []string {
    var chain []string
    for err != nil {
        if errorX, ok := err.(*errorsx.ErrorX); ok {
            chain = append(chain, fmt.Sprintf("%s: %s", errorX.Reason, errorX.Message))
            err = errorX.Unwrap()
        } else {
            chain = append(chain, err.Error())
            break
        }
    }
    return chain
}
```

## 扩展指南

### 1. 自定义错误类型

```go
// 定义新的业务错误
package payment

import "github.com/costa92/go-protoc/pkg/errorsx"

// 支付相关错误
var (
    ErrPaymentFailed = errorsx.BadRequest().
        WithReason("PAYMENT_FAILED").
        WithI18nKey("errors.payment.failed").
        Build()
        
    ErrInsufficientFunds = errorsx.BadRequest().
        WithReason("INSUFFICIENT_FUNDS").
        WithI18nKey("errors.payment.insufficient_funds").
        Build()
)

// 构建函数
func NewPaymentFailedError(transactionID string, reason string) *errorsx.ErrorX {
    return ErrPaymentFailed.WithMetadata(map[string]any{
        "transaction_id": transactionID,
        "failure_reason": reason,
    })
}
```

### 2. 自定义错误处理器

```go
type CustomErrorHandler struct {
    logger  *log.Logger
    metrics *prometheus.Registry
}

func (h *CustomErrorHandler) HandleError(ctx context.Context, err error) *errorsx.ErrorResponse {
    // 自定义错误处理逻辑
    errorX := h.convertError(err)
    
    // 记录日志
    h.logError(ctx, errorX)
    
    // 记录指标
    h.recordMetrics(ctx, errorX)
    
    // 发送告警
    if errorX.Code >= 500 {
        h.sendAlert(ctx, errorX)
    }
    
    return &errorsx.ErrorResponse{
        Code:      errorX.Code,
        Reason:    errorX.Reason,
        Message:   errorX.Message,
        Metadata:  errorX.Metadata,
        RequestID: getRequestID(ctx),
        Timestamp: time.Now().Unix(),
    }
}
```

### 3. 中间件扩展

```go
// 自定义 gRPC 拦截器
func CustomGRPCErrorInterceptor(logger *log.Logger) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        resp, err := handler(ctx, req)
        
        if err != nil {
            // 转换错误
            if errorX, ok := err.(*errorsx.ErrorX); ok {
                // 记录详细日志
                logger.WithContext(ctx).WithFields(map[string]interface{}{
                    "grpc_method": info.FullMethod,
                    "error_code":  errorX.Code,
                    "error_reason": errorX.Reason,
                }).Error("gRPC error")
                
                // 返回 gRPC 状态
                return resp, errorX.GRPCStatus().Err()
            }
        }
        
        return resp, err
    }
}
```

## 总结

本错误处理系统通过统一的架构设计，提供了：

1. **一致性**: 统一的错误格式和处理流程
2. **可维护性**: 清晰的模块分离和集中管理
3. **国际化**: 深度集成的多语言支持
4. **可扩展性**: 灵活的扩展机制
5. **高性能**: 优化的内存使用和处理效率
6. **可观测性**: 完善的监控和日志记录

通过遵循本文档中的设计原则和最佳实践，可以构建出健壮、可维护的错误处理系统，提升整体系统的质量和用户体验。