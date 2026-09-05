package model

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"gorm.io/gorm"
)

// Task 任务记录
type Task struct {
	ID             int        `json:"id" gorm:"primaryKey"`
	TaskID         string     `json:"task_id" gorm:"uniqueIndex;size:64;not null"`
	UserID         int        `json:"user_id" gorm:"index;not null;default:0"`             // 用户ID
	VendorID       int        `json:"vendor_id" gorm:"index;not null;default:0"`
	ModelID        int        `json:"model_id" gorm:"index;not null;default:0"`
	EndpointID     int        `json:"endpoint_id" gorm:"index;not null;default:0"`
	Status         string     `json:"status" gorm:"size:32;not null;default:'submitted'"` // submitted, completed, failed
	QuotaAmount    int64      `json:"quota_amount" gorm:"default:0"`                      // 消耗的积分数量
	QuotaRefunded  bool       `json:"quota_refunded" gorm:"default:false"`                // 积分是否已退还
	VendorResponse string     `json:"vendor_response" gorm:"type:text"`                   // 供应商任务提交响应 JSON
	QueryResponse  string     `json:"query_response" gorm:"type:text"`                    // 供应商任务查询响应 JSON
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

// GenerateTaskID 生成唯一任务ID
func GenerateTaskID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return "task_" + hex.EncodeToString(b)
}

// CreateTask 创建任务
func CreateTask(task *Task) error {
	return DB.Create(task).Error
}

// GetTaskByTaskID 根据任务ID获取任务
func GetTaskByTaskID(taskID string) (*Task, error) {
	var task Task
	err := DB.Where("task_id = ?", taskID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// UpdateTaskStatus 更新任务状态和查询响应
func UpdateTaskStatus(taskID string, status string, queryResponse string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if queryResponse != "" {
		updates["query_response"] = queryResponse
	}
	return DB.Model(&Task{}).Where("task_id = ?", taskID).Updates(updates).Error
}

// UpdateTaskStatusWithRefund 更新任务状态，如果失败则退还积分
func UpdateTaskStatusWithRefund(taskID string, status string, queryResponse string) error {
	// 获取任务信息
	task, err := GetTaskByTaskID(taskID)
	if err != nil {
		return err
	}

	// 更新任务状态
	updates := map[string]interface{}{
		"status": status,
	}
	if queryResponse != "" {
		updates["query_response"] = queryResponse
	}

	// 如果任务失败且积分未退还，则退还积分
	failedStatuses := []string{"failed", "cancelled", "call_fail"}
	isFailed := false
	for _, s := range failedStatuses {
		if status == s {
			isFailed = true
			break
		}
	}

	if isFailed && !task.QuotaRefunded && task.QuotaAmount > 0 {
		// 退还积分
		if err := RefundCredits(task.UserID, taskID, task.QuotaAmount, "任务失败退还"); err != nil {
			return err
		}
		updates["quota_refunded"] = true
	}

	return DB.Model(&Task{}).Where("task_id = ?", taskID).Updates(updates).Error
}

// GetPendingTasks 获取所有未完成的任务（用于启动时恢复轮询）
func GetPendingTasks() ([]Task, error) {
	var tasks []Task
	err := DB.Where("status NOT IN ?", []string{"completed", "failed", "cancelled", "call_fail"}).
		Order("id ASC").Find(&tasks).Error
	return tasks, err
}

// GetTaskLogs 获取任务日志（分页）
func GetTaskLogs(page, pageSize int, status string) ([]Task, int64, error) {
	var tasks []Task
	var total int64

	query := DB.Model(&Task{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// GetTaskByID 根据 ID 获取任务
func GetTaskByID(id int) (*Task, error) {
	var task Task
	err := DB.Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}
