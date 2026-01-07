package api

import (
	"log/slog"
	"net/http"

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

	// 2. 获取mysql连接池
	db, err := mysql.GetDB("ali")
	if err != nil {
		slog.Error("数据库连接失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器内部错误"})
		return
	}

	// 3. 定义结构体：拆分订单基础信息和包含订单项的完整订单
	// 只用于接收orders表的基础数据（无嵌套字段，避免GORM扫描错误）
	type OrderBase struct {
		OrderId     uint64  `json:"order_id"`
		OrderSn     string  `json:"order_sn"`
		TotalAmount float64 `json:"total_amount"`
		PayAmount   float64 `json:"pay_amount"`
		OrderStatus uint    `json:"order_status"`
		CreateTime  string  `json:"create_time"`
	}

	// 订单项结构体（不变）
	type OrderItem struct {
		PlantName      string  `json:"plant_name"`
		PlantLatinName string  `json:"plant_latin_name"`
		SkuSize        string  `json:"sku_size"`
		MainImgUrl     string  `json:"main_img_url"`
		Price          float64 `json:"price"`
		Quantity       int     `json:"quantity"`
	}

	// 最终返回的完整订单结构体（包含订单项）
	type Order struct {
		OrderBase              // 嵌入基础订单信息
		OrderItems []OrderItem `json:"order_items"` // 订单项列表
	}

	// 查询订单基础数据
	var orderBaseList []OrderBase
	query := `
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
	LIMIT 20
	;`
	queryResult := db.Debug().Raw(query, uid).Scan(&orderBaseList)
	if queryResult.Error != nil {
		slog.Error("获取用户的订单失败", slog.Any("error", queryResult.Error))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取用户的订单失败",
		})
		return
	}

	// 构建完整的订单列表（包含订单项）
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

	// 遍历基础订单，补充订单项
	for _, base := range orderBaseList {
		// 初始化完整订单，先赋值基础信息
		order := Order{
			OrderBase: base,
		}

		// 查询当前订单的订单项
		var items []OrderItem
		itemResult := db.Debug().Raw(itemQuery, base.OrderId).Scan(&items)
		if itemResult.Error != nil {
			slog.Error("获取订单订单项失败",
				slog.Any("order_id", base.OrderId),
				slog.Any("error", itemResult.Error))
			// 即使订单项查询失败，也保留基础订单信息
			order.OrderItems = []OrderItem{}
		} else {
			order.OrderItems = items
		}

		orderList = append(orderList, order)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    orderList,
	})
}
