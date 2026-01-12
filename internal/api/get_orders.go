package api

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sunzhaoc/plant_be/pkg/db/mysql"
)

func GetOrders(c *gin.Context) {
	slog.Info("获取历史订单数据")

	// 1. 获取并校验用户 ID
	uid, exists := c.Get("userId")
	if !exists || uid == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "用户未登录",
			"error":   "Unauthorized",
		})
		return
	}

	// 2. 获取并校验分页参数（page：当前页，默认1；pageSize：每页条数，默认10）
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	// 转换为整数并校验
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1 // 非法参数默认第一页
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 50 { // 限制最大每页50条，避免性能问题
		pageSize = 10
	}

	// 计算偏移量：OFFSET = (页码-1) * 每页条数
	offset := (page - 1) * pageSize

	// 3. 获取mysql连接池
	db, err := mysql.GetDB("ali")
	if err != nil {
		slog.Error("数据库连接失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器内部错误"})
		return
	}

	// 4. 定义结构体（保持原有结构不变）
	type OrderBase struct {
		OrderId     uint64  `json:"order_id"`
		OrderSn     string  `json:"order_sn"`
		TotalAmount float64 `json:"total_amount"`
		PayAmount   float64 `json:"pay_amount"`
		OrderStatus uint    `json:"order_status"`
		CreateTime  string  `json:"create_time"`
	}

	type OrderItem struct {
		OrderId        uint64  `json:"-"` // 新增：用于分组，不返回给前端
		PlantName      string  `json:"plant_name"`
		PlantLatinName string  `json:"plant_latin_name"`
		SkuSize        string  `json:"sku_size"`
		MainImgUrl     string  `json:"main_img_url"`
		Price          float64 `json:"price"`
		Quantity       int     `json:"quantity"`
	}

	type Order struct {
		OrderBase
		OrderItems []OrderItem `json:"order_items"`
	}

	// 5. 先查询总订单数（用于分页计算）
	var total int64
	countQuery := `SELECT COUNT(*) FROM plant.orders WHERE user_id = ?;`
	countResult := db.Raw(countQuery, uid).Scan(&total)
	if countResult.Error != nil {
		slog.Error("查询订单总数失败", slog.Any("error", countResult.Error))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取订单总数失败",
		})
		return
	}

	// 6. 查询当前页的订单基础数据
	var orderBaseList []OrderBase
	orderQuery := `
	SELECT
	    id order_id,
		order_sn,
		total_amount,
		pay_amount,
		order_status,
		DATE_FORMAT(create_time, '%Y-%m-%d %H:%i:%s') create_time
	FROM plant.orders
	WHERE user_id = ?
	ORDER BY create_time DESC
	LIMIT ? OFFSET ?
	;`
	queryResult := db.Raw(orderQuery, uid, pageSize, offset).Scan(&orderBaseList)
	if queryResult.Error != nil {
		slog.Error("获取用户的订单失败", slog.Any("error", queryResult.Error))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取用户的订单失败",
		})
		return
	}

	// 7. 批量查询订单项（核心优化点）
	// 7.1 提取当前页所有订单ID
	var orderIds []uint64
	for _, base := range orderBaseList {
		orderIds = append(orderIds, base.OrderId)
	}
	// 如果没有订单，直接返回空列表，避免无效查询
	if len(orderIds) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "",
			"data": gin.H{
				"list":  []Order{},
				"total": total,
			},
		})
		return
	}

	// 7.2 批量查询所有订单项（仅1次SQL）
	var allOrderItems []OrderItem
	itemQuery := `
	SELECT
		order_id,  -- 新增：查询订单ID，用于分组
		plant_name,
		plant_latin_name,
		sku_size,
		main_img_url,
		price,
		quantity
	FROM plant.order_items
	WHERE order_id IN (?)  -- 批量查询
	ORDER BY id DESC
	;
	`
	itemResult := db.Raw(itemQuery, orderIds).Scan(&allOrderItems)
	if itemResult.Error != nil {
		slog.Error("批量获取订单订单项失败", slog.Any("error", itemResult.Error))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取订单详情失败",
		})
		return
	}

	// 7.3 将订单项按订单ID分组（map映射）
	itemMap := make(map[uint64][]OrderItem)
	for _, item := range allOrderItems {
		itemMap[item.OrderId] = append(itemMap[item.OrderId], item)
	}

	// 8. 组装最终订单数据（从map取订单项，无循环SQL）
	var orderList []Order
	for _, base := range orderBaseList {
		order := Order{
			OrderBase:  base,
			OrderItems: itemMap[base.OrderId],
		}
		orderList = append(orderList, order)
	}

	// 9. 返回数据
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"list":  orderList, // 当前页订单列表
			"total": total,     // 订单总条数
		},
	})
}
