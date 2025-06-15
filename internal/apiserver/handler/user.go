package handler

import (
	"context"

	v1 "github.com/costa92/go-protoc/v2/pkg/api/apiserver/v1"
)

func (h *Handler) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.GetUserResponse, error) {
	// TODO: 实现实际的业务逻辑
	return &v1.GetUserResponse{
		Id:    req.Id,
		Name:  "Test User",
		Email: "test@example.com",
	}, nil
}
