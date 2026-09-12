package controller

import (
	"net/http"

	"github.com/FutureAI/token-hub/model"
	"github.com/gin-gonic/gin"
)

// GetPricing 获取模型定价信息
func GetPricing(c *gin.Context) {
	// 获取所有可用模型（带供应商信息）
	models, err := model.GetPricingModels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取模型列表失败",
		})
		return
	}

	// 获取所有供应商（仅公开信息：不含 api_key / base_url）
	vendors, err := model.GetPublicVendors()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取供应商列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"models":  models,
			"vendors": vendors,
		},
	})
}
