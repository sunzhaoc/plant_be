package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// 定义配置文件结构体
type IPBlacklistConfig struct {
	Blacklist []string `json:"blacklist"`
}

// 全局变量：内存中的黑名单（加读写锁保证并发安全）
var (
	ipBlackList   = make(map[string]bool)
	ipListMutex   sync.RWMutex
	configPath    = "config/ip_black_list.yaml" // 配置文件路径
	refreshPeriod = 60 * time.Second            // 定时刷新间隔（30秒）
)

// 初始化函数：程序启动时加载配置+启动定时刷新
func init() {
	// 加载初始黑名单
	if err := loadIPBlacklist(); err != nil {
		fmt.Printf("加载IP黑名单配置失败: %v\n", err)
	}

	go startIPBlacklistRefresh()
}

// 从配置文件加载黑名单到内存
func loadIPBlacklist() error {
	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析JSON
	var config IPBlacklistConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 加写锁更新内存中的黑名单
	ipListMutex.Lock()
	defer ipListMutex.Unlock()

	// 清空旧数据，写入新数据
	clear(ipBlackList)
	for _, ip := range config.Blacklist {
		ipBlackList[ip] = true
	}
	return nil
}

// 启动定时刷新任务
func startIPBlacklistRefresh() {
	ticker := time.NewTicker(refreshPeriod)
	defer ticker.Stop()

	for range ticker.C {
		if err := loadIPBlacklist(); err != nil {
			fmt.Printf("定时刷新IP黑名单失败: %v\n", err)
		}
	}
}

// IPBlacklistMiddleware IP黑名单中间件（逻辑不变，仅加读锁）
func IpBlackMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// 加读锁读取黑名单（不阻塞写操作，保证并发安全）
		ipListMutex.RLock()
		isBlocked := ipBlackList[clientIP]
		ipListMutex.RUnlock()

		if isBlocked {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "您的IP已被限制访问",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
