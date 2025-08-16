package handler

import (
	"context"
	"errors"

	v1 "github.com/costa92/go-protoc/v2/pkg/api/apiserver/v1"
	"github.com/costa92/go-protoc/v2/pkg/log"
)

func (h *Handler) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.GetUserResponse, error) {
	log.Errorw(errors.New("test code error"), "GetUser")
	// 从请求中获取用户ID
	return h.biz.UserV1().Get(ctx, req)
}

// CreateUser 创建用户
func (h *Handler) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
	return nil, nil
}
