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
	OSSEndpoint     string // 格式：oss-cn-beijing.aliyuncs.com（无http://）
	OSSBucketName   string
}

// 获取STS临时凭证（修复类型不匹配问题）
func GetSTSCredentials(config AliConfig) (*sts.AssumeRoleResponse, error) {
	// 地域必须匹配：OSS是北京，STS客户端也用cn-beijing
	client, err := sts.NewClientWithAccessKey("cn-beijing", config.AccessKeyID, config.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("创建STS客户端失败: %v", err)
	}

	request := sts.CreateAssumeRoleRequest()
	request.Scheme = "https"
	request.RoleArn = config.RoleARN
	request.RoleSessionName = config.RoleSessionName
	// 核心修复：使用requests.Integer封装900（适配SDK类型要求）
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

	// 创建OSS客户端（Endpoint不要加http://）
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

	// 下载图片到缓冲区
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
	// 从环境变量读取敏感配置（自定义环境变量名，建议大写+下划线）
	accessKeyID := os.Getenv("ALI_OSS_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("ALI_OSS_ACCESS_KEY_SECRET")
	roleARN := os.Getenv("ALI_OSS_ROLE_ARN")

	// 校验环境变量是否配置，未配置则直接退出
	if accessKeyID == "" || accessKeySecret == "" || roleARN == "" {
		log.Fatalf("错误：必须配置以下环境变量后运行！\n" +
			"  ALI_OSS_ACCESS_KEY_ID\n" +
			"  ALI_OSS_ACCESS_KEY_SECRET\n" +
			"  ALI_OSS_ROLE_ARN")
	}

	// 从环境变量读取配置（避免硬编码）
	config := AliConfig{
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
		RoleARN:         roleARN,
		RoleSessionName: "oss-sts-session-123",
		OSSEndpoint:     "oss-cn-beijing.aliyuncs.com", // 正确格式（无http://）
		OSSBucketName:   "public-plant-images",
	}

	// 图片路径（确认Bucket中存在该文件）
	objectKey := "plant/squamellaria/squamellaria_grayi/img.png"

	// 下载图片
	imageBytes, err := GetOSSImageBySTS(config, objectKey)
	if err != nil {
		log.Fatalf("获取图片失败: %v", err)
	}

	// 写入本地文件
	err = os.WriteFile("downloaded.jpg", imageBytes, 0644)
	if err != nil {
		log.Fatalf("写入文件失败: %v", err)
	}

	log.Printf("成功！图片大小：%d 字节，已保存为downloaded.jpg", len(imageBytes))
}
