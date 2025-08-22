package validate

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"

	"github.com/costa92/go-protoc/v2/pkg/api/errno"
)

// RequestValidator 定义了用于自定义验证的接口.
type RequestValidator interface {
	Validate(ctx context.Context, rq any) error
}

// Validator is a validator middleware.
func Validator(validator RequestValidator) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, rq any) (reply any, err error) {
			// Custom validation, specific to the API interface
			if err := validator.Validate(ctx, rq); err != nil {
				if se := new(errors.Error); errors.As(err, &se) {
					return nil, se
				}
				// If err is not a Kratos error, it's likely a validation error from
				// business logic (e.g., an i18n error).
				// We use its message as the primary message for the Kratos error.
				errMsg := err.Error()
				if errMsg == "" {
					errMsg = "validation failed" // Default message if original error message is empty
				}
				kratosValidationErr := errno.ErrorInvalidParameter("%s", errMsg)
				// Attach the original error as the cause for richer debugging information.
				return nil, kratosValidationErr.WithCause(err)
			}

			return handler(ctx, rq)
		}
	}
}
