package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	// ★★★ 核心修复：显式定义 ID 并加 json:"id" 标签 ★★★
	// 之前直接用 gorm.Model 会导致返回 "ID" (大写)，前端读不到
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// ★★★ 新增：支付截止时间截 ★★★
	ExpiredAt time.Time `json:"expired_at" gorm:"index"`

	OrderNo   string  `json:"order_no" gorm:"unique"`  // 订单号
	UserID    uint    `json:"user_id"`                 // 买家ID
	SellerID  uint    `json:"seller_id"`               // 卖家ID
	ProductID uint    `json:"product_id"`              // 商品ID
	Price     float64 `json:"price"`                   // 成交价格
	Status    int     `json:"status" gorm:"default:1"` // 1:待支付 2:待发货 ...

	// 关联信息
	Product Product `json:"product"`
	User    User    `json:"user"`                              // 买家信息
	Seller  User    `json:"seller" gorm:"foreignKey:SellerID"` // 卖家信息
}

func (Order) TableName() string {
	return "orders"
}
