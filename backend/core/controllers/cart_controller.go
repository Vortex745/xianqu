package controllers

import (
	"gotest/config"
	"gotest/core/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CartController struct {
	// 移除 Service 依赖，直接用 DB 操作
}

// Add 添加到购物车 (支持数量累加)
func (cc *CartController) Add(c *gin.Context) {
	var input struct {
		ProductID uint `json:"product_id"`
		Count     int  `json:"count"` // 接收前端传来的数量
	}
	// 绑定参数
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 默认数量为 1
	if input.Count <= 0 {
		input.Count = 1
	}

	// 获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	uid := userID.(uint)

	// 1. 检查商品是否存在
	var product models.Product
	if err := config.DB.First(&product, input.ProductID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "商品不存在"})
		return
	}

	// ★★★ 新增：禁止将自己的商品加入购物车 ★★★
	if product.UserID == uid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能将自己的商品加入购物车"})
		return
	}

	// 2. 检查购物车是否已存在该商品
	var cartItem models.Cart
	err := config.DB.Where("user_id = ? AND product_id = ?", uid, input.ProductID).First(&cartItem).Error

	if err == nil {
		// ★★★ 如果已存在，则累加数量 ★★★
		cartItem.Count += input.Count
		config.DB.Save(&cartItem)
		c.JSON(http.StatusOK, gin.H{"message": "购物车数量已更新", "data": cartItem})
		return
	}

	// 3. 不存在，创建新记录
	newCart := models.Cart{
		UserID:    uid,
		ProductID: input.ProductID,
		Count:     input.Count, // 使用传入的数量
	}

	if err := config.DB.Create(&newCart).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "添加成功", "data": newCart})
}

// List 获取购物车列表
func (cc *CartController) List(c *gin.Context) {
	// 获取当前用户
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var list []models.Cart
	// 预加载 Product 与卖家信息，确保前端可拿到 product.seller
	if err := config.DB.
		Preload("Product").
		Preload("Product.User", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("id", "username", "nickname", "avatar")
		}).
		Where("user_id = ?", userID).
		Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取购物车失败"})
		return
	}

	// 处理图片路径
	for i := range list {
		if list[i].Product.Image == "" {
			list[i].Product.Image = "/uploads/default_product.png"
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": list})
}

// Delete 删除购物车项
func (cc *CartController) Delete(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	// 执行删除 (需匹配 UserID，防止删除别人的数据)
	result := config.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Cart{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
