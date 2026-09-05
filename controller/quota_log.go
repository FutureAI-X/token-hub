package controller

import (
	"net/http"
	"strconv"

	"github.com/FutureAI/token-hub/model"
	"github.com/gin-gonic/gin"
)

// GetQuotaLogs 获取当前用户的积分日志
func GetQuotaLogs(c *gin.Context) {
	uid, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "未登录"})
		return
	}
	userID := uid.(int)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := model.GetQuotaLogsByUserID(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "获取积分日志失败"})
		return
	}

	type logResponse struct {
		ID        int    `json:"id"`
		TaskID    string `json:"task_id"`
		Amount    int64  `json:"amount"`
		Type      string `json:"type"`
		Remark    string `json:"remark"`
		CreatedAt string `json:"created_at"`
	}

	items := make([]logResponse, len(logs))
	for i, log := range logs {
		items[i] = logResponse{
			ID:        log.ID,
			TaskID:    log.TaskID,
			Amount:    log.Amount,
			Type:      log.Type,
			Remark:    log.Remark,
			CreatedAt: log.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"items":     items,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}
