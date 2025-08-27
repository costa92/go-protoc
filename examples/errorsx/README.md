# ErrorsX Examples

展示项目统一错误处理系统 `pkg/errorsx` 的使用示例和最佳实践。

## 📁 文件结构

```
examples/errorsx/
├── README.md          # 本文档
└── examples.go        # 错误处理示例
```

## 🏗️ ErrorsX 系统特性

- **多语言支持**: 基于 context 的国际化错误信息
- **结构化错误**: 统一的错误码和错误分类
- **链式操作**: 支持错误包装和上下文信息添加
- **gRPC 集成**: 自动转换为标准 gRPC 错误码
- **日志集成**: 自动记录错误详情和调用链

## 🚀 快速开始

```bash
cd examples/errorsx
go run examples.go
```

## 📊 示例内容

### 1. 基础错误创建和处理

```go
// 创建基础错误
err := errorsx.New("USER_NOT_FOUND", "用户未找到")

// 创建带参数的错误  
err := errorsx.Newf("INVALID_INPUT", "无效输入: %s", input)

// 错误包装
err := errorsx.Wrap(originalErr, "OPERATION_FAILED", "操作失败")
```

### 2. 多语言错误信息

```go
// 设置语言上下文
ctx := context.WithValue(context.Background(), "language", "en")

// 错误会自动根据上下文返回对应语言
err := errorsx.NewWithContext(ctx, "USER_NOT_FOUND")
// 中文: "用户未找到"
// 英文: "User not found"
```

### 3. gRPC 错误转换

```go
// 自动转换为 gRPC 状态码
grpcErr := errorsx.ToGRPCError(err)

// 从 gRPC 错误恢复
originalErr := errorsx.FromGRPCError(grpcErr)
```

### 4. 错误分类和处理

```go
// 检查错误类型
if errorsx.IsNotFound(err) {
    // 处理未找到错误
}

if errorsx.IsValidation(err) {
    // 处理验证错误
}

// 获取错误详情
details := errorsx.GetDetails(err)
fmt.Printf("Code: %s, Message: %s", details.Code, details.Message)
```

## 🔧 集成示例

### HTTP 处理器中的使用

```go
func (h *UserHandler) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.GetUserResponse, error) {
    user, err := h.userService.GetUser(ctx, req.UserId)
    if err != nil {
        // 自动处理和转换错误
        return nil, errorsx.Wrap(err, "GET_USER_FAILED", "获取用户失败")
    }
    
    return &v1.GetUserResponse{User: user}, nil
}
```

### 业务逻辑中的使用

```go  
func (s *UserService) CreateUser(ctx context.Context, user *User) error {
    // 验证输入
    if user.Email == "" {
        return errorsx.NewValidation("INVALID_EMAIL", "邮箱不能为空")
    }
    
    // 检查重复
    exists, err := s.userRepo.ExistsByEmail(ctx, user.Email)
    if err != nil {
        return errorsx.Wrap(err, "CHECK_USER_EXISTS_FAILED", "检查用户存在性失败")
    }
    
    if exists {
        return errorsx.NewConflict("USER_EXISTS", "用户已存在")
    }
    
    // 创建用户
    if err := s.userRepo.Create(ctx, user); err != nil {
        return errorsx.Wrap(err, "CREATE_USER_FAILED", "创建用户失败")
    }
    
    return nil
}
```

## 📈 错误码规范

### 命名约定

- **格式**: `CATEGORY_SPECIFIC_ERROR`
- **示例**: 
  - `USER_NOT_FOUND` - 用户未找到
  - `INVALID_INPUT_FORMAT` - 输入格式无效
  - `DATABASE_CONNECTION_FAILED` - 数据库连接失败

### 分类体系

| 分类 | 前缀 | gRPC 状态码 | HTTP 状态码 |
|------|------|-------------|-------------|
| 验证错误 | `INVALID_*` | InvalidArgument | 400 |
| 未找到 | `*_NOT_FOUND` | NotFound | 404 |
| 冲突 | `*_EXISTS` | AlreadyExists | 409 |
| 权限 | `ACCESS_*` | PermissionDenied | 403 |
| 系统错误 | `SYSTEM_*` | Internal | 500 |

## 🛠️ 最佳实践

### 1. 错误创建

```go
// ✅ 推荐: 语义化错误码
err := errorsx.New("USER_EMAIL_INVALID", "用户邮箱格式无效")

// ❌ 不推荐: 通用错误码  
err := errorsx.New("ERROR", "错误")
```

### 2. 错误包装

```go
// ✅ 推荐: 保留原始错误信息
if err != nil {
    return errorsx.Wrap(err, "OPERATION_FAILED", "操作失败")
}

// ❌ 不推荐: 丢失原始错误
if err != nil {
    return errorsx.New("OPERATION_FAILED", "操作失败")
}
```

### 3. 上下文传递

```go
// ✅ 推荐: 使用带上下文的错误创建
func HandleRequest(ctx context.Context) error {
    return errorsx.NewWithContext(ctx, "REQUEST_INVALID")
}

// ✅ 推荐: 传递语言信息
ctx = context.WithValue(ctx, "language", "en")
```

### 4. 错误检查

```go
// ✅ 推荐: 使用类型检查方法
if errorsx.IsNotFound(err) {
    return handleNotFound()
}

// ❌ 不推荐: 字符串匹配
if strings.Contains(err.Error(), "not found") {
    return handleNotFound()
}
```

## 🔍 调试和日志

### 错误详情输出

```go
err := errorsx.New("USER_CREATE_FAILED", "创建用户失败")
errorsx.AddDetail(err, "user_id", "12345")
errorsx.AddDetail(err, "email", "test@example.com")

// 输出结构化错误信息
logger.Errorw("Operation failed", "error", errorsx.ToMap(err))
```

### 调用链追踪

```go
// ErrorsX 自动记录错误调用链
err1 := doStep1() // line 10
err2 := errorsx.Wrap(err1, "STEP2_FAILED", "步骤2失败") // line 15  
err3 := errorsx.Wrap(err2, "OPERATION_FAILED", "操作失败") // line 20

// 获取完整调用链
chain := errorsx.GetCallChain(err3)
// 输出: [file.go:10, file.go:15, file.go:20]
```

## 📚 相关文档

- [ErrorsX 包文档](../../pkg/errorsx/README.md)
- [国际化配置](../../pkg/i18n/README.md)  
- [日志集成](../../pkg/logger/README.md)