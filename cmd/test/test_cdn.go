package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type CdnAuthConfigTypeA struct {
	Domain  string
	AuthKey string
}

func (c *CdnAuthConfigTypeA) GenerageCdnAuthUrlTypeA(rawUrl string, timestamp int64) (string, error) {
	const (
		uid  = "0"
		rand = "0"
	)

	// 1. 分离 Path 和 Query 参数
	var path, query string
	if idx := strings.Index(rawUrl, "?"); idx != -1 {
		path = rawUrl[:idx]
		query = rawUrl[idx+1:]
	} else {
		path = rawUrl
	}

	// 2. 确保路径以 / 开头
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// 3. 构造签名字符串 (仅使用 Path 部分)
	// 格式: /URI-Timestamp-rand-uid-PrivateKey
	var sb strings.Builder
	sb.WriteString(path)
	sb.WriteByte('-')
	sb.WriteString(fmt.Sprintf("%d", timestamp))
	sb.WriteString("-" + rand + "-" + uid + "-")
	sb.WriteString(c.AuthKey)

	// 4. 计算 MD5
	hash := md5.Sum([]byte(sb.String()))
	md5hash := hex.EncodeToString(hash[:])

	// 5. 拼接 auth_key 字段
	authKeyParam := fmt.Sprintf("%d-%s-%s-%s", timestamp, rand, uid, md5hash)

	// 6. 组合最终 URL
	// 如果原 URL 有参数，用 & 连接 auth_key；如果没有，用 ? 连接
	connector := "?"
	if query != "" {
		connector = "?" + query + "&"
	}

	finalURL := fmt.Sprintf("https://%s%s%sauth_key=%s", c.Domain, path, connector, authKeyParam)

	return finalURL, nil
}

func main() {
	config := CdnAuthConfigTypeA{
		Domain:  "image.antplant.store",
		AuthKey: "sunzhaochuan",
	}

	// 包含图片处理参数的测试
	uri := "/plant/test/test.jpg?image_process=resize,h_200"
	//uri := "/plant/test/test.jpg"

	timestamp := time.Now().Unix() + 3600

	authURL, err := config.GenerageCdnAuthUrlTypeA(uri, timestamp)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("生成的鉴权 URL:")
	fmt.Println(authURL)
}
