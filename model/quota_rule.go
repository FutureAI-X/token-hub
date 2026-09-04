package model

import (
	"time"

	"gorm.io/gorm"
)

// QuotaRuleType 计费规则类型
type QuotaRuleType string

const (
	// QuotaRuleTypePerRequest 按次计费
	QuotaRuleTypePerRequest QuotaRuleType = "per_request"
)

// QuotaRule 积分扣除规则（每个模型一条）
type QuotaRule struct {
	// 规则唯一标识，自增主键
	ID int `json:"id" gorm:"primaryKey"`

	// 关联的模型 ID（唯一）
	ModelID int `json:"model_id" gorm:"uniqueIndex;not null"`

	// 规则类型：per_request=按次计费
	RuleType QuotaRuleType `json:"rule_type" gorm:"size:32;not null;default:'per_request'"`

	// 基础积分价格（每次请求扣除的积分数量）
	BasePrice float64 `json:"base_price" gorm:"not null;default:0"`

	// 规则描述
	Description string `json:"description,omitempty" gorm:"type:text"`

	// 规则状态：1=启用, 2=禁用
	Status int `json:"status" gorm:"default:1"`

	// 记录创建时间
	CreatedAt time.Time `json:"created_at"`

	// 记录最后更新时间
	UpdatedAt time.Time `json:"updated_at"`

	// 软删除时间戳
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联的参数价格映射
	Items []QuotaRuleItem `json:"items,omitempty" gorm:"foreignKey:RuleID"`
}

// QuotaRuleItem 参数价格映射项
type QuotaRuleItem struct {
	// 项唯一标识，自增主键
	ID int `json:"id" gorm:"primaryKey"`

	// 关联的规则 ID
	RuleID int `json:"rule_id" gorm:"index;not null"`

	// 请求参数路径（如 "size", "quality", "model"）
	ParamPath string `json:"param_path" gorm:"size:255;not null"`

	// 参数值（如 "1024x1024", "high", "gpt-4"）
	ParamValue string `json:"param_value" gorm:"size:255;not null"`

	// 该参数值对应的积分价格
	Price float64 `json:"price" gorm:"not null;default:0"`

	// 记录创建时间
	CreatedAt time.Time `json:"created_at"`

	// 记录最后更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// GetQuotaRuleByModelID 获取指定模型的积分规则（包含参数映射）
func GetQuotaRuleByModelID(modelID int) (*QuotaRule, error) {
	var rule QuotaRule
	err := DB.Preload("Items").Where("model_id = ? AND status = ?", modelID, 1).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// GetQuotaRuleByID 根据 ID 获取积分规则
func GetQuotaRuleByID(id int) (*QuotaRule, error) {
	var rule QuotaRule
	err := DB.Preload("Items").Where("id = ?", id).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// CreateQuotaRule 创建积分规则（包含参数映射）
func CreateQuotaRule(rule *QuotaRule) error {
	return DB.Create(rule).Error
}

// UpdateQuotaRule 更新积分规则
func UpdateQuotaRule(id int, updates map[string]interface{}) error {
	return DB.Model(&QuotaRule{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteQuotaRule 删除积分规则（软删除，会级联删除参数映射）
func DeleteQuotaRule(id int) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		// 先删除参数映射
		if err := tx.Where("rule_id = ?", id).Delete(&QuotaRuleItem{}).Error; err != nil {
			return err
		}
		// 再删除规则
		return tx.Delete(&QuotaRule{}, id).Error
	})
}

// DeleteQuotaRuleByModelID 删除指定模型的积分规则
func DeleteQuotaRuleByModelID(modelID int) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		// 先获取规则 ID
		var rule QuotaRule
		if err := tx.Where("model_id = ?", modelID).First(&rule).Error; err != nil {
			return err
		}
		// 删除参数映射
		if err := tx.Where("rule_id = ?", rule.ID).Delete(&QuotaRuleItem{}).Error; err != nil {
			return err
		}
		// 删除规则
		return tx.Delete(&QuotaRule{}, rule.ID).Error
	})
}

// ReplaceQuotaRuleItems 替换规则的所有参数映射
func ReplaceQuotaRuleItems(ruleID int, items []QuotaRuleItem) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		// 删除旧的映射
		if err := tx.Where("rule_id = ?", ruleID).Delete(&QuotaRuleItem{}).Error; err != nil {
			return err
		}
		// 创建新的映射
		if len(items) > 0 {
			for i := range items {
				items[i].RuleID = ruleID
			}
			return tx.Create(&items).Error
		}
		return nil
	})
}
