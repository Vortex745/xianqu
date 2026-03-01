package services

import (
	"errors"
	"gotest/config"
	"gotest/core/models"
	"gotest/core/utils"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
)

type UserService struct{}

// Register 注册
func (s *UserService) Register(username, password, phone string) error {
	if strings.TrimSpace(username) == "" {
		return errors.New("用户名不能为空")
	}
	if utf8.RuneCountInString(password) < 4 {
		return errors.New("密码至少4位")
	}

	var existingUser models.User
	// 检查用户名是否已存在
	if err := config.DB.Where("username = ?", username).First(&existingUser).Error; err == nil {
		return errors.New("用户名已存在")
	}

	// 手机号为选填，仅在填写时检查是否冲突
	if strings.TrimSpace(phone) != "" {
		if err := config.DB.Where("phone = ?", phone).First(&existingUser).Error; err == nil {
			return errors.New("手机号已存在")
		}
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return err
	}
	// 默认昵称
	nickname := "用户" + username
	if len(username) > 4 {
		nickname = "用户" + username[:4]
	}

	user := models.User{
		Username: username,
		Password: hashedPassword,
		Phone:    phone,
		Nickname: nickname,
		Status:   1,
		Avatar:   "https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png", // 默认头像
	}
	return config.DB.Create(&user).Error
}

// Login 登录 (支持 用户名 或 手机号)
func (s *UserService) Login(loginID, password string) (string, *models.User, error) {
	var user models.User

	// 同时查询 username 或 phone
	if err := config.DB.Where("username = ? OR phone = ?", loginID, loginID).First(&user).Error; err != nil {
		return "", nil, errors.New("账号不存在")
	}

	// 检查封禁状态
	if user.Status == 0 {
		return "", nil, errors.New("账号已被封禁，请联系管理员")
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return "", nil, errors.New("密码错误")
	}

	token, err := utils.GenerateToken(user.ID)
	return token, &user, err
}

// UpdateProfile 更新个人资料 (包含手机号)
func (s *UserService) UpdateProfile(user *models.User) error {
	// 使用 Updates 动态更新非零值字段
	return config.DB.Model(user).Updates(map[string]interface{}{
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
		"phone":    user.Phone,
	}).Error
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(userID uint, oldPwd, newPwd string) error {
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	// 1. 验证旧密码
	if !utils.CheckPasswordHash(oldPwd, user.Password) {
		return errors.New("旧密码不正确")
	}

	// 2. 加密新密码
	newHash, err := utils.HashPassword(newPwd)
	if err != nil {
		return errors.New("加密失败")
	}

	// 3. 更新数据库
	return config.DB.Model(&user).Update("password", newHash).Error
}

// GetMyProducts 获取我的发布
func (s *UserService) GetMyProducts(userID uint) ([]models.Product, error) {
	var products []models.Product
	err := config.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&products).Error
	return products, err
}

// GetMyFavorites 获取我的收藏
func (s *UserService) GetMyFavorites(userID uint) ([]models.Favorite, error) {
	var favs []models.Favorite
	// 预加载 Product 信息，以便在收藏列表中显示商品详情
	err := config.DB.Preload("Product").Where("user_id = ?", userID).Order("created_at desc").Find(&favs).Error
	return favs, err
}

// ToggleFavorite 切换收藏状态 (收藏 <-> 取消)
func (s *UserService) ToggleFavorite(userID, productID uint) error {
	var fav models.Favorite
	err := config.DB.Where("user_id = ? AND product_id = ?", userID, productID).First(&fav).Error

	if err == nil {
		// 存在则删除 (取消收藏)
		return config.DB.Delete(&fav).Error
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// 不存在则创建 (添加收藏)
		fav = models.Favorite{UserID: userID, ProductID: productID}
		return config.DB.Create(&fav).Error
	}
	return err
}

// IsFavorited [新增] 检查单个商品是否被收藏
// 这是配合 user_controller.go 中 CheckFavoriteStatus 方法必须的逻辑
func (s *UserService) IsFavorited(userID, productID uint) (bool, error) {
	var fav models.Favorite
	err := config.DB.Where("user_id = ? AND product_id = ?", userID, productID).First(&fav).Error

	if err == nil {
		return true, nil // 找到了记录，说明已收藏
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil // 没找到记录，说明未收藏 (不是错误)
	}

	return false, err // 数据库查询出错
}
