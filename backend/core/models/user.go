package models

import "time"

// User 用户模型，对应数据库中的 users 表
type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`            // 主键ID
	Username string `gorm:"unique;not null" json:"username"` // 用户名，唯一
	Password string `gorm:"not null" json:"-"`               // 密码，json:"-" 表示返回前端时不显示
	Nickname string `json:"nickname"`                        // 昵称
	Avatar   string `json:"avatar"`                          // 头像
	Phone    string `json:"phone"`                           // 手机号（选填）
	Email    string `json:"email"`                           // 邮箱

	// ★★★ 新增 Role 字段 (修复 "field Role unknown" 报错) ★★★
	Role string `gorm:"default:'user'" json:"role"` // 角色: user 或 admin

	Status    int       `gorm:"default:1" json:"status"` // 状态 1:正常 0:禁用
	CreatedAt time.Time `json:"created_at"`              // 创建时间
	UpdatedAt time.Time `json:"updated_at"`              // 更新时间
}

// TableName 指定数据库表名为 users
func (User) TableName() string {
	return "users"
}
