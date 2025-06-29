# 错误处理系统使用指南

## 概述

本指南详细介绍如何在 go-protoc 项目中使用统一错误处理系统。通过本指南，您将学会如何创建、处理和管理各种类型的错误。

## 快速开始

### 1. 基本错误创建

```go
package main

import (
    "github.com/costa92/go-protoc/pkg/errorsx"
    "github.com/costa92/go-protoc/pkg/errors"
)

func main() {
    // 方式1: 使用预定义错误
    err1 := errors.ErrUserNotFound
    
    // 方式2: 使用构建器创建
    err2 := errorsx.NotFound().
        WithReason("USER_NOT_FOUND").
        WithMessage("User not found").
        Build()
    
    // 方式3: 使用便捷函数
    err3 := errors.NewUserNotFoundError("123")
    
    fmt.Println(err1) // User not found
    fmt.Println(err2) // User not found  
    fmt.Println(err3) // User not found
}
```

### 2. 错误处理

```go
func handleError(err error) {
    // 检查错误类型
    if errorsx.IsCode(err, 404) {
        fmt.Println("资源不存在")
    }
    
    // 检查错误原因
    if errorsx.IsReason(err, "USER_NOT_FOUND") {
        fmt.Println("用户不存在")
    }
    
    // 获取错误详情
    if errorX, ok := err.(*errorsx.ErrorX); ok {
        fmt.Printf("Code: %d, Reason: %s, Message: %s\n", 
            errorX.Code, errorX.Reason, errorX.Message)
    }
}
```

## 详细使用说明

### 1. 错误创建方式

#### 1.1 使用构建器模式

```go
// 基本构建
err := errorsx.BadRequest().
    WithReason("INVALID_PARAMETER").
    WithMessage("Invalid email format").
    Build()

// 添加元数据
err := errorsx.BadRequest().
    WithReason("VALIDATION_FAILED").
    WithMessage("Validation failed").
    AddMetadata("field", "email").
    AddMetadata("value", "invalid-email").
    AddMetadata("rule", "email_format").
    Build()

// 包装原始错误
originalErr := sql.ErrNoRows
err := errorsx.NotFound().
    WithReason("USER_NOT_FOUND").
    WithMessage("User not found").
    WithCause(originalErr).
    Build()

// 国际化支持
err := errorsx.BadRequest().
    WithReason("INVALID_PARAMETER").
    WithI18nKey("errors.validation.invalid_email").
    AddMetadata("field", "email").
    Build()
```

#### 1.2 使用预定义错误

```go
// 用户相关错误
err1 := errors.ErrUserNotFound
err2 := errors.ErrUserAlreadyExists
err3 := errors.ErrUserDisabled

// 认证相关错误
err4 := errors.ErrTokenInvalid
err5 := errors.ErrTokenExpired
err6 := errors.ErrInsufficientPermissions

// 通用错误
err7 := errors.ErrResourceNotFound
err8 := errors.ErrInvalidRequest
err9 := errors.ErrRateLimitExceeded
```

#### 1.3 使用便捷构建函数

```go
// 用户错误
err1 := errors.NewUserNotFoundError("user123")
err2 := errors.NewUserValidationError("email", "invalid format")
err3 := errors.NewUserPermissionError("delete_user")

// 认证错误
err4 := errors.NewTokenInvalidError("expired")
err5 := errors.NewLoginFailedError("invalid_credentials")

// 通用错误
err6 := errors.NewResourceNotFoundError("user", "123")
err7 := errors.NewInvalidParameterError("email", "invalid format")
```

### 2. 错误检查和处理

#### 2.1 类型检查

```go
func processError(err error) {
    // 检查是否为 ErrorX 类型
    if errorX, ok := err.(*errorsx.ErrorX); ok {
        fmt.Printf("ErrorX: %+v\n", errorX)
    }
    
    // 使用 errors.As 进行类型转换
    var errorX *errorsx.ErrorX
    if errors.As(err, &errorX) {
        fmt.Printf("Code: %d\n", errorX.Code)
    }
}
```

#### 2.2 错误码检查

```go
func handleByCode(err error) {
    switch {
    case errorsx.IsCode(err, 400):
        fmt.Println("客户端错误")
    case errorsx.IsCode(err, 401):
        fmt.Println("认证失败")
    case errorsx.IsCode(err, 403):
        fmt.Println("权限不足")
    case errorsx.IsCode(err, 404):
        fmt.Println("资源不存在")
    case errorsx.IsCode(err, 500):
        fmt.Println("服务器错误")
    default:
        fmt.Println("未知错误")
    }
}
```

#### 2.3 错误原因检查

```go
func handleByReason(err error) {
    switch {
    case errorsx.IsReason(err, "USER_NOT_FOUND"):
        fmt.Println("用户不存在")
    case errorsx.IsReason(err, "TOKEN_EXPIRED"):
        fmt.Println("令牌过期")
    case errorsx.IsReason(err, "INSUFFICIENT_PERMISSIONS"):
        fmt.Println("权限不足")
    case errorsx.IsReason(err, "RATE_LIMIT_EXCEEDED"):
        fmt.Println("请求过于频繁")
    }
}
```

#### 2.4 错误链处理

```go
func handleErrorChain(err error) {
    // 检查错误链
    if errors.Is(err, sql.ErrNoRows) {
        fmt.Println("数据库中没有找到记录")
    }
    
    // 获取原始错误
    if errorX, ok := err.(*errorsx.ErrorX); ok {
        if cause := errorX.Unwrap(); cause != nil {
            fmt.Printf("原始错误: %v\n", cause)
        }
    }
    
    // 遍历错误链
    for err != nil {
        fmt.Printf("错误: %v\n", err)
        if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
            err = unwrapper.Unwrap()
        } else {
            break
        }
    }
}
```

### 3. 元数据使用

#### 3.1 添加元数据

```go
// 单个添加
err := errorsx.BadRequest().
    WithReason("VALIDATION_FAILED").
    AddMetadata("field", "email").
    AddMetadata("value", "invalid@").
    AddMetadata("rule", "email_format").
    Build()

// 批量添加
metadata := map[string]any{
    "field": "email",
    "value": "invalid@",
    "rule":  "email_format",
    "line":  42,
}
err := errorsx.BadRequest().
    WithReason("VALIDATION_FAILED").
    WithMetadata(metadata).
    Build()
```

#### 3.2 获取元数据

```go
func extractMetadata(err error) {
    if errorX, ok := err.(*errorsx.ErrorX); ok {
        // 获取所有元数据
        metadata := errorX.Metadata
        fmt.Printf("元数据: %+v\n", metadata)
        
        // 获取特定字段
        if field, exists := metadata["field"]; exists {
            fmt.Printf("字段: %v\n", field)
        }
        
        // 类型安全的获取
        if fieldStr, ok := metadata["field"].(string); ok {
            fmt.Printf("字段名: %s\n", fieldStr)
        }
    }
}
```

### 4. 国际化集成

#### 4.1 设置国际化键

```go
// 创建带国际化键的错误
err := errorsx.BadRequest().
    WithReason("VALIDATION_FAILED").
    WithI18nKey("errors.validation.required_field").
    AddMetadata("field", "username").
    Build()

// 使用参数化消息
err := errorsx.BadRequest().
    WithReason("VALIDATION_FAILED").
    WithI18nKey("errors.validation.min_length").
    AddMetadata("field", "password").
    AddMetadata("min_length", 8).
    Build()
```

#### 4.2 本地化错误消息

```go
func localizeError(ctx context.Context, err error) error {
    if errorX, ok := err.(*errorsx.ErrorX); ok {
        // 自动本地化
        return errorX.Localize(ctx)
    }
    return err
}

// 使用示例
func handleRequest(ctx context.Context) {
    err := someBusinessLogic()
    if err != nil {
        // 本地化错误消息
        localizedErr := localizeError(ctx, err)
        
        // 返回本地化的错误
        respondWithError(ctx, localizedErr)
    }
}
```

### 5. 错误注册器使用

#### 5.1 注册错误模板

```go
func init() {
    // 注册用户相关错误模板
    errorsx.RegisterGlobal("USER_NOT_FOUND", &errorsx.ErrorTemplate{
        Code:    404,
        Reason:  "USER_NOT_FOUND",
        I18nKey: "errors.user.not_found",
    })
    
    errorsx.RegisterGlobal("USER_VALIDATION_FAILED", &errorsx.ErrorTemplate{
        Code:    400,
        Reason:  "USER_VALIDATION_FAILED",
        I18nKey: "errors.user.validation_failed",
    })
}
```

#### 5.2 使用注册的模板

```go
func createErrorFromTemplate() {
    // 从模板创建错误
    err, _ := errorsx.CreateGlobal("USER_NOT_FOUND", map[string]any{
        "user_id": "123",
    })
    
    // 必须成功创建（panic if template not found）
    err2 := errorsx.MustCreateGlobal("USER_VALIDATION_FAILED", map[string]any{
        "field": "email",
        "value": "invalid@",
    })
    
    fmt.Println(err)  // User not found
    fmt.Println(err2) // User validation failed
}
```

### 6. 中间件集成

#### 6.1 Gin 中间件

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/costa92/go-protoc/pkg/errorsx"
)

func main() {
    r := gin.Default()
    
    // 使用默认错误处理器
    r.Use(errorsx.GinErrorMiddleware(errorsx.NewDefaultErrorHandler()))
    
    // 或使用自定义错误处理器
    customHandler := &CustomErrorHandler{}
    r.Use(errorsx.GinErrorMiddleware(customHandler))
    
    r.GET("/users/:id", getUserHandler)
    
    r.Run(":8080")
}

func getUserHandler(c *gin.Context) {
    userID := c.Param("id")
    
    user, err := getUserByID(userID)
    if err != nil {
        // 错误会被中间件自动处理
        c.Error(err)
        return
    }
    
    c.JSON(200, user)
}

func getUserByID(id string) (*User, error) {
    // 模拟业务逻辑
    if id == "" {
        return nil, errorsx.BadRequest().
            WithReason("INVALID_PARAMETER").
            WithMessage("User ID is required").
            AddMetadata("parameter", "id").
            Build()
    }
    
    if id == "999" {
        return nil, errors.NewUserNotFoundError(id)
    }
    
    return &User{ID: id, Name: "John"}, nil
}
```

#### 6.2 标准 HTTP 中间件

```go
package main

import (
    "net/http"
    "github.com/costa92/go-protoc/pkg/errorsx"
)

func main() {
    mux := http.NewServeMux()
    
    // 包装处理器
    handler := errorsx.HTTPErrorMiddleware(
        errorsx.NewDefaultErrorHandler(),
    )(getUserHandler)
    
    mux.Handle("/users/", handler)
    
    http.ListenAndServe(":8080", mux)
}

func getUserHandler(w http.ResponseWriter, r *http.Request) error {
    userID := r.URL.Path[len("/users/"):]
    
    user, err := getUserByID(userID)
    if err != nil {
        return err // 错误会被中间件处理
    }
    
    // 正常响应
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
    return nil
}
```

#### 6.3 Panic 恢复中间件

```go
func main() {
    r := gin.Default()
    
    // 添加 panic 恢复中间件
    r.Use(errorsx.RecoverMiddleware())
    
    // 添加错误处理中间件
    r.Use(errorsx.GinErrorMiddleware(errorsx.NewDefaultErrorHandler()))
    
    r.GET("/panic", func(c *gin.Context) {
        panic("something went wrong") // 会被恢复并转换为 500 错误
    })
    
    r.Run(":8080")
}
```

### 7. gRPC 集成

#### 7.1 gRPC 错误转换

```go
package main

import (
    "context"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "github.com/costa92/go-protoc/pkg/errorsx"
)

// gRPC 服务实现
type UserService struct {
    // ...
}

func (s *UserService) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
    user, err := s.getUserByID(req.Id)
    if err != nil {
        // ErrorX 会自动转换为 gRPC 状态
        if errorX, ok := err.(*errorsx.ErrorX); ok {
            return nil, errorX.GRPCStatus().Err()
        }
        return nil, status.Error(codes.Internal, "Internal error")
    }
    
    return &GetUserResponse{User: user}, nil
}

// 从 gRPC 错误转换回 ErrorX
func handleGRPCError(err error) {
    if st, ok := status.FromError(err); ok {
        errorX := errorsx.FromGRPCStatus(st)
        fmt.Printf("Code: %d, Reason: %s\n", errorX.Code, errorX.Reason)
    }
}
```

#### 7.2 gRPC 拦截器

```go
func errorInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        resp, err := handler(ctx, req)
        
        if err != nil {
            // 记录错误日志
            if errorX, ok := err.(*errorsx.ErrorX); ok {
                log.Printf("gRPC error: %s - %s", errorX.Reason, errorX.Message)
                return resp, errorX.GRPCStatus().Err()
            }
        }
        
        return resp, err
    }
}

func main() {
    s := grpc.NewServer(
        grpc.UnaryInterceptor(errorInterceptor()),
    )
    
    // 注册服务...
    
    lis, _ := net.Listen("tcp", ":9090")
    s.Serve(lis)
}
```

### 8. 业务场景示例

#### 8.1 用户管理服务

```go
package user

import (
    "context"
    "database/sql"
    "github.com/costa92/go-protoc/pkg/errorsx"
    "github.com/costa92/go-protoc/pkg/errors"
)

type Service struct {
    repo *Repository
}

func (s *Service) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // 参数验证
    if err := s.validateCreateUserRequest(req); err != nil {
        return nil, err
    }
    
    // 检查用户是否已存在
    existing, err := s.repo.GetByEmail(ctx, req.Email)
    if err != nil && !errors.Is(err, sql.ErrNoRows) {
        return nil, errorsx.InternalError().
            WithReason("DATABASE_ERROR").
            WithMessage("Failed to check user existence").
            WithCause(err).
            Build()
    }
    
    if existing != nil {
        return nil, errors.NewUserAlreadyExistsError(req.Email)
    }
    
    // 创建用户
    user, err := s.repo.Create(ctx, req)
    if err != nil {
        return nil, errorsx.InternalError().
            WithReason("USER_CREATION_FAILED").
            WithMessage("Failed to create user").
            WithCause(err).
            AddMetadata("email", req.Email).
            Build()
    }
    
    return user, nil
}

func (s *Service) validateCreateUserRequest(req *CreateUserRequest) error {
    var validationErrors []string
    
    if req.Email == "" {
        validationErrors = append(validationErrors, "email is required")
    } else if !isValidEmail(req.Email) {
        return errors.NewUserValidationError("email", "invalid format")
    }
    
    if req.Password == "" {
        validationErrors = append(validationErrors, "password is required")
    } else if len(req.Password) < 8 {
        return errorsx.BadRequest().
            WithReason("PASSWORD_TOO_SHORT").
            WithI18nKey("errors.user.password_too_short").
            AddMetadata("min_length", 8).
            AddMetadata("actual_length", len(req.Password)).
            Build()
    }
    
    if len(validationErrors) > 0 {
        return errorsx.BadRequest().
            WithReason("VALIDATION_FAILED").
            WithMessage("Validation failed").
            AddMetadata("errors", validationErrors).
            Build()
    }
    
    return nil
}

func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    if id == "" {
        return nil, errors.NewInvalidParameterError("id", "cannot be empty")
    }
    
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, errors.NewUserNotFoundError(id)
        }
        return nil, errorsx.InternalError().
            WithReason("DATABASE_ERROR").
            WithCause(err).
            Build()
    }
    
    return user, nil
}

func (s *Service) UpdateUser(ctx context.Context, id string, req *UpdateUserRequest) (*User, error) {
    // 检查用户是否存在
    existing, err := s.GetUser(ctx, id)
    if err != nil {
        return nil, err // 直接返回，保持错误链
    }
    
    // 检查权限
    if !s.hasPermission(ctx, "update_user", existing) {
        return nil, errors.NewUserPermissionError("update_user")
    }
    
    // 更新用户
    user, err := s.repo.Update(ctx, id, req)
    if err != nil {
        return nil, errorsx.InternalError().
            WithReason("USER_UPDATE_FAILED").
            WithCause(err).
            AddMetadata("user_id", id).
            Build()
    }
    
    return user, nil
}

func (s *Service) DeleteUser(ctx context.Context, id string) error {
    // 检查用户是否存在
    user, err := s.GetUser(ctx, id)
    if err != nil {
        return err
    }
    
    // 检查权限
    if !s.hasPermission(ctx, "delete_user", user) {
        return errors.NewUserPermissionError("delete_user")
    }
    
    // 软删除用户
    if err := s.repo.SoftDelete(ctx, id); err != nil {
        return errorsx.InternalError().
            WithReason("USER_DELETION_FAILED").
            WithCause(err).
            AddMetadata("user_id", id).
            Build()
    }
    
    return nil
}
```

#### 8.2 认证服务

```go
package auth

import (
    "context"
    "time"
    "github.com/costa92/go-protoc/pkg/errorsx"
    "github.com/costa92/go-protoc/pkg/errors"
)

type Service struct {
    userRepo  *UserRepository
    tokenRepo *TokenRepository
}

func (s *Service) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
    // 参数验证
    if email == "" || password == "" {
        return nil, errorsx.BadRequest().
            WithReason("MISSING_CREDENTIALS").
            WithI18nKey("errors.auth.missing_credentials").
            Build()
    }
    
    // 获取用户
    user, err := s.userRepo.GetByEmail(ctx, email)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, errors.NewLoginFailedError("invalid_credentials")
        }
        return nil, errorsx.InternalError().
            WithReason("DATABASE_ERROR").
            WithCause(err).
            Build()
    }
    
    // 检查用户状态
    if user.Status == "disabled" {
        return nil, errors.NewUserDisabledError(user.ID)
    }
    
    // 验证密码
    if !s.verifyPassword(password, user.PasswordHash) {
        // 记录失败尝试
        s.recordFailedLogin(ctx, email)
        
        return nil, errors.NewLoginFailedError("invalid_credentials")
    }
    
    // 检查账户锁定
    if s.isAccountLocked(ctx, user.ID) {
        return nil, errorsx.Forbidden().
            WithReason("ACCOUNT_LOCKED").
            WithI18nKey("errors.auth.account_locked").
            AddMetadata("user_id", user.ID).
            AddMetadata("locked_until", s.getLockExpiry(ctx, user.ID)).
            Build()
    }
    
    // 生成令牌
    token, err := s.generateToken(ctx, user)
    if err != nil {
        return nil, errorsx.InternalError().
            WithReason("TOKEN_GENERATION_FAILED").
            WithCause(err).
            Build()
    }
    
    return &LoginResponse{
        Token: token,
        User:  user,
    }, nil
}

func (s *Service) ValidateToken(ctx context.Context, tokenString string) (*User, error) {
    if tokenString == "" {
        return nil, errors.NewTokenMissingError()
    }
    
    // 解析令牌
    token, err := s.parseToken(tokenString)
    if err != nil {
        return nil, errors.NewTokenInvalidError("malformed")
    }
    
    // 检查过期
    if token.ExpiresAt.Before(time.Now()) {
        return nil, errors.NewTokenExpiredError(token.ExpiresAt)
    }
    
    // 检查是否被撤销
    if s.isTokenRevoked(ctx, token.ID) {
        return nil, errors.NewTokenInvalidError("revoked")
    }
    
    // 获取用户信息
    user, err := s.userRepo.GetByID(ctx, token.UserID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, errors.NewTokenInvalidError("user_not_found")
        }
        return nil, errorsx.InternalError().
            WithReason("DATABASE_ERROR").
            WithCause(err).
            Build()
    }
    
    return user, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
    // 验证刷新令牌
    token, err := s.validateRefreshToken(ctx, refreshToken)
    if err != nil {
        return nil, err
    }
    
    // 检查用户状态
    user, err := s.userRepo.GetByID(ctx, token.UserID)
    if err != nil {
        return nil, errorsx.InternalError().
            WithReason("DATABASE_ERROR").
            WithCause(err).
            Build()
    }
    
    if user.Status == "disabled" {
        return nil, errors.NewUserDisabledError(user.ID)
    }
    
    // 生成新令牌
    newToken, err := s.generateToken(ctx, user)
    if err != nil {
        return nil, errorsx.InternalError().
            WithReason("TOKEN_GENERATION_FAILED").
            WithCause(err).
            Build()
    }
    
    // 撤销旧令牌
    if err := s.revokeToken(ctx, refreshToken); err != nil {
        // 记录警告，但不影响新令牌生成
        log.Warnf("Failed to revoke old refresh token: %v", err)
    }
    
    return &TokenResponse{
        AccessToken:  newToken.AccessToken,
        RefreshToken: newToken.RefreshToken,
        ExpiresAt:    newToken.ExpiresAt,
    }, nil
}
```

### 9. 性能优化技巧

#### 9.1 错误对象复用

```go
// 预定义常用错误
var (
    errInvalidEmail = errorsx.BadRequest().
        WithReason("INVALID_EMAIL").
        WithI18nKey("errors.validation.invalid_email").
        Build()
        
    errPasswordTooShort = errorsx.BadRequest().
        WithReason("PASSWORD_TOO_SHORT").
        WithI18nKey("errors.validation.password_too_short").
        Build()
)

// 复用错误对象
func validateEmail(email string) error {
    if !isValidEmail(email) {
        return errInvalidEmail.WithMetadata(map[string]any{
            "email": email,
        })
    }
    return nil
}
```

#### 9.2 延迟国际化

```go
// 只在需要时进行国际化
func handleError(ctx context.Context, err error) {
    if errorX, ok := err.(*errorsx.ErrorX); ok {
        // 只有在需要返回给客户端时才进行国际化
        if needsLocalization(ctx) {
            err = errorX.Localize(ctx)
        }
    }
    
    respondWithError(ctx, err)
}
```

#### 9.3 批量错误处理

```go
// 批量验证
func validateUserBatch(users []*User) error {
    var errors []string
    
    for i, user := range users {
        if err := validateUser(user); err != nil {
            errors = append(errors, fmt.Sprintf("user[%d]: %v", i, err))
        }
    }
    
    if len(errors) > 0 {
        return errorsx.BadRequest().
            WithReason("BATCH_VALIDATION_FAILED").
            WithMessage("Multiple validation errors").
            AddMetadata("errors", errors).
            AddMetadata("total_errors", len(errors)).
            Build()
    }
    
    return nil
}
```

## 常见问题和解决方案

### 1. 错误消息不显示

**问题**: 创建的错误没有显示预期的消息

**解决方案**:
```go
// ❌ 错误做法 - 没有设置消息
err := errorsx.BadRequest().
    WithReason("INVALID_PARAMETER").
    Build()

// ✅ 正确做法 - 设置消息或国际化键
err := errorsx.BadRequest().
    WithReason("INVALID_PARAMETER").
    WithMessage("Invalid parameter").
    Build()

// 或者使用国际化
err := errorsx.BadRequest().
    WithReason("INVALID_PARAMETER").
    WithI18nKey("errors.validation.invalid_parameter").
    Build()
```

### 2. 错误链丢失

**问题**: 包装错误时丢失了原始错误信息

**解决方案**:
```go
// ❌ 错误做法 - 丢失原始错误
if err != nil {
    return errorsx.InternalError().
        WithMessage("Database error").
        Build()
}

// ✅ 正确做法 - 保持错误链
if err != nil {
    return errorsx.InternalError().
        WithMessage("Database error").
        WithCause(err).
        Build()
}
```

### 3. 元数据类型错误

**问题**: 元数据类型不匹配导致序列化失败

**解决方案**:
```go
// ❌ 错误做法 - 使用不可序列化的类型
err := errorsx.BadRequest().
    AddMetadata("func", func() {}).
    Build()

// ✅ 正确做法 - 使用可序列化的类型
err := errorsx.BadRequest().
    AddMetadata("field", "email").
    AddMetadata("value", "invalid@").
    AddMetadata("line", 42).
    Build()
```

### 4. 国际化不生效

**问题**: 设置了国际化键但消息没有本地化

**解决方案**:
```go
// 确保国际化系统已初始化
func init() {
    i18n.LoadLocales("locales")
}

// 在处理错误时进行本地化
func handleError(ctx context.Context, err error) {
    if errorX, ok := err.(*errorsx.ErrorX); ok {
        localizedErr := errorX.Localize(ctx)
        respondWithError(ctx, localizedErr)
    }
}
```

### 5. 性能问题

**问题**: 频繁创建错误对象导致性能下降

**解决方案**:
```go
// 使用错误对象复用
var errPool = sync.Pool{
    New: func() interface{} {
        return &errorsx.ErrorX{}
    },
}

func getError() *errorsx.ErrorX {
    return errPool.Get().(*errorsx.ErrorX)
}

func putError(err *errorsx.ErrorX) {
    err.Reset()
    errPool.Put(err)
}
```

## 总结

通过本使用指南，您应该能够：

1. 熟练使用各种错误创建方式
2. 正确处理和检查错误
3. 有效利用元数据和国际化功能
4. 集成中间件进行统一错误处理
5. 在实际业务场景中应用错误处理最佳实践
6. 优化错误处理性能
7. 解决常见问题

记住，良好的错误处理是构建健壮应用程序的关键。始终遵循一致性、可维护性和用户友好性的原则。