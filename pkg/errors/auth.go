package errors

import (
	"time"
	
	"github.com/costa92/go-protoc/v2/pkg/errorsx"
)

// 认证相关错误定义
var (
	// ErrTokenInvalid JWT Token 无效
	ErrTokenInvalid = errorsx.New(401, "TOKEN_INVALID", "Invalid token").WithI18nKey("errors.auth.token_invalid")
	
	// ErrTokenExpired JWT Token 已过期
	ErrTokenExpired = errorsx.New(401, "TOKEN_EXPIRED", "Token has expired").WithI18nKey("errors.auth.token_expired")
	
	// ErrTokenMalformed JWT Token 格式错误
	ErrTokenMalformed = errorsx.New(401, "TOKEN_MALFORMED", "Malformed token").WithI18nKey("errors.auth.token_malformed")
	
	// ErrTokenMissing JWT Token 缺失
	ErrTokenMissing = errorsx.New(401, "TOKEN_MISSING", "Token is missing").WithI18nKey("errors.auth.token_missing")
	
	// ErrRefreshTokenInvalid 刷新令牌无效
	ErrRefreshTokenInvalid = errorsx.New(401, "REFRESH_TOKEN_INVALID", "Invalid refresh token").WithI18nKey("errors.auth.refresh_token_invalid")
	
	// ErrRefreshTokenExpired 刷新令牌已过期
	ErrRefreshTokenExpired = errorsx.New(401, "REFRESH_TOKEN_EXPIRED", "Refresh token has expired").WithI18nKey("errors.auth.refresh_token_expired")
	
	// ErrInsufficientPermissions 权限不足
	ErrInsufficientPermissions = errorsx.New(403, "INSUFFICIENT_PERMISSIONS", "Insufficient permissions").WithI18nKey("errors.auth.insufficient_permissions")
	
	// ErrAccountNotActivated 账户未激活
	ErrAccountNotActivated = errorsx.New(403, "ACCOUNT_NOT_ACTIVATED", "Account is not activated").WithI18nKey("errors.auth.account_not_activated")
	
	// ErrTwoFactorRequired 需要双因子认证
	ErrTwoFactorRequired = errorsx.New(403, "TWO_FACTOR_REQUIRED", "Two-factor authentication required").WithI18nKey("errors.auth.two_factor_required")
	
	// ErrTwoFactorInvalid 双因子认证码无效
	ErrTwoFactorInvalid = errorsx.New(401, "TWO_FACTOR_INVALID", "Invalid two-factor authentication code").WithI18nKey("errors.auth.two_factor_invalid")
	
	// ErrLoginAttemptExceeded 登录尝试次数超限
	ErrLoginAttemptExceeded = errorsx.New(429, "LOGIN_ATTEMPT_EXCEEDED", "Too many login attempts").WithI18nKey("errors.auth.login_attempt_exceeded")
	
	// ErrPasswordResetRequired 需要重置密码
	ErrPasswordResetRequired = errorsx.New(403, "PASSWORD_RESET_REQUIRED", "Password reset required").WithI18nKey("errors.auth.password_reset_required")
	
	// ErrLoginFailed 登录失败
	ErrLoginFailed = errorsx.New(401, "LOGIN_FAILED", "Login failed").WithI18nKey("errors.auth.login_failed")
	
	// ErrAccountLocked 账户被锁定
	ErrAccountLocked = errorsx.New(423, "ACCOUNT_LOCKED", "Account is locked").WithI18nKey("errors.auth.account_locked")
	
	// ErrSessionExpired 会话已过期
	ErrSessionExpired = errorsx.New(401, "SESSION_EXPIRED", "Session has expired").WithI18nKey("errors.auth.session_expired")
)

// 认证错误构建器函数

// NewTwoFactorRequiredError 创建双因子认证需求错误
func NewTwoFactorRequiredError(methods []string) *errorsx.ErrorX {
	return ErrTwoFactorRequired.AddMetadata("available_methods", methods)
}

// NewAccountNotActivatedError 创建账户未激活错误
func NewAccountNotActivatedError(userID, activationMethod string) *errorsx.ErrorX {
	return ErrAccountNotActivated.
		AddMetadata("user_id", userID).
		AddMetadata("activation_method", activationMethod)
}

// NewTokenInvalidError 创建 Token 无效错误（带token和reason参数）
func NewTokenInvalidError(token, reason string) *errorsx.ErrorX {
	return ErrTokenInvalid.
		AddMetadata("token", token).
		AddMetadata("reason", reason)
}

// NewTokenExpiredError 创建 Token 过期错误（带token和过期时间）
func NewTokenExpiredError(token string, expiredAt time.Time) *errorsx.ErrorX {
	return ErrTokenExpired.
		AddMetadata("token", token).
		AddMetadata("expired_at", expiredAt.Format(time.RFC3339))
}

// NewTokenMissingError 创建 Token 缺失错误
func NewTokenMissingError(header string) *errorsx.ErrorX {
	return ErrTokenMissing.AddMetadata("header", header)
}

// NewInsufficientPermissionsError 创建权限不足错误（带详细参数）
func NewInsufficientPermissionsError(userID, requiredPermission string, userPermissions []string) *errorsx.ErrorX {
	return ErrInsufficientPermissions.
		AddMetadata("user_id", userID).
		AddMetadata("required_permission", requiredPermission).
		AddMetadata("user_permissions", userPermissions)
}

// NewLoginFailedError 创建登录失败错误
func NewLoginFailedError(username, reason string, attemptCount int) *errorsx.ErrorX {
	return ErrLoginFailed.
		AddMetadata("username", username).
		AddMetadata("reason", reason).
		AddMetadata("attempt_count", attemptCount)
}

// NewAccountLockedError 创建账户锁定错误
func NewAccountLockedError(userID, lockReason string, unlockAt time.Time) *errorsx.ErrorX {
	return ErrAccountLocked.
		AddMetadata("user_id", userID).
		AddMetadata("lock_reason", lockReason).
		AddMetadata("unlock_at", unlockAt.Format(time.RFC3339))
}

// NewSessionExpiredError 创建会话过期错误
func NewSessionExpiredError(sessionID string, expiredAt time.Time) *errorsx.ErrorX {
	return ErrSessionExpired.
		AddMetadata("session_id", sessionID).
		AddMetadata("expired_at", expiredAt.Format(time.RFC3339))
}

// NewLoginAttemptExceededError 创建登录尝试超限错误（带详细参数）
func NewLoginAttemptExceededError(username string, maxAttempts, currentAttempts int, lockDuration time.Duration) *errorsx.ErrorX {
	return ErrLoginAttemptExceeded.
		AddMetadata("username", username).
		AddMetadata("max_attempts", maxAttempts).
		AddMetadata("current_attempts", currentAttempts).
		AddMetadata("lock_duration", lockDuration.String())
}