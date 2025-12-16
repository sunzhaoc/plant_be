package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sunzhaoc/plant_be/internal/api"
)

func InitRouter() {
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		// 允许前端域名（生产环境替换为你的前端实际域名，如http://localhost:3000）
		c.Header("Access-Control-Allow-Origin", "*")
		// 允许的请求方法（必须包含OPTIONS，fetch会先发OPTIONS预检请求）
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		// 允许的请求头
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
		// 处理OPTIONS预检请求（fetch必触发，不处理会拦截）
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent) // 204
			return
		}
		c.Next()
	})

	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	r.GET("/api/plant-image", api.GetPlantImageHandler)

	r.POST("/api/register", api.PostRegister)

	r.Run(":8080")
}
