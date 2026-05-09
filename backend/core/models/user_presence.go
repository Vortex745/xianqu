package models

import "time"

type UserPresence struct {
	UserID     uint      `gorm:"primaryKey" json:"user_id"`
	Status     string    `gorm:"size:16;not null;default:'offline'" json:"status"`
	LastSeenAt time.Time `gorm:"not null;index" json:"last_seen_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (UserPresence) TableName() string {
	return "user_presences"
}
