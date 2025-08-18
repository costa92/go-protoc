package handler

import (
	"context"
	"errors"

	v1 "github.com/costa92/go-protoc/v2/pkg/api/apiserver/v1"
	"github.com/costa92/go-protoc/v2/pkg/logger"
)

func (h *Handler) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.GetUserResponse, error) {
	// 使用简单错误，让 Zap 自动生成堆栈跟踪
	err := errors.New("business logic error")
	logger.Errorw("GetUser failed", "error", err.Error(), "user_id", req.Id)
	// 从请求中获取用户ID
	return h.biz.UserV1().Get(ctx, req)
}

// CreateUser 创建用户
func (h *Handler) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
	return nil, nil
}
