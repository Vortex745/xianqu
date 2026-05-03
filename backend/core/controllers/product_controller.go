package controllers

import (
	"encoding/json"
	"errors"
	"gotest/config"
	"gotest/core/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductController struct {
	// 移除 Service 依赖，直接使用 config.DB
}

func normalizeReportReason(value string) (string, error) {
	reason := strings.TrimSpace(value)
	if len([]rune(reason)) < 2 {
		return "", errors.New("请填写具体举报理由")
	}
	if len([]rune(reason)) > 200 {
		return "", errors.New("举报理由不能超过200字")
	}
	return reason, nil
}

// List 获取商品列表
func (p *ProductController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	category := c.Query("category")
	search := c.Query("search")
	isRandom := c.Query("is_random") == "true"
	area := c.Query("area")
	userID := c.Query("user_id")

	var products []models.Product
	var total int64

	db := config.DB.Model(&models.Product{})
	hasUserFilter := userID != ""

	if hasUserFilter {
		db = db.Where("user_id = ?", userID)
	} else {
		db = db.Where("status = ?", 1)
	}

	if area != "" && !hasUserFilter {
		db = db.Where("area = ?", area)
	}

	if search != "" {
		db = db.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if category != "" && category != "全部" {
		db = db.Where("category = ?", category)
	}

	if isRandom && !hasUserFilter {
		if err := db.Order("RANDOM()").Limit(pageSize).Preload("User").Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取推荐失败"})
			return
		}
		total = int64(len(products))
	} else {
		db.Count(&total)
		offset := (page - 1) * pageSize
		if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Preload("User").Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取列表失败"})
			return
		}
	}

	for i := range products {
		if products[i].Image == "" {
			products[i].Image = "/uploads/default_product.png"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"list":  products,
		"total": total,
		"page":  page,
	})
}

// GetDetail 获取商品详情
func (p *ProductController) GetDetail(c *gin.Context) {
	id := c.Param("id")
	var product models.Product

	// 增加浏览量
	config.DB.Model(&models.Product{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))

	if err := config.DB.Preload("User").First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "商品不存在"})
		return
	}

	// 查询真实收藏数
	var collectCount int64
	config.DB.Model(&models.Favorite{}).Where("product_id = ?", id).Count(&collectCount)

	// 兼容历史服务：确保新交易字段始终出现在详情响应中
	dataMap := gin.H{}
	if raw, err := json.Marshal(product); err == nil {
		_ = json.Unmarshal(raw, &dataMap)
	}
	dataMap["is_home_delivery"] = product.IsHomeDelivery
	dataMap["is_self_pickup"] = product.IsSelfPickup
	dataMap["is_negotiable"] = product.IsNegotiable

	c.JSON(http.StatusOK, gin.H{
		"data":          dataMap,
		"collect_count": collectCount,
	})
}

// Categories 获取分类列表
func (p *ProductController) Categories(c *gin.Context) {
	categories := []string{"数码", "书籍", "生活", "服饰", "运动", "美妆", "乐器", "手游", "兼职", "其他"}
	c.JSON(http.StatusOK, gin.H{"data": categories})
}

// Create 发布商品
func (p *ProductController) Create(c *gin.Context) {
	var input models.Product
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 绑定当前登录用户 ID
	input.UserID = userID.(uint)
	input.Status = 1
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	// 注意：input 结构体中已经包含了 Area 字段（需要在 model 中定义），
	// 这里 ShouldBindJSON 会自动把前端传来的 area 存进去，不需要额外写代码。

	if err := config.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发布失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "发布成功", "data": input})
}

// Update 更新商品
func (p *ProductController) Update(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	var product models.Product
	if err := config.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "商品不存在"})
		return
	}

	// 权限校验
	if product.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权修改此商品"})
		return
	}

	var input models.Product
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateData := map[string]interface{}{
		"name":             input.Name,
		"description":      input.Description,
		"price":            input.Price,
		"count":            input.Count,
		"image":            input.Image,
		"category":         input.Category,
		"area":             input.Area,
		"status":           input.Status,
		"is_free_shipping": input.IsFreeShipping,
		"is_negotiable":    input.IsNegotiable,
		"is_home_delivery": input.IsHomeDelivery,
		"is_self_pickup":   input.IsSelfPickup,
		"updated_at":       time.Now(),
	}

	if err := config.DB.Model(&product).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	// 返回最新数据
	_ = config.DB.First(&product, id).Error

	c.JSON(http.StatusOK, gin.H{"message": "更新成功", "data": product})
}

// Report 举报商品
func (p *ProductController) Report(c *gin.Context) {
	id := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var product models.Product
	if err := config.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "商品不存在"})
		return
	}

	reporterID := userID.(uint)
	if product.UserID == reporterID {
		c.JSON(http.StatusForbidden, gin.H{"error": "不能举报自己的商品"})
		return
	}

	var input struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	reason, err := normalizeReportReason(input.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report := models.ProductReport{
		ProductID: product.ID,
		UserID:    reporterID,
		Reason:    reason,
		CreatedAt: time.Now(),
	}
	if err := config.DB.Create(&report).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "您已举报过该商品"})
		return
	}

	var reportCount int64
	config.DB.Model(&models.ProductReport{}).Where("product_id = ?", product.ID).Count(&reportCount)

	c.JSON(http.StatusOK, gin.H{
		"message":       "举报已提交",
		"report_count":  reportCount,
		"warning_level": productWarningLevel(reportCount),
	})
}
