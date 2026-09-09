package model

import (
	"time"

	"gorm.io/gorm"
)

// 积分操作类型
const (
	CreditLogTypeDeduct = "deduct" // 扣除
	CreditLogTypeRefund = "refund" // 退还
)

// CreditLog 积分日志
type CreditLog struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	UserID    int       `json:"user_id" gorm:"index;not null"`
	TaskID    string    `json:"task_id" gorm:"index;size:64"`
	Credits   int64     `json:"credits" gorm:"not null"`      // 积分数量（正数）
	Type      string    `json:"type" gorm:"size:32;not null"` // deduct=扣除, refund=退还
	Remark    string    `json:"remark" gorm:"size:255"`       // 备注
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名（原为 quota_logs）
func (CreditLog) TableName() string {
	return "credit_logs"
}

// DeductCredits 扣除用户积分
func DeductCredits(userID int, taskID string, amount int64, remark string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		return deductCreditsTx(tx, userID, taskID, amount, remark)
	})
}

// deductCreditsTx 在给定事务中扣除用户积分（同事务维护 used_credits）
func deductCreditsTx(tx *gorm.DB, userID int, taskID string, amount int64, remark string) error {
	// 扣除用户积分并累计已用积分
	result := tx.Model(&User{}).Where("id = ? AND credits >= ?", userID, amount).
		Updates(map[string]interface{}{
			"credits":      gorm.Expr("credits - ?", amount),
			"used_credits": gorm.Expr("used_credits + ?", amount),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrInsufficientCredits
	}

	// 记录积分日志
	log := CreditLog{
		UserID:  userID,
		TaskID:  taskID,
		Credits: amount,
		Type:    CreditLogTypeDeduct,
		Remark:  remark,
	}
	return tx.Create(&log).Error
}

// RefundCredits 退还用户积分
func RefundCredits(userID int, taskID string, amount int64, remark string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		return refundCreditsTx(tx, userID, taskID, amount, remark)
	})
}

// refundCreditsTx 在给定事务中退还用户积分（同事务维护 used_credits）
func refundCreditsTx(tx *gorm.DB, userID int, taskID string, amount int64, remark string) error {
	// 增加用户积分并减少已用积分（已用不足时按0计算，避免负值）
	if err := tx.Model(&User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"credits":      gorm.Expr("credits + ?", amount),
			"used_credits": gorm.Expr("GREATEST(used_credits - ?, 0)", amount),
		}).Error; err != nil {
		return err
	}

	// 记录积分日志
	log := CreditLog{
		UserID:  userID,
		TaskID:  taskID,
		Credits: amount,
		Type:    CreditLogTypeRefund,
		Remark:  remark,
	}
	return tx.Create(&log).Error
}

// GetCreditLogsByUserID 获取用户积分日志
func GetCreditLogsByUserID(userID int, page, pageSize int) ([]CreditLog, int64, error) {
	var logs []CreditLog
	var total int64

	query := DB.Model(&CreditLog{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetCreditLogByTaskID 根据任务ID获取积分日志
func GetCreditLogByTaskID(taskID string) (*CreditLog, error) {
	var log CreditLog
	err := DB.Where("task_id = ? AND type = ?", taskID, CreditLogTypeDeduct).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// UpdateCreditLogTaskID 更新积分日志的任务ID
func UpdateCreditLogTaskID(userID int, taskID string) error {
	return DB.Model(&CreditLog{}).
		Where("user_id = ? AND task_id = ? AND type = ?", userID, "", CreditLogTypeDeduct).
		Update("task_id", taskID).Error
}
