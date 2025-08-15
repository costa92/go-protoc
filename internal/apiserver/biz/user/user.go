package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/costa92/go-protoc/v2/internal/apiserver/model"
	"github.com/costa92/go-protoc/v2/internal/apiserver/store"
	v1 "github.com/costa92/go-protoc/v2/pkg/api/apiserver/v1"
	"gorm.io/gorm"
)

type UserBiz interface {

	// List retrieves a list of users and their total count based on the provided request parameters.
	CreateUser(ctx context.Context, rq *v1.CreateUserRequest) (*v1.CreateUserResponse, error)

	// Get retrieves the details of a specific user based on the provided request parameters.
	Get(ctx context.Context, rq *v1.GetUserRequest) (*v1.GetUserResponse, error)
}

// userBiz is the implementation of the UserBiz.
type userBiz struct {
	store store.IStore
}

// Ensure that *userBiz implements the UserBiz.
var _ UserBiz = (*userBiz)(nil)

// New creates and returns a new instance of *userBiz.
func New(store store.IStore) *userBiz {
	return &userBiz{store: store}
}

func (u *userBiz) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
	// 1. 数据验证
	if err := validateCreateUserRequest(req); err != nil {
		return nil, err
	}

	// 2. 创建用户模型
	user := &model.User{
		Username: req.Name,
		Email:    req.Email,
		Status:   model.UserStatusEnabled,
	}

	// 3. 在事务中创建用户
	err := u.store.TX(ctx, func(ctx context.Context) error {
		// 检查用户名是否已存在
		var count int64
		if err := u.store.DB(ctx).Model(&model.User{}).Where("username = ?", user.Username).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("username %s already exists", user.Username)
		}

		// 创建用户
		return u.store.DB(ctx).Create(user).Error
	})

	if err != nil {
		return nil, err
	}

	// 4. 返回响应
	return &v1.CreateUserResponse{
		Id: fmt.Sprintf("%d", user.ID),
	}, nil
}

func (u *userBiz) Get(ctx context.Context, req *v1.GetUserRequest) (*v1.GetUserResponse, error) {
	// 1. 数据验证
	if req.Id == "" {
		return nil, fmt.Errorf("invalid user id")
	}

	// 2. 查询用户
	var user model.User
	if err := u.store.DB(ctx).Where("id = ?", req.Id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	// 3. 返回响应
	return &v1.GetUserResponse{
		Id:    fmt.Sprintf("%d", user.ID),
		Name:  user.Username,
		Email: user.Email,
	}, nil
}

// validateCreateUserRequest validates the create user request
func validateCreateUserRequest(req *v1.CreateUserRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	return nil
}
