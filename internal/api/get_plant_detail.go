package api

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sunzhaoc/plant_be/pkg/db/mysql"
)

func GetPlantDetail(c *gin.Context) {
	//slog.Info("获取植物详情")
	plantId := c.Param("plantId")

	db, err := mysql.GetDB("ali")
	if err != nil {
		slog.Error("数据库连接失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器内部错误"})
		return
	}

	// 获取植物规格数据
	type PlantSku = struct {
		Size  string  `json:"size"`
		Price float64 `json:"price"`
		Stock uint    `json:"stock"`
	}
	var plantSkuList []PlantSku
	query := "SELECT `size`, price, stock FROM plant.plant_sku WHERE plant_id = ? ORDER BY sort;"
	skuResult := db.Raw(query, plantId).Scan(&plantSkuList)
	if skuResult.Error != nil {
		slog.Error("查询植物SKU列表失败", slog.Any("error", skuResult.Error))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询植物SKU列表失败",
		})
		return
	}

	// 获取植物规格图片
	type PlantImage = struct {
		ImgUrl string `json:"img_url"`
	}
	var plantImageList []PlantImage
	query = "SELECT img_url FROM plant.plant_image WHERE plant_id = ? ORDER BY sort;"
	imageResult := db.Raw(query, plantId).Scan(&plantImageList)
	if imageResult.Error != nil {
		slog.Error("查询植物图片列表失败", slog.Any("error", imageResult.Error))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询植物图片列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取植物详情成功",
		"data": gin.H{
			"skus":   plantSkuList,
			"images": plantImageList,
		},
	})
}
