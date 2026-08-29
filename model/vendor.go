package model

import (
	"time"

	"gorm.io/gorm"
)

// Vendor 供应商信息
type Vendor struct {
	// 供应商唯一标识，自增主键
	ID int `json:"id" gorm:"primaryKey"`

	// 供应商名称
	Name string `json:"name" gorm:"size:128;not null"`

	// 供应商描述
	Description string `json:"description,omitempty" gorm:"type:text"`

	// 供应商图标标识
	Icon string `json:"icon,omitempty" gorm:"size:128"`

	// API 基础地址
	BaseURL string `json:"base_url" gorm:"size:512"`

	// API 密钥
	APIKey string `json:"api_key" gorm:"size:512"`

	// 协议类型：openai-chat, openai-responses, anthropic-messages
	ProtocolType string `json:"protocol_type" gorm:"size:64;default:'openai-chat'"`

	// 供应商状态：1=启用, 2=禁用
	Status int `json:"status" gorm:"default:1"`

	// 记录创建时间
	CreatedAt time.Time `json:"created_at"`

	// 记录最后更新时间
	UpdatedAt time.Time `json:"updated_at"`

	// 软删除时间戳
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// GetVendors 获取所有启用的供应商
func GetVendors() ([]Vendor, error) {
	var vendors []Vendor
	err := DB.Where("status = ?", 1).Order("id ASC").Find(&vendors).Error
	return vendors, err
}

// GetVendorMap 获取供应商名称到对象的映射
func GetVendorMap() (map[string]Vendor, error) {
	var vendors []Vendor
	err := DB.Where("status = ?", 1).Find(&vendors).Error
	if err != nil {
		return nil, err
	}
	m := make(map[string]Vendor, len(vendors))
	for _, v := range vendors {
		m[v.Name] = v
	}
	return m, nil
}

// GetVendorMapByID 获取供应商 ID 到对象的映射
func GetVendorMapByID() (map[int]Vendor, error) {
	var vendors []Vendor
	err := DB.Where("status = ?", 1).Find(&vendors).Error
	if err != nil {
		return nil, err
	}
	m := make(map[int]Vendor, len(vendors))
	for _, v := range vendors {
		m[v.ID] = v
	}
	return m, nil
}

// AdminGetVendors 管理员获取全部供应商（含禁用）
func AdminGetVendors() ([]Vendor, error) {
	var vendors []Vendor
	err := DB.Order("id ASC").Find(&vendors).Error
	return vendors, err
}

// GetVendorByID 根据 ID 获取供应商
func GetVendorByID(id int) (*Vendor, error) {
	var vendor Vendor
	err := DB.Where("id = ?", id).First(&vendor).Error
	if err != nil {
		return nil, err
	}
	return &vendor, nil
}

// CreateVendor 创建供应商
func CreateVendor(vendor *Vendor) error {
	return DB.Create(vendor).Error
}

// UpdateVendor 更新供应商
func UpdateVendor(id int, updates map[string]interface{}) error {
	return DB.Model(&Vendor{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateVendorStatus 更新供应商状态
func UpdateVendorStatus(id int, status int) error {
	return DB.Model(&Vendor{}).Where("id = ?", id).Update("status", status).Error
}

// DeleteVendor 删除供应商（软删除）
func DeleteVendor(id int) error {
	return DB.Delete(&Vendor{}, id).Error
}

