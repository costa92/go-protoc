package http

import (
	"net/http"

	"github.com/costa92/go-protoc/pkg/log"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

var validate = validator.New()

type validationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message"`
}

func (v *validationError) Error() string {
	return v.Message
}

// ValidationMiddleware 创建一个验证中间件
func ValidationMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				log.L().Errorf("Failed to parse form: %v", err)
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			if err := validate.Struct(r); err != nil {
				log.L().Errorf("Validation failed: %v", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// NewValidator 创建一个新的验证器实例
func NewValidator() *validator.Validate {
	return validate
}
