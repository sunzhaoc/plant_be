package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sunzhaoc/plant_be/pkg/db/redis"
)

// 1. 购物车新增/修改项请求结构体（明确语义，专用于新增/更新场景）
type CartAddOrUpdateItemReq struct {
	Id       int    `json:"id" binding:"required"`
	Size     string `json:"size"`
	Quantity int    `json:"quantity" binding:"required,min=1"`
}

// 2. 购物车删除项请求结构体（独立封装，专用于删除场景）
type CartDeleteItemReq struct {
	Id   int    `json:"id" binding:"required"` // 商品ID
	Size string `json:"size"`                  // 商品规格（删除需匹配ID+规格）
}

// 3. 购物车增量同步顶层请求结构体（聚合上述两个结构体，作为接口入参）
type CartIncrementalSyncReq struct {
	AddedOrUpdatedItems []CartAddOrUpdateItemReq `json:"addedOrUpdatedItems"` // 新增/修改的项
	DeletedItems        []CartDeleteItemReq      `json:"deletedItems"`
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
	var req CartIncrementalSyncReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数校验失败",
			"error":   err.Error(),
		})
		return
	}

	fmt.Println("req:", req)

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

	// 若无任何增量数据，直接返回成功
	if len(req.AddedOrUpdatedItems) == 0 && len(req.DeletedItems) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "无增量数据，无需同步",
		})
		return
	}

	pipe := rdb.Pipeline()

	// 处理新增/修改项
	if len(req.AddedOrUpdatedItems) > 0 {
		cartData := make(map[string]interface{}, len(req.AddedOrUpdatedItems))
		for _, item := range req.AddedOrUpdatedItems {
			field := fmt.Sprintf("%d:%s", item.Id, item.Size)
			cartData[field] = item.Quantity
		}
		pipe.HSet(ctx, redisKey, cartData)
	}

	// 处理删除项（HDel 删除指定字段）
	if len(req.DeletedItems) > 0 {
		delFields := make([]string, len(req.DeletedItems))
		for i, item := range req.DeletedItems {
			delFields[i] = fmt.Sprintf("%d:%s", item.Id, item.Size)
		}
		pipe.HDel(ctx, redisKey, delFields...)
	}

	pipe.Expire(ctx, redisKey, CartExpireTime)

	if _, err := pipe.Exec(ctx); err != nil {
		slog.Error("Redis增量同步操作失败", "uid", uid, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "增量同步失败",
			"error":   err.Error(),
		})
		return
	}

	// 日志记录同步结果
	slog.Info("购物车增量同步成功",
		"uid", uid,
		"addedOrUpdatedCount", len(req.AddedOrUpdatedItems),
		"deletedCount", len(req.DeletedItems))

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "增量同步成功",
		"data": gin.H{
			"addedOrUpdatedCount": len(req.AddedOrUpdatedItems),
			"deletedCount":        len(req.DeletedItems),
		},
	})
}
