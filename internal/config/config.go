package config

import (
	"log"
	"os"
)

type AliConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	RoleARN         string
	RoleSessionName string
	OSSEndpoint     string
	OSSBucketName   string
}

func LoadAliConfig() AliConfig {
	// 从环境变量读取敏感配置（避免硬编码）
	accessKeyID := os.Getenv("ALI_OSS_ACCESS_KEY_ID")         // 阿里云访问密钥ID
	accessKeySecret := os.Getenv("ALI_OSS_ACCESS_KEY_SECRET") // 阿里云访问密钥Secret
	roleARN := os.Getenv("ALI_OSS_ROLE_ARN")                  // 阿里云角色ARN

	// 校验必填配置
	if accessKeyID == "" || accessKeySecret == "" || roleARN == "" {
		log.Fatalf("错误：必须配置以下环境变量后运行！\n" +
			"  ALI_OSS_ACCESS_KEY_ID\n" + // 访问密钥ID
			"  ALI_OSS_ACCESS_KEY_SECRET\n" + // 访问密钥Secret
			"  ALI_OSS_ROLE_ARN") // 角色ARN
	}

	return AliConfig{
		AccessKeyID:     accessKeyID,              // 访问密钥ID
		AccessKeySecret: accessKeySecret,          // 访问密钥Secret
		RoleARN:         roleARN,                  // 角色ARN
		RoleSessionName: "aliyun-sts-session-123", // STS临时会话名称
		OSSEndpoint:     "oss-cn-beijing.aliyuncs.com",
		OSSBucketName:   "public-plant-images", // OSS存储桶名称
	}
}
