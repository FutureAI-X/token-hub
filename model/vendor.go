package model

import (
	"time"
)

// Vendor 供应商信息
type Vendor struct {
	// 唯一标识，自增主键
	ID int `json:"id" gorm:"primaryKey"`

	// 名称
	Name string `json:"name" gorm:"size:128;not null"`

	// 描述
	Description string `json:"description,omitempty" gorm:"type:text"`

	// API 基础地址
	BaseURL string `json:"base_url" gorm:"size:512"`

	// API 密钥
	APIKey string `json:"api_key" gorm:"size:512"`

	// 状态：1=启用, 2=禁用, 3=已删除
	Status int `json:"status" gorm:"default:1"`

	// 记录创建时间
	CreatedAt time.Time `json:"created_at"`

	// 记录最后更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// Vendor 状态常量
const (
	VendorStatusEnabled  = 1 // 启用
	VendorStatusDisabled = 2 // 禁用
	VendorStatusDeleted  = 3 // 已删除
)

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

// AdminGetVendors 管理员获取全部供应商（不含已删除）
func AdminGetVendors() ([]Vendor, error) {
	var vendors []Vendor
	err := DB.Where("status != ?", VendorStatusDeleted).Order("id ASC").Find(&vendors).Error
	return vendors, err
}

// GetVendorByID 根据 ID 获取供应商（不含已删除）
func GetVendorByID(id int) (*Vendor, error) {
	var vendor Vendor
	err := DB.Where("id = ? AND status != ?", id, VendorStatusDeleted).First(&vendor).Error
	if err != nil {
		return nil, err
	}
	return &vendor, nil
}

// GetEnabledVendorByNameInsensitive 根据名称获取启用的供应商（名称忽略大小写）
func GetEnabledVendorByNameInsensitive(name string) (*Vendor, error) {
	var vendor Vendor
	err := DB.Where("LOWER(name) = LOWER(?) AND status = ?", name, VendorStatusEnabled).First(&vendor).Error
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

// DeleteVendor 删除供应商（置为已删除状态，非物理删除）
func DeleteVendor(id int) error {
	return DB.Model(&Vendor{}).Where("id = ?", id).Update("status", VendorStatusDeleted).Error
}

