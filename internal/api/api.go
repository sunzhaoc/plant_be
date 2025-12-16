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
