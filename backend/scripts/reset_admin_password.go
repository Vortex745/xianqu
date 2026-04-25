package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
	Role     string `gorm:"default:'user'"`
}

func (User) TableName() string {
	return "users"
}

func main() {
	dbURL := os.Getenv("DB_URL")

	if dbURL == "" {
		log.Fatal("missing DB_URL")
	}

	fmt.Println("🚀 正在连接 PostgreSQL 数据库...")
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}

	// 生成 bcrypt 密码哈希
	newPassword := "123456"
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 14)
	if err != nil {
		log.Fatalf("❌ 密码加密失败: %v", err)
	}

	// 查找 admin 账号
	var user User
	result := db.Where("username = ?", "admin").First(&user)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// 不存在则创建
			admin := User{
				Username: "admin",
				Password: string(hash),
				Role:     "admin",
			}
			if err := db.Create(&admin).Error; err != nil {
				log.Fatalf("❌ 创建管理员账号失败: %v", err)
			}
			fmt.Println("✅ 管理员账号创建成功!")
			fmt.Printf("   账号: admin\n")
			fmt.Printf("   密码: %s\n", newPassword)
		} else {
			log.Fatalf("❌ 查询管理员账号失败: %v", result.Error)
		}
	} else {
		// 存在则更新密码和角色
		updates := map[string]interface{}{
			"password": string(hash),
			"role":     "admin",
		}
		if err := db.Model(&user).Updates(updates).Error; err != nil {
			log.Fatalf("❌ 更新管理员密码失败: %v", err)
		}
		fmt.Println("✅ 管理员密码重置成功!")
		fmt.Printf("   账号: admin\n")
		fmt.Printf("   新密码: %s\n", newPassword)
	}

	// 验证新密码
	var verifyUser User
	if err := db.Where("username = ?", "admin").First(&verifyUser).Error; err != nil {
		log.Fatalf("❌ 验证失败: 无法查询到管理员账号: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(verifyUser.Password), []byte(newPassword)); err != nil {
		log.Fatalf("❌ 验证失败: 新密码哈希不匹配")
	}

	fmt.Println("✅ 新密码验证通过，可以使用新密码登录系统管理后台")
}
