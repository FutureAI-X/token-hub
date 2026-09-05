package controller

import (
	"net/http"
	"strconv"

	"github.com/FutureAI/token-hub/model"
	"github.com/gin-gonic/gin"
)

// AdminGetTaskLogs 管理员获取任务日志列表
func AdminGetTaskLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	tasks, total, err := model.GetTaskLogs(page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "获取任务日志失败"})
		return
	}

	type taskResponse struct {
		ID         int    `json:"id"`
		TaskID     string `json:"task_id"`
		VendorID   int    `json:"vendor_id"`
		ModelID    int    `json:"model_id"`
		EndpointID int    `json:"endpoint_id"`
		Status     string `json:"status"`
		CreatedAt  string `json:"created_at"`
		UpdatedAt  string `json:"updated_at"`
	}

	items := make([]taskResponse, len(tasks))
	for i, t := range tasks {
		items[i] = taskResponse{
			ID:         t.ID,
			TaskID:     t.TaskID,
			VendorID:   t.VendorID,
			ModelID:    t.ModelID,
			EndpointID: t.EndpointID,
			Status:     t.Status,
			CreatedAt:  t.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:  t.UpdatedAt.Format("2006-01-02 15:04:05"),
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

// AdminGetTaskLogDetail 管理员获取任务日志详情
func AdminGetTaskLogDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的任务ID"})
		return
	}

	task, err := model.GetTaskByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "任务不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": task})
}
