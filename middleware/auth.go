package middleware

import (
	"net/http"
	"strings"
	"time"

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

// RootAuth 超级管理员认证中间件（仅 role=100）
func RootAuth() gin.HandlerFunc {
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

		claims, err := common.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "无效的Token",
			})
			c.Abort()
			return
		}

		if claims.Role < model.RoleRootUser {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

// APIAuth API Key 认证中间件
// 校验 /v1 接口的 Authorization: Bearer <token>
// <token> 对应 api_keys 表中的 key 字段（即 API Key）
func APIAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := extractToken(c)
		if key == "" {
			abortAPIAuth(c, "未提供 API Key", "authentication_error")
			return
		}

		// 根据 Key 获取 Token（GetTokenByKey 已过滤 status=1 启用状态）
		token, err := model.GetTokenByKey(key)
		if err != nil {
			abortAPIAuth(c, "无效的 API Key", "authentication_error")
			return
		}

		// 检查是否过期（-1 表示永不过期）
		if token.ExpiredTime != -1 && token.ExpiredTime < time.Now().Unix() {
			abortAPIAuth(c, "API Key 已过期", "authentication_error")
			return
		}

		// 将身份信息存入上下文供后续处理使用
		c.Set("user_id", token.UserID)
		c.Set("token_id", token.ID)
		c.Set("api_key", token.Key)

		c.Next()
	}
}

// abortAPIAuth 以 OpenAI 兼容格式终止 API 请求
func abortAPIAuth(c *gin.Context, message, errorType string) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"error": gin.H{
			"message": message,
			"type":    errorType,
		},
	})
	c.Abort()
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
