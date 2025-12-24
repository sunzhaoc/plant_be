package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const TokenExpire = 24 * time.Hour // Token 过期时间
//const TokenExpire = 1 * time.Minute // Token 过期时间

// Claims 自定义JWT声明，包含用户非敏感身份信息
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GetJWTSecretKey() []byte {
	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		return []byte("default-secret_key-antplant-store-forever")
	}
	return []byte(secret)
}

// GenerateToken 生成JWT Token
func GenerateToken(userID uint, username string) (string, error) {
	// 1. 构建自定义声明
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpire)), // 过期时间（24小时）
			IssuedAt:  jwt.NewNumericDate(time.Now()),                  // 签发时间
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(GetJWTSecretKey())
}
