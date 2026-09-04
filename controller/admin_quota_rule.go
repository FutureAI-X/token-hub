package controller

import (
	"math"
	"net/http"
	"strconv"

	"github.com/FutureAI/token-hub/model"
	"github.com/gin-gonic/gin"
)

// validateDecimalPlaces 验证小数位数不超过2位
func validateDecimalPlaces(value float64) bool {
	rounded := math.Round(value*100) / 100
	return math.Abs(value-rounded) < 1e-9
}

// AdminGetQuotaRule 获取模型的积分规则
func AdminGetQuotaRule(c *gin.Context) {
	modelID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的模型ID"})
		return
	}

	rule, err := model.GetQuotaRuleByModelID(modelID)
	if err != nil {
		// 没有找到规则返回空
		c.JSON(http.StatusOK, gin.H{"success": true, "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": rule})
}

// AdminSaveQuotaRuleRequest 保存积分规则请求
type AdminSaveQuotaRuleRequest struct {
	RuleType    string                   `json:"rule_type" binding:"required"`
	BasePrice   float64                  `json:"base_price"`
	Description string                   `json:"description"`
	Items       []QuotaRuleItemRequest   `json:"items"`
}

// QuotaRuleItemRequest 参数映射项请求
type QuotaRuleItemRequest struct {
	ParamPath  string  `json:"param_path" binding:"required"`
	ParamValue string  `json:"param_value" binding:"required"`
	Price      float64 `json:"price"`
}

// AdminSaveQuotaRule 保存积分规则（创建或更新）
func AdminSaveQuotaRule(c *gin.Context) {
	modelID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的模型ID"})
		return
	}

	var req AdminSaveQuotaRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请填写完整信息"})
		return
	}

	// 验证规则类型
	validTypes := map[string]bool{"per_request": true}
	if !validTypes[req.RuleType] {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "规则类型无效，目前仅支持 per_request"})
		return
	}

	// 验证基础价格
	if req.BasePrice <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "基础价格必须大于 0"})
		return
	}

	// 验证小数位数不超过2位
	if !validateDecimalPlaces(req.BasePrice) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "基础价格最多支持2位小数"})
		return
	}

	// 验证模型是否存在
	_, err = model.GetModelByID(modelID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "模型不存在"})
		return
	}

	// 验证参数映射项
	for i, item := range req.Items {
		if item.ParamPath == "" || item.ParamValue == "" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数路径和参数值不能为空"})
			return
		}
		if item.Price <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "第 " + strconv.Itoa(i+1) + " 项的价格必须大于 0"})
			return
		}
		if !validateDecimalPlaces(item.Price) {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "第 " + strconv.Itoa(i+1) + " 项的价格最多支持2位小数"})
			return
		}
	}

	// 查找现有规则
	existingRule, _ := model.GetQuotaRuleByModelID(modelID)

	if existingRule != nil {
		// 更新现有规则
		updates := map[string]interface{}{
			"rule_type":    req.RuleType,
			"base_price":   req.BasePrice,
			"description": req.Description,
		}
		if err := model.UpdateQuotaRule(existingRule.ID, updates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新积分规则失败"})
			return
		}

		// 替换参数映射
		items := make([]model.QuotaRuleItem, len(req.Items))
		for i, item := range req.Items {
			items[i] = model.QuotaRuleItem{
				ParamPath:  item.ParamPath,
				ParamValue: item.ParamValue,
				Price:      item.Price,
			}
		}
		if err := model.ReplaceQuotaRuleItems(existingRule.ID, items); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新参数映射失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "积分规则更新成功"})
	} else {
		// 创建新规则
		rule := &model.QuotaRule{
			ModelID:     modelID,
			RuleType:    model.QuotaRuleType(req.RuleType),
			BasePrice:   req.BasePrice,
			Description: req.Description,
			Status:      1,
		}

		// 构建参数映射
		rule.Items = make([]model.QuotaRuleItem, len(req.Items))
		for i, item := range req.Items {
			rule.Items[i] = model.QuotaRuleItem{
				ParamPath:  item.ParamPath,
				ParamValue: item.ParamValue,
				Price:      item.Price,
			}
		}

		if err := model.CreateQuotaRule(rule); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建积分规则失败: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "积分规则创建成功"})
	}
}

// AdminDeleteQuotaRule 删除积分规则
func AdminDeleteQuotaRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的规则ID"})
		return
	}

	if err := model.DeleteQuotaRule(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除积分规则失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "积分规则已删除"})
}

// AdminDeleteModelQuotaRule 删除模型的积分规则
func AdminDeleteModelQuotaRule(c *gin.Context) {
	modelID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的模型ID"})
		return
	}

	if err := model.DeleteQuotaRuleByModelID(modelID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除积分规则失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "积分规则已删除"})
}

// AdminUpdateQuotaRuleStatus 更新积分规则状态
func AdminUpdateQuotaRuleStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的规则ID"})
		return
	}

	var req struct {
		Status int `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "状态值无效"})
		return
	}

	if req.Status != 1 && req.Status != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "状态值必须为 1(启用) 或 2(禁用)"})
		return
	}

	if err := model.UpdateQuotaRule(id, map[string]interface{}{"status": req.Status}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "状态已更新"})
}
