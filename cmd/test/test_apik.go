//package main
//
//import (
//	"encoding/json"
//	"fmt"
//	"log/slog"
//
//	"github.com/sunzhaoc/plant_be/pkg/db/mysql"
//)
//
//// --- 模型定义 ---
//
//type CartItemRequest struct {
//	Id       uint64 `json:"id,string"` // 使用 ,string 标签自动处理前端传来的字符串
//	SkuId    uint64 `json:"skuId,string"`
//	Quantity uint64 `json:"quantity,string"`
//}
//
//type CartSyncStockRequest struct {
//	CartItems []CartItemRequest `json:"cartItems"`
//}
//
//type PlantSku struct {
//	PlantId uint64 `gorm:"column:plant_id"`
//	Id      uint64 `gorm:"column:id"`
//	Stock   uint64 `gorm:"column:stock"`
//}
//
//type CartItemResult struct {
//	Id          uint64 `json:"id,string"`
//	SkuId       uint64 `json:"skuId,string"`
//	OldQuantity uint64 `json:"oldQuantity,string"`
//	NewQuantity uint64 `json:"newQuantity,string"`
//	Stock       uint64 `json:"stock,string"`
//}
//
//func main() {
//	// 1. 初始化 (建议放在 init 或单独的 setup 函数中)
//	if err := mysql.Init(mysql.Load(), []string{"ali"}); err != nil {
//		slog.Error("数据库初始化失败", "err", err)
//		return
//	}
//	db, _ := mysql.GetDB("ali")
//
//	// 模拟输入
//	req := CartSyncStockRequest{
//		CartItems: []CartItemRequest{
//			{Id: 77, SkuId: 96, Quantity: 10},
//			{Id: 1, SkuId: 2, Quantity: 3},
//		},
//	}
//
//	if len(req.CartItems) == 0 {
//		return
//	}
//
//	// 2. 构造查询条件 (Tuple Comparison)
//	// 构建形式如: SELECT * FROM table WHERE (plant_id, id) IN ((77, 96), (1, 2))
//	var pairs [][]interface{}
//	for _, item := range req.CartItems {
//		pairs = append(pairs, []interface{}{item.Id, item.SkuId})
//	}
//
//	var skus []PlantSku
//	err := db.Table("plant.plant_sku").
//		Select("plant_id, id, stock").
//		Where("(plant_id, id) IN ?", pairs).
//		Find(&skus).Error
//
//	if err != nil {
//		slog.Error("查询库存失败", "error", err)
//		return
//	}
//
//	// 3. 构建 Map 提高检索效率 (Key: plantId_skuId)
//	stockMap := make(map[string]uint64, len(skus))
//	for _, s := range skus {
//		key := fmt.Sprintf("%d_%d", s.PlantId, s.Id)
//		stockMap[key] = s.Stock
//	}
//
//	// 4. 处理业务逻辑：数量修正
//	var cartResults []CartItemResult
//	for _, item := range req.CartItems {
//		key := fmt.Sprintf("%d_%d", item.Id, item.SkuId)
//		currentStock := stockMap[key] // 若 key 不存在，默认就是 0
//
//		newQty := item.Quantity
//		if item.Quantity > currentStock {
//			newQty = currentStock
//		}
//
//		cartResults = append(cartResults, CartItemResult{
//			Id:          item.Id,
//			SkuId:       item.SkuId,
//			OldQuantity: item.Quantity,
//			NewQuantity: newQty,
//			Stock:       currentStock,
//		})
//	}
//
//	// 5. 输出结果
//	output, _ := json.MarshalIndent(cartResults, "", "  ")
//	fmt.Println(string(output))
//}

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sunzhaoc/plant_be/pkg/db/mysql"
)

// 定义常量，避免硬编码，方便维护
const (
	DBInstanceAli      = "ali"             // 阿里云数据库实例名
	TablePlantSku      = "plant.plant_sku" // 商品SKU表名
	ErrMsgInternal     = "服务器内部错误"         // 通用内部错误提示
	ErrMsgInvalidParam = "参数格式错误"          // 参数校验错误提示
)

// CartItemRequest 购物车项请求参数
// 注：若Id对应plant_id、SkuId对应sku表的id，建议字段名改为PlantId/SkuId以提高可读性
type CartItemRequest struct {
	Id       uint64 `json:"id,string" binding:"required,min=1"`       // 关联plant_id，必填且≥1
	SkuId    uint64 `json:"skuId,string" binding:"required,min=1"`    // SKU ID，必填且≥1
	Quantity uint64 `json:"quantity,string" binding:"required,min=0"` // 购物车数量，必填且≥0
}

// CartSyncStockRequest 购物车库存同步请求体
type CartSyncStockRequest struct {
	CartItems []CartItemRequest `json:"cartItems" binding:"required,dive"` // 购物车项列表，必填且内部元素需校验
}

// PlantSku 商品SKU库存模型（与数据库映射）
type PlantSku struct {
	PlantId uint64 `gorm:"column:plant_id"` // 商品ID
	Id      uint64 `gorm:"column:id"`       // SKU ID
	Stock   uint64 `gorm:"column:stock"`    // 库存数量
}

// CartItemResult 购物车库存同步结果
type CartItemResult struct {
	Id          uint64 `json:"id,string"`          // 商品ID（plant_id）
	SkuId       uint64 `json:"skuId,string"`       // SKU ID
	OldQuantity uint64 `json:"oldQuantity,string"` // 原购物车数量
	NewQuantity uint64 `json:"newQuantity,string"` // 修正后数量（不超过库存）
	Stock       uint64 `json:"stock,string"`       // 当前SKU库存
}

// SyncCartStockResponse 统一响应结构体
type SyncCartStockResponse struct {
	Success bool             `json:"success"` // 是否成功
	Message string           `json:"message"` // 提示信息
	Data    []CartItemResult `json:"data"`    // 同步结果
}

// SyncCartStock 同步购物车库存（修正超出库存的购物车数量）
// @Summary 同步购物车库存
// @Description 根据商品SKU库存，修正购物车中超出库存的商品数量
// @Tags 购物车
// @Accept json
// @Produce json
// @Param request body CartSyncStockRequest true "购物车库存同步请求"
// @Success 200 {object} SyncCartStockResponse "同步成功"
// @Failure 400 {object} SyncCartStockResponse "参数错误"
// @Failure 500 {object} SyncCartStockResponse "服务器内部错误"
// @Router /cart/sync-stock [post]
func SyncCartStock(c *gin.Context) {
	// 1. 统一响应函数，简化重复代码
	resp := func(code int, success bool, msg string, data []CartItemResult) {
		c.JSON(code, SyncCartStockResponse{
			Success: success,
			Message: msg,
			Data:    data,
		})
	}

	// 2. 获取数据库连接（带上下文日志）
	reqID := c.GetString("X-Request-ID") // 假设gin中间件注入了请求ID，便于链路追踪
	db, err := mysql.GetDB(DBInstanceAli)
	if err != nil {
		slog.Error("获取数据库连接失败",
			"request_id", reqID,
			"error", err,
			"db_instance", DBInstanceAli)
		resp(http.StatusInternalServerError, false, ErrMsgInternal, nil)
		return
	}

	// 3. 绑定并严格校验参数（利用gin的binding标签）
	var req CartSyncStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("参数绑定/校验失败",
			"request_id", reqID,
			"error", err,
			"request_body", c.Request.Body)
		resp(http.StatusBadRequest, false, fmt.Sprintf("%s：%v", ErrMsgInvalidParam, err.Error()), nil)
		return
	}

	// 4. 空列表处理（友好返回，而非直接return）
	cartItemLen := len(req.CartItems)
	if cartItemLen == 0 {
		slog.Info("购物车项列表为空", "request_id", reqID)
		resp(http.StatusOK, true, "购物车项列表为空，无需同步", nil)
		return
	}

	// 5. 构造SKU查询条件（plant_id, sku_id）元组，明确命名提高可读性
	var skuIdsPairs [][]interface{}
	for _, item := range req.CartItems {
		skuIdsPairs = append(skuIdsPairs, []interface{}{item.Id, item.SkuId})
	}

	// 6. 查询SKU库存（带上下文日志，失败时返回前端）
	var skus []PlantSku
	if err := db.Table(TablePlantSku).
		Select("plant_id, id, stock").
		Where("(plant_id, id) IN ?", skuIdsPairs).
		Find(&skus).Error; err != nil {
		slog.Error("查询SKU库存失败",
			"request_id", reqID,
			"error", err,
			"sku_pairs_count", cartItemLen)
		resp(http.StatusInternalServerError, false, ErrMsgInternal, nil)
		return
	}

	// 7. 构建库存映射表（抽离key构造逻辑，提高可维护性）
	stockMap := makeStockMap(skus)

	// 8. 处理库存修正逻辑（清晰的业务逻辑拆分）
	cartResults := make([]CartItemResult, 0, cartItemLen) // 预分配容量，提升性能
	for _, item := range req.CartItems {
		key := buildSkuKey(item.Id, item.SkuId)
		currentStock := stockMap[key] // 不存在则默认0

		// 修正数量：不超过当前库存
		newQty := item.Quantity
		if item.Quantity > currentStock {
			newQty = currentStock
			slog.Debug("购物车数量超出库存，自动修正",
				"request_id", reqID,
				"plant_id", item.Id,
				"sku_id", item.SkuId,
				"old_quantity", item.Quantity,
				"new_quantity", newQty,
				"stock", currentStock)
		}

		cartResults = append(cartResults, CartItemResult{
			Id:          item.Id,
			SkuId:       item.SkuId,
			OldQuantity: item.Quantity,
			NewQuantity: newQty,
			Stock:       currentStock,
		})
	}

	// 9. 调试日志（生产环境可通过日志级别控制）
	if len(cartResults) > 0 {
		resultJSON, err := json.MarshalIndent(cartResults, "", "  ")
		if err != nil {
			slog.Warn("序列化库存同步结果失败", "request_id", reqID, "error", err)
		} else {
			slog.Debug("购物车库存同步结果", "request_id", reqID, "result", string(resultJSON))
		}
	}

	// 10. 返回成功响应（核心：前端必须收到响应）
	resp(http.StatusOK, true, "库存同步成功", cartResults)
}

// buildSkuKey 构造库存映射表的key（抽离重复逻辑，提高可维护性）
func buildSkuKey(plantId, skuId uint64) string {
	return fmt.Sprintf("%d_%d", plantId, skuId)
}

// makeStockMap 构建SKU库存映射表（抽离逻辑，提高可读性）
func makeStockMap(skus []PlantSku) map[string]uint64 {
	stockMap := make(map[string]uint64, len(skus))
	for _, sku := range skus {
		key := buildSkuKey(sku.PlantId, sku.Id)
		stockMap[key] = sku.Stock
	}
	return stockMap
}
