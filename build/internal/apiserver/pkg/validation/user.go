package validation

import (
	"context"
	"log"

	v1 "github.com/costa92/go-protoc/v2/pkg/api/apiserver/v1"
	"github.com/costa92/go-protoc/v2/pkg/errors"
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

	// 模拟检查用户是否已存在的逻辑
	// 这里可以添加实际的数据库查询逻辑
	if rq.Name == "existing_user" {
		return errors.NewUserAlreadyExistsError(rq.Name)
	}

	return nil
}
