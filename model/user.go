package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           int            `json:"id" gorm:"primaryKey"`
	Username     string         `json:"username" gorm:"uniqueIndex;size:32;not null"`
	Password     string         `json:"-" gorm:"not null"`
	DisplayName  string         `json:"display_name" gorm:"size:64"`
	Role         int            `json:"role" gorm:"default:1"` // 1=普通用户, 10=管理员, 100=root
	Status       int            `json:"status" gorm:"default:1"` // 1=启用, 2=禁用
	Email        string         `json:"email" gorm:"size:64"`
	Quota        int64          `json:"quota" gorm:"default:0"`
	UsedQuota    int64          `json:"used_quota" gorm:"default:0"`
	AccessToken  *string        `json:"access_token" gorm:"size:256"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// Token API Token 模型
type Token struct {
	ID          int            `json:"id" gorm:"primaryKey"`
	UserID      int            `json:"user_id" gorm:"index"`
	Key         string         `json:"key" gorm:"uniqueIndex;size:64;not null"`
	Name        string         `json:"name" gorm:"size:64"`
	Status      int            `json:"status" gorm:"default:1"` // 1=启用, 2=禁用
	ExpiredTime int64          `json:"expired_time" gorm:"default:-1"` // -1=永不过期
	RemainQuota int64          `json:"remain_quota" gorm:"default:-1"` // -1=无限制
	UsedQuota   int64          `json:"used_quota" gorm:"default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}
