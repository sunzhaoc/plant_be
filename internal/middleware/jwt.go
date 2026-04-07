package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sunzhaoc/plant_be/pkg/utils"
)

var ADMIN_ALLOW_LIST = []string{"御品汤包", "Utsugi"}

// JWTAuthMiddleware 验证JWT Token的中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenStr string
		var useCookie bool

		// 方式1：从Cookie获取Token
		if cookieToken, err := c.Cookie("plant_token"); err == nil {
			tokenStr = cookieToken
			useCookie = true
		} else {
			// 方式2：从Authorization头获取Token（兼容前端手动携带）
			slog.Warn("从Authorization头获取Token")
			authHeader := c.GetHeader("Authorization")
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				tokenStr = authHeader[7:]
				useCookie = false
			}
		}

		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "未携带有效Token"})
			c.Abort()
			return
		}

		// 2. 解析 Token
		claims := &utils.Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return utils.GetJWTSecretKey(), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Token无效或已过期"})
			c.Abort()
			return
		}

		// 3. 自动刷新逻辑（核心修改）
		if utils.ShouldRefresh(claims) {
			newToken, err := utils.GenerateToken(claims.UserID, claims.Username)
			if err == nil {
				// 如果原本是 Cookie 方式，自动写回 Cookie（前端无感）
				if useCookie {
					c.SetCookie("plant_token", newToken, int(utils.TokenExpire.Seconds()), "/", "", false, true)
				}
				// 同时也放在 Header 中，方便前端通过拦截器手动更新（如果是 Header 方式）
				//c.Header("New-Token", newToken)
				//c.Header("Access-Control-Expose-Headers", "New-Token")
			}
		}

		// 4. 设置上下文
		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, exists := c.Get("username")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "用户信息未解析",
			})
			c.Abort()
			return
		}

		// 校验用户名是否在管理员白名单中
		usernameStr, ok := username.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "用户名格式错误",
			})
			c.Abort()
			return
		}

		isAdmin := false
		for _, admin := range ADMIN_ALLOW_LIST {
			if admin == usernameStr {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "无管理员权限，禁止访问",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
