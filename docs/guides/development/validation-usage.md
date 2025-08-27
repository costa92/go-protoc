# 接口验证方法调用文档

## 概述

本文档详细说明了项目中接口如何调用 `internal/apiserver/pkg/validation/user.go` 文件中的验证方法。

## 验证架构

### 1. 验证流程图

```
请求 → 中间件 → 通用验证器 → 具体验证方法 → 业务处理器
```

### 2. 核心组件

#### 2.1 验证中间件 (`validate_kratos.go`)

位置：`internal/pkg/middleware/validate/validate_kratos.go`

```go
// RequestValidator 定义了用于自定义验证的接口
type RequestValidator interface {
    Validate(ctx context.Context, rq any) error
}

// Validator 是一个验证中间件
func Validator(validator RequestValidator) middleware.Middleware {
    return func(handler middleware.Handler) middleware.Handler {
        return func(ctx context.Context, rq any) (reply any, err error) {
            // 自定义验证，特定于 API 接口
            if err := validator.Validate(ctx, rq); err != nil {
                // 错误处理逻辑
                return nil, errno.ErrorInvalidParameter("validation failed").WithCause(err)
            }
            return handler(ctx, rq)
        }
    }
}
```

#### 2.2 通用验证器 (`validation.go`)

位置：`pkg/validation/validation.go`

```go
type Validator struct {
    registry map[string]reflect.Value
}

// Validate 使用适当的验证方法验证请求
func (v *Validator) Validate(ctx context.Context, request any) error {
    validationFunc, ok := v.registry[reflect.TypeOf(request).Elem().Name()]
    if !ok {
        return nil // 未找到该请求类型的验证函数
    }

    result := validationFunc.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(request)})
    if !result[0].IsNil() {
        return result[0].Interface().(error)
    }

    return nil
}
```

#### 2.3 用户验证方法 (`user.go`)

位置：`internal/apiserver/pkg/validation/user.go`

```go
// ValidateCreateUserRequest 验证 CreateUserRequest 的字段
func (v *Validator) ValidateCreateUserRequest(ctx context.Context, rq *v1.CreateUserRequest) error {
    return i18n.FromContext(ctx).E(locales.UserAlreadyExists)
    // return nil
}

// ValidateUserRules 验证用户规则
func (v *Validator) ValidateUserRules() genericvalidation.Rules {
    return genericvalidation.Rules{}
}
```

## 调用流程详解

### 1. 依赖注入配置

在 `wire.go` 中配置依赖注入：

```go
func InitializeWebServer(cfg *config.Config) (*kratos.App, func(), error) {
    wire.Build(
        // ... 其他提供者
        validation.ProviderSet,  // 验证器提供者
        // ...
    )
    return nil, nil, nil
}
```

### 2. 中间件注册

在 `server.go` 中注册验证中间件：

```go
func NewMiddlewares(logger krtlog.Logger, val validate.RequestValidator) []middleware.Middleware {
    return []middleware.Middleware{
        validate.Validator(val),  // 注册验证中间件
        // ... 其他中间件
    }
}
```

### 3. 验证方法注册机制

通用验证器通过反射机制自动注册验证方法：

1. **方法命名约定**：验证方法必须以 `Validate` 开头，后跟请求类型名称
   - `ValidateCreateUserRequest` 对应 `CreateUserRequest`
   - `ValidateGetUserRequest` 对应 `GetUserRequest`

2. **方法签名要求**：
   ```go
   func (v *Validator) Validate{RequestType}(ctx context.Context, rq *{RequestType}) error
   ```

3. **自动注册过程**：
   ```go
   func extractValidationMethods(customValidator any) map[string]reflect.Value {
       funcs := make(map[string]reflect.Value)
       validatorType := reflect.TypeOf(customValidator)
       validatorValue := reflect.ValueOf(customValidator)
   
       for i := 0; i < validatorType.NumMethod(); i++ {
           method := validatorType.Method(i)
           // 检查方法名是否以 "Validate" 开头
           if !strings.HasPrefix(method.Name, "Validate") {
               continue
           }
           // 验证方法签名
           // 注册到 registry 中
           funcs[requestTypeName] = methodValue
       }
       return funcs
   }
   ```

### 4. 请求处理流程

以 `CreateUser` 接口为例：

1. **请求到达**：客户端发送 `CreateUserRequest`
2. **中间件拦截**：验证中间件拦截请求
3. **类型匹配**：通用验证器根据请求类型 `CreateUserRequest` 查找对应的验证方法
4. **方法调用**：调用 `ValidateCreateUserRequest` 方法
5. **验证执行**：执行具体的验证逻辑
6. **结果处理**：
   - 验证通过：继续执行业务逻辑
   - 验证失败：返回错误响应

## API 接口定义

### 请求结构体

```protobuf
message CreateUserRequest {
  string name = 1;
  string email = 2;
}

message GetUserRequest {
  string id = 1;
}
```

### 对应的验证方法

```go
// 对应 CreateUserRequest
func (v *Validator) ValidateCreateUserRequest(ctx context.Context, rq *v1.CreateUserRequest) error {
    // 验证用户创建请求
    // 可以验证 name、email 等字段
    return nil
}

// 对应 GetUserRequest (如果需要)
func (v *Validator) ValidateGetUserRequest(ctx context.Context, rq *v1.GetUserRequest) error {
    // 验证用户获取请求
    // 可以验证 id 字段
    return nil
}
```

## 扩展验证方法

### 1. 添加新的验证方法

要为新的请求类型添加验证，只需在 `user.go` 中添加相应的方法：

```go
// 为 UpdateUserRequest 添加验证
func (v *Validator) ValidateUpdateUserRequest(ctx context.Context, rq *v1.UpdateUserRequest) error {
    // 验证逻辑
    return nil
}
```

### 2. 使用通用验证函数

可以使用项目提供的通用验证函数：

```go
func (v *Validator) ValidateCreateUserRequest(ctx context.Context, rq *v1.CreateUserRequest) error {
    // 验证必需字段
    if err := genericvalidation.ValidRequired(rq, "Name", "Email"); err != nil {
        return err
    }
    
    // 其他自定义验证逻辑
    return nil
}
```

## 错误处理

验证失败时，可以返回国际化的错误信息：

```go
func (v *Validator) ValidateCreateUserRequest(ctx context.Context, rq *v1.CreateUserRequest) error {
    if rq.Name == "" {
        return i18n.FromContext(ctx).E(locales.InvalidParameter)
    }
    return nil
}
```

## 总结

1. **自动化**：验证方法通过反射机制自动注册，无需手动配置
2. **约定优于配置**：遵循命名约定即可自动生效
3. **类型安全**：编译时检查方法签名
4. **扩展性**：易于添加新的验证方法
5. **国际化**：支持多语言错误信息

这种设计使得验证逻辑与业务逻辑分离，提高了代码的可维护性和可测试性。