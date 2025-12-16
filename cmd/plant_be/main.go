// 简单的Gin框架示例
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sunzhaoc/plant_be/pkg/aliyun"
)

// ImageResponse 图片URL响应结构
type ImageResponse struct {
	URL string `json:"url"`
}

// ImagesResponse 批量图片响应
type ImagesResponse struct {
	Urls []string `json:"urls"`
}

func main() {
	r := gin.Default()
	cfg := aliyun.LoadAliConfig()
	r.Use(func(c *gin.Context) {
		// 允许前端域名（生产环境替换为你的前端实际域名，如http://localhost:3000）
		c.Header("Access-Control-Allow-Origin", "*")
		// 允许的请求方法（必须包含OPTIONS，fetch会先发OPTIONS预检请求）
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		// 允许的请求头
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
		// 处理OPTIONS预检请求（fetch必触发，不处理会拦截）
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent) // 204
			return
		}
		c.Next()
	})
	// 获取单张图片
	r.GET("/api/plant-image", func(c *gin.Context) {
		imgUrl := c.Query("imgUrl")

		//ossUrl := "plant/squamellaria/squamellaria_grayi/img.png"
		signedURL, err := aliyun.GetOssUrl(cfg, imgUrl, 290, 260)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ImageResponse{URL: ""})
			return
		}
		c.JSON(http.StatusOK, ImageResponse{URL: signedURL})
	})

	// 获取多张图片
	r.GET("/api/plant-images", func(c *gin.Context) {
		plantId := c.Query("plantId")
		// 实际应用中根据plantId查询数据库获取图片列表
		urls := []string{
			"https://your-bucket.oss-cn-beijing.aliyuncs.com/plants/" + plantId + "/1.png",
			"https://your-bucket.oss-cn-beijing.aliyuncs.com/plants/" + plantId + "/2.png",
		}

		c.JSON(http.StatusOK, ImagesResponse{Urls: urls})
	})

	r.Run(":8080")
}
