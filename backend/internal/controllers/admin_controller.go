package controllers

import (
	"gotest/config"
	"gotest/internal/middleware"
	"gotest/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AdminController struct{}

// Login 管理员登录
func (a *AdminController) Login(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	var user models.User
	// 1. 查询用户
	if err := config.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "管理员账号不存在"}) // 报 401
		return
	}

	// 2. 检查权限
	if user.Role != "admin" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无权访问后台"}) // 报 401
		return
	}

	// 3. 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"}) // 报 401
		return
	}

	// 4. 生成 Token
	token, _ := middleware.GenerateToken(user.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"token":   token,
		"admin":   user,
	})
}

// Init 初始化管理员账号 (后门接口，用于快速创建管理员)
// 访问地址: http://localhost:8081/api/admin/init
func (a *AdminController) Init(c *gin.Context) {
	var count int64
	config.DB.Model(&models.User{}).Where("username = ?", "admin").Count(&count)
	if count > 0 {
		c.JSON(http.StatusOK, gin.H{"message": "管理员已存在，无需初始化"})
		return
	}

	// 创建默认管理员
	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	admin := models.User{
		Username: "admin",
		Password: string(hash),
		Role:     "admin", // 关键：设置为管理员角色
		Nickname: "系统管理员",
		Avatar:   "https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png",
	}

	if err := config.DB.Create(&admin).Error; err != nil {
		c.JSON(500, gin.H{"error": "初始化失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "管理员创建成功! 账号: admin, 密码: 123456"})
}

// GetInfo 获取管理员信息
func (a *AdminController) GetInfo(c *gin.Context) {
	userID, _ := c.Get("userID") // 从中间件获取
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(404, gin.H{"error": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"name":   user.Nickname,
		"avatar": user.Avatar,
		"roles":  []string{user.Role}, // 前端通常需要数组格式的角色
	})
}

// GetStats 获取统计数据
func (a *AdminController) GetStats(c *gin.Context) {
	var userCount, productCount, orderCount int64
	var tradeAmount float64

	config.DB.Model(&models.User{}).Count(&userCount)
	config.DB.Model(&models.Product{}).Count(&productCount)
	config.DB.Model(&models.Order{}).Count(&orderCount)

	// 统计总交易额 (已完成的订单)
	type Result struct{ Total float64 }
	var res Result
	config.DB.Model(&models.Order{}).Select("sum(price) as total").Where("status > 1").Scan(&res)
	tradeAmount = res.Total

	c.JSON(http.StatusOK, gin.H{
		"user_count":    userCount,
		"product_count": productCount,
		"order_count":   orderCount,
		"trade_amount":  tradeAmount,
	})
}

// GetUsers 获取用户列表
func (a *AdminController) GetUsers(c *gin.Context) {
	var users []models.User
	config.DB.Find(&users)
	c.JSON(http.StatusOK, gin.H{"data": users})
}

// UpdateUserStatus 封禁/解封用户
func (a *AdminController) UpdateUserStatus(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Status int `json:"status"`
	}
	if c.ShouldBindJSON(&input) != nil {
		return
	}

	config.DB.Model(&models.User{}).Where("id = ?", id).Update("status", input.Status)
	c.JSON(http.StatusOK, gin.H{"message": "操作成功"})
}

// GetProducts 获取商品列表
func (a *AdminController) GetProducts(c *gin.Context) {
	var products []models.Product
	config.DB.Preload("User").Order("created_at desc").Find(&products)
	c.JSON(http.StatusOK, gin.H{"data": products})
}

// AuditProduct 审核/下架商品
func (a *AdminController) AuditProduct(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Status int `json:"status"`
	}
	if c.ShouldBindJSON(&input) != nil {
		return
	}

	config.DB.Model(&models.Product{}).Where("id = ?", id).Update("status", input.Status)
	c.JSON(http.StatusOK, gin.H{"message": "操作成功"})
}

// GetOrders 获取订单列表
func (a *AdminController) GetOrders(c *gin.Context) {
	var orders []models.Order
	config.DB.Preload("Product").Preload("User").Order("created_at desc").Find(&orders)
	c.JSON(http.StatusOK, gin.H{"data": orders})
}
