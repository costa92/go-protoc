# 错误处理模块重新设计方案

## 概述

本文档描述了对当前错误处理模块的重新设计方案，旨在解决现有架构中的问题，构建统一、可维护、国际化的错误处理机制。

**状态**: ✅ 已完成实现
**版本**: v2.0
**最后更新**: 2024年

## 现状分析

### 当前架构问题

1. **模块职责重叠**
   - `pkg/errorsx` 和 `pkg/api/errno` 功能重复
   - `pkg/errors` 定义业务错误但缺乏统一标准
   - 错误处理逻辑分散在多个包中

2. **调用流程混乱**
   - 缺乏统一的错误创建入口
   - 错误转换逻辑不一致
   - 国际化集成不够深入

3. **维护成本高**
   - 错误定义分散，难以管理
   - 缺乏统一的错误码规范
   - 测试覆盖不完整

## 设计目标

1. **统一性**: 提供统一的错误处理接口和格式
2. **可维护性**: 清晰的模块职责和简洁的调用流程
3. **国际化**: 深度集成国际化支持
4. **可扩展性**: 支持业务错误的灵活扩展
5. **性能**: 优化错误处理的性能开销

## 重新设计方案

### 1. 核心架构

```
pkg/errorsx/          # 核心错误处理模块 ✅
├── errorsx.go        # 统一错误接口 ✅
├── builder.go        # 错误构建器 ✅
├── code.go           # 标准错误码 ✅
├── i18n.go          # 国际化集成 ✅
├── middleware.go     # 中间件支持 ✅
├── registry.go       # 错误注册器 ✅
├── examples.go       # 使用示例 ✅
├── errorsx_test.go   # 核心测试 ✅
├── builder_test.go   # 构建器测试 ✅
├── registry_test.go  # 注册器测试 ✅
└── middleware_test.go # 中间件测试 ✅

pkg/errors/           # 业务错误定义 ✅
├── errors.go         # 兼容性错误定义 ✅
├── user.go           # 用户相关错误 ✅
├── auth.go           # 认证相关错误 ✅
├── common.go         # 通用业务错误 ✅
├── user_test.go      # 用户错误测试 ✅
├── auth_test.go      # 认证错误测试 ✅
└── common_test.go    # 通用错误测试 ✅
```

### 2. 核心组件设计

#### 2.1 统一错误接口 (pkg/errorsx/errorsx.go)

```go
// ErrorX 统一错误接口
type ErrorX struct {
    Code     int32             `json:"code"`
    Reason   string            `json:"reason"`
    Message  string            `json:"message"`
    Metadata map[string]any    `json:"metadata,omitempty"`
    
    // 内部字段
    i18nKey  string            // 国际化键
    cause    error             // 原始错误
}

// 核心方法
func (e *ErrorX) Error() string
func (e *ErrorX) Unwrap() error
func (e *ErrorX) Is(target error) bool
func (e *ErrorX) As(target any) bool
func (e *ErrorX) WithMetadata(key string, value any) *ErrorX
func (e *ErrorX) WithCause(err error) *ErrorX
func (e *ErrorX) Localize(ctx context.Context) *ErrorX
```

#### 2.2 错误构建器 (pkg/errorsx/builder.go)

```go
// Builder 错误构建器
type Builder struct {
    code     int32
    reason   string
    i18nKey  string
    metadata map[string]any
    cause    error
}

// 构建器方法
func NewBuilder(code int32, reason string) *Builder
func (b *Builder) WithI18nKey(key string) *Builder
func (b *Builder) WithMessage(message string) *Builder
func (b *Builder) WithMetadata(key string, value any) *Builder
func (b *Builder) WithCause(err error) *Builder
func (b *Builder) Build() *ErrorX
func (b *Builder) BuildWithContext(ctx context.Context) *ErrorX

// 便捷方法
func BadRequest(reason string) *Builder
func Unauthorized(reason string) *Builder
func Forbidden(reason string) *Builder
func NotFound(reason string) *Builder
func Conflict(reason string) *Builder
func InternalError(reason string) *Builder
```

#### 2.3 国际化集成 (pkg/errorsx/i18n.go)

```go
// I18nError 国际化错误
type I18nError struct {
    *ErrorX
    i18n *i18n.I18n
}

// 国际化方法
func (e *I18nError) LocalizeMessage(ctx context.Context) string
func (e *I18nError) LocalizeWithParams(ctx context.Context, params map[string]any) string

// 全局国际化函数
func SetGlobalI18n(i18n *i18n.I18n)
func LocalizeError(ctx context.Context, err *ErrorX) *ErrorX
```

#### 2.4 错误注册器 (pkg/errorsx/registry.go)

```go
// Registry 错误注册器
type Registry struct {
    errors map[string]*ErrorTemplate
    mutex  sync.RWMutex
}

// ErrorTemplate 错误模板
type ErrorTemplate struct {
    Code    int32
    Reason  string
    I18nKey string
}

// 注册方法
func (r *Registry) Register(reason string, template *ErrorTemplate)
func (r *Registry) Get(reason string) (*ErrorTemplate, bool)
func (r *Registry) MustGet(reason string) *ErrorTemplate

// 全局注册器
var GlobalRegistry = NewRegistry()

func Register(reason string, code int32, i18nKey string)
func MustCreate(reason string) *Builder
```

### 3. 业务错误重构

#### 3.1 用户相关错误 (internal/apiserver/pkg/errors/user.go)

```go
package errors

import (
    "github.com/costa92/go-protoc/pkg/errorsx"
    "github.com/costa92/go-protoc/internal/apiserver/pkg/locales"
)

// 用户错误码
const (
    ReasonUserNotFound      = "UserNotFound"
    ReasonUserAlreadyExists = "UserAlreadyExists"
    ReasonUserDisabled      = "UserDisabled"
    ReasonUserPasswordWeak  = "UserPasswordWeak"
)

// 注册用户错误
func init() {
    errorsx.Register(ReasonUserNotFound, 404, locales.UserNotFound)
    errorsx.Register(ReasonUserAlreadyExists, 409, locales.UserAlreadyExists)
    errorsx.Register(ReasonUserDisabled, 403, locales.UserDisabled)
    errorsx.Register(ReasonUserPasswordWeak, 400, locales.UserPasswordWeak)
}

// 便捷构造函数
func UserNotFound() *errorsx.Builder {
    return errorsx.MustCreate(ReasonUserNotFound)
}

func UserAlreadyExists() *errorsx.Builder {
    return errorsx.MustCreate(ReasonUserAlreadyExists)
}

func UserDisabled() *errorsx.Builder {
    return errorsx.MustCreate(ReasonUserDisabled)
}

func UserPasswordWeak() *errorsx.Builder {
    return errorsx.MustCreate(ReasonUserPasswordWeak)
}
```

#### 3.2 认证相关错误 (internal/apiserver/pkg/errors/auth.go)

```go
package errors

import (
    "github.com/costa92/go-protoc/pkg/errorsx"
    "github.com/costa92/go-protoc/internal/apiserver/pkg/locales"
)

// 认证错误码
const (
    ReasonTokenExpired    = "TokenExpired"
    ReasonTokenInvalid    = "TokenInvalid"
    ReasonLoginRequired   = "LoginRequired"
    ReasonPermissionDenied = "PermissionDenied"
)

// 注册认证错误
func init() {
    errorsx.Register(ReasonTokenExpired, 401, locales.TokenExpired)
    errorsx.Register(ReasonTokenInvalid, 401, locales.TokenInvalid)
    errorsx.Register(ReasonLoginRequired, 401, locales.LoginRequired)
    errorsx.Register(ReasonPermissionDenied, 403, locales.PermissionDenied)
}

// 便捷构造函数
func TokenExpired() *errorsx.Builder {
    return errorsx.MustCreate(ReasonTokenExpired)
}

func TokenInvalid() *errorsx.Builder {
    return errorsx.MustCreate(ReasonTokenInvalid)
}

func LoginRequired() *errorsx.Builder {
    return errorsx.MustCreate(ReasonLoginRequired)
}

func PermissionDenied() *errorsx.Builder {
    return errorsx.MustCreate(ReasonPermissionDenied)
}
```

### 4. 使用示例

#### 4.1 基本使用

```go
package handler

import (
    "context"
    "github.com/costa92/go-protoc/pkg/errorsx"
    "github.com/costa92/go-protoc/internal/apiserver/pkg/errors"
)

func (h *UserHandler) GetUser(ctx context.Context, id string) (*User, error) {
    user, err := h.userRepo.GetByID(ctx, id)
    if err != nil {
        if isNotFoundError(err) {
            // 返回业务错误，自动国际化
            return nil, errors.UserNotFound().
                WithMetadata("user_id", id).
                WithCause(err).
                BuildWithContext(ctx)
        }
        // 返回系统错误
        return nil, errorsx.InternalError("DatabaseError").
            WithCause(err).
            BuildWithContext(ctx)
    }
    return user, nil
}
```

#### 4.2 参数验证错误

```go
func (h *UserHandler) CreateUser(ctx context.Context, req *CreateUserRequest) error {
    // 参数验证
    if req.Username == "" {
        return errorsx.BadRequest("InvalidParameter").
            WithI18nKey(locales.UsernameRequired).
            WithMetadata("field", "username").
            BuildWithContext(ctx)
    }
    
    if len(req.Username) < 3 {
        return errorsx.BadRequest("InvalidParameter").
            WithI18nKey(locales.UsernameMinLength).
            WithMetadata("field", "username").
            WithMetadata("min_length", 3).
            WithMetadata("actual_length", len(req.Username)).
            BuildWithContext(ctx)
    }
    
    // 业务逻辑...
}
```

#### 4.3 错误处理中间件

```go
package middleware

import (
    "github.com/costa92/go-protoc/pkg/errorsx"
)

func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            
            // 转换为统一错误格式
            var errorX *errorsx.ErrorX
            if !errors.As(err, &errorX) {
                // 未知错误转换为内部错误
                errorX = errorsx.InternalError("UnknownError").
                    WithCause(err).
                    BuildWithContext(c.Request.Context())
            }
            
            // 本地化错误消息
            localizedError := errorX.Localize(c.Request.Context())
            
            c.JSON(int(localizedError.Code), localizedError)
        }
    }
}
```

## 迁移计划

### 阶段一：核心模块重构 (1-2周)

1. 重构 `pkg/errorsx` 模块
   - 实现新的 ErrorX 结构
   - 实现错误构建器
   - 集成国际化支持
   - 实现错误注册器

2. 更新 protobuf 错误定义
   - 简化 errno.proto
   - 重新生成错误码

### 阶段二：业务错误迁移 (1周)

1. 重构业务错误定义
   - 迁移用户相关错误
   - 迁移认证相关错误
   - 迁移其他业务错误

2. 更新错误处理中间件

### 阶段三：应用层适配 (1-2周)

1. 更新所有 handler
2. 更新所有 service
3. 更新错误处理逻辑

### 阶段四：测试和优化 (1周)

1. 完善单元测试
2. 集成测试
3. 性能优化
4. 文档更新

## 预期收益

1. **开发效率提升 30%**
   - 统一的错误处理接口
   - 自动化的国际化支持
   - 简化的错误创建流程

2. **维护成本降低 40%**
   - 集中的错误定义管理
   - 清晰的模块职责
   - 完善的测试覆盖

3. **用户体验改善**
   - 一致的错误响应格式
   - 准确的国际化消息
   - 丰富的错误上下文信息

4. **系统可观测性增强**
   - 统一的错误日志格式
   - 完整的错误链追踪
   - 便于监控和告警

## 风险评估

### 技术风险

1. **兼容性风险**: 中等
   - 现有 API 响应格式保持不变
   - 渐进式迁移降低风险

2. **性能风险**: 低
   - 错误对象复用机制
   - 延迟国际化策略

### 业务风险

1. **迁移风险**: 低
   - 分阶段迁移
   - 充分的测试覆盖

2. **学习成本**: 低
   - 简化的 API 设计
   - 完善的文档和示例

## 总结

本重新设计方案通过统一错误处理接口、优化模块架构、深度集成国际化支持，将显著提升错误处理的一致性、可维护性和用户体验。通过分阶段的迁移计划，可以在控制风险的同时，逐步实现架构升级。