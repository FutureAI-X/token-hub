package model

import (
	"time"

	"gorm.io/gorm"
)

// VendorEndpointField 供应商端点字段定义
type VendorEndpointField struct {
	ID               int        `json:"id" gorm:"primaryKey"`
	VendorEndpointID int        `json:"vendor_endpoint_id" gorm:"index;not null"`
	EndpointFieldID  *int       `json:"endpoint_field_id" gorm:"index"` // 关联端点字段，nil=手动新增
	FieldKey         string     `json:"field_key" gorm:"size:128;not null"`
	FieldName        string     `json:"field_name" gorm:"size:128;not null"`
	FieldType        string     `json:"field_type" gorm:"size:32;not null;default:'string'"`
	Required         bool       `json:"required" gorm:"default:false"`
	Description      string     `json:"description" gorm:"size:512"`
	SortOrder        int        `json:"sort_order" gorm:"default:0"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
}

// GetVendorEndpointFields 获取供应商端点字段列表
func GetVendorEndpointFields(veID int) ([]VendorEndpointField, error) {
	var fields []VendorEndpointField
	err := DB.Where("vendor_endpoint_id = ?", veID).Order("sort_order ASC, id ASC").Find(&fields).Error
	return fields, err
}

// CreateVendorEndpointField 创建字段
func CreateVendorEndpointField(field *VendorEndpointField) error {
	return DB.Create(field).Error
}

// UpdateVendorEndpointField 更新字段
func UpdateVendorEndpointField(id int, updates map[string]interface{}) error {
	return DB.Model(&VendorEndpointField{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteVendorEndpointField 删除字段
func DeleteVendorEndpointField(id int) error {
	return DB.Delete(&VendorEndpointField{}, id).Error
}

// SyncVEFieldsFromEndpoint 从端点同步字段到供应商端点（追加）
func SyncVEFieldsFromEndpoint(veID int, endpointID int) error {
	epFields, err := GetEndpointFields(endpointID)
	if err != nil {
		return err
	}

	existing, _ := GetVendorEndpointFields(veID)
	existingKeys := make(map[string]bool)
	for _, f := range existing {
		existingKeys[f.FieldKey] = true
	}

	for _, ef := range epFields {
		if existingKeys[ef.FieldKey] {
			continue
		}
		efID := ef.ID
		vef := VendorEndpointField{
			VendorEndpointID: veID,
			EndpointFieldID:  &efID,
			FieldKey:         ef.FieldKey,
			FieldName:        ef.FieldName,
			FieldType:        ef.FieldType,
			Required:         ef.Required,
			Description:      ef.Description,
			SortOrder:        ef.SortOrder,
		}
		if err := DB.Create(&vef).Error; err != nil {
			return err
		}
	}
	return nil
}
