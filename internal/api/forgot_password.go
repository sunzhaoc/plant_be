package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sunzhaoc/plant_be/pkg/db/mysql"
	"github.com/sunzhaoc/plant_be/pkg/email"
	"github.com/sunzhaoc/plant_be/pkg/utils"
)

type GetVerifyCodeRequest struct {
	Email string `json:"email" binding:"required"`
}

// SendVerificationCodeHandler 发送重置密码验证码
func SendVerificationCodeHandler(c *gin.Context) {
	db, err := mysql.GetDB("ali")
	if err != nil {
		slog.Info("数据库连接失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器内部错误"})
		return
	}

	// 2. 绑定并校验参数
	var req GetVerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数校验失败",
			"error":   err.Error(),
		})
		return
	}

	// 查询这个邮件是否存在
	type UserTable struct {
		Email string
	}
	var user UserTable
	query := "SELECT `email` FROM plant.users WHERE `email` = ? LIMIT 1;"
	result := db.Raw(query, req.Email).Scan(&user)
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

	// 2. 生成验证码
	code := utils.GenerateVerifyCode()

	// 3. 存储验证码到Redis
	ctx := context.Background()
	if err := utils.SetVerifyCode(ctx, req.Email, code); err != nil {
		slog.Error("存储验证码到Redis失败", "error", err.Error(), "email", req.Email)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "验证码存储失败",
		})
		return
	}

	// 4. 发送验证码邮件
	if err := email.SendVerificationCode(req.Email, code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "验证码发送失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "验证码已发送，请查收邮箱",
	})
}

// ResetPasswordHandler 重置密码
func ResetPasswordHandler(c *gin.Context) {
	type Request struct {
		Email            string `json:"email" binding:"required,email"`
		VerificationCode string `json:"code" binding:"required,len=6"`
		NewPassword      string `json:"newPassword" binding:"required,min=6"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数错误：" + err.Error(),
		})
		return
	}

	// 1. 验证验证码
	ctx := context.Background()
	code, err := utils.GetVerifyCode(ctx, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "服务器内部错误",
		})
		return
	}
	slog.Warn(code)
	if code == "" || code != req.VerificationCode {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "验证码错误或已过期",
		})
		return
	}

	// 2. 加密新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		// 加密失败
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器内部错误"})
	}

	slog.Info(hashedPassword)
	// 3. 更新数据库密码
	db, err := mysql.GetDB("ali")
	if err != nil {
		//slog("数据库连接失败", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器内部错误"})
		return
	}
	updateQuery := "UPDATE plant.users SET password = ? WHERE email = ?;"
	result := db.Exec(updateQuery, hashedPassword, req.Email)
	if result.Error != nil {
		slog.Error("更新密码失败", "error", result.Error, "email", req.Email)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "密码重置失败",
		})
		return
	}
	if result.RowsAffected == 0 {
		slog.Warn("密码更新无影响行数", "email", req.Email)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "用户不存在或密码未变更",
		})
		return
	}

	// 4. 删除已使用的验证码
	_ = utils.DelVerifyCode(ctx, req.Email)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "密码重置成功，请使用新密码登录",
	})
}
