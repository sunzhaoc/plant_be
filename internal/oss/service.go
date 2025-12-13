package oss

import (
	"bytes"
	"fmt"
	"io"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/sunzhaoc/plant_be/config"
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

// DownloadImage 通过STS凭证下载OSS图片
func DownloadImage(cfg config.AliConfig, objectKey string) ([]byte, error) {
	stsResp, err := GetSTSCredentials(cfg)
	if err != nil {
		return nil, err
	}

	// 创建OSS客户端
	client, err := oss.New(
		cfg.OSSEndpoint,
		stsResp.Credentials.AccessKeyId,
		stsResp.Credentials.AccessKeySecret,
		oss.SecurityToken(stsResp.Credentials.SecurityToken),
	)
	if err != nil {
		return nil, fmt.Errorf("创建OSS客户端失败: %w", err)
	}

	// 获取Bucket
	bucket, err := client.Bucket(cfg.OSSBucketName)
	if err != nil {
		return nil, fmt.Errorf("获取Bucket失败: %w", err)
	}

	// 读取文件内容到缓冲区
	var buf bytes.Buffer
	objectReader, err := bucket.GetObject(objectKey)
	if err != nil {
		return nil, fmt.Errorf("获取ObjectReader失败: %w", err)
	}
	defer objectReader.Close()

	if _, err = io.Copy(&buf, objectReader); err != nil {
		return nil, fmt.Errorf("复制图片内容失败: %w", err)
	}

	return buf.Bytes(), nil
}
