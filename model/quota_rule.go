package model

import (
	"time"

	"gorm.io/gorm"
)

// CreditRuleType 计费规则类型
type CreditRuleType string

const (
	// CreditRuleTypePerRequest 按次计费
	CreditRuleTypePerRequest CreditRuleType = "per_request"
)

// CreditRule 积分扣除规则（每个模型一条）
type CreditRule struct {
	// 规则唯一标识，自增主键
	ID int `json:"id" gorm:"primaryKey"`

	// 关联的模型 ID（唯一）
	ModelID int `json:"model_id" gorm:"uniqueIndex;not null"`

	// 规则类型：per_request=按次计费
	RuleType CreditRuleType `json:"rule_type" gorm:"size:32;not null;default:'per_request'"`

	// 基础积分（每次请求扣除的积分数量）
	BaseCredits float64 `json:"base_credits" gorm:"not null;default:0"`

	// 规则描述
	Description string `json:"description,omitempty" gorm:"type:text"`

	// 规则状态：1=启用, 2=禁用
	Status int `json:"status" gorm:"default:1"`

	// 记录创建时间
	CreatedAt time.Time `json:"created_at"`

	// 记录最后更新时间
	UpdatedAt time.Time `json:"updated_at"`

	// 关联的参数组合积分映射
	Items []CreditRuleItem `json:"items,omitempty" gorm:"foreignKey:RuleID"`
}

// TableName 指定表名（原为 quota_rules）
func (CreditRule) TableName() string {
	return "credit_rules"
}

// CreditRuleItem 参数组合积分映射项（一组条件的 AND 命中部请求时，使用该积分）
type CreditRuleItem struct {
	// 项唯一标识，自增主键
	ID int `json:"id" gorm:"primaryKey"`

	// 关联的规则 ID
	RuleID int `json:"rule_id" gorm:"index;not null"`

	// 该组合命中部请求时对应的积分
	Credits float64 `json:"credits" gorm:"not null;default:0"`

	// 记录创建时间
	CreatedAt time.Time `json:"created_at"`

	// 记录最后更新时间
	UpdatedAt time.Time `json:"updated_at"`

	// 该组合的 AND 条件（至少 1 条，无上限）
	Conditions []CreditRuleCondition `json:"conditions,omitempty" gorm:"foreignKey:ItemID"`
}

// TableName 指定表名（原为 quota_rule_items）
func (CreditRuleItem) TableName() string {
	return "credit_rule_items"
}

// CreditRuleCondition 参数映射条件（属于某个映射项，全部 AND）
type CreditRuleCondition struct {
	// 条件唯一标识，自增主键
	ID int `json:"id" gorm:"primaryKey"`

	// 所属映射项 ID
	ItemID int `json:"item_id" gorm:"index;not null"`

	// 请求参数路径（如 "resolution", "quality", "model"）
	ParamPath string `json:"param_path" gorm:"size:255;not null"`

	// 参数值（如 "1k", "low", "gpt-4"）
	ParamValue string `json:"param_value" gorm:"size:255;not null"`

	// 记录创建时间
	CreatedAt time.Time `json:"created_at"`

	// 记录最后更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (CreditRuleCondition) TableName() string {
	return "credit_rule_conditions"
}

// GetCreditRuleByModelID 获取指定模型的积分规则（含参数组合映射）
func GetCreditRuleByModelID(modelID int) (*CreditRule, error) {
	var rule CreditRule
	err := DB.Preload("Items.Conditions").Where("model_id = ? AND status = ?", modelID, 1).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// GetCreditRuleByID 根据 ID 获取积分规则
func GetCreditRuleByID(id int) (*CreditRule, error) {
	var rule CreditRule
	err := DB.Preload("Items.Conditions").Where("id = ?", id).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// CreateCreditRule 创建积分规则（含参数组合映射）
func CreateCreditRule(rule *CreditRule) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Items").Create(rule).Error; err != nil {
			return err
		}
		// 创建映射项及条件
		for i := range rule.Items {
			rule.Items[i].RuleID = rule.ID
			if err := tx.Omit("Conditions").Create(&rule.Items[i]).Error; err != nil {
				return err
			}
			for j := range rule.Items[i].Conditions {
				rule.Items[i].Conditions[j].ItemID = rule.Items[i].ID
			}
			if len(rule.Items[i].Conditions) > 0 {
				if err := tx.Create(&rule.Items[i].Conditions).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// UpdateCreditRule 更新积分规则
func UpdateCreditRule(id int, updates map[string]interface{}) error {
	return DB.Model(&CreditRule{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteCreditRule 删除积分规则（硬删除，会级联删除参数组合映射）
func DeleteCreditRule(id int) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		// 先取此项规则的所有映射项 id
		var itemIDs []int
		if err := tx.Model(&CreditRuleItem{}).Where("rule_id = ?", id).Pluck("id", &itemIDs).Error; err != nil {
			return err
		}
		// 删除映射项的条件
		if len(itemIDs) > 0 {
			if err := tx.Where("item_id IN ?", itemIDs).Delete(&CreditRuleCondition{}).Error; err != nil {
				return err
			}
		}
		// 删除映射项
		if err := tx.Where("rule_id = ?", id).Delete(&CreditRuleItem{}).Error; err != nil {
			return err
		}
		// 删除规则
		return tx.Delete(&CreditRule{}, id).Error
	})
}

// DeleteCreditRuleByModelID 删除指定模型的积分规则（硬删除，会级联删除参数组合映射）
func DeleteCreditRuleByModelID(modelID int) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		// 先获取规则 ID
		var rule CreditRule
		if err := tx.Where("model_id = ?", modelID).First(&rule).Error; err != nil {
			return err
		}
		return DeleteCreditRule(rule.ID)
	})
}

// ReplaceCreditRuleItems 替换规则的所有参数组合映射
func ReplaceCreditRuleItems(ruleID int, items []CreditRuleItem) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		// 先取旧项的 id，用于级联删除条件
		var oldIDs []int
		if err := tx.Model(&CreditRuleItem{}).Where("rule_id = ?", ruleID).Pluck("id", &oldIDs).Error; err != nil {
			return err
		}

		// 删除旧的映射项及其条件
		if len(oldIDs) > 0 {
			if err := tx.Where("item_id IN ?", oldIDs).Delete(&CreditRuleCondition{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("rule_id = ?", ruleID).Delete(&CreditRuleItem{}).Error; err != nil {
			return err
		}

		// 创建新的映射项及条件
		for i := range items {
			items[i].RuleID = ruleID
			items[i].ID = 0
			items[i].CreatedAt = time.Time{}
			items[i].UpdatedAt = time.Time{}
			if err := tx.Omit("Conditions").Create(&items[i]).Error; err != nil {
				return err
			}
			for j := range items[i].Conditions {
				items[i].Conditions[j].ItemID = items[i].ID
				items[i].Conditions[j].ID = 0
				items[i].Conditions[j].CreatedAt = time.Time{}
				items[i].Conditions[j].UpdatedAt = time.Time{}
			}
			if len(items[i].Conditions) > 0 {
				if err := tx.Create(&items[i].Conditions).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
