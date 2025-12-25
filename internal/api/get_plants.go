package api

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sunzhaoc/plant_be/pkg/db/mysql"
)

func GetPlants(c *gin.Context) {
	db, err := mysql.GetDB("ali")
	if err != nil {
		slog.Info("数据库连接失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器内部错误"})
		return
	}

	type Plant = struct {
		PlantId    uint64  `json:"plant_id"`     // 改为uint64匹配数据库bigint unsigned类型
		Name       string  `json:"name"`         // 中文名
		LatinName  string  `json:"latin_name"`   // 拉丁学名
		MainImgUrl string  `json:"main_img_url"` // 主图地址
		MinPrice   float64 `json:"min_price"`    // 起始价格
	}
	var plantList []Plant

	query := "SELECT id plant_id, name, latin_name, main_img_url, min_price FROM plant.plants WHERE is_on_sale = 1;"
	result := db.Raw(query).Scan(&plantList)

	if result.Error != nil {
		slog.Error("查询植物列表失败", slog.Any("error", result.Error))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询植物列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "查询上架植物列表成功",
		"data":    plantList,
		"count":   len(plantList),
	})
}
