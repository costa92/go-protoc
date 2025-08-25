// Package errorsx 提供功能丰富且高性能的错误处理框架。
//
// errorsx 是项目错误处理架构的框架层，专注于提供错误处理的核心机制和基础设施。
// 它与 errors 包协作，形成完整的两层错误处理体系：
//   - errorsx (框架层): 提供 ErrorX 结构体、中间件、国际化、性能优化等核心功能
//   - errors (业务层): 基于 errorsx 构建业务特定的错误定义和构建器函数
//
// 核心特性：
//
// 1. 结构化错误信息：
//    - Code: HTTP状态码，用于API交互
//    - Reason: 业务错误码，用于精准定位和程序判断
//    - Message: 用户友好的错误消息
//    - Metadata: 附加的上下文信息和调试数据
//    - RequestID: 请求追踪标识
//    - i18nKey: 国际化键，支持多语言
//    - cause: 原始错误链，支持错误包装
//
// 2. 国际化支持：
//    - 内置 i18n 键管理
//    - 基于 Accept-Language 的自动语言检测
//    - 上下文感知的错误本地化
//    - 支持占位符和参数化消息
//
// 3. 性能优化：
//    - 对象池复用（StringBuilder、Metadata Map）
//    - 错误码字符串缓存，减少运行时格式化
//    - 智能日志采样，避免高频错误造成日志洪流
//    - 预分配缓冲区，减少内存分配
//
// 4. 协议支持：
//    - 自动 HTTP 状态码映射
//    - gRPC Status 和 ErrorDetails 转换
//    - 与 Kratos 框架深度集成
//    - 支持 OpenTelemetry 链路追踪
//
// 5. 中间件生态：
//    - Gin 错误处理和恢复中间件
//    - HTTP 标准库中间件支持
//    - 自动 panic 恢复和错误转换
//    - 统一的错误响应格式
//
// 基础用法：
//
//	// 创建简单错误
//	err := errorsx.New(404, "USER_NOT_FOUND", "User not found")
//
//	// 链式调用添加信息
//	err = err.WithI18nKey("errors.user.not_found").
//	    AddMetadata("user_id", "12345").
//	    WithRequestID(requestID)
//
//	// 包装现有错误
//	wrappedErr := errorsx.Wrap(dbErr, 500, "DATABASE_ERROR", "Database query failed")
//
// 高级用法：
//
//	// 构建富上下文错误
//	err := errorsx.New(422, "VALIDATION_ERROR", "Validation failed").
//	    WithI18nKey("errors.validation.failed").
//	    WithMetadata(map[string]any{
//	        "field": "email",
//	        "rule": "required",
//	        "value": userInput.Email,
//	    })
//
//	// 错误判断和类型转换
//	if errorsx.Code(err) == 404 {
//	    // 处理未找到错误
//	}
//
//	if reason := errorsx.Reason(err); reason == "USER_NOT_FOUND" {
//	    // 精确的业务错误判断
//	}
//
// 中间件使用：
//
//	// Gin 错误处理
//	errorHandler := errorsx.NewDefaultErrorHandler(logger)
//	router.Use(errorsx.GinErrorMiddleware(errorHandler))
//	router.Use(errorsx.RecoverMiddleware(errorHandler))
//
//	// HTTP 标准库
//	handler = errorsx.HTTPErrorMiddleware(errorHandler)(handler)
//
// 预定义错误：
//
// errorsx 提供了一套基础的预定义错误，适用于通用场景：
//   - ErrInternal: 内部服务器错误 (500)
//   - ErrNotFound: 资源未找到 (404)
//   - ErrBind: 请求参数绑定失败 (400)
//   - ErrInvalidArgument: 无效参数 (400)
//   - ErrUnauthenticated: 未认证 (401)
//   - ErrPermissionDenied: 权限拒绝 (403)
//   - ErrValidation: 验证失败 (422)
//   - ErrConflict: 资源冲突 (409)
//   - ErrTooManyRequests: 请求过多 (429)
//   - ErrServiceUnavailable: 服务不可用 (503)
//   - ErrTimeout: 请求超时 (408)
//
// 架构设计：
//
// errorsx 专注于提供错误处理的机制，而不关心具体的业务语义。
// 对于业务特定的错误，应该使用 errors 包中的预定义错误：
//
//	// ❌ 直接使用 errorsx（缺少业务语义）
//	return errorsx.New(404, "NOT_FOUND", "User not found")
//
//	// ✅ 使用 errors 包的业务错误（推荐）
//	return errors.NewUserNotFoundError(userID)
//
// 性能注意事项：
//   - ErrorX 实例设计为写时复制，链式调用安全且高效
//   - 高频错误会触发采样机制，避免日志系统过载
//   - 对象池会在合适的时机自动清理，无需手动管理
//   - 在热路径中复用 ErrorX 实例可以进一步提升性能
//
// 错误处理最佳实践：
//   1. 在框架层使用 errorsx，在业务层使用 errors
//   2. 始终为错误添加适当的上下文信息
//   3. 使用合适的 HTTP 状态码和业务错误码
//   4. 为面向用户的错误提供国际化支持
//   5. 在日志中记录完整的错误链和元数据
//
// 线程安全：
// ErrorX 的所有方法都是线程安全的，可以在并发环境中安全使用。
// 链式调用会创建新的实例，不会修改原始错误对象。
package errorsx // import "github.com/costa92/go-protoc/v2/pkg/errorsx"
