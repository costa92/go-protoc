package http

import (
	"encoding/json"
	"net/http"

	"github.com/costa92/go-protoc/pkg/logger"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

// Validator 定义验证接口
type Validator interface {
	Validate() error
}

var validate = validator.New()

// ValidationMiddleware 创建一个验证中间件
func ValidationMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var err error

			// 根据 Content-Type 处理不同类型的请求
			contentType := r.Header.Get("Content-Type")
			switch contentType {
			case "application/json":
				if r.Body == nil {
					http.Error(w, "请求体不能为空", http.StatusBadRequest)
					return
				}
				defer r.Body.Close()

				var v Validator
				if err = json.NewDecoder(r.Body).Decode(&v); err != nil {
					logger.L().Errorf("解析 JSON 失败: %v", err)
					http.Error(w, "无效的 JSON 格式", http.StatusBadRequest)
					return
				}

				if err = v.Validate(); err != nil {
					logger.L().Errorf("验证失败: %v", err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

			default:
				// 处理表单数据
				if err = r.ParseForm(); err != nil {
					logger.L().Errorf("解析表单失败: %v", err)
					http.Error(w, "无效的表单数据", http.StatusBadRequest)
					return
				}

				if err = validate.Struct(r.Form); err != nil {
					logger.L().Errorf("验证失败: %v", err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// NewValidator 创建一个新的验证器实例
func NewValidator() *validator.Validate {
	return validate
}

// RegisterCustomValidation 注册自定义验证规则
func RegisterCustomValidation(tag string, fn validator.Func) error {
	return validate.RegisterValidation(tag, fn)
}
