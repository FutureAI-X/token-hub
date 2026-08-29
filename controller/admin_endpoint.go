package controller

import (
	"net/http"
	"strconv"

	"github.com/FutureAI/token-hub/model"
	"github.com/gin-gonic/gin"
)

// AdminGetEndpoints 获取端点列表
func AdminGetEndpoints(c *gin.Context) {
	endpoints, err := model.AdminGetEndpoints()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "获取端点列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": endpoints})
}

// AdminCreateEndpointRequest 创建端点请求
type AdminCreateEndpointRequest struct {
	Path        string `json:"path" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// AdminCreateEndpoint 创建端点
func AdminCreateEndpoint(c *gin.Context) {
	var req AdminCreateEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请填写完整信息"})
		return
	}

	ep := model.Endpoint{
		Path:        req.Path,
		Name:        req.Name,
		Description: req.Description,
		Status:      1,
	}

	if err := model.CreateEndpoint(&ep); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "端点创建成功"})
}

// AdminUpdateEndpointRequest 更新端点请求
type AdminUpdateEndpointRequest struct {
	Path        string `json:"path"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AdminUpdateEndpoint 更新端点
func AdminUpdateEndpoint(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的ID"})
		return
	}

	var req AdminUpdateEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数无效"})
		return
	}

	updates := map[string]interface{}{}
	if req.Path != "" {
		updates["path"] = req.Path
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "没有需要更新的字段"})
		return
	}

	if err := model.UpdateEndpoint(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "更新成功"})
}

// AdminUpdateEndpointStatus 更新端点状态
func AdminUpdateEndpointStatus(c *gin.Context) {
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

	if err := model.UpdateEndpointStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "状态已更新"})
}

// AdminDeleteEndpoint 删除端点
func AdminDeleteEndpoint(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的ID"})
		return
	}

	if err := model.DeleteEndpoint(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已删除"})
}

// ── 端点字段 ──

// GetEndpointFields 获取端点字段列表
func GetEndpointFields(c *gin.Context) {
	epID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的ID"})
		return
	}

	section := c.Query("section")
	fields, err := model.GetEndpointFields(epID, section)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "获取字段失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": fields})
}

// CreateEndpointFieldRequest 创建端点字段请求
type CreateEndpointFieldRequest struct {
	FieldKey    string `json:"field_key" binding:"required"`
	FieldName   string `json:"field_name" binding:"required"`
	FieldType   string `json:"field_type" binding:"required"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	Section     string `json:"section"`
}

// CreateEndpointField 创建端点字段
func CreateEndpointField(c *gin.Context) {
	epID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的ID"})
		return
	}

	var req CreateEndpointFieldRequest
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

	field := model.EndpointField{
		EndpointID:  epID,
		Section:     section,
		FieldKey:    req.FieldKey,
		FieldName:   req.FieldName,
		FieldType:   req.FieldType,
		Required:    req.Required,
		Description: req.Description,
		SortOrder:   req.SortOrder,
	}

	if err := model.CreateEndpointField(&field); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "字段创建成功", "data": field})
}

// UpdateEndpointFieldRequest 更新端点字段请求
type UpdateEndpointFieldRequest struct {
	FieldKey    string `json:"field_key"`
	FieldName   string `json:"field_name"`
	FieldType   string `json:"field_type"`
	Required    *bool  `json:"required"`
	Description string `json:"description"`
	SortOrder   *int   `json:"sort_order"`
}

// UpdateEndpointField 更新端点字段
func UpdateEndpointField(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("fieldId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的字段ID"})
		return
	}

	var req UpdateEndpointFieldRequest
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

	if err := model.UpdateEndpointField(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "更新成功"})
}

// DeleteEndpointField 删除端点字段
func DeleteEndpointField(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("fieldId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的字段ID"})
		return
	}

	if err := model.DeleteEndpointField(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已删除"})
}
