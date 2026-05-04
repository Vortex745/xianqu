package models

import (
	"time"

	"gorm.io/gorm"
)

type SystemNotification struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID    uint   `gorm:"index;not null" json:"user_id"`
	Type      string `gorm:"size:40;not null" json:"type"`
	Title     string `gorm:"size:80;not null" json:"title"`
	Content   string `gorm:"size:500;not null" json:"content"`
	ProductID *uint  `gorm:"index" json:"product_id"`
	OrderID   *uint  `gorm:"index" json:"order_id"`
	IsRead    bool   `gorm:"default:false;index" json:"is_read"`

	Product Product `gorm:"foreignKey:ProductID" json:"product"`
	Order   Order   `gorm:"foreignKey:OrderID" json:"order"`
}

func (SystemNotification) TableName() string {
	return "system_notifications"
}
