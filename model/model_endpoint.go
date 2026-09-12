package model

import (
	"time"

	"gorm.io/gorm"
)

// ModelEndpoint 模型端点关联
// (model_id, endpoint_id) 唯一：重复关联会让「该模型是否支持某端点」的判断与
// 关联列表出现冗余行。
type ModelEndpoint struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	ModelID    int       `json:"model_id" gorm:"uniqueIndex:idx_model_endpoint;not null"`
	EndpointID int       `json:"endpoint_id" gorm:"uniqueIndex:idx_model_endpoint;not null"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// 非数据库字段
	EndpointPath string `json:"endpoint_path,omitempty" gorm:"-"`
	EndpointName string `json:"endpoint_name,omitempty" gorm:"-"`
}

// GetModelEndpoints 获取模型的端点关联列表
func GetModelEndpoints(modelID int) ([]ModelEndpoint, error) {
	var items []ModelEndpoint
	err := DB.Where("model_id = ?", modelID).Order("id ASC").Find(&items).Error
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

// SyncModelEndpoints 同步模型端点（全量替换，整体在一个事务内完成）
func SyncModelEndpoints(modelID int, endpointIDs []int) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("model_id = ?", modelID).Delete(&ModelEndpoint{}).Error; err != nil {
			return err
		}

		// 去重：同一 endpointID 重复出现会违反唯一索引
		seen := make(map[int]struct{}, len(endpointIDs))
		for _, eid := range endpointIDs {
			if _, dup := seen[eid]; dup {
				continue
			}
			seen[eid] = struct{}{}

			me := ModelEndpoint{
				ModelID:    modelID,
				EndpointID: eid,
			}
			if err := tx.Create(&me).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
