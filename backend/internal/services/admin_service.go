package services

import (
	"errors"
	"gotest/config"
	"gotest/internal/models"
	"gotest/internal/utils" // 确保这里引入的是 utils 包
)

type AdminService struct{}

// Login 管理员登录
func (s *AdminService) Login(username, password string) (string, *models.Admin, error) {
	var admin models.Admin
	// 1. 查询管理员
	if err := config.DB.Where("username = ?", username).First(&admin).Error; err != nil {
		return "", nil, errors.New("管理员不存在")
	}

	// 2. 验证密码 (使用 utils/encryption.go 中的方法)
	if !utils.CheckPasswordHash(password, admin.Password) {
		return "", nil, errors.New("密码错误")
	}

	// 3. 生成管理员专属 Token
	token, err := utils.GenerateAdminToken(admin.ID)
	if err != nil {
		return "", nil, errors.New("Token生成失败")
	}

	return token, &admin, nil
}

// CreateAdmin 创建管理员 (用于初始化)
func (s *AdminService) CreateAdmin(username, password string) error {
	// 加密密码
	hash, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	admin := models.Admin{
		Username: username,
		Password: hash,
		// ★★★ 修复点：将 Role 改为 RoleID，并赋值为整数 1 ★★★
		RoleID: 1,
		Avatar: "https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png",
		Status: 1,
	}
	return config.DB.Create(&admin).Error
}

// GetDashboardStats 获取仪表盘统计数据
func (s *AdminService) GetDashboardStats() (map[string]interface{}, error) {
	var userCount, productCount, orderCount int64
	var totalSales float64

	// 1. 统计用户总数
	if err := config.DB.Model(&models.User{}).Count(&userCount).Error; err != nil {
		return nil, err
	}

	// 2. 统计商品总数
	if err := config.DB.Model(&models.Product{}).Count(&productCount).Error; err != nil {
		return nil, err
	}

	// 3. 统计订单总数
	if err := config.DB.Model(&models.Order{}).Count(&orderCount).Error; err != nil {
		return nil, err
	}

	// 4. 统计总交易额 (注意处理 NULL 情况，如果没有订单 sum 为 null)
	// 使用 COALESCE 函数确保返回 0 而不是 null
	if err := config.DB.Model(&models.Order{}).Select("COALESCE(SUM(amount), 0)").Scan(&totalSales).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"user_count":    userCount,
		"product_count": productCount,
		"order_count":   orderCount,
		"total_sales":   totalSales,
	}, nil
}

// GetUserList 获取用户列表 (分页 + 搜索)
func (s *AdminService) GetUserList(page, pageSize int, keyword string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	db := config.DB.Model(&models.User{})

	if keyword != "" {
		db = db.Where("username LIKE ? OR nickname LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 计算总数
	db.Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UpdateUserStatus 修改用户状态 (封号/解封)
func (s *AdminService) UpdateUserStatus(id uint, status int) error {
	// status: 1=正常, 0=封禁 (根据你的 User 模型定义，假设默认是 1)
	return config.DB.Model(&models.User{}).Where("id = ?", id).Update("status", status).Error
}

// GetAdminProductList 管理员获取商品列表 (包含所有状态)
func (s *AdminService) GetAdminProductList(page, pageSize int, keyword string) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	db := config.DB.Model(&models.Product{})

	// 搜索逻辑 (匹配标题或描述)
	if keyword != "" {
		db = db.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 计算总数
	db.Count(&total)

	// 分页查询 (预加载卖家信息)
	offset := (page - 1) * pageSize
	if err := db.Preload("User").Offset(offset).Limit(pageSize).Order("created_at desc").Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// AuditProduct 审核商品 (修改状态)
// status: 1=在售, 2=已售, 3=下架(违规)
func (s *AdminService) AuditProduct(id uint, status int) error {
	return config.DB.Model(&models.Product{}).Where("id = ?", id).Update("status", status).Error
}

// GetAdminOrderList 管理员获取订单列表 (包含买家、商品、卖家信息)
func (s *AdminService) GetAdminOrderList(page, pageSize int, keyword string) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	db := config.DB.Model(&models.Order{})

	// 搜索逻辑 (按订单号搜索)
	if keyword != "" {
		db = db.Where("order_no LIKE ?", "%"+keyword+"%")
	}

	// 计算总数
	db.Count(&total)

	// 分页查询
	// Preload链解释：
	// 1. Preload("User"): 加载买家信息
	// 2. Preload("Product"): 加载商品信息
	// 3. Preload("Product.User"): 加载商品关联的卖家信息
	offset := (page - 1) * pageSize
	if err := db.Preload("User").Preload("Product").Preload("Product.User").
		Offset(offset).Limit(pageSize).Order("created_at desc").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}
