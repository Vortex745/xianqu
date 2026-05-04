package controllers

import (
	"gotest/config"
	"gotest/core/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	notificationTypeShipment        = "shipment"
	notificationTypeProductTakedown = "product_takedown"
	productStatusViolationTakedown  = 3
)

type SystemNotificationController struct{}

func buildShipmentNotification(order models.Order, product models.Product) models.SystemNotification {
	productID := product.ID
	orderID := order.ID
	return models.SystemNotification{
		UserID:    order.SellerID,
		Type:      notificationTypeShipment,
		Title:     "你有新的待发货订单",
		Content:   "买家已支付商品「" + product.Name + "」，请尽快进入我卖出的页面安排发货。",
		ProductID: &productID,
		OrderID:   &orderID,
		IsRead:    false,
	}
}

func buildProductTakedownNotification(product models.Product) models.SystemNotification {
	productID := product.ID
	return models.SystemNotification{
		UserID:    product.UserID,
		Type:      notificationTypeProductTakedown,
		Title:     "商品已被下架",
		Content:   "你发布的商品「" + product.Name + "」已被管理员下架，请检查商品信息后整改。",
		ProductID: &productID,
		IsRead:    false,
	}
}

func shouldNotifyProductTakedown(oldStatus, newStatus int) bool {
	return oldStatus != productStatusViolationTakedown && newStatus == productStatusViolationTakedown
}

func createSystemNotification(tx *gorm.DB, notification models.SystemNotification) error {
	return tx.Create(&notification).Error
}

func (n *SystemNotificationController) List(c *gin.Context) {
	userID, _ := c.Get("userID")

	var notifications []models.SystemNotification
	if err := config.DB.
		Preload("Product").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取系统通知失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": notifications})
}

func (n *SystemNotificationController) UnreadCount(c *gin.Context) {
	userID, _ := c.Get("userID")

	var count int64
	if err := config.DB.Model(&models.SystemNotification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取未读通知失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (n *SystemNotificationController) MarkRead(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "通知不存在"})
		return
	}

	result := config.DB.Model(&models.SystemNotification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "标记通知失败"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "通知不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "操作成功"})
}
