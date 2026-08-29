package controller

import (
	"net/http"
	"strconv"

	"github.com/FutureAI/token-hub/common"
	"github.com/FutureAI/token-hub/model"
	"github.com/gin-gonic/gin"
)

// AdminGetVendors 管理员获取供应商列表
func AdminGetVendors(c *gin.Context) {
	vendors, err := model.AdminGetVendors()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "获取供应商列表失败"})
		return
	}

	type vendorResponse struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		Description  string `json:"description"`
		BaseURL      string `json:"base_url"`
		APIKey       string `json:"api_key"`
		ProtocolType string `json:"protocol_type"`
		Status       int    `json:"status"`
		CreatedAt    string `json:"created_at"`
	}

	items := make([]vendorResponse, len(vendors))
	for i, v := range vendors {
		apiKey := ""
		if v.APIKey != "" {
			if decrypted, err := common.DecryptSecret(v.APIKey); err == nil {
				apiKey = decrypted
			}
		}
		items[i] = vendorResponse{
			ID:           v.ID,
			Name:         v.Name,
			Description:  v.Description,
			BaseURL:      v.BaseURL,
			APIKey:       apiKey,
			ProtocolType: v.ProtocolType,
			Status:       v.Status,
			CreatedAt:    v.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
	})
}

// AdminCreateVendorRequest 创建供应商请求
type AdminCreateVendorRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	BaseURL      string `json:"base_url" binding:"required"`
	APIKey       string `json:"api_key" binding:"required"`
	ProtocolType string `json:"protocol_type" binding:"required"`
	DataKey      string `json:"data_key" binding:"required"`
}

// AdminCreateVendor 创建供应商
func AdminCreateVendor(c *gin.Context) {
	var req AdminCreateVendorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请填写完整信息"})
		return
	}

	if req.ProtocolType != "openai-chat" && req.ProtocolType != "openai-responses" && req.ProtocolType != "anthropic-messages" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "协议类型无效"})
		return
	}

	// 用 data_key 解密前端传来的 API Key
	apiKey, err := common.DecryptWithKey(req.APIKey, req.DataKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "API Key 解密失败"})
		return
	}

	// 用服务端密钥加密存储
	encryptedKey, err := common.EncryptSecret(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "API Key 加密失败"})
		return
	}

	vendor := model.Vendor{
		Name:         req.Name,
		Description:  req.Description,
		BaseURL:      req.BaseURL,
		APIKey:       encryptedKey,
		ProtocolType: req.ProtocolType,
		Status:       model.UserStatusEnabled,
	}

	if err := model.CreateVendor(&vendor); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建供应商失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "供应商创建成功"})
}

// AdminUpdateVendorRequest 更新供应商请求
type AdminUpdateVendorRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	BaseURL      string `json:"base_url"`
	APIKey       string `json:"api_key"`
	ProtocolType string `json:"protocol_type"`
	DataKey      string `json:"data_key"`
}

// AdminUpdateVendor 更新供应商
func AdminUpdateVendor(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的供应商ID"})
		return
	}

	var req AdminUpdateVendorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数无效"})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.BaseURL != "" {
		updates["base_url"] = req.BaseURL
	}
	if req.APIKey != "" {
		if req.DataKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "更新 API Key 需要提供 data_key"})
			return
		}
		apiKey, err := common.DecryptWithKey(req.APIKey, req.DataKey)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "API Key 解密失败"})
			return
		}
		encryptedKey, err := common.EncryptSecret(apiKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "API Key 加密失败"})
			return
		}
		updates["api_key"] = encryptedKey
	}
	if req.ProtocolType != "" {
		if req.ProtocolType != "openai-chat" && req.ProtocolType != "openai-responses" && req.ProtocolType != "anthropic-messages" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "协议类型无效"})
			return
		}
		updates["protocol_type"] = req.ProtocolType
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "没有需要更新的字段"})
		return
	}

	if err := model.UpdateVendor(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新供应商失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "供应商更新成功"})
}

// AdminUpdateVendorStatus 更新供应商状态
func AdminUpdateVendorStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的供应商ID"})
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

	if err := model.UpdateVendorStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "状态已更新"})
}

// AdminDeleteVendor 删除供应商
func AdminDeleteVendor(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的供应商ID"})
		return
	}

	if err := model.DeleteVendor(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除供应商失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "供应商已删除"})
}
