package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword 将明文密码加密
func HashPassword(password string) (string, error) {
	// 14 是加密难度，数字越大越安全但也越慢
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash 验证密码是否正确
// password: 用户输入的明文密码
// hash: 数据库中存储的加密密码
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
