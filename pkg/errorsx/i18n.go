package errorsx

import (
	"context"

	"github.com/costa92/go-protoc/v2/pkg/i18n"
)

// I18nError 国际化错误包装器
type I18nError struct {
	*ErrorX
	i18n *i18n.I18n
}

// NewI18nError 创建国际化错误
func NewI18nError(err *ErrorX, i18n *i18n.I18n) *I18nError {
	return &I18nError{
		ErrorX: err,
		i18n:   i18n,
	}
}

// LocalizeMessage 本地化错误消息
func (e *I18nError) LocalizeMessage(ctx context.Context) string {
	if e.i18nKey != "" {
		return e.i18n.T(e.i18nKey)
	}
	return e.Message
}

// LocalizeWithParams 使用参数本地化错误消息
func (e *I18nError) LocalizeWithParams(ctx context.Context, params map[string]any) string {
	if e.i18nKey != "" {
		// 合并元数据和参数
		allParams := make(map[string]any)
		for k, v := range e.Metadata {
			allParams[k] = v
		}
		for k, v := range params {
			allParams[k] = v
		}
		
		// 使用 i18n 的模板功能（如果支持）
		return e.i18n.T(e.i18nKey)
	}
	return e.Message
}

// 全局国际化实例
var globalI18n *i18n.I18n

// SetGlobalI18n 设置全局国际化实例
func SetGlobalI18n(i18nInstance *i18n.I18n) {
	globalI18n = i18nInstance
}

// GetGlobalI18n 获取全局国际化实例
func GetGlobalI18n() *i18n.I18n {
	return globalI18n
}

// LocalizeError 本地化错误对象
func LocalizeError(ctx context.Context, err *ErrorX) *ErrorX {
	if err == nil {
		return nil
	}
	
	// 创建错误副本
	localizedErr := &ErrorX{
		Code:     err.Code,
		Reason:   err.Reason,
		Message:  err.Message,
		Metadata: err.Metadata,
		i18nKey:  err.i18nKey,
		cause:    err.cause,
	}
	
	// 如果有国际化键，进行本地化
	if err.i18nKey != "" {
		if translator := i18n.FromContext(ctx); translator != nil {
			localizedErr.Message = translator.T(err.i18nKey)
		} else if globalI18n != nil {
			localizedErr.Message = globalI18n.T(err.i18nKey)
		}
	}
	
	return localizedErr
}

// LocalizeErrorWithParams 使用参数本地化错误对象
func LocalizeErrorWithParams(ctx context.Context, err *ErrorX, params map[string]any) *ErrorX {
	if err == nil {
		return nil
	}
	
	// 创建错误副本
	localizedErr := &ErrorX{
		Code:     err.Code,
		Reason:   err.Reason,
		Message:  err.Message,
		Metadata: err.Metadata,
		i18nKey:  err.i18nKey,
		cause:    err.cause,
	}
	
	// 合并元数据和参数
	allParams := make(map[string]any)
	for k, v := range err.Metadata {
		allParams[k] = v
	}
	for k, v := range params {
		allParams[k] = v
	}
	localizedErr.Metadata = allParams
	
	// 如果有国际化键，进行本地化
	if err.i18nKey != "" {
		if translator := i18n.FromContext(ctx); translator != nil {
			localizedErr.Message = translator.T(err.i18nKey)
		} else if globalI18n != nil {
			localizedErr.Message = globalI18n.T(err.i18nKey)
		}
	}
	
	return localizedErr
}

// MustLocalizeError 本地化错误，如果失败则使用原始消息
func MustLocalizeError(ctx context.Context, err *ErrorX) *ErrorX {
	if err == nil {
		return nil
	}
	
	localizedErr := LocalizeError(ctx, err)
	
	// 如果本地化后消息为空，使用原始消息
	if localizedErr.Message == "" {
		localizedErr.Message = err.Message
	}
	
	// 如果仍然为空，使用默认消息
	if localizedErr.Message == "" {
		localizedErr.Message = "An error occurred"
	}
	
	return localizedErr
}