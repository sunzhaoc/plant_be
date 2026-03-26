package manage

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/sunzhaoc/plant_be/pkg/db/mysql"
	"github.com/sunzhaoc/plant_be/pkg/db/mysql/models"
)

type PlantSku struct {
	PlantSkuId    uint64          `json:"skuId"`
	PlantSkuSize  string          `json:"size"`
	PlantSkuPrice decimal.Decimal `json:"price"`
	PlantSkuStock uint            `json:"stock"`
	PlantSkuSort  uint8           `json:"sort"`
}

type PlantImage struct {
	PlantImgUrl  string `json:"img_url"`
	PlantImgSort uint8  `json:"sort"`
}

type SavePlantRequest struct {
	PlantId        int          `json:"id"`
	PlantName      string       `json:"name"`
	PlantLatinName string       `json:"latinName"`
	PlantIsOnSale  bool         `json:"isOnSale"`
	PlantDetail    string       `json:"detail"`
	PlantSkus      []PlantSku   `json:"skus"`
	PlantImages    []PlantImage `json:"images"`
}

func ApiPlantManageSavePlant(c *gin.Context) {
	// 绑定并校验参数
	var req SavePlantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("参数绑定失败", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数校验失败",
			"error":   err.Error(),
		})
		return
	}

	fmt.Println("req", req)

	db, err := mysql.GetDB("ali")
	if err != nil {
		slog.Error("数据库连接失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器内部错误"})
		return
	}

	if req.PlantId == -1 {
		// 新增一个植物
		// 开启事务
		tx := db.Begin()
		if tx.Error != nil {
			slog.Error("开启事务失败", "error", tx.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器内部错误", "error": tx.Error.Error()})
			return
		}

		newPlant := models.Plants{
			Name:      strings.TrimSpace(req.PlantName),
			LatinName: strings.TrimSpace(req.PlantLatinName),
			IsOnSale:  req.PlantIsOnSale,
			Detail:    req.PlantDetail,
		}

		if err := tx.Create(&newPlant).Error; err != nil {
			tx.Rollback()
			slog.Error("新增植物失败: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "新增植物失败"})
			return
		}

		// 新增植物的 SKU
		plantId := newPlant.Id // 新增植物的自增ID
		var newPlantSkuList []models.PlantSku
		for _, sku := range req.PlantSkus {
			newPlantSkuList = append(newPlantSkuList, models.PlantSku{
				PlantId: plantId,
				Size:    strings.TrimSpace(sku.PlantSkuSize),
				Price:   sku.PlantSkuPrice,
				Stock:   sku.PlantSkuStock,
				Sort:    sku.PlantSkuSort,
			})
		}
		if err := tx.CreateInBatches(&newPlantSkuList, 10).Error; err != nil { // 每10条一批插入
			tx.Rollback() // 回滚事务
			slog.Error("新增SKU失败", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "新增规格失败"})
			return
		}

		// 新增植物的 图片
		var newPlantImageList []models.PlantImage
		for _, image := range req.PlantImages {
			newPlantImageList = append(newPlantImageList, models.PlantImage{
				PlantId: plantId,
				ImgUrl:  image.PlantImgUrl,
				Sort:    int8(image.PlantImgSort),
			})
		}
		if err := tx.CreateInBatches(&newPlantImageList, 10).Error; err != nil { // 每10条一批插入
			tx.Rollback() // 回滚事务
			slog.Error("新增图片失败", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "新增规格失败"})
			return
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			slog.Error("提交事务失败", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "保存数据失败"})
			return
		}
		slog.Info("新增植物及SKU成功", "plant_id", plantId, "sku_count", len(newPlantSkuList))
	} else {
		// 更新一个植物
		plantId := uint64(req.PlantId)

		// 判断这个植物是否存在
		var exists bool
		query := "SELECT EXISTS(SELECT 1 FROM plant.plants WHERE id = ?);"
		db.Raw(query, plantId).Scan(&exists)
		if !exists {
			c.JSON(http.StatusConflict, gin.H{"success": false, "message": "植物不存在"})
			return
		}

		tx := db.Begin()
		if tx.Error != nil {
			slog.Error("开启事务失败", "error", tx.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器内部错误", "error": tx.Error.Error()})
			return
		}
		// 更新植物主表的信息
		updateSql := `
			UPDATE plant.plants 
			SET name = ?, latin_name = ?, is_on_sale = ?, detail = ?
			WHERE id = ?
		`
		result := tx.Exec(updateSql,
			strings.TrimSpace(req.PlantName), strings.TrimSpace(req.PlantLatinName), req.PlantIsOnSale, req.PlantDetail,
			plantId,
		)
		if result.Error != nil {
			slog.Error("更新植物信息失败")
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新植物信息失败"})
			return
		}

		// 修改 SKU
		// 获取现有数据库中的植物规格SIZE 艺术感
		var existingSkus []models.PlantSku
		if err := tx.Where("plant_id = ?", plantId).Find(&existingSkus).Error; err != nil {
			tx.Rollback()
			slog.Error("查询现有SKU失败", "error", err, "plant_id", plantId)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "获取原有规格失败"})
			return
		}

		// 构建数据库中老的SKU的size映射表
		existingSizeMap := make(map[string]models.PlantSku, len(existingSkus))
		for _, sku := range existingSkus {
			existingSizeMap[strings.TrimSpace(sku.Size)] = sku
		}

		// 处理新SKU列表：相同SIZE更新，不同SIZE新增
		newSizeSet := make(map[string]bool)
		for _, newSku := range req.PlantSkus {
			size := strings.TrimSpace(newSku.PlantSkuSize)
			newSizeSet[size] = true
			if existingSku, ok := existingSizeMap[size]; ok {
				// 如果该size已存在，执行更新
				updateData := map[string]interface{}{
					"price": newSku.PlantSkuPrice,
					"stock": newSku.PlantSkuStock,
					"sort":  newSku.PlantSkuSort,
				}
				if err := tx.Model(&models.PlantSku{}).Where("id = ?", existingSku.Id).Updates(updateData).Error; err != nil {
					tx.Rollback()
					slog.Error("更新SKU失败", "error", err, "size", size, "plant_id", plantId)
					c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新规格失败：" + size})
					return
				}
				delete(existingSizeMap, size)
			} else {
				// 该size不存在，执行新增
				newPlantSku := models.PlantSku{
					PlantId: plantId,
					Size:    size,
					Price:   newSku.PlantSkuPrice,
					Stock:   newSku.PlantSkuStock,
					Sort:    newSku.PlantSkuSort,
				}
				if err := tx.Create(&newPlantSku).Error; err != nil {
					tx.Rollback()
					slog.Error("新增SKU失败", "error", err, "size", size, "plant_id", plantId)
					c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "新增规格失败：" + size})
					return
				}
			}
		}

		// 删除原有SKU中，未出现在新列表中的SKU（即existingSizeMap中剩余的）
		for size, existingSku := range existingSizeMap {
			deleteSql := "DELETE FROM plant.plant_sku WHERE id = ?"
			result := tx.Exec(deleteSql, existingSku.Id)
			if result.Error != nil {
				tx.Rollback()
				slog.Error("删除多余 SKU 失败", "error", result.Error, "size", size, "plant_id", plantId)
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除废弃规格失败：" + size})
				return
			}
			// 检查是否真的删除了记录（防止主键不存在但无报错的情况）
			if result.RowsAffected == 0 {
				tx.Rollback()
				slog.Warn("未找到要删除的SKU记录", "sku_id", existingSku.Id, "size", size)
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除废弃规格失败：未找到规格" + size})
				return
			}
		}

		// 修改 Image
		var existingImages []models.PlantImage
		if err := tx.Where("plant_id = ?", plantId).Find(&existingImages).Error; err != nil {
			tx.Rollback()
			slog.Error("查询现有image失败", "error", err, "plant_id", plantId)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "获取原有图片失败"})
			return
		}

		// 构建数据库中老的Image的url映射表
		existingImageMap := make(map[string]models.PlantImage, len(existingImages))
		for _, __ := range existingImages {
			existingImageMap[strings.TrimSpace(__.ImgUrl)] = __
		}

		// 处理新Image列表：相同img更新，不同img新增
		newImageSet := make(map[string]bool)
		for _, newImage := range req.PlantImages {
			imgUrl := strings.TrimSpace(newImage.PlantImgUrl)
			newImageSet[imgUrl] = true
			if existingImage, ok := existingImageMap[imgUrl]; ok {
				// 如果该image已存在，执行更新
				updateData := map[string]interface{}{
					"img_url": newImage.PlantImgUrl,
					"sort":    newImage.PlantImgSort,
				}
				if err := tx.Model(&models.PlantImage{}).Where("id = ?", existingImage.Id).Updates(updateData).Error; err != nil {
					tx.Rollback()
					slog.Error("更新image失败", "error", err, "url", imgUrl, "plant_id", plantId)
					c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新图片失败：" + imgUrl})
					return
				}
				delete(existingImageMap, imgUrl)
			} else {
				// 该image不存在，执行新增
				newPlantImage := models.PlantImage{
					PlantId: plantId,
					ImgUrl:  newImage.PlantImgUrl,
					Sort:    int8(newImage.PlantImgSort),
				}
				if err := tx.Create(&newPlantImage).Error; err != nil {
					tx.Rollback()
					slog.Error("新增图片失败", "error", err, "size", imgUrl, "plant_id", plantId)
					c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "新增图片失败：" + imgUrl})
					return
				}
			}
		}

		// 删除原有Image中，未出现在新列表中的Image
		for image, existingImage := range existingImageMap {
			deleteSql := "DELETE FROM plant.plant_image WHERE id = ?"
			result := tx.Exec(deleteSql, existingImage.Id)
			if result.Error != nil {
				tx.Rollback()
				slog.Error("删除多余图片失败", "error", result.Error, "size", image, "plant_id", plantId)
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除废弃图片失败：" + image})
				return
			}
			// 检查是否真的删除了记录（防止主键不存在但无报错的情况）
			if result.RowsAffected == 0 {
				tx.Rollback()
				slog.Warn("未找到要删除的图片记录", "sku_id", existingImage.Id, "size", image)
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除废弃图片失败：未找到图片" + image})
				return
			}
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			slog.Error("提交更新事务失败", "error", err, "plant_id", plantId)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "保存更新数据失败"})
			return
		}

		slog.Info("更新植物及SKU成功", "plant_id", plantId, "new_sku_count", len(req.PlantSkus))
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "保存成功"})
}
