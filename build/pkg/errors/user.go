package errors

import (
	"github.com/costa92/go-protoc/v2/pkg/errorsx"
)

// 用户相关错误定义
var (
	// ErrUserNotFound 用户不存在
	ErrUserNotFound = errorsx.New(404, "USER_NOT_FOUND", "User not found").WithI18nKey("errors.user.not_found")

	// ErrUserAlreadyExists 用户已存在
	ErrUserAlreadyExists = errorsx.New(409, "USER_ALREADY_EXISTS", "User already exists").WithI18nKey("errors.user.already_exists")

	// ErrUserInvalidCredentials 用户凭据无效
	ErrUserInvalidCredentials = errorsx.New(401, "USER_INVALID_CREDENTIALS", "Invalid credentials").WithI18nKey("errors.user.invalid_credentials")

	// ErrUserAccountLocked 用户账户被锁定
	ErrUserAccountLocked = errorsx.New(423, "USER_ACCOUNT_LOCKED", "User account is locked").WithI18nKey("errors.user.account_locked")

	// ErrUserAccountDisabled 用户账户被禁用
	ErrUserAccountDisabled = errorsx.New(403, "USER_ACCOUNT_DISABLED", "User account is disabled").WithI18nKey("errors.user.account_disabled")

	// ErrUserPasswordExpired 用户密码已过期
	ErrUserPasswordExpired = errorsx.New(403, "USER_PASSWORD_EXPIRED", "User password has expired").WithI18nKey("errors.user.password_expired")

	// ErrUserInvalidEmail 用户邮箱格式无效
	ErrUserInvalidEmail = errorsx.New(400, "USER_INVALID_EMAIL", "Invalid email format").WithI18nKey("errors.user.invalid_email")

	// ErrUserInvalidPassword 用户密码格式无效
	ErrUserInvalidPassword = errorsx.New(400, "USER_INVALID_PASSWORD", "Invalid password format").WithI18nKey("errors.user.invalid_password")

	// ErrUserPermissionDenied 用户权限不足
	ErrUserPermissionDenied = errorsx.New(403, "USER_PERMISSION_DENIED", "Permission denied").WithI18nKey("errors.user.permission_denied")

	// ErrUserSessionExpired 用户会话已过期
	ErrUserSessionExpired = errorsx.New(401, "USER_SESSION_EXPIRED", "Session has expired").WithI18nKey("errors.user.session_expired")
	// ErrUserValidationFailed 用户验证失败
	ErrUserValidationFailed = errorsx.New(400, "USER_VALIDATION_FAILED", "User validation failed").WithI18nKey("errors.user.validation_failed")
	// ErrUserInactive 用户不活跃
	ErrUserInactive = errorsx.New(403, "USER_INACTIVE", "User account is inactive").WithI18nKey("errors.user.inactive")
	// ErrUserPasswordInvalid 用户密码无效
	ErrUserPasswordInvalid = errorsx.New(401, "USER_PASSWORD_INVALID", "User password invalid").WithI18nKey("errors.user.password_invalid")
	// ErrUserEmailInvalid 用户邮箱无效
	ErrUserEmailInvalid = errorsx.New(400, "USER_EMAIL_INVALID", "User email invalid").WithI18nKey("errors.user.email_invalid")
	// ErrUserOperationFailed 用户操作失败
	ErrUserOperationFailed = errorsx.New(500, "USER_OPERATION_FAILED", "User operation failed").WithI18nKey("errors.user.operation_failed")
)

// 用户错误构建器函数

// NewUserNotFoundError 创建用户不存在错误
func NewUserNotFoundError(userID string) *errorsx.ErrorX {
	return ErrUserNotFound.WithMessage("User with ID %s not found", userID).AddMetadata("user_id", userID)
}

// NewUserAlreadyExistsError 创建用户已存在错误
func NewUserAlreadyExistsError(email string) *errorsx.ErrorX {
	return ErrUserAlreadyExists.WithMessage("User with email %s already exists", email).AddMetadata("email", email)
}

// NewUserInvalidCredentialsError 创建用户凭据无效错误
func NewUserInvalidCredentialsError(email string) *errorsx.ErrorX {
	return ErrUserInvalidCredentials.AddMetadata("email", email)
}

// NewUserAccountLockedError 创建用户账户锁定错误
func NewUserAccountLockedError(userID string, lockUntil string) *errorsx.ErrorX {
	return ErrUserAccountLocked.AddMetadata("user_id", userID).AddMetadata("lock_until", lockUntil)
}

// NewUserPermissionDeniedError 创建用户权限不足错误
func NewUserPermissionDeniedError(userID, resource, action string) *errorsx.ErrorX {
	return ErrUserPermissionDenied.WithMessage("User %s denied %s access to %s", userID, action, resource).
		AddMetadata("user_id", userID).
		AddMetadata("resource", resource).
		AddMetadata("action", action)
}

// NewUserValidationError 创建用户验证错误
func NewUserValidationError(field, reason string) *errorsx.ErrorX {
	return errorsx.ErrValidation.
		WithI18nKey("errors.user.validation").
		AddMetadata("field", field).
		AddMetadata("reason", reason)
}

// NewUserInactiveError 创建用户不活跃错误
func NewUserInactiveError(userID, status string) *errorsx.ErrorX {
	return ErrUserInactive.WithMessage("User with ID %s is inactive (status: %s)", userID, status).
		AddMetadata("user_id", userID).
		AddMetadata("status", status)
}

// NewUserPasswordInvalidError 创建用户密码无效错误
func NewUserPasswordInvalidError(userID, reason string) *errorsx.ErrorX {
	return ErrUserPasswordInvalid.WithMessage("Password for user %s is invalid: %s", userID, reason).
		AddMetadata("user_id", userID).
		AddMetadata("reason", reason)
}

// NewUserEmailInvalidError 创建用户邮箱无效错误
func NewUserEmailInvalidError(email, reason string) *errorsx.ErrorX {
	return ErrUserEmailInvalid.WithMessage("Email %s is invalid: %s", email, reason).
		AddMetadata("email", email).
		AddMetadata("reason", reason)
}

// NewUserOperationFailedError 创建用户操作失败错误
func NewUserOperationFailedError(userID, operation, reason string) *errorsx.ErrorX {
	return ErrUserOperationFailed.WithMessage("User %s operation %s failed: %s", userID, operation, reason).
		AddMetadata("user_id", userID).
		AddMetadata("operation", operation).
		AddMetadata("reason", reason)
}
