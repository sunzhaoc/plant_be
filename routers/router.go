package routers

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sunzhaoc/plant_be/internal/api"
	"github.com/sunzhaoc/plant_be/internal/middleware"
)

func InitRouter() {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},                   // 允许的前端域名
		AllowCredentials: true,                                                // 开启允许携带凭证（Cookie）
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "PUT", "DELETE"}, // 允许的请求方法
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},        // 允许的请求头
		MaxAge:           12 * time.Hour,                                      // 预检请求的有效期（可选，默认8小时）
	}))

	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	r.GET("/api/plant-image", middleware.JWTAuthMiddleware(), api.GetPlantImageHandler)

	r.POST("/api/register", api.PostRegister)

	r.POST("/api/login", api.PostLogin)

	err := r.Run(":8080")
	if err != nil {
		return
	}
}
