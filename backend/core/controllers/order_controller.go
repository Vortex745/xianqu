package controllers

import (
	"fmt"
	"gotest/config"
	"gotest/core/models"
	"gotest/core/services"
	"net/http"
	"strconv" // ★★★ 新增：用于订单号拼接
	"time"

	"github.com/gin-gonic/gin"
)

type OrderController struct {
	// 直接使用 config.DB
}

// Create 创建订单 (单商品直接购买)
func (o *OrderController) Create(c *gin.Context) {
	var input struct {
		ProductID uint `json:"product_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	userID, _ := c.Get("userID")
	uid := userID.(uint)

	// 1. 开启事务
	tx := config.DB.Begin()

	// 2. 查询商品
	var product models.Product
	// 加锁查询防止超卖
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&product, input.ProductID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "商品不存在"})
		return
	}

	// 3. 校验状态
	if product.Status != 1 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "商品已售出或下架"})
		return
	}
	if product.UserID == uid {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能购买自己的商品"})
		return
	}

	// 4. 创建订单
	order := models.Order{
		OrderNo:   generateOrderNo(),
		UserID:    uid,
		SellerID:  product.UserID,
		ProductID: product.ID,
		Price:     product.Price,
		Status:    1,                                // 1: 待支付
		ExpiredAt: time.Now().Add(30 * time.Minute), // 新增：30分钟后超时
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建订单失败"})
		return
	}

	// 5. 更新商品状态为已售出 (2)
	if err := tx.Model(&product).Update("status", 2).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新商品状态失败"})
		return
	}

	tx.Commit()
	_ = services.RecordUserBehavior(config.DB, uid, product.ID, "buy", "order_create")
	c.JSON(http.StatusOK, gin.H{"message": "下单成功", "data": order})
}

func (o *OrderController) List(c *gin.Context) {
	userID, _ := c.Get("userID")
	role := c.Query("role")

	var orders []models.Order

	// ★★★ 核心修复：必须 Preload 三层数据 ★★★
	// 1. Product: 获取商品名、图片
	// 2. Seller: 获取卖家头像、昵称
	// 3. User: 获取买家信息
	db := config.DB.Preload("Product").Preload("User").Preload("Seller")

	if role == "seller" {
		db = db.Where("seller_id = ?", userID)
	} else {
		db = db.Where("user_id = ?", userID)
	}

	// 按时间倒序
	if err := db.Order("created_at desc").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取订单失败"})
		return
	}

	// 图片路径处理 (防止前端裂图)
	for i := range orders {
		if orders[i].Product.Image == "" {
			orders[i].Product.Image = "/uploads/default_product.png"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        orders,
		"system_time": time.Now(), // ★★★ 新增：附加系统当前时间
	})
}

// Pay 发起支付申请 (返回 Mock 收银台地址)
func (o *OrderController) Pay(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	var order models.Order
	if err := config.DB.Preload("Product").First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}

	// 校验权限
	if order.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作"})
		return
	}

	if order.Status != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "订单状态异常"})
		return
	}

	// ★★★ 新增：超时拦截 ★★★
	if time.Now().After(order.ExpiredAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "订单已失效，系统正在自动取消中"})
		return
	}

	// 返回收银台跳转链接 (携带必要参数)
	payURL := fmt.Sprintf("/pay/mock?order_id=%d&amount=%.2f", order.ID, order.Price)
	c.JSON(http.StatusOK, gin.H{
		"message": "跳转支付",
		"pay_url": payURL,
	})
}

// ConfirmPay 确认支付 (由 Mock 收银台调用)
func (o *OrderController) ConfirmPay(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	tx := config.DB.Begin()

	var order models.Order
	// 加写锁，防止定时任务正好在这时把订单取消掉
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&order, id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}

	if order.UserID != userID.(uint) {
		tx.Rollback()
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作"})
		return
	}

	if order.Status == 5 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "订单已超时关闭，将自动退款"})
		return
	}

	if order.Status != 1 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "订单已支付或状态异常"})
		return
	}

	// 仅更新状态为待发货 (2)
	// 在实际业务中，这里应该校验第三方回调签名
	if err := tx.Model(&order).Updates(map[string]interface{}{
		"status":     2,
		"updated_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "支付处理失败"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "支付成功"})
}

// ConfirmReceive 确认收货
func (o *OrderController) ConfirmReceive(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	var order models.Order
	if err := config.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}

	if order.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作"})
		return
	}

	// 状态检查：必须是已支付/待发货(2)或运输中(3)
	if order.Status != 2 && order.Status != 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "订单状态不满足收货条件"})
		return
	}

	tx := config.DB.Begin()

	// 1. 更新订单状态为已完成 (4)
	if err := tx.Model(&order).Updates(map[string]interface{}{
		"status":     4,
		"updated_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}

	// 2. 更新商品状态为已完成 (3)
	if err := tx.Model(&models.Product{}).Where("id = ?", order.ProductID).Update("status", 3).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新商品状态失败"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "确认收货成功"})
}

// Refund 申请退单/取消订单
func (o *OrderController) Refund(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	var order models.Order
	if err := config.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}

	if order.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作"})
		return
	}

	// 状态检查：仅 待付款(1)、待发货(2) 支持退单，已发货(3)暂时也允许(简单处理)，已完成(4)不可退
	if order.Status == 4 || order.Status == 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该订单状态无法退单"})
		return
	}

	tx := config.DB.Begin()

	// 1. 更新订单状态为 已取消/已退款 (5)
	if err := tx.Model(&order).Updates(map[string]interface{}{
		"status":     5,
		"updated_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}

	// 2. 恢复商品库存/状态为 上架中 (1)
	if err := tx.Model(&models.Product{}).Where("id = ?", order.ProductID).Update("status", 1).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "恢复商品状态失败"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "退单成功，金额已原路退回"})
}

// BatchCreate 购物车批量结算 (★★★ 核心修复实现 ★★★)
func (o *OrderController) BatchCreate(c *gin.Context) {
	// 1. 定义接收格式
	var input struct {
		CartIDs []uint `json:"cart_ids"` // 前端传来的购物车ID数组
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if len(input.CartIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要结算的商品"})
		return
	}

	userID, _ := c.Get("userID")
	uid := userID.(uint)

	// 2. 开启事务 (Transaction)
	tx := config.DB.Begin()

	// 准备一个数组存放生成的订单，用于返回给前端
	var createdOrders []models.Order

	for _, cartID := range input.CartIDs {
		// A. 查找购物车记录 (确保是自己的)
		var cartItem models.Cart
		if err := tx.Preload("Product").Where("id = ? AND user_id = ?", cartID, uid).First(&cartItem).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "购物车记录不存在或已删除"})
			return
		}

		// B. 检查商品状态
		if cartItem.Product.Status != 1 {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "商品 [" + cartItem.Product.Name + "] 已失效，无法购买"})
			return
		}

		// C. 防自己买自己
		if cartItem.Product.UserID == uid {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "不能购买自己的商品"})
			return
		}

		// D. 创建订单
		order := models.Order{
			OrderNo:   generateOrderNo() + strconv.Itoa(int(cartID)), // 防止高并发下订单号冲突
			UserID:    uid,
			SellerID:  cartItem.Product.UserID,
			ProductID: cartItem.Product.ID,
			Price:     cartItem.Product.Price,
			Status:    1,                                // 待支付
			ExpiredAt: time.Now().Add(30 * time.Minute), // 新增：30分钟后超时
		}

		if err := tx.Create(&order).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建订单失败"})
			return
		}
		createdOrders = append(createdOrders, order)

		// E. 标记商品为已售出 (2)
		if err := tx.Model(&cartItem.Product).Update("status", 2).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "锁定商品失败"})
			return
		}

		// F. 从购物车移除该条目
		if err := tx.Delete(&cartItem).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "移除购物车失败"})
			return
		}
	}

	// 3. 提交事务
	tx.Commit()
	for _, order := range createdOrders {
		_ = services.RecordUserBehavior(config.DB, uid, order.ProductID, "buy", "order_batch")
	}

	c.JSON(http.StatusOK, gin.H{"message": "结算成功", "data": createdOrders})
}

func generateOrderNo() string {
	return time.Now().Format("20060102150405") + "001"
}
