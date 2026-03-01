package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 定义密钥
var jwtKey = []byte("your_secret_key_xianqu_2025")

// Claims 定义 Token 里包含的信息，新增 Role 字段
type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"` // "user" or "admin"
	jwt.RegisteredClaims
}

// GenerateToken 生成普通用户 Token (保持原有签名不变，兼容旧代码)
func GenerateToken(userID uint) (string, error) {
	return generateTokenWithRole(userID, "user")
}

// GenerateAdminToken 生成管理员 Token (新增)
func GenerateAdminToken(adminID uint) (string, error) {
	return generateTokenWithRole(adminID, "admin")
}

// 内部通用生成逻辑
func generateTokenWithRole(id uint, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: id,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    "xianqu-app",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ParseToken 解析 Token
func ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
