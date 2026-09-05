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
		case model.ErrUserDeleted:
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "用户已删除",
			})
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

	// 生成数据加密密钥
	dataKey, err := common.GenerateDataKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "生成密钥失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "登录成功",
		"data": gin.H{
			"token":     token,
			"data_key":  dataKey,
			"user": gin.H{
				"id":           user.ID,
				"username":     user.Username,
				"display_name": user.DisplayName,
				"role":         user.Role,
				"status":       user.Status,
				"email":        user.Email,
				"credits":      user.Credits,
				"used_credits": user.UsedCredits,
			},
		},
	})
}

// UpdateProfileRequest 更新个人资料请求
type UpdateProfileRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

// UpdateProfile 更新当前登录用户资料
func UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未登录",
		})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数无效",
		})
		return
	}

	updates := map[string]interface{}{}
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.DisplayName != "" {
		updates["display_name"] = req.DisplayName
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "没有需要更新的字段",
		})
		return
	}

	err := model.DB.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "资料更新成功",
	})
}

// ResetMyPasswordRequest 用户重置自己密码请求
type ResetMyPasswordRequest struct {
	DataKey string `json:"data_key" binding:"required"`
}

// ResetMyPassword 用户重置自己的密码（随机生成，AES-GCM 加密返回）
func ResetMyPassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "未登录"})
		return
	}

	var req ResetMyPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请提供 data_key"})
		return
	}

	newPassword := common.GenerateRandomPassword(12)

	hashed, err := common.Password2Hash(newPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "密码加密失败"})
		return
	}

	if err := model.UpdateUser(userID.(int), map[string]interface{}{"password": hashed}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "重置密码失败"})
		return
	}

	encrypted, err := common.EncryptWithKey(newPassword, req.DataKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "密码加密传输失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "密码重置成功",
		"data": gin.H{
			"encrypted_password": encrypted,
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
			"credits":      user.Credits,
			"used_credits": user.UsedCredits,
		},
	})
}
