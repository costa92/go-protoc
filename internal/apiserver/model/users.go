package model

import "time"

const TableNameUser = "users"

type User struct {
	ID            uint64     `gorm:"column:id" json:"id"`                           // 用户ID，主键自增
	Username      string     `gorm:"column:username" json:"username"`               // 用户名，唯一标识
	Email         string     `gorm:"column:email" json:"email"`                     // 邮箱地址
	PasswordHash  string     `gorm:"column:password_hash" json:"password_hash"`     // 密码哈希值
	Phone         *string    `gorm:"column:phone" json:"phone,omitempty"`           // 手机号码
	FullName      *string    `gorm:"column:full_name" json:"full_name,omitempty"`   // 用户全名
	AvatarURL     *string    `gorm:"column:avatar_url" json:"avatar_url,omitempty"` // 头像URL
	Status        int8       `gorm:"column:status" json:"status"`                   // 用户状态: 0=禁用, 1=启用, 2=锁定
	EmailVerified bool       `gorm:"column:email_verified" json:"email_verified"`   // 邮箱是否已验证
	PhoneVerified bool       `gorm:"column:phone_verified" json:"phone_verified"`   // 手机是否已验证
	LoginCount    uint32     `gorm:"column:login_count" json:"login_count"`         // 登录次数
	LastLoginAt   *time.Time `gorm:"column:last_login_at" json:"last_login_at"`     // 最后登录时间
	LastLoginIP   *string    `gorm:"column:last_login_ip" json:"last_login_ip"`     // 最后登录IP
	CreatedAt     time.Time  `gorm:"column:created_at" json:"created_at"`           // 创建时间
	UpdatedAt     time.Time  `gorm:"column:updated_at" json:"updated_at"`           // 更新时间
	DeletedAt     *time.Time `gorm:"column:deleted_at" json:"deleted_at,omitempty"` // 软删除时间
}

// TableName sets the insert table name for this struct type
func (User) TableName() string {
	return TableNameUser
}

// 用户状态常量
const (
	UserStatusDisabled int8 = 0 // 禁用
	UserStatusEnabled  int8 = 1 // 启用
	UserStatusLocked   int8 = 2 // 锁定
)
