package models

import "time"

// VerificationCode 验证码模型
type VerificationCode struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Email       string     `gorm:"index:idx_email_created,priority:1;not null" json:"email"`
	Code        string     `gorm:"not null" json:"-"`
	ExpiresAt   time.Time  `json:"expires_at"`
	VerifiedAt  *time.Time `json:"verified_at"`
	FailedCount int        `gorm:"default:0" json:"failed_count"`
	CreatedAt   time.Time  `gorm:"index:idx_email_created,priority:2,sort:desc" json:"created_at"`
}

func (VerificationCode) TableName() string {
	return "verification_codes"
}
