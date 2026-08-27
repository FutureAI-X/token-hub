package common

import "golang.org/x/crypto/bcrypt"

// Password2Hash 将密码转换为 bcrypt 哈希
func Password2Hash(password string) (string, error) {
	passwordBytes := []byte(password)
	hashedPassword, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	return string(hashedPassword), err
}

// ValidatePasswordAndHash 验证密码是否匹配哈希
func ValidatePasswordAndHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
