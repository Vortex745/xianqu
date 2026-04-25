package models

import "time"

type Product struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// ★★★ 修改1: 显式添加 column:name，强制数据库列名为 name ★★★
	Name string `gorm:"column:name;not null" json:"name"`

	Description string  `gorm:"column:description" json:"description"`
	Price       float64 `gorm:"column:price;not null" json:"price"`
	Image       string  `gorm:"column:image" json:"image"`
	Area        string  `gorm:"column:area" json:"area"`

	// ★★★ 修改2: 显式指定 column:category ★★★
	Category string `gorm:"column:category" json:"category"`

	Status int `gorm:"column:status;default:1" json:"status"`

	// ★★★ 修改3: 显式指定 column:view_count ★★★
	ViewCount int `gorm:"column:view_count;default:0" json:"view_count"`

	// 其他字段保持原样
	Count          int  `json:"count" gorm:"default:1"`
	IsFreeShipping bool `json:"is_free_shipping" gorm:"default:false"`
	IsNegotiable   bool `json:"is_negotiable" gorm:"default:false"`
	IsHomeDelivery bool `json:"is_home_delivery" gorm:"default:false"`
	IsSelfPickup   bool `json:"is_self_pickup" gorm:"default:false"`

	UserID    uint      `gorm:"column:user_id;not null" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联字段
	User User `gorm:"foreignKey:UserID" json:"seller"`
}

func (Product) TableName() string {
	return "products"
}
