package errorsx

import (
	"fmt"
	"sync"
)

// ErrorTemplate 错误模板，用于定义错误的基本信息
type ErrorTemplate struct {
	Code    int32  `json:"code"`
	Reason  string `json:"reason"`
	I18nKey string `json:"i18n_key,omitempty"`
}

// Registry 错误注册器，用于管理错误模板
type Registry struct {
	errors map[string]*ErrorTemplate
	mutex  sync.RWMutex
}

// NewRegistry 创建新的错误注册器
func NewRegistry() *Registry {
	return &Registry{
		errors: make(map[string]*ErrorTemplate),
	}
}

// Register 注册错误模板
func (r *Registry) Register(reason string, template *ErrorTemplate) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.errors[reason] = template
}

// Get 获取错误模板
func (r *Registry) Get(reason string) (*ErrorTemplate, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	template, exists := r.errors[reason]
	return template, exists
}

// MustGet 获取错误模板，如果不存在则 panic
func (r *Registry) MustGet(reason string) *ErrorTemplate {
	template, exists := r.Get(reason)
	if !exists {
		panic(fmt.Sprintf("error template not found: %s", reason))
	}
	return template
}

// List 列出所有注册的错误模板
func (r *Registry) List() map[string]*ErrorTemplate {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	result := make(map[string]*ErrorTemplate)
	for reason, template := range r.errors {
		result[reason] = template
	}
	return result
}

// Exists 检查错误模板是否存在
func (r *Registry) Exists(reason string) bool {
	_, exists := r.Get(reason)
	return exists
}

// 全局注册器
var GlobalRegistry = NewRegistry()

// Register 在全局注册器中注册错误模板
func Register(reason string, code int32, i18nKey string) {
	GlobalRegistry.Register(reason, &ErrorTemplate{
		Code:    code,
		Reason:  reason,
		I18nKey: i18nKey,
	})
}

// MustCreate 使用全局注册器创建错误构建器
func MustCreate(reason string) *Builder {
	template := GlobalRegistry.MustGet(reason)
	builder := NewBuilder(template.Code, template.Reason)
	if template.I18nKey != "" {
		builder = builder.WithI18nKey(template.I18nKey)
	}
	return builder
}

// Create 使用全局注册器创建错误构建器，如果模板不存在则返回 nil
func Create(reason string) *Builder {
	template, exists := GlobalRegistry.Get(reason)
	if !exists {
		return nil
	}
	builder := NewBuilder(template.Code, template.Reason)
	if template.I18nKey != "" {
		builder = builder.WithI18nKey(template.I18nKey)
	}
	return builder
}

// GetTemplate 获取错误模板
func GetTemplate(reason string) (*ErrorTemplate, bool) {
	return GlobalRegistry.Get(reason)
}

// ListTemplates 列出所有错误模板
func ListTemplates() map[string]*ErrorTemplate {
	return GlobalRegistry.List()
}

// TemplateExists 检查错误模板是否存在
func TemplateExists(reason string) bool {
	return GlobalRegistry.Exists(reason)
}