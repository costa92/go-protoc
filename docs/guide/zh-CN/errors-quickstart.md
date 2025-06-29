# 错误处理快速开始

## 5分钟快速上手

本指南将帮助你在 5 分钟内掌握项目中基于新 `ErrorX` 架构的错误处理基本用法。

## 第一步：了解错误响应格式

当 API 调用失败时，你会收到如下格式的错误响应：

```json
{
  "code": 409,
  "reason": "USER_ALREADY_EXISTS",
  "message": "用户已存在",
  "i18n_key": "errors.user.already_exists",
  "metadata": {
    "username": "john_doe",
    "service": "user_service"
  }
}
```

- `code`: HTTP 状态码
- `reason`: 错误原因码（用于程序判断，采用大写下划线格式）
- `message`: 人类可读的错误消息（支持国际化）
- `i18n_key`: 国际化键（用于客户端本地化）
- `metadata`: 附加元数据信息（可选，提供更多上下文）

## 第二步：在业务代码中返回错误

### 使用 ErrorX 构建器返回错误

```go
package user

import (
    "context"
    "github.com/costa92/go-protoc/v2/pkg/errorsx"
    "github.com/costa92/go-protoc/v2/pkg/errors"
)

func CreateUser(ctx context.Context, req *CreateUserRequest) error {
    // 检查用户是否已存在
    if userExists(req.Name) {
        // 使用 ErrorX 构建器返回业务错误（HTTP 409）
        return errorsx.Conflict().
            WithReason("USER_ALREADY_EXISTS").
            WithMessage("用户已存在").
            WithI18nKey("errors.user.already_exists").
            AddMetadata("username", req.Name).
            AddMetadata("service", "user_service").
            Build()
    }
    
    // 或者使用预定义的业务错误
    if userExists(req.Name) {
        return errors.NewUserAlreadyExistsError(req.Name)
    }
    
    return nil
}
```

### 返回通用错误

```go
import "github.com/costa92/go-protoc/v2/pkg/errorsx"

func ValidateInput(input string) error {
    if input == "" {
        // 使用 ErrorX 构建器返回参数错误（HTTP 400）
        return errorsx.BadRequest().
            WithReason("EMPTY_INPUT").
            WithMessage("输入不能为空").
            WithI18nKey("errors.validation.empty_input").
            AddMetadata("field", "input").
            Build()
    }
    return nil
}
```

## 第三步：客户端处理错误

### JavaScript 示例

```javascript
// ErrorX 响应接口定义
interface ErrorXResponse {
  code: number;
  reason: string;
  message: string;
  i18n_key?: string;
  metadata?: Record<string, any>;
}

async function createUser(userData) {
  try {
    const response = await fetch('/api/v1/users', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(userData)
    });
    
    if (!response.ok) {
      const error: ErrorXResponse = await response.json();
      
      // 根据错误原因码进行不同处理
      switch (error.reason) {
        case 'USER_ALREADY_EXISTS':
          const username = error.metadata?.username || '该用户';
          alert(`用户名 "${username}" 已被使用，请选择其他用户名`);
          break;
        case 'INVALID_USERNAME_LENGTH':
          const minLength = error.metadata?.min_length || 3;
          alert(`用户名长度至少需要 ${minLength} 个字符`);
          break;
        case 'EMPTY_USERNAME':
          alert('用户名不能为空');
          break;
        default:
          // 使用国际化键进行客户端本地化（如果支持）
          const localizedMessage = error.i18n_key 
            ? i18n.t(error.i18n_key) 
            : error.message;
          alert(localizedMessage || '操作失败');
      }
      return;
    }
    
    const user = await response.json();
    console.log('用户创建成功:', user);
  } catch (err) {
    console.error('网络错误:', err);
  }
}
```

### Go 客户端示例

```go
// ErrorX 响应结构体
type ErrorXResponse struct {
    Code     int                    `json:"code"`
    Reason   string                 `json:"reason"`
    Message  string                 `json:"message"`
    I18nKey  string                 `json:"i18n_key,omitempty"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func handleApiError(resp *http.Response) error {
    var apiError ErrorXResponse
    
    if err := json.NewDecoder(resp.Body).Decode(&apiError); err != nil {
        return fmt.Errorf("failed to decode error response: %w", err)
    }
    
    switch apiError.Reason {
    case "USER_ALREADY_EXISTS":
        username := "unknown"
        if u, ok := apiError.Metadata["username"].(string); ok {
            username = u
        }
        return fmt.Errorf("用户 %s 已存在: %s", username, apiError.Message)
    case "INVALID_USERNAME_LENGTH":
        minLength := 3
        if ml, ok := apiError.Metadata["min_length"].(float64); ok {
            minLength = int(ml)
        }
        return fmt.Errorf("用户名长度至少需要 %d 个字符: %s", minLength, apiError.Message)
    case "EMPTY_USERNAME":
        return fmt.Errorf("用户名不能为空: %s", apiError.Message)
    default:
        return fmt.Errorf("API错误 [%d] %s: %s", apiError.Code, apiError.Reason, apiError.Message)
    }
}

// 带重试的错误处理示例
func createUserWithRetry(client *http.Client, userData interface{}) error {
    maxRetries := 3
    
    for i := 0; i < maxRetries; i++ {
        resp, err := client.Post("/api/v1/users", "application/json", nil)
        if err != nil {
            return fmt.Errorf("network error: %w", err)
        }
        
        if resp.StatusCode == http.StatusOK {
            return nil // 成功
        }
        
        apiErr := handleApiError(resp)
        
        // 根据错误类型决定是否重试
        var errorX ErrorXResponse
        json.NewDecoder(resp.Body).Decode(&errorX)
        
        switch errorX.Reason {
        case "USER_ALREADY_EXISTS", "INVALID_USERNAME_LENGTH", "EMPTY_USERNAME":
            // 客户端错误，不重试
            return apiErr
        case "INTERNAL_ERROR", "DATABASE_CONNECTION_FAILED":
            // 服务器错误，可以重试
            if i == maxRetries-1 {
                return apiErr
            }
            time.Sleep(time.Duration(i+1) * time.Second)
            continue
        default:
            return apiErr
        }
    }
    
    return fmt.Errorf("max retries exceeded")
}
```

## 第四步：常用错误码速查

| 错误原因码 | HTTP码 | 使用场景 | ErrorX 构建器示例 |
|-----------|--------|----------|------------------|
| `INVALID_PARAMETER` | 400 | 参数验证失败 | `errorsx.BadRequest().WithReason("INVALID_PARAMETER").WithMessage("参数无效").Build()` |
| `UNAUTHORIZED` | 401 | 未登录/认证失败 | `errorsx.Unauthorized().WithReason("UNAUTHORIZED").WithMessage("请先登录").Build()` |
| `FORBIDDEN` | 403 | 权限不足 | `errorsx.Forbidden().WithReason("FORBIDDEN").WithMessage("权限不足").Build()` |
| `NOT_FOUND` | 404 | 资源不存在 | `errorsx.NotFound().WithReason("NOT_FOUND").WithMessage("资源未找到").Build()` |
| `USER_ALREADY_EXISTS` | 409 | 用户已存在 | `errorsx.Conflict().WithReason("USER_ALREADY_EXISTS").WithMessage("用户已存在").Build()` |
| `INTERNAL_ERROR` | 500 | 服务器内部错误 | `errorsx.InternalError().WithReason("INTERNAL_ERROR").WithMessage("服务器错误").Build()` |

### 预定义业务错误

| 业务错误 | 使用场景 | 示例代码 |
|---------|----------|----------|
| `UserAlreadyExistsError` | 用户已存在 | `errors.NewUserAlreadyExistsError(username)` |
| `UserNotFoundError` | 用户不存在 | `errors.NewUserNotFoundError(userID)` |
| `InvalidCredentialsError` | 认证失败 | `errors.NewInvalidCredentialsError()` |

## 第五步：添加国际化支持

### 1. 使用内置国际化键

```go
// ErrorX 内置国际化支持
return errorsx.BadRequest().
    WithReason("CUSTOM_ERROR").
    WithMessage("自定义错误消息").
    WithI18nKey("errors.custom.error"). // 设置国际化键
    Build()
```

### 2. 配置多语言消息

```yaml
# zh-CN.yaml
errors:
  custom:
    error: "自定义错误消息"

# en-US.yaml  
errors:
  custom:
    error: "Custom error message"
```

### 3. 客户端本地化

```go
// 服务端延迟国际化
func CreateError() error {
    return errorsx.BadRequest().
        WithReason("CUSTOM_ERROR").
        WithI18nKey("errors.custom.error"). // 只设置键，不设置消息
        Build()
}

// 在需要时进行本地化
func LocalizeError(ctx context.Context, err error) error {
    if errorX, ok := err.(*errorsx.ErrorX); ok {
        if errorX.I18nKey != "" {
            message := i18n.FromContext(ctx).T(errorX.I18nKey)
            return errorX.WithMessage(message)
        }
    }
    return err
}
```

## 完整示例

这是一个完整的用户创建接口示例，展示了新 ErrorX 架构的使用：

```go
package handler

import (
    "context"
    "github.com/costa92/go-protoc/v2/pkg/errorsx"
    "github.com/costa92/go-protoc/v2/pkg/errors"
    "github.com/costa92/go-protoc/v2/pkg/i18n"
)

func (h *UserHandler) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
    // 1. 参数验证
    if req.Name == "" {
        return nil, errorsx.BadRequest().
            WithReason("EMPTY_USERNAME").
            WithMessage("用户名不能为空").
            WithI18nKey("errors.validation.empty_username").
            AddMetadata("field", "name").
            Build()
    }
    
    if len(req.Name) < 3 {
        return nil, errorsx.BadRequest().
            WithReason("INVALID_USERNAME_LENGTH").
            WithMessage("用户名长度至少3个字符").
            WithI18nKey("errors.validation.username_too_short").
            AddMetadata("field", "name").
            AddMetadata("min_length", 3).
            AddMetadata("actual_length", len(req.Name)).
            Build()
    }
    
    // 2. 业务逻辑检查
    if h.userService.UserExists(ctx, req.Name) {
        // 使用预定义业务错误
        return nil, errors.NewUserAlreadyExistsError(req.Name)
        
        // 或者使用 ErrorX 构建器
        // return nil, errorsx.Conflict().
        //     WithReason("USER_ALREADY_EXISTS").
        //     WithMessage("用户已存在").
        //     WithI18nKey("errors.user.already_exists").
        //     AddMetadata("username", req.Name).
        //     Build()
    }
    
    // 3. 创建用户
    user, err := h.userService.CreateUser(ctx, req)
    if err != nil {
        // 记录详细错误日志
        h.logger.WithError(err).Error("创建用户失败")
        
        // 包装底层错误并返回
        return nil, errorsx.InternalError().
            WithReason("USER_CREATE_FAILED").
            WithMessage("创建用户失败").
            WithI18nKey("errors.user.create_failed").
            AddMetadata("operation", "create_user").
            AddMetadata("username", req.Name).
            WithCause(err). // 保留原始错误链
            Build()
    }
    
    return &CreateUserResponse{
        User: user,
    }, nil
}
```

## 下一步

- 阅读 [错误处理使用指南](./errors-usage.md) 了解更多高级用法
- 查看 [API 错误码列表](./api/errors-code/apiserver/v1/errors_code.md) 了解所有可用错误码
- 学习 [验证使用指南](../../validation-usage.md) 了解请求验证

## 常见问题

**Q: 为什么我的错误码总是返回 50000？**

A: 这通常是因为错误码映射器未正确注册。确保在应用启动时导入了 `internal/apiserver` 包。

**Q: 如何自定义错误消息？**

A: 使用国际化机制，在 `locales.go` 中定义键，在配置文件中定义消息，然后使用 `i18n.FromContext(ctx).T(key)` 获取消息。

**Q: 客户端如何区分不同的错误？**

A: 使用 `reason` 字段进行判断，不要依赖 `message` 字段，因为消息可能因语言而异。