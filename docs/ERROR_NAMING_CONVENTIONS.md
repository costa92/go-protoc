# 错误命名规范指南

本文档定义了项目中错误定义和构建器函数的命名规范，确保代码的一致性和可读性。

## 命名规范概览

### 1. 预定义错误变量

```go
// 格式：Err + 领域前缀 + 具体错误
var ErrUserNotFound = errorsx.New(404, "USER_NOT_FOUND", "User not found")
var ErrUserAlreadyExists = errorsx.New(409, "USER_ALREADY_EXISTS", "User already exists")
var ErrTokenExpired = errorsx.New(401, "TOKEN_EXPIRED", "Token has expired")
```

**规则**：
- 以 `Err` 开头
- 使用 PascalCase 命名
- 包含领域前缀（User、Token、Order等）
- 描述具体的错误情况

### 2. 构建器函数

```go
// 格式：New + 领域前缀 + 具体错误 + Error
func NewUserNotFoundError(userID string) *errorsx.ErrorX
func NewUserAlreadyExistsError(email string) *errorsx.ErrorX  
func NewTokenExpiredError(token string, expiredAt time.Time) *errorsx.ErrorX
```

**规则**：
- 以 `New` 开头
- 使用 PascalCase 命名
- 包含领域前缀
- 以 `Error` 结尾
- 参数名使用 camelCase

### 3. 业务错误码（Reason）

```go
// 格式：领域前缀_具体错误_描述
"USER_NOT_FOUND"           // 用户未找到
"USER_ALREADY_EXISTS"      // 用户已存在
"TOKEN_EXPIRED"            // Token过期
"VALIDATION_FAILED"        // 验证失败
"PERMISSION_DENIED"        // 权限拒绝
```

**规则**：
- 全大写字母
- 使用下划线分隔
- 包含领域前缀
- 简洁明确地描述错误

## 领域分类

### 用户相关 (User)

| 预定义错误 | 构建器函数 | 错误码 | 说明 |
|------------|------------|--------|------|
| `ErrUserNotFound` | `NewUserNotFoundError` | `USER_NOT_FOUND` | 用户不存在 |
| `ErrUserAlreadyExists` | `NewUserAlreadyExistsError` | `USER_ALREADY_EXISTS` | 用户已存在 |
| `ErrUserInvalidCredentials` | `NewUserInvalidCredentialsError` | `USER_INVALID_CREDENTIALS` | 凭据无效 |
| `ErrUserAccountLocked` | `NewUserAccountLockedError` | `USER_ACCOUNT_LOCKED` | 账户锁定 |
| `ErrUserPermissionDenied` | `NewUserPermissionDeniedError` | `USER_PERMISSION_DENIED` | 权限不足 |

### 认证相关 (Token/Auth)

| 预定义错误 | 构建器函数 | 错误码 | 说明 |
|------------|------------|--------|------|
| `ErrTokenInvalid` | `NewTokenInvalidError` | `TOKEN_INVALID` | Token无效 |
| `ErrTokenExpired` | `NewTokenExpiredError` | `TOKEN_EXPIRED` | Token过期 |
| `ErrTokenMissing` | `NewTokenMissingError` | `TOKEN_MISSING` | Token缺失 |
| `ErrAccountLocked` | `NewAccountLockedError` | `ACCOUNT_LOCKED` | 账户锁定 |
| `ErrLoginFailed` | `NewLoginFailedError` | `LOGIN_FAILED` | 登录失败 |

### 通用业务 (Resource/Parameter)

| 预定义错误 | 构建器函数 | 错误码 | 说明 |
|------------|------------|--------|------|
| `ErrResourceNotFound` | `NewResourceNotFoundError` | `RESOURCE_NOT_FOUND` | 资源不存在 |
| `ErrResourceConflict` | `NewResourceConflictError` | `RESOURCE_CONFLICT` | 资源冲突 |
| `ErrInvalidParameter` | `NewInvalidParameterError` | `INVALID_PARAMETER` | 参数无效 |
| `ErrMissingParameter` | `NewMissingParameterError` | `MISSING_PARAMETER` | 参数缺失 |
| `ErrRateLimitExceeded` | `NewRateLimitExceededError` | `RATE_LIMIT_EXCEEDED` | 限流超限 |

## 命名规范细则

### 1. 错误变量命名

**格式**: `Err + [Domain] + [Specific]`

```go
// ✅ 正确示例
var ErrUserNotFound = errorsx.New(...)      // 用户未找到
var ErrOrderCancelled = errorsx.New(...)    // 订单已取消
var ErrPaymentFailed = errorsx.New(...)     // 支付失败
var ErrFileTooBig = errorsx.New(...)        // 文件过大

// ❌ 错误示例
var UserNotFoundError = errorsx.New(...)    // 缺少 Err 前缀
var ErrNotFound = errorsx.New(...)          // 缺少领域前缀
var ERROR_USER_NOT_FOUND = errorsx.New(...) // 使用下划线
```

### 2. 构建器函数命名

**格式**: `New + [Domain] + [Specific] + Error`

```go
// ✅ 正确示例
func NewUserNotFoundError(userID string) *errorsx.ErrorX { ... }
func NewOrderCancelledError(orderID, reason string) *errorsx.ErrorX { ... }
func NewPaymentFailedError(paymentID string, cause error) *errorsx.ErrorX { ... }

// ❌ 错误示例
func UserNotFoundError(userID string) *errorsx.ErrorX { ... }     // 缺少 New 前缀
func NewNotFoundError(userID string) *errorsx.ErrorX { ... }      // 缺少领域前缀
func NewUserNotFound(userID string) *errorsx.ErrorX { ... }       // 缺少 Error 后缀
func CreateUserNotFoundError(userID string) *errorsx.ErrorX { ... } // 使用 Create 而非 New
```

### 3. 错误码命名

**格式**: `[DOMAIN]_[SPECIFIC]_[DESCRIPTION]`

```go
// ✅ 正确示例
"USER_NOT_FOUND"                 // 用户未找到
"ORDER_PAYMENT_FAILED"           // 订单支付失败
"FILE_SIZE_EXCEEDED"             // 文件大小超限
"RATE_LIMIT_EXCEEDED"            // 请求频率超限

// ❌ 错误示例
"NotFound"                       // 缺少领域前缀，格式不对
"user_not_found"                 // 应该全大写
"USER-NOT-FOUND"                 // 使用连字符而非下划线
"USERNOTFOUND"                   // 缺少分隔符
```

### 4. 参数命名

构建器函数参数使用 camelCase：

```go
// ✅ 正确示例
func NewUserNotFoundError(userID string) *errorsx.ErrorX { ... }
func NewPaymentFailedError(paymentID string, failureReason string) *errorsx.ErrorX { ... }
func NewRateLimitExceededError(requestLimit int, timeWindow string) *errorsx.ErrorX { ... }

// ❌ 错误示例
func NewUserNotFoundError(UserID string) *errorsx.ErrorX { ... }     // PascalCase
func NewUserNotFoundError(user_id string) *errorsx.ErrorX { ... }    // snake_case
func NewUserNotFoundError(userid string) *errorsx.ErrorX { ... }     // 无分隔符
```

## 特殊情况处理

### 1. 复合领域错误

当错误涉及多个领域时，使用主要领域作为前缀：

```go
// 用户订单相关 - 以 User 为主
var ErrUserOrderNotFound = errorsx.New(404, "USER_ORDER_NOT_FOUND", "User order not found")
func NewUserOrderNotFoundError(userID, orderID string) *errorsx.ErrorX { ... }

// 订单支付相关 - 以 Order 为主  
var ErrOrderPaymentFailed = errorsx.New(402, "ORDER_PAYMENT_FAILED", "Order payment failed")
func NewOrderPaymentFailedError(orderID, reason string) *errorsx.ErrorX { ... }
```

### 2. 通用错误的特化

从通用错误创建特定错误时：

```go
// 通用错误
var ErrValidationFailed = errorsx.New(422, "VALIDATION_FAILED", "Validation failed")

// 特定验证错误
var ErrEmailValidationFailed = errorsx.New(422, "EMAIL_VALIDATION_FAILED", "Email validation failed")
var ErrPasswordValidationFailed = errorsx.New(422, "PASSWORD_VALIDATION_FAILED", "Password validation failed")
```

### 3. HTTP状态码对应

不同HTTP状态码的错误命名：

```go
// 4xx 客户端错误
var ErrInvalidRequest = errorsx.New(400, "INVALID_REQUEST", "Invalid request")
var ErrUnauthorized = errorsx.New(401, "UNAUTHORIZED", "Unauthorized access")
var ErrForbidden = errorsx.New(403, "FORBIDDEN", "Access forbidden")
var ErrNotFound = errorsx.New(404, "NOT_FOUND", "Resource not found")
var ErrConflict = errorsx.New(409, "CONFLICT", "Resource conflict")
var ErrValidationFailed = errorsx.New(422, "VALIDATION_FAILED", "Validation failed")
var ErrRateLimited = errorsx.New(429, "RATE_LIMITED", "Rate limit exceeded")

// 5xx 服务器错误
var ErrInternalServer = errorsx.New(500, "INTERNAL_SERVER_ERROR", "Internal server error")
var ErrBadGateway = errorsx.New(502, "BAD_GATEWAY", "Bad gateway")
var ErrServiceUnavailable = errorsx.New(503, "SERVICE_UNAVAILABLE", "Service unavailable")
```

## 文件组织规范

### 目录结构

```
pkg/errors/
├── doc.go           # 包文档
├── common.go        # 通用业务错误
├── user.go          # 用户相关错误  
├── auth.go          # 认证相关错误
├── order.go         # 订单相关错误（示例）
├── payment.go       # 支付相关错误（示例）
└── errors.go        # 特定领域错误和向后兼容
```

### 文件内组织

每个错误文件的结构：

```go
package errors

import (
    "github.com/costa92/go-protoc/v2/pkg/errorsx"
)

// 1. 预定义错误变量（按字母顺序）
var (
    ErrUserAccountLocked     = errorsx.New(423, "USER_ACCOUNT_LOCKED", "User account is locked")
    ErrUserAlreadyExists     = errorsx.New(409, "USER_ALREADY_EXISTS", "User already exists") 
    ErrUserNotFound          = errorsx.New(404, "USER_NOT_FOUND", "User not found")
    ErrUserPermissionDenied  = errorsx.New(403, "USER_PERMISSION_DENIED", "Permission denied")
)

// 2. 构建器函数（按字母顺序）
func NewUserAccountLockedError(userID string, lockUntil string) *errorsx.ErrorX { ... }
func NewUserAlreadyExistsError(email string) *errorsx.ErrorX { ... }
func NewUserNotFoundError(userID string) *errorsx.ErrorX { ... }
func NewUserPermissionDeniedError(userID, resource, action string) *errorsx.ErrorX { ... }
```

## 检查清单

在添加新错误时，请确保：

- [ ] 错误变量名符合 `Err + Domain + Specific` 格式
- [ ] 构建器函数名符合 `New + Domain + Specific + Error` 格式  
- [ ] 错误码使用全大写和下划线分隔
- [ ] 参数名使用 camelCase
- [ ] 包含适当的 i18n 键
- [ ] 错误消息用户友好
- [ ] 添加了适当的元数据支持
- [ ] HTTP状态码正确
- [ ] 有对应的单元测试
- [ ] 更新了相关文档

## 示例模板

### 新增错误模板

```go
// 在 pkg/errors/domain.go 中

// 预定义错误
var ErrDomainSpecificError = errorsx.New(
    400,                                    // HTTP状态码
    "DOMAIN_SPECIFIC_ERROR",               // 业务错误码  
    "Domain specific error occurred",       // 用户友好消息
).WithI18nKey("errors.domain.specific")    // 国际化键

// 构建器函数
func NewDomainSpecificError(param1, param2 string) *errorsx.ErrorX {
    return ErrDomainSpecificError.
        WithMessage("Domain specific error for %s: %s", param1, param2).
        AddMetadata("param1", param1).
        AddMetadata("param2", param2)
}
```

### 测试模板

```go
// 在 pkg/errors/domain_test.go 中

func TestNewDomainSpecificError(t *testing.T) {
    err := NewDomainSpecificError("value1", "value2")
    
    // 测试状态码
    assert.Equal(t, int32(400), errorsx.Code(err))
    
    // 测试错误码
    assert.Equal(t, "DOMAIN_SPECIFIC_ERROR", errorsx.Reason(err))
    
    // 测试元数据
    assert.Equal(t, "value1", err.Metadata["param1"])
    assert.Equal(t, "value2", err.Metadata["param2"])
    
    // 测试错误链判断
    assert.True(t, errors.Is(err, ErrDomainSpecificError))
}
```

遵循这些命名规范可以确保项目错误处理的一致性、可读性和可维护性。