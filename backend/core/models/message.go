package models

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// ★★★ 核心修复：添加 json 标签，强制转为小写 ★★★
	SenderID   uint   `json:"sender_id"`
	ReceiverID uint   `json:"receiver_id"`
	Content    string `json:"content"` // 这里的 json:"content" 让前端能读到 msg.content
	Type       int    `json:"type"`    // 1:文本 2:图片
	IsRead     bool   `json:"is_read" gorm:"default:false"`
}

func (Message) TableName() string {
	return "messages"
}
