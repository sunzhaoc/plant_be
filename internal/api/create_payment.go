package api

import (
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sunzhaoc/plant_be/pkg/db/mysql"
	"github.com/sunzhaoc/plant_be/pkg/db/mysql/models"
)

type Address struct {
	Receiver      string `json:"receiver"`      // 收货人姓名
	Phone         string `json:"phone"`         // 联系电话
	Province      string `json:"province"`      // 省
	City          string `json:"city"`          // 市
	Area          string `json:"area"`          // 区/县
	DetailAddress string `json:"detailAddress"` // 详细地址
}

type PaymentRequest struct {
	CartItems []CartItem `json:"cartItems"`
	Address   Address    `json:"address"` // 收货地址
}

// CartItem 对应前端 cartItems 数组中的单个元素
type CartItem struct {
	PlantId  uint64 `json:"plantId"`
	SkuId    uint64 `json:"skuId"`
	Quantity uint   `json:"quantity"`
}

func generateOrderSn(uid uint64) string {
	now := time.Now()
	timestamp := now.Format("20060102150405") // 年月日时分秒
	randNum := rand.Intn(900000) + 100000     // 6位随机数（100000-999999）
	uidSuffix := uid % 1000000                // 用户ID后6位（防止ID过长）
	return fmt.Sprintf("%s%06d%06d", timestamp, randNum, uidSuffix)
}

func CreatePayment(c *gin.Context) {
	uidRaw, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未登录或ID无效"})
		return
	}
	userId, ok := uidRaw.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户信息解析失败"})
		return
	}
	userId64 := uint64(userId)

	// 2. 绑定参数
	var req PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数校验失败",
			"error":   err.Error(),
		})
		return
	}

	// 3. 获取mysql连接池
	db, err := mysql.GetDB("ali")
	if err != nil {
		slog.Error("数据库连接失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器内部错误"})
		return
	}

	// 4. 开启数据库事务，保证库存操作原子性
	tx := db.Begin()
	if tx.Error != nil {
		slog.Error("开启事务失败", "error", tx.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器内部错误", "error": tx.Error.Error()})
		return
	}

	// 5. 定义事务回滚的延迟处理（确保任何失败都能回滚）
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			slog.Error("付款流程panic", "error", r)
		}
	}()

	// 6. 获取sku的信息
	skuParams := make([]interface{}, 0, len(req.CartItems))
	for _, item := range req.CartItems {
		skuParams = append(skuParams, item.SkuId)
	}
	skuPlaceholders := strings.TrimSuffix(strings.Repeat("?,", len(skuParams)), ",")
	skuSql := fmt.Sprintf(`SELECT plant_id, id sku_id, size sku_size, stock, price FROM plant.plant_sku WHERE id IN (%s) FOR UPDATE`, skuPlaceholders)
	type SkuInfo struct {
		PlantId uint64  `gorm:"column:plant_id"`
		SkuId   uint64  `gorm:"column:sku_id"`
		SkuSize string  `gorm:"column:sku_size"`
		Stock   uint    `gorm:"column:stock"`
		Price   float64 `gorm:"column:price"`
	}
	var skuList []SkuInfo
	if err := tx.Raw(skuSql, skuParams...).Scan(&skuList).Error; err != nil {
		tx.Rollback()
		slog.Error("批量查询SKU失败", "skuIDs", skuParams, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "系统内部错误",
			"error":   err.Error(),
		})
		return
	}
	skuMap := make(map[uint64]SkuInfo, len(skuList))
	for _, sku := range skuList {
		skuMap[sku.SkuId] = sku
	}

	// 7. 库存减扣˚
	updateParams := make([]interface{}, 0, len(req.CartItems)*2) // 批量更新的参数
	skuIdsForWhere := make([]interface{}, 0, len(req.CartItems)) // WHERE IN的SKU ID参数
	var caseWhenBuilder strings.Builder                          // 构建CASE WHEN语句
	for _, item := range req.CartItems {
		// 检查购物车传进来的 sku id 是否存在
		sku, ok := skuMap[item.SkuId]
		if !ok || item.Quantity <= 0 || sku.Stock < item.Quantity {
			tx.Rollback()
			msg := "库存不足或参数错误"
			if !ok {
				msg = fmt.Sprintf("规格ID %d 不存在", item.SkuId)
			} else if item.Quantity <= 0 {
				msg = "购买数量必须大于0"
			}
			slog.Error("校验失败", "skuId", item.SkuId, "reason", msg)
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": msg})
			return
		}
		caseWhenBuilder.WriteString("WHEN id = ? THEN stock - ? ")
		updateParams = append(updateParams, item.SkuId, item.Quantity)
		skuIdsForWhere = append(skuIdsForWhere, item.SkuId)
	}
	wherePlaceholders := strings.TrimSuffix(strings.Repeat("?,", len(skuIdsForWhere)), ",")
	batchUpdateSql := fmt.Sprintf(`
		UPDATE plant.plant_sku 
		SET stock = CASE %s ELSE stock END 
		WHERE id IN (%s)
	`, caseWhenBuilder.String(), wherePlaceholders)
	finalParams := append(updateParams, skuIdsForWhere...)
	if err := tx.Exec(batchUpdateSql, finalParams...).Error; err != nil {
		tx.Rollback()
		slog.Error("批量扣减库存失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "库存扣减失败",
			"error":   err.Error(),
		})
		return
	}

	// ---------------------- 8. 订单主表写入逻辑开始 ----------------------
	var totalAmount float64
	for _, item := range req.CartItems {
		totalAmount += skuMap[item.SkuId].Price * float64(item.Quantity)
	}
	orderSn := generateOrderSn(userId64)

	order := &models.Orders{
		OrderSn:         orderSn,
		UserId:          userId64,
		TotalAmount:     totalAmount,
		PayAmount:       totalAmount,
		OrderStatus:     0,
		ReceiverName:    req.Address.Receiver,
		ReceiverPhone:   req.Address.Phone,
		ReceiverAddress: req.Address.Province + req.Address.City + req.Address.Area + req.Address.DetailAddress,
	}
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		slog.Error("插入订单主表失败", "orderSN", orderSn, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "创建订单失败",
			"error":   err.Error(),
		})
		return
	}

	// ---------------------- 9：订单详情快照开始 ----------------------
	plantIds := make([]interface{}, 0)
	plantIdMap := make(map[uint64]struct{})
	for _, sku := range skuList {
		if _, exists := plantIdMap[sku.PlantId]; !exists {
			plantIdMap[sku.PlantId] = struct{}{}
			plantIds = append(plantIds, sku.PlantId)
		}
	}
	plantPlaceholders := strings.TrimSuffix(strings.Repeat("?,", len(plantIds)), ",")
	plantSql := fmt.Sprintf("SELECT id plant_id, name, latin_name, main_img_url FROM plant.plants WHERE id IN (%s)", plantPlaceholders)
	type PlantInfo struct {
		PlantId    uint64 `gorm:"column:plant_id"`
		Name       string `gorm:"column:name"`
		LatinName  string `gorm:"column:latin_name"`
		MainImgUrl string `gorm:"column:main_img_url"`
	}
	var plantList []PlantInfo
	if err := tx.Raw(plantSql, plantIds...).Scan(&plantList).Error; err != nil {
		tx.Rollback()
		slog.Error("批量查询植物信息失败", "plantIds", plantIds, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取植物信息失败",
			"error":   err.Error(),
		})
		return
	}
	plantMap := make(map[uint64]PlantInfo, len(plantList))
	for _, plant := range plantList {
		plantMap[plant.PlantId] = plant
	}

	var orderItems []models.OrderItem
	for _, item := range req.CartItems {
		sku := skuMap[item.SkuId]
		plant := plantMap[sku.PlantId]

		orderItem := models.OrderItem{
			OrderId:        order.Id,         // 关联订单ID
			PlantId:        sku.PlantId,      // 植物ID
			SkuId:          sku.SkuId,        // 规格ID
			PlantName:      plant.Name,       // 植物名称（快照）
			PlantLatinName: plant.LatinName,  // 拉丁学名（快照）
			SkuSize:        sku.SkuSize,      // 规格名称（快照）
			MainImgUrl:     plant.MainImgUrl, // 主图（快照）
			Price:          sku.Price,        // 下单时单价（快照）
			Quantity:       item.Quantity,    // 购买数量
		}
		orderItems = append(orderItems, orderItem)
	}

	// 7. 批量插入订单商品详情表
	if err := tx.CreateInBatches(&orderItems, 100).Error; err != nil { // 批量插入（每次最多100条）
		tx.Rollback()
		slog.Error("批量插入订单项失败", "orderID", order.OrderSn, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "创建订单项失败",
			"error":   err.Error(),
		})
		return
	}

	// 11. 所有操作成功，提交事务
	if err := tx.Commit().Error; err != nil {
		slog.Error("提交事务失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "系统内部错误",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
