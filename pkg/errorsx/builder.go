package errorsx

import (
	"context"
	"net/http"

	"github.com/costa92/go-protoc/v2/pkg/i18n"
)

// Builder 错误构建器，用于链式构建错误对象
type Builder struct {
	code     int32
	reason   string
	i18nKey  string
	message  string
	metadata map[string]any
	cause    error
}

// NewBuilder 创建一个新的错误构建器
func NewBuilder(code int32, reason string) *Builder {
	return &Builder{
		code:     code,
		reason:   reason,
		metadata: make(map[string]any),
	}
}

// WithI18nKey 设置国际化键
func (b *Builder) WithI18nKey(key string) *Builder {
	b.i18nKey = key
	return b
}

// WithMessage 设置错误消息
func (b *Builder) WithMessage(message string) *Builder {
	b.message = message
	return b
}

// WithMetadata 添加元数据
func (b *Builder) WithMetadata(key string, value any) *Builder {
	if b.metadata == nil {
		b.metadata = make(map[string]any)
	}
	b.metadata[key] = value
	return b
}

// WithCause 设置原始错误
func (b *Builder) WithCause(err error) *Builder {
	b.cause = err
	return b
}

// Build 构建错误对象
func (b *Builder) Build() *ErrorX {
	return &ErrorX{
		Code:     b.code,
		Reason:   b.reason,
		Message:  b.message,
		Metadata: b.metadata,
		i18nKey:  b.i18nKey,
		cause:    b.cause,
	}
}

// BuildWithContext 使用上下文构建错误对象，自动进行国际化
func (b *Builder) BuildWithContext(ctx context.Context) *ErrorX {
	err := b.Build()
	
	// 如果设置了国际化键且消息为空，则进行国际化
	if b.i18nKey != "" && b.message == "" {
		if translator := i18n.FromContext(ctx); translator != nil {
			err.Message = translator.T(b.i18nKey)
		}
	}
	
	return err
}

// 便捷构造函数

// BadRequest 创建 400 错误构建器
func BadRequest(reason string) *Builder {
	return NewBuilder(http.StatusBadRequest, reason)
}

// Unauthorized 创建 401 错误构建器
func Unauthorized(reason string) *Builder {
	return NewBuilder(http.StatusUnauthorized, reason)
}

// Forbidden 创建 403 错误构建器
func Forbidden(reason string) *Builder {
	return NewBuilder(http.StatusForbidden, reason)
}

// NotFound 创建 404 错误构建器
func NotFound(reason string) *Builder {
	return NewBuilder(http.StatusNotFound, reason)
}

// Conflict 创建 409 错误构建器
func Conflict(reason string) *Builder {
	return NewBuilder(http.StatusConflict, reason)
}

// InternalError 创建 500 错误构建器
func InternalError(reason string) *Builder {
	return NewBuilder(http.StatusInternalServerError, reason)
}

// UnprocessableEntity 创建 422 错误构建器
func UnprocessableEntity(reason string) *Builder {
	return NewBuilder(http.StatusUnprocessableEntity, reason)
}

// TooManyRequests 创建 429 错误构建器
func TooManyRequests(reason string) *Builder {
	return NewBuilder(http.StatusTooManyRequests, reason)
}