// Package errors 提供业务相关的预定义错误类型和构建器函数。
//
// 本包基于 errorsx 包构建，采用两层架构设计：
//   - errorsx: 错误处理引擎（框架层）- 提供 ErrorX 结构体、中间件、国际化等核心功能
//   - errors: 业务错误目录（应用层）- 提供领域特定的错误定义和构建器函数
//
// 设计原则：
//   1. 框架与业务分离：errorsx 专注机制，errors 专注语义
//   2. 类型安全：使用预定义错误避免魔法字符串
//   3. 上下文丰富：通过构建器函数添加元数据和上下文信息
//   4. 国际化支持：所有错误都包含 i18n 键
//
// 快速开始：
//
//	import "github.com/costa92/go-protoc/v2/pkg/errors"
//
//	// 使用预定义错误
//	return errors.ErrUserNotFound
//
//	// 使用构建器函数添加上下文
//	return errors.NewUserNotFoundError("user123")
//
//	// 创建带有详细信息的错误
//	return errors.NewUserPermissionDeniedError("user123", "orders", "delete")
//
// 错误分类：
//   - 用户相关（user.go）: 用户认证、授权、CRUD操作相关错误
//   - 认证相关（auth.go）: JWT、会话、双因子认证等错误
//   - 通用业务（common.go）: 资源操作、参数验证、限流等错误
//   - 矿工相关（errors.go）: 特定业务领域错误（支持向后兼容）
//
// 最佳实践：
//   1. 优先使用预定义错误而非直接创建 errorsx.New()
//   2. 使用构建器函数添加特定上下文信息
//   3. 在业务层使用，Handler层直接透传
//   4. 错误信息要对用户友好，避免暴露内部实现细节
//
// 示例：
//
//	// ✅ 推荐：使用预定义错误和构建器
//	if user == nil {
//	    return errors.NewUserNotFoundError(userID)
//	}
//
//	// ✅ 推荐：添加丰富的上下文信息
//	return errors.NewUserPermissionDeniedError(userID, "orders", "delete")
//
//	// ❌ 不推荐：直接使用 errorsx（应该在 errors 包中预定义）
//	return errorsx.New(404, "USER_NOT_FOUND", "User not found")
//
// 性能注意事项：
//   - 错误创建使用对象池优化，性能开销很小
//   - 构建器函数会复制基础错误实例，添加特定元数据
//   - 高频错误会自动采样，避免日志洪流
package errors
