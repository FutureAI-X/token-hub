package model

import (
	"time"

	"gorm.io/gorm"
)

// ModelEndpoint 模型端点关联
type ModelEndpoint struct {
	ID         int        `json:"id" gorm:"primaryKey"`
	ModelID    int        `json:"model_id" gorm:"index;not null"`
	EndpointID int        `json:"endpoint_id" gorm:"index;not null"`
	Priority   int        `json:"priority" gorm:"default:0"` // 优先级，数值越小越优先
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// 非数据库字段
	EndpointPath string `json:"endpoint_path,omitempty" gorm:"-"`
	EndpointName string `json:"endpoint_name,omitempty" gorm:"-"`
}

// GetModelEndpoints 获取模型的端点关联列表
func GetModelEndpoints(modelID int) ([]ModelEndpoint, error) {
	var items []ModelEndpoint
	err := DB.Where("model_id = ?", modelID).Order("priority ASC, id ASC").Find(&items).Error
	if err != nil {
		return nil, err
	}

	epMap, _ := GetEndpointMapByID()
	for i := range items {
		if ep, ok := epMap[items[i].EndpointID]; ok {
			items[i].EndpointPath = ep.Path
			items[i].EndpointName = ep.Name
		}
	}
	return items, nil
}

// GetEndpointMapByID 获取端点 ID 映射
func GetEndpointMapByID() (map[int]Endpoint, error) {
	var endpoints []Endpoint
	err := DB.Find(&endpoints).Error
	if err != nil {
		return nil, err
	}
	m := make(map[int]Endpoint, len(endpoints))
	for _, ep := range endpoints {
		m[ep.ID] = ep
	}
	return m, nil
}

// SyncModelEndpoints 同步模型端点（全量替换）
func SyncModelEndpoints(modelID int, endpointIDs []int) error {
	if err := DB.Where("model_id = ?", modelID).Delete(&ModelEndpoint{}).Error; err != nil {
		return err
	}

	for i, eid := range endpointIDs {
		me := ModelEndpoint{
			ModelID:    modelID,
			EndpointID: eid,
			Priority:   i,
		}
		if err := DB.Create(&me).Error; err != nil {
			return err
		}
	}
	return nil
}
