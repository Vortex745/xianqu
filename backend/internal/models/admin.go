package models

import "gorm.io/gorm"

type Admin struct {
	gorm.Model
	// 核心认证信息
	Username string `gorm:"type:varchar(32);unique;not null;comment:用户名" json:"username"`
	Password string `gorm:"type:varchar(128);not null;comment:加密密码" json:"-"` // json:"-" 确保密码不会返回给前端

	// 基本信息
	Nickname string `gorm:"type:varchar(32);comment:昵称" json:"nickname"`
	Avatar   string `gorm:"type:varchar(255);comment:头像" json:"avatar"`
	Phone    string `gorm:"type:varchar(20);comment:手机号" json:"phone"`
	Email    string `gorm:"type:varchar(100);comment:邮箱" json:"email"`

	// 权限与状态
	RoleID int `gorm:"default:1;comment:角色ID(1:超级管理员 2:普通管理员)" json:"role_id"`
	Status int `gorm:"default:1;comment:状态(1:正常 0:禁用)" json:"status"`

	// 登录记录 (安全审计用)
	LastLoginTime *string `gorm:"type:datetime;comment:最后登录时间" json:"last_login_time"`
	LastLoginIP   string  `gorm:"type:varchar(50);comment:最后登录IP" json:"last_login_ip"`
}

func (Admin) TableName() string {
	return "admins"
}
