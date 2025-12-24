package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sunzhaoc/plant_be/pkg/db/redis"
)

// CartItemReq 购物车单项请求参数
type CartItemReq struct {
	Id       int    `json:"id" binding:"required"`
	Size     string `json:"size"`
	Quantity int    `json:"quantity" binding:"required,min=1"`
}

// SyncReq 购物车同步请求体
type SyncReq struct {
	CartItems []CartItemReq `json:"cartItems"`
}

const (
	CartExpireTime = 7 * 24 * time.Hour
)

func SyncCartToRedis(c *gin.Context) {
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

	// 2. 绑定并校验参数
	var req SyncReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数校验失败",
			"error":   err.Error(),
		})
		return
	}

	// 3. 获取 Redis 客户端并检查错误
	rdb, err := redis.GetDb("ali")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "数据库连接失败",
			"error":   err.Error(),
		})
		return
	}

	// 4. 执行 Redis 操作
	ctx := c.Request.Context()
	redisKey := fmt.Sprintf("cart:u:%v", uid)

	if len(req.CartItems) == 0 {
		if err := rdb.Del(ctx, redisKey).Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "清除购物车失败",
				"error":   err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "购物车已清空"})
		return
	}

	pipe := rdb.Pipeline()
	pipe.Del(ctx, redisKey) // 清空旧数据：全量同步策略

	// 构建 HSet 数据 map，减少 Pipeline 中的命令数量
	cartData := make(map[string]interface{}, len(req.CartItems))
	for _, item := range req.CartItems {
		// 组合 Key: id:size
		field := fmt.Sprintf("%d:%s", item.Id, item.Size)
		cartData[field] = item.Quantity
	}
	pipe.HSet(ctx, redisKey, cartData)
	pipe.Expire(ctx, redisKey, CartExpireTime)
	if _, err := pipe.Exec(ctx); err != nil {
		slog.Error("Redis同步操作失败", "uid", uid, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "同步失败",
			"error":   err.Error(),
		})
		return
	}
	slog.Info("购物车同步成功", "uid", uid, "count", len(req.CartItems))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "同步成功",
	})
}
