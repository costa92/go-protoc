package errors

import (
	"github.com/costa92/go-protoc/v2/pkg/errorsx"
)

// 通用业务错误定义
var (
	// ErrResourceNotFound 资源不存在
	ErrResourceNotFound = errorsx.New(404, "RESOURCE_NOT_FOUND", "Resource not found").WithI18nKey("errors.common.resource_not_found")
	
	// ErrResourceAlreadyExists 资源已存在
	ErrResourceAlreadyExists = errorsx.New(409, "RESOURCE_ALREADY_EXISTS", "Resource already exists").WithI18nKey("errors.common.resource_already_exists")
	
	// ErrResourceConflict 资源冲突
	ErrResourceConflict = errorsx.New(409, "RESOURCE_CONFLICT", "Resource conflict").WithI18nKey("errors.common.resource_conflict")
	
	// ErrInvalidRequest 请求无效
	ErrInvalidRequest = errorsx.New(400, "INVALID_REQUEST", "Invalid request").WithI18nKey("errors.common.invalid_request")
	
	// ErrMissingParameter 缺少必需参数
	ErrMissingParameter = errorsx.New(400, "MISSING_PARAMETER", "Missing required parameter").WithI18nKey("errors.common.missing_parameter")
	
	// ErrInvalidParameter 参数无效
	ErrInvalidParameter = errorsx.New(400, "INVALID_PARAMETER", "Invalid parameter").WithI18nKey("errors.common.invalid_parameter")
	
	// ErrParameterOutOfRange 参数超出范围
	ErrParameterOutOfRange = errorsx.New(400, "PARAMETER_OUT_OF_RANGE", "Parameter out of range").WithI18nKey("errors.common.parameter_out_of_range")
	
	// ErrRateLimitExceeded 请求频率超限
	ErrRateLimitExceeded = errorsx.New(429, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded").WithI18nKey("errors.common.rate_limit_exceeded")
	
	// ErrServiceUnavailable 服务不可用
	ErrServiceUnavailable = errorsx.New(503, "SERVICE_UNAVAILABLE", "Service temporarily unavailable").WithI18nKey("errors.common.service_unavailable")
	
	// ErrDatabaseError 数据库错误
	ErrDatabaseError = errorsx.New(500, "DATABASE_ERROR", "Database operation failed").WithI18nKey("errors.common.database_error")
	
	// ErrExternalServiceError 外部服务错误
	ErrExternalServiceError = errorsx.New(502, "EXTERNAL_SERVICE_ERROR", "External service error").WithI18nKey("errors.common.external_service_error")
	
	// ErrConfigurationError 配置错误
	ErrConfigurationError = errorsx.New(500, "CONFIGURATION_ERROR", "Configuration error").WithI18nKey("errors.common.configuration_error")
	
	// ErrFileNotFound 文件不存在
	ErrFileNotFound = errorsx.New(404, "FILE_NOT_FOUND", "File not found").WithI18nKey("errors.common.file_not_found")
	
	// ErrFileUploadFailed 文件上传失败
	ErrFileUploadFailed = errorsx.New(400, "FILE_UPLOAD_FAILED", "File upload failed").WithI18nKey("errors.common.file_upload_failed")
	
	// ErrFileSizeExceeded 文件大小超限
	ErrFileSizeExceeded = errorsx.New(413, "FILE_SIZE_EXCEEDED", "File size exceeded").WithI18nKey("errors.common.file_size_exceeded")
	
	// ErrUnsupportedFileType 不支持的文件类型
	ErrUnsupportedFileType = errorsx.New(415, "UNSUPPORTED_FILE_TYPE", "Unsupported file type").WithI18nKey("errors.common.unsupported_file_type")
)

// 通用错误构建器函数

// NewResourceNotFoundError 创建资源不存在错误
func NewResourceNotFoundError(resourceType, resourceID string) *errorsx.ErrorX {
	return ErrResourceNotFound.
		AddMetadata("resource_type", resourceType).
		AddMetadata("resource_id", resourceID)
}

// NewResourceAlreadyExistsError 创建资源已存在错误
func NewResourceAlreadyExistsError(resourceType, resourceID string) *errorsx.ErrorX {
	return ErrResourceAlreadyExists.
		AddMetadata("resource_type", resourceType).
		AddMetadata("resource_id", resourceID)
}

// NewMissingParameterError 创建缺少参数错误
func NewMissingParameterError(paramName string) *errorsx.ErrorX {
	return ErrMissingParameter.AddMetadata("parameter", paramName)
}

// NewInvalidParameterError 创建无效参数错误
func NewInvalidParameterError(paramName, reason string) *errorsx.ErrorX {
	return ErrInvalidParameter.
		AddMetadata("parameter", paramName).
		AddMetadata("reason", reason)
}

// NewParameterOutOfRangeError 创建参数超出范围错误
func NewParameterOutOfRangeError(paramName string, min, max, actual any) *errorsx.ErrorX {
	return ErrParameterOutOfRange.
		AddMetadata("parameter", paramName).
		AddMetadata("min", min).
		AddMetadata("max", max).
		AddMetadata("actual", actual)
}

// NewRateLimitExceededError 创建请求频率超限错误
func NewRateLimitExceededError(limit int, window string, retryAfter int) *errorsx.ErrorX {
	return ErrRateLimitExceeded.
		AddMetadata("limit", limit).
		AddMetadata("window", window).
		AddMetadata("retry_after_seconds", retryAfter)
}

// NewDatabaseError 创建数据库错误
func NewDatabaseError(operation string, err error) *errorsx.ErrorX {
	return ErrDatabaseError.
		AddMetadata("operation", operation).
		WithCause(err)
}

// NewExternalServiceError 创建外部服务错误
func NewExternalServiceError(service string, err error) *errorsx.ErrorX {
	return ErrExternalServiceError.
		AddMetadata("service", service).
		WithCause(err)
}

// NewFileUploadError 创建文件上传错误
func NewFileUploadError(filename, reason string) *errorsx.ErrorX {
	return ErrFileUploadFailed.
		AddMetadata("filename", filename).
		AddMetadata("reason", reason)
}

// NewFileSizeExceededError 创建文件大小超限错误
func NewFileSizeExceededError(filename string, size, maxSize int64) *errorsx.ErrorX {
	return ErrFileSizeExceeded.
		AddMetadata("filename", filename).
		AddMetadata("size", size).
		AddMetadata("max_size", maxSize)
}

// NewUnsupportedFileTypeError 创建不支持的文件类型错误
func NewUnsupportedFileTypeError(filename, fileType string, supportedTypes []string) *errorsx.ErrorX {
	return ErrUnsupportedFileType.
		AddMetadata("filename", filename).
		AddMetadata("file_type", fileType).
		AddMetadata("supported_types", supportedTypes)
}