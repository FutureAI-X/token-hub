package controller

import (
	"net/http"
	"strconv"

	"github.com/FutureAI/token-hub/model"
	"github.com/gin-gonic/gin"
)

// AdminGetModels 管理员获取模型列表
func AdminGetModels(c *gin.Context) {
	models, err := model.AdminGetModels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "获取模型列表失败"})
		return
	}

	type modelResponse struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Owner       string `json:"owner"`
		Description string `json:"description"`
		Tags        string `json:"tags"`
		Status      int    `json:"status"`
		CreatedAt   string `json:"created_at"`
	}

	items := make([]modelResponse, len(models))
	for i, m := range models {
		items[i] = modelResponse{
			ID:          m.ID,
			Name:        m.Name,
			Owner:       m.Owner,
			Description: m.Description,
			Tags:        m.Tags,
			Status:      m.Status,
			CreatedAt:   m.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": items})
}

// AdminCreateModelRequest 创建模型请求
type AdminCreateModelRequest struct {
	Name        string `json:"name" binding:"required"`
	Owner       string `json:"owner" binding:"required"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
}

// AdminCreateModel 创建模型
func AdminCreateModel(c *gin.Context) {
	var req AdminCreateModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请填写完整信息"})
		return
	}

	m := model.Model{
		Name:        req.Name,
		Owner:       req.Owner,
		Description: req.Description,
		Tags:        req.Tags,
		Status:      1,
	}

	if err := model.CreateModel(&m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建模型失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "模型创建成功"})
}

// AdminUpdateModelRequest 更新模型请求
type AdminUpdateModelRequest struct {
	Name        string `json:"name"`
	Owner       string `json:"owner"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
}

// AdminUpdateModel 更新模型
func AdminUpdateModel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的模型ID"})
		return
	}

	var req AdminUpdateModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数无效"})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Owner != "" {
		updates["owner"] = req.Owner
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Tags != "" {
		updates["tags"] = req.Tags
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "没有需要更新的字段"})
		return
	}

	if err := model.UpdateModel(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新模型失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "模型更新成功"})
}

// AdminUpdateModelStatus 更新模型状态
func AdminUpdateModelStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的模型ID"})
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

	if err := model.UpdateModelStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "状态已更新"})
}

// AdminDeleteModel 删除模型
func AdminDeleteModel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的模型ID"})
		return
	}

	if err := model.DeleteModel(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除模型失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "模型已删除"})
}
