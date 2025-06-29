package errorsx_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"

	"github.com/costa92/go-protoc/v2/pkg/errorsx"
)

func TestErrorX_NewAndToString(t *testing.T) {
	// 创建一个 ErrorX 错误
	errx := errorsx.New(500, "InternalError.DBConnection", "Database connection failed: %s", "timeout")

	// 检查字段值
	assert.Equal(t, 500, errx.Code)
	assert.Equal(t, "InternalError.DBConnection", errx.Reason)
	assert.Equal(t, "Database connection failed: timeout", errx.Message)

	// 检查字符串表示
	expected := `error: code = 500 reason = InternalError.DBConnection message = Database connection failed: timeout metadata = map[]`
	assert.Equal(t, expected, errx.Error())
}

func TestErrorX_WithMessage(t *testing.T) {
	// 创建一个基础错误
	errx := errorsx.New(400, "BadRequest.InvalidInput", "Invalid input for field %s", "username")

	// 更新错误的消息
	errx.WithMessage("New error message: %s", "retry failed")

	// 验证变更
	assert.Equal(t, "New error message: retry failed", errx.Message)
	assert.Equal(t, 400, errx.Code)                         // Code 不变
	assert.Equal(t, "BadRequest.InvalidInput", errx.Reason) // Reason 不变
}

func TestErrorX_WithMetadata(t *testing.T) {
	// 创建基础错误
	errx := errorsx.New(400, "BadRequest.InvalidInput", "Invalid input")

	// 添加元数据
	errx.WithMetadata(map[string]string{
		"field": "username",
		"type":  "empty",
	})

	// 验证元数据
	assert.Equal(t, "username", errx.Metadata["field"])
	assert.Equal(t, "empty", errx.Metadata["type"])

	// 动态添加更多元数据
	errx.KV("user_id", "12345", "trace_id", "xyz-789")
	assert.Equal(t, "12345", errx.Metadata["user_id"])
	assert.Equal(t, "xyz-789", errx.Metadata["trace_id"])
}

func TestErrorX_Is(t *testing.T) {
	// 定义两个预定义错误
	err1 := errorsx.New(404, "NotFound.User", "User not found")
	err2 := errorsx.New(404, "NotFound.User", "Another message")
	err3 := errorsx.New(403, "Forbidden", "Access denied")

	// 验证两个错误均被认为是同一种类型的错误（Code 和 Reason 相等）
	assert.True(t, err1.Is(err2))  // Message 不影响匹配
	assert.False(t, err1.Is(err3)) // Reason 不同
}

func TestErrorX_FromError_WithPlainError(t *testing.T) {
	// 创建一个普通的 Go 错误
	plainErr := errors.New("Something went wrong")

	// 转换为 ErrorX
	errx := errorsx.FromError(plainErr)

	// 检查转换后的 ErrorX
	assert.Equal(t, errorsx.UnknownCode, errx.Code)       // 默认 500
	assert.Equal(t, errorsx.UnknownReason, errx.Reason)   // 默认 ""
	assert.Equal(t, "Something went wrong", errx.Message) // 转换时保留原始错误消息
}

func TestErrorX_FromError_WithGRPCError(t *testing.T) {
	// 创建一个 gRPC 错误
	grpcErr := status.New(3, "Invalid argument").Err() // gRPC INVALID_ARGUMENT = 3

	// 转换为 ErrorX
	errx := errorsx.FromError(grpcErr)

	// 检查转换后的 ErrorX
	assert.Equal(t, 400, errx.Code) // httpstatus.FromGRPCCode(3) 对应 HTTP 400
	assert.Equal(t, "Invalid argument", errx.Message)

	// 没有附加的元数据
	assert.Nil(t, errx.Metadata)
}

func TestErrorX_FromError_WithGRPCErrorDetails(t *testing.T) {
	// 创建带有详细信息的 gRPC 错误
	st := status.New(3, "Invalid argument")
	grpcErr, err := st.WithDetails(&errdetails.ErrorInfo{
		Reason:   "InvalidInput",
		Metadata: map[string]string{"field": "name", "type": "required"},
	})
	assert.NoError(t, err) // 确保 gRPC 错误创建成功

	// 转换为 ErrorX
	errx := errorsx.FromError(grpcErr.Err())

	// 检查转换后的 ErrorX
	assert.Equal(t, 400, errx.Code) // gRPC INVALID_ARGUMENT = HTTP 400
	assert.Equal(t, "Invalid argument", errx.Message)
	assert.Equal(t, "InvalidInput", errx.Reason) // 从 gRPC ErrorInfo 中提取

	// 检查元数据
	assert.Equal(t, "name", errx.Metadata["field"])
	assert.Equal(t, "required", errx.Metadata["type"])
}

func TestErrorX_WithI18nKey(t *testing.T) {
	// 测试国际化键设置
	err := errorsx.New(404, "NOT_FOUND", "Resource not found")
	err = err.WithI18nKey("errors.resource.not_found")
	
	assert.Equal(t, "errors.resource.not_found", err.GetI18nKey())
}

func TestErrorX_WithCause(t *testing.T) {
	// 测试原始错误设置
	originalErr := fmt.Errorf("database connection failed")
	err := errorsx.New(500, "DATABASE_ERROR", "Database operation failed")
	err = err.WithCause(originalErr)
	
	assert.Equal(t, originalErr, err.GetCause())
	
	// 测试 Unwrap 接口
	unwrapped := errors.Unwrap(err)
	assert.Equal(t, originalErr, unwrapped)
}

func TestErrorX_AddMetadata(t *testing.T) {
	// 测试添加元数据
	err := errorsx.New(400, "VALIDATION_ERROR", "Validation failed")
	err = err.AddMetadata("field", "email")
	err = err.AddMetadata("value", "invalid-email")
	
	assert.Equal(t, "email", err.Metadata["field"])
	assert.Equal(t, "invalid-email", err.Metadata["value"])
}

func TestCode(t *testing.T) {
	// 测试错误码提取
	err := errorsx.New(404, "NOT_FOUND", "Resource not found")
	code := errorsx.Code(err)
	assert.Equal(t, int32(404), code)
	
	// 测试 nil 错误
	code = errorsx.Code(nil)
	assert.Equal(t, int32(200), code)
	
	// 测试标准错误
	stdErr := fmt.Errorf("standard error")
	code = errorsx.Code(stdErr)
	assert.Equal(t, errorsx.UnknownCode, code)
}

func TestReason(t *testing.T) {
	// 测试错误原因提取
	err := errorsx.New(404, "NOT_FOUND", "Resource not found")
	reason := errorsx.Reason(err)
	assert.Equal(t, "NOT_FOUND", reason)
	
	// 测试 nil 错误
	reason = errorsx.Reason(nil)
	assert.Equal(t, "", reason)
	
	// 测试标准错误
	stdErr := fmt.Errorf("standard error")
	reason = errorsx.Reason(stdErr)
	assert.Equal(t, errorsx.UnknownReason, reason)
}

func TestWrap(t *testing.T) {
	// 测试错误包装
	originalErr := fmt.Errorf("database connection failed")
	wrappedErr := errorsx.Wrap(originalErr, 500, "DATABASE_ERROR", "Database operation failed")
	
	assert.Equal(t, int32(500), wrappedErr.Code)
	assert.Equal(t, "DATABASE_ERROR", wrappedErr.Reason)
	assert.Equal(t, originalErr, wrappedErr.GetCause())
	
	// 测试 nil 错误包装
	wrappedErr = errorsx.Wrap(nil, 500, "TEST", "Test")
	assert.Nil(t, wrappedErr)
}

func TestWrapf(t *testing.T) {
	// 测试格式化错误包装
	originalErr := fmt.Errorf("connection failed")
	wrappedErr := errorsx.Wrapf(originalErr, 500, "DATABASE_ERROR", "Database operation failed: %s", "timeout")
	
	expectedMessage := "Database operation failed: timeout"
	assert.Equal(t, expectedMessage, wrappedErr.Message)
	assert.Equal(t, originalErr, wrappedErr.GetCause())
}

func TestBuilder(t *testing.T) {
	// 测试错误构建器
	err := errorsx.BadRequest().
		WithReason("VALIDATION_FAILED").
		WithMessage("Validation failed").
		WithI18nKey("errors.validation.failed").
		AddMetadata("field", "email").
		Build()
	
	assert.Equal(t, int32(400), err.Code)
	assert.Equal(t, "VALIDATION_FAILED", err.Reason)
	assert.Equal(t, "Validation failed", err.Message)
	assert.Equal(t, "errors.validation.failed", err.GetI18nKey())
	assert.Equal(t, "email", err.Metadata["field"])
}

func TestBuilderWithContext(t *testing.T) {
	// 测试带上下文的构建器
	ctx := context.Background()
	
	err := errorsx.BadRequest().
		WithReason("VALIDATION_FAILED").
		WithI18nKey("errors.validation.failed").
		BuildWithContext(ctx)
	
	assert.Equal(t, int32(400), err.Code)
	assert.Equal(t, "errors.validation.failed", err.GetI18nKey())
}

func TestPredefinedErrors(t *testing.T) {
	// 测试预定义错误
	tests := []struct {
		name     string
		err      *errorsx.ErrorX
		expCode  int32
		expReason string
	}{
		{"OK", errorsx.OK, 200, "OK"},
		{"ErrInternal", errorsx.ErrInternal, 500, "INTERNAL_ERROR"},
		{"ErrNotFound", errorsx.ErrNotFound, 404, "NOT_FOUND"},
		{"ErrBind", errorsx.ErrBind, 400, "BIND_ERROR"},
		{"ErrInvalidArgument", errorsx.ErrInvalidArgument, 400, "INVALID_ARGUMENT"},
		{"ErrUnauthenticated", errorsx.ErrUnauthenticated, 401, "UNAUTHENTICATED"},
		{"ErrPermissionDenied", errorsx.ErrPermissionDenied, 403, "PERMISSION_DENIED"},
		{"ErrOperationFailed", errorsx.ErrOperationFailed, 500, "OPERATION_FAILED"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expCode, tt.err.Code)
			assert.Equal(t, tt.expReason, tt.err.Reason)
		})
	}
}

func TestErrorChaining(t *testing.T) {
	// 测试错误链
	originalErr := fmt.Errorf("original error")
	err1 := errorsx.Wrap(originalErr, 500, "LEVEL1", "Level 1 error")
	err2 := errorsx.Wrap(err1, 500, "LEVEL2", "Level 2 error")
	
	// 检查错误链
	assert.True(t, errors.Is(err2, originalErr))
	assert.True(t, errors.Is(err2, err1))
	
	// 检查直接原因
	assert.Equal(t, err1, err2.GetCause())
	assert.Equal(t, originalErr, err1.GetCause())
}

func TestBuilderMethods(t *testing.T) {
	// 测试各种构建器方法
	tests := []struct {
		name    string
		builder func() *errorsx.Builder
		expCode int32
	}{
		{"BadRequest", errorsx.BadRequest, 400},
		{"Unauthorized", errorsx.Unauthorized, 401},
		{"Forbidden", errorsx.Forbidden, 403},
		{"NotFound", errorsx.NotFound, 404},
		{"Conflict", errorsx.Conflict, 409},
		{"UnprocessableEntity", errorsx.UnprocessableEntity, 422},
		{"TooManyRequests", errorsx.TooManyRequests, 429},
		{"InternalError", errorsx.InternalError, 500},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.builder().WithMessage("Test message").Build()
			assert.Equal(t, tt.expCode, err.Code)
			assert.Equal(t, "Test message", err.Message)
		})
	}
}
