package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	// 用户唯一标识，自增主键
	ID int `json:"id" gorm:"primaryKey"`

	// 用户名，用于登录，全局唯一，长度限制32字符
	Username string `json:"username" gorm:"uniqueIndex;size:32;not null"`

	// 密码哈希值，使用 bcrypt 加密存储，JSON 序列化时忽略
	Password string `json:"-" gorm:"not null"`

	// 显示名称，用于界面展示，长度限制64字符
	DisplayName string `json:"display_name" gorm:"size:64"`

	// 用户角色：1=普通用户, 10=管理员, 100=root 超级管理员
	Role int `json:"role" gorm:"default:1"`

	// 用户状态：1=启用, 2=禁用
	Status int `json:"status" gorm:"default:1"`

	// 用户邮箱，用于通知和找回密码，长度限制64字符
	Email string `json:"email" gorm:"size:64"`

	// 用户总配额，单位为 tokens，0 表示无配额
	Quota int64 `json:"quota" gorm:"default:0"`

	// 已使用配额，单位为 tokens
	UsedQuota int64 `json:"used_quota" gorm:"default:0"`

	// 用户访问令牌，用于 API 认证，可为空
	AccessToken *string `json:"access_token" gorm:"size:256"`

	// 记录创建时间，自动设置
	CreatedAt time.Time `json:"created_at"`

	// 记录最后更新时间，自动更新
	UpdatedAt time.Time `json:"updated_at"`

	// 软删除时间戳，非空表示已删除
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Token API Token 模型
type Token struct {
	// Token 唯一标识，自增主键
	ID int `json:"id" gorm:"primaryKey"`

	// 所属用户ID，关联 users 表
	UserID int `json:"user_id" gorm:"index"`

	// Token 密钥，用于 API 认证，全局唯一
	Key string `json:"key" gorm:"uniqueIndex;size:64;not null"`

	// Token 名称，便于用户识别
	Name string `json:"name" gorm:"size:64"`

	// Token 状态：1=启用, 2=禁用
	Status int `json:"status" gorm:"default:1"`

	// 过期时间戳，-1 表示永不过期
	ExpiredTime int64 `json:"expired_time" gorm:"default:-1"`

	// 剩余配额，-1 表示无限制
	RemainQuota int64 `json:"remain_quota" gorm:"default:-1"`

	// 已使用配额
	UsedQuota int64 `json:"used_quota" gorm:"default:0"`

	// 记录创建时间
	CreatedAt time.Time `json:"created_at"`

	// 记录最后更新时间
	UpdatedAt time.Time `json:"updated_at"`

	// 软删除时间戳
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
