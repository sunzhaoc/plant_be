package gift

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sunzhaoc/plant_be/pkg/db/mysql"
	"github.com/sunzhaoc/plant_be/pkg/db/mysql/models"
	"github.com/sunzhaoc/plant_be/pkg/db/redis"
)

// GetTodayLotteryDraw 抽奖接口
// 规则：
// 1. 每个用户每小时最多抽 1 次
// 2. 中奖概率 100%
// 3. 全局每天只能中 1 次奖，一旦有人中，所有人当天都不能再中
func GetTodayLotteryDraw(c *gin.Context) {
	// 1. 获取用户ID
	uid, exists := c.Get("userId")
	//fmt.Print("中奖用户：", uid) // TODO 调试用
	if !exists || uid == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "用户未登录",
			"error":   "Unauthorized",
		})
		return
	}
	// 2. 获取Redis
	rdb, err := redis.GetDb("ali")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Redis连接失败",
			"error":   err.Error(),
		})
		return
	}

	// 3. 基础变量
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	now := time.Now()
	today := now.Format("20060102")                                        // 日期维度（用于全局中奖标记）
	hourKey := now.Format("2006010215")                                    // 小时维度（2006010215 代表2006年1月2日15点）
	userHourlyKey := fmt.Sprintf("lottery:user:%v:count:%s", uid, hourKey) // 用户当前小时抽奖次数
	globalWinKey := "lottery:global:win:" + today                          // 全局今天是否已中奖
	expireHour := 1 * time.Hour                                            // 过期时间 1 小时
	expireDay := 24 * time.Hour                                            // 全局中奖标记过期时间 1 天

	// --------------------------
	// 规则1：先判断今天全局是否已经有人中奖
	// --------------------------
	globalExists, _ := rdb.Exists(ctx, globalWinKey).Result()
	if globalExists > 0 {
		// 今天已经有人中奖 → 所有人都不能再中
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"is_win":  false,
			"message": "今日大奖已被抽走，欢迎明天再来",
		})
		return
	}

	// --------------------------
	// 规则2：判断用户当前小时是否已经抽过（最多1次）
	// --------------------------
	count, _ := rdb.Get(ctx, userHourlyKey).Int64()
	if count >= 1 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"is_win":  false,
			"message": "本小时抽奖次数已用完（每小时仅限1次），请下一小时再试",
		})
		return
	}

	// --------------------------
	// 次数+1（原子操作）
	// --------------------------
	rdb.Incr(ctx, userHourlyKey)
	rdb.Expire(ctx, userHourlyKey, expireHour) // 小时维度键1小时后过期

	// --------------------------
	// 规则3：1% 概率中奖（修正原逻辑错误）
	// --------------------------
	rand.Seed(time.Now().UnixNano())
	randomNum := rand.Intn(100) + 1 // 0~99
	isWin := randomNum <= 50

	// --------------------------
	// 如果中奖：设置全局中奖标记（今天所有人都不能再中）
	// --------------------------
	if isWin {
		rdb.SetNX(ctx, globalWinKey, "1", expireDay)
	}

	// --------------------------
	// 返回结果
	// --------------------------
	winCode := uuid.New().String()
	msg := "谢谢参与"
	if isWin {
		msg = "恭喜您中奖了！"
		db, err := mysql.GetDB("ali")
		if err != nil {
			slog.Error("数据库连接失败", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器内部错误"})
			return
		}
		lotteryWinningRecord := models.LotteryWinningRecord{
			UserId:      uint64(uid.(uint)),
			ActivityId:  1,
			WinningTime: time.Now(),
			WinCode:     winCode,
			Remark:      "每日抽奖活动",
		}
		if err := db.Create(&lotteryWinningRecord).Error; err != nil {
			slog.Error("记录中奖信息失败", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "注册失败，请稍后再试"})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"is_win":       isWin,
		"message":      msg,
		"win_code":     winCode,
		"hour_count":   count + 1,           // 当前小时第几次抽奖（固定为1）
		"current_hour": now.Format("15:00"), // 当前小时（便于前端展示）
	})
}
