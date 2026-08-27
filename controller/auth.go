package controller

import (
	"net/http"

	"github.com/FutureAI/token-hub/common"
	"github.com/FutureAI/token-hub/model"
	"github.com/gin-gonic/gin"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 用户登录
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "用户名和密码不能为空",
		})
		return
	}

	// 验证用户名密码
	user := model.User{
		Username: req.Username,
		Password: req.Password,
	}

	if err := user.ValidateAndFill(); err != nil {
		switch err {
		case model.ErrUserDisabled:
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "用户已被禁用",
			})
		default:
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "用户名或密码错误",
			})
		}
		return
	}

	// 生成 JWT Token
	token, err := common.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "生成Token失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "登录成功",
		"data": gin.H{
			"token": token,
			"user": gin.H{
				"id":           user.ID,
				"username":     user.Username,
				"display_name": user.DisplayName,
				"role":         user.Role,
				"status":       user.Status,
				"email":        user.Email,
				"quota":        user.Quota,
				"used_quota":   user.UsedQuota,
			},
		},
	})
}

// GetUserInfo 获取当前登录用户信息
func GetUserInfo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未登录",
		})
		return
	}

	user, err := model.GetUserByID(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取用户信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":           user.ID,
			"username":     user.Username,
			"display_name": user.DisplayName,
			"role":         user.Role,
			"status":       user.Status,
			"email":        user.Email,
			"quota":        user.Quota,
			"used_quota":   user.UsedQuota,
		},
	})
}
