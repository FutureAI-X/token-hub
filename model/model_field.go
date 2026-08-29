package model

import (
	"time"

	"gorm.io/gorm"
)

// ModelField 模型字段定义（请求体/响应体）
type ModelField struct {
	ID          int    `json:"id" gorm:"primaryKey"`
	ModelID     int    `json:"model_id" gorm:"index;not null"`
	Section     string `json:"section" gorm:"size:32;not null;default:'request'"` // request 或 response
	FieldKey    string `json:"field_key" gorm:"size:128;not null"`
	FieldName   string `json:"field_name" gorm:"size:128;not null"`
	FieldType   string `json:"field_type" gorm:"size:32;not null;default:'string'"`
	Required    bool   `json:"required" gorm:"default:false"`
	Description string `json:"description" gorm:"size:512"`
	SortOrder   int    `json:"sort_order" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// GetModelFields 获取模型的字段列表（可选 section 过滤）
func GetModelFields(modelID int, section ...string) ([]ModelField, error) {
	var fields []ModelField
	query := DB.Where("model_id = ?", modelID)
	if len(section) > 0 && section[0] != "" {
		query = query.Where("section = ?", section[0])
	}
	err := query.Order("sort_order ASC, id ASC").Find(&fields).Error
	return fields, err
}

// CreateModelField 创建模型字段
func CreateModelField(field *ModelField) error {
	return DB.Create(field).Error
}

// UpdateModelField 更新模型字段
func UpdateModelField(id int, updates map[string]interface{}) error {
	return DB.Model(&ModelField{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteModelField 删除模型字段
func DeleteModelField(id int) error {
	return DB.Delete(&ModelField{}, id).Error
}

// DeleteModelFieldsByModelID 删除模型的所有字段
func DeleteModelFieldsByModelID(modelID int) error {
	return DB.Where("model_id = ?", modelID).Delete(&ModelField{}).Error
}
