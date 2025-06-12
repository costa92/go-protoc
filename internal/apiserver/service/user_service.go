package service

import (
	"context"

	userv1 "github.com/costa92/go-protoc/pkg/api/user/v1"
	"github.com/costa92/go-protoc/pkg/logger"
)

// UserService 实现 UserServiceServer 接口
type UserService struct {
	userv1.UnimplementedUserServiceServer
	logger logger.Logger
}

// NewUserService 创建一个新的 UserService 实例
func NewUserService(logger logger.Logger) *UserService {
	return &UserService{
		logger: logger,
	}
}

// CreateUser 实现创建用户接口
func (s *UserService) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	s.logger.Infow("收到创建用户请求", "username", req.GetUsername())
	// TODO: 实现实际的用户创建逻辑
	return &userv1.CreateUserResponse{
		User: &userv1.User{
			UserId:    "user-123",
			Username:  req.GetUsername(),
			Email:     req.GetEmail(),
			Age:       req.GetAge(),
			CreatedAt: 0,
			UpdatedAt: 0,
		},
	}, nil
}

// GetUser 实现获取用户接口
func (s *UserService) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	s.logger.Infow("收到获取用户请求", "user_id", req.GetUserId())
	// TODO: 实现实际的用户获取逻辑
	return &userv1.GetUserResponse{
		User: &userv1.User{
			UserId:    req.GetUserId(),
			Username:  "test_user",
			Email:     "test@example.com",
			Age:       25,
			CreatedAt: 0,
			UpdatedAt: 0,
		},
	}, nil
}

// UpdateUser 实现更新用户接口
func (s *UserService) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	s.logger.Infow("收到更新用户请求", "user_id", req.GetUserId())
	// TODO: 实现实际的用户更新逻辑
	return &userv1.UpdateUserResponse{
		User: &userv1.User{
			UserId:    req.GetUserId(),
			Username:  req.GetUsername(),
			Email:     req.GetEmail(),
			Age:       req.GetAge(),
			CreatedAt: 0,
			UpdatedAt: 0,
		},
	}, nil
}

// DeleteUser 实现删除用户接口
func (s *UserService) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) (*userv1.DeleteUserResponse, error) {
	s.logger.Infow("收到删除用户请求", "user_id", req.GetUserId())
	// TODO: 实现实际的用户删除逻辑
	return &userv1.DeleteUserResponse{
		Success: true,
	}, nil
}

// ListUsers 实现列出用户接口
func (s *UserService) ListUsers(ctx context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	s.logger.Infow("收到列出用户请求", "page", req.GetPage(), "page_size", req.GetPageSize())
	// TODO: 实现实际的用户列表逻辑
	return &userv1.ListUsersResponse{
		Users: []*userv1.User{
			{
				UserId:    "user-1",
				Username:  "user1",
				Email:     "user1@example.com",
				Age:       25,
				CreatedAt: 0,
				UpdatedAt: 0,
			},
			{
				UserId:    "user-2",
				Username:  "user2",
				Email:     "user2@example.com",
				Age:       30,
				CreatedAt: 0,
				UpdatedAt: 0,
			},
		},
		Total: 2,
	}, nil
}
