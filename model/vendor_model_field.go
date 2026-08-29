package model

import (
	"time"

	"gorm.io/gorm"
)

// VendorModelField 供应商模型字段定义（独立于模型字段，支持从模型同步或手动新增）
type VendorModelField struct {
	ID            int        `json:"id" gorm:"primaryKey"`
	VendorModelID int        `json:"vendor_model_id" gorm:"index;not null"`
	ModelFieldID  *int       `json:"model_field_id" gorm:"index"` // 关联的模型字段ID，nil 表示手动新增
	FieldKey      string     `json:"field_key" gorm:"size:128;not null"`
	FieldName     string     `json:"field_name" gorm:"size:128;not null"`
	FieldType     string     `json:"field_type" gorm:"size:32;not null;default:'string'"`
	Required      bool       `json:"required" gorm:"default:false"`
	Description   string     `json:"description" gorm:"size:512"`
	SortOrder     int        `json:"sort_order" gorm:"default:0"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// GetVendorModelFields 获取供应商模型的字段列表
func GetVendorModelFields(vendorModelID int) ([]VendorModelField, error) {
	var fields []VendorModelField
	err := DB.Where("vendor_model_id = ?", vendorModelID).Order("sort_order ASC, id ASC").Find(&fields).Error
	return fields, err
}

// CreateVendorModelField 创建字段
func CreateVendorModelField(field *VendorModelField) error {
	return DB.Create(field).Error
}

// UpdateVendorModelField 更新字段
func UpdateVendorModelField(id int, updates map[string]interface{}) error {
	return DB.Model(&VendorModelField{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteVendorModelField 删除字段
func DeleteVendorModelField(id int) error {
	return DB.Delete(&VendorModelField{}, id).Error
}

// SyncFieldsFromModel 从模型同步字段到供应商模型（追加，不覆盖已有）
func SyncFieldsFromModel(vendorModelID int, modelID int) error {
	modelFields, err := GetModelFields(modelID)
	if err != nil {
		return err
	}

	// 获取已有的 field_key
	existing, _ := GetVendorModelFields(vendorModelID)
	existingKeys := make(map[string]bool)
	for _, f := range existing {
		existingKeys[f.FieldKey] = true
	}

	for _, mf := range modelFields {
		if existingKeys[mf.FieldKey] {
			continue // 跳过已存在的
		}
		mfID := mf.ID
		vmf := VendorModelField{
			VendorModelID: vendorModelID,
			ModelFieldID:  &mfID,
			FieldKey:      mf.FieldKey,
			FieldName:     mf.FieldName,
			FieldType:     mf.FieldType,
			Required:      mf.Required,
			Description:   mf.Description,
			SortOrder:     mf.SortOrder,
		}
		if err := DB.Create(&vmf).Error; err != nil {
			return err
		}
	}
	return nil
}
