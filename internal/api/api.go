package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sunzhaoc/plant_be/pkg/aliyun"
)

// ImageResponse 图片URL响应结构
type ImageResponse struct {
	URL string `json:"url"`
}

// GetPlantImageHandler 获取植物图像的签名URL
//
// 从查询参数中获取图片URL，调用阿里云OSS服务生成带有效期的签名URL
//
// 参数:
//
//	c - gin.Context对象，包含HTTP请求上下文和查询参数
//
// 返回值:
//
//	通过JSON返回包含签名URL的响应
//	成功: 200状态码和签名URL
//	失败: 500状态码和空URL
func GetPlantImageHandler(c *gin.Context) {
	cfg := aliyun.LoadAliConfig()
	imgUrl := c.Query("imgUrl")
	signedURL, err := aliyun.GetOssUrl(cfg, imgUrl, 290, 260)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ImageResponse{URL: ""})
		return
	}
	c.JSON(http.StatusOK, ImageResponse{URL: signedURL})
}

// PostRegister 处理用户注册请求
//
// 参数:
//
//	c *gin.Context: Gin 框架的上下文对象
//
// 返回:
//
//	返回 JSON 格式的测试消息
func PostRegister(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "test"})
}
