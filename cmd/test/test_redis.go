package main

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/sunzhaoc/plant_be/pkg/db/redis"
)

func main() {
	SyncCartRedisToMySQL()
}

func SyncCartRedisToMySQL() {
	ctx := context.Background()
	startTime := time.Now()
	slog.Info("开始执行购物车Redis -> MySQL同步任务")

	// 1. 获取Redis客户端
	rdb, err := redis.GetDb("ali")
	if err != nil {
		slog.Error("获取Redis客户端失败", "error", err)
		return
	}

	// 2. 扫描Redis中所有购物车Key（cart:u:xxx）
	var cartKeys []string
	iter := rdb.ScanIterator(ctx, 0, "cart:u:*", 100) // 每次扫描100个Key，分页获取
	for iter.Next(ctx) {
		key := iter.Val()
		cartKeys = append(cartKeys, key)
	}
	if err := iter.Err(); err != nil {
		slog.Error("扫描Redis购物车Key失败", "error", err)
		return
	}
	if len(cartKeys) == 0 {
		slog.Info("Redis中无购物车数据，无需同步")
		return
	}
	slog.Info("扫描到待同步的购物车Key数量", "count", len(cartKeys))

	// 3. 遍历每个用户的购物车，批量同步
	syncUserCount := 0
	failUserCount := 0
	for _, key := range cartKeys {
		// 解析用户ID（从key: cart:u:123中提取123）
		uidStr := strings.TrimPrefix(key, "cart:u:")
		uid, err := strconv.ParseUint(uidStr, 10, 64)
		if err != nil {
			slog.Warn("解析购物车Key中的用户ID失败", "key", key, "error", err)
			continue
		}
		// 单用户同步（带事务保证数据一致性）
		if err := syncSingleUserCart(ctx, rdb, key, uid); err != nil {
			slog.Error("单用户购物车同步失败", "uid", uid, "key", key, "error", err)
			failUserCount++
			continue
		}
		syncUserCount++
	}

	// 4. 打印同步统计信息
	costTime := time.Since(startTime)
	slog.Info("购物车Redis->MySQL同步任务执行完成",
		"totalUserCount", len(cartKeys),
		"successUserCount", syncUserCount,
		"failUserCount", failUserCount,
		"costTime", costTime.String(),
	)
}
