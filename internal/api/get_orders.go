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

	// 7. 补充订单项数据
	var orderList []Order
	itemQuery := `
	SELECT
		plant_name,
		plant_latin_name,
		sku_size,
		main_img_url,
		price,
		quantity
	FROM plant.order_items
	WHERE order_id = ?
	ORDER BY id DESC
	;
	`

	for _, base := range orderBaseList {
		order := Order{
			OrderBase: base,
		}

		var items []OrderItem
		itemResult := db.Raw(itemQuery, base.OrderId).Scan(&items)
		if itemResult.Error != nil {
			slog.Error("获取订单订单项失败",
				slog.Any("order_id", base.OrderId),
				slog.Any("error", itemResult.Error))
			order.OrderItems = []OrderItem{}
		} else {
			order.OrderItems = items
		}

		orderList = append(orderList, order)
	}

	// 8. 返回数据
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"list":  orderList, // 当前页订单列表
			"total": total,     // 订单总条数
		},
	})
}
