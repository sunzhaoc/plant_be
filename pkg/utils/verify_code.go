package utils

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"time"

	"github.com/sunzhaoc/plant_be/pkg/db/redis"
)

// GenerateVerifyCode 生成6位数字验证码
func GenerateVerifyCode() string {
	rand.Seed(time.Now().UnixNano())
	code := rand.Intn(900000) + 100000 // 100000-999999
	return strconv.Itoa(code)
}

// SetVerifyCode 存储验证码到Redis，有效期5分钟
func SetVerifyCode(ctx context.Context, email, code string) error {
	rdb, err := redis.GetDb("ali")
	if err != nil {
		return err
	}
	redisKey := "verify_code:" + email
	return rdb.Set(ctx, redisKey, code, 5*time.Minute).Err() // 设置5分钟有效期（验证码常规有效期）
}

// GetVerifyCode 获取并验证验证码
func GetVerifyCode(ctx context.Context, email string) (string, error) {
	rdb, err := redis.GetDb("ali") // 从Redis连接池获取"ali"配置的连接
	if err != nil {                // 检查连接是否成功
		return "", err // 连接失败，返回空字符串和错误
	}
	redisKey := "verify_code:" + email           // 构建Redis键：verify_code:邮箱地址
	code, err := rdb.Get(ctx, redisKey).Result() // 执行GET命令获取验证码
	if errors.Is(err, redis.Nil) {
		return "", nil // 验证码不存在/过期，返回空字符串+nil错误
	}
	return code, err
}

// DelVerifyCode 验证成功后删除验证码
func DelVerifyCode(ctx context.Context, email string) error {
	rdb, err := redis.GetDb("ali")
	if err != nil {
		return err
	}
	redisKey := "verify_code:" + email
	return rdb.Del(ctx, redisKey).Err()
}
