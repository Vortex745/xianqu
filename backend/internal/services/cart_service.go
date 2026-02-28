package services

import (
	"errors"
	"gotest/config"
	"gotest/internal/models"
)

type CartService struct{}

// AddToCart 添加到购物车
func (s *CartService) AddToCart(userID, productID uint) error {
	// 1. 检查商品是否存在且在售
	var product models.Product
	if err := config.DB.First(&product, productID).Error; err != nil {
		return errors.New("商品不存在")
	}
	if product.Status != 1 {
		return errors.New("商品已失效")
	}
	if product.UserID == userID {
		return errors.New("不能添加自己的商品")
	}

	// 2. 检查是否已在购物车
	var count int64
	config.DB.Model(&models.Cart{}).Where("user_id = ? AND product_id = ?", userID, productID).Count(&count)
	if count > 0 {
		return errors.New("已经在购物车里啦")
	}

	// 3. 添加
	cart := models.Cart{UserID: userID, ProductID: productID}
	return config.DB.Create(&cart).Error
}

// GetCartList 获取我的购物车
func (s *CartService) GetCartList(userID uint) ([]models.Cart, error) {
	var list []models.Cart
	err := config.DB.Preload("Product").Preload("Product.User").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&list).Error
	return list, err
}

// RemoveFromCart 移除购物车项
func (s *CartService) RemoveFromCart(id uint) error {
	return config.DB.Delete(&models.Cart{}, id).Error
}
