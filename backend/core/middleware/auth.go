package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// 定义加密密钥 (需与 Controller 中生成 Token 的密钥一致)
var JwtSecret = []byte("xianqu_super_secret_key_2024")

// JWT Claims 结构体
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 Token (供 User/Admin Controller 调用)
func GenerateToken(userID uint) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24小时有效
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "xianqu",
		},
	}
	// 使用 HS256 签名
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JwtSecret)
}

// ParseToken 解析 Token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return JwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, err
}

// Auth 鉴权中间件
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 尝试从 Header 获取
		tokenString := c.GetHeader("Authorization")

		// 2. 如果 Header 没有，尝试从 URL 参数获取 (适配 WebSocket)
		if tokenString == "" {
			tokenString = c.Query("token")
		}

		// 3. 处理 Bearer 前缀
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录或Token缺失"})
			c.Abort()
			return
		}

		// 智能去除 Bearer 前缀
		if len(tokenString) > 7 && strings.ToUpper(tokenString[0:7]) == "BEARER " {
			tokenString = tokenString[7:]
		}

		// 4. 解析 Token
		claims, err := ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "登录已过期，请重新登录"})
			c.Abort()
			return
		}

		// 5. 将 UserID 存入上下文
		c.Set("userID", claims.UserID)
		c.Next()
	}
}

// AdminAuth 管理员中间件 (目前复用普通验证，由 request.js 分离 Token)
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		Auth()(c)
	}
}
