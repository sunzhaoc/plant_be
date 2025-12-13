package aliyun

import (
	"bytes"
	"fmt"
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/sunzhaoc/plant_be/internal/config"
)

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

func GetOssUrl(cfg config.AliConfig, objectKey string) (string, error) {
	stsResp, err := GetSTSCredentials(cfg)
	if err != nil {
		return "", err
	}

	// 创建OSS客户端
	client, err := oss.New(
		cfg.OSSEndpoint,
		stsResp.Credentials.AccessKeyId,
		stsResp.Credentials.AccessKeySecret,
		oss.SecurityToken(stsResp.Credentials.SecurityToken),
	)
	if err != nil {
		return "", fmt.Errorf("创建OSS客户端失败: %w", err)
	}

	// 获取Bucket
	bucket, err := client.Bucket(cfg.OSSBucketName)
	if err != nil {
		return "", fmt.Errorf("获取Bucket失败: %w", err)
	}

	signedURL, err := bucket.SignURL(objectKey, oss.HTTPGet, 1000)
	if err != nil {
		return "", fmt.Errorf("生成签名URL失败: %v", err)
	}
	fmt.Println(signedURL)
	return signedURL, nil
}
