package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sunzhaoc/plant_be/pkg/aliyun"
	"github.com/sunzhaoc/plant_be/pkg/db/mysql"
	"github.com/sunzhaoc/plant_be/pkg/utils"
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
	Phone    string `json:"phone" binding:"required,len=11"`          // 手机号 11位
}

type User struct {
	ID       uint   `gorm:"primaryKey"` // 手动指定主键
	Username string `gorm:"column:username;type:varchar(50);uniqueIndex;not null"`
	Email    string `gorm:"column:email;type:varchar(100);uniqueIndex;not null"`
	Phone    string `gorm:"column:phone;type:varchar(100);uniqueIndex;not null"`
	Password string `gorm:"column:password;type:varchar(100);not null"`
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
	// 1. 立即获取数据库连接，若失败直接返回
	db, err := mysql.GetDB("ali")
	if err != nil {
		log.Printf("数据库连接失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器内部错误"})
		return
	}

	// 2. 绑定并校验参数
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数校验失败",
			"error":   err.Error(),
		})
		return
	}

	// 3. 业务逻辑：检查用户名或手机号是否已存在
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM plant.users WHERE username = ? OR email = ? OR phone = ?)"
	db.Raw(query, req.Username, req.Email, req.Phone).Scan(&exists)
	if exists {
		c.JSON(http.StatusConflict, gin.H{"success": false, "message": "用户名、邮箱或手机号已存在"})
		return
	}

	// 4. 密码加密处理
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		// 加密失败
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器内部错误"})
	}
	// 4. 执行入库操作
	newUser := User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Phone:    req.Phone,
	}
	if err := db.Create(&newUser).Error; err != nil {
		log.Printf("创建用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "注册失败，请稍后再试"})
		return
	}

	// 6. 成功返回
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "注册成功"})
}
