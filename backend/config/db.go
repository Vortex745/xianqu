package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gotest/internal/models" // 确保这里的 module 名和你 go.mod 一致

	"github.com/glebarez/sqlite" // 保持使用纯 Go 驱动，避免 CGO 问题
	"gorm.io/gorm"
	"gorm.io/gorm/logger" // ★★★ 引入日志包
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() {
	// 1. 动态获取当前运行路径
	workDir, _ := os.Getwd()
	dbFileName := "xianqu.db"
	dbPath := filepath.Join(workDir, dbFileName)

	// ★★★ 2. 配置 GORM 日志 (强烈建议) ★★★
	// 这样控制台会打印出 SQL 语句，方便你调试 500 错误
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second, // 慢 SQL 阈值
			LogLevel:                  logger.Info, // 级别：Info (打印所有 SQL)
			IgnoreRecordNotFoundError: true,        // 忽略记录未找到错误
			Colorful:                  true,        // 彩色打印
		},
	)

	// 3. 连接 SQLite
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: newLogger, // 挂载日志配置
	})
	if err != nil {
		log.Fatal("❌ 数据库连接失败: ", err)
	}

	// 4. 开启 SQLite 外键约束 (推荐)
	// SQLite 默认关闭外键，开启后能防止脏数据
	DB.Exec("PRAGMA foreign_keys = ON")

	// 5. 设置连接池
	sqlDB, _ := DB.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(10 * time.Second)

	fmt.Println("✅ SQLite 数据库连接成功！路径:", dbPath)

	// 6. ★★★ 自动迁移表结构 ★★★
	// 注意：请确保 internal/models 下真的有这些结构体，否则会报错
	err = DB.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Order{},    // 启用订单表
		&models.OrderLog{}, // 新增订单日志表
		&models.Cart{},     // 启用购物车表
		&models.Favorite{}, // 启用收藏表
		&models.Message{},
		&models.VerificationCode{}, // ★ 邮箱验证码表
		&models.AIModel{},          // ★ AI模型配置表
		&models.AIUsageLog{},       // ★ AI用量记录表
		&models.AIUsageDailyStat{}, // ★ AI用量日统计表
	)

	if err != nil {
		log.Fatal("❌ 自动建表失败: ", err)
	}

	// 7. 兼容历史数据库：补齐后续新增字段（避免旧库缺列导致功能失效）
	if err := ensureProductColumns(); err != nil {
		log.Fatal("❌ 商品表字段补齐失败: ", err)
	}

	// 8. 兼容历史数据库：手机号改为选填后，移除 phone 的唯一约束
	if err := ensureUserPhoneOptional(); err != nil {
		log.Fatal("❌ 用户表手机号约束迁移失败: ", err)
	}

	fmt.Println("✅ 数据库表结构已同步完成")
}

// ensureProductColumns 兼容旧版 SQLite 库，补齐 products 表缺失列
func ensureProductColumns() error {
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

// ensureUserPhoneOptional 将 users.phone 从唯一约束迁移为普通可选字段
// 兼容历史数据库中由 gorm:"unique" 产生的约束
func ensureUserPhoneOptional() error {
	hasUniquePhone, err := hasUniqueIndexOnUsersPhone()
	if err != nil {
		return err
	}
	if !hasUniquePhone {
		return nil
	}

	if err := DB.Exec("PRAGMA foreign_keys = OFF").Error; err != nil {
		return err
	}
	defer DB.Exec("PRAGMA foreign_keys = ON")

	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	execOrRollback := func(sql string) error {
		if err := tx.Exec(sql).Error; err != nil {
			tx.Rollback()
			return err
		}
		return nil
	}

	createSQL := `
CREATE TABLE users_new (
	id integer PRIMARY KEY AUTOINCREMENT,
	username text NOT NULL,
	password text NOT NULL,
	nickname text,
	avatar text,
	phone text,
	email text,
	role text DEFAULT "user",
	status integer DEFAULT 1,
	created_at datetime,
	updated_at datetime,
	CONSTRAINT uni_users_username UNIQUE (username)
)`

	if err := execOrRollback(createSQL); err != nil {
		return err
	}

	if err := execOrRollback(`
INSERT INTO users_new (
	id, username, password, nickname, avatar, phone, email, role, status, created_at, updated_at
) SELECT
	id, username, password, nickname, avatar, phone, email, role, status, created_at, updated_at
FROM users`); err != nil {
		return err
	}

	if err := execOrRollback("DROP TABLE users"); err != nil {
		return err
	}

	if err := execOrRollback("ALTER TABLE users_new RENAME TO users"); err != nil {
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	fmt.Println("✅ 已迁移 users.phone：移除唯一约束，支持手机号选填")
	return nil
}

func hasUniqueIndexOnUsersPhone() (bool, error) {
	rows, err := DB.Raw("PRAGMA index_list('users')").Rows()
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var seq int
		var name string
		var unique int
		var origin string
		var partial int
		if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
			return false, err
		}
		if unique != 1 {
			continue
		}

		indexInfoSQL := fmt.Sprintf("PRAGMA index_info('%s')", strings.ReplaceAll(name, "'", "''"))
		infoRows, err := DB.Raw(indexInfoSQL).Rows()
		if err != nil {
			return false, err
		}

		hasPhone := false
		for infoRows.Next() {
			var colSeq int
			var colID int
			var colName string
			if err := infoRows.Scan(&colSeq, &colID, &colName); err != nil {
				infoRows.Close()
				return false, err
			}
			if strings.EqualFold(colName, "phone") {
				hasPhone = true
			}
		}
		infoRows.Close()

		if hasPhone {
			return true, nil
		}
	}
	return false, nil
}
