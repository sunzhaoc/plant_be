package aliyun

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gin-gonic/gin"
)

type OSSUploadPolicy struct {
	AccessKeyId   string `json:"accessKeyId"`
	Policy        string `json:"policy"`
	Signature     string `json:"signature"`
	Dir           string `json:"dir"`           // 上传目录
	Host          string `json:"host"`          // OSS访问域名
	Expire        int64  `json:"expire"`        // Policy过期时间（秒）
	SecurityToken string `json:"securityToken"` // 新增：STS临时凭证的Token
}

func generatePostPolicySignature(policyBase64, accessKeySecret string) (string, error) {
	h := hmac.New(sha1.New, []byte(accessKeySecret))
	_, err := h.Write([]byte(policyBase64))
	if err != nil {
		return "", fmt.Errorf("签名计算失败: %v", err)
	}
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}

func GetOSSUploadPolicy(c *gin.Context) {
	cfg := LoadAliConfig()
	stsResp, err := GetSTSCredentials(cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "获取STS凭证失败", "err": err.Error()})
		return
	}

	expireEndTime := time.Now().Unix() + int64(300) // 5分钟
	conditions := []interface{}{
		[]interface{}{"starts-with", "$key", "plant/"},             // 限制上传目录
		[]interface{}{"content-length-range", 0, 10 * 1024 * 1024}, // 限制文件大小（10MB）
	}

	// 构造Policy
	policyMap := map[string]interface{}{
		"expiration": time.Unix(expireEndTime, 0).UTC().Format("2006-01-02T15:04:05Z"), // 强制UTC时间，避免时区问题
		"conditions": conditions,
	}
	policyJson, err := json.Marshal(policyMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Policy序列化失败", "err": err.Error()})
		return
	}
	policyBase64 := base64.StdEncoding.EncodeToString(policyJson)

	// 4. 生成Signature
	signature, err := generatePostPolicySignature(policyBase64, stsResp.Credentials.AccessKeySecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "生成Policy签名失败", "err": err.Error()})
		return
	}

	// 5. 组装返回数据
	policy := OSSUploadPolicy{
		AccessKeyId:   stsResp.Credentials.AccessKeyId,
		Policy:        policyBase64,
		Signature:     signature,
		Dir:           "plant",
		Host:          fmt.Sprintf("https://%s.%s", cfg.OSSBucketName, cfg.OSSEndpoint),
		Expire:        expireEndTime,
		SecurityToken: stsResp.Credentials.SecurityToken,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": policy,
	})
}

// DownloadImage 通过STS凭证下载OSS图片
func DownloadImage(cfg AliConfig, objectKey string) ([]byte, error) {
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

func GetOssUrl(cfg AliConfig, objectKey string, width int, height int) (string, error) {
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

	var imgProcess string
	if width > 0 && height > 0 {
		// 等比缩放，宽高不超过指定值（推荐）
		imgProcess = fmt.Sprintf("image/resize,lfit,w_%d,h_%d/quality,q_80", width, height)
	} else if width > 0 {
		// 仅指定宽度，高度等比
		imgProcess = fmt.Sprintf("image/resize,lfit,w_%d/quality,q_80", width)
	} else if height > 0 {
		// 仅指定高度，宽度等比
		imgProcess = fmt.Sprintf("image/resize,lfit,h_%d/quality,q_80", height)
	} else {
		// 无缩放，仅压缩质量（可选）
		imgProcess = "image/quality,q_80"
	}

	if imgProcess != "" {
		imgProcess += "/format,webp"
	} else {
		imgProcess = "image/quality,q_8/format,webp"
	}

	signedURL, err := bucket.SignURL(objectKey, oss.HTTPGet, 1000, oss.Process(imgProcess))
	if err != nil {
		return "", fmt.Errorf("生成签名URL失败: %v", err)
	}
	return signedURL, nil
}
