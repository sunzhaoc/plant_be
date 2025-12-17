package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sunzhaoc/plant_be/pkg/aliyun"
	"github.com/sunzhaoc/plant_be/pkg/db/mysql"
	"gorm.io/gorm"
)

// ImageResponse 图片URL响应结构
type ImageResponse struct {
	URL string `json:"url"`
}

// GetPlantImageHandler 获取植物图像的签名URL
//
// 从查询参数中获取图片URL，调用阿里云OSS服务生成带有效期的签名URL
//
// 参数:
//
//	c - gin.Context对象，包含HTTP请求上下文和查询参数
//
// 返回值:
//
//	通过JSON返回包含签名URL的响应
//	成功: 200状态码和签名URL
//	失败: 500状态码和空URL
func GetPlantImageHandler(c *gin.Context) {
	cfg := aliyun.LoadAliConfig()
	imgUrl := c.Query("imgUrl")
	signedURL, err := aliyun.GetOssUrl(cfg, imgUrl, 290, 260)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ImageResponse{URL: ""})
		return
	}
	c.JSON(http.StatusOK, ImageResponse{URL: signedURL})
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"` // 用户名必填，3-20位
	Email    string `json:"email" binding:"required,email"`           // 邮箱必填，格式验证
	Password string `json:"password" binding:"required,min=6"`        // 密码必填，至少6位
}

type User struct {
	gorm.Model        // 内置字段：ID、CreatedAt、UpdatedAt、DeletedAt
	Username   string `gorm:"column:username;type:varchar(50);uniqueIndex;not null"` // 唯一索引，非空
	Email      string `gorm:"column:email;type:varchar(100);uniqueIndex;not null"`   // 唯一索引，非空
	Password   string `gorm:"column:password;type:varchar(100);not null"`            // 加密后的密码
}

// PostRegister 处理用户注册请求
//
// 参数:
//
//	c *gin.Context: Gin 框架的上下文对象
//
// 返回:
//
//	返回 JSON 格式的测试消息
func PostRegister(c *gin.Context) {
	db, err := mysql.GetDB("ali")
	fmt.Println("开始处理注册")
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": fmt.Sprintf("参数错误：%v", err),
		})
		return
	}

	fmt.Println(req.Username)
	fmt.Println(req.Email)
	fmt.Println(req.Password)

	var existingUser User
	if err := db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户名已存在",
		})
		return
	}

	if err != nil {
		log.Printf("获取数据库连接失败: %v", err)
		c.JSON(http.StatusInternalServerError, ImageResponse{URL: ""})
	}
	//fmt.Println(db)
	c.JSON(http.StatusOK, gin.H{"message": "test"})
}
