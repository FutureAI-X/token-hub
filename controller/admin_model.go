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
		ID           int     `json:"id"`
		Name         string  `json:"name"`
		Owner        string  `json:"owner"`
		ModelType    string  `json:"model_type"`
		RequestPath  string  `json:"request_path"`
		Description  string  `json:"description"`
		QuotaType    int     `json:"quota_type"`
		ModelRatio   float64 `json:"model_ratio"`
		CompletionRatio float64 `json:"completion_ratio"`
		ContextLength    int   `json:"context_length"`
		MaxOutputTokens  int   `json:"max_output_tokens"`
		Status       int     `json:"status"`
		CreatedAt    string  `json:"created_at"`
	}

	items := make([]modelResponse, len(models))
	for i, m := range models {
		items[i] = modelResponse{
			ID:              m.ID,
			Name:            m.Name,
			Owner:           m.Owner,
			ModelType:       m.ModelType,
			RequestPath:     m.RequestPath,
			Description:     m.Description,
			QuotaType:       m.QuotaType,
			ModelRatio:      m.ModelRatio,
			CompletionRatio: m.CompletionRatio,
			ContextLength:   m.ContextLength,
			MaxOutputTokens: m.MaxOutputTokens,
			Status:          m.Status,
			CreatedAt:       m.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": items})
}

// AdminCreateModelRequest 创建模型请求
type AdminCreateModelRequest struct {
	Name            string  `json:"name" binding:"required"`
	Owner           string  `json:"owner" binding:"required"`
	ModelType       string  `json:"model_type" binding:"required"`
	RequestPath     string  `json:"request_path"`
	Description     string  `json:"description"`
	ContextLength   int     `json:"context_length"`
	MaxOutputTokens int     `json:"max_output_tokens"`
}

// AdminCreateModel 创建模型
func AdminCreateModel(c *gin.Context) {
	var req AdminCreateModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请填写完整信息"})
		return
	}

	validTypes := map[string]bool{"text": true, "image": true, "video": true, "audio": true}
	if !validTypes[req.ModelType] {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "模型类型无效，必须为 text/image/video/audio"})
		return
	}

	if req.ContextLength == 0 {
		req.ContextLength = 4096
	}
	if req.MaxOutputTokens == 0 {
		req.MaxOutputTokens = 4096
	}

	m := model.Model{
		Name:            req.Name,
		Owner:           req.Owner,
		ModelType:       req.ModelType,
		RequestPath:     req.RequestPath,
		Description:     req.Description,
		ContextLength:   req.ContextLength,
		MaxOutputTokens: req.MaxOutputTokens,
		Status:          1,
	}

	if err := model.CreateModel(&m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建模型失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "模型创建成功"})
}

// AdminUpdateModelRequest 更新模型请求
type AdminUpdateModelRequest struct {
	Name            string  `json:"name"`
	Owner           string  `json:"owner"`
	ModelType       string  `json:"model_type"`
	RequestPath     string  `json:"request_path"`
	Description     string  `json:"description"`
	ContextLength   int     `json:"context_length"`
	MaxOutputTokens int     `json:"max_output_tokens"`
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
	if req.ModelType != "" {
		validTypes := map[string]bool{"text": true, "image": true, "video": true, "audio": true}
		if !validTypes[req.ModelType] {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "模型类型无效"})
			return
		}
		updates["model_type"] = req.ModelType
	}
	if req.RequestPath != "" {
		updates["request_path"] = req.RequestPath
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.ContextLength > 0 {
		updates["context_length"] = req.ContextLength
	}
	if req.MaxOutputTokens > 0 {
		updates["max_output_tokens"] = req.MaxOutputTokens
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
