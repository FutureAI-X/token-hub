package model

import (
	"time"

	"gorm.io/gorm"
)

// Model 模型信息
type Model struct {
	ID        int            `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"uniqueIndex;size:64;not null"`
	Owner     string         `json:"owner" gorm:"size:64;default:token-hub"`
	Status    int            `json:"status" gorm:"default:1"` // 1=启用, 2=禁用
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// GetModels 获取所有可用模型
func GetModels() ([]Model, error) {
	var models []Model
	err := DB.Where("status = ?", 1).Find(&models).Error
	return models, err
}

// GetModelByName 根据名称获取模型
func GetModelByName(name string) (*Model, error) {
	var model Model
	err := DB.Where("name = ? AND status = ?", name, 1).First(&model).Error
	if err != nil {
		return nil, err
	}
	return &model, nil
}

// createDefaultModels 创建默认模型数据
func createDefaultModels() error {
	// 检查是否已有模型数据
	var count int64
	DB.Model(&Model{}).Count(&count)
	if count > 0 {
		return nil
	}

	// 创建默认模型
	defaultModels := []Model{
		{Name: "deepseek-v4-flash", Owner: "token-hub"},
	}

	for _, m := range defaultModels {
		if err := DB.Create(&m).Error; err != nil {
			return err
		}
	}

	return nil
}
