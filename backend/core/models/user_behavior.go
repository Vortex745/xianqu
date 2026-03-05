package models

import "time"

// UserBehavior 记录用户行为事件，用于推荐与行为分析。
type UserBehavior struct {
	ID uint `gorm:"primaryKey" json:"id"`

	UserID    uint    `gorm:"not null;index:idx_user_behaviors_user_created,priority:1" json:"user_id"`
	ProductID uint    `gorm:"not null;index:idx_user_behaviors_product_action,priority:1" json:"product_id"`
	Action    string  `gorm:"type:varchar(16);not null;index:idx_user_behaviors_action_created,priority:1;index:idx_user_behaviors_product_action,priority:2" json:"action"`
	Weight    float64 `gorm:"type:decimal(6,2);not null" json:"weight"`
	Source    string  `gorm:"type:varchar(64)" json:"source"`

	CreatedAt time.Time `gorm:"index:idx_user_behaviors_created_at;index:idx_user_behaviors_action_created,priority:2;index:idx_user_behaviors_user_created,priority:2" json:"created_at"`
}

func (UserBehavior) TableName() string {
	return "user_behaviors"
}
