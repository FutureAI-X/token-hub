package controller

import (
	"net/http"
	"strconv"

	"github.com/FutureAI/token-hub/model"
	"github.com/gin-gonic/gin"
)

// AdminGetUsers 管理员获取用户列表
func AdminGetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("p", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	users, total, err := model.GetUsers(page, pageSize, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取用户列表失败",
		})
		return
	}

	// 构建响应，隐藏密码等敏感信息
	type userResponse struct {
		ID          int    `json:"id"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		Role        int    `json:"role"`
		Status      int    `json:"status"`
		Email       string `json:"email"`
		Quota       int64  `json:"quota"`
		UsedQuota   int64  `json:"used_quota"`
		CreatedAt   string `json:"created_at"`
	}

	items := make([]userResponse, len(users))
	for i, u := range users {
		items[i] = userResponse{
			ID:          u.ID,
			Username:    u.Username,
			DisplayName: u.DisplayName,
			Role:        u.Role,
			Status:      u.Status,
			Email:       u.Email,
			Quota:       u.Quota,
			UsedQuota:   u.UsedQuota,
			CreatedAt:   u.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"items": items,
			"total": total,
		},
	})
}

// AdminCreateUserRequest 创建用户请求
type AdminCreateUserRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	DisplayName string `json:"display_name"`
	Role        int    `json:"role"`
	Quota       int64  `json:"quota"`
}

// AdminCreateUser 管理员创建用户
func AdminCreateUser(c *gin.Context) {
	var req AdminCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "用户名和密码不能为空",
		})
		return
	}

	// 验证角色：不允许创建 root 用户
	if req.Role >= model.RoleRootUser {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "不允许创建超级管理员用户",
		})
		return
	}

	// 默认角色为普通用户
	if req.Role == 0 {
		req.Role = model.RoleCommonUser
	}

	user := model.User{
		Username:    req.Username,
		Password:    req.Password,
		DisplayName: req.DisplayName,
		Role:        req.Role,
		Status:      model.UserStatusEnabled,
		Quota:       req.Quota,
	}

	if err := model.AdminCreateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "创建用户失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户创建成功",
	})
}

// AdminDeleteUser 管理员删除用户
func AdminDeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的用户ID",
		})
		return
	}

	// 不允许删除 root 用户
	user, err := model.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "用户不存在",
		})
		return
	}

	if user.Role >= model.RoleRootUser {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "不允许删除超级管理员",
		})
		return
	}

	if err := model.AdminDeleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "删除用户失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户已删除",
	})
}

// AdminUpdateUserStatusRequest 更新用户状态请求
type AdminUpdateUserStatusRequest struct {
	Status int `json:"status" binding:"required"`
}

// AdminUpdateUserStatus 管理员更新用户状态
func AdminUpdateUserStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的用户ID",
		})
		return
	}

	var req AdminUpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "状态值无效",
		})
		return
	}

	if req.Status != model.UserStatusEnabled && req.Status != model.UserStatusDisabled {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "状态值无效，必须为 1(启用) 或 2(禁用)",
		})
		return
	}

	// 不允许禁用 root 用户
	user, err := model.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "用户不存在",
		})
		return
	}

	if user.Role >= model.RoleRootUser {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "不允许修改超级管理员状态",
		})
		return
	}

	if err := model.UpdateUserStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新用户状态失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户状态已更新",
	})
}

// AdminAdjustUserQuotaRequest 积分调整请求
type AdminAdjustUserQuotaRequest struct {
	Mode  string `json:"mode" binding:"required"`
	Value int64  `json:"value" binding:"required"`
}

// AdminAdjustUserQuota 管理员调整用户积分
func AdminAdjustUserQuota(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的用户ID",
		})
		return
	}

	var req AdminAdjustUserQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请提供调整模式和数值",
		})
		return
	}

	if req.Mode != "add" && req.Mode != "subtract" && req.Mode != "override" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "调整模式无效，必须为 add, subtract 或 override",
		})
		return
	}

	if req.Value < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "数值不能为负数",
		})
		return
	}

	if err := model.AdjustUserQuota(id, req.Mode, req.Value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "积分调整失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "积分调整成功",
	})
}
