package errors_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/costa92/go-protoc/v2/pkg/errors"
	"github.com/costa92/go-protoc/v2/pkg/errorsx"
)

func TestCommonErrors_Predefined(t *testing.T) {
	// 测试预定义的通用错误
	tests := []struct {
		name      string
		err       *errorsx.ErrorX
		expCode   int32
		expReason string
		expI18nKey string
	}{
		{
			name:      "ErrResourceNotFound",
			err:       errors.ErrResourceNotFound,
			expCode:   404,
			expReason: "RESOURCE_NOT_FOUND",
			expI18nKey: "errors.common.resource_not_found",
		},
		{
			name:      "ErrInvalidRequest",
			err:       errors.ErrInvalidRequest,
			expCode:   400,
			expReason: "INVALID_REQUEST",
			expI18nKey: "errors.common.invalid_request",
		},
		{
			name:      "ErrInternalServer",
			err:       errors.ErrInternalServer,
			expCode:   500,
			expReason: "INTERNAL_SERVER_ERROR",
			expI18nKey: "errors.common.internal_server_error",
		},
		{
			name:      "ErrRateLimitExceeded",
			err:       errors.ErrRateLimitExceeded,
			expCode:   429,
			expReason: "RATE_LIMIT_EXCEEDED",
			expI18nKey: "errors.common.rate_limit_exceeded",
		},
		{
			name:      "ErrServiceUnavailable",
			err:       errors.ErrServiceUnavailable,
			expCode:   503,
			expReason: "SERVICE_UNAVAILABLE",
			expI18nKey: "errors.common.service_unavailable",
		},
		{
			name:      "ErrRequestTimeout",
			err:       errors.ErrRequestTimeout,
			expCode:   408,
			expReason: "REQUEST_TIMEOUT",
			expI18nKey: "errors.common.request_timeout",
		},
		{
			name:      "ErrDatabaseConnection",
			err:       errors.ErrDatabaseConnection,
			expCode:   500,
			expReason: "DATABASE_CONNECTION_ERROR",
			expI18nKey: "errors.common.database_connection_error",
		},
		{
			name:      "ErrExternalService",
			err:       errors.ErrExternalService,
			expCode:   502,
			expReason: "EXTERNAL_SERVICE_ERROR",
			expI18nKey: "errors.common.external_service_error",
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

func TestNewResourceNotFoundError(t *testing.T) {
	// 测试资源未找到错误构建
	resourceType := "user"
	resourceID := "123"
	
	err := errors.NewResourceNotFoundError(resourceType, resourceID)
	
	assert.Equal(t, int32(404), err.Code)
	assert.Equal(t, "RESOURCE_NOT_FOUND", err.Reason)
	assert.Contains(t, err.Message, resourceType)
	assert.Contains(t, err.Message, resourceID)
	assert.Equal(t, resourceType, err.Metadata["resource_type"])
	assert.Equal(t, resourceID, err.Metadata["resource_id"])
	assert.Equal(t, "errors.common.resource_not_found", err.GetI18nKey())
}

func TestNewInvalidParameterError(t *testing.T) {
	// 测试无效参数错误构建
	paramName := "email"
	paramValue := "invalid-email"
	reason := "invalid email format"
	
	err := errors.NewInvalidParameterError(paramName, paramValue, reason)
	
	assert.Equal(t, int32(400), err.Code)
	assert.Equal(t, "INVALID_REQUEST", err.Reason)
	assert.Contains(t, err.Message, paramName)
	assert.Contains(t, err.Message, reason)
	assert.Equal(t, paramName, err.Metadata["parameter_name"])
	assert.Equal(t, paramValue, err.Metadata["parameter_value"])
	assert.Equal(t, reason, err.Metadata["reason"])
	assert.Equal(t, "errors.common.invalid_request", err.GetI18nKey())
}

func TestNewMissingParameterError(t *testing.T) {
	// 测试缺失参数错误构建
	paramName := "username"
	
	err := errors.NewMissingParameterError(paramName)
	
	assert.Equal(t, int32(400), err.Code)
	assert.Equal(t, "INVALID_REQUEST", err.Reason)
	assert.Contains(t, err.Message, paramName)
	assert.Contains(t, err.Message, "required")
	assert.Equal(t, paramName, err.Metadata["parameter_name"])
	assert.Equal(t, "errors.common.invalid_request", err.GetI18nKey())
}

func TestNewRateLimitExceededError(t *testing.T) {
	// 测试速率限制超限错误构建
	limit := 100
	window := time.Hour
	retryAfter := time.Minute * 30
	
	err := errors.NewRateLimitExceededError(limit, window, retryAfter)
	
	assert.Equal(t, int32(429), err.Code)
	assert.Equal(t, "RATE_LIMIT_EXCEEDED", err.Reason)
	assert.Contains(t, err.Message, "rate limit")
	assert.Equal(t, limit, err.Metadata["limit"])
	assert.Equal(t, window.String(), err.Metadata["window"])
	assert.Equal(t, retryAfter.String(), err.Metadata["retry_after"])
	assert.Equal(t, "errors.common.rate_limit_exceeded", err.GetI18nKey())
}

func TestNewDatabaseError(t *testing.T) {
	// 测试数据库错误构建
	operation := "SELECT"
	table := "users"
	originalErr := sql.ErrNoRows
	
	err := errors.NewDatabaseError(operation, table, originalErr)
	
	assert.Equal(t, int32(500), err.Code)
	assert.Equal(t, "DATABASE_CONNECTION_ERROR", err.Reason)
	assert.Contains(t, err.Message, operation)
	assert.Contains(t, err.Message, table)
	assert.Equal(t, operation, err.Metadata["operation"])
	assert.Equal(t, table, err.Metadata["table"])
	assert.Equal(t, originalErr.Error(), err.Metadata["original_error"])
	assert.Equal(t, originalErr, err.GetCause())
	assert.Equal(t, "errors.common.database_connection_error", err.GetI18nKey())
}

func TestNewExternalServiceError(t *testing.T) {
	// 测试外部服务错误构建
	serviceName := "payment-service"
	endpoint := "/api/v1/payments"
	statusCode := 502
	responseBody := "Bad Gateway"
	
	err := errors.NewExternalServiceError(serviceName, endpoint, statusCode, responseBody)
	
	assert.Equal(t, int32(502), err.Code)
	assert.Equal(t, "EXTERNAL_SERVICE_ERROR", err.Reason)
	assert.Contains(t, err.Message, serviceName)
	assert.Contains(t, err.Message, endpoint)
	assert.Equal(t, serviceName, err.Metadata["service_name"])
	assert.Equal(t, endpoint, err.Metadata["endpoint"])
	assert.Equal(t, statusCode, err.Metadata["status_code"])
	assert.Equal(t, responseBody, err.Metadata["response_body"])
	assert.Equal(t, "errors.common.external_service_error", err.GetI18nKey())
}

func TestNewTimeoutError(t *testing.T) {
	// 测试超时错误构建
	operation := "database query"
	timeout := time.Second * 30
	
	err := errors.NewTimeoutError(operation, timeout)
	
	assert.Equal(t, int32(408), err.Code)
	assert.Equal(t, "REQUEST_TIMEOUT", err.Reason)
	assert.Contains(t, err.Message, operation)
	assert.Contains(t, err.Message, "timeout")
	assert.Equal(t, operation, err.Metadata["operation"])
	assert.Equal(t, timeout.String(), err.Metadata["timeout"])
	assert.Equal(t, "errors.common.request_timeout", err.GetI18nKey())
}

func TestCommonErrors_ErrorChaining(t *testing.T) {
	// 测试通用错误链
	originalErr := assert.AnError
	resourceType := "order"
	resourceID := "456"
	
	// 创建带原始错误的资源未找到错误
	err := errors.NewResourceNotFoundError(resourceType, resourceID)
	err = err.WithCause(originalErr)
	
	assert.Equal(t, originalErr, err.GetCause())
	assert.Equal(t, resourceType, err.Metadata["resource_type"])
	assert.Equal(t, resourceID, err.Metadata["resource_id"])
	
	// 测试错误链
	assert.True(t, errorsx.Is(err, errors.ErrResourceNotFound))
}

func TestCommonErrors_MetadataExtension(t *testing.T) {
	// 测试通用错误元数据扩展
	resourceType := "product"
	resourceID := "789"
	err := errors.NewResourceNotFoundError(resourceType, resourceID)
	
	// 添加额外的元数据
	err = err.AddMetadata("request_id", "req-456")
	err = err.AddMetadata("user_id", "user-123")
	err = err.AddMetadata("search_criteria", map[string]any{
		"category": "electronics",
		"price_range": []int{100, 500},
		"in_stock": true,
	})
	err = err.AddMetadata("timestamp", time.Now().Format(time.RFC3339))
	
	assert.Equal(t, resourceType, err.Metadata["resource_type"])
	assert.Equal(t, resourceID, err.Metadata["resource_id"])
	assert.Equal(t, "req-456", err.Metadata["request_id"])
	assert.Equal(t, "user-123", err.Metadata["user_id"])
	assert.NotNil(t, err.Metadata["search_criteria"])
	assert.NotEmpty(t, err.Metadata["timestamp"])
}

func TestCommonErrors_I18nIntegration(t *testing.T) {
	// 测试通用错误国际化集成
	tests := []struct {
		name   string
		err    *errorsx.ErrorX
		i18nKey string
	}{
		{"ResourceNotFound", errors.ErrResourceNotFound, "errors.common.resource_not_found"},
		{"InvalidRequest", errors.ErrInvalidRequest, "errors.common.invalid_request"},
		{"InternalServer", errors.ErrInternalServer, "errors.common.internal_server_error"},
		{"RateLimitExceeded", errors.ErrRateLimitExceeded, "errors.common.rate_limit_exceeded"},
		{"ServiceUnavailable", errors.ErrServiceUnavailable, "errors.common.service_unavailable"},
		{"RequestTimeout", errors.ErrRequestTimeout, "errors.common.request_timeout"},
		{"DatabaseConnection", errors.ErrDatabaseConnection, "errors.common.database_connection_error"},
		{"ExternalService", errors.ErrExternalService, "errors.common.external_service_error"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.i18nKey, tt.err.GetI18nKey())
			assert.NotEmpty(t, tt.err.GetI18nKey())
		})
	}
}

func TestCommonErrors_BuilderPattern(t *testing.T) {
	// 测试通用错误构建器模式
	serviceName := "user-service"
	endpoint := "/api/v1/users/123"
	statusCode := 503
	
	// 使用构建器模式创建复杂的外部服务错误
	err := errorsx.BadGateway().
		WithReason("EXTERNAL_SERVICE_ERROR").
		WithMessage("External service is temporarily unavailable").
		WithI18nKey("errors.common.external_service_error").
		AddMetadata("service_name", serviceName).
		AddMetadata("endpoint", endpoint).
		AddMetadata("status_code", statusCode).
		AddMetadata("retry_count", 3).
		AddMetadata("last_attempt", time.Now().Format(time.RFC3339)).
		AddMetadata("circuit_breaker_state", "OPEN").
		Build()
	
	assert.Equal(t, int32(502), err.Code)
	assert.Equal(t, "EXTERNAL_SERVICE_ERROR", err.Reason)
	assert.Equal(t, serviceName, err.Metadata["service_name"])
	assert.Equal(t, endpoint, err.Metadata["endpoint"])
	assert.Equal(t, statusCode, err.Metadata["status_code"])
	assert.Equal(t, 3, err.Metadata["retry_count"])
	assert.NotEmpty(t, err.Metadata["last_attempt"])
	assert.Equal(t, "OPEN", err.Metadata["circuit_breaker_state"])
}

func TestCommonErrors_ValidationScenarios(t *testing.T) {
	// 测试通用错误验证场景
	
	// 1. 多个参数验证错误
	validationErrors := []*errorsx.ErrorX{
		errors.NewInvalidParameterError("email", "invalid-email", "invalid format"),
		errors.NewMissingParameterError("password"),
		errors.NewInvalidParameterError("age", "-5", "must be positive"),
	}
	
	for _, err := range validationErrors {
		assert.Equal(t, int32(400), err.Code)
		assert.Equal(t, "INVALID_REQUEST", err.Reason)
		assert.NotEmpty(t, err.Metadata["parameter_name"])
	}
	
	// 2. 嵌套资源未找到
	parentType := "organization"
	parentID := "org-123"
	childType := "project"
	childID := "proj-456"
	
	err := errors.NewResourceNotFoundError(childType, childID).
		AddMetadata("parent_type", parentType).
		AddMetadata("parent_id", parentID).
		AddMetadata("hierarchy", []string{parentType, childType})
	
	assert.Equal(t, childType, err.Metadata["resource_type"])
	assert.Equal(t, childID, err.Metadata["resource_id"])
	assert.Equal(t, parentType, err.Metadata["parent_type"])
	assert.Equal(t, parentID, err.Metadata["parent_id"])
	assert.NotNil(t, err.Metadata["hierarchy"])
}

func TestCommonErrors_PerformanceConsiderations(t *testing.T) {
	// 测试通用错误性能考虑
	
	// 1. 错误对象复用
	baseErr := errors.ErrResourceNotFound
	
	// 创建多个基于同一模板的错误实例
	err1 := baseErr.WithMetadata(map[string]any{
		"resource_type": "user",
		"resource_id":   "123",
	})
	
	err2 := baseErr.WithMetadata(map[string]any{
		"resource_type": "order",
		"resource_id":   "456",
	})
	
	// 验证它们是不同的实例但共享相同的基础属性
	assert.Equal(t, baseErr.Code, err1.Code)
	assert.Equal(t, baseErr.Code, err2.Code)
	assert.Equal(t, baseErr.Reason, err1.Reason)
	assert.Equal(t, baseErr.Reason, err2.Reason)
	assert.NotEqual(t, err1.Metadata["resource_type"], err2.Metadata["resource_type"])
	
	// 2. 批量错误处理
	resources := []struct {
		resourceType string
		resourceID   string
	}{
		{"user", "1"},
		{"user", "2"},
		{"order", "100"},
		{"product", "200"},
	}
	
	var batchErrors []*errorsx.ErrorX
	for _, resource := range resources {
		err := errors.NewResourceNotFoundError(resource.resourceType, resource.resourceID)
		batchErrors = append(batchErrors, err)
	}
	
	assert.Len(t, batchErrors, 4)
	for _, err := range batchErrors {
		assert.Equal(t, int32(404), err.Code)
		assert.Equal(t, "RESOURCE_NOT_FOUND", err.Reason)
		assert.NotEmpty(t, err.Metadata["resource_type"])
		assert.NotEmpty(t, err.Metadata["resource_id"])
	}
}

func TestCommonErrors_ErrorAggregation(t *testing.T) {
	// 测试通用错误聚合
	
	// 创建多个验证错误
	validationErrors := []*errorsx.ErrorX{
		errors.NewInvalidParameterError("email", "invalid", "invalid format"),
		errors.NewMissingParameterError("name"),
		errors.NewInvalidParameterError("age", "abc", "must be number"),
	}
	
	// 创建聚合错误
	aggregatedErr := errorsx.BadRequest().
		WithReason("VALIDATION_FAILED").
		WithMessage("Multiple validation errors occurred").
		WithI18nKey("errors.common.validation_failed").
		AddMetadata("error_count", len(validationErrors)).
		AddMetadata("validation_errors", validationErrors).
		Build()
	
	assert.Equal(t, int32(400), aggregatedErr.Code)
	assert.Equal(t, "VALIDATION_FAILED", aggregatedErr.Reason)
	assert.Equal(t, len(validationErrors), aggregatedErr.Metadata["error_count"])
	assert.Equal(t, validationErrors, aggregatedErr.Metadata["validation_errors"])
}