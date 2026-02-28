package controllers

import (
	"fmt"
	"gotest/config"
	"gotest/internal/middleware"
	"gotest/internal/models"
	"gotest/internal/services"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthController struct{}

// 邮箱格式校验正则
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// SendCode 发送验证码
func (a *AuthController) SendCode(c *gin.Context) {
	var input struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据格式错误"})
		return
	}

	email := strings.TrimSpace(strings.ToLower(input.Email))
	if email == "" || !emailRegex.MatchString(email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱格式不正确"})
		return
	}

	// 60秒防刷：查询该邮箱最近一条验证码
	var lastCode models.VerificationCode
	result := config.DB.Where("email = ?", email).Order("created_at desc").First(&lastCode)
	if result.Error == nil {
		elapsed := time.Since(lastCode.CreatedAt)
		if elapsed < 60*time.Second {
			remaining := 60 - int(elapsed.Seconds())
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":     fmt.Sprintf("发送过于频繁，请 %d 秒后再试", remaining),
				"countdown": remaining,
			})
			return
		}
	}

	// 生成 6 位随机数字验证码
	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	// 写入数据库
	newCode := models.VerificationCode{
		Email:     email,
		Code:      code,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	if err := config.DB.Create(&newCode).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统错误，请重试"})
		return
	}

	// 4. 发送邮件 (同步调用以便将第三方服务错误返回给前端)
	err := services.SendVerificationEmail(email, newCode.Code)
	if err != nil {
		fmt.Printf("❌ Failed to send verification email to %s: %v\n", email, err)
		// 删除刚刚生成的验证码（可选，或者保留等它过期都行）
		c.JSON(http.StatusInternalServerError, gin.H{"error": "邮件发送失败，可能是由于尚未绑定域名或者发送限制: " + err.Error()})
		return
	}

	fmt.Printf("✅ Verification email sent to %s successfully\n", email)

	c.JSON(http.StatusOK, gin.H{
		"message": "验证码已发送，请检查您的邮箱",
	})
}

// VerifyLogin 验证码校验并登录/注册
func (a *AuthController) VerifyLogin(c *gin.Context) {
	var input struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据格式错误"})
		return
	}

	email := strings.TrimSpace(strings.ToLower(input.Email))
	inputCode := strings.TrimSpace(input.Code)

	if email == "" || inputCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱和验证码不能为空"})
		return
	}

	// 查询最新的未验证的验证码记录
	var record models.VerificationCode
	err := config.DB.Where("email = ? AND verified_at IS NULL", email).
		Order("created_at desc").First(&record).Error

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "验证码不存在或已被使用，请重新获取"})
		return
	}

	// 检查是否已过期
	if time.Now().After(record.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "验证码已过期，请重新获取"})
		return
	}

	// 检查失败次数是否超过 3 次
	if record.FailedCount >= 3 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "验证码已失效（错误次数过多），请重新获取"})
		return
	}

	// 校验验证码
	if inputCode != record.Code {
		// 递增失败次数
		config.DB.Model(&record).Update("failed_count", record.FailedCount+1)
		remaining := 3 - (record.FailedCount + 1)
		if remaining > 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": fmt.Sprintf("验证码错误，还剩 %d 次机会", remaining),
			})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "验证码已失效（错误次数过多），请重新获取",
			})
		}
		return
	}

	// 验证成功：标记已使用
	now := time.Now()
	config.DB.Model(&record).Update("verified_at", now)

	// 查找或自动注册用户（邮箱登录的核心：无需密码）
	var user models.User
	result := config.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		// 用户不存在，自动注册
		// 从邮箱提取默认用户名
		defaultUsername := strings.Split(email, "@")[0]
		// 确保用户名唯一
		var count int64
		config.DB.Model(&models.User{}).Where("username = ?", defaultUsername).Count(&count)
		if count > 0 {
			defaultUsername = fmt.Sprintf("%s_%d", defaultUsername, time.Now().UnixMilli()%10000)
		}

		user = models.User{
			Username: defaultUsername,
			Password: "", // 邮箱登录用户无需密码
			Email:    email,
			Nickname: defaultUsername,
			Role:     "user",
			Avatar:   "https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png",
		}
		if err := config.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "用户注册失败，请重试"})
			return
		}
	}

	// 生成 JWT Token
	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统错误：Token 生成失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"token":   token,
		"user":    user,
	})
}
