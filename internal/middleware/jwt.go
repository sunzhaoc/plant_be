package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sunzhaoc/plant_be/pkg/utils"
)

var ADMIN_ALLOW_LIST = []string{"御品汤包", "Utsugi"}

// JWTAuthMiddleware 验证JWT Token的中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 方式1：从Cookie获取Token
		tokenStr, err := c.Cookie("plant_token")
		// 方式2：从Authorization头获取Token（兼容前端手动携带）
		if err != nil {
			authHeader := c.GetHeader("Authorization")
			if len(authHeader) <= 7 || authHeader[:7] != "Bearer " {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "未携带有效Token",
				})
				c.Abort()
				return
			}
			tokenStr = authHeader[7:]
		}

		// 解析Token
		token, err := jwt.ParseWithClaims(tokenStr, &utils.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return utils.GetJWTSecretKey(), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Token无效或已过期",
			})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(*utils.Claims); ok {
			c.Set("userId", claims.UserID)
			c.Set("username", claims.Username)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Token解析失败",
			})
			c.Abort()

		}
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
