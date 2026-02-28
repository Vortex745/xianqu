package services

import (
	"fmt"
	"log"
	"time"

	"gotest/config"
	"gotest/internal/models"
)

// StartOrderSweeper 开启订单超时自动取消轮询线程
func StartOrderSweeper() {
	ticker := time.NewTicker(30 * time.Second)
	// 启动时不阻塞，跑在一个 goroutine 里
	go func() {
		for range ticker.C {
			processTimeoutOrders()
		}
	}()
	fmt.Println(">>> Order Timeout Sweeper Started")
}

func processTimeoutOrders() {
	tx := config.DB.Begin()

	// 查出已超时且状态仍为 "待付款(1)" 的订单
	// 加排他写锁，同时由于有其他节点可能在跑扫表，用 SKIP LOCKED (如果支持，虽然sqlite不支持，但不影响功能) 避免锁等。
	// SQLite 由于是整个库级别的读写锁，不需要写 SKIP LOCKED
	var orders []models.Order
	if err := tx.Where("status = ? AND expired_at <= ?", 1, time.Now()).Find(&orders).Error; err != nil {
		tx.Rollback()
		return
	}

	if len(orders) == 0 {
		tx.Rollback()
		return
	}

	for _, order := range orders {
		// a. 更新订单状态为 "已取消(5)"
		if err := tx.Model(&order).Update("status", 5).Error; err != nil {
			log.Println("❌ 更新订单取消状态失败", err)
			continue
		}

		// b. 自动释放库存 (当前为单点库存，故恢复 status=1 可售状态)
		if err := tx.Model(&models.Product{}).Where("id = ?", order.ProductID).Update("status", 1).Error; err != nil {
			log.Println("❌ 恢复订单库存状态失败", err)
			continue
		}

		// c. 插入系统操作日志
		logEntry := models.OrderLog{
			OrderID:  order.ID,
			Action:   "TIMEOUT_CANCEL",
			Operator: "SYSTEM",
			Detail:   "订单已超过30分钟未支付，系统自动取消并挂起商品出售",
		}
		if err := tx.Create(&logEntry).Error; err != nil {
			log.Println("❌ 写入订单操作日志失败", err)
			continue
		}

		fmt.Printf("✅ 订单自动取消成功: #%s, 商品ID: %d \n", order.OrderNo, order.ProductID)
	}

	tx.Commit()
}
