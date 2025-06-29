package errorsx_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/costa92/go-protoc/v2/pkg/errorsx"
)

func TestBuilder_Basic(t *testing.T) {
	// 测试基本构建器功能
	builder := errorsx.NewBuilder(400, "VALIDATION_ERROR")
	err := builder.
		WithMessage("Validation failed").
		Build()
	
	assert.Equal(t, int32(400), err.Code)
	assert.Equal(t, "VALIDATION_ERROR", err.Reason)
	assert.Equal(t, "Validation failed", err.Message)
}

func TestBuilder_WithI18nKey(t *testing.T) {
	// 测试国际化键设置
	builder := errorsx.NewBuilder(404, "NOT_FOUND")
	err := builder.
		WithI18nKey("errors.resource.not_found").
		Build()
	
	assert.Equal(t, "errors.resource.not_found", err.GetI18nKey())
}

func TestBuilder_AddMetadata(t *testing.T) {
	// 测试元数据添加
	builder := errorsx.NewBuilder(400, "VALIDATION_ERROR")
	err := builder.
		WithMetadata(map[string]any{"field": "email", "value": "invalid"}).
		Build()
	
	assert.Equal(t, "email", err.Metadata["field"])
	assert.Equal(t, "invalid", err.Metadata["value"])
}

func TestBuilder_WithMetadata(t *testing.T) {
	// 测试批量元数据设置
	metadata := map[string]any{
		"field": "username",
		"type":  "required",
		"count": 5,
	}
	
	builder := errorsx.NewBuilder(400, "VALIDATION_ERROR")
	err := builder.
		WithMetadata(metadata).
		Build()
	
	assert.Equal(t, "username", err.Metadata["field"])
	assert.Equal(t, "required", err.Metadata["type"])
	assert.Equal(t, 5, err.Metadata["count"])
}

func TestBuilder_WithCause(t *testing.T) {
	// 测试原始错误设置
	originalErr := assert.AnError
	
	builder := errorsx.NewBuilder(500, "INTERNAL_ERROR")
	err := builder.
		WithCause(originalErr).
		Build()
	
	assert.Equal(t, originalErr, err.GetCause())
}

func TestBuilder_BuildWithContext(t *testing.T) {
	// 测试带上下文构建
	ctx := context.Background()
	
	builder := errorsx.NewBuilder(400, "VALIDATION_ERROR")
	err := builder.
		WithI18nKey("errors.validation.failed").
		BuildWithContext(ctx)
	
	assert.Equal(t, int32(400), err.Code)
	assert.Equal(t, "VALIDATION_ERROR", err.Reason)
	assert.Equal(t, "errors.validation.failed", err.GetI18nKey())
}

func TestBuilder_ChainedCalls(t *testing.T) {
	// 测试链式调用
	err := errorsx.NewBuilder(422, "UNPROCESSABLE_ENTITY").
		WithMessage("Entity validation failed").
		WithI18nKey("errors.entity.validation").
		WithMetadata(map[string]any{
			"entity": "user",
			"field": "email",
			"constraint": "unique",
		}).
		Build()
	
	assert.Equal(t, int32(422), err.Code)
	assert.Equal(t, "UNPROCESSABLE_ENTITY", err.Reason)
	assert.Equal(t, "Entity validation failed", err.Message)
	assert.Equal(t, "errors.entity.validation", err.GetI18nKey())
	assert.Equal(t, "user", err.Metadata["entity"])
	assert.Equal(t, "email", err.Metadata["field"])
	assert.Equal(t, "unique", err.Metadata["constraint"])
}

func TestBuilder_ConvenienceMethods(t *testing.T) {
	// 测试便捷方法
	tests := []struct {
		name    string
		builder func(reason string) *errorsx.Builder
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
			err := tt.builder("TEST_REASON").
				WithMessage("Test message").
				Build()
			
			assert.Equal(t, tt.expCode, err.Code)
			assert.Equal(t, "TEST_REASON", err.Reason)
			assert.Equal(t, "Test message", err.Message)
		})
	}
}

func TestBuilder_EmptyBuild(t *testing.T) {
	// 测试空构建器
	builder := errorsx.NewBuilder(500, "INTERNAL_ERROR")
	err := builder.Build()
	
	// 应该有默认值
	assert.Equal(t, int32(500), err.Code) // 默认内部错误
	assert.NotEmpty(t, err.Reason)        // 应该有默认原因
	assert.NotEmpty(t, err.Message)       // 应该有默认消息
}

func TestBuilder_OverwriteValues(t *testing.T) {
	// 测试值覆盖
	builder := errorsx.NewBuilder(400, "FIRST_REASON")
	err := builder.
		WithMessage("First message").
		WithMessage("Second message"). // 覆盖前面的值
		Build()
	
	assert.Equal(t, int32(400), err.Code)
	assert.Equal(t, "FIRST_REASON", err.Reason)
	assert.Equal(t, "Second message", err.Message)
}

func TestBuilder_MetadataOverwrite(t *testing.T) {
	// 测试元数据覆盖
	builder := errorsx.NewBuilder(400, "TEST_REASON")
	err := builder.
		WithMetadata(map[string]any{"key": "value1"}).
		WithMetadata(map[string]any{"key": "value2"}). // 覆盖前面的值
		Build()
	
	assert.Equal(t, "value2", err.Metadata["key"])
}

func TestBuilder_WithMetadataOverwrite(t *testing.T) {
	// 测试批量元数据覆盖
	builder := errorsx.NewBuilder(400, "TEST_REASON")
	err := builder.
		WithMetadata(map[string]any{
			"key1": "value1",
			"key2": "value2",
		}).
		WithMetadata(map[string]any{
			"key1": "new_value1", // 覆盖
			"key3": "value3",    // 新增
		}).
		Build()
	
	assert.Equal(t, "new_value1", err.Metadata["key1"])
	assert.Equal(t, "value2", err.Metadata["key2"]) // 保持不变
	assert.Equal(t, "value3", err.Metadata["key3"])
}