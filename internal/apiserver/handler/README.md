# Handler Layer 最佳实践指南

本文档详细说明了 Handler 层在项目清洁架构中的职责、设计原则和关联表业务处理的最佳实践。

## 目录

- [架构概览](#架构概览)
- [Handler层职责](#handler层职责)
- [基础用法](#基础用法)
- [关联表业务处理](#关联表业务处理)
- [错误处理](#错误处理)
- [性能优化](#性能优化)
- [最佳实践](#最佳实践)
- [常见问题](#常见问题)

## 架构概览

Handler 层作为项目清洁架构的**展示层(Presentation Layer)**，负责处理外部请求和响应：

```
┌─────────────────────────────────────┐
│           Handler Layer             │ ← gRPC/HTTP API接口层
├─────────────────────────────────────┤
│          Business Layer             │ ← 业务逻辑层 (Biz)
├─────────────────────────────────────┤
│          Storage Layer              │ ← 数据存储层 (Store)
└─────────────────────────────────────┘
```

**依赖流向**: Handler → Biz → Store → Database

## Handler层职责

### ✅ 应该做的事情

- **协议转换**: gRPC/HTTP 请求/响应的格式转换
- **简单委托**: 将请求直接委托给业务层处理
- **接口实现**: 实现 protobuf 生成的服务接口

### ❌ 不应该做的事情

- **业务逻辑处理**: 复杂的业务规则和计算
- **数据库操作**: 直接访问数据库或存储层
- **事务管理**: 跨表操作的事务控制
- **数据验证**: 复杂的业务规则验证

## 基础用法

### 标准Handler结构

```go
// handler.go
type Handler struct {
    v1.UnimplementedApiServerServer
    biz biz.IBiz  // 业务层依赖
}

// NewHandler 创建Handler实例
func NewHandler(biz biz.IBiz) *Handler {
    return &Handler{
        biz: biz,
    }
}
```

### 简单CRUD操作

```go
// ✅ 正确示例 - 保持简洁
func (h *Handler) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.GetUserResponse, error) {
    return h.biz.UserV1().Get(ctx, req)
}

func (h *Handler) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
    return h.biz.UserV1().CreateUser(ctx, req)
}

func (h *Handler) UpdateUser(ctx context.Context, req *v1.UpdateUserRequest) (*v1.UpdateUserResponse, error) {
    return h.biz.UserV1().Update(ctx, req)
}

func (h *Handler) DeleteUser(ctx context.Context, req *v1.DeleteUserRequest) (*v1.DeleteUserResponse, error) {
    return h.biz.UserV1().Delete(ctx, req)
}
```

## 关联表业务处理

### 一对一关联 (User -> UserProfile)

```go
// 获取用户及其配置信息
func (h *Handler) GetUserWithProfile(ctx context.Context, req *v1.GetUserWithProfileRequest) (*v1.GetUserWithProfileResponse, error) {
    return h.biz.UserV1().GetWithProfile(ctx, req)
}

// 创建用户及配置信息
func (h *Handler) CreateUserWithProfile(ctx context.Context, req *v1.CreateUserWithProfileRequest) (*v1.CreateUserWithProfileResponse, error) {
    return h.biz.UserV1().CreateWithProfile(ctx, req)
}
```

### 一对多关联 (User -> Orders)

```go
// 获取用户的所有订单
func (h *Handler) GetUserOrders(ctx context.Context, req *v1.GetUserOrdersRequest) (*v1.GetUserOrdersResponse, error) {
    return h.biz.UserV1().GetOrders(ctx, req)
}

// 分页查询用户订单
func (h *Handler) ListUserOrders(ctx context.Context, req *v1.ListUserOrdersRequest) (*v1.ListUserOrdersResponse, error) {
    return h.biz.UserV1().ListOrders(ctx, req)
}
```

### 多对多关联 (User <-> Roles)

```go
// 获取用户角色
func (h *Handler) GetUserRoles(ctx context.Context, req *v1.GetUserRolesRequest) (*v1.GetUserRolesResponse, error) {
    return h.biz.UserV1().GetRoles(ctx, req)
}

// 更新用户角色
func (h *Handler) UpdateUserRoles(ctx context.Context, req *v1.UpdateUserRolesRequest) (*v1.UpdateUserRolesResponse, error) {
    return h.biz.UserV1().UpdateRoles(ctx, req)
}

// 批量查询用户及其角色
func (h *Handler) ListUsersWithRoles(ctx context.Context, req *v1.ListUsersWithRolesRequest) (*v1.ListUsersWithRolesResponse, error) {
    return h.biz.UserV1().ListWithRoles(ctx, req)
}
```

### 复杂关联查询

```go
// 获取用户的完整信息（包含多种关联）
func (h *Handler) GetUserFullProfile(ctx context.Context, req *v1.GetUserFullProfileRequest) (*v1.GetUserFullProfileResponse, error) {
    return h.biz.UserV1().GetFullProfile(ctx, req)
}

// 用户统计信息（聚合查询）
func (h *Handler) GetUserStats(ctx context.Context, req *v1.GetUserStatsRequest) (*v1.GetUserStatsResponse, error) {
    return h.biz.UserV1().GetStats(ctx, req)
}
```

## 错误处理

### Handler层错误处理原则

**Handler层不处理具体错误，直接透传业务层错误**：

```go
// ✅ 正确 - 简单透传
func (h *Handler) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.GetUserResponse, error) {
    return h.biz.UserV1().Get(ctx, req)  // 直接返回业务层错误
}

// ❌ 错误 - Handler层处理错误
func (h *Handler) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.GetUserResponse, error) {
    resp, err := h.biz.UserV1().Get(ctx, req)
    if err != nil {
        // 不要在Handler层转换错误
        return nil, fmt.Errorf("get user failed: %v", err)
    }
    return resp, nil
}
```

### 统一错误处理

项目使用统一的 `errorsx` 包和中间件处理错误：

- **业务层**: 使用 `errorsx.New()` 创建标准错误
- **中间件**: 自动捕获和转换错误格式
- **Handler层**: 直接透传，无需额外处理

## 性能优化

### 避免在Handler层做优化

```go
// ❌ 错误 - Handler层处理性能优化
func (h *Handler) ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
    // 不要在Handler层做缓存
    if cached := h.cache.Get(req.String()); cached != nil {
        return cached.(*v1.ListUsersResponse), nil
    }

    resp, err := h.biz.UserV1().List(ctx, req)
    if err == nil {
        h.cache.Set(req.String(), resp)  // 错误！
    }
    return resp, err
}

// ✅ 正确 - 委托给业务层
func (h *Handler) ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
    return h.biz.UserV1().List(ctx, req)  // 业务层处理缓存
}
```

### 批量操作

```go
// 批量创建用户
func (h *Handler) BatchCreateUsers(ctx context.Context, req *v1.BatchCreateUsersRequest) (*v1.BatchCreateUsersResponse, error) {
    return h.biz.UserV1().BatchCreate(ctx, req)
}

// 批量更新用户状态
func (h *Handler) BatchUpdateUserStatus(ctx context.Context, req *v1.BatchUpdateUserStatusRequest) (*v1.BatchUpdateUserStatusResponse, error) {
    return h.biz.UserV1().BatchUpdateStatus(ctx, req)
}
```

## 最佳实践

### 1. 保持方法简洁

每个Handler方法应该是一行委托：

```go
// ✅ 理想的Handler方法
func (h *Handler) MethodName(ctx context.Context, req *v1.Request) (*v1.Response, error) {
    return h.biz.ServiceV1().Method(ctx, req)
}
```

### 2. 使用一致的命名

```go
// Handler方法名与业务方法名保持一致
func (h *Handler) GetUser(...) -> h.biz.UserV1().Get(...)
func (h *Handler) CreateUser(...) -> h.biz.UserV1().CreateUser(...)
func (h *Handler) ListUsers(...) -> h.biz.UserV1().List(...)
```

### 3. 按业务模块组织

```
handler/
├── user.go          # 用户相关的Handler
├── order.go         # 订单相关的Handler
├── product.go       # 产品相关的Handler
└── handler.go       # Handler结构体定义
```

### 4. 接口设计原则

```go
// ✅ 推荐 - 明确的业务语义
GetUserWithProfile    // 获取用户及配置
GetUserOrders        // 获取用户订单
UpdateUserRoles      // 更新用户角色

// ❌ 避免 - 模糊的通用接口
GetUserData          // 不明确获取什么数据
UpdateUser           // 不明确更新哪些字段
```

### 5. 参数验证

```go
// ❌ 错误 - Handler层验证业务规则
func (h *Handler) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
    if req.Email == "" {
        return nil, fmt.Errorf("email is required")  // 业务逻辑验证
    }
    return h.biz.UserV1().CreateUser(ctx, req)
}

// ✅ 正确 - 委托给业务层
func (h *Handler) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
    return h.biz.UserV1().CreateUser(ctx, req)  // 业务层处理验证
}
```

## 常见问题

### Q: Handler层可以处理简单的参数检查吗？

A: **不推荐**。即使是简单的非空检查也应该在业务层处理，保持Handler层的职责单一性。

### Q: 如何处理需要调用多个业务方法的场景？

A: 创建专门的业务方法来协调多个操作，而不是在Handler层调用多个业务方法：

```go
// ❌ 错误
func (h *Handler) ComplexOperation(ctx context.Context, req *v1.Request) (*v1.Response, error) {
    user, err := h.biz.UserV1().Get(ctx, userReq)
    if err != nil {
        return nil, err
    }

    orders, err := h.biz.OrderV1().List(ctx, orderReq)
    if err != nil {
        return nil, err
    }

    // 组装响应... 错误！
}

// ✅ 正确
func (h *Handler) ComplexOperation(ctx context.Context, req *v1.Request) (*v1.Response, error) {
    return h.biz.UserV1().GetComplexData(ctx, req)  // 业务层协调
}
```

### Q: 如何处理分页查询？

A: 分页逻辑在业务层处理，Handler层只做委托：

```go
func (h *Handler) ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
    return h.biz.UserV1().List(ctx, req)  // 业务层处理分页、筛选、排序
}
```

### Q: 可以在Handler层记录日志吗？

A: **不推荐**。项目已经有中间件自动记录请求日志，Handler层应该保持简洁。特殊情况下的业务日志应该在业务层记录。

## 代码示例

完整的用户Handler示例：

```go
// user.go
package handler

import (
    "context"

    "github.com/costa92/go-protoc/v2/internal/apiserver/biz"
    v1 "github.com/costa92/go-protoc/v2/pkg/api/apiserver/v1"
)

// 用户相关的Handler方法
// 所有方法都遵循简单委托原则

// 基础CRUD操作
func (h *Handler) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.GetUserResponse, error) {
    return h.biz.UserV1().Get(ctx, req)
}

func (h *Handler) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
    return h.biz.UserV1().CreateUser(ctx, req)
}

func (h *Handler) UpdateUser(ctx context.Context, req *v1.UpdateUserRequest) (*v1.UpdateUserResponse, error) {
    return h.biz.UserV1().Update(ctx, req)
}

func (h *Handler) DeleteUser(ctx context.Context, req *v1.DeleteUserRequest) (*v1.DeleteUserResponse, error) {
    return h.biz.UserV1().Delete(ctx, req)
}

func (h *Handler) ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
    return h.biz.UserV1().List(ctx, req)
}

// 关联表操作
func (h *Handler) GetUserWithProfile(ctx context.Context, req *v1.GetUserWithProfileRequest) (*v1.GetUserWithProfileResponse, error) {
    return h.biz.UserV1().GetWithProfile(ctx, req)
}

func (h *Handler) CreateUserWithProfile(ctx context.Context, req *v1.CreateUserWithProfileRequest) (*v1.CreateUserWithProfileResponse, error) {
    return h.biz.UserV1().CreateWithProfile(ctx, req)
}

func (h *Handler) UpdateUserRoles(ctx context.Context, req *v1.UpdateUserRolesRequest) (*v1.UpdateUserRolesResponse, error) {
    return h.biz.UserV1().UpdateRoles(ctx, req)
}

func (h *Handler) GetUserRoles(ctx context.Context, req *v1.GetUserRolesRequest) (*v1.GetUserRolesResponse, error) {
    return h.biz.UserV1().GetRoles(ctx, req)
}

func (h *Handler) ListUsersWithRoles(ctx context.Context, req *v1.ListUsersWithRolesRequest) (*v1.ListUsersWithRolesResponse, error) {
    return h.biz.UserV1().ListWithRoles(ctx, req)
}

// 批量操作
func (h *Handler) BatchCreateUsers(ctx context.Context, req *v1.BatchCreateUsersRequest) (*v1.BatchCreateUsersResponse, error) {
    return h.biz.UserV1().BatchCreate(ctx, req)
}

func (h *Handler) BatchUpdateUserStatus(ctx context.Context, req *v1.BatchUpdateUserStatusRequest) (*v1.BatchUpdateUserStatusResponse, error) {
    return h.biz.UserV1().BatchUpdateStatus(ctx, req)
}

// 复杂查询
func (h *Handler) GetUserStats(ctx context.Context, req *v1.GetUserStatsRequest) (*v1.GetUserStatsResponse, error) {
    return h.biz.UserV1().GetStats(ctx, req)
}

func (h *Handler) SearchUsers(ctx context.Context, req *v1.SearchUsersRequest) (*v1.SearchUsersResponse, error) {
    return h.biz.UserV1().Search(ctx, req)
}
```

---

**总结**: Handler 层应该保持极简，专注于协议转换和请求委托。复杂的业务逻辑、数据处理、关联查询都应该在业务层(Biz)处理。这样的设计保证了代码的可维护性、可测试性和扩展性。
