package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenExpire      = 7 * 24 * time.Hour // Token 总有效期
	RefreshThreshold = 6 * 24 * time.Hour // 剩余多长时间触发刷新
)

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
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpire)), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),                  // 签发时间
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(GetJWTSecretKey())
}

// ShouldRefresh 检查是否需要刷新 Token
func ShouldRefresh(claims *Claims) bool {
	if claims.ExpiresAt == nil {
		return false
	}
	// 计算剩余有效时间
	remainingTime := time.Until(claims.ExpiresAt.Time)
	// 如果剩余时间小于设定的阈值，则建议刷新
	return remainingTime > 0 && remainingTime < RefreshThreshold
}
