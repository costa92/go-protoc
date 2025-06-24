# 错误处理使用指南

## 概述

本文档详细说明了项目中错误处理机制的设计、使用方法和最佳实践。项目采用了基于 Kratos 框架的统一错误处理架构，支持国际化和多种错误码映射。

## 错误处理架构

### 1. 架构概览

```
业务逻辑 → 错误生成 → 错误码映射 → HTTP 响应编码 → 客户端
```

### 2. 核心组件

#### 2.1 错误定义层

- **通用错误** (`pkg/api/errno/errno.proto`): 定义系统级通用错误
- **业务错误** (`pkg/api/apiserver/v1/errors.proto`): 定义具体业务错误

#### 2.2 错误处理层

- **HTTP 编码器** (`pkg/server/http_codec.go`): 将 Kratos 错误转换为 HTTP 响应
- **错误码映射器** (`internal/apiserver/error_mapper.go`): 注册业务错误码映射

#### 2.3 国际化层

- **国际化支持** (`pkg/i18n/`): 提供多语言错误消息支持
- **本地化常量** (`internal/apiserver/pkg/locales/locales.go`): 定义国际化键值

## 错误码定义

### 1. 错误码结构

每个错误包含以下字段：

```json
{
  "code": 409,                    // HTTP 状态码
  "reason": "UserAlreadyExists",  // 错误原因标识
  "message": "用户已存在",         // 错误消息（支持国际化）
  "metadata": {}                  // 附加元数据
}
```

### 2. 错误码分类

#### 2.1 通用错误码 (errno)

| 错误原因 | HTTP 状态码 | 描述 |
|---------|------------|------|
| Unknown | 500 | 未知错误 |
| InvalidParameter | 400 | 参数无效 |
| Unauthorized | 401 | 未授权 |
| Forbidden | 403 | 禁止访问 |
| NotFound | 404 | 资源未找到 |
| InternalError | 500 | 内部服务器错误 |

#### 2.2 业务错误码 (apiserver/v1)

| 错误原因 | HTTP 状态码 | 描述 |
|---------|------------|------|
| UserLoginFailed | 401 | 用户登录失败 |
| UserAlreadyExists | 409 | 用户已存在 |
| UserNotFound | 404 | 用户未找到 |
| UserCreateFailed | 541 | 创建用户失败 |
| UserOperationForbidden | 403 | 用户操作被禁止 |
| SecretReachMaxCount | 400 | 密钥达到最大数量 |
| SecretNotFound | 404 | 密钥未找到 |
| SecretCreateFailed | 541 | 创建密钥失败 |

## 使用方法

### 1. 在业务逻辑中返回错误

#### 1.1 使用预定义错误

```go
package user

import (
    "context"
    v1 "github.com/costa92/go-protoc/v2/pkg/api/apiserver/v1"
    "github.com/costa92/go-protoc/v2/pkg/i18n"
    "github.com/costa92/go-protoc/v2/internal/apiserver/pkg/locales"
)

func (s *UserService) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
    // 检查用户是否已存在
    if userExists(req.Name) {
        // 获取国际化消息
        message := i18n.FromContext(ctx).T(locales.UserAlreadyExists)
        // 返回业务错误
        return nil, v1.ErrorUserAlreadyExists(message)
    }
    
    // 业务逻辑...
    return &v1.CreateUserResponse{}, nil
}
```

#### 1.2 使用通用错误

```go
import (
    "github.com/costa92/go-protoc/v2/pkg/api/errno"
)

func (s *Service) ValidateRequest(req *Request) error {
    if req.Name == "" {
        return errno.ErrorInvalidParameter("name is required")
    }
    return nil
}
```

### 2. 在验证层中使用错误

```go
package validation

import (
    "context"
    v1 "github.com/costa92/go-protoc/v2/pkg/api/apiserver/v1"
    "github.com/costa92/go-protoc/v2/pkg/i18n"
    "github.com/costa92/go-protoc/v2/internal/apiserver/pkg/locales"
)

func (v *UserValidator) ValidateCreateUserRequest(ctx context.Context, rq any) error {
    req, ok := rq.(*v1.CreateUserRequest)
    if !ok {
        return errno.ErrorInvalidParameter("invalid request type")
    }
    
    // 模拟用户已存在的情况
    if req.Name == "existing_user" {
        message := i18n.FromContext(ctx).T(locales.UserAlreadyExists)
        return v1.ErrorUserAlreadyExists(message)
    }
    
    return nil
}
```

### 3. 错误码映射配置

#### 3.1 注册业务错误码映射

在 `internal/apiserver/error_mapper.go` 中：

```go
package apiserver

import (
    "github.com/costa92/go-protoc/v2/pkg/api/apiserver/v1"
    "github.com/costa92/go-protoc/v2/pkg/server"
)

// init 函数在包导入时自动执行
func init() {
    // 注册 v1 包的错误码映射器
    mapper := server.NewHTTPStatusCodeMapper(v1.ErrorReason_value)
    server.RegisterErrorCodeMapper(mapper)
}
```

#### 3.2 确保映射器初始化

在应用启动文件中导入包：

```go
import (
    _ "github.com/costa92/go-protoc/v2/internal/apiserver" // 导入以执行 init 函数
)
```

## 国际化支持

### 1. 定义国际化键值

在 `internal/apiserver/pkg/locales/locales.go` 中：

```go
package locales

const (
    UserAlreadyExists = "user.already.exists"
    UserNotFound      = "user.not.found"
    // 更多常量...
)
```

### 2. 配置国际化消息

在国际化配置文件中定义消息：

```yaml
# zh-CN.yaml
user:
  already:
    exists: "用户已存在"
  not:
    found: "用户未找到"

# en-US.yaml
user:
  already:
    exists: "User already exists"
  not:
    found: "User not found"
```

### 3. 在代码中使用国际化

```go
// 获取国际化实例
i18nInstance := i18n.FromContext(ctx)

// 翻译消息
message := i18nInstance.T(locales.UserAlreadyExists)

// 返回错误
return v1.ErrorUserAlreadyExists(message)
```

## 客户端错误处理

### 1. 错误响应格式

客户端接收到的错误响应格式：

```json
{
  "code": 409,
  "reason": "UserAlreadyExists",
  "message": "用户已存在",
  "metadata": {}
}
```

### 2. 客户端处理示例

#### 2.1 JavaScript/TypeScript

```typescript
interface ApiError {
  code: number;
  reason: string;
  message: string;
  metadata: Record<string, any>;
}

async function createUser(userData: UserData) {
  try {
    const response = await fetch('/api/v1/users', {
      method: 'POST',
      body: JSON.stringify(userData),
    });
    
    if (!response.ok) {
      const error: ApiError = await response.json();
      
      switch (error.reason) {
        case 'UserAlreadyExists':
          throw new Error('用户已存在，请使用其他用户名');
        case 'InvalidParameter':
          throw new Error('请求参数无效');
        default:
          throw new Error(error.message || '未知错误');
      }
    }
    
    return await response.json();
  } catch (error) {
    console.error('创建用户失败:', error);
    throw error;
  }
}
```

#### 2.2 Go 客户端

```go
type ApiError struct {
    Code     int                    `json:"code"`
    Reason   string                 `json:"reason"`
    Message  string                 `json:"message"`
    Metadata map[string]interface{} `json:"metadata"`
}

func (c *Client) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    resp, err := c.httpClient.Post("/api/v1/users", req)
    if err != nil {
        return nil, err
    }
    
    if resp.StatusCode != http.StatusOK {
        var apiErr ApiError
        if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
            return nil, fmt.Errorf("failed to decode error response: %w", err)
        }
        
        switch apiErr.Reason {
        case "UserAlreadyExists":
            return nil, fmt.Errorf("user already exists: %s", apiErr.Message)
        case "InvalidParameter":
            return nil, fmt.Errorf("invalid parameter: %s", apiErr.Message)
        default:
            return nil, fmt.Errorf("api error [%d]: %s", apiErr.Code, apiErr.Message)
        }
    }
    
    var user User
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return &user, nil
}
```

## 最佳实践

### 1. 错误定义原则

- **语义明确**: 错误原因应该清晰表达具体的错误情况
- **分层设计**: 区分通用错误和业务错误
- **国际化友好**: 所有用户可见的错误消息都应支持国际化

### 2. 错误处理原则

- **早期验证**: 在请求处理的早期阶段进行参数验证
- **统一格式**: 使用统一的错误响应格式
- **详细日志**: 记录详细的错误信息用于调试

### 3. 性能考虑

- **避免频繁创建**: 复用错误实例，避免频繁创建新的错误对象
- **延迟国际化**: 只在需要时进行国际化翻译
- **缓存映射**: 错误码映射结果可以适当缓存

### 4. 安全考虑

- **信息泄露**: 避免在错误消息中暴露敏感信息
- **错误码一致性**: 确保相同的错误情况返回一致的错误码
- **日志记录**: 记录错误的上下文信息用于安全审计

## 故障排查

### 1. 常见问题

#### 1.1 错误码被默认值覆盖

**问题**: 业务错误返回默认错误码 50000 而不是预期的错误码

**原因**: 错误码映射器未正确注册

**解决方案**:
1. 确保 `error_mapper.go` 中的 `init` 函数被执行
2. 在应用启动文件中添加包导入
3. 检查错误码映射器的注册逻辑

#### 1.2 国际化消息未生效

**问题**: 错误消息始终显示为英文或默认语言

**原因**: 国际化上下文未正确设置或配置文件缺失

**解决方案**:
1. 检查请求头中的语言设置
2. 确认国际化中间件已正确配置
3. 验证国际化配置文件是否存在且格式正确

### 2. 调试技巧

#### 2.1 启用详细日志

```go
// 在错误处理中添加详细日志
log.WithFields(log.Fields{
    "error_reason": err.Reason,
    "error_code":   err.Code,
    "request_id":   requestID,
}).Error("API error occurred")
```

#### 2.2 错误码映射检查

```go
// 检查错误码是否正确映射
if code, ok := v1.ErrorReason_value["UserAlreadyExists"]; ok {
    log.Infof("UserAlreadyExists mapped to code: %d", code)
} else {
    log.Error("UserAlreadyExists not found in mapping")
}
```

## 参考资料

- [Kratos 错误处理文档](https://go-kratos.dev/docs/component/errors)
- [Protocol Buffers 错误定义](https://developers.google.com/protocol-buffers/docs/proto3)
- [HTTP 状态码规范](https://tools.ietf.org/html/rfc7231#section-6)
- [国际化最佳实践](https://github.com/nicksnyder/go-i18n)

## 更新日志

- **v1.0.0** (2024-01-XX): 初始版本，包含基础错误处理机制
- **v1.1.0** (2024-01-XX): 添加国际化支持和错误码映射机制
- **v1.2.0** (2024-01-XX): 优化错误处理架构，支持可扩展的错误码映射