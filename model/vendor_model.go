package model

import (
	"time"
)

// VendorModel 供应商模型关联（供应商 + 模型 + 供应商侧的模型ID）
// (vendor_id, model_id) 唯一：重复行会让 image.go 中「取第一个供应商」的路由
// 结果依赖物理行序，产生不确定的上游选择。
type VendorModel struct {
	ID            int       `json:"id" gorm:"primaryKey"`
	VendorID      int       `json:"vendor_id" gorm:"uniqueIndex:idx_vendor_model;not null"`
	ModelID       int       `json:"model_id" gorm:"uniqueIndex:idx_vendor_model;not null"`
	VendorModelID string    `json:"vendor_model_id" gorm:"size:256;not null"`
	Status        int       `json:"status" gorm:"default:1"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// 非数据库字段：关联查询时填充
	VendorName string `json:"vendor_name,omitempty" gorm:"-"`
	ModelName  string `json:"model_name,omitempty" gorm:"-"`
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

// DeleteVendorModel 删除（硬删除，直接从表中删除）
func DeleteVendorModel(id int) error {
	return DB.Delete(&VendorModel{}, id).Error
}

// GetVendorModelsByModelID 根据模型ID获取可用的供应商模型列表
func GetVendorModelsByModelID(modelID int) ([]VendorModel, error) {
	var items []VendorModel
	err := DB.Where("model_id = ? AND status = ?", modelID, 1).
		Order("id ASC"). // 显式排序：调用方取 [0] 作为实际路由目标，无序会导致结果不确定
		Find(&items).Error
	return items, err
}

// GetEnabledVendorModelsByModelID 获取模型下「关联启用 且 供应商本身也启用」的供应商模型。
// 仅过滤关联表状态是不够的——供应商被禁用后，其关联行仍是启用状态，
// 会导致流量继续打向已停用的供应商。
func GetEnabledVendorModelsByModelID(modelID int) ([]VendorModel, error) {
	var items []VendorModel
	err := DB.Model(&VendorModel{}).
		Joins("JOIN vendors ON vendors.id = vendor_models.vendor_id").
		Where("vendor_models.model_id = ? AND vendor_models.status = ? AND vendors.status = ?",
			modelID, VendorStatusEnabled, VendorStatusEnabled).
		Order("vendor_models.id ASC").
		Find(&items).Error
	return items, err
}
