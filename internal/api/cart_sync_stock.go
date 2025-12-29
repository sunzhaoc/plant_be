package api

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sunzhaoc/plant_be/pkg/db/mysql"
)

type CartItemRequest struct {
	Id       uint64 `json:"id,uint64"`
	SkuId    uint64 `json:"skuId,uint64"`
	Quantity uint64 `json:"quantity,uint"`
}

type CartSyncStockRequest struct {
	CartItems []CartItemRequest `json:"cartItems"`
}

type PlantSku struct {
	PlantId uint64 `gorm:"column:plant_id"`
	Id      uint64 `gorm:"column:id"`
	Stock   uint64 `gorm:"column:stock"`
}

type CartItemResult struct {
	Id          uint64 `json:"id,string"`
	SkuId       uint64 `json:"skuId,string"`
	OldQuantity uint64 `json:"oldQuantity,string"`
	NewQuantity uint64 `json:"newQuantity,string"`
	Stock       uint64 `json:"stock,string"`
}

func SyncCartStock(c *gin.Context) {
	// 1. 立即获取数据库连接，若失败直接返回
	db, err := mysql.GetDB("ali")
	if err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器内部错误"})
		return
	}
	// 2. 绑定并校验参数
	var req CartSyncStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数校验失败",
			"error":   err.Error(),
		})
		return
	}
	if len(req.CartItems) == 0 {
		return
	}

	// 3. 构造查询条件 (Tuple Comparison)
	// 构建形式如: SELECT * FROM table WHERE (plant_id, id) IN ((77, 96), (1, 2))
	var pairs [][]interface{}
	for _, item := range req.CartItems {
		pairs = append(pairs, []interface{}{item.Id, item.SkuId})
	}

	var skus []PlantSku
	err = db.Table("plant.plant_sku").
		Select("plant_id, id, stock").
		Where("(plant_id, id) IN ?", pairs).
		Find(&skus).Error
	if err != nil {
		slog.Error("查询库存失败", "error", err)
		return
	}

	// 4. 构建 Map 提高检索效率 (Key: plantId_skuId)
	stockMap := make(map[string]uint64, len(skus))
	for _, s := range skus {
		key := fmt.Sprintf("%d_%d", s.PlantId, s.Id)
		stockMap[key] = s.Stock
	}

	// 5. 处理业务逻辑：数量修正
	var cartResults []CartItemResult
	for _, item := range req.CartItems {
		key := fmt.Sprintf("%d_%d", item.Id, item.SkuId)
		currentStock := stockMap[key]

		newQty := item.Quantity
		if item.Quantity > currentStock {
			newQty = currentStock
		}

		cartResults = append(cartResults, CartItemResult{
			Id:          item.Id,
			SkuId:       item.SkuId,
			OldQuantity: item.Quantity,
			NewQuantity: newQty,
			Stock:       currentStock,
		})
	}

	// 5. 输出结果
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "修正植物库存成功",
		"data": gin.H{
			"stockInfo": cartResults,
		},
	})
}
