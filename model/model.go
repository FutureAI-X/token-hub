package model

import (
	"time"

	"gorm.io/gorm"
)

// Model 模型信息
type Model struct {
	// 模型唯一标识，自增主键
	ID int `json:"id" gorm:"primaryKey"`

	// 模型名称，全局唯一，用于 API 调用
	Name string `json:"name" gorm:"uniqueIndex;size:64;not null"`

	// 模型描述
	Description string `json:"description,omitempty" gorm:"type:text"`

	// 模型图标标识
	Icon string `json:"icon,omitempty" gorm:"size:128"`

	// 模型标签，逗号分隔
	Tags string `json:"tags,omitempty" gorm:"size:255"`

	// 供应商 ID
	VendorID int `json:"vendor_id,omitempty" gorm:"index"`

	// 模型所有者/提供商名称（兼容旧数据）
	Owner string `json:"owner" gorm:"size:64;default:token-hub"`

	// 模型类型：text, image, video, audio
	ModelType string `json:"model_type" gorm:"size:32;default:'text'"`

	// 模型请求路径（转发到上游的实际路径）
	RequestPath string `json:"request_path" gorm:"size:256"`

	// 计费类型：0=按量计费(基于token), 1=按次计费
	QuotaType int `json:"quota_type" gorm:"default:0"`

	// 输入 token 倍率（按量计费时使用）
	ModelRatio float64 `json:"model_ratio" gorm:"default:1"`

	// 输出 token 倍率（相对于输入的倍数）
	CompletionRatio float64 `json:"completion_ratio" gorm:"default:2"`

	// 按次计费价格（quota_type=1 时使用，单位：分）
	ModelPrice float64 `json:"model_price" gorm:"default:0"`

	// 最大上下文长度
	ContextLength int `json:"context_length" gorm:"default:4096"`

	// 最大输出 token 数
	MaxOutputTokens int `json:"max_output_tokens" gorm:"default:4096"`

	// 模型状态：1=启用, 2=禁用
	Status int `json:"status" gorm:"default:1"`

	// 记录创建时间
	CreatedAt time.Time `json:"created_at"`

	// 记录最后更新时间
	UpdatedAt time.Time `json:"updated_at"`

	// 软删除时间戳
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 非数据库字段：供应商名称（查询时填充）
	VendorName string `json:"vendor_name,omitempty" gorm:"-"`
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

// GetPricingModels 获取所有可用模型（带供应商信息）
func GetPricingModels() ([]Model, error) {
	var models []Model
	err := DB.Where("status = ?", 1).
		Order("vendor_id ASC, name ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	// 填充供应商名称
	vendorMap, err := GetVendorMapByID()
	if err == nil && vendorMap != nil {
		for i := range models {
			if v, ok := vendorMap[models[i].VendorID]; ok {
				models[i].VendorName = v.Name
			}
		}
	}

	return models, nil
}

// AdminGetModels 管理员获取全部模型（含禁用）
func AdminGetModels() ([]Model, error) {
	var models []Model
	err := DB.Order("id ASC").Find(&models).Error
	return models, err
}

// GetModelByID 根据 ID 获取模型
func GetModelByID(id int) (*Model, error) {
	var m Model
	err := DB.Where("id = ?", id).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// CreateModel 创建模型
func CreateModel(m *Model) error {
	return DB.Create(m).Error
}

// UpdateModel 更新模型
func UpdateModel(id int, updates map[string]interface{}) error {
	return DB.Model(&Model{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateModelStatus 更新模型状态
func UpdateModelStatus(id int, status int) error {
	return DB.Model(&Model{}).Where("id = ?", id).Update("status", status).Error
}

// DeleteModel 删除模型（软删除）
func DeleteModel(id int) error {
	return DB.Delete(&Model{}, id).Error
}

// createDefaultModels 创建默认模型数据
func createDefaultModels() error {
	// 检查是否已有模型数据
	var count int64
	DB.Model(&Model{}).Count(&count)
	if count > 0 {
		return nil
	}

	// 获取供应商 ID 映射
	vendorMap, _ := GetVendorMap()

	getVendorID := func(name string) int {
		if v, ok := vendorMap[name]; ok {
			return v.ID
		}
		return 0
	}

	// 创建默认模型
	defaultModels := []Model{
		// DeepSeek
		{Name: "deepseek-chat", Description: "DeepSeek V3 对话模型，擅长中文理解和代码生成", Tags: "对话,代码", VendorID: getVendorID("DeepSeek"), QuotaType: 0, ModelRatio: 1, CompletionRatio: 2, ContextLength: 65536, MaxOutputTokens: 8192},
		{Name: "deepseek-reasoner", Description: "DeepSeek R1 推理模型，支持深度思考", Tags: "推理,思考", VendorID: getVendorID("DeepSeek"), QuotaType: 0, ModelRatio: 2, CompletionRatio: 4, ContextLength: 65536, MaxOutputTokens: 8192},

		// OpenAI
		{Name: "gpt-4o", Description: "OpenAI 多模态旗舰模型，支持文本和图像输入", Tags: "多模态,对话", VendorID: getVendorID("OpenAI"), QuotaType: 0, ModelRatio: 10, CompletionRatio: 3, ContextLength: 128000, MaxOutputTokens: 16384},
		{Name: "gpt-4o-mini", Description: "OpenAI 轻量模型，性价比极高", Tags: "轻量,快速", VendorID: getVendorID("OpenAI"), QuotaType: 0, ModelRatio: 0.6, CompletionRatio: 2.5, ContextLength: 128000, MaxOutputTokens: 16384},
		{Name: "o3-mini", Description: "OpenAI 推理模型，支持复杂推理任务", Tags: "推理", VendorID: getVendorID("OpenAI"), QuotaType: 0, ModelRatio: 4.4, CompletionRatio: 3, ContextLength: 200000, MaxOutputTokens: 100000},

		// Claude
		{Name: "claude-sonnet-4-20250514", Description: "Anthropic Claude 4 Sonnet，平衡性能与速度", Tags: "对话,代码", VendorID: getVendorID("Anthropic"), QuotaType: 0, ModelRatio: 12, CompletionRatio: 5, ContextLength: 200000, MaxOutputTokens: 16000},
		{Name: "claude-haiku-3-5", Description: "Anthropic 轻量模型，响应极快", Tags: "快速,轻量", VendorID: getVendorID("Anthropic"), QuotaType: 0, ModelRatio: 3.2, CompletionRatio: 4, ContextLength: 200000, MaxOutputTokens: 8192},

		// Gemini
		{Name: "gemini-2.5-flash", Description: "Google Gemini 2.5 Flash，高速多模态模型", Tags: "快速,多模态", VendorID: getVendorID("Google"), QuotaType: 0, ModelRatio: 0.6, CompletionRatio: 4, ContextLength: 1048576, MaxOutputTokens: 65536},
		{Name: "gemini-2.5-pro", Description: "Google Gemini 2.5 Pro，最强推理能力", Tags: "推理,多模态", VendorID: getVendorID("Google"), QuotaType: 0, ModelRatio: 5, CompletionRatio: 4, ContextLength: 1048576, MaxOutputTokens: 65536},
	}

	for _, m := range defaultModels {
		if m.Owner == "" {
			m.Owner = "token-hub"
		}
		if err := DB.Create(&m).Error; err != nil {
			return err
		}
	}

	return nil
}
