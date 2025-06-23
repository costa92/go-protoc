# 错误处理最佳实践

## 概述

本文档提供了在项目中实施错误处理的最佳实践，帮助开发团队构建健壮、可维护的错误处理机制。

## 设计原则

### 1. 一致性原则

**所有 API 错误响应必须使用统一格式**

```json
{
  "code": 400,
  "reason": "InvalidParameter",
  "message": "参数验证失败",
  "metadata": {
    "field": "username",
    "constraint": "min_length_3"
  }
}
```

✅ **正确做法**:
```go
// 使用统一的错误创建方式
return errno.ErrorInvalidParameter("用户名长度不能少于3个字符")
```

❌ **错误做法**:
```go
// 直接返回不同格式的错误
return errors.New("用户名太短")
return fmt.Errorf("validation failed: %s", field)
```

### 2. 分层原则

**区分系统错误和业务错误**

```go
// 系统级错误 - 使用 errno 包
errno.ErrorInternalError("数据库连接失败")
errno.ErrorInvalidParameter("请求格式错误")

// 业务级错误 - 使用具体业务包
v1.ErrorUserAlreadyExists("用户已存在")
v1.ErrorSecretNotFound("密钥未找到")
```

### 3. 国际化原则

**所有面向用户的错误消息都应支持国际化**

✅ **正确做法**:
```go
message := i18n.FromContext(ctx).T(locales.UserAlreadyExists)
return v1.ErrorUserAlreadyExists(message)
```

❌ **错误做法**:
```go
// 硬编码中文消息
return v1.ErrorUserAlreadyExists("用户已存在")
```

## 错误定义最佳实践

### 1. 错误命名规范

**使用清晰、具体的错误名称**

✅ **推荐命名**:
- `UserAlreadyExists` - 明确表示用户已存在
- `SecretReachMaxCount` - 明确表示密钥数量达到上限
- `UserOperationForbidden` - 明确表示用户操作被禁止

❌ **不推荐命名**:
- `UserError` - 过于宽泛
- `Failed` - 没有具体信息
- `Error1`, `Error2` - 无意义的命名

### 2. HTTP 状态码选择

**选择合适的 HTTP 状态码**

| 状态码 | 使用场景 | 示例 |
|--------|----------|------|
| 400 | 客户端请求错误 | 参数格式错误、必填参数缺失 |
| 401 | 认证失败 | 未登录、token 过期 |
| 403 | 权限不足 | 无权限访问资源 |
| 404 | 资源不存在 | 用户不存在、页面不存在 |
| 409 | 资源冲突 | 用户名已存在、重复创建 |
| 422 | 语义错误 | 业务规则验证失败 |
| 429 | 请求过多 | 超出速率限制 |
| 500 | 服务器错误 | 数据库错误、第三方服务异常 |

### 3. 错误消息设计

**编写有用的错误消息**

✅ **好的错误消息**:
```go
// 具体、可操作
"用户名长度必须在 3-20 个字符之间"
"邮箱格式不正确，请检查后重试"
"密钥数量已达上限（10个），请删除不需要的密钥后重试"
```

❌ **不好的错误消息**:
```go
// 模糊、无用
"操作失败"
"系统错误"
"请联系管理员"
```

## 代码实现最佳实践

### 1. 错误处理层次

**在不同层次正确处理错误**

```go
// 1. 数据访问层 - 转换底层错误
func (r *UserRepo) GetUser(ctx context.Context, id string) (*User, error) {
    var user User
    err := r.db.Where("id = ?", id).First(&user).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            // 转换为业务错误
            message := i18n.FromContext(ctx).T(locales.UserNotFound)
            return nil, v1.ErrorUserNotFound(message)
        }
        // 系统错误保持原样或包装
        return nil, fmt.Errorf("database query failed: %w", err)
    }
    return &user, nil
}

// 2. 业务逻辑层 - 处理业务规则
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // 业务规则验证
    if len(req.Username) < 3 {
        return nil, errno.ErrorInvalidParameter("用户名长度不能少于3个字符")
    }
    
    // 检查用户是否存在
    existing, err := s.repo.GetUserByName(ctx, req.Username)
    if err != nil && !v1.IsUserNotFound(err) {
        // 非"用户不存在"的其他错误，直接返回
        return nil, err
    }
    if existing != nil {
        message := i18n.FromContext(ctx).T(locales.UserAlreadyExists)
        return nil, v1.ErrorUserAlreadyExists(message)
    }
    
    // 创建用户
    return s.repo.CreateUser(ctx, req)
}

// 3. 接口层 - 最终错误处理
func (h *UserHandler) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
    user, err := h.service.CreateUser(ctx, &CreateUserRequest{
        Username: req.Username,
        Email:    req.Email,
    })
    if err != nil {
        // 记录错误日志（包含详细信息）
        h.logger.WithFields(logrus.Fields{
            "username": req.Username,
            "error":    err.Error(),
            "trace_id": contextx.TraceID(ctx),
        }).Error("Failed to create user")
        
        // 返回错误（可能经过包装）
        return nil, err
    }
    
    return &v1.CreateUserResponse{User: convertUser(user)}, nil
}
```

### 2. 错误包装和展开

**正确使用错误包装**

```go
// 包装错误以添加上下文
func (s *Service) ProcessData(ctx context.Context, data []byte) error {
    if err := s.validateData(data); err != nil {
        return fmt.Errorf("data validation failed: %w", err)
    }
    
    if err := s.saveData(ctx, data); err != nil {
        return fmt.Errorf("failed to save data: %w", err)
    }
    
    return nil
}

// 检查特定错误类型
func (h *Handler) HandleRequest(ctx context.Context, req *Request) error {
    err := h.service.ProcessData(ctx, req.Data)
    if err != nil {
        // 检查是否是特定的业务错误
        if v1.IsUserAlreadyExists(err) {
            // 特殊处理
            return h.handleUserExists(ctx, err)
        }
        
        // 检查是否是参数错误
        if errno.IsInvalidParameter(err) {
            // 记录参数错误
            h.logger.Warn("Invalid parameter", "error", err)
            return err
        }
        
        // 其他错误转换为内部错误
        h.logger.Error("Internal error", "error", err)
        return errno.ErrorInternalError("处理请求失败")
    }
    
    return nil
}
```

### 3. 错误日志记录

**记录有用的错误信息**

```go
func (s *Service) CreateUser(ctx context.Context, req *CreateUserRequest) error {
    logger := s.logger.WithFields(logrus.Fields{
        "operation": "create_user",
        "username":  req.Username,
        "trace_id":  contextx.TraceID(ctx),
        "user_id":   contextx.UserID(ctx),
    })
    
    if err := s.validateUser(req); err != nil {
        // 参数错误 - 使用 Warn 级别
        logger.WithField("validation_error", err.Error()).Warn("User validation failed")
        return errno.ErrorInvalidParameter(err.Error())
    }
    
    user, err := s.repo.CreateUser(ctx, req)
    if err != nil {
        if v1.IsUserAlreadyExists(err) {
            // 业务错误 - 使用 Info 级别
            logger.Info("User already exists")
            return err
        }
        
        // 系统错误 - 使用 Error 级别
        logger.WithField("db_error", err.Error()).Error("Failed to create user in database")
        return errno.ErrorInternalError("创建用户失败")
    }
    
    logger.WithField("user_id", user.ID).Info("User created successfully")
    return nil
}
```

## 测试最佳实践

### 1. 错误场景测试

**全面测试各种错误场景**

```go
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name          string
        request       *CreateUserRequest
        mockSetup     func(*MockUserRepo)
        expectedError string
        expectedCode  int
    }{
        {
            name: "success",
            request: &CreateUserRequest{
                Username: "testuser",
                Email:    "test@example.com",
            },
            mockSetup: func(repo *MockUserRepo) {
                repo.EXPECT().GetUserByName(gomock.Any(), "testuser").Return(nil, v1.ErrorUserNotFound("user not found"))
                repo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(&User{ID: "123"}, nil)
            },
            expectedError: "",
        },
        {
            name: "user_already_exists",
            request: &CreateUserRequest{
                Username: "existing",
                Email:    "test@example.com",
            },
            mockSetup: func(repo *MockUserRepo) {
                repo.EXPECT().GetUserByName(gomock.Any(), "existing").Return(&User{}, nil)
            },
            expectedError: "UserAlreadyExists",
            expectedCode:  409,
        },
        {
            name: "invalid_username",
            request: &CreateUserRequest{
                Username: "ab", // 太短
                Email:    "test@example.com",
            },
            expectedError: "InvalidParameter",
            expectedCode:  400,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 设置 mock
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()
            
            repo := NewMockUserRepo(ctrl)
            if tt.mockSetup != nil {
                tt.mockSetup(repo)
            }
            
            service := NewUserService(repo)
            
            // 执行测试
            _, err := service.CreateUser(context.Background(), tt.request)
            
            // 验证结果
            if tt.expectedError == "" {
                assert.NoError(t, err)
            } else {
                assert.Error(t, err)
                
                // 验证错误类型
                var kratosErr *errors.Error
                if errors.As(err, &kratosErr) {
                    assert.Equal(t, tt.expectedError, kratosErr.Reason)
                    assert.Equal(t, tt.expectedCode, int(kratosErr.Code))
                }
            }
        })
    }
}
```

### 2. 错误传播测试

**测试错误在不同层之间的传播**

```go
func TestErrorPropagation(t *testing.T) {
    // 测试数据库错误如何传播到 API 层
    t.Run("database_error_propagation", func(t *testing.T) {
        ctrl := gomock.NewController(t)
        defer ctrl.Finish()
        
        repo := NewMockUserRepo(ctrl)
        repo.EXPECT().GetUserByName(gomock.Any(), "test").Return(nil, errors.New("database connection failed"))
        
        service := NewUserService(repo)
        handler := NewUserHandler(service)
        
        req := &v1.CreateUserRequest{Username: "test"}
        _, err := handler.CreateUser(context.Background(), req)
        
        // 应该返回内部服务器错误
        assert.Error(t, err)
        var kratosErr *errors.Error
        if errors.As(err, &kratosErr) {
            assert.Equal(t, "InternalError", kratosErr.Reason)
            assert.Equal(t, 500, int(kratosErr.Code))
        }
    })
}
```

## 监控和告警

### 1. 错误指标收集

**收集关键错误指标**

```go
var (
    errorCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "api_errors_total",
            Help: "Total number of API errors",
        },
        []string{"method", "reason", "code"},
    )
    
    errorDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "api_error_duration_seconds",
            Help: "Duration of API calls that resulted in errors",
        },
        []string{"method", "reason"},
    )
)

func recordError(method, reason string, code int, duration time.Duration) {
    errorCounter.WithLabelValues(method, reason, strconv.Itoa(code)).Inc()
    errorDuration.WithLabelValues(method, reason).Observe(duration.Seconds())
}
```

### 2. 错误告警规则

**设置合理的告警阈值**

```yaml
# Prometheus 告警规则
groups:
- name: api_errors
  rules:
  # 5xx 错误率过高
  - alert: HighServerErrorRate
    expr: |
      (
        sum(rate(api_errors_total{code=~"5.."}[5m])) by (method)
        /
        sum(rate(api_requests_total[5m])) by (method)
      ) > 0.05
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "High server error rate for {{ $labels.method }}"
      description: "Error rate is {{ $value | humanizePercentage }} for method {{ $labels.method }}"
  
  # 特定业务错误激增
  - alert: HighUserAlreadyExistsErrors
    expr: |
      rate(api_errors_total{reason="UserAlreadyExists"}[5m]) > 10
    for: 1m
    labels:
      severity: warning
    annotations:
      summary: "High rate of UserAlreadyExists errors"
      description: "{{ $value }} UserAlreadyExists errors per second"
```

## 性能优化

### 1. 错误对象复用

**避免频繁创建错误对象**

```go
// 预定义常用错误
var (
    ErrInvalidUsername = errno.ErrorInvalidParameter("用户名格式不正确")
    ErrInvalidEmail    = errno.ErrorInvalidParameter("邮箱格式不正确")
)

// 复用错误对象
func validateUser(user *User) error {
    if !isValidUsername(user.Username) {
        return ErrInvalidUsername
    }
    if !isValidEmail(user.Email) {
        return ErrInvalidEmail
    }
    return nil
}
```

### 2. 延迟国际化

**只在需要时进行国际化翻译**

```go
// 错误的做法 - 总是进行翻译
func badExample(ctx context.Context) error {
    message := i18n.FromContext(ctx).T(locales.UserNotFound) // 即使不需要也翻译
    return v1.ErrorUserNotFound(message)
}

// 正确的做法 - 延迟翻译
func goodExample(ctx context.Context) error {
    // 只传递键，在需要时才翻译
    return v1.ErrorUserNotFoundWithKey(locales.UserNotFound)
}

// 在 HTTP 编码器中进行翻译
func (e *ErrorEncoder) Encode(ctx context.Context, err error) {
    if kratosErr, ok := err.(*errors.Error); ok {
        if key, hasKey := kratosErr.Metadata["i18n_key"]; hasKey {
            kratosErr.Message = i18n.FromContext(ctx).T(key.(string))
        }
    }
}
```

## 安全考虑

### 1. 信息泄露防护

**避免在错误消息中暴露敏感信息**

```go
// ❌ 错误做法 - 可能泄露敏感信息
func badExample(userID string) error {
    return fmt.Errorf("user %s not found in database table users", userID)
}

// ✅ 正确做法 - 通用错误消息
func goodExample(userID string) error {
    // 详细信息记录在日志中
    logger.WithField("user_id", userID).Warn("User not found")
    // 返回通用错误消息
    return v1.ErrorUserNotFound("用户不存在")
}
```

### 2. 错误码一致性

**确保相同错误情况返回一致的错误码**

```go
// 定义错误码常量
const (
    CodeUserNotFound = 404
    CodeUserExists   = 409
)

// 在所有相关接口中使用相同的错误码
func GetUser(id string) error {
    return v1.ErrorUserNotFound("用户不存在") // 总是返回 404
}

func DeleteUser(id string) error {
    return v1.ErrorUserNotFound("用户不存在") // 总是返回 404
}
```

## 总结

遵循这些最佳实践可以帮助你构建：

1. **一致的错误处理机制** - 统一的错误格式和处理流程
2. **可维护的错误代码** - 清晰的错误分层和命名规范
3. **用户友好的错误体验** - 有意义的错误消息和国际化支持
4. **可观测的错误系统** - 完善的日志记录和监控指标
5. **安全的错误处理** - 避免信息泄露和安全风险

记住，好的错误处理不仅仅是技术实现，更是用户体验和系统可靠性的重要组成部分。