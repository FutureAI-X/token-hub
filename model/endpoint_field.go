package model

import (
	"time"

	"gorm.io/gorm"
)

// EndpointField 端点字段定义（请求体/响应体）
type EndpointField struct {
	ID          int    `json:"id" gorm:"primaryKey"`
	EndpointID  int    `json:"endpoint_id" gorm:"index;not null"`
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

// GetEndpointFields 获取端点字段列表
func GetEndpointFields(endpointID int, section ...string) ([]EndpointField, error) {
	var fields []EndpointField
	query := DB.Where("endpoint_id = ?", endpointID)
	if len(section) > 0 && section[0] != "" {
		query = query.Where("section = ?", section[0])
	}
	err := query.Order("sort_order ASC, id ASC").Find(&fields).Error
	return fields, err
}

// CreateEndpointField 创建端点字段
func CreateEndpointField(field *EndpointField) error {
	return DB.Create(field).Error
}

// UpdateEndpointField 更新端点字段
func UpdateEndpointField(id int, updates map[string]interface{}) error {
	return DB.Model(&EndpointField{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteEndpointField 删除端点字段
func DeleteEndpointField(id int) error {
	return DB.Delete(&EndpointField{}, id).Error
}
