package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gotest/core/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	dbURL := databaseURL()
	if dbURL == "" {
		log.Fatal("missing PostgreSQL database URL: set DB_URL or DATABASE_URL")
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	var err error
	fmt.Println("🚀 正在连接 PostgreSQL...")
	DB, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("❌ PostgreSQL 数据库连接失败: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("❌ 获取数据库连接池失败: %v", err)
	}

	// Serverless 优化：减少连接数，避免在并发请求时耗尽 pooler 资源
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(2)
	sqlDB.SetConnMaxLifetime(30 * time.Second)

	err = DB.AutoMigrate(
		&models.Admin{},
		&models.User{},
		&models.Product{},
		&models.Order{},
		&models.OrderLog{},
		&models.Cart{},
		&models.Favorite{},
		&models.ProductReport{},
		&models.SystemNotification{},
		&models.UserBehavior{},
		&models.Message{},
		&models.VerificationCode{},
		&models.AIModel{},
		&models.AIUsageLog{},
		&models.AIUsageDailyStat{},
	)

	if err != nil {
		log.Fatalf("❌ 自动建表失败: %v", err)
	}

	if err := ensureProductColumns(); err != nil {
		fmt.Printf("❌ 商品表字段补齐失败: %v\n", err)
	}

	fmt.Println("✅ PostgreSQL 数据库初始化完成")
}

func databaseURL() string {
	for _, key := range []string{"DB_URL", "DATABASE_URL", "POSTGRES_DSN"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
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
