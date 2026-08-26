package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListModels 返回可用的模型列表
// 固定返回 deepseek-v4-flash
func ListModels(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data": []gin.H{
			{
				"id":       "deepseek-v4-flash",
				"object":   "model",
				"owned_by": "token-hub",
			},
		},
	})
}
