package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// 阿里云配置
type AliConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	RoleARN         string
	RoleSessionName string
	OSSEndpoint     string
	OSSBucketName   string
}

// 获取STS临时凭证（修复类型不匹配问题）
func GetSTSCredentials(config AliConfig) (*sts.AssumeRoleResponse, error) {
	client, err := sts.NewClientWithAccessKey("cn-beijing", config.AccessKeyID, config.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("创建STS客户端失败: %v", err)
	}

	request := sts.CreateAssumeRoleRequest()
	request.Scheme = "https"
	request.RoleArn = config.RoleARN
	request.RoleSessionName = config.RoleSessionName
	request.DurationSeconds = "900"

	response, err := client.AssumeRole(request)
	if err != nil {
		return nil, fmt.Errorf("获取STS临时凭证失败: %v", err)
	}

	return response, nil
}

// 通过STS凭证下载OSS图片
func GetOSSImageBySTS(config AliConfig, objectKey string) ([]byte, error) {
	stsResp, err := GetSTSCredentials(config)
	if err != nil {
		return nil, err
	}

	client, err := oss.New(
		config.OSSEndpoint,
		stsResp.Credentials.AccessKeyId,
		stsResp.Credentials.AccessKeySecret,
		oss.SecurityToken(stsResp.Credentials.SecurityToken),
	)
	if err != nil {
		return nil, fmt.Errorf("创建OSS客户端失败: %v", err)
	}

	bucket, err := client.Bucket(config.OSSBucketName)
	if err != nil {
		return nil, fmt.Errorf("获取Bucket失败: %v", err)
	}

	var buf bytes.Buffer
	objectReader, err := bucket.GetObject(objectKey)
	if err != nil {
		return nil, fmt.Errorf("获取ObjectReader失败: %v", err)
	}
	defer objectReader.Close()

	_, err = io.Copy(&buf, objectReader)
	if err != nil {
		return nil, fmt.Errorf("复制图片内容失败: %v", err)
	}

	return buf.Bytes(), nil
}

func main() {
	// 从环境变量读取敏感配置（避免硬编码）
	accessKeyID := os.Getenv("ALI_OSS_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("ALI_OSS_ACCESS_KEY_SECRET")
	roleARN := os.Getenv("ALI_OSS_ROLE_ARN")
	if accessKeyID == "" || accessKeySecret == "" || roleARN == "" {
		log.Fatalf("错误：必须配置以下环境变量后运行！\n" +
			"  ALI_OSS_ACCESS_KEY_ID\n" +
			"  ALI_OSS_ACCESS_KEY_SECRET\n" +
			"  ALI_OSS_ROLE_ARN")
	}

	// 从环境变量读取配置
	config := AliConfig{
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
		RoleARN:         roleARN,
		RoleSessionName: "oss-sts-session-123",
		OSSEndpoint:     "oss-cn-beijing.aliyuncs.com",
		OSSBucketName:   "public-plant-images",
	}

	objectKey := "plant/squamellaria/squamellaria_grayi/img.png" // 图片路径
	imageBytes, err := GetOSSImageBySTS(config, objectKey)       // 下载图片
	if err != nil {
		log.Fatalf("获取图片失败: %v", err)
	}

	err = os.WriteFile("downloaded.jpg", imageBytes, 0644) // 写入本地文件
	if err != nil {
		log.Fatalf("写入文件失败: %v", err)
	}
	log.Printf("成功！图片大小：%d 字节，已保存为downloaded.jpg", len(imageBytes))
}
