package controllers

import (
	"encoding/json"
	"gotest/config"
	"gotest/core/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductController struct {
	// 移除 Service 依赖，直接使用 config.DB
}

// List 获取商品列表
func (p *ProductController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	category := c.Query("category")
	search := c.Query("search")
	isRandom := c.Query("is_random") == "true"

	// ★★★ 修改1: 获取前端传来的区域参数 ★★★
	area := c.Query("area")

	var products []models.Product
	var total int64

	db := config.DB.Model(&models.Product{})

	// 1. 过滤状态：只显示在售商品
	db = db.Where("status = ?", 1)

	// ★★★ 修改2: 增加区域过滤逻辑 ★★★
	if area != "" {
		db = db.Where("area = ?", area)
	}

	// 2. 搜索逻辑
	if search != "" {
		db = db.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 3. 分类筛选
	if category != "" && category != "全部" {
		db = db.Where("category = ?", category)
	}

	// 4. 核心逻辑
	if isRandom {
		// 随机推荐也要 Preload("User")
		// 注意：如果是 MySQL 请用 RAND()，SQLite 用 RANDOM()
		if err := db.Order("RANDOM()").Limit(pageSize).Preload("User").Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取推荐失败"})
			return
		}
		total = int64(len(products))
	} else {
		// 普通分页列表
		db.Count(&total)
		offset := (page - 1) * pageSize
		if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Preload("User").Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取列表失败"})
			return
		}
	}

	// 图片路径处理
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
