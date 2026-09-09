package common

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT 密钥，从环境变量获取（延迟加载，确保 .env 已先被加载）
var (
	jwtOnce   sync.Once
	jwtSecret []byte
)

// getJWTSecret 获取 JWT 密钥（首次调用时加载 .env）
func getJWTSecret() []byte {
	jwtOnce.Do(func() {
		ensureEnvLoaded()
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "token-hub-jwt-secret-change-me"
		}
		if secret == "token-hub-jwt-secret-change-me" {
			log.Printf("[TOKEN-HUB] [安全] JWT_SECRET 未设置或为默认值，JWT 可被伪造，请设置强随机密钥!")
		}
		jwtSecret = []byte(secret)
	})
	return jwtSecret
}

// Claims JWT 声明
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     int    `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT Token
func GenerateToken(userID int, username string, role int) (string, error) {
	// Token 过期时间：24小时
	expireTime := time.Now().Add(24 * time.Hour)

	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "token-hub",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

// ParseToken 解析 JWT Token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 校验签名算法，仅接受 HMAC，防止算法混淆攻击
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return getJWTSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
