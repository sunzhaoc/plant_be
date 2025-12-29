package routers

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/natefinch/lumberjack"
	"github.com/sunzhaoc/plant_be/internal/api"
	"github.com/sunzhaoc/plant_be/internal/middleware"
)

// setupLogger 配置Gin日志输出到文件并实现拆分
func setupLogger() {
	// 配置日志轮转规则
	logWriter := &lumberjack.Logger{
		Filename:   "./logs/plant_be.log", // 日志文件路径
		MaxSize:    100,                   // 单个日志文件最大大小（MB）
		MaxBackups: 10,                    // 保留的旧日志文件最大数量
		MaxAge:     7,                     // 日志文件保留天数
		Compress:   true,                  // 是否压缩旧日志文件
	}

	// 将Gin的日志输出重定向到轮转的日志文件
	//gin.DefaultWriter = logWriter
	// 如果需要同时输出到控制台+文件，可以用io.MultiWriter
	gin.DefaultWriter = io.MultiWriter(os.Stdout, logWriter)
	gin.ForceConsoleColor()
}

func InitRouter() {
	// 第一步：初始化日志配置（输出到文件并拆分）
	setupLogger()

	// 第二步：创建Gin引擎（保留默认的Recovery中间件，但日志已替换为文件输出）
	r := gin.Default()

	// 第三步：全局使用IP黑名单中间件（也可针对特定路由单独使用）
	r.Use(middleware.IpBlackMiddleware())

	// 第四步：配置CORS（保留原有配置）
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://antplant.store/", "http://antplant.store/", "http://localhost:5173"}, // 允许的前端域名
		AllowCredentials: true,                                                                                   // 开启允许携带凭证（Cookie）
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "PUT", "DELETE"},                                    // 允许的请求方法
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},                                           // 允许的请求头
		MaxAge:           12 * time.Hour,                                                                         // 预检请求的有效期（可选，默认8小时）
	}))

	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	//r.GET("/api/plant-image", api.GetPlantImageHandler)

	r.GET("/api/plants", api.GetPlants)

	r.GET("/api/plant-detail/:plantId", middleware.JWTAuthMiddleware(), api.GetPlantDetail)

	r.POST("/api/cart/sync-stock", middleware.JWTAuthMiddleware(), api.SyncCartStock)

	r.POST("/api/register", api.PostRegister)

	r.POST("/api/login", api.PostLogin)

	r.POST("/api/cart/sync-redis", middleware.JWTAuthMiddleware(), api.SyncCartToRedis)

	err := r.Run(":8080")
	if err != nil {
		return
	}
}
