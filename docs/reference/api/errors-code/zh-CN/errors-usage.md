# 错误处理使用指南

## 概述

本文档详细说明了项目中基于 **ErrorX** 架构的错误处理机制的设计、使用方法和最佳实践。ErrorX 是一个现代化的错误处理框架，提供了统一的错误结构、强大的构建器模式、内置国际化支持和丰富的元数据功能。

> **注意**: 本文档描述的是新的 ErrorX 架构。如果您正在从旧的 Kratos 错误系统迁移，请参考 [错误处理最佳实践](./errors-best-practices.md) 中的迁移指南。

## ErrorX 架构

### 1. 架构概览

```
业务逻辑 → ErrorX 构建 → 错误注册 → HTTP 响应编码 → 客户端
     ↓           ↓            ↓             ↓
  验证层    →  构建器模式  →  错误注册器  →  统一响应格式
```

### 2. 核心组件

#### 2.1 ErrorX 核心层

- **ErrorX 结构体** (`pkg/errorsx/errorx.go`): 统一的错误数据结构
- **构建器模式** (`pkg/errorsx/builders.go`): 流式 API 构建错误
- **预定义错误** (`pkg/errors/`): 业务领域特定错误

#### 2.2 错误处理层

- **错误注册器** (`pkg/errorsx/registry.go`): 统一错误注册和管理
- **HTTP 编码器** (`pkg/server/http_codec.go`): ErrorX 到 HTTP 响应转换
- **中间件** (`pkg/middleware/errors.go`): 错误处理中间件

#### 2.3 国际化层

- **内置 i18n 支持**: ErrorX 原生支持国际化键
- **延迟翻译**: 支持在上下文中进行翻译
- **多语言配置** (`configs/i18n/`): 多语言错误消息配置

## ErrorX 结构定义

### 1. ErrorX 数据结构

每个 ErrorX 实例包含以下字段：

```json
{
  "code": 409,                           // HTTP 状态码
  "reason": "USER_ALREADY_EXISTS",       // 错误原因码（大写下划线格式）
  "message": "用户已存在",                // 错误消息
  "i18n_key": "user.already.exists",    // 国际化键（可选）
  "metadata": {                          // 元数据（可选）
    "username": "john_doe",
    "service": "user-service",
    "trace_id": "abc123"
  }
}
```

### 2. ErrorX 分类

#### 2.1 通用 ErrorX 构建器

| 构建器方法 | HTTP 状态码 | 使用场景 |
|-----------|------------|----------|
| `errorsx.BadRequest()` | 400 | 客户端请求错误 |
| `errorsx.Unauthorized()` | 401 | 未授权访问 |
| `errorsx.Forbidden()` | 403 | 禁止访问 |
| `errorsx.NotFound()` | 404 | 资源未找到 |
| `errorsx.Conflict()` | 409 | 资源冲突 |
| `errorsx.InternalError()` | 500 | 内部服务器错误 |
| `errorsx.ServiceUnavailable()` | 503 | 服务不可用 |

#### 2.2 预定义业务错误

| 错误函数 | HTTP 状态码 | 错误原因码 | 描述 |
|---------|------------|-----------|------|
| `errors.NewUserLoginFailedError()` | 401 | USER_LOGIN_FAILED | 用户登录失败 |
| `errors.NewUserAlreadyExistsError()` | 409 | USER_ALREADY_EXISTS | 用户已存在 |
| `errors.NewUserNotFoundError()` | 404 | USER_NOT_FOUND | 用户未找到 |
| `errors.NewUserCreateFailedError()` | 500 | USER_CREATE_FAILED | 创建用户失败 |
| `errors.NewUserOperationForbiddenError()` | 403 | USER_OPERATION_FORBIDDEN | 用户操作被禁止 |
| `errors.NewSecretReachMaxCountError()` | 400 | SECRET_REACH_MAX_COUNT | 密钥达到最大数量 |
| `errors.NewSecretNotFoundError()` | 404 | SECRET_NOT_FOUND | 密钥未找到 |
| `errors.NewSecretCreateFailedError()` | 500 | SECRET_CREATE_FAILED | 创建密钥失败 |

## 使用方法

### 1. 在业务逻辑中使用 ErrorX

#### 1.1 使用预定义业务错误

```go
package user

import (
    "context"
    "github.com/costa92/go-protoc/v2/pkg/errors"
    "github.com/costa92/go-protoc/v2/pkg/errorsx"
)

func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
    // 检查用户是否已存在
    if userExists(req.Username) {
        // 使用预定义业务错误
        return nil, errors.NewUserAlreadyExistsError(
            "用户已存在",
            errorsx.WithI18nKey("user.already.exists"),
            errorsx.WithMetadata(map[string]interface{}{
                "username": req.Username,
                "service":  "user-service",
            }),
        )
    }
    
    // 业务逻辑...
    return &CreateUserResponse{}, nil
}
```

#### 1.2 使用 ErrorX 构建器

```go
import (
    "github.com/costa92/go-protoc/v2/pkg/errorsx"
)

func (s *Service) ValidateRequest(req *Request) error {
    if req.Username == "" {
        return errorsx.BadRequest().
            WithReason("EMPTY_USERNAME").
            WithMessage("用户名不能为空").
            WithI18nKey("validation.username.empty").
            WithMetadata(map[string]interface{}{
                "field": "username",
            }).
            Build()
    }
    
    if len(req.Username) < 3 {
        return errorsx.BadRequest().
            WithReason("INVALID_USERNAME_LENGTH").
            WithMessage("用户名长度至少需要3个字符").
            WithI18nKey("validation.username.min_length").
            WithMetadata(map[string]interface{}{
                "field":      "username",
                "min_length": 3,
                "actual":     len(req.Username),
            }).
            Build()
    }
    
    return nil
}
```

### 2. 在验证层中使用 ErrorX

```go
package validation

import (
    "context"
    "github.com/costa92/go-protoc/v2/pkg/errors"
    "github.com/costa92/go-protoc/v2/pkg/errorsx"
)

func (v *UserValidator) ValidateCreateUserRequest(ctx context.Context, req interface{}) error {
    userReq, ok := req.(*CreateUserRequest)
    if !ok {
        return errorsx.BadRequest().
            WithReason("INVALID_REQUEST_TYPE").
            WithMessage("无效的请求类型").
            WithI18nKey("validation.request.invalid_type").
            Build()
    }
    
    // 用户名验证
    if err := v.validateUsername(userReq.Username); err != nil {
        return err
    }
    
    // 邮箱验证
    if err := v.validateEmail(userReq.Email); err != nil {
        return err
    }
    
    return nil
}

func (v *UserValidator) validateUsername(username string) error {
    if username == "" {
        return errorsx.BadRequest().
            WithReason("EMPTY_USERNAME").
            WithMessage("用户名不能为空").
            WithI18nKey("validation.username.empty").
            Build()
    }
    
    if len(username) < 3 || len(username) > 50 {
        return errorsx.BadRequest().
            WithReason("INVALID_USERNAME_LENGTH").
            WithMessage("用户名长度必须在3-50个字符之间").
            WithI18nKey("validation.username.length").
            WithMetadata(map[string]interface{}{
                "min_length": 3,
                "max_length": 50,
                "actual":     len(username),
            }).
            Build()
    }
    
    return nil
}

func (v *UserValidator) validateEmail(email string) error {
    if email == "" {
        return errorsx.BadRequest().
            WithReason("EMPTY_EMAIL").
            WithMessage("邮箱不能为空").
            WithI18nKey("validation.email.empty").
            Build()
    }
    
    // 简单的邮箱格式验证
    if !strings.Contains(email, "@") {
        return errorsx.BadRequest().
            WithReason("INVALID_EMAIL_FORMAT").
            WithMessage("邮箱格式无效").
            WithI18nKey("validation.email.invalid_format").
            WithMetadata(map[string]interface{}{
                "email": email,
            }).
            Build()
    }
    
    return nil
}
```

### 3. ErrorX 注册配置

#### 3.1 注册预定义业务错误

在 `pkg/errors/registry.go` 中：

```go
package errors

import (
    "github.com/costa92/go-protoc/v2/pkg/errorsx"
)

// init 函数在包导入时自动执行
func init() {
    // 注册用户相关错误
    errorsx.RegisterError("USER_LOGIN_FAILED", 401)
    errorsx.RegisterError("USER_ALREADY_EXISTS", 409)
    errorsx.RegisterError("USER_NOT_FOUND", 404)
    errorsx.RegisterError("USER_CREATE_FAILED", 500)
    errorsx.RegisterError("USER_OPERATION_FORBIDDEN", 403)
    
    // 注册密钥相关错误
    errorsx.RegisterError("SECRET_REACH_MAX_COUNT", 400)
    errorsx.RegisterError("SECRET_NOT_FOUND", 404)
    errorsx.RegisterError("SECRET_CREATE_FAILED", 500)
    
    // 注册验证相关错误
    errorsx.RegisterError("EMPTY_USERNAME", 400)
    errorsx.RegisterError("INVALID_USERNAME_LENGTH", 400)
    errorsx.RegisterError("EMPTY_EMAIL", 400)
    errorsx.RegisterError("INVALID_EMAIL_FORMAT", 400)
    errorsx.RegisterError("INVALID_REQUEST_TYPE", 400)
}
```

#### 3.2 确保错误注册器初始化

在应用启动文件中导入包：

```go
import (
    _ "github.com/costa92/go-protoc/v2/pkg/errors" // 导入以执行 init 函数
)

func main() {
    // 初始化 ErrorX 注册器
    errorsx.InitRegistry()
    
    // 应用启动逻辑...
}
```

#### 3.3 动态注册错误

```go
package main

import (
    "github.com/costa92/go-protoc/v2/pkg/errorsx"
)

func init() {
    // 批量注册错误
    errors := map[string]int{
        "PAYMENT_FAILED":           402,
        "PAYMENT_TIMEOUT":          408,
        "PAYMENT_INSUFFICIENT":     400,
        "ORDER_NOT_FOUND":          404,
        "ORDER_ALREADY_PROCESSED":  409,
        "INVENTORY_INSUFFICIENT":   400,
    }
    
    for reason, code := range errors {
        errorsx.RegisterError(reason, code)
    }
}
```

## 国际化支持

### 1. ErrorX 内置国际化

ErrorX 原生支持国际化键，无需额外的国际化库：

```go
// 创建带国际化键的错误
err := errorsx.BadRequest().
    WithReason("USER_ALREADY_EXISTS").
    WithMessage("用户已存在").
    WithI18nKey("user.already.exists").
    Build()

// 客户端可以使用 i18n_key 进行本地化
fmt.Println(err.I18nKey) // "user.already.exists"
```

### 2. 配置国际化消息文件

在 `configs/i18n/` 目录下创建语言文件：

```yaml
# configs/i18n/zh-CN.yaml
user:
  already:
    exists: "用户已存在"
  not:
    found: "用户未找到"
  login:
    failed: "登录失败"
    
validation:
  username:
    empty: "用户名不能为空"
    length: "用户名长度必须在{min_length}-{max_length}个字符之间"
  email:
    empty: "邮箱不能为空"
    invalid_format: "邮箱格式无效"

# configs/i18n/en-US.yaml
user:
  already:
    exists: "User already exists"
  not:
    found: "User not found"
  login:
    failed: "Login failed"
    
validation:
  username:
    empty: "Username cannot be empty"
    length: "Username length must be between {min_length}-{max_length} characters"
  email:
    empty: "Email cannot be empty"
    invalid_format: "Invalid email format"
```

### 3. 服务端国际化处理

```go
package i18n

import (
    "context"
    "github.com/costa92/go-protoc/v2/pkg/errorsx"
)

// I18nMiddleware 国际化中间件
func I18nMiddleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 从请求头获取语言
            lang := r.Header.Get("Accept-Language")
            if lang == "" {
                lang = "zh-CN" // 默认语言
            }
            
            // 设置语言到上下文
            ctx := context.WithValue(r.Context(), "lang", lang)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// TranslateError 翻译错误消息
func TranslateError(ctx context.Context, err *errorsx.ErrorX) *errorsx.ErrorX {
    if err.I18nKey == "" {
        return err // 没有国际化键，直接返回
    }
    
    lang := ctx.Value("lang").(string)
    translatedMessage := translateMessage(err.I18nKey, lang, err.Metadata)
    
    // 创建新的错误实例，替换消息
    return errorsx.New(err.Code, err.Reason, translatedMessage).
        WithI18nKey(err.I18nKey).
        WithMetadata(err.Metadata).
        Build()
}

func translateMessage(key, lang string, metadata map[string]interface{}) string {
    // 实现消息翻译逻辑
    // 可以使用 go-i18n 或其他国际化库
    // 这里简化处理
    
    messages := map[string]map[string]string{
        "zh-CN": {
            "user.already.exists":           "用户已存在",
            "validation.username.empty":     "用户名不能为空",
            "validation.username.length":    "用户名长度必须在{min_length}-{max_length}个字符之间",
            "validation.email.invalid_format": "邮箱格式无效",
        },
        "en-US": {
            "user.already.exists":           "User already exists",
            "validation.username.empty":     "Username cannot be empty",
            "validation.username.length":    "Username length must be between {min_length}-{max_length} characters",
            "validation.email.invalid_format": "Invalid email format",
        },
    }
    
    if langMessages, ok := messages[lang]; ok {
        if message, ok := langMessages[key]; ok {
            // 替换占位符
            return replacePlaceholders(message, metadata)
        }
    }
    
    // 回退到默认语言
    if defaultMessages, ok := messages["zh-CN"]; ok {
        if message, ok := defaultMessages[key]; ok {
            return replacePlaceholders(message, metadata)
        }
    }
    
    return key // 最后回退到键本身
}

func replacePlaceholders(message string, metadata map[string]interface{}) string {
    // 实现占位符替换逻辑
    // 例如：将 {min_length} 替换为实际值
    result := message
    for key, value := range metadata {
        placeholder := "{" + key + "}"
        result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
    }
    return result
}
```

### 4. 客户端国际化处理

```javascript
// JavaScript 客户端示例
class I18nErrorHandler {
    constructor(locale = 'zh-CN') {
        this.locale = locale;
        this.messages = {
            'zh-CN': {
                'user.already.exists': '用户已存在',
                'validation.username.empty': '用户名不能为空',
                'validation.username.length': '用户名长度必须在{min_length}-{max_length}个字符之间'
            },
            'en-US': {
                'user.already.exists': 'User already exists',
                'validation.username.empty': 'Username cannot be empty',
                'validation.username.length': 'Username length must be between {min_length}-{max_length} characters'
            }
        };
    }
    
    translateError(error) {
        if (!error.i18n_key) {
            return error.message; // 没有国际化键，使用原始消息
        }
        
        const messages = this.messages[this.locale] || this.messages['zh-CN'];
        let message = messages[error.i18n_key] || error.message;
        
        // 替换占位符
        if (error.metadata) {
            Object.keys(error.metadata).forEach(key => {
                const placeholder = `{${key}}`;
                message = message.replace(new RegExp(placeholder, 'g'), error.metadata[key]);
            });
        }
        
        return message;
    }
}

// 使用示例
const i18nHandler = new I18nErrorHandler('zh-CN');

fetch('/api/v1/users', { method: 'POST', body: userData })
    .then(response => {
        if (!response.ok) {
            return response.json().then(error => {
                const localizedMessage = i18nHandler.translateError(error);
                throw new Error(localizedMessage);
            });
        }
        return response.json();
    })
    .catch(error => {
        console.error('Error:', error.message);
    });
```

## 客户端错误处理

### 1. ErrorX 响应格式

客户端接收到的 ErrorX 响应格式：

```json
{
  "code": 409,
  "reason": "USER_ALREADY_EXISTS",
  "message": "用户已存在",
  "i18n_key": "user.already.exists",
  "metadata": {
    "username": "john_doe",
    "user_id": "12345",
    "timestamp": "2024-01-15T10:30:00Z",
    "request_id": "req_abc123"
  }
}
```

### 2. 客户端处理示例

#### 2.1 JavaScript/TypeScript

```typescript
interface ErrorXResponse {
  code: number;
  reason: string;
  message: string;
  i18n_key?: string;
  metadata?: Record<string, any>;
}

class ApiClient {
  private i18nHandler: I18nErrorHandler;
  
  constructor(locale: string = 'zh-CN') {
    this.i18nHandler = new I18nErrorHandler(locale);
  }

  async createUser(userData: any): Promise<any> {
    try {
      const response = await fetch('/api/v1/users', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept-Language': this.i18nHandler.locale,
        },
        body: JSON.stringify(userData),
      });

      if (!response.ok) {
        const error: ErrorXResponse = await response.json();
        throw new ApiError(error, this.i18nHandler);
      }

      return await response.json();
    } catch (error) {
      if (error instanceof ApiError) {
        this.handleApiError(error);
      }
      throw error;
    }
  }

  private handleApiError(error: ApiError) {
    const localizedMessage = error.getLocalizedMessage();
    
    switch (error.reason) {
      case 'USER_ALREADY_EXISTS':
        console.error('用户已存在:', {
          message: localizedMessage,
          userId: error.metadata?.user_id,
          username: error.metadata?.username
        });
        break;
      case 'INVALID_USERNAME_LENGTH':
        console.error('用户名长度无效:', {
          message: localizedMessage,
          minLength: error.metadata?.min_length,
          maxLength: error.metadata?.max_length,
          actualLength: error.metadata?.actual_length
        });
        break;
      case 'VALIDATION_FAILED':
        console.error('验证失败:', {
          message: localizedMessage,
          field: error.metadata?.field,
          value: error.metadata?.value
        });
        break;
      default:
        console.error('未知错误:', localizedMessage);
    }
  }
}

class ApiError extends Error {
  constructor(
    public errorResponse: ErrorXResponse,
    private i18nHandler: I18nErrorHandler
  ) {
    super(errorResponse.message);
    this.name = 'ApiError';
  }

  get code(): number {
    return this.errorResponse.code;
  }

  get reason(): string {
    return this.errorResponse.reason;
  }

  get metadata(): Record<string, any> | undefined {
    return this.errorResponse.metadata;
  }

  get i18nKey(): string | undefined {
    return this.errorResponse.i18n_key;
  }

  getLocalizedMessage(): string {
    return this.i18nHandler.translateError(this.errorResponse);
  }

  // 检查是否为特定类型的错误
  isUserError(): boolean {
    return this.reason.startsWith('USER_');
  }

  isValidationError(): boolean {
    return this.reason.startsWith('VALIDATION_') || this.reason.includes('_INVALID');
  }

  isRetryable(): boolean {
    // 5xx 错误通常可以重试
    return this.code >= 500 && this.code < 600;
  }
}
```

#### 2.2 Go 客户端

```go
type ErrorXResponse struct {
    Code     int                    `json:"code"`
    Reason   string                 `json:"reason"`
    Message  string                 `json:"message"`
    I18nKey  string                 `json:"i18n_key,omitempty"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type ApiError struct {
    *ErrorXResponse
    RequestID string
    Timestamp time.Time
}

func (e *ApiError) Error() string {
    return fmt.Sprintf("[%s] %s (code: %d, reason: %s)", 
        e.RequestID, e.Message, e.Code, e.Reason)
}

// IsUserError 检查是否为用户相关错误
func (e *ApiError) IsUserError() bool {
    return strings.HasPrefix(e.Reason, "USER_")
}

// IsValidationError 检查是否为验证错误
func (e *ApiError) IsValidationError() bool {
    return strings.HasPrefix(e.Reason, "VALIDATION_") || 
           strings.Contains(e.Reason, "_INVALID")
}

// IsRetryable 检查错误是否可重试
func (e *ApiError) IsRetryable() bool {
    return e.Code >= 500 && e.Code < 600
}

type Client struct {
    baseURL string
    client  *http.Client
    locale  string
}

func NewClient(baseURL, locale string) *Client {
    return &Client{
        baseURL: baseURL,
        client:  &http.Client{Timeout: 30 * time.Second},
        locale:  locale,
    }
}

func (c *Client) CreateUser(ctx context.Context, userData interface{}) error {
    resp, err := c.postWithRetry(ctx, "/api/v1/users", userData, 3)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        var errorResp ErrorXResponse
        if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
            return fmt.Errorf("failed to decode error response: %w", err)
        }
        
        apiErr := &ApiError{
            ErrorXResponse: &errorResp,
            RequestID:      resp.Header.Get("X-Request-ID"),
            Timestamp:      time.Now(),
        }
        
        return apiErr
    }

    return nil
}

// postWithRetry 带重试机制的 POST 请求
func (c *Client) postWithRetry(ctx context.Context, path string, data interface{}, maxRetries int) (*http.Response, error) {
    var lastErr error
    
    for attempt := 0; attempt <= maxRetries; attempt++ {
        resp, err := c.post(ctx, path, data)
        if err != nil {
            lastErr = err
            continue
        }
        
        // 如果是客户端错误（4xx），不重试
        if resp.StatusCode >= 400 && resp.StatusCode < 500 {
            return resp, nil
        }
        
        // 如果是服务器错误（5xx）且还有重试次数，则重试
        if resp.StatusCode >= 500 && attempt < maxRetries {
            resp.Body.Close()
            time.Sleep(time.Duration(attempt+1) * time.Second) // 指数退避
            continue
        }
        
        return resp, nil
    }
    
    return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries+1, lastErr)
}

func (c *Client) post(ctx context.Context, path string, data interface{}) (*http.Response, error) {
    // 实现 POST 请求逻辑
    // 设置 Accept-Language 头
    req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+path, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Accept-Language", c.locale)
    req.Header.Set("Content-Type", "application/json")
    
    return c.client.Do(req)
}

func (c *Client) HandleError(err error) {
    if apiErr, ok := err.(*ApiError); ok {
        switch apiErr.Reason {
        case "USER_ALREADY_EXISTS":
            fmt.Printf("用户已存在: %v (用户名: %v)\n", 
                apiErr.Metadata["user_id"], apiErr.Metadata["username"])
        case "INVALID_USERNAME_LENGTH":
            fmt.Printf("用户名长度无效: 实际长度 %v，要求 %v-%v\n",
                apiErr.Metadata["actual_length"],
                apiErr.Metadata["min_length"],
                apiErr.Metadata["max_length"])
        case "VALIDATION_FAILED":
            fmt.Printf("验证失败: 字段 %v，值 %v\n",
                apiErr.Metadata["field"], apiErr.Metadata["value"])
        default:
            fmt.Printf("API 错误: %s\n", apiErr.Error())
        }
        
        // 记录详细信息用于调试
        fmt.Printf("请求ID: %s, 时间: %s\n", apiErr.RequestID, apiErr.Timestamp.Format(time.RFC3339))
        if apiErr.I18nKey != "" {
            fmt.Printf("国际化键: %s\n", apiErr.I18nKey)
        }
    }
}
```

## 最佳实践

### 1. ErrorX 错误定义

1. **使用预定义业务错误**：
   ```go
   // 好的做法 - 使用预定义业务错误
   return errors.NewUserAlreadyExistsError("john_doe")
   
   // 好的做法 - 使用 ErrorX 构建器
   return errorsx.BadRequest().
       WithReason("INVALID_USERNAME_LENGTH").
       WithMessage("用户名长度无效").
       WithI18nKey("validation.username.length").
       WithMetadata(map[string]interface{}{
           "min_length": 3,
           "max_length": 20,
           "actual_length": len(username),
       }).
       Build()
   
   // 避免的做法 - 直接使用通用错误
   return errorsx.BadRequest().WithMessage("错误").Build()
   ```

2. **提供丰富的元数据**：
   ```go
   return errorsx.BadRequest().
       WithReason("VALIDATION_FAILED").
       WithMessage("参数验证失败").
       WithI18nKey("validation.failed").
       WithMetadata(map[string]interface{}{
           "field":        "email",
           "value":        email,
           "expected":     "valid email format",
           "request_id":   requestID,
           "timestamp":    time.Now().Unix(),
       }).
       Build()
   ```

3. **使用一致的错误原因格式**：
   ```go
   // 好的做法 - 使用大写下划线格式
   const (
       ReasonUserAlreadyExists     = "USER_ALREADY_EXISTS"
       ReasonInvalidUsernameLength = "INVALID_USERNAME_LENGTH"
       ReasonValidationFailed      = "VALIDATION_FAILED"
   )
   ```

### 2. ErrorX 错误处理

1. **在适当的层级处理错误**：
   ```go
   // 在 service 层返回业务错误
   func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) error {
       if s.userExists(req.Username) {
           return errors.NewUserAlreadyExistsError(req.Username)
       }
       
       if err := s.validateUser(req); err != nil {
           return err // ErrorX 错误直接返回
       }
       
       return nil
   }
   
   // 在 handler 层处理和记录错误
   func (h *UserHandler) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
       if err := h.userService.CreateUser(ctx, req); err != nil {
           // 记录错误用于监控
           h.logError(ctx, "CreateUser", err)
           return nil, err // ErrorX 错误直接返回，框架会自动处理
       }
       return &CreateUserResponse{}, nil
   }
   ```

2. **错误包装和上下文传递**：
   ```go
   func (s *UserService) processUser(ctx context.Context, userID string) error {
       user, err := s.repo.GetUser(ctx, userID)
       if err != nil {
           // 包装底层错误，添加上下文
           return errorsx.InternalError().
               WithReason("USER_FETCH_FAILED").
               WithMessage("获取用户信息失败").
               WithI18nKey("user.fetch.failed").
               WithMetadata(map[string]interface{}{
                   "user_id": userID,
                   "operation": "GetUser",
               }).
               WithCause(err). // 保留原始错误
               Build()
       }
       
       return nil
   }
   ```

### 3. 性能优化

1. **预定义常用错误**：
   ```go
   // 在 pkg/errors/common.go 中预定义
   var (
       ErrUserNotFound     = errors.NewUserNotFoundError("")
       ErrInvalidToken     = errorsx.Unauthorized().WithReason("INVALID_TOKEN").Build()
       ErrInternalError    = errorsx.InternalError().WithReason("INTERNAL_ERROR").Build()
   )
   
   // 使用时克隆并添加具体信息
   func (s *UserService) GetUser(userID string) error {
       return ErrUserNotFound.WithMetadata(map[string]interface{}{
           "user_id": userID,
       })
   }
   ```

2. **使用错误池减少内存分配**：
   ```go
   var errorXPool = sync.Pool{
       New: func() interface{} {
           return &errorsx.ErrorX{}
       },
   }
   
   func getPooledError() *errorsx.ErrorX {
       return errorXPool.Get().(*errorsx.ErrorX)
   }
   
   func putPooledError(err *errorsx.ErrorX) {
       err.Reset() // 重置错误状态
       errorXPool.Put(err)
   }
   ```

- **延迟国际化**: 只在需要时进行国际化翻译
- **缓存映射**: 错误码映射结果可以适当缓存

### 4. 安全考虑

1. **避免泄露敏感信息**：
   ```go
   func (s *UserService) authenticateUser(username, password string) error {
       user, err := s.repo.GetUserByUsername(username)
       if err != nil {
           // 好的做法 - 不泄露用户是否存在
           return errorsx.Unauthorized().
               WithReason("AUTHENTICATION_FAILED").
               WithMessage("用户名或密码错误").
               WithI18nKey("auth.failed").
               Build()
       }
       
       if !s.verifyPassword(password, user.PasswordHash) {
           // 记录详细信息用于安全审计
           s.auditLogger.LogFailedLogin(username, "invalid_password")
           
           // 返回通用错误消息
           return errorsx.Unauthorized().
               WithReason("AUTHENTICATION_FAILED").
               WithMessage("用户名或密码错误").
               WithI18nKey("auth.failed").
               Build()
       }
       
       return nil
   }
   ```

2. **敏感信息脱敏**：
   ```go
   func (s *UserService) createUser(req *CreateUserRequest) error {
       if err := s.validateEmail(req.Email); err != nil {
           return errorsx.BadRequest().
               WithReason("INVALID_EMAIL_FORMAT").
               WithMessage("邮箱格式无效").
               WithI18nKey("validation.email.invalid").
               WithMetadata(map[string]interface{}{
                   "email": maskEmail(req.Email), // 脱敏处理
                   "field": "email",
               }).
               Build()
       }
       return nil
   }
   
   func maskEmail(email string) string {
       parts := strings.Split(email, "@")
       if len(parts) != 2 {
           return "***@***"
       }
       
       username := parts[0]
       domain := parts[1]
       
       if len(username) <= 2 {
           return "***@" + domain
       }
       
       return username[:1] + "***" + username[len(username)-1:] + "@" + domain
   }
   ```

3. **错误审计和监控**：
   ```go
   type ErrorAuditor struct {
       logger *zap.Logger
   }
   
   func (a *ErrorAuditor) AuditError(ctx context.Context, operation string, err *errorsx.ErrorX) {
       fields := []zap.Field{
           zap.String("operation", operation),
           zap.Int("code", err.Code),
           zap.String("reason", err.Reason),
           zap.String("message", err.Message),
           zap.String("i18n_key", err.I18nKey),
       }
       
       // 添加请求上下文信息
       if requestID := ctx.Value("request_id"); requestID != nil {
           fields = append(fields, zap.String("request_id", requestID.(string)))
       }
       
       if userID := ctx.Value("user_id"); userID != nil {
           fields = append(fields, zap.String("user_id", userID.(string)))
       }
       
       // 根据错误级别记录
       if err.Code >= 500 {
           a.logger.Error("Server error occurred", fields...)
       } else if err.Code >= 400 {
           a.logger.Warn("Client error occurred", fields...)
       } else {
           a.logger.Info("Error occurred", fields...)
       }
   }
   ```

- **错误码一致性**: 确保相同的错误情况返回一致的错误码
- **日志记录**: 记录错误的上下文信息用于安全审计

## 故障排查

### 1. 常见问题

#### 1.1 ErrorX 构建器使用错误

**问题**: ErrorX 构建器链式调用后忘记调用 `Build()` 方法。

**错误示例**:
```go
// 错误：忘记调用 Build()
return errorsx.BadRequest().
    WithReason("USER_ALREADY_EXISTS").
    WithMessage("用户已存在")
```

**正确做法**:
```go
// 正确：必须调用 Build() 完成构建
return errorsx.BadRequest().
    WithReason("USER_ALREADY_EXISTS").
    WithMessage("用户已存在").
    Build()
```

#### 1.2 预定义业务错误未注册

**问题**: 使用预定义业务错误时出现 "error not registered" 错误。

**原因**: 错误注册器未正确初始化或错误未注册。

**解决方案**:
```go
// 1. 确保在 main.go 中导入错误包
import _ "github.com/costa92/go-protoc/v2/pkg/errors"

// 2. 确保调用初始化函数
func main() {
    // 初始化错误注册器
    errorsx.InitRegistry()
    
    // 其他初始化代码...
}

// 3. 检查错误是否正确注册
func checkErrorRegistration() {
    registry := errorsx.GetRegistry()
    if !registry.IsRegistered("USER_ALREADY_EXISTS") {
        log.Warn("错误 USER_ALREADY_EXISTS 未注册")
    }
}
```

#### 1.3 国际化键未生效

**问题**: 设置了 `i18n_key` 但客户端仍显示原始消息。

**原因**: 客户端国际化处理逻辑有问题或消息文件缺失。

**解决方案**:
```go
// 服务端：确保正确设置国际化键
func createUserError() error {
    return errorsx.BadRequest().
        WithReason("USER_ALREADY_EXISTS").
        WithMessage("用户已存在"). // 默认消息
        WithI18nKey("user.already.exists"). // 国际化键
        Build()
}

// 客户端：检查国际化处理逻辑
function translateError(error) {
    console.log('Error i18n_key:', error.i18n_key); // 调试输出
    
    if (!error.i18n_key) {
        return error.message; // 回退到原始消息
    }
    
    const messages = getI18nMessages();
    return messages[error.i18n_key] || error.message;
}
```

#### 1.4 元数据丢失或格式错误

**问题**: 错误的元数据在传输过程中丢失或格式不正确。

**解决方案**:
```go
// 确保元数据可序列化
func createValidationError(field, value string) error {
    metadata := map[string]interface{}{
        "field": field,
        "value": value,
        "timestamp": time.Now().Unix(), // 使用 Unix 时间戳而不是 time.Time
        "request_id": getRequestID(),
    }
    
    // 验证元数据可序列化
    if _, err := json.Marshal(metadata); err != nil {
        log.Errorf("元数据序列化失败: %v", err)
        metadata = map[string]interface{}{
            "field": field,
            "error": "metadata_serialization_failed",
        }
    }
    
    return errorsx.BadRequest().
        WithReason("VALIDATION_FAILED").
        WithMessage("参数验证失败").
        WithI18nKey("validation.failed").
        WithMetadata(metadata).
        Build()
}
```

### 2. 调试技巧

#### 2.1 启用 ErrorX 详细日志

```go
// 在开发环境启用 ErrorX 详细日志
func enableErrorXDebugging() {
    // 设置 ErrorX 调试模式
    errorsx.SetDebugMode(true)
    
    // 添加错误拦截器
    errorsx.AddInterceptor(func(err *errorsx.ErrorX) {
        log.Debugf("ErrorX created: code=%d, reason=%s, message=%s, i18n_key=%s, metadata=%+v",
            err.Code, err.Reason, err.Message, err.I18nKey, err.Metadata)
    })
}

// 中间件记录所有 ErrorX 错误
func ErrorXLoggingMiddleware() gin.HandlerFunc {
    return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
        if err, ok := recovered.(*errorsx.ErrorX); ok {
            log.Errorf("ErrorX panic: %+v", err)
            c.JSON(err.Code, err)
            return
        }
        
        // 处理其他类型的 panic
        c.AbortWithStatus(http.StatusInternalServerError)
    })
}
```

#### 2.2 错误注册状态检查

```go
// 检查错误注册状态的工具函数
func validateErrorXRegistry() {
    registry := errorsx.GetRegistry()
    
    // 检查预定义错误
    predefinedErrors := []string{
        "USER_ALREADY_EXISTS",
        "USER_NOT_FOUND",
        "INVALID_USERNAME_LENGTH",
        "VALIDATION_FAILED",
    }
    
    for _, reason := range predefinedErrors {
        if !registry.IsRegistered(reason) {
            log.Warnf("预定义错误未注册: %s", reason)
        } else {
            errorFunc := registry.GetErrorFunc(reason)
            if errorFunc == nil {
                log.Warnf("错误构造函数为空: %s", reason)
            }
        }
    }
    
    // 输出所有已注册的错误
    log.Infof("已注册的错误数量: %d", registry.Count())
    for _, reason := range registry.ListRegistered() {
        log.Debugf("已注册错误: %s", reason)
    }
}
```

#### 2.3 客户端错误处理调试

```javascript
// JavaScript 客户端调试工具
class ErrorXDebugger {
    static enableDebug() {
        // 拦截所有 fetch 请求
        const originalFetch = window.fetch;
        window.fetch = async function(...args) {
            const response = await originalFetch.apply(this, args);
            
            if (!response.ok) {
                const errorData = await response.clone().json();
                console.group('🚨 ErrorX Debug Info');
                console.log('URL:', args[0]);
                console.log('Status:', response.status);
                console.log('ErrorX Data:', errorData);
                
                // 验证 ErrorX 格式
                ErrorXDebugger.validateErrorXFormat(errorData);
                console.groupEnd();
            }
            
            return response;
        };
    }
    
    static validateErrorXFormat(errorData) {
        const requiredFields = ['code', 'reason', 'message'];
        const optionalFields = ['i18n_key', 'metadata'];
        
        console.log('📋 ErrorX Format Validation:');
        
        requiredFields.forEach(field => {
            if (errorData[field] === undefined) {
                console.error(`❌ Missing required field: ${field}`);
            } else {
                console.log(`✅ ${field}:`, errorData[field]);
            }
        });
        
        optionalFields.forEach(field => {
            if (errorData[field] !== undefined) {
                console.log(`📎 ${field}:`, errorData[field]);
            }
        });
        
        // 检查 reason 格式
        if (errorData.reason && !/^[A-Z_]+$/.test(errorData.reason)) {
            console.warn('⚠️ Reason should be in UPPER_SNAKE_CASE format');
        }
    }
}

// 在开发环境启用调试
if (process.env.NODE_ENV === 'development') {
    ErrorXDebugger.enableDebug();
}
```

#### 2.4 性能监控和分析

```go
// ErrorX 性能监控
type ErrorXMetrics struct {
    errorCounts    map[string]int64
    creationTimes  map[string]time.Duration
    mutex          sync.RWMutex
}

func (m *ErrorXMetrics) RecordError(reason string, creationTime time.Duration) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    
    m.errorCounts[reason]++
    m.creationTimes[reason] = creationTime
}

func (m *ErrorXMetrics) GetStats() map[string]interface{} {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    
    return map[string]interface{}{
        "error_counts":    m.errorCounts,
        "creation_times": m.creationTimes,
    }
}

// 使用示例
var metrics = &ErrorXMetrics{
    errorCounts:   make(map[string]int64),
    creationTimes: make(map[string]time.Duration),
}

func createMonitoredError(reason, message string) error {
    start := time.Now()
    
    err := errorsx.BadRequest().
        WithReason(reason).
        WithMessage(message).
        Build()
    
    metrics.RecordError(reason, time.Since(start))
    return err
}
```

## 参考资料

### 核心文档
- [ErrorX 架构设计文档](./errorx-architecture.md)
- [错误处理快速开始](./errors-quickstart.md)
- [Kratos 框架文档](https://go-kratos.dev/docs/)
- [gRPC 状态码规范](https://grpc.github.io/grpc/core/md_doc_statuscodes.html)
- [HTTP 状态码规范](https://tools.ietf.org/html/rfc7231#section-6)

### 国际化相关
- [Go 国际化库 go-i18n](https://github.com/nicksnyder/go-i18n)
- [国际化最佳实践](https://phrase.com/blog/posts/i18n-best-practices/)
- [Unicode CLDR 规范](https://cldr.unicode.org/)

### 错误处理最佳实践
- [Google API 设计指南 - 错误处理](https://cloud.google.com/apis/design/errors)
- [微服务错误处理模式](https://microservices.io/patterns/reliability/circuit-breaker.html)
- [RESTful API 错误处理](https://blog.restcase.com/rest-api-error-codes-101/)

### 相关工具
- [ErrorX 代码生成器](./tools/errorx-generator.md)
- [错误码检查工具](./tools/error-validator.md)
- [国际化消息提取工具](./tools/i18n-extractor.md)

## 更新日志

### v3.0.0 (2024-03-15) - ErrorX 架构
**重大更新**
- 🎉 引入全新的 ErrorX 架构，替代传统 Kratos 错误系统
- ✨ 新增构建器模式，支持链式调用创建错误
- 🌍 内置国际化支持，自动生成 `i18n_key`
- 📊 增强元数据功能，支持结构化错误信息
- 🔧 新增错误注册器，统一管理预定义业务错误
- 📝 完善的类型安全，编译时错误检查
- 🚀 性能优化，支持错误对象池

**迁移指南**
- 旧版本错误创建方式仍然兼容
- 建议逐步迁移到 ErrorX 构建器模式
- 详见 [迁移指南](./migration-guide.md)

### v2.2.1 (2024-02-20)
**Bug 修复**
- 🐛 修复国际化消息在某些场景下不生效的问题
- 🔧 优化错误码映射性能
- 📚 更新文档示例

### v2.2.0 (2024-01-30)
**功能增强**
- ✨ 新增错误链追踪功能
- 🔍 增强调试信息输出
- 📊 添加错误统计和监控支持
- 🌐 扩展国际化语言支持

### v2.1.0 (2024-01-15)
**功能更新**
- ✨ 新增错误码映射配置
- 🌍 优化国际化支持
- 📱 添加客户端错误处理示例
- 🔧 改进错误响应格式

### v2.0.0 (2023-12-01)
**架构重构**
- 🏗️ 重构错误处理架构
- 📋 统一错误响应格式
- 🌐 新增多语言支持
- 🔒 增强安全性和性能

### v1.0.0 (2023-10-01)
**初始发布**
- 🎉 初始版本发布
- ⚡ 基础错误处理功能
- 📖 完整的文档和示例

---

> **注意**: 从 v3.0.0 开始，推荐使用 ErrorX 架构进行错误处理。旧版本的 Kratos 错误系统仍然支持，但建议在新项目中使用 ErrorX。
> 
> 如有问题或建议，请提交 [Issue](https://github.com/costa92/go-protoc/issues) 或查看 [FAQ](./faq.md)。