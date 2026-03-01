package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var aesKey = []byte("xianqu_ai_model_enc_key_2024!!!!")

func encrypt(plainText string) string {
	block, _ := aes.NewCipher(aesKey)
	aesGCM, _ := cipher.NewGCM(block)
	nonce := make([]byte, aesGCM.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	cipherText := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(cipherText)
}

type AIModel struct {
	gorm.Model
	Provider    string
	ModelName   string
	APIKey      string
	BaseURL     string
	PricePerK   float64
	Status      int
	Priority    int
	Description string
}

type Product struct {
	ID             uint    `gorm:"primaryKey"`
	Name           string  `gorm:"column:name;not null"`
	Description    string  `gorm:"column:description"`
	Price          float64 `gorm:"column:price;not null"`
	Image          string  `gorm:"column:image"`
	Area           string  `gorm:"column:area"`
	Category       string  `gorm:"column:category"`
	Status         int     `gorm:"column:status;default:1"`
	ViewCount      int     `gorm:"column:view_count;default:0"`
	Count          int     `gorm:"column:count;default:1"`
	IsFreeShipping bool    `gorm:"column:is_free_shipping;default:false"`
	IsNegotiable   bool    `gorm:"column:is_negotiable;default:false"`
	IsHomeDelivery bool    `gorm:"column:is_home_delivery;default:false"`
	IsSelfPickup   bool    `gorm:"column:is_self_pickup;default:false"`
	UserID         uint    `gorm:"column:user_id;not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func main() {
	dsn := "postgresql://neondb_owner:npg_FzBwHbUl3G1M@ep-silent-sky-ajvcov11-pooler.c-3.us-east-2.aws.neon.tech/neondb?sslmode=require"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// 1. Clean and insert AI Models
	db.Exec("TRUNCATE TABLE ai_models CASCADE")
	models := []AIModel{
		{
			Provider:    "deepseek",
			ModelName:   "deepseek-chat",
			APIKey:      encrypt("sk-deepseek-dummy-key-123456789"),
			BaseURL:     "https://api.deepseek.com/v1",
			PricePerK:   0.001,
			Status:      1,
			Priority:    100,
			Description: "主要对话模型",
		},
		{
			Provider:    "deepseek",
			ModelName:   "deepseek-reasoner",
			APIKey:      encrypt("sk-deepseek-dummy-key-123456789"),
			BaseURL:     "https://api.deepseek.com/v1",
			PricePerK:   0.002,
			Status:      1,
			Priority:    50,
			Description: "逻辑推理模型",
		},
		{
			Provider:    "zhipu",
			ModelName:   "glm-4",
			APIKey:      encrypt("sk-zhipu-dummy-key-555555555"),
			BaseURL:     "https://open.bigmodel.cn/api/paas/v4",
			PricePerK:   0.012,
			Status:      1,
			Priority:    30,
			Description: "国产智谱AI模型",
		},
	}
	if err := db.Create(&models).Error; err != nil {
		log.Printf("Error creating models: %v", err)
	} else {
		fmt.Println("AI models populated.")
	}

	// 2. Insert Products
	db.Exec("DELETE FROM products") // Avoid TRUNCATE CASCADE issues if any
	products := []Product{
		{
			Name:           "iPhone 15 Pro Max 256G",
			Description:    "自用闲置，成色极佳，无划痕。电池健康98%。配件齐全，送保护壳。",
			Price:          6500,
			Image:          "https://picsum.photos/seed/iphone/600/600",
			Area:           "上海 浦东新区",
			Category:       "数码",
			Status:         1,
			ViewCount:      1250,
			Count:          1,
			IsNegotiable:   true,
			IsHomeDelivery: false,
			IsSelfPickup:   true,
			UserID:         1,
		},
		{
			Name:           "2023款 MacBook Air M2",
			Description:    "深空灰色，8+256G版本。仅作为会议办公使用，几乎全新。欢迎同城自提。",
			Price:          5200,
			Image:          "https://picsum.photos/seed/macbook/600/600",
			Area:           "北京 朝阳区",
			Category:       "数码",
			Status:         1,
			ViewCount:      890,
			Count:          1,
			IsNegotiable:   true,
			IsHomeDelivery: false,
			IsSelfPickup:   true,
			UserID:         1,
		},
		{
			Name:           "索尼 PS5 光驱版 游戏机",
			Description:    "国行版，购买半年。送底座和两个原装手柄（白色/黑色）。已破解可装双系统。",
			Price:          2800,
			Image:          "https://picsum.photos/seed/ps5/600/600",
			Area:           "广州 天河区",
			Category:       "玩具",
			Status:         1,
			ViewCount:      2300,
			Count:          1,
			IsNegotiable:   false,
			IsHomeDelivery: true,
			IsSelfPickup:   true,
			UserID:         1,
		},
		{
			Name:           "始祖鸟(Arc'teryx) 硬壳冲锋衣",
			Description:    "Alpha SV 系列，买大了。全新仅试穿，吊牌齐全。黑色 M码。保真。",
			Price:          4500,
			Image:          "https://picsum.photos/seed/arcteryx/600/600",
			Area:           "杭州 西湖区",
			Category:       "服饰",
			Status:         1,
			ViewCount:      450,
			Count:          1,
			IsFreeShipping: true,
			IsNegotiable:   false,
			UserID:         1,
		},
		{
			Name:         "大疆 DJI Mini 4 Pro 无人机",
			Description:  "畅飞套装，三枚电池。循环次数极少，功能完全正常，未炸机。含遥控器。",
			Price:        4200,
			Image:        "https://picsum.photos/seed/dji/600/600",
			Area:         "深圳 南山区",
			Category:     "数码",
			Status:       1,
			ViewCount:    670,
			Count:        1,
			IsNegotiable: true,
			UserID:       1,
		},
		{
			Name:           "任天堂 Switch OLED 白色版",
			Description:    "全套原装配件，日版单机。屏幕无黄，按键清脆。附送一个便携收纳包。",
			Price:          1600,
			Image:          "https://picsum.photos/seed/switch/600/600",
			Area:           "成都 武侯区",
			Category:       "玩具",
			Status:         1,
			ViewCount:      1100,
			Count:          1,
			IsHomeDelivery: true,
			IsSelfPickup:   true,
			UserID:         1,
		},
		{
			Name:           "哈利波特 全集 精装套装",
			Description:    "人文版，七册齐全。买来收藏的，品相近乎全新。爱书之人请进。",
			Price:          120,
			Image:          "https://picsum.photos/seed/harrypotter/600/600",
			Area:           "武汉 洪山区",
			Category:       "图书",
			Status:         1,
			ViewCount:      210,
			Count:          1,
			IsFreeShipping: true,
			UserID:         1,
		},
		{
			Name:        "机械工业出版社 - 编译原理 第2版",
			Description: "龙书，经典教材。几乎没动过（太难了）。买书送笔记。",
			Price:       35,
			Image:       "https://picsum.photos/seed/dragonbook/600/600",
			Area:        "西安 雁塔区",
			Category:    "图书",
			Status:      1,
			ViewCount:   150,
			Count:       1,
			UserID:      1,
		},
	}
	if err := db.Create(&products).Error; err != nil {
		log.Printf("Error creating products: %v", err)
	} else {
		fmt.Println("Products populated.")
	}
}
