package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 定义IP黑名单列表（可扩展为从配置文件/数据库读取）
var ipBlackList = map[string]bool{
	//"127.0.0.1": false, // 示例：放行本地IP
	"124.126.139.34": true, // 示例：拉黑这个IP
	//"10.0.0.5":      true,  // 示例：拉黑这个IP
}

// IPBlacklistMiddleware IP黑名单中间件
func IpBlackMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		if ipBlackList[clientIP] {
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
