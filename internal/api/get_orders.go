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

	// 3.
	type Order struct {
		OrderSn     string  `json:"order_sn"`
		TotalAmount float64 `json:"total_amount"`
		PayAmount   float64 `json:"pay_amount"`
		OrderStatus uint    `json:"order_status"`
	}
	var orderList []Order
	query := `
	SELECT
		order_sn,
		total_amount,
		pay_amount,
		order_status
	FROM plant.orders
	WHERE user_id = ?
	ORDER BY create_time DESC
	;`
	queryResult := db.Debug().Raw(query, uid).Scan(&orderList)
	if queryResult.Error != nil {
		slog.Error("获取用户的订单失败", slog.Any("error", queryResult.Error))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取用户的订单失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    orderList,
	})
}
