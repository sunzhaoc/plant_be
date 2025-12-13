package oss

import (
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
	"github.com/sunzhaoc/plant_be/internal/config"
)

// GetSTSCredentials 获取STS临时凭证
func GetSTSCredentials(cfg config.AliConfig) (*sts.AssumeRoleResponse, error) {
	client, err := sts.NewClientWithAccessKey("cn-beijing", cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("创建STS客户端失败: %w", err) // 使用%w包装错误，便于上层处理
	}

	request := sts.CreateAssumeRoleRequest()
	request.Scheme = "https"
	request.RoleArn = cfg.RoleARN
	request.RoleSessionName = cfg.RoleSessionName
	request.DurationSeconds = "900"

	response, err := client.AssumeRole(request)
	if err != nil {
		return nil, fmt.Errorf("获取STS临时凭证失败: %w", err)
	}

	return response, nil
}
