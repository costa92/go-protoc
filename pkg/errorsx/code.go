package errorsx

// 预定义的错误码和原因
const (
	// UnknownCode 表示未知的错误代码.
	UnknownCode int32 = 500

	// UnknownReason 表示未知的错误原因.
	UnknownReason = "UNKNOWN"
)

var (
	// OK 表示成功状态.
	OK = New(200, "OK", "Success")

	// ErrInternal 表示内部服务器错误.
	ErrInternal = New(500, "INTERNAL_ERROR", "Internal server error")

	// ErrNotFound 表示资源未找到错误.
	ErrNotFound = New(404, "NOT_FOUND", "Resource not found")

	// ErrBind 表示请求参数绑定错误.
	ErrBind = New(400, "BIND_ERROR", "Request parameter binding failed")

	// ErrInvalidArgument 表示无效参数错误.
	ErrInvalidArgument = New(400, "INVALID_ARGUMENT", "Invalid argument")

	// ErrUnauthenticated 表示未认证错误.
	ErrUnauthenticated = New(401, "UNAUTHENTICATED", "Authentication required")

	// ErrPermissionDenied 表示权限拒绝错误.
	ErrPermissionDenied = New(403, "PERMISSION_DENIED", "Permission denied")

	// ErrOperationFailed 表示操作失败错误.
	ErrOperationFailed = New(500, "OPERATION_FAILED", "Operation failed")

	// ErrValidation 表示验证错误.
	ErrValidation = New(422, "VALIDATION_ERROR", "Validation failed")

	// ErrConflict 表示资源冲突错误.
	ErrConflict = New(409, "CONFLICT", "Resource conflict")

	// ErrTooManyRequests 表示请求过多错误.
	ErrTooManyRequests = New(429, "TOO_MANY_REQUESTS", "Too many requests")

	// ErrServiceUnavailable 表示服务不可用错误.
	ErrServiceUnavailable = New(503, "SERVICE_UNAVAILABLE", "Service unavailable")

	// ErrTimeout 表示超时错误.
	ErrTimeout = New(408, "TIMEOUT", "Request timeout")
)
