# 错误处理最佳实践

## 概述

本文档提供了在 go-protoc 项目中使用统一错误处理系统的最佳实践指南。基于新的 ErrorX 架构，这些实践将帮助您构建更加健壮、可维护和用户友好的应用程序。

> **注意**: 本文档基于新的 `pkg/errorsx` 错误处理架构。如果您还在使用旧的错误处理方式，请参考 [错误重新设计文档](./errors-redesign.md) 进行迁移。

## 设计原则

### 1. 一致性原则

**所有错误都应使用统一的 ErrorX 结构**

```json
{
  "code": 400,
  "reason": "INVALID_PARAMETER",
  "message": "参数验证失败",
  "i18n_key": "errors.validation.invalid_parameter",
  "metadata": {
    "field": "username",
    "min_length": 3,
    "actual_length": 2
  },
  "timestamp": "2024-01-15T10:30:00Z",
  "request_id": "req_123456789"
}
```

✅ **正确做法**:
```go
// 使用 ErrorX 构建器
return errorsx.BadRequest().
    WithReason("INVALID_PARAMETER").
    WithMessage("用户名长度不能少于3个字符").
    WithI18nKey("errors.validation.username_too_short").
    AddMetadata("field", "username").
    AddMetadata("min_length", 3).
    AddMetadata("actual_length", len(username)).
    Build()

// 或使用预定义错误
return errors.NewInvalidParameterError("username", "too short")
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
// 系统级错误 - 使用 errorsx 包
errorsx.InternalError().
    WithReason("DATABASE_CONNECTION_FAILED").
    WithMessage("数据库连接失败").
    Build()

errorsx.BadRequest().
    WithReason("INVALID_REQUEST_FORMAT").
    WithMessage("请求格式错误").
    Build()

// 业务级错误 - 使用 errors 包的预定义错误
errors.NewUserAlreadyExistsError("user@example.com")
errors.NewUserNotFoundError("123")

// 或使用错误注册器
registry := errorsx.NewRegistry()
registry.Register("USER_NOT_FOUND", errorsx.NotFound(), "User not found", "errors.user.not_found")
err := registry.Create("USER_NOT_FOUND").AddMetadata("user_id", "123").Build()
```

### 3. 国际化原则

**所有面向用户的错误消息都应支持国际化**

✅ **正确做法**:
```go
// 使用 i18n 键，延迟翻译
return errorsx.Conflict().
    WithReason("USER_ALREADY_EXISTS").
    WithI18nKey("errors.user.already_exists").
    AddMetadata("email", email).
    Build()

// 或在上下文中进行翻译
err := errors.NewUserAlreadyExistsError(email)
return err.Localize(ctx) // 根据上下文语言进行翻译
```

❌ **错误做法**:
```go
// 硬编码中文消息
return errorsx.Conflict().
    WithMessage("用户已存在").
    Build()
```

## 错误定义最佳实践

### 1. 错误命名规范

**使用清晰、具体的错误原因码**

✅ **推荐命名** (使用大写下划线格式):
- `USER_ALREADY_EXISTS` - 明确表示用户已存在
- `SECRET_LIMIT_EXCEEDED` - 明确表示密钥数量达到上限
- `INSUFFICIENT_PERMISSIONS` - 明确表示权限不足
- `INVALID_EMAIL_FORMAT` - 明确表示邮箱格式错误
- `PASSWORD_TOO_WEAK` - 明确表示密码强度不够

❌ **不推荐命名**:
- `USER_ERROR` - 过于宽泛
- `FAILED` - 没有具体信息
- `ERROR_1`, `ERROR_2` - 无意义的命名
- `userAlreadyExists` - 不符合命名规范

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
// 具体、可操作，包含上下文信息
errorsx.BadRequest().
    WithReason("INVALID_USERNAME_LENGTH").
    WithMessage("用户名长度必须在 3-20 个字符之间").
    WithI18nKey("errors.validation.username_length").
    AddMetadata("field", "username").
    AddMetadata("min_length", 3).
    AddMetadata("max_length", 20).
    AddMetadata("actual_length", len(username)).
    Build()

errorsx.BadRequest().
    WithReason("INVALID_EMAIL_FORMAT").
    WithMessage("邮箱格式不正确，请检查后重试").
    WithI18nKey("errors.validation.invalid_email").
    AddMetadata("field", "email").
    AddMetadata("value", email).
    AddMetadata("expected_format", "user@domain.com").
    Build()

errorsx.BadRequest().
    WithReason("SECRET_LIMIT_EXCEEDED").
    WithMessage("密钥数量已达上限，请删除不需要的密钥后重试").
    WithI18nKey("errors.business.secret_limit_exceeded").
    AddMetadata("current_count", currentCount).
    AddMetadata("max_allowed", maxAllowed).
    Build()
```

❌ **不好的错误消息**:
```go
// 模糊、无用
errorsx.InternalError().
    WithMessage("操作失败").
    Build()

errorsx.InternalError().
    WithMessage("系统错误").
    Build()

errorsx.InternalError().
    WithMessage("请联系管理员").
    Build()
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
            return nil, errors.NewUserNotFoundError(id)
        }
        // 系统错误包装为 ErrorX
        return nil, errorsx.InternalError().
            WithReason("DATABASE_QUERY_FAILED").
            WithMessage("Database query failed").
            WithCause(err).
            AddMetadata("operation", "get_user").
            AddMetadata("user_id", id).
            Build()
    }
    return &user, nil
}

// 2. 业务逻辑层 - 处理业务规则
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // 业务规则验证
    if len(req.Username) < 3 {
        return nil, errorsx.BadRequest().
            WithReason("INVALID_USERNAME_LENGTH").
            WithMessage("用户名长度不能少于3个字符").
            WithI18nKey("errors.validation.username_too_short").
            AddMetadata("field", "username").
            AddMetadata("min_length", 3).
            AddMetadata("actual_length", len(req.Username)).
            Build()
    }
    
    // 检查用户是否存在
    existing, err := s.repo.GetUserByName(ctx, req.Username)
    if err != nil {
        var userNotFoundErr *errors.UserNotFoundError
        if !errors.As(err, &userNotFoundErr) {
            // 非"用户不存在"的其他错误，直接返回
            return nil, err
        }
    }
    if existing != nil {
        return nil, errors.NewUserAlreadyExistsError(req.Username)
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
        if errorX, ok := err.(*errorsx.ErrorX); ok {
            h.logger.WithFields(logrus.Fields{
                "username":   req.Username,
                "error_code": errorX.Code,
                "error_reason": errorX.Reason,
                "error_message": errorX.Message,
                "metadata":   errorX.Metadata,
                "trace_id":   contextx.TraceID(ctx),
                "request_id": contextx.RequestID(ctx),
            }).Error("Failed to create user")
        } else {
            h.logger.WithFields(logrus.Fields{
                "username": req.Username,
                "error":    err.Error(),
                "trace_id": contextx.TraceID(ctx),
            }).Error("Failed to create user")
        }
        
        // 返回错误（可能经过本地化处理）
        if errorX, ok := err.(*errorsx.ErrorX); ok {
            return nil, errorX.Localize(ctx)
        }
        return nil, err
    }
    
    return &v1.CreateUserResponse{User: convertUser(user)}, nil
}
```

### 2. 错误包装和展开

**正确使用错误包装和检查**

```go
// 包装错误以添加上下文
func (s *Service) ProcessData(ctx context.Context, data []byte) error {
    if err := s.validateData(data); err != nil {
        return errorsx.BadRequest().
            WithReason("DATA_VALIDATION_FAILED").
            WithMessage("Data validation failed").
            WithCause(err).
            AddMetadata("data_size", len(data)).
            Build()
    }
    
    if err := s.saveData(ctx, data); err != nil {
        return errorsx.InternalError().
            WithReason("DATA_SAVE_FAILED").
            WithMessage("Failed to save data").
            WithCause(err).
            AddMetadata("operation", "save_data").
            AddMetadata("data_size", len(data)).
            Build()
    }
    
    return nil
}

// 检查特定错误类型
func (h *Handler) HandleRequest(ctx context.Context, req *Request) error {
    err := h.service.ProcessData(ctx, req.Data)
    if err != nil {
        // 检查是否是特定的业务错误
        var userExistsErr *errors.UserAlreadyExistsError
        if errors.As(err, &userExistsErr) {
            // 特殊处理
            return h.handleUserExists(ctx, userExistsErr)
        }
        
        // 检查是否是 ErrorX 类型
        if errorX, ok := err.(*errorsx.ErrorX); ok {
            if errorX.Code == 400 {
                // 记录参数错误
                h.logger.Warn("Invalid parameter", "error", errorX)
                return errorX
            }
            
            if errorX.Code >= 500 {
                // 记录系统错误
                h.logger.Error("Internal error", "error", errorX)
                return errorX
            }
        }
        
        // 其他错误转换为内部错误
        h.logger.Error("Unknown error", "error", err)
        return errorsx.InternalError().
            WithReason("REQUEST_PROCESSING_FAILED").
            WithMessage("处理请求失败").
            WithCause(err).
            Build()
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
        "request_id": contextx.RequestID(ctx),
    })
    
    if err := s.validateUser(req); err != nil {
        // 参数错误 - 使用 Warn 级别
        if errorX, ok := err.(*errorsx.ErrorX); ok {
            logger.WithFields(logrus.Fields{
                "error_code": errorX.Code,
                "error_reason": errorX.Reason,
                "error_metadata": errorX.Metadata,
            }).Warn("User validation failed")
        } else {
            logger.WithField("validation_error", err.Error()).Warn("User validation failed")
        }
        return err
    }
    
    user, err := s.repo.CreateUser(ctx, req)
    if err != nil {
        var userExistsErr *errors.UserAlreadyExistsError
        if errors.As(err, &userExistsErr) {
            // 业务错误 - 使用 Info 级别
            logger.WithField("existing_username", userExistsErr.Username).Info("User already exists")
            return err
        }
        
        // 系统错误 - 使用 Error 级别
        if errorX, ok := err.(*errorsx.ErrorX); ok {
            logger.WithFields(logrus.Fields{
                "error_code": errorX.Code,
                "error_reason": errorX.Reason,
                "error_metadata": errorX.Metadata,
                "error_cause": errorX.Cause,
            }).Error("Failed to create user in database")
        } else {
            logger.WithField("db_error", err.Error()).Error("Failed to create user in database")
        }
        return errorsx.InternalError().
            WithReason("USER_CREATION_FAILED").
            WithMessage("创建用户失败").
            WithCause(err).
            Build()
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
        checkError    func(t *testing.T, err error)
    }{
        {
            name: "success",
            request: &CreateUserRequest{
                Username: "testuser",
                Email:    "test@example.com",
            },
            mockSetup: func(repo *MockUserRepo) {
                repo.EXPECT().GetUserByName(gomock.Any(), "testuser").Return(nil, errors.NewUserNotFoundError("testuser"))
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
            expectedError: "USER_ALREADY_EXISTS",
            expectedCode:  409,
            checkError: func(t *testing.T, err error) {
                var userExistsErr *errors.UserAlreadyExistsError
                assert.True(t, errors.As(err, &userExistsErr))
                assert.Equal(t, "existing", userExistsErr.Username)
            },
        },
        {
            name: "invalid_username",
            request: &CreateUserRequest{
                Username: "ab", // 太短
                Email:    "test@example.com",
            },
            expectedError: "INVALID_USERNAME_LENGTH",
            expectedCode:  400,
            checkError: func(t *testing.T, err error) {
                errorX, ok := err.(*errorsx.ErrorX)
                assert.True(t, ok)
                assert.Equal(t, 400, errorX.Code)
                assert.Equal(t, "INVALID_USERNAME_LENGTH", errorX.Reason)
                assert.Contains(t, errorX.Metadata, "field")
                assert.Equal(t, "username", errorX.Metadata["field"])
                assert.Equal(t, 3, errorX.Metadata["min_length"])
                assert.Equal(t, 2, errorX.Metadata["actual_length"])
            },
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
                
                // 使用自定义检查函数
                if tt.checkError != nil {
                    tt.checkError(t, err)
                }
                
                // 验证错误类型
                if errorX, ok := err.(*errorsx.ErrorX); ok {
                    assert.Equal(t, tt.expectedError, errorX.Reason)
                    assert.Equal(t, tt.expectedCode, errorX.Code)
                } else if businessErr, ok := err.(interface{ GetReason() string }); ok {
                    assert.Equal(t, tt.expectedError, businessErr.GetReason())
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
    
    // 测试错误如何从数据层传播到API层
    t.Run("error_chain_propagation", func(t *testing.T) {
        // 模拟数据库错误
        dbErr := errors.New("connection timeout")
        
        // 数据层包装错误
        repoErr := errorsx.InternalError().
            WithReason("DATABASE_CONNECTION_TIMEOUT").
            WithMessage("Database connection timeout").
            WithCause(dbErr).
            AddMetadata("operation", "query_user").
            AddMetadata("timeout", "30s").
            Build()
        
        // 服务层处理错误
        serviceErr := errorsx.InternalError().
            WithReason("USER_QUERY_FAILED").
            WithMessage("用户查询失败").
            WithCause(repoErr).
            AddMetadata("service", "user_service").
            Build()
        
        // API层最终错误（可能进行本地化）
        ctx := context.WithValue(context.Background(), "lang", "zh-CN")
        apiErr := serviceErr.Localize(ctx)
        
        // 验证错误链
        assert.True(t, errors.Is(serviceErr, dbErr))
        assert.True(t, errors.Is(repoErr, dbErr))
        
        // 验证 ErrorX 结构
        if errorX, ok := serviceErr.(*errorsx.ErrorX); ok {
            assert.Equal(t, 500, errorX.Code)
            assert.Equal(t, "USER_QUERY_FAILED", errorX.Reason)
            assert.Equal(t, "用户查询失败", errorX.Message)
            assert.NotNil(t, errorX.Cause)
            assert.Contains(t, errorX.Metadata, "service")
            
            // 验证原因错误也是 ErrorX
            if causeErrorX, ok := errorX.Cause.(*errorsx.ErrorX); ok {
                assert.Equal(t, "DATABASE_CONNECTION_TIMEOUT", causeErrorX.Reason)
                assert.True(t, errors.Is(causeErrorX, dbErr))
            }
        }
        
        // 验证本地化后的错误
        if localizedErr, ok := apiErr.(*errorsx.ErrorX); ok {
            assert.Equal(t, serviceErr.(*errorsx.ErrorX).Code, localizedErr.Code)
            assert.Equal(t, serviceErr.(*errorsx.ErrorX).Reason, localizedErr.Reason)
            // 消息可能已被本地化
        }
    })
}
```

## 监控和告警

### 1. 错误指标收集

**收集关键错误指标**

```go
// Prometheus 指标定义
var (
    apiErrorsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "api_errors_total",
            Help: "Total number of API errors",
        },
        []string{"method", "endpoint", "error_code", "error_reason", "error_type", "service"},
    )
    
    apiErrorDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "api_error_duration_seconds",
            Help: "Duration of API calls that resulted in errors",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint", "error_code", "error_reason"},
    )
    
    businessErrorsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "business_errors_total",
            Help: "Total number of business logic errors",
        },
        []string{"error_reason", "service", "operation"},
    )
)

// 错误指标记录
func recordErrorMetrics(ctx context.Context, err error, method, endpoint string, duration time.Duration) {
    var errorCode, errorReason, errorType, service string
    
    // 处理 ErrorX 类型
    if errorX, ok := err.(*errorsx.ErrorX); ok {
        errorCode = strconv.Itoa(errorX.Code)
        errorReason = errorX.Reason
        
        // 根据 HTTP 状态码分类
        switch {
        case errorX.Code >= 500:
            errorType = "server_error"
        case errorX.Code >= 400:
            errorType = "client_error"
        default:
            errorType = "success"
        }
        
        // 从元数据中获取服务信息
        if svc, exists := errorX.Metadata["service"]; exists {
            if svcStr, ok := svc.(string); ok {
                service = svcStr
            }
        }
        
        // 记录业务错误
        if errorX.Code < 500 {
            operation := "unknown"
            if op, exists := errorX.Metadata["operation"]; exists {
                if opStr, ok := op.(string); ok {
                    operation = opStr
                }
            }
            businessErrorsTotal.WithLabelValues(errorReason, service, operation).Inc()
        }
    } else if businessErr, ok := err.(interface{ GetReason() string }); ok {
        // 处理业务错误类型
        errorReason = businessErr.GetReason()
        errorCode = "400" // 默认为客户端错误
        errorType = "business_error"
    } else {
        // 未知错误类型
        errorCode = "500"
        errorReason = "UNKNOWN_ERROR"
        errorType = "system_error"
    }
    
    if service == "" {
        service = "unknown"
    }
    
    apiErrorsTotal.WithLabelValues(method, endpoint, errorCode, errorReason, errorType, service).Inc()
    apiErrorDuration.WithLabelValues(method, endpoint, errorCode, errorReason).Observe(duration.Seconds())
}
```

### 2. 错误告警规则

**设置合理的告警阈值**

```yaml
# Prometheus 告警规则
groups:
  - name: api_errors
    rules:
      - alert: HighServerErrorRate
        expr: |
          (
            sum(rate(api_errors_total{error_type="server_error"}[5m])) by (service)
            /
            sum(rate(api_requests_total[5m])) by (service)
          ) > 0.05
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High 5xx error rate detected"
          description: "Service {{ $labels.service }} has {{ $value | humanizePercentage }} 5xx error rate"
      
      - alert: HighClientErrorRate
        expr: |
          (
            sum(rate(api_errors_total{error_type="client_error"}[5m])) by (service, error_reason)
            /
            sum(rate(api_requests_total[5m])) by (service)
          ) > 0.20
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High 4xx error rate detected"
          description: "Service {{ $labels.service }} has {{ $value | humanizePercentage }} 4xx error rate for {{ $labels.error_reason }}"
      
      - alert: BusinessErrorSpike
        expr: |
          sum(rate(business_errors_total{error_reason="USER_ALREADY_EXISTS"}[5m])) > 10
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Business error spike detected"
          description: "USER_ALREADY_EXISTS error rate is {{ $value }} per second"
      
      - alert: DatabaseErrorSpike
        expr: |
          sum(rate(api_errors_total{error_reason=~"DATABASE_.*"}[5m])) > 5
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Database error spike detected"
          description: "Database-related errors rate is {{ $value }} per second"
      
      - alert: ValidationErrorSpike
        expr: |
          sum(rate(api_errors_total{error_reason=~"INVALID_.*"}[5m])) > 50
        for: 3m
        labels:
          severity: warning
        annotations:
          summary: "Validation error spike detected"
          description: "Validation errors rate is {{ $value }} per second, possible attack or client issue"
```

## 性能优化

### 1. 错误对象复用

**避免频繁创建错误对象**

```go
// 预定义常用错误
var (
    ErrInvalidUsername = errorsx.BadRequest().
        WithReason("INVALID_USERNAME").
        WithMessage("用户名格式不正确").
        WithI18nKey("errors.validation.invalid_username").
        Build()
    ErrInvalidEmail = errorsx.BadRequest().
        WithReason("INVALID_EMAIL").
        WithMessage("邮箱格式不正确").
        WithI18nKey("errors.validation.invalid_email").
        Build()
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

// 使用对象池减少内存分配
var errorXPool = sync.Pool{
    New: func() interface{} {
        return &errorsx.ErrorX{
            Metadata: make(map[string]interface{}),
        }
    },
}

// 从池中获取 ErrorX 对象
func getPooledErrorX() *errorsx.ErrorX {
    return errorXPool.Get().(*errorsx.ErrorX)
}

// 释放 ErrorX 对象到池中
func releaseErrorX(err *errorsx.ErrorX) {
    if err != nil {
        // 重置对象状态
        err.Code = 0
        err.Reason = ""
        err.Message = ""
        err.I18nKey = ""
        err.Cause = nil
        err.Timestamp = time.Time{}
        
        // 清空但保留 map 容量
        for k := range err.Metadata {
            delete(err.Metadata, k)
        }
        
        errorXPool.Put(err)
    }
}
```

### 2. 延迟国际化

**只在需要时进行国际化翻译**

```go
// 错误的做法 - 总是进行翻译
func badExample(ctx context.Context) error {
    message := i18n.FromContext(ctx).T("errors.user.not_found") // 即使不需要也翻译
    return errorsx.NotFound().
        WithReason("USER_NOT_FOUND").
        WithMessage(message).
        Build()
}

// 正确的做法 - 延迟翻译
func goodExample(ctx context.Context) error {
    // 只传递键，在需要时才翻译
    return errorsx.NotFound().
        WithReason("USER_NOT_FOUND").
        WithMessage("User not found").
        WithI18nKey("errors.user.not_found").
        Build()
}

// 高效的国际化缓存
type I18nCache struct {
    cache sync.Map // map[string]map[string]string
    mutex sync.RWMutex
}

var globalI18nCache = &I18nCache{}

func (c *I18nCache) Get(lang, key string) (string, bool) {
    if langCache, ok := c.cache.Load(lang); ok {
        if cache, ok := langCache.(map[string]string); ok {
            if value, exists := cache[key]; exists {
                return value, true
            }
        }
    }
    return "", false
}

func (c *I18nCache) Set(lang, key, value string) {
    langCacheInterface, _ := c.cache.LoadOrStore(lang, make(map[string]string))
    langCache := langCacheInterface.(map[string]string)
    
    c.mutex.Lock()
    langCache[key] = value
    c.mutex.Unlock()
}

// 优化的本地化方法
func (e *ErrorX) LocalizeWithCache(ctx context.Context) *ErrorX {
    if e.I18nKey == "" {
        return e
    }
    
    lang := i18n.GetLanguageFromContext(ctx)
    if lang == "" {
        return e
    }
    
    // 尝试从缓存获取
    if cached, found := globalI18nCache.Get(lang, e.I18nKey); found {
        localizedErr := *e // 浅拷贝
        localizedErr.Message = cached
        return &localizedErr
    }
    
    // 缓存未命中，进行翻译并缓存
    if translator := i18n.FromContext(ctx); translator != nil {
        translated := translator.T(e.I18nKey)
        globalI18nCache.Set(lang, e.I18nKey, translated)
        
        localizedErr := *e // 浅拷贝
        localizedErr.Message = translated
        return &localizedErr
    }
    
    return e
}
```

## 安全考虑

### 1. 信息泄露防护

**避免在错误消息中暴露敏感信息**

```go
// ❌ 错误做法 - 可能泄露敏感信息
func badExample(userID string) error {
    return fmt.Errorf("failed to query user %s from database: connection string mysql://user:password@localhost/db", userID)
}

// ✅ 正确做法 - 隐藏敏感信息
func goodExample(userID string) error {
    // 记录详细错误到日志（包含敏感信息）
    logger.Error("Database connection failed", 
        "user_id", userID,
        "connection", "mysql://user:***@localhost/db",
        "error_code", "DB_CONNECTION_FAILED",
    )
    
    // 返回安全的错误消息（不包含敏感信息）
    return errorsx.InternalError().
        WithReason("DATABASE_CONNECTION_FAILED").
        WithMessage("数据库连接失败").
        WithI18nKey("errors.database.connection_failed").
        AddMetadata("operation", "user_query").
        AddMetadata("safe_user_id", maskUserID(userID)). // 脱敏后的用户ID
        Build()
}

// 敏感信息脱敏函数
func maskUserID(userID string) string {
    if len(userID) <= 4 {
        return "****"
    }
    return userID[:2] + "****" + userID[len(userID)-2:]
}

// 错误信息过滤器
type ErrorSanitizer struct {
    sensitivePatterns []*regexp.Regexp
}

func NewErrorSanitizer() *ErrorSanitizer {
    patterns := []*regexp.Regexp{
        regexp.MustCompile(`password[=:]\s*\S+`),
        regexp.MustCompile(`token[=:]\s*\S+`),
        regexp.MustCompile(`key[=:]\s*\S+`),
        regexp.MustCompile(`secret[=:]\s*\S+`),
        regexp.MustCompile(`mysql://[^:]+:[^@]+@`),
        regexp.MustCompile(`postgres://[^:]+:[^@]+@`),
    }
    
    return &ErrorSanitizer{
        sensitivePatterns: patterns,
    }
}

func (s *ErrorSanitizer) SanitizeMessage(message string) string {
    result := message
    for _, pattern := range s.sensitivePatterns {
        result = pattern.ReplaceAllString(result, "[REDACTED]")
    }
    return result
}

// 安全的错误构建器
func SafeInternalError(reason, message string) *errorsx.ErrorX {
    sanitizer := NewErrorSanitizer()
    
    return errorsx.InternalError().
        WithReason(reason).
        WithMessage(sanitizer.SanitizeMessage(message)).
        Build()
}
```

### 2. 错误码一致性和审计

**确保相同错误情况返回一致的错误码**

```go
// 定义错误码常量
const (
    CodeUserNotFound = 404
    CodeUserExists   = 409
)

// 在所有相关接口中使用相同的错误码
func GetUser(id string) error {
    return errors.NewUserNotFoundError(id) // 总是返回 404
}

func DeleteUser(id string) error {
    return errors.NewUserNotFoundError(id) // 总是返回 404
}

// 错误审计日志
type ErrorAuditor struct {
    logger *logrus.Logger
}

func NewErrorAuditor() *ErrorAuditor {
    return &ErrorAuditor{
        logger: logrus.New(),
    }
}

func (a *ErrorAuditor) AuditError(ctx context.Context, err error, operation string) {
    if errorX, ok := err.(*errorsx.ErrorX); ok {
        // 记录错误审计信息
        a.logger.WithFields(logrus.Fields{
            "timestamp":    time.Now().UTC(),
            "operation":    operation,
            "error_code":   errorX.Code,
            "error_reason": errorX.Reason,
            "user_id":      contextx.UserID(ctx),
            "request_id":   contextx.RequestID(ctx),
            "trace_id":     contextx.TraceID(ctx),
            "ip_address":   contextx.ClientIP(ctx),
            "user_agent":   contextx.UserAgent(ctx),
            "metadata":     errorX.Metadata,
        }).Info("Error audit log")
        
        // 对于敏感操作的错误，记录额外审计信息
        if isSensitiveOperation(operation) {
            a.logger.WithFields(logrus.Fields{
                "security_event": "sensitive_operation_error",
                "operation":      operation,
                "error_reason":   errorX.Reason,
                "user_id":        contextx.UserID(ctx),
                "timestamp":      time.Now().UTC(),
            }).Warn("Sensitive operation failed")
        }
    }
}

func isSensitiveOperation(operation string) bool {
    sensitiveOps := []string{
        "login", "password_change", "permission_grant",
        "data_export", "admin_action", "payment_process",
    }
    
    for _, op := range sensitiveOps {
        if strings.Contains(operation, op) {
            return true
        }
    }
    return false
}
```

## 总结

本文档基于新的 `ErrorX` 架构提供了全面的错误处理最佳实践指南，涵盖了从错误定义到监控告警的完整流程。遵循这些最佳实践可以帮助你构建：

1. **统一的错误处理机制** - 基于 `ErrorX` 结构的一致错误格式和处理流程
2. **可维护的错误代码** - 清晰的错误分层和命名规范，支持丰富的元数据
3. **用户友好的错误体验** - 有意义的错误消息和内置国际化支持
4. **可观测的错误系统** - 结构化错误信息，完善的日志记录和监控指标
5. **安全的错误处理** - 敏感信息脱敏和错误审计机制
6. **高性能的错误处理** - 错误对象池化和延迟国际化优化

**核心要点：**

- **统一架构**：使用 `ErrorX` 结构提供一致的错误格式和丰富的元数据支持
- **分层处理**：在不同层次正确使用 `errorsx` 构建器和业务错误类型
- **国际化支持**：内置 i18n 键支持，实现延迟翻译和多语言错误消息
- **可观测性**：结构化错误信息，完善的日志记录、指标收集和监控告警
- **性能优化**：错误对象池化、延迟国际化和批量错误处理
- **安全考虑**：敏感信息脱敏、错误审计和安全的错误消息构建

**迁移建议：**

- 从旧的错误处理系统迁移到 `ErrorX` 架构时，建议逐步替换
- 优先迁移核心业务逻辑和 API 层的错误处理
- 保持向后兼容性，确保现有错误处理逻辑正常工作
- 充分利用 `ErrorX` 的元数据和国际化特性提升用户体验

记住，好的错误处理不仅仅是技术实现，更是用户体验和系统可靠性的重要组成部分。通过遵循这些基于 `ErrorX` 架构的最佳实践，你的错误处理系统将更加专业、可靠和易于维护。