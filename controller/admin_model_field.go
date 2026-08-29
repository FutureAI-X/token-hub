package controller

import (
	"net/http"
	"strconv"

	"github.com/FutureAI/token-hub/model"
	"github.com/gin-gonic/gin"
)

// GetModelFields 获取模型的字段列表
func GetModelFields(c *gin.Context) {
	modelID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的模型ID"})
		return
	}

	section := c.Query("section") // 可选：request 或 response
	fields, err := model.GetModelFields(modelID, section)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "获取字段列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": fields})
}

// CreateModelFieldRequest 创建字段请求
type CreateModelFieldRequest struct {
	FieldKey    string `json:"field_key" binding:"required"`
	FieldName   string `json:"field_name" binding:"required"`
	FieldType   string `json:"field_type" binding:"required"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	Section     string `json:"section"` // request 或 response，默认 request
}

// CreateModelField 创建模型字段
func CreateModelField(c *gin.Context) {
	modelID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的模型ID"})
		return
	}

	var req CreateModelFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请填写完整信息"})
		return
	}

	validTypes := map[string]bool{
		"string": true, "number": true, "boolean": true, "array": true, "object": true,
	}
	if !validTypes[req.FieldType] {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "字段类型无效"})
		return
	}

	section := req.Section
	if section != "response" {
		section = "request"
	}

	field := model.ModelField{
		ModelID:     modelID,
		Section:     section,
		FieldKey:    req.FieldKey,
		FieldName:   req.FieldName,
		FieldType:   req.FieldType,
		Required:    req.Required,
		Description: req.Description,
		SortOrder:   req.SortOrder,
	}

	if err := model.CreateModelField(&field); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建字段失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "字段创建成功", "data": field})
}

// UpdateModelFieldRequest 更新字段请求
type UpdateModelFieldRequest struct {
	FieldKey    string `json:"field_key"`
	FieldName   string `json:"field_name"`
	FieldType   string `json:"field_type"`
	Required    *bool  `json:"required"`
	Description string `json:"description"`
	SortOrder   *int   `json:"sort_order"`
}

// UpdateModelField 更新模型字段
func UpdateModelField(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("fieldId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的字段ID"})
		return
	}

	var req UpdateModelFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数无效"})
		return
	}

	updates := map[string]interface{}{}
	if req.FieldKey != "" {
		updates["field_key"] = req.FieldKey
	}
	if req.FieldName != "" {
		updates["field_name"] = req.FieldName
	}
	if req.FieldType != "" {
		validTypes := map[string]bool{
			"string": true, "number": true, "boolean": true, "array": true, "object": true,
		}
		if !validTypes[req.FieldType] {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "字段类型无效"})
			return
		}
		updates["field_type"] = req.FieldType
	}
	if req.Required != nil {
		updates["required"] = *req.Required
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "没有需要更新的字段"})
		return
	}

	if err := model.UpdateModelField(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新字段失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "字段更新成功"})
}

// DeleteModelField 删除模型字段
func DeleteModelField(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("fieldId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的字段ID"})
		return
	}

	if err := model.DeleteModelField(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除字段失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "字段已删除"})
}
