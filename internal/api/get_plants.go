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

	isNewStr := c.Query("is_new")
	genus := c.Query("genus")

	// 校验参数互斥性：两个参数不能同时存在
	if isNewStr != "" && genus != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "is_new 和 genus 参数不能同时存在"})
		return
	}

	// 校验参数必填性：至少传递其中一个参数
	if isNewStr == "" && genus == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "必须传递 is_new 或 genus 参数"})
		return
	}

	type Plant = struct {
		PlantId    uint64  `json:"plant_id"`     // 改为uint64匹配数据库bigint unsigned类型
		Name       string  `json:"name"`         // 中文名
		LatinName  string  `json:"latin_name"`   // 拉丁学名
		MainImgUrl string  `json:"main_img_url"` // 主图地址
		MinPrice   float64 `json:"min_price"`    // 起始价格
		Stock      int     `json:"stock"`        // 库存
		Tag        string  `json:"tag"`          // 标签
	}
	var plantList []Plant
	var query string
	var args []interface{}

	if genus != "" {
		// genus参数存在
		query = `
			SELECT id plant_id, name, latin_name, main_img_url, min_price, stock, tag 
			FROM plant.plants 
			WHERE is_on_sale = 1 AND genus = ? 
			ORDER BY CASE WHEN stock IS NULL THEN 0 WHEN stock > 0 THEN -1 ELSE 0 END, tag DESC, id
			;
		`
		args = []interface{}{genus}
	} else {
		// is_new参数存在
		if isNewStr != "true" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "is_new 参数值必须为 'true'"})
			return
		}
		query = `
			SELECT id plant_id, name, latin_name, main_img_url, min_price, stock, tag 
			FROM plant.plants 
			WHERE is_on_sale = 1 AND stock > 0 AND tag = '新品'
			ORDER BY id DESC
			LIMIT 50
			;
		`
		args = []interface{}{}
	}

	// 执行查询
	result := db.Raw(query, args...).Scan(&plantList)
	if result.Error != nil {
		slog.Error("查询植物列表失败", slog.Any("error", result.Error))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询植物列表失败",
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "查询植物列表成功",
		"data":    plantList,
		"count":   len(plantList),
	})
}
