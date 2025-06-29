package errors_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/costa92/go-protoc/v2/pkg/errors"
	"github.com/costa92/go-protoc/v2/pkg/errorsx"
)

func TestUserErrors_Predefined(t *testing.T) {
	// 测试预定义的用户错误
	tests := []struct {
		name      string
		err       *errorsx.ErrorX
		expCode   int32
		expReason string
		expI18nKey string
	}{
		{
			name:      "ErrUserNotFound",
			err:       errors.ErrUserNotFound,
			expCode:   404,
			expReason: "USER_NOT_FOUND",
			expI18nKey: "errors.user.not_found",
		},
		{
			name:      "ErrUserAlreadyExists",
			err:       errors.ErrUserAlreadyExists,
			expCode:   409,
			expReason: "USER_ALREADY_EXISTS",
			expI18nKey: "errors.user.already_exists",
		},
		{
			name:      "ErrUserValidationFailed",
			err:       errors.ErrUserValidationFailed,
			expCode:   400,
			expReason: "USER_VALIDATION_FAILED",
			expI18nKey: "errors.user.validation_failed",
		},
		{
			name:      "ErrUserInactive",
			err:       errors.ErrUserInactive,
			expCode:   403,
			expReason: "USER_INACTIVE",
			expI18nKey: "errors.user.inactive",
		},
		{
			name:      "ErrUserPasswordInvalid",
			err:       errors.ErrUserPasswordInvalid,
			expCode:   401,
			expReason: "USER_PASSWORD_INVALID",
			expI18nKey: "errors.user.password_invalid",
		},
		{
			name:      "ErrUserEmailInvalid",
			err:       errors.ErrUserEmailInvalid,
			expCode:   400,
			expReason: "USER_EMAIL_INVALID",
			expI18nKey: "errors.user.email_invalid",
		},
		{
			name:      "ErrUserPermissionDenied",
			err:       errors.ErrUserPermissionDenied,
			expCode:   403,
			expReason: "USER_PERMISSION_DENIED",
			expI18nKey: "errors.user.permission_denied",
		},
		{
			name:      "ErrUserOperationFailed",
			err:       errors.ErrUserOperationFailed,
			expCode:   500,
			expReason: "USER_OPERATION_FAILED",
			expI18nKey: "errors.user.operation_failed",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expCode, tt.err.Code)
			assert.Equal(t, tt.expReason, tt.err.Reason)
			assert.Equal(t, tt.expI18nKey, tt.err.GetI18nKey())
		})
	}
}

func TestNewUserNotFoundError(t *testing.T) {
	// 测试用户未找到错误构建
	userID := "user123"
	err := errors.NewUserNotFoundError(userID)
	
	assert.Equal(t, int32(404), err.Code)
	assert.Equal(t, "USER_NOT_FOUND", err.Reason)
	assert.Contains(t, err.Message, userID)
	assert.Equal(t, userID, err.Metadata["user_id"])
	assert.Equal(t, "errors.user.not_found", err.GetI18nKey())
}

func TestNewUserAlreadyExistsError(t *testing.T) {
	// 测试用户已存在错误构建
	email := "test@example.com"
	err := errors.NewUserAlreadyExistsError(email)
	
	assert.Equal(t, int32(409), err.Code)
	assert.Equal(t, "USER_ALREADY_EXISTS", err.Reason)
	assert.Contains(t, err.Message, email)
	assert.Equal(t, email, err.Metadata["email"])
	assert.Equal(t, "errors.user.already_exists", err.GetI18nKey())
}

func TestNewUserValidationError(t *testing.T) {
	// 测试用户验证错误构建
	field := "email"
	value := "invalid-email"
	reason := "format invalid"
	
	err := errors.NewUserValidationError(field, value, reason)
	
	assert.Equal(t, int32(400), err.Code)
	assert.Equal(t, "USER_VALIDATION_FAILED", err.Reason)
	assert.Contains(t, err.Message, field)
	assert.Contains(t, err.Message, reason)
	assert.Equal(t, field, err.Metadata["field"])
	assert.Equal(t, value, err.Metadata["value"])
	assert.Equal(t, reason, err.Metadata["reason"])
	assert.Equal(t, "errors.user.validation_failed", err.GetI18nKey())
}

func TestNewUserInactiveError(t *testing.T) {
	// 测试用户未激活错误构建
	userID := "user456"
	status := "suspended"
	
	err := errors.NewUserInactiveError(userID, status)
	
	assert.Equal(t, int32(403), err.Code)
	assert.Equal(t, "USER_INACTIVE", err.Reason)
	assert.Contains(t, err.Message, userID)
	assert.Contains(t, err.Message, status)
	assert.Equal(t, userID, err.Metadata["user_id"])
	assert.Equal(t, status, err.Metadata["status"])
	assert.Equal(t, "errors.user.inactive", err.GetI18nKey())
}

func TestNewUserPasswordInvalidError(t *testing.T) {
	// 测试用户密码无效错误构建
	userID := "user789"
	reason := "password too weak"
	
	err := errors.NewUserPasswordInvalidError(userID, reason)
	
	assert.Equal(t, int32(401), err.Code)
	assert.Equal(t, "USER_PASSWORD_INVALID", err.Reason)
	assert.Contains(t, err.Message, reason)
	assert.Equal(t, userID, err.Metadata["user_id"])
	assert.Equal(t, reason, err.Metadata["reason"])
	assert.Equal(t, "errors.user.password_invalid", err.GetI18nKey())
}

func TestNewUserEmailInvalidError(t *testing.T) {
	// 测试用户邮箱无效错误构建
	email := "invalid@email"
	reason := "domain not allowed"
	
	err := errors.NewUserEmailInvalidError(email, reason)
	
	assert.Equal(t, int32(400), err.Code)
	assert.Equal(t, "USER_EMAIL_INVALID", err.Reason)
	assert.Contains(t, err.Message, email)
	assert.Contains(t, err.Message, reason)
	assert.Equal(t, email, err.Metadata["email"])
	assert.Equal(t, reason, err.Metadata["reason"])
	assert.Equal(t, "errors.user.email_invalid", err.GetI18nKey())
}

func TestNewUserPermissionDeniedError(t *testing.T) {
	// 测试用户权限拒绝错误构建
	userID := "user999"
	action := "delete_user"
	resource := "admin_panel"
	
	err := errors.NewUserPermissionDeniedError(userID, action, resource)
	
	assert.Equal(t, int32(403), err.Code)
	assert.Equal(t, "USER_PERMISSION_DENIED", err.Reason)
	assert.Contains(t, err.Message, action)
	assert.Contains(t, err.Message, resource)
	assert.Equal(t, userID, err.Metadata["user_id"])
	assert.Equal(t, action, err.Metadata["action"])
	assert.Equal(t, resource, err.Metadata["resource"])
	assert.Equal(t, "errors.user.permission_denied", err.GetI18nKey())
}

func TestNewUserOperationFailedError(t *testing.T) {
	// 测试用户操作失败错误构建
	userID := "user111"
	operation := "update_profile"
	reason := "database timeout"
	
	err := errors.NewUserOperationFailedError(userID, operation, reason)
	
	assert.Equal(t, int32(500), err.Code)
	assert.Equal(t, "USER_OPERATION_FAILED", err.Reason)
	assert.Contains(t, err.Message, operation)
	assert.Contains(t, err.Message, reason)
	assert.Equal(t, userID, err.Metadata["user_id"])
	assert.Equal(t, operation, err.Metadata["operation"])
	assert.Equal(t, reason, err.Metadata["reason"])
	assert.Equal(t, "errors.user.operation_failed", err.GetI18nKey())
}

func TestUserErrors_ErrorChaining(t *testing.T) {
	// 测试用户错误链
	originalErr := assert.AnError
	userID := "user123"
	
	// 创建带原始错误的用户错误
	err := errors.NewUserNotFoundError(userID)
	err = err.WithCause(originalErr)
	
	assert.Equal(t, originalErr, err.GetCause())
	assert.Equal(t, userID, err.Metadata["user_id"])
	
	// 测试错误链
	assert.True(t, errorsx.Is(err, errors.ErrUserNotFound))
}

func TestUserErrors_MetadataExtension(t *testing.T) {
	// 测试用户错误元数据扩展
	userID := "user123"
	err := errors.NewUserNotFoundError(userID)
	
	// 添加额外的元数据
	err = err.AddMetadata("request_id", "req-456")
	err = err.AddMetadata("timestamp", "2023-01-01T00:00:00Z")
	err = err.AddMetadata("source", "user_service")
	
	assert.Equal(t, userID, err.Metadata["user_id"])
	assert.Equal(t, "req-456", err.Metadata["request_id"])
	assert.Equal(t, "2023-01-01T00:00:00Z", err.Metadata["timestamp"])
	assert.Equal(t, "user_service", err.Metadata["source"])
}

func TestUserErrors_I18nIntegration(t *testing.T) {
	// 测试用户错误国际化集成
	tests := []struct {
		name   string
		err    *errorsx.ErrorX
		i18nKey string
	}{
		{"UserNotFound", errors.ErrUserNotFound, "errors.user.not_found"},
		{"UserAlreadyExists", errors.ErrUserAlreadyExists, "errors.user.already_exists"},
		{"UserValidationFailed", errors.ErrUserValidationFailed, "errors.user.validation_failed"},
		{"UserInactive", errors.ErrUserInactive, "errors.user.inactive"},
		{"UserPasswordInvalid", errors.ErrUserPasswordInvalid, "errors.user.password_invalid"},
		{"UserEmailInvalid", errors.ErrUserEmailInvalid, "errors.user.email_invalid"},
		{"UserPermissionDenied", errors.ErrUserPermissionDenied, "errors.user.permission_denied"},
		{"UserOperationFailed", errors.ErrUserOperationFailed, "errors.user.operation_failed"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.i18nKey, tt.err.GetI18nKey())
			assert.NotEmpty(t, tt.err.GetI18nKey())
		})
	}
}

func TestUserErrors_BuilderPattern(t *testing.T) {
	// 测试用户错误构建器模式
	userID := "user123"
	email := "test@example.com"
	
	// 使用构建器模式创建复杂的用户错误
	err := errorsx.BadRequest().
		WithReason("USER_VALIDATION_FAILED").
		WithMessage("User validation failed for multiple fields").
		WithI18nKey("errors.user.validation_failed").
		AddMetadata("user_id", userID).
		AddMetadata("email", email).
		AddMetadata("fields", []string{"email", "password"}).
		AddMetadata("validation_errors", map[string]string{
			"email":    "invalid format",
			"password": "too weak",
		}).
		Build()
	
	assert.Equal(t, int32(400), err.Code)
	assert.Equal(t, "USER_VALIDATION_FAILED", err.Reason)
	assert.Equal(t, userID, err.Metadata["user_id"])
	assert.Equal(t, email, err.Metadata["email"])
	assert.NotNil(t, err.Metadata["fields"])
	assert.NotNil(t, err.Metadata["validation_errors"])
}