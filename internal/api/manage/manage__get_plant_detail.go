package manage

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sunzhaoc/plant_be/pkg/db/mysql"
)

func ApiPlantManageGetPlantDetail(c *gin.Context) {
	plantId := c.Param("plantId")

	db, err := mysql.GetDB("ali")
	if err != nil {
		slog.Error("数据库连接失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器内部错误"})
		return
	}

	// 获取植物规格数据
	type PlantSku = struct {
		SkuId uint64  `json:"sku_id"`
		Size  string  `json:"size"`
		Price float64 `json:"price"`
		Stock uint    `json:"stock"`
		Sort  uint8   `json:"sort"`
	}
	var plantSkuList []PlantSku
	query := "SELECT `id` sku_id, `size`, price, stock, sort FROM plant.plant_sku WHERE plant_id = ? ORDER BY sort;"
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
		ImageId uint64 `json:"image_id"`
		ImgUrl  string `json:"img_url"`
		Sort    uint8  `json:"sort"`
	}
	var plantImageList []PlantImage
	query = "SELECT id `image_id`, img_url, sort FROM plant.plant_image WHERE plant_id = ? ORDER BY sort;"
	imageResult := db.Raw(query, plantId).Scan(&plantImageList)
	if imageResult.Error != nil {
		slog.Error("查询植物图片列表失败", slog.Any("error", imageResult.Error))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询植物图片列表失败",
		})
		return
	}

	// 获取植物的介绍
	type PlantMain = struct {
		Name      string `json:"name"`
		LatinName string `json:"latin_name"`
		Detail    string `json:"detail"`
	}
	var plantDetail PlantMain
	query = "SELECT `name`, `latin_name`, `detail` FROM plant.plants WHERE id = ?;"
	detailResult := db.Raw(query, plantId).Scan(&plantDetail)
	if detailResult.Error != nil {
		slog.Error("查询植物介绍失败", slog.Any("error", detailResult.Error))
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
			"plant_name":       plantDetail.Name,
			"plant_latin_name": plantDetail.LatinName,
			"plant_detail":     plantDetail.Detail,
			"skus":             plantSkuList,
			"images":           plantImageList,
		},
	})
}
