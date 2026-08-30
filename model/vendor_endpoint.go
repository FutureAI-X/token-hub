package model

import (
	"time"

	"gorm.io/gorm"
)

// VendorEndpoint 供应商端点（供应商 + 端点 + 供应商侧自定义配置）
type VendorEndpoint struct {
	ID          int        `json:"id" gorm:"primaryKey"`
	VendorID    int        `json:"vendor_id" gorm:"index;not null"`
	EndpointID  int        `json:"endpoint_id" gorm:"index;not null"`
	Path        string     `json:"path" gorm:"size:512"`         // 供应商侧路径，为空则用端点默认路径
	Name        string     `json:"name" gorm:"size:128"`         // 供应商侧名称，为空则用端点默认名称
	Description string     `json:"description" gorm:"size:512"`  // 供应商侧描述
	IsAsync     bool       `json:"is_async" gorm:"default:false"`
	Status      int        `json:"status" gorm:"default:1"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 非数据库字段
	VendorName    string `json:"vendor_name" gorm:"-"`
	EndpointPath  string `json:"endpoint_path" gorm:"-"`
	EndpointName  string `json:"endpoint_name" gorm:"-"`
}

// AdminGetVendorEndpoints 获取全部供应商端点
func AdminGetVendorEndpoints() ([]VendorEndpoint, error) {
	var items []VendorEndpoint
	err := DB.Order("id ASC").Find(&items).Error
	if err != nil {
		return nil, err
	}

	vendorMap, _ := GetVendorMapByID()
	epMap, _ := GetEndpointMapByID()

	for i := range items {
		if v, ok := vendorMap[items[i].VendorID]; ok {
			items[i].VendorName = v.Name
		}
		if ep, ok := epMap[items[i].EndpointID]; ok {
			items[i].EndpointPath = ep.Path
			items[i].EndpointName = ep.Name
		}
	}
	return items, nil
}

// GetVendorEndpointByID 根据 ID 获取供应商端点
func GetVendorEndpointByID(id int) (*VendorEndpoint, error) {
	var ve VendorEndpoint
	err := DB.Where("id = ?", id).First(&ve).Error
	if err != nil {
		return nil, err
	}
	return &ve, nil
}

// CreateVendorEndpoint 创建供应商端点
func CreateVendorEndpoint(ve *VendorEndpoint) error {
	return DB.Create(ve).Error
}

// UpdateVendorEndpoint 更新供应商端点
func UpdateVendorEndpoint(id int, updates map[string]interface{}) error {
	return DB.Model(&VendorEndpoint{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateVendorEndpointStatus 更新状态
func UpdateVendorEndpointStatus(id int, status int) error {
	return DB.Model(&VendorEndpoint{}).Where("id = ?", id).Update("status", status).Error
}

// DeleteVendorEndpoint 删除
func DeleteVendorEndpoint(id int) error {
	return DB.Delete(&VendorEndpoint{}, id).Error
}
