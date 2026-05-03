package models

import "time"

type ProductReport struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProductID uint      `gorm:"not null;index:idx_product_report_user,unique" json:"product_id"`
	UserID    uint      `gorm:"not null;index:idx_product_report_user,unique" json:"user_id"`
	Reason    string    `gorm:"type:varchar(200);not null" json:"reason"`
	CreatedAt time.Time `json:"created_at"`

	Product Product `gorm:"foreignKey:ProductID" json:"product"`
	User    User    `gorm:"foreignKey:UserID" json:"user"`
}

func (ProductReport) TableName() string {
	return "product_reports"
}
