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
	VendorID       int        `json:"vendor_id" gorm:"index;not null"`
	ModelID        int        `json:"model_id" gorm:"index;not null"`
	EndpointID     int        `json:"endpoint_id" gorm:"index;not null"`
	Status         string     `json:"status" gorm:"size:32;not null;default:'submitted'"` // submitted, completed, failed
	VendorResponse string     `json:"vendor_response" gorm:"type:text"`                   // 供应商原始响应 JSON
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
