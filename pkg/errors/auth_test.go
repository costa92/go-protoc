package errors_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/costa92/go-protoc/v2/pkg/errors"
	"github.com/costa92/go-protoc/v2/pkg/errorsx"
)

func TestAuthErrors_Predefined(t *testing.T) {
	// 测试预定义的认证错误
	tests := []struct {
		name      string
		err       *errorsx.ErrorX
		expCode   int32
		expReason string
		expI18nKey string
	}{
		{
			name:      "ErrTokenInvalid",
			err:       errors.ErrTokenInvalid,
			expCode:   401,
			expReason: "TOKEN_INVALID",
			expI18nKey: "errors.auth.token_invalid",
		},
		{
			name:      "ErrTokenExpired",
			err:       errors.ErrTokenExpired,
			expCode:   401,
			expReason: "TOKEN_EXPIRED",
			expI18nKey: "errors.auth.token_expired",
		},
		{
			name:      "ErrTokenMissing",
			err:       errors.ErrTokenMissing,
			expCode:   401,
			expReason: "TOKEN_MISSING",
			expI18nKey: "errors.auth.token_missing",
		},
		{
			name:      "ErrInsufficientPermissions",
			err:       errors.ErrInsufficientPermissions,
			expCode:   403,
			expReason: "INSUFFICIENT_PERMISSIONS",
			expI18nKey: "errors.auth.insufficient_permissions",
		},
		{
			name:      "ErrLoginFailed",
			err:       errors.ErrLoginFailed,
			expCode:   401,
			expReason: "LOGIN_FAILED",
			expI18nKey: "errors.auth.login_failed",
		},
		{
			name:      "ErrAccountLocked",
			err:       errors.ErrAccountLocked,
			expCode:   423,
			expReason: "ACCOUNT_LOCKED",
			expI18nKey: "errors.auth.account_locked",
		},
		{
			name:      "ErrSessionExpired",
			err:       errors.ErrSessionExpired,
			expCode:   401,
			expReason: "SESSION_EXPIRED",
			expI18nKey: "errors.auth.session_expired",
		},
		{
			name:      "ErrLoginAttemptExceeded",
			err:       errors.ErrLoginAttemptExceeded,
			expCode:   429,
			expReason: "LOGIN_ATTEMPT_EXCEEDED",
			expI18nKey: "errors.auth.login_attempt_exceeded",
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

func TestNewTokenInvalidError(t *testing.T) {
	// 测试无效令牌错误构建
	token := "invalid.jwt.token"
	reason := "signature verification failed"
	
	err := errors.NewTokenInvalidError(token, reason)
	
	assert.Equal(t, int32(401), err.Code)
	assert.Equal(t, "TOKEN_INVALID", err.Reason)
	assert.Contains(t, err.Message, reason)
	assert.Equal(t, token, err.Metadata["token"])
	assert.Equal(t, reason, err.Metadata["reason"])
	assert.Equal(t, "errors.auth.token_invalid", err.GetI18nKey())
}

func TestNewTokenExpiredError(t *testing.T) {
	// 测试令牌过期错误构建
	token := "expired.jwt.token"
	expiredAt := time.Now().Add(-time.Hour)
	
	err := errors.NewTokenExpiredError(token, expiredAt)
	
	assert.Equal(t, int32(401), err.Code)
	assert.Equal(t, "TOKEN_EXPIRED", err.Reason)
	assert.Contains(t, err.Message, "expired")
	assert.Equal(t, token, err.Metadata["token"])
	assert.Equal(t, expiredAt.Format(time.RFC3339), err.Metadata["expired_at"])
	assert.Equal(t, "errors.auth.token_expired", err.GetI18nKey())
}

func TestNewTokenMissingError(t *testing.T) {
	// 测试令牌缺失错误构建
	header := "Authorization"
	
	err := errors.NewTokenMissingError(header)
	
	assert.Equal(t, int32(401), err.Code)
	assert.Equal(t, "TOKEN_MISSING", err.Reason)
	assert.Contains(t, err.Message, header)
	assert.Equal(t, header, err.Metadata["header"])
	assert.Equal(t, "errors.auth.token_missing", err.GetI18nKey())
}

func TestNewInsufficientPermissionsError(t *testing.T) {
	// 测试权限不足错误构建
	userID := "user123"
	requiredPermission := "admin:write"
	userPermissions := []string{"user:read", "user:write"}
	
	err := errors.NewInsufficientPermissionsError(userID, requiredPermission, userPermissions)
	
	assert.Equal(t, int32(403), err.Code)
	assert.Equal(t, "INSUFFICIENT_PERMISSIONS", err.Reason)
	assert.Contains(t, err.Message, requiredPermission)
	assert.Equal(t, userID, err.Metadata["user_id"])
	assert.Equal(t, requiredPermission, err.Metadata["required_permission"])
	assert.Equal(t, userPermissions, err.Metadata["user_permissions"])
	assert.Equal(t, "errors.auth.insufficient_permissions", err.GetI18nKey())
}

func TestNewLoginFailedError(t *testing.T) {
	// 测试登录失败错误构建
	username := "testuser"
	reason := "invalid password"
	attemptCount := 3
	
	err := errors.NewLoginFailedError(username, reason, attemptCount)
	
	assert.Equal(t, int32(401), err.Code)
	assert.Equal(t, "LOGIN_FAILED", err.Reason)
	assert.Contains(t, err.Message, reason)
	assert.Equal(t, username, err.Metadata["username"])
	assert.Equal(t, reason, err.Metadata["reason"])
	assert.Equal(t, attemptCount, err.Metadata["attempt_count"])
	assert.Equal(t, "errors.auth.login_failed", err.GetI18nKey())
}

func TestNewAccountLockedError(t *testing.T) {
	// 测试账户锁定错误构建
	userID := "user456"
	lockReason := "too many failed login attempts"
	unlockAt := time.Now().Add(time.Hour)
	
	err := errors.NewAccountLockedError(userID, lockReason, unlockAt)
	
	assert.Equal(t, int32(423), err.Code)
	assert.Equal(t, "ACCOUNT_LOCKED", err.Reason)
	assert.Contains(t, err.Message, lockReason)
	assert.Equal(t, userID, err.Metadata["user_id"])
	assert.Equal(t, lockReason, err.Metadata["lock_reason"])
	assert.Equal(t, unlockAt.Format(time.RFC3339), err.Metadata["unlock_at"])
	assert.Equal(t, "errors.auth.account_locked", err.GetI18nKey())
}

func TestNewSessionExpiredError(t *testing.T) {
	// 测试会话过期错误构建
	sessionID := "session789"
	expiredAt := time.Now().Add(-time.Minute * 30)
	
	err := errors.NewSessionExpiredError(sessionID, expiredAt)
	
	assert.Equal(t, int32(401), err.Code)
	assert.Equal(t, "SESSION_EXPIRED", err.Reason)
	assert.Contains(t, err.Message, "expired")
	assert.Equal(t, sessionID, err.Metadata["session_id"])
	assert.Equal(t, expiredAt.Format(time.RFC3339), err.Metadata["expired_at"])
	assert.Equal(t, "errors.auth.session_expired", err.GetI18nKey())
}

func TestNewLoginAttemptExceededError(t *testing.T) {
	// 测试登录尝试超限错误构建
	username := "testuser"
	maxAttempts := 5
	currentAttempts := 6
	lockDuration := time.Hour
	
	err := errors.NewLoginAttemptExceededError(username, maxAttempts, currentAttempts, lockDuration)
	
	assert.Equal(t, int32(429), err.Code)
	assert.Equal(t, "LOGIN_ATTEMPT_EXCEEDED", err.Reason)
	assert.Contains(t, err.Message, "exceeded")
	assert.Equal(t, username, err.Metadata["username"])
	assert.Equal(t, maxAttempts, err.Metadata["max_attempts"])
	assert.Equal(t, currentAttempts, err.Metadata["current_attempts"])
	assert.Equal(t, lockDuration.String(), err.Metadata["lock_duration"])
	assert.Equal(t, "errors.auth.login_attempt_exceeded", err.GetI18nKey())
}

func TestAuthErrors_ErrorChaining(t *testing.T) {
	// 测试认证错误链
	originalErr := assert.AnError
	token := "invalid.token"
	reason := "signature failed"
	
	// 创建带原始错误的认证错误
	err := errors.NewTokenInvalidError(token, reason)
	err = err.WithCause(originalErr)
	
	assert.Equal(t, originalErr, err.GetCause())
	assert.Equal(t, token, err.Metadata["token"])
	assert.Equal(t, reason, err.Metadata["reason"])
	
	// 测试错误链
	assert.True(t, errorsx.Is(err, errors.ErrTokenInvalid))
}

func TestAuthErrors_MetadataExtension(t *testing.T) {
	// 测试认证错误元数据扩展
	token := "test.token"
	reason := "test reason"
	err := errors.NewTokenInvalidError(token, reason)
	
	// 添加额外的元数据
	err = err.AddMetadata("request_id", "req-789")
	err = err.AddMetadata("client_ip", "192.168.1.1")
	err = err.AddMetadata("user_agent", "TestAgent/1.0")
	err = err.AddMetadata("timestamp", time.Now().Format(time.RFC3339))
	
	assert.Equal(t, token, err.Metadata["token"])
	assert.Equal(t, reason, err.Metadata["reason"])
	assert.Equal(t, "req-789", err.Metadata["request_id"])
	assert.Equal(t, "192.168.1.1", err.Metadata["client_ip"])
	assert.Equal(t, "TestAgent/1.0", err.Metadata["user_agent"])
	assert.NotEmpty(t, err.Metadata["timestamp"])
}

func TestAuthErrors_I18nIntegration(t *testing.T) {
	// 测试认证错误国际化集成
	tests := []struct {
		name   string
		err    *errorsx.ErrorX
		i18nKey string
	}{
		{"TokenInvalid", errors.ErrTokenInvalid, "errors.auth.token_invalid"},
		{"TokenExpired", errors.ErrTokenExpired, "errors.auth.token_expired"},
		{"TokenMissing", errors.ErrTokenMissing, "errors.auth.token_missing"},
		{"InsufficientPermissions", errors.ErrInsufficientPermissions, "errors.auth.insufficient_permissions"},
		{"LoginFailed", errors.ErrLoginFailed, "errors.auth.login_failed"},
		{"AccountLocked", errors.ErrAccountLocked, "errors.auth.account_locked"},
		{"SessionExpired", errors.ErrSessionExpired, "errors.auth.session_expired"},
		{"LoginAttemptExceeded", errors.ErrLoginAttemptExceeded, "errors.auth.login_attempt_exceeded"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.i18nKey, tt.err.GetI18nKey())
			assert.NotEmpty(t, tt.err.GetI18nKey())
		})
	}
}

func TestAuthErrors_BuilderPattern(t *testing.T) {
	// 测试认证错误构建器模式
	userID := "user123"
	requiredRole := "admin"
	userRoles := []string{"user", "editor"}
	
	// 使用构建器模式创建复杂的认证错误
	err := errorsx.Forbidden("INSUFFICIENT_PERMISSIONS").
		WithMessage("User does not have required permissions").
		WithI18nKey("errors.auth.insufficient_permissions").
		WithMetadata("user_id", userID).
		WithMetadata("required_role", requiredRole).
		WithMetadata("user_roles", userRoles).
		WithMetadata("resource", "admin_panel").
		WithMetadata("action", "delete_user").
		WithMetadata("timestamp", time.Now().Format(time.RFC3339)).
		Build()
	
	assert.Equal(t, int32(403), err.Code)
	assert.Equal(t, "INSUFFICIENT_PERMISSIONS", err.Reason)
	assert.Equal(t, userID, err.Metadata["user_id"])
	assert.Equal(t, requiredRole, err.Metadata["required_role"])
	assert.Equal(t, userRoles, err.Metadata["user_roles"])
	assert.Equal(t, "admin_panel", err.Metadata["resource"])
	assert.Equal(t, "delete_user", err.Metadata["action"])
	assert.NotEmpty(t, err.Metadata["timestamp"])
}

func TestAuthErrors_SecurityConsiderations(t *testing.T) {
	// 测试认证错误的安全考虑
	
	// 1. 敏感信息不应该暴露在错误消息中
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.sensitive.data"
	reason := "signature verification failed"
	
	err := errors.NewTokenInvalidError(token, reason)
	
	// 错误消息不应该包含完整的令牌
	assert.NotContains(t, err.Message, token)
	assert.Contains(t, err.Message, reason)
	
	// 但元数据中可以包含（用于日志记录等内部用途）
	assert.Equal(t, token, err.Metadata["token"])
	
	// 2. 登录失败不应该暴露用户是否存在
	username := "nonexistent@example.com"
	loginErr := errors.NewLoginFailedError(username, "invalid credentials", 1)
	
	// 错误消息应该是通用的
	assert.Contains(t, loginErr.Message, "invalid credentials")
	assert.NotContains(t, loginErr.Message, "user not found")
	assert.NotContains(t, loginErr.Message, "password incorrect")
}

func TestAuthErrors_TimeHandling(t *testing.T) {
	// 测试认证错误中的时间处理
	now := time.Now()
	expiredAt := now.Add(-time.Hour)
	unlockAt := now.Add(time.Hour)
	
	// 测试令牌过期时间格式
	tokenErr := errors.NewTokenExpiredError("test.token", expiredAt)
	assert.Equal(t, expiredAt.Format(time.RFC3339), tokenErr.Metadata["expired_at"])
	
	// 测试账户解锁时间格式
	accountErr := errors.NewAccountLockedError("user123", "too many attempts", unlockAt)
	assert.Equal(t, unlockAt.Format(time.RFC3339), accountErr.Metadata["unlock_at"])
	
	// 测试会话过期时间格式
	sessionErr := errors.NewSessionExpiredError("session123", expiredAt)
	assert.Equal(t, expiredAt.Format(time.RFC3339), sessionErr.Metadata["expired_at"])
}