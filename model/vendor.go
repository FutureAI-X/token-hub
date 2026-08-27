package model

import (
	"time"

	"gorm.io/gorm"
)

// Vendor 供应商信息
type Vendor struct {
	// 供应商唯一标识，自增主键
	ID int `json:"id" gorm:"primaryKey"`

	// 供应商名称，全局唯一
	Name string `json:"name" gorm:"uniqueIndex;size:128;not null"`

	// 供应商描述
	Description string `json:"description,omitempty" gorm:"type:text"`

	// 供应商图标标识
	Icon string `json:"icon,omitempty" gorm:"size:128"`

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

// createDefaultVendors 创建默认供应商数据
func createDefaultVendors() error {
	var count int64
	DB.Model(&Vendor{}).Count(&count)
	if count > 0 {
		return nil
	}

	defaultVendors := []Vendor{
		{Name: "OpenAI", Description: "GPT 系列模型提供商", Icon: "openai"},
		{Name: "Anthropic", Description: "Claude 系列模型提供商", Icon: "anthropic"},
		{Name: "Google", Description: "Gemini 系列模型提供商", Icon: "google"},
		{Name: "DeepSeek", Description: "DeepSeek 系列模型提供商", Icon: "deepseek"},
	}

	for _, v := range defaultVendors {
		if err := DB.Create(&v).Error; err != nil {
			return err
		}
	}

	return nil
}
