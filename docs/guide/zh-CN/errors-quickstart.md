# 错误处理快速开始

## 5分钟快速上手

本指南将帮助你在 5 分钟内掌握项目中错误处理的基本用法。

## 第一步：了解错误响应格式

当 API 调用失败时，你会收到如下格式的错误响应：

```json
{
  "code": 409,
  "reason": "UserAlreadyExists",
  "message": "用户已存在",
  "metadata": {}
}
```

- `code`: HTTP 状态码
- `reason`: 错误原因标识（用于程序判断）
- `message`: 人类可读的错误消息（支持国际化）
- `metadata`: 附加信息（可选）

## 第二步：在业务代码中返回错误

### 返回业务错误

```go
package user

import (
    v1 "github.com/costa92/go-protoc/v2/pkg/api/apiserver/v1"
    "github.com/costa92/go-protoc/v2/pkg/i18n"
    "github.com/costa92/go-protoc/v2/internal/apiserver/pkg/locales"
)

func CreateUser(ctx context.Context, req *v1.CreateUserRequest) error {
    // 检查用户是否已存在
    if userExists(req.Name) {
        // 获取国际化消息
        message := i18n.FromContext(ctx).T(locales.UserAlreadyExists)
        // 返回错误（HTTP 409）
        return v1.ErrorUserAlreadyExists(message)
    }
    return nil
}
```

### 返回通用错误

```go
import "github.com/costa92/go-protoc/v2/pkg/api/errno"

func ValidateInput(input string) error {
    if input == "" {
        // 返回参数错误（HTTP 400）
        return errno.ErrorInvalidParameter("输入不能为空")
    }
    return nil
}
```

## 第三步：客户端处理错误

### JavaScript 示例

```javascript
async function createUser(userData) {
  try {
    const response = await fetch('/api/v1/users', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(userData)
    });
    
    if (!response.ok) {
      const error = await response.json();
      
      // 根据错误原因进行不同处理
      switch (error.reason) {
        case 'UserAlreadyExists':
          alert('用户名已被使用，请选择其他用户名');
          break;
        case 'InvalidParameter':
          alert('请检查输入参数');
          break;
        default:
          alert(error.message || '操作失败');
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
func handleApiError(resp *http.Response) error {
    var apiError struct {
        Code    int    `json:"code"`
        Reason  string `json:"reason"`
        Message string `json:"message"`
    }
    
    json.NewDecoder(resp.Body).Decode(&apiError)
    
    switch apiError.Reason {
    case "UserAlreadyExists":
        return fmt.Errorf("用户已存在: %s", apiError.Message)
    case "InvalidParameter":
        return fmt.Errorf("参数错误: %s", apiError.Message)
    default:
        return fmt.Errorf("API错误 [%d]: %s", apiError.Code, apiError.Message)
    }
}
```

## 第四步：常用错误码速查

| 错误原因 | HTTP码 | 使用场景 | 示例代码 |
|---------|--------|----------|----------|
| `InvalidParameter` | 400 | 参数验证失败 | `errno.ErrorInvalidParameter("参数无效")` |
| `Unauthorized` | 401 | 未登录/认证失败 | `errno.ErrorUnauthorized("请先登录")` |
| `Forbidden` | 403 | 权限不足 | `errno.ErrorForbidden("权限不足")` |
| `NotFound` | 404 | 资源不存在 | `errno.ErrorNotFound("资源未找到")` |
| `UserAlreadyExists` | 409 | 用户已存在 | `v1.ErrorUserAlreadyExists(message)` |
| `InternalError` | 500 | 服务器内部错误 | `errno.ErrorInternalError("服务器错误")` |

## 第五步：添加国际化支持

### 1. 定义国际化键

在 `locales.go` 中添加：

```go
const (
    MyCustomError = "my.custom.error"
)
```

### 2. 配置多语言消息

```yaml
# zh-CN.yaml
my:
  custom:
    error: "自定义错误消息"

# en-US.yaml  
my:
  custom:
    error: "Custom error message"
```

### 3. 在代码中使用

```go
message := i18n.FromContext(ctx).T(locales.MyCustomError)
return v1.ErrorCustomError(message)
```

## 完整示例

这是一个完整的用户创建接口示例：

```go
package handler

import (
    "context"
    v1 "github.com/costa92/go-protoc/v2/pkg/api/apiserver/v1"
    "github.com/costa92/go-protoc/v2/pkg/api/errno"
    "github.com/costa92/go-protoc/v2/pkg/i18n"
    "github.com/costa92/go-protoc/v2/internal/apiserver/pkg/locales"
)

func (h *UserHandler) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
    // 1. 参数验证
    if req.Name == "" {
        return nil, errno.ErrorInvalidParameter("用户名不能为空")
    }
    
    if len(req.Name) < 3 {
        return nil, errno.ErrorInvalidParameter("用户名长度至少3个字符")
    }
    
    // 2. 业务逻辑检查
    if h.userService.UserExists(ctx, req.Name) {
        message := i18n.FromContext(ctx).T(locales.UserAlreadyExists)
        return nil, v1.ErrorUserAlreadyExists(message)
    }
    
    // 3. 创建用户
    user, err := h.userService.CreateUser(ctx, req)
    if err != nil {
        // 记录详细错误日志
        h.logger.Errorf("创建用户失败: %v", err)
        // 返回通用错误给客户端
        return nil, errno.ErrorInternalError("创建用户失败")
    }
    
    return &v1.CreateUserResponse{
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