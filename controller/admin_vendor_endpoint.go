package controller

import (
	"net/http"
	"strconv"

	"github.com/FutureAI/token-hub/model"
	"github.com/gin-gonic/gin"
)

// ── 供应商端点 CRUD ──

// AdminGetVendorEndpoints 获取供应商端点列表
func AdminGetVendorEndpoints(c *gin.Context) {
	items, err := model.AdminGetVendorEndpoints()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "获取列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": items})
}

// AdminCreateVendorEndpointRequest 创建请求
type AdminCreateVendorEndpointRequest struct {
	VendorID    int    `json:"vendor_id" binding:"required"`
	EndpointID  int    `json:"endpoint_id" binding:"required"`
	Path        string `json:"path"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsAsync     bool   `json:"is_async"`
}

// AdminCreateVendorEndpoint 创建供应商端点
func AdminCreateVendorEndpoint(c *gin.Context) {
	var req AdminCreateVendorEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请填写完整信息"})
		return
	}

	ve := model.VendorEndpoint{
		VendorID:    req.VendorID,
		EndpointID:  req.EndpointID,
		Path:        req.Path,
		Name:        req.Name,
		Description: req.Description,
		IsAsync:     req.IsAsync,
		Status:      1,
	}

	if err := model.CreateVendorEndpoint(&ve); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "创建成功"})
}

// AdminUpdateVendorEndpointRequest 更新请求
type AdminUpdateVendorEndpointRequest struct {
	VendorID    int    `json:"vendor_id"`
	EndpointID  int    `json:"endpoint_id"`
	Path        string `json:"path"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsAsync     *bool  `json:"is_async"`
}

// AdminUpdateVendorEndpoint 更新供应商端点
func AdminUpdateVendorEndpoint(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的ID"})
		return
	}

	var req AdminUpdateVendorEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数无效"})
		return
	}

	updates := map[string]interface{}{}
	if req.VendorID > 0 {
		updates["vendor_id"] = req.VendorID
	}
	if req.EndpointID > 0 {
		updates["endpoint_id"] = req.EndpointID
	}
	if req.Path != "" {
		updates["path"] = req.Path
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.IsAsync != nil {
		updates["is_async"] = *req.IsAsync
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "没有需要更新的字段"})
		return
	}

	if err := model.UpdateVendorEndpoint(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "更新成功"})
}

// AdminUpdateVendorEndpointStatus 更新状态
func AdminUpdateVendorEndpointStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的ID"})
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
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "状态值必须为 1 或 2"})
		return
	}

	if err := model.UpdateVendorEndpointStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "状态已更新"})
}

// AdminDeleteVendorEndpoint 删除
func AdminDeleteVendorEndpoint(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的ID"})
		return
	}

	if err := model.DeleteVendorEndpoint(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已删除"})
}

// ── 供应商端点字段 ──

// GetVendorEndpointFields 获取供应商端点字段列表
func GetVendorEndpointFields(c *gin.Context) {
	veID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的ID"})
		return
	}

	fields, err := model.GetVendorEndpointFields(veID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "获取字段失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": fields})
}

// SyncVEFieldsFromEndpoint 从端点同步字段
func SyncVEFieldsFromEndpoint(c *gin.Context) {
	veID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的ID"})
		return
	}

	ve, err := model.GetVendorEndpointByID(veID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "供应商端点不存在"})
		return
	}

	if err := model.SyncVEFieldsFromEndpoint(veID, ve.EndpointID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "同步失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "字段同步成功"})
}

// CreateVEFieldRequest 创建字段请求
type CreateVEFieldRequest struct {
	FieldKey        string `json:"field_key" binding:"required"`
	FieldName       string `json:"field_name" binding:"required"`
	FieldType       string `json:"field_type" binding:"required"`
	Required        bool   `json:"required"`
	Description     string `json:"description"`
	EndpointFieldID *int   `json:"endpoint_field_id"`
}

// CreateVEField 创建字段
func CreateVEField(c *gin.Context) {
	veID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的ID"})
		return
	}

	var req CreateVEFieldRequest
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

	field := model.VendorEndpointField{
		VendorEndpointID: veID,
		EndpointFieldID:  req.EndpointFieldID,
		FieldKey:         req.FieldKey,
		FieldName:        req.FieldName,
		FieldType:        req.FieldType,
		Required:         req.Required,
		Description:      req.Description,
	}

	if err := model.CreateVendorEndpointField(&field); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "字段创建成功", "data": field})
}

// UpdateVEFieldRequest 更新字段请求
type UpdateVEFieldRequest struct {
	FieldKey        string `json:"field_key"`
	FieldName       string `json:"field_name"`
	FieldType       string `json:"field_type"`
	Required        *bool  `json:"required"`
	Description     string `json:"description"`
	EndpointFieldID *int   `json:"endpoint_field_id"`
}

// UpdateVEField 更新字段
func UpdateVEField(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("fieldId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的字段ID"})
		return
	}

	var req UpdateVEFieldRequest
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
			updates["endpoint_field_id"] = nil
		} else {
			updates["endpoint_field_id"] = req.EndpointFieldID
		}
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "没有需要更新的字段"})
		return
	}

	if err := model.UpdateVendorEndpointField(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "更新成功"})
}

// DeleteVEField 删除字段
func DeleteVEField(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("fieldId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的字段ID"})
		return
	}

	if err := model.DeleteVendorEndpointField(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已删除"})
}
