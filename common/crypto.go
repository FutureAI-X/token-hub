package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"math/big"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// serverKey 服务端加密密钥（从环境变量派生，用于加密存储敏感数据）
var serverKey string

func init() {
	secret := os.Getenv("SECRET_KEY")
	if secret == "" {
		secret = "token-hub-secret-change-me"
	}
	hash := sha256.Sum256([]byte(secret))
	serverKey = base64.StdEncoding.EncodeToString(hash[:])
}

// EncryptSecret 使用服务端密钥加密敏感数据
func EncryptSecret(plaintext string) (string, error) {
	return EncryptWithKey(plaintext, serverKey)
}

// DecryptSecret 使用服务端密钥解密敏感数据
func DecryptSecret(ciphertext string) (string, error) {
	return DecryptWithKey(ciphertext, serverKey)
}

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

// GenerateRandomPassword 生成指定长度的随机密码
func GenerateRandomPassword(length int) string {
	const chars = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789!@#$"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		result[i] = chars[n.Int64()]
	}
	return string(result)
}

// GenerateDataKey 生成 32 字节随机数据密钥，返回 base64 编码字符串
func GenerateDataKey() (string, error) {
	key := make([]byte, 32) // AES-256
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// EncryptWithKey 使用 AES-GCM 加密明文，key 为 base64 编码的 32 字节密钥
func EncryptWithKey(plaintext string, keyBase64 string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptWithKey 使用 AES-GCM 解密密文，key 为 base64 编码的 32 字节密钥
func DecryptWithKey(ciphertextBase64 string, keyBase64 string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return "", err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", err
	}

	nonce, encryptedData := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
