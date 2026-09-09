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

// AdminGetCreditRule 获取模型的积分规则
func AdminGetCreditRule(c *gin.Context) {
	modelID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的模型ID"})
		return
	}

	rule, err := model.GetCreditRuleByModelID(modelID)
	if err != nil {
		// 没有找到规则返回空
		c.JSON(http.StatusOK, gin.H{"success": true, "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": rule})
}

// AdminSaveCreditRuleRequest 保存积分规则请求
type AdminSaveCreditRuleRequest struct {
	RuleType    string                  `json:"rule_type" binding:"required"`
	BaseCredits float64                 `json:"base_credits"`
	Description string                  `json:"description"`
	Items       []CreditRuleItemRequest `json:"items"`
}

// CreditRuleItemRequest 参数组合映射项请求（一组 AND 条件 → 积分）
type CreditRuleItemRequest struct {
	Credits    float64                       `json:"credits"`
	Conditions []CreditRuleConditionRequest  `json:"conditions"`
}

// CreditRuleConditionRequest 参数映射条件请求
type CreditRuleConditionRequest struct {
	ParamPath  string `json:"param_path" binding:"required"`
	ParamValue string `json:"param_value" binding:"required"`
}

// AdminSaveCreditRule 保存积分规则（创建或更新）
func AdminSaveCreditRule(c *gin.Context) {
	modelID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的模型ID"})
		return
	}

	var req AdminSaveCreditRuleRequest
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

	// 验证基础积分
	if req.BaseCredits <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "基础积分必须大于 0"})
		return
	}

	// 验证小数位数不超过2位
	if !validateDecimalPlaces(req.BaseCredits) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "基础积分最多支持2位小数"})
		return
	}

	// 验证模型是否存在
	_, err = model.GetModelByID(modelID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "模型不存在"})
		return
	}

	// 验证参数组合映射项
	for i, item := range req.Items {
		if item.Credits <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "第 " + strconv.Itoa(i+1) + " 项的积分必须大于 0"})
			return
		}
		if !validateDecimalPlaces(item.Credits) {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "第 " + strconv.Itoa(i+1) + " 项的积分最多支持2位小数"})
			return
		}
		// 至少 1 个条件，无上限
		if len(item.Conditions) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "第 " + strconv.Itoa(i+1) + " 项至少需要 1 个参数条件"})
			return
		}
		for j, cond := range item.Conditions {
			if cond.ParamPath == "" || cond.ParamValue == "" {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "第 " + strconv.Itoa(i+1) + " 项的第 " + strconv.Itoa(j+1) + " 个条件参数路径和参数值不能为空"})
				return
			}
		}
	}

	// 查找现有规则
	existingRule, _ := model.GetCreditRuleByModelID(modelID)

	if existingRule != nil {
		// 更新现有规则
		updates := map[string]interface{}{
			"rule_type":     req.RuleType,
			"base_credits":  req.BaseCredits,
			"description":   req.Description,
		}
		if err := model.UpdateCreditRule(existingRule.ID, updates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新积分规则失败"})
			return
		}

		// 替换参数组合映射
		items := buildCreditRuleItems(req.Items)
		if err := model.ReplaceCreditRuleItems(existingRule.ID, items); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新参数映射失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "积分规则更新成功"})
	} else {
		// 创建新规则
		rule := &model.CreditRule{
			ModelID:     modelID,
			RuleType:    model.CreditRuleType(req.RuleType),
			BaseCredits: req.BaseCredits,
			Description: req.Description,
			Status:      1,
		}

		// 构建参数组合映射
		rule.Items = buildCreditRuleItems(req.Items)

		if err := model.CreateCreditRule(rule); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建积分规则失败: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "积分规则创建成功"})
	}
}

// buildCreditRuleItems 将请求的组合映射转换为模型切片
func buildCreditRuleItems(items []CreditRuleItemRequest) []model.CreditRuleItem {
	result := make([]model.CreditRuleItem, len(items))
	for i, item := range items {
		result[i] = model.CreditRuleItem{
			Credits: item.Credits,
		}
		conds := make([]model.CreditRuleCondition, len(item.Conditions))
		for j, cond := range item.Conditions {
			conds[j] = model.CreditRuleCondition{
				ParamPath:  cond.ParamPath,
				ParamValue: cond.ParamValue,
			}
		}
		result[i].Conditions = conds
	}
	return result
}

// AdminDeleteCreditRule 删除积分规则
func AdminDeleteCreditRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的规则ID"})
		return
	}

	if err := model.DeleteCreditRule(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除积分规则失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "积分规则已删除"})
}

// AdminDeleteModelCreditRule 删除模型的积分规则
func AdminDeleteModelCreditRule(c *gin.Context) {
	modelID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的模型ID"})
		return
	}

	if err := model.DeleteCreditRuleByModelID(modelID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除积分规则失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "积分规则已删除"})
}

// AdminUpdateCreditRuleStatus 更新积分规则状态
func AdminUpdateCreditRuleStatus(c *gin.Context) {
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

	if err := model.UpdateCreditRule(id, map[string]interface{}{"status": req.Status}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "状态已更新"})
}
