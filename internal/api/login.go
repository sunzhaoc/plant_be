package api

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sunzhaoc/plant_be/pkg/db/mysql"
	"github.com/sunzhaoc/plant_be/pkg/db/mysql/models"
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

	type UserTable = struct {
		ID       uint
		Username string
		Email    string
		Phone    string
		Password string
	}
	var user UserTable
	query := "SELECT id, username, email, password FROM plant.users WHERE username = ? OR email = ? OR phone = ? LIMIT 1;"
	result := db.Raw(query, req.Account, req.Account, req.Account).Scan(&user)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "数据库服务异常",
		})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "账号不存在",
		})
		return
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "账号或密码错误",
		})
		return
	}

	// 保存用户的登录数据
	newUserLogin := models.UserLogin{
		UserId: user.ID,
	}
	if err := db.Create(&newUserLogin).Error; err != nil {
		slog.Info(fmt.Sprintf("用户【%v】登录数据保存失败: %v", newUserLogin.UserId, err))
		return
	}

	// 生成 JWT token
	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		slog.Error(fmt.Sprintf("生成[%d][%v]的JWT Token失败", user.ID, user.Username), err)
		return
	}

	// 设置HttpOnly Cookie
	c.SetCookie(
		"plant_token",                    // Cookie名称
		token,                            // Cookie值（Token）
		int(utils.TokenExpire.Seconds()), // Token 过期时间（秒）
		"/",                              // 生效路径（全站）
		"",                               // 生效域名（空表示当前域名）
		true,                             // 是否开启HTTPS（生产环境建议true，开发环境false）
		true,                             // 是否开启HttpOnly（防止XSS攻击，无法通过JS获取）
	)

	slog.Info(fmt.Sprintf("[%d][%v]登录成功", user.ID, user.Username))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "登录成功",
		//"token":   token, // 响应体中返回 Token
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"phone":    user.Phone,
		},
	})
}
