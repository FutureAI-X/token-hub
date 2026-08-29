package controller

import (
	"net/http"
	"strconv"

	"github.com/FutureAI/token-hub/model"
	"github.com/gin-gonic/gin"
)

// GetVendorModelFields 获取供应商模型的字段列表
func GetVendorModelFields(c *gin.Context) {
	vmID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的ID"})
		return
	}

	fields, err := model.GetVendorModelFields(vmID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "获取字段失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": fields})
}

// SyncFieldsFromEndpoint 从端点同步字段
func SyncFieldsFromEndpoint(c *gin.Context) {
	vmID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的ID"})
		return
	}

	var req struct {
		EndpointID int `json:"endpoint_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请提供 endpoint_id"})
		return
	}

	if err := model.SyncFieldsFromEndpoint(vmID, req.EndpointID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "同步失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "字段同步成功"})
}

// CreateVMFieldRequest 创建字段请求
type CreateVMFieldRequest struct {
	FieldKey         string `json:"field_key" binding:"required"`
	FieldName        string `json:"field_name" binding:"required"`
	FieldType        string `json:"field_type" binding:"required"`
	Required         bool   `json:"required"`
	Description      string `json:"description"`
	EndpointFieldID  *int   `json:"endpoint_field_id"` // 可选，关联端点字段
}

// CreateVMField 手动创建字段
func CreateVMField(c *gin.Context) {
	vmID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的ID"})
		return
	}

	var req CreateVMFieldRequest
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

	field := model.VendorModelField{
		VendorModelID:   vmID,
		EndpointFieldID: req.EndpointFieldID,
		FieldKey:      req.FieldKey,
		FieldName:     req.FieldName,
		FieldType:     req.FieldType,
		Required:      req.Required,
		Description:   req.Description,
	}

	if err := model.CreateVendorModelField(&field); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "字段创建成功", "data": field})
}

// UpdateVMFieldRequest 更新字段请求
type UpdateVMFieldRequest struct {
	FieldKey         string `json:"field_key"`
	FieldName        string `json:"field_name"`
	FieldType        string `json:"field_type"`
	Required         *bool  `json:"required"`
	Description      string `json:"description"`
	EndpointFieldID  *int   `json:"endpoint_field_id"`
}

// UpdateVMField 更新字段
func UpdateVMField(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("fieldId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的字段ID"})
		return
	}

	var req UpdateVMFieldRequest
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
	if req.EndpointFieldID != nil {
		if *req.EndpointFieldID == -1 {
			updates["endpoint_field_id"] = nil // 取消绑定
		} else {
			updates["endpoint_field_id"] = req.EndpointFieldID
		}
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "没有需要更新的字段"})
		return
	}

	if err := model.UpdateVendorModelField(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "更新成功"})
}

// DeleteVMField 删除字段
func DeleteVMField(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("fieldId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的字段ID"})
		return
	}

	if err := model.DeleteVendorModelField(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已删除"})
}
