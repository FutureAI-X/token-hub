package middleware

import (
	"net/http"
	"strings"

	"github.com/FutureAI/token-hub/common"
	"github.com/FutureAI/token-hub/model"
	"github.com/gin-gonic/gin"
)

// UserAuth 用户认证中间件
func UserAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "未提供认证Token",
			})
			c.Abort()
			return
		}

		// 解析 Token
		claims, err := common.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "无效的Token",
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

// AdminAuth 管理员认证中间件
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "未提供认证Token",
			})
			c.Abort()
			return
		}

		// 解析 Token
		claims, err := common.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "无效的Token",
			})
			c.Abort()
			return
		}

		// 检查管理员权限
		if claims.Role < model.RoleAdminUser {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

// extractToken 从请求中提取 Token
func extractToken(c *gin.Context) string {
	// 从 Authorization header 获取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// 从查询参数获取
	return c.Query("token")
}
