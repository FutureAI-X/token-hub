package model

import (
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"github.com/FutureAI/token-hub/common"
	"gorm.io/gorm"
)

// 用户角色常量
const (
	RoleCommonUser = 1   // 普通用户
	RoleAdminUser  = 10  // 管理员
	RoleRootUser   = 100 // 超级管理员
)

// 用户状态常量
const (
	UserStatusEnabled  = 1 // 启用
	UserStatusDisabled = 2 // 禁用
)

// 错误定义
var (
	ErrInvalidCredentials  = errors.New("invalid username or password")
	ErrUserDisabled        = errors.New("user is disabled")
	ErrUserEmptyCredentials = errors.New("username or password is empty")
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

// GetTokensByUserID 获取用户的 Token 列表
func GetTokensByUserID(userID int) ([]Token, error) {
	var tokens []Token
	err := DB.Where("user_id = ?", userID).Order("id ASC").Find(&tokens).Error
	return tokens, err
}

// GetTokenByID 根据 ID 获取 Token
func GetTokenByID(id int) (*Token, error) {
	var token Token
	err := DB.Where("id = ?", id).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// CreateToken 创建 Token
func CreateToken(token *Token) error {
	return DB.Create(token).Error
}

// UpdateTokenName 更新 Token 名称
func UpdateTokenName(id int, name string) error {
	return DB.Model(&Token{}).Where("id = ?", id).Update("name", name).Error
}

// DeleteToken 删除 Token（软删除）
func DeleteToken(id int) error {
	return DB.Delete(&Token{}, id).Error
}

// IsTokenKeyExists 检查 Token Key 是否已存在
func IsTokenKeyExists(key string) bool {
	var count int64
	DB.Model(&Token{}).Where("key = ?", key).Count(&count)
	return count > 0
}

// GenerateUniqueTokenKey 生成唯一的 Token Key（sk- + 32位小写字母数字）
func GenerateUniqueTokenKey() (string, error) {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	for i := 0; i < 10; i++ { // 最多重试10次
		result := make([]byte, 32)
		for j := range result {
			n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
			result[j] = chars[n.Int64()]
		}
		key := "sk-" + string(result)
		if !IsTokenKeyExists(key) {
			return key, nil
		}
	}
	return "", errors.New("failed to generate unique token key")
}

// ValidateAndFill 验证用户名密码并填充用户信息
func (user *User) ValidateAndFill() error {
	password := user.Password
	if user.Username == "" || password == "" {
		return ErrUserEmptyCredentials
	}

	// 根据用户名查询用户
	err := DB.Where("username = ?", user.Username).First(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidCredentials
		}
		return err
	}

	// 验证密码
	if !common.ValidatePasswordAndHash(password, user.Password) {
		return ErrInvalidCredentials
	}

	// 检查用户状态
	if user.Status != UserStatusEnabled {
		return ErrUserDisabled
	}

	return nil
}

// GetUserByUsername 根据用户名获取用户
func GetUserByUsername(username string) (*User, error) {
	var user User
	err := DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID 根据ID获取用户
func GetUserByID(id int) (*User, error) {
	var user User
	err := DB.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUsers 分页查询用户列表，支持关键词搜索
func GetUsers(page, pageSize int, keyword string) ([]User, int64, error) {
	var users []User
	var total int64

	query := DB.Model(&User{})
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("username LIKE ? OR display_name LIKE ? OR email LIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id ASC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// AdminCreateUser 管理员创建用户
func AdminCreateUser(user *User) error {
	hashedPassword, err := common.Password2Hash(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword
	return DB.Create(user).Error
}

// AdminDeleteUser 管理员删除用户（软删除）
func AdminDeleteUser(id int) error {
	return DB.Delete(&User{}, id).Error
}

// UpdateUserStatus 更新用户状态
func UpdateUserStatus(id int, status int) error {
	return DB.Model(&User{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateUser 更新用户信息
func UpdateUser(id int, updates map[string]interface{}) error {
	return DB.Model(&User{}).Where("id = ?", id).Updates(updates).Error
}

// AdjustUserQuota 调整用户积分
func AdjustUserQuota(id int, mode string, value int64) error {
	user, err := GetUserByID(id)
	if err != nil {
		return err
	}

	switch mode {
	case "add":
		return DB.Model(user).Update("quota", user.Quota+value).Error
	case "subtract":
		newQuota := user.Quota - value
		if newQuota < 0 {
			newQuota = 0
		}
		return DB.Model(user).Update("quota", newQuota).Error
	case "override":
		return DB.Model(user).Update("quota", value).Error
	default:
		return errors.New("invalid mode: must be add, subtract, or override")
	}
}

// CreateRootUserIfNeed 创建 root 用户（如果不存在）
func CreateRootUserIfNeed() error {
	var count int64
	DB.Model(&User{}).Count(&count)
	if count > 0 {
		return nil
	}

	// 创建默认 root 用户
	hashedPassword, err := common.Password2Hash("123456")
	if err != nil {
		return err
	}

	rootUser := User{
		Username:    "root",
		Password:    hashedPassword,
		DisplayName: "Root User",
		Role:        RoleRootUser,
		Status:      UserStatusEnabled,
		Quota:       100000000,
	}

	if err := DB.Create(&rootUser).Error; err != nil {
		return err
	}

	common.SysLog("created default root user (username: root, password: 123456)")
	return nil
}
