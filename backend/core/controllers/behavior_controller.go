package controllers

import (
	"net/http"

	"gotest/config"
	"gotest/core/models"
	"gotest/core/services"

	"github.com/gin-gonic/gin"
)

type BehaviorController struct{}

// Report 上报用户行为，接口：POST /api/behavior
func (bc *BehaviorController) Report(c *gin.Context) {
	var input struct {
		ProductID uint   `json:"product_id"`
		Action    string `json:"action"`
		Source    string `json:"source"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if input.ProductID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id 不能为空"})
		return
	}

	action, ok := services.NormalizeBehaviorAction(input.Action)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "action 非法，仅支持 view/like/cart/buy"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var product models.Product
	if err := config.DB.Select("id").First(&product, input.ProductID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "商品不存在"})
		return
	}

	if err := services.RecordUserBehavior(config.DB, userID.(uint), input.ProductID, action, input.Source); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "行为上报失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
