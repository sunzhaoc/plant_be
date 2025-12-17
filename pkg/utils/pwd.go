package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword 使用bcrypt算法对密码进行哈希处理
//
// 参数:
//
//	password: 需要哈希的原始密码字符串
//
// 返回值:
//
//	string: 哈希后的密码字符串
//	error: 如果哈希过程中出现错误则返回错误信息
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash 比较明文密码与bcrypt哈希是否匹配
//
// 参数:
//
//	password: 待验证的明文密码
//	hash:     存储的bcrypt密码哈希
//
// 返回值:
//
//	bool: 如果密码匹配返回true，否则返回false
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
