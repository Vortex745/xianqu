package models

import "time"

type OrderLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OrderID   uint      `json:"order_id" gorm:"index"`
	Action    string    `json:"action"`     // 操作类型，例如 "TIMEOUT_CANCEL"
	Operator  string    `json:"operator"`   // 操作人，例如 "SYSTEM" 或 "USER"
	Detail    string    `json:"detail"`     // 详情描述
	CreatedAt time.Time `json:"created_at"`
}

func (OrderLog) TableName() string {
	return "order_logs"
}
