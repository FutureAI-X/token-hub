package controller

import (
	"net/http"
	"strconv"

	"github.com/FutureAI/token-hub/model"
	"github.com/gin-gonic/gin"
)

// AdminGetVendorModels 获取供应商模型列表
func AdminGetVendorModels(c *gin.Context) {
	items, err := model.AdminGetVendorModels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "获取列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": items})
}

// AdminCreateVendorModelRequest 创建请求
type AdminCreateVendorModelRequest struct {
	VendorID      int    `json:"vendor_id" binding:"required"`
	ModelID       int    `json:"model_id" binding:"required"`
	VendorModelID string `json:"vendor_model_id" binding:"required"`
	RequestPath   string `json:"request_path"`
	IsAsync       bool   `json:"is_async"`
}

// AdminCreateVendorModel 创建供应商模型
func AdminCreateVendorModel(c *gin.Context) {
	var req AdminCreateVendorModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请填写完整信息"})
		return
	}

	vm := model.VendorModel{
		VendorID:      req.VendorID,
		ModelID:       req.ModelID,
		VendorModelID: req.VendorModelID,
		RequestPath:   req.RequestPath,
		IsAsync:       req.IsAsync,
		Status:        1,
	}

	if err := model.CreateVendorModel(&vm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "创建成功"})
}

// AdminUpdateVendorModelRequest 更新请求
type AdminUpdateVendorModelRequest struct {
	VendorID      int    `json:"vendor_id"`
	ModelID       int    `json:"model_id"`
	VendorModelID string `json:"vendor_model_id"`
	RequestPath   string `json:"request_path"`
	IsAsync       *bool  `json:"is_async"`
}

// AdminUpdateVendorModel 更新供应商模型
func AdminUpdateVendorModel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的ID"})
		return
	}

	var req AdminUpdateVendorModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数无效"})
		return
	}

	updates := map[string]interface{}{}
	if req.VendorID > 0 {
		updates["vendor_id"] = req.VendorID
	}
	if req.ModelID > 0 {
		updates["model_id"] = req.ModelID
	}
	if req.VendorModelID != "" {
		updates["vendor_model_id"] = req.VendorModelID
	}
	if req.RequestPath != "" {
		updates["request_path"] = req.RequestPath
	}
	if req.IsAsync != nil {
		updates["is_async"] = *req.IsAsync
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "没有需要更新的字段"})
		return
	}

	if err := model.UpdateVendorModel(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "更新成功"})
}

// AdminUpdateVendorModelStatus 更新状态
func AdminUpdateVendorModelStatus(c *gin.Context) {
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

	if err := model.UpdateVendorModelStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "状态已更新"})
}

// AdminDeleteVendorModel 删除
func AdminDeleteVendorModel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的ID"})
		return
	}

	if err := model.DeleteVendorModel(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已删除"})
}
