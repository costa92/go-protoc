package examples

import (
	"context"
	"fmt"
	"time"

	"github.com/costa92/go-protoc/v2/pkg/errors"
	"github.com/costa92/go-protoc/v2/pkg/errorsx"
)

// 使用示例

// ExampleBasicUsage 基本使用示例
func ExampleBasicUsage() {
	// 1. 创建基本错误
	err := errorsx.New(400, "INVALID_INPUT", "Invalid input parameter")
	fmt.Printf("Basic error: %v\n", err)
	
	// 2. 添加元数据
	err = err.AddMetadata("field", "email").AddMetadata("value", "invalid-email")
	fmt.Printf("Error with metadata: %v\n", err)
	
	// 3. 使用构建器
	err2 := errorsx.BadRequest("VALIDATION_FAILED").WithMessage("输入参数无效")
	fmt.Printf("Builder error: %v\n", err2)
}

// ExampleBusinessErrors 业务错误使用示例
func ExampleBusinessErrors() {
	// 1. 用户相关错误
	userErr := errors.NewUserNotFoundError("user123")
	fmt.Printf("User error: %v\n", userErr)
	
	// 2. 认证相关错误
	authErr := errors.NewTokenExpiredError("expired_token", time.Now().Add(-time.Hour))
	fmt.Printf("Auth error: %v\n", authErr)
	
	// 3. 通用错误
	commonErr := errors.NewResourceNotFoundError("order", "order123")
	fmt.Printf("Common error: %v\n", commonErr)
}

// ExampleErrorWrapping 错误包装示例
func ExampleErrorWrapping() {
	// 1. 包装标准错误
	originalErr := fmt.Errorf("database connection failed")
	wrappedErr := errorsx.Wrap(originalErr, 500, "DATABASE_ERROR", "Database operation failed")
	fmt.Printf("Wrapped error: %v\n", wrappedErr)
	fmt.Printf("Original cause: %v\n", wrappedErr.GetCause())
	
	// 2. 错误链
	err1 := errorsx.New(500, "EXTERNAL_API_ERROR", "External API failed")
	err2 := errorsx.Wrap(err1, 500, "SERVICE_ERROR", "Service operation failed")
	fmt.Printf("Error chain: %v\n", err2)
}

// ExampleI18nIntegration 国际化集成示例
func ExampleI18nIntegration() {
	ctx := context.Background()
	
	// 1. 创建带国际化键的错误
	err := errorsx.New(404, "USER_NOT_FOUND", "User not found").WithI18nKey("errors.user.not_found")
	
	// 2. 本地化错误
	localizedErr := errorsx.LocalizeError(ctx, err)
	fmt.Printf("Localized error: %v\n", localizedErr)
	
	// 3. 使用参数本地化
	params := map[string]any{
		"user_id": "123",
		"email":   "user@example.com",
	}
	localizedWithParams := errorsx.LocalizeErrorWithParams(ctx, err, params)
	fmt.Printf("Localized with params: %v\n", localizedWithParams)
}

// ExampleErrorRegistry 错误注册器示例
func ExampleErrorRegistry() {
	// 1. 注册错误模板
	errorsx.Register("CUSTOM_VALIDATION_ERROR", 400, "Custom validation failed")
	
	// 2. 使用模板创建错误
	err := errorsx.MustCreate("CUSTOM_VALIDATION_ERROR")
	fmt.Printf("Template error: %v\n", err)
	
	// 3. 检查模板是否存在
	if errorsx.TemplateExists("CUSTOM_VALIDATION_ERROR") {
		fmt.Println("Template exists")
	}
	
	// 4. 列出所有模板
	templates := errorsx.ListTemplates()
	fmt.Printf("Total templates: %d\n", len(templates))
}

// ExampleErrorHandling 错误处理示例
func ExampleErrorHandling() {
	// 1. 错误类型判断
	err := errors.ErrUserNotFound
	if errorsx.Is(err, errors.ErrUserNotFound) {
		fmt.Println("This is a user not found error")
	}
	
	// 2. 获取错误码
	code := errorsx.Code(err)
	fmt.Printf("Error code: %d\n", code)
	
	// 3. 获取错误原因
	reason := errorsx.Reason(err)
	fmt.Printf("Error reason: %s\n", reason)
	
	// 4. 从任意错误转换
	stdErr := fmt.Errorf("standard error")
	errorX := errorsx.FromError(stdErr)
	fmt.Printf("Converted error: %v\n", errorX)
}

// ExampleValidationErrors 验证错误示例
func ExampleValidationErrors() {
	// 1. 单个字段验证错误
	err := errors.NewUserValidationError("email", "invalid format")
	fmt.Printf("Validation error: %v\n", err)
	
	// 2. 多个字段验证错误
	validationErr := errorsx.New(400, "VALIDATION_ERROR", "Validation failed").WithI18nKey("errors.validation.multiple")
	validationErr = validationErr.AddMetadata("fields", map[string]string{
		"email":    "invalid format",
		"password": "too short",
		"age":      "must be positive",
	})
	fmt.Printf("Multiple validation errors: %v\n", validationErr)
	
	// 3. 参数范围错误
	rangeErr := errors.NewParameterOutOfRangeError("age", 18, 100, 15)
	fmt.Printf("Range error: %v\n", rangeErr)
}

// ExampleGRPCIntegration gRPC 集成示例
func ExampleGRPCIntegration() {
	// 1. 创建错误
	err := errorsx.New(404, "RESOURCE_NOT_FOUND", "Resource not found")
	err = err.AddMetadata("resource_id", "123")
	
	// 2. 转换为 gRPC 状态
	grpcStatus := err.GRPCStatus()
	fmt.Printf("gRPC status code: %v\n", grpcStatus.Code())
	fmt.Printf("gRPC message: %s\n", grpcStatus.Message())
	
	// 3. 从 gRPC 状态转换回错误
	convertedErr := errorsx.FromError(grpcStatus.Err())
	fmt.Printf("Converted back: %v\n", convertedErr)
}

// ExamplePerformanceOptimization 性能优化示例
func ExamplePerformanceOptimization() {
	// 1. 错误对象复用
	baseErr := errors.ErrResourceNotFound // 复用预定义错误
	customErr := baseErr.AddMetadata("resource_id", "123")
	fmt.Printf("Reused error: %v\n", customErr)
	
	// 2. 延迟国际化
	err := errorsx.New(400, "VALIDATION_ERROR", "Validation failed").WithI18nKey("errors.validation")
	// 只在需要时才进行国际化
	ctx := context.Background()
	localizedErr := errorsx.LocalizeError(ctx, err)
	fmt.Printf("Lazy localized: %v\n", localizedErr)
	
	// 3. 批量错误处理
	errs := []*errorsx.ErrorX{
		errorsx.New(400, "ERROR1", "Error 1"),
		errorsx.New(400, "ERROR2", "Error 2"),
		errorsx.New(400, "ERROR3", "Error 3"),
	}
	
	for _, err := range errs {
		localizedErr := errorsx.LocalizeError(ctx, err)
		fmt.Printf("Batch localized: %v\n", localizedErr)
	}
}