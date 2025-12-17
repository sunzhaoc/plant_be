package api

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sunzhaoc/plant_be/pkg/db/mysql"
	"github.com/sunzhaoc/plant_be/pkg/utils"
)

type LoginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required,min=6"` // 密码必填，至少6位
}

func PostLogin(c *gin.Context) {
	db, err := mysql.GetDB("ali")
	if err != nil {
		slog.Info("数据库连接失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器内部错误"})
		return
	}

	// 2. 绑定并校验参数
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数校验失败",
			"error":   err.Error(),
		})
		return
	}

	var hashedPassword string
	query := "SELECT password FROM plant.users WHERE username = ? OR email = ? OR phone = ? LIMIT 1;"
	query_result := db.Raw(query, req.Account, req.Account, req.Account).Scan(&hashedPassword)
	if query_result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "数据库服务异常",
		})
		return
	}
	if query_result.RowsAffected == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "账号不存在",
		})
		return
	}

	if !utils.CheckPasswordHash(req.Password, hashedPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "账号或密码错误",
		})
		return
	}
	slog.Info("登录成功")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "登录成功",
	})
}
