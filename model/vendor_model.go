package model

import (
	"time"

	"gorm.io/gorm"
)

// VendorModel 供应商模型关联（供应商 + 模型 + 供应商侧的模型ID）
type VendorModel struct {
	ID              int    `json:"id" gorm:"primaryKey"`
	VendorID        int    `json:"vendor_id" gorm:"index;not null"`
	ModelID         int    `json:"model_id" gorm:"index;not null"`
	VendorModelID   string `json:"vendor_model_id" gorm:"size:256;not null"`
	Status          int    `json:"status" gorm:"default:1"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`

	// 非数据库字段：关联查询时填充
	VendorName      string `json:"vendor_name,omitempty" gorm:"-"`
	ModelName       string `json:"model_name,omitempty" gorm:"-"`
}

// AdminGetVendorModels 获取全部供应商模型
func AdminGetVendorModels() ([]VendorModel, error) {
	var items []VendorModel
	err := DB.Order("id ASC").Find(&items).Error
	if err != nil {
		return nil, err
	}

	vendorMap, _ := GetVendorMapByID()
	modelMap, _ := GetModelMapByID()

	for i := range items {
		if v, ok := vendorMap[items[i].VendorID]; ok {
			items[i].VendorName = v.Name
		}
		if m, ok := modelMap[items[i].ModelID]; ok {
			items[i].ModelName = m.Name
		}
	}
	return items, nil
}

// GetModelMapByID 获取模型 ID 到对象的映射
func GetModelMapByID() (map[int]Model, error) {
	var models []Model
	err := DB.Find(&models).Error
	if err != nil {
		return nil, err
	}
	m := make(map[int]Model, len(models))
	for _, v := range models {
		m[v.ID] = v
	}
	return m, nil
}

// GetVendorModelByID 根据 ID 获取供应商模型
func GetVendorModelByID(id int) (*VendorModel, error) {
	var vm VendorModel
	err := DB.Where("id = ?", id).First(&vm).Error
	if err != nil {
		return nil, err
	}
	return &vm, nil
}

// CreateVendorModel 创建供应商模型
func CreateVendorModel(vm *VendorModel) error {
	return DB.Create(vm).Error
}

// UpdateVendorModel 更新供应商模型
func UpdateVendorModel(id int, updates map[string]interface{}) error {
	return DB.Model(&VendorModel{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateVendorModelStatus 更新状态
func UpdateVendorModelStatus(id int, status int) error {
	return DB.Model(&VendorModel{}).Where("id = ?", id).Update("status", status).Error
}

// DeleteVendorModel 删除
func DeleteVendorModel(id int) error {
	return DB.Delete(&VendorModel{}, id).Error
}

// GetVendorModelsByModelID 根据模型ID获取可用的供应商模型列表
func GetVendorModelsByModelID(modelID int) ([]VendorModel, error) {
	var items []VendorModel
	err := DB.Where("model_id = ? AND status = ?", modelID, 1).Find(&items).Error
	return items, err
}
