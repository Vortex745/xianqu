package models

import "time"

type Cart struct {
	ID        uint `gorm:"primaryKey" json:"id"`
	UserID    uint `gorm:"not null" json:"user_id"`
	ProductID uint `gorm:"not null" json:"product_id"`

	// ★★★ 新增：购买数量字段 ★★★
	Count int `gorm:"default:1" json:"count"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联字段：预加载商品信息
	Product Product `gorm:"foreignKey:ProductID" json:"product"`
}

// TableName 指定数据库表名为 carts
func (Cart) TableName() string {
	return "carts"
}
