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

				return nil, errno.ErrorInvalidParameter("validation failed").WithCause(err)
			}

			return handler(ctx, rq)
		}
	}
}
