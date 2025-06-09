package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Error 定义了标准错误响应结构
type Error struct {
	Code    int         `json:"code"`              // 错误码
	Message string      `json:"message"`           // 错误信息
	Details interface{} `json:"details,omitempty"` // 错误详情
}

// 实现 error 接口
func (e *Error) Error() string {
	return fmt.Sprintf("错误码: %d, 信息: %s", e.Code, e.Message)
}

// NewError 创建一个新的错误实例
func NewError(code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// WithDetails 添加错误详情
func (e *Error) WithDetails(details interface{}) *Error {
	e.Details = details
	return e
}

// HTTPStatusCode 根据错误码获取对应的 HTTP 状态码
func (e *Error) HTTPStatusCode() int {
	switch e.Code / 100 {
	case 101:
		return http.StatusBadRequest // 10100-10199 参数验证错误
	case 201:
		return http.StatusNotFound // 20100-20199 用户模块错误
	case 202:
		return http.StatusUnauthorized // 20200-20299 认证模块错误
	case 203:
		return http.StatusForbidden // 20300-20399 授权模块错误
	case 300:
		return http.StatusBadGateway // 30000-30099 外部API错误
	case 400:
		return http.StatusInternalServerError // 40000-40099 数据库错误
	case 500:
		return http.StatusInternalServerError // 50000-50099 缓存错误
	default:
		return http.StatusInternalServerError
	}
}

// 预定义的系统错误
var (
	// 系统级错误
	ErrInternal = NewError(10000, "系统内部错误")
	ErrService  = NewError(10001, "服务暂时不可用")
	ErrTimeout  = NewError(10002, "请求超时")

	// 参数验证错误
	ErrValidation   = NewError(10100, "参数验证失败")
	ErrInvalidType  = NewError(10101, "参数类型错误")
	ErrMissingParam = NewError(10102, "必填参数缺失")

	// 用户模块错误
	ErrUserNotFound = NewError(20100, "用户未找到")
	ErrUserExists   = NewError(20101, "用户已存在")
	ErrInvalidUser  = NewError(20102, "用户名无效")

	// 认证模块错误
	ErrUnauthorized = NewError(20200, "未授权访问")
	ErrTokenExpired = NewError(20201, "访问令牌过期")
	ErrInvalidToken = NewError(20202, "无效的访问令牌")

	// 授权模块错误
	ErrPermissionDenied = NewError(20300, "权限不足")
	ErrRateLimit        = NewError(20301, "超出访问限制")

	// 第三方服务错误
	ErrThirdParty        = NewError(30000, "第三方服务异常")
	ErrThirdPartyTimeout = NewError(30001, "第三方服务超时")

	// 数据库错误
	ErrDBConnection = NewError(40000, "数据库连接错误")
	ErrDBQuery      = NewError(40001, "数据库查询错误")
	ErrDBNotFound   = NewError(40002, "数据不存在")

	// 缓存错误
	ErrCacheService  = NewError(50000, "缓存服务错误")
	ErrCacheNotFound = NewError(50001, "缓存键不存在")
)

// WriteJSON 将错误信息写入 HTTP 响应
func WriteJSON(w http.ResponseWriter, err *Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPStatusCode())
	json.NewEncoder(w).Encode(err)
}

// FromError 从普通错误转换为自定义错误
func FromError(err error) *Error {
	if err == nil {
		return nil
	}

	if e, ok := err.(*Error); ok {
		return e
	}

	return ErrInternal.WithDetails(err.Error())
}

// IsNotFound 判断是否为"未找到"类型的错误
func IsNotFound(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == ErrUserNotFound.Code || e.Code == ErrDBNotFound.Code || e.Code == ErrCacheNotFound.Code
	}
	return false
}

// IsUnauthorized 判断是否为未授权错误
func IsUnauthorized(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code >= 20200 && e.Code < 20300
	}
	return false
}

// IsValidationError 判断是否为验证错误
func IsValidationError(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code >= 10100 && e.Code < 10200
	}
	return false
}

// IsInternalError 判断是否为内部错误
func IsInternalError(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == ErrInternal.Code
	}
	return false
}
