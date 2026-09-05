package model

import (
	"time"

	"gorm.io/gorm"
)

// 积分操作类型
const (
	QuotaLogTypeDeduct  = "deduct"  // 扣除
	QuotaLogTypeRefund  = "refund"  // 退还
)

// QuotaLog 积分日志
type QuotaLog struct {
	ID         int        `json:"id" gorm:"primaryKey"`
	UserID     int        `json:"user_id" gorm:"index;not null"`
	TaskID     string     `json:"task_id" gorm:"index;size:64"`
	Amount     int64      `json:"amount" gorm:"not null"`              // 积分数量（正数）
	Type       string     `json:"type" gorm:"size:32;not null"`        // deduct=扣除, refund=退还
	Remark     string     `json:"remark" gorm:"size:255"`              // 备注
	CreatedAt  time.Time  `json:"created_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// DeductQuota 扣除用户积分
func DeductQuota(userID int, taskID string, amount int64, remark string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		// 扣除用户积分
		result := tx.Model(&User{}).Where("id = ? AND quota >= ?", userID, amount).
			Update("quota", gorm.Expr("quota - ?", amount))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrInsufficientQuota
		}

		// 记录积分日志
		log := QuotaLog{
			UserID: userID,
			TaskID: taskID,
			Amount: amount,
			Type:   QuotaLogTypeDeduct,
			Remark: remark,
		}
		return tx.Create(&log).Error
	})
}

// RefundQuota 退还用户积分
func RefundQuota(userID int, taskID string, amount int64, remark string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		// 增加用户积分
		if err := tx.Model(&User{}).Where("id = ?", userID).
			Update("quota", gorm.Expr("quota + ?", amount)).Error; err != nil {
			return err
		}

		// 记录积分日志
		log := QuotaLog{
			UserID: userID,
			TaskID: taskID,
			Amount: amount,
			Type:   QuotaLogTypeRefund,
			Remark: remark,
		}
		return tx.Create(&log).Error
	})
}

// GetQuotaLogsByUserID 获取用户积分日志
func GetQuotaLogsByUserID(userID int, page, pageSize int) ([]QuotaLog, int64, error) {
	var logs []QuotaLog
	var total int64

	query := DB.Model(&QuotaLog{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetQuotaLogByTaskID 根据任务ID获取积分日志
func GetQuotaLogByTaskID(taskID string) (*QuotaLog, error) {
	var log QuotaLog
	err := DB.Where("task_id = ? AND type = ?", taskID, QuotaLogTypeDeduct).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// UpdateQuotaLogTaskID 更新积分日志的任务ID
func UpdateQuotaLogTaskID(userID int, taskID string) error {
	return DB.Model(&QuotaLog{}).
		Where("user_id = ? AND task_id = ? AND type = ?", userID, "", QuotaLogTypeDeduct).
		Update("task_id", taskID).Error
}
