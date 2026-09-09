package model

import (
	"errors"
	"time"
)

// ErrEndpointInUse 端点仍被模型引用
var ErrEndpointInUse = errors.New("该端点仍被模型引用，无法删除")

// Endpoint 端点定义
type Endpoint struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Path        string    `json:"path" gorm:"size:256;not null"`
	Name        string    `json:"name" gorm:"size:128;not null"`
	Description string    `json:"description" gorm:"size:512"`
	Status      int       `json:"status" gorm:"default:1"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AdminGetEndpoints 获取全部端点
func AdminGetEndpoints() ([]Endpoint, error) {
	var endpoints []Endpoint
	err := DB.Order("id ASC").Find(&endpoints).Error
	return endpoints, err
}

// GetEndpointByID 根据 ID 获取端点
func GetEndpointByID(id int) (*Endpoint, error) {
	var ep Endpoint
	err := DB.Where("id = ?", id).First(&ep).Error
	if err != nil {
		return nil, err
	}
	return &ep, nil
}

// GetEndpointByPath 根据路径获取端点
func GetEndpointByPath(path string) (*Endpoint, error) {
	var ep Endpoint
	err := DB.Where("path = ? AND status = ?", path, 1).First(&ep).Error
	if err != nil {
		return nil, err
	}
	return &ep, nil
}

// CreateEndpoint 创建端点
func CreateEndpoint(ep *Endpoint) error {
	return DB.Create(ep).Error
}

// UpdateEndpoint 更新端点
func UpdateEndpoint(id int, updates map[string]interface{}) error {
	return DB.Model(&Endpoint{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateEndpointStatus 更新端点状态
func UpdateEndpointStatus(id int, status int) error {
	return DB.Model(&Endpoint{}).Where("id = ?", id).Update("status", status).Error
}

// DeleteEndpoint 删除端点（物理删除；若仍被模型引用则拒绝）
func DeleteEndpoint(id int) error {
	// 检查是否仍被模型端点关联引用
	var count int64
	DB.Model(&ModelEndpoint{}).Where("endpoint_id = ?", id).Count(&count)
	if count > 0 {
		return ErrEndpointInUse
	}

	return DB.Delete(&Endpoint{}, id).Error
}
