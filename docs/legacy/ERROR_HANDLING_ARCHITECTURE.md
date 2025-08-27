# 错误处理架构设计文档

本文档详细说明项目的错误处理架构设计、实现原理和最佳实践。

## 目录

- [架构概览](#架构概览)
- [设计原则](#设计原则)
- [两层架构详解](#两层架构详解)
- [核心组件](#核心组件)
- [使用指南](#使用指南)
- [性能优化](#性能优化)
- [最佳实践](#最佳实践)
- [常见问题](#常见问题)
- [迁移指南](#迁移指南)

## 架构概览

项目采用**两层错误处理架构**，实现了框架与业务的完美分离：

```
┌─────────────────────────────────────┐
│           业务应用层                │
├─────────────────────────────────────┤
│         pkg/errors (业务层)        │ ← 业务错误目录
├─────────────────────────────────────┤  
│        pkg/errorsx (框架层)        │ ← 错误处理引擎
├─────────────────────────────────────┤
│            底层基础设施              │
└─────────────────────────────────────┘
```

### 核心组件关系

```go
业务代码 → pkg/errors → pkg/errorsx → HTTP/gRPC响应
   ↓           ↓            ↓            ↓
验证逻辑    领域错误目录   错误处理引擎   协议转换
```

**优势**：
- ✅ **关注点分离**: 框架专注机制，业务专注语义
- ✅ **可扩展性强**: 业务错误独立演化，不影响框架
- ✅ **维护性好**: 职责清晰，降低耦合度
- ✅ **性能优化**: 框架层统一优化，业务层受益

## 设计原则

### 1. 框架与业务分离

**框架层 (errorsx)**：
- 提供错误处理的核心机制
- 不包含具体业务语义
- 专注性能优化和协议支持

**业务层 (errors)**：
- 定义领域特定错误
- 提供业务上下文信息
- 基于框架层构建

### 2. 结构化错误信息

每个错误都包含完整的结构化信息：

```go
type ErrorX struct {
    Code      int32          // HTTP状态码 (API交互)
    Reason    string         // 业务错误码 (精准定位)
    Message   string         // 用户友好消息
    Metadata  map[string]any // 上下文信息
    RequestID string         // 请求追踪
    i18nKey   string         // 国际化键
    cause     error          // 原始错误链
}
```

### 3. 链式构建模式

支持流畅的链式API，逐步丰富错误信息：

```go
err := errorsx.New(404, "USER_NOT_FOUND", "User not found").
    WithI18nKey("errors.user.not_found").
    AddMetadata("user_id", userID).
    WithRequestID(requestID)
```

### 4. 性能优先设计

- **对象池复用**: StringBuilder、Metadata Map
- **字符串缓存**: 预建常用错误码字符串  
- **智能采样**: 高频错误自动采样，避免日志洪流
- **写时复制**: 链式调用高效且安全

## 两层架构详解

### pkg/errorsx (框架层)

**职责**：
- 提供 `ErrorX` 核心结构体
- 实现错误处理中间件
- 支持国际化和本地化
- HTTP/gRPC 协议转换
- 性能优化和对象池管理

**核心API**：

```go
// 基础错误创建
func New(code int32, reason, message string) *ErrorX

// 错误包装
func Wrap(err error, code int32, reason, message string) *ErrorX

// 错误转换
func FromError(err error) *ErrorX

// 状态码提取
func Code(err error) int32
func Reason(err error) string
```

**预定义错误**：
```go
var (
    ErrInternal           = New(500, "INTERNAL_ERROR", "Internal server error")
    ErrNotFound           = New(404, "NOT_FOUND", "Resource not found")
    ErrValidation         = New(422, "VALIDATION_ERROR", "Validation failed")
    ErrUnauthenticated    = New(401, "UNAUTHENTICATED", "Authentication required")
    ErrPermissionDenied   = New(403, "PERMISSION_DENIED", "Permission denied")
    // ... 更多通用错误
)
```

### pkg/errors (业务层)

**职责**：
- 定义具体业务错误
- 提供丰富的构建器函数
- 维护错误分类和组织
- 支持业务特定的上下文信息

**组织结构**：
```
pkg/errors/
├── doc.go       # 包文档和使用指南
├── user.go      # 用户相关错误
├── auth.go      # 认证相关错误  
├── common.go    # 通用业务错误
└── errors.go    # 特定领域错误（兼容旧版）
```

**错误定义模式**：

```go
// 1. 预定义基础错误
var ErrUserNotFound = errorsx.New(404, "USER_NOT_FOUND", "User not found").
    WithI18nKey("errors.user.not_found")

// 2. 构建器函数添加上下文
func NewUserNotFoundError(userID string) *errorsx.ErrorX {
    return ErrUserNotFound.
        WithMessage("User with ID %s not found", userID).
        AddMetadata("user_id", userID)
}

// 3. 复杂构建器函数
func NewUserPermissionDeniedError(userID, resource, action string) *errorsx.ErrorX {
    return ErrUserPermissionDenied.
        WithMessage("User %s denied %s access to %s", userID, action, resource).
        AddMetadata("user_id", userID).
        AddMetadata("resource", resource).
        AddMetadata("action", action)
}
```

## 核心组件

### 1. ErrorX 结构体

项目错误处理的核心数据结构：

```go
type ErrorX struct {
    Code      int32          `json:"code,omitempty"`      // HTTP状态码
    Reason    string         `json:"reason,omitempty"`    // 业务错误码
    Message   string         `json:"message,omitempty"`   // 错误消息
    Metadata  map[string]any `json:"metadata,omitempty"`  // 元数据
    RequestID string         `json:"request_id,omitempty"`// 请求ID
    
    // 内部字段
    i18nKey string // 国际化键
    cause   error  // 原始错误
}
```

**关键方法**：
- `Error() string`: 实现标准错误接口
- `WithMessage()`: 设置错误消息
- `AddMetadata()`: 添加元数据
- `WithI18nKey()`: 设置国际化键
- `WithCause()`: 包装原始错误
- `GRPCStatus()`: 转换为gRPC状态

### 2. 中间件系统

#### Gin 中间件

```go
// 错误处理中间件
func GinErrorMiddleware(handler ErrorHandler) gin.HandlerFunc

// Panic 恢复中间件  
func RecoverMiddleware(handler ErrorHandler) gin.HandlerFunc

// 使用示例
errorHandler := errorsx.NewDefaultErrorHandler(logger)
router.Use(errorsx.GinErrorMiddleware(errorHandler))
router.Use(errorsx.RecoverMiddleware(errorHandler))
```

#### HTTP 标准库中间件

```go
// HTTP错误处理中间件
func HTTPErrorMiddleware(handler ErrorHandler) func(http.Handler) http.Handler

// 使用示例
handler = errorsx.HTTPErrorMiddleware(errorHandler)(handler)
```

### 3. 国际化系统

#### 上下文感知的本地化

```go
// 自动检测语言并本地化错误
func LocalizeError(ctx context.Context, err *ErrorX) *ErrorX

// 使用示例
localizedErr := errorsx.LocalizeError(ctx, err)
```

#### 支持的特性

- 基于 `Accept-Language` 头的自动检测
- 占位符和参数化消息支持
- 回退机制（英语作为默认语言）
- 上下文感知的语言切换

### 4. 性能优化组件

#### 对象池系统

```go
// 字符串构建器对象池
var stringBuilderPool = sync.Pool{
    New: func() interface{} {
        builder := &strings.Builder{}
        builder.Grow(256) // 预分配缓冲区
        return builder
    },
}

// 元数据对象池
var metadataPool = sync.Pool{
    New: func() interface{} {
        return make(map[string]any, 8)
    },
}
```

#### 智能日志采样

```go
// 错误采样，避免高频错误造成日志洪流
func shouldSkipLogging(err *ErrorX) bool {
    // 500级别错误始终记录
    if err.Code >= 500 {
        return false
    }
    
    // 前10次记录，之后每100次记录一次
    // ...
}
```

## 使用指南

### 基础用法

#### 1. 使用预定义错误

```go
import "github.com/costa92/go-protoc/v2/pkg/errors"

// 简单使用
if user == nil {
    return errors.ErrUserNotFound
}

// 添加上下文
return errors.NewUserNotFoundError(userID)
```

#### 2. 错误判断

```go
import (
    "github.com/costa92/go-protoc/v2/pkg/errors"
    "github.com/costa92/go-protoc/v2/pkg/errorsx"
)

// 按状态码判断
if errorsx.Code(err) == 404 {
    // 处理未找到错误
}

// 按业务错误码判断
if errorsx.Reason(err) == "USER_NOT_FOUND" {
    // 精确的业务错误处理
}

// 错误链判断
if errors.Is(err, errors.ErrUserNotFound) {
    // 使用标准库的错误判断
}
```

#### 3. 业务层错误创建

```go
// 在 biz 层创建业务错误
func (b *userBiz) GetUser(ctx context.Context, id string) (*User, error) {
    user, err := b.store.GetUser(ctx, id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.NewUserNotFoundError(id)
        }
        return nil, errorsx.Wrap(err, 500, "DATABASE_ERROR", "Failed to query user")
    }
    
    if user.Status == UserStatusDisabled {
        return nil, errors.NewUserInactiveError(id, "disabled")
    }
    
    return user, nil
}
```

### 高级用法

#### 1. 自定义错误处理器

```go
type CustomErrorHandler struct {
    logger logger.Logger
    tracer trace.Tracer
}

func (h *CustomErrorHandler) HandleError(ctx context.Context, err error) *errorsx.ErrorResponse {
    errorX := errorsx.FromError(err)
    
    // 添加链路追踪信息
    if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
        errorX = errorX.AddMetadata("trace_id", span.SpanContext().TraceID().String())
    }
    
    // 自定义日志记录
    h.logError(ctx, errorX)
    
    // 构建响应
    return &errorsx.ErrorResponse{
        Code:      errorX.Code,
        Reason:    errorX.Reason,
        Message:   errorX.Message,
        Metadata:  errorX.Metadata,
        Timestamp: time.Now().Format(time.RFC3339),
    }
}
```

#### 2. 批量错误处理

```go
func ValidateUserInput(input *UserInput) error {
    var errs []error
    
    if input.Email == "" {
        errs = append(errs, errors.NewUserEmailInvalidError(input.Email, "required"))
    }
    
    if input.Password == "" {
        errs = append(errs, errors.NewUserPasswordInvalidError("", "required"))
    }
    
    if len(errs) > 0 {
        return errorsx.New(400, "VALIDATION_ERROR", "Multiple validation errors").
            WithMetadata(map[string]any{"errors": errs})
    }
    
    return nil
}
```

#### 3. 错误监控和指标

```go
// 错误指标收集
func (err *ErrorX) Track() *ErrorX {
    // Prometheus 指标
    errorCounter.WithLabelValues(err.Reason, strconv.Itoa(int(err.Code))).Inc()
    
    // OpenTelemetry 事件
    span := trace.SpanFromContext(context.Background())
    span.AddEvent("error.occurred", trace.WithAttributes(
        attribute.String("error.reason", err.Reason),
        attribute.Int("error.code", int(err.Code)),
    ))
    
    return err
}
```

## 性能优化

### 1. 内存优化

**对象池复用**：
- `StringBuilder`: 减少字符串拼接分配
- `Metadata Map`: 减少map创建开销
- 预分配缓冲区大小，减少扩容

**字符串缓存**：
```go
// 预建常用错误码字符串
var errorCodeStrings = map[int32]string{
    200: "200", 400: "400", 401: "401", 403: "403", 404: "404",
    500: "500", 502: "502", 503: "503",
}
```

### 2. 日志优化

**智能采样**：
- 5xx错误始终记录
- 4xx错误前10次记录，之后每100次记录一次
- 定期清理采样计数器

**字段限制**：
```go
// 只在必要时添加元数据，避免过多日志
if len(err.Metadata) > 0 && len(err.Metadata) < 10 {
    // 记录元数据
}
```

### 3. 协议转换优化

**gRPC状态转换**：
```go
func (err *ErrorX) GRPCStatus() *status.Status {
    // 高效的元数据转换
    strMetadata := make(map[string]string, len(err.Metadata))
    for k, v := range err.Metadata {
        // 类型判断优化
        if str, ok := v.(string); ok {
            strMetadata[k] = str
        } else {
            strMetadata[k] = fmt.Sprintf("%v", v)
        }
    }
    // ...
}
```

## 最佳实践

### 1. 错误创建

```go
// ✅ 推荐：使用预定义错误
return errors.NewUserNotFoundError(userID)

// ✅ 推荐：包装数据库错误  
return errorsx.Wrap(err, 500, "DATABASE_ERROR", "Query failed")

// ❌ 不推荐：直接使用 errorsx（缺少业务语义）
return errorsx.New(404, "NOT_FOUND", "User not found")

// ❌ 不推荐：魔法字符串
return fmt.Errorf("user %s not found", userID)
```

### 2. 错误处理

```go
// ✅ 推荐：分层错误处理
func (h *Handler) GetUser(ctx context.Context, req *Request) (*Response, error) {
    return h.biz.GetUser(ctx, req)  // Handler层直接透传
}

func (b *Biz) GetUser(ctx context.Context, req *Request) (*Response, error) {
    // 业务层处理错误转换和上下文添加
    user, err := b.store.GetUser(ctx, req.ID)
    if err != nil {
        return nil, errors.NewUserNotFoundError(req.ID)
    }
    return user, nil
}
```

### 3. 错误信息

```go
// ✅ 推荐：用户友好的错误信息
"User account is temporarily locked due to multiple failed login attempts"

// ✅ 推荐：提供解决建议
"Invalid email format. Please enter a valid email address like user@example.com"

// ❌ 不推荐：暴露内部实现
"Database connection timeout on host 192.168.1.100:3306"

// ❌ 不推荐：技术细节过多
"gorm.ErrRecordNotFound: record not found in users table"
```

### 4. 元数据添加

```go
// ✅ 推荐：添加有用的调试信息
return errors.NewUserPermissionDeniedError(userID, "orders", "delete").
    AddMetadata("user_role", user.Role).
    AddMetadata("required_permission", "orders:delete")

// ✅ 推荐：添加业务上下文
return errors.NewRateLimitExceededError(100, "per_hour", 3600).
    AddMetadata("user_id", userID).
    AddMetadata("endpoint", "/api/v1/users")

// ❌ 不推荐：敏感信息
err.AddMetadata("password", user.Password)        // 泄露密码
err.AddMetadata("api_key", request.ApiKey)       // 泄露密钥
```

## 常见问题

### Q1: 什么时候使用 errorsx，什么时候使用 errors？

**A**: 
- **业务代码**: 始终使用 `pkg/errors` 中的预定义错误
- **框架代码**: 使用 `pkg/errorsx` 的基础功能
- **中间件**: 使用 `pkg/errorsx` 的中间件和处理器

### Q2: 如何添加新的业务错误？

**A**: 在 `pkg/errors` 中添加：

```go
// 1. 定义预定义错误
var ErrOrderNotFound = errorsx.New(404, "ORDER_NOT_FOUND", "Order not found").
    WithI18nKey("errors.order.not_found")

// 2. 创建构建器函数
func NewOrderNotFoundError(orderID string) *errorsx.ErrorX {
    return ErrOrderNotFound.AddMetadata("order_id", orderID)
}
```

### Q3: 如何处理错误链？

**A**: 使用 `WithCause()` 保持错误链：

```go
user, err := b.store.GetUser(ctx, id)
if err != nil {
    return nil, errorsx.Wrap(err, 500, "DATABASE_ERROR", "Failed to query user")
    // 或者
    return nil, errors.ErrDatabaseError.WithCause(err)
}
```

### Q4: 错误性能开销大吗？

**A**: 
- **创建开销**: 极小，使用对象池优化
- **内存开销**: 很小，写时复制设计
- **日志开销**: 智能采样，避免洪流

### Q5: 如何做错误测试？

**A**: 

```go
func TestUserNotFound(t *testing.T) {
    err := errors.NewUserNotFoundError("user123")
    
    // 测试错误码
    assert.Equal(t, int32(404), errorsx.Code(err))
    
    // 测试业务错误码
    assert.Equal(t, "USER_NOT_FOUND", errorsx.Reason(err))
    
    // 测试元数据
    assert.Equal(t, "user123", err.Metadata["user_id"])
    
    // 测试错误链判断
    assert.True(t, errors.Is(err, errors.ErrUserNotFound))
}
```

## 迁移指南

### 从字符串错误迁移

**旧代码**：
```go
return fmt.Errorf("user %s not found", userID)
```

**新代码**：
```go
return errors.NewUserNotFoundError(userID)
```

### 从简单错误迁移

**旧代码**：
```go
var ErrUserNotFound = errors.New("user not found")
```

**新代码**：
```go
var ErrUserNotFound = errorsx.New(404, "USER_NOT_FOUND", "User not found").
    WithI18nKey("errors.user.not_found")
```

### 批量迁移策略

1. **渐进式迁移**: 逐个模块替换，保持向后兼容
2. **保留旧接口**: 使用 `Deprecated` 标记
3. **添加新接口**: 提供基于 errorsx 的新实现
4. **文档更新**: 更新使用示例和最佳实践
5. **测试覆盖**: 确保新旧接口都有完整测试

---

## 总结

项目的错误处理架构具有以下特点：

1. **架构清晰**: 两层设计，框架与业务分离
2. **功能完整**: 支持国际化、性能优化、协议转换
3. **易于使用**: 链式API，预定义错误，构建器函数
4. **性能优化**: 对象池、字符串缓存、智能采样
5. **扩展性强**: 支持自定义处理器、中间件、监控

通过遵循本文档的指导和最佳实践，可以构建健壮、高性能、易维护的错误处理系统。