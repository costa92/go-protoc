package handler

import (
	"context"

	"github.com/costa92/go-protoc/v2/internal/apiserver/pkg/locales"
	v1 "github.com/costa92/go-protoc/v2/pkg/api/apiserver/v1"
	"github.com/costa92/go-protoc/v2/pkg/i18n"
	"github.com/costa92/go-protoc/v2/pkg/log"
)

func (h *Handler) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.GetUserResponse, error) {
	// 从请求中获取用户ID
	userID := req.Id

	// 验证用户ID是否为空
	if userID == "11" {
		return nil, i18n.FromContext(ctx).E(locales.RecordNotFound)
	}

	log.Infow("get user", "user_id", userID)

	// TODO: 实现实际的业务逻辑
	return &v1.GetUserResponse{
		Id:    req.Id,
		Name:  "Test User",
		Email: "test@example.com",
	}, nil
}
