package errorsx_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/costa92/go-protoc/v2/pkg/errorsx"
)

func TestErrorTemplate(t *testing.T) {
	// 测试错误模板创建
	template := &errorsx.ErrorTemplate{
		Code:    400,
		Reason:  "VALIDATION_ERROR",
		I18nKey: "errors.validation.failed",
	}

	assert.Equal(t, int32(400), template.Code)
	assert.Equal(t, "VALIDATION_ERROR", template.Reason)
	assert.Equal(t, "errors.validation.failed", template.I18nKey)
}

func TestRegistry_Register(t *testing.T) {
	// 创建新的注册器实例
	registry := errorsx.NewRegistry()

	// 注册错误模板
	template := &errorsx.ErrorTemplate{
		Code:    404,
		Reason:  "USER_NOT_FOUND",
		I18nKey: "errors.user.not_found",
	}

	err := registry.Register(template)
	assert.NoError(t, err)

	// 验证注册成功
	assert.True(t, registry.TemplateExists("USER_NOT_FOUND"))
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	// 创建新的注册器实例
	registry := errorsx.NewRegistry()

	// 注册第一个模板
	template1 := &errorsx.ErrorTemplate{
		Code:    404,
		Reason:  "USER_NOT_FOUND",
		I18nKey: "errors.user.not_found",
	}

	err := registry.Register(template1)
	assert.NoError(t, err)

	// 尝试注册重复的模板
	template2 := &errorsx.ErrorTemplate{
		Code:    400,
		Reason:  "USER_NOT_FOUND", // 相同的 Reason
		I18nKey: "errors.user.not_found_v2",
	}

	err = registry.Register(template2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestRegistry_GetTemplate(t *testing.T) {
	// 创建新的注册器实例
	registry := errorsx.NewRegistry()

	// 注册错误模板
	originalTemplate := &errorsx.ErrorTemplate{
		Code:    422,
		Reason:  "VALIDATION_FAILED",
		I18nKey: "errors.validation.failed",
	}

	err := registry.Register(originalTemplate)
	assert.NoError(t, err)

	// 获取模板
	retrievedTemplate, exists := registry.GetTemplate("VALIDATION_FAILED")
	assert.True(t, exists)
	assert.Equal(t, originalTemplate.Code, retrievedTemplate.Code)
	assert.Equal(t, originalTemplate.Reason, retrievedTemplate.Reason)
	assert.Equal(t, originalTemplate.I18nKey, retrievedTemplate.I18nKey)

	// 获取不存在的模板
	_, exists = registry.GetTemplate("NON_EXISTENT")
	assert.False(t, exists)
}

func TestRegistry_ListTemplates(t *testing.T) {
	// 创建新的注册器实例
	registry := errorsx.NewRegistry()

	// 注册多个模板
	templates := []*errorsx.ErrorTemplate{
		{Code: 400, Reason: "BAD_REQUEST", I18nKey: "errors.bad_request"},
		{Code: 401, Reason: "UNAUTHORIZED", I18nKey: "errors.unauthorized"},
		{Code: 403, Reason: "FORBIDDEN", I18nKey: "errors.forbidden"},
	}

	for _, template := range templates {
		err := registry.Register(template)
		assert.NoError(t, err)
	}

	// 列出所有模板
	allTemplates := registry.ListTemplates()
	assert.Len(t, allTemplates, 3)

	// 验证所有模板都存在
	reasons := make(map[string]bool)
	for _, template := range allTemplates {
		reasons[template.Reason] = true
	}

	assert.True(t, reasons["BAD_REQUEST"])
	assert.True(t, reasons["UNAUTHORIZED"])
	assert.True(t, reasons["FORBIDDEN"])
}

func TestRegistry_TemplateExists(t *testing.T) {
	// 创建新的注册器实例
	registry := errorsx.NewRegistry()

	// 检查不存在的模板
	assert.False(t, registry.TemplateExists("NON_EXISTENT"))

	// 注册模板
	template := &errorsx.ErrorTemplate{
		Code:    500,
		Reason:  "INTERNAL_ERROR",
		I18nKey: "errors.internal",
	}

	err := registry.Register(template)
	assert.NoError(t, err)

	// 检查存在的模板
	assert.True(t, registry.TemplateExists("INTERNAL_ERROR"))
}

func TestRegistry_Create(t *testing.T) {
	// 创建新的注册器实例
	registry := errorsx.NewRegistry()

	// 注册模板
	template := &errorsx.ErrorTemplate{
		Code:    404,
		Reason:  "RESOURCE_NOT_FOUND",
		I18nKey: "errors.resource.not_found",
	}

	err := registry.Register(template)
	assert.NoError(t, err)

	// 使用模板创建错误
	errorX, err := registry.Create("RESOURCE_NOT_FOUND")
	assert.NoError(t, err)
	assert.NotNil(t, errorX)

	assert.Equal(t, int32(404), errorX.Code)
	assert.Equal(t, "RESOURCE_NOT_FOUND", errorX.Reason)
	assert.Equal(t, "errors.resource.not_found", errorX.GetI18nKey())

	// 尝试创建不存在的错误
	_, err = registry.Create("NON_EXISTENT")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRegistry_MustCreate(t *testing.T) {
	// 创建新的注册器实例
	registry := errorsx.NewRegistry()

	// 注册模板
	template := &errorsx.ErrorTemplate{
		Code:    409,
		Reason:  "CONFLICT",
		I18nKey: "errors.conflict",
	}

	err := registry.Register(template)
	assert.NoError(t, err)

	// 使用模板创建错误（不应该 panic）
	errorX := registry.MustCreate("CONFLICT")
	assert.NotNil(t, errorX)
	assert.Equal(t, int32(409), errorX.Code)
	assert.Equal(t, "CONFLICT", errorX.Reason)

	// 测试 panic 情况
	assert.Panics(t, func() {
		registry.MustCreate("NON_EXISTENT")
	})
}

func TestGlobalRegistry(t *testing.T) {
	// 测试全局注册器函数

	// 注册模板
	template := &errorsx.ErrorTemplate{
		Code:    429,
		Reason:  "RATE_LIMITED",
		I18nKey: "errors.rate_limit",
	}

	err := errorsx.Register(template)
	assert.NoError(t, err)

	// 检查模板存在
	assert.True(t, errorsx.TemplateExists("RATE_LIMITED"))

	// 获取模板
	retrievedTemplate, exists := errorsx.GetTemplate("RATE_LIMITED")
	assert.True(t, exists)
	assert.Equal(t, template.Code, retrievedTemplate.Code)
	assert.Equal(t, template.Reason, retrievedTemplate.Reason)
	assert.Equal(t, template.I18nKey, retrievedTemplate.I18nKey)

	// 创建错误
	errorX, err := errorsx.Create("RATE_LIMITED")
	assert.NoError(t, err)
	assert.Equal(t, int32(429), errorX.Code)
	assert.Equal(t, "RATE_LIMITED", errorX.Reason)

	// MustCreate
	errorX2 := errorsx.MustCreate("RATE_LIMITED")
	assert.Equal(t, int32(429), errorX2.Code)
	assert.Equal(t, "RATE_LIMITED", errorX2.Reason)

	// 列出模板
	allTemplates := errorsx.ListTemplates()
	assert.NotEmpty(t, allTemplates)

	// 验证我们注册的模板在列表中
	found := false
	for _, tmpl := range allTemplates {
		if tmpl.Reason == "RATE_LIMITED" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	// 测试并发访问安全性
	registry := errorsx.NewRegistry()

	// 注册一些模板
	templates := []*errorsx.ErrorTemplate{
		{Code: 400, Reason: "ERROR_1", I18nKey: "errors.1"},
		{Code: 401, Reason: "ERROR_2", I18nKey: "errors.2"},
		{Code: 402, Reason: "ERROR_3", I18nKey: "errors.3"},
	}

	for _, template := range templates {
		err := registry.Register(template)
		assert.NoError(t, err)
	}

	// 并发读取
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// 并发获取模板
			for j := 1; j <= 3; j++ {
				reason := fmt.Sprintf("ERROR_%d", j)
				_, exists := registry.GetTemplate(reason)
				assert.True(t, exists)

				// 并发创建错误
				_, err := registry.Create(reason)
				assert.NoError(t, err)

				// 并发检查存在性
				assert.True(t, registry.TemplateExists(reason))
			}

			// 并发列出模板
			allTemplates := registry.ListTemplates()
			assert.Len(t, allTemplates, 3)
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestRegistry_EmptyRegistry(t *testing.T) {
	// 测试空注册器
	registry := errorsx.NewRegistry()

	// 列出模板应该返回空切片
	allTemplates := registry.ListTemplates()
	assert.Empty(t, allTemplates)

	// 获取不存在的模板
	_, exists := registry.GetTemplate("NON_EXISTENT")
	assert.False(t, exists)

	// 检查不存在的模板
	assert.False(t, registry.TemplateExists("NON_EXISTENT"))

	// 创建不存在的错误应该失败
	_, err := registry.Create("NON_EXISTENT")
	assert.Error(t, err)
}
