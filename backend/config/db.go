package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gotest/core/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	dbURL := os.Getenv("DB_URL")
	var dialector gorm.Dialector

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	if dbURL != "" {
		fmt.Println("🚀 检测到 DB_URL，正在连接 Postgres...")
		dialector = postgres.Open(dbURL)
	} else {
		fmt.Println("⚠️ 未检测到 DB_URL，数据库初始化将跳过")
		return
	}

	var err error
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		fmt.Printf("❌ 数据库连接失败: %v\n", err)
		return
	}

	sqlDB, _ := DB.DB()
	// Serverless 优化：减少连接数，避免在并发请求时耗尽 pooler 资源
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(2)
	sqlDB.SetConnMaxLifetime(30 * time.Second)

	err = DB.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Order{},
		&models.OrderLog{},
		&models.Cart{},
		&models.Favorite{},
		&models.Message{},
		&models.VerificationCode{},
		&models.AIModel{},
		&models.AIUsageLog{},
		&models.AIUsageDailyStat{},
	)

	if err != nil {
		fmt.Printf("❌ 自动建表失败: %v\n", err)
		return
	}

	if err := ensureProductColumns(); err != nil {
		fmt.Printf("❌ 商品表字段补齐失败: %v\n", err)
	}

	fmt.Println("✅ 数据库初始化流程完成（可能部分失败）")
}

func ensureProductColumns() error {
	if DB == nil {
		return fmt.Errorf("DB is nil")
	}
	product := &models.Product{}
	extraColumns := []struct {
		FieldName  string
		ColumnName string
	}{
		{FieldName: "IsHomeDelivery", ColumnName: "is_home_delivery"},
		{FieldName: "IsSelfPickup", ColumnName: "is_self_pickup"},
	}

	for _, col := range extraColumns {
		if !DB.Migrator().HasColumn(product, col.ColumnName) {
			if err := DB.Migrator().AddColumn(product, col.FieldName); err != nil {
				return fmt.Errorf("add column %s failed: %w", col.ColumnName, err)
			}
			fmt.Printf("✅ 已补齐字段: products.%s\n", col.ColumnName)
		}
	}

	return nil
}
