package validation

import (
	"context"
	"log"

	"github.com/costa92/go-protoc/v2/internal/apiserver/pkg/locales"
	v1 "github.com/costa92/go-protoc/v2/pkg/api/apiserver/v1"
	"github.com/costa92/go-protoc/v2/pkg/i18n"
	genericvalidation "github.com/costa92/go-protoc/v2/pkg/validation"
)

// ValidateUserRules validates the user rules.
func (v *Validator) ValidateUserRules() genericvalidation.Rules {
	return genericvalidation.Rules{}
}

// ValidateCreateUserRequest validates the fields of a CreateUserRequest.
func (v *Validator) ValidateCreateUserRequest(ctx context.Context, rq *v1.CreateUserRequest) error {
	log.Println("ValidateCreateUserRequest")

	// 设置默认值
	if rq.Name == "" {
		rq.Name = "default_user"
	}

	if rq.Email == "" {
		rq.Email = "default@example.com"
	}

	// 检查用户是否已存在
	return i18n.FromContext(ctx).E(locales.UserAlreadyExists)
}
