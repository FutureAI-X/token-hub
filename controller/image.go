package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/FutureAI/token-hub/common"
	"github.com/FutureAI/token-hub/model"
	"github.com/FutureAI/token-hub/supplier"
	"github.com/gin-gonic/gin"
)

// ImageGenerate 图像生成端点
// POST /v1/images/generations
func ImageGenerate(c *gin.Context) {
	// 解析请求体
	var reqBody map[string]interface{}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		common.SysErrorf("[ImageGenerate] 请求体解析失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": "fail", "message": "请求体格式错误"})
		return
	}

	// 提取 model 字段
	modelName, _ := reqBody["model"].(string)
	if modelName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "fail", "message": "缺少 model 参数"})
		return
	}

	common.SysLogf("[ImageGenerate] 收到请求: model=%s", modelName)

	// 1. 根据端点路径查找端点
	endpoint, err := model.GetEndpointByPath("/v1/images/generations")
	if err != nil {
		common.SysErrorf("[ImageGenerate] 端点不存在: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"code": "fail", "message": "端点不存在"})
		return
	}

	// 2. 根据 model 名称查找模型
	m, err := model.GetModelByName(modelName)
	if err != nil {
		common.SysErrorf("[ImageGenerate] 模型不存在或已禁用: model=%s, err=%v", modelName, err)
		c.JSON(http.StatusBadRequest, gin.H{"code": "fail", "message": "模型不存在或已禁用"})
		return
	}

	// 3. 验证模型是否支持此端点
	modelEndpoints, err := model.GetModelEndpoints(m.ID)
	if err != nil {
		common.SysErrorf("[ImageGenerate] 查询端点关联失败: modelID=%d, err=%v", m.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "fail", "message": "查询端点关联失败"})
		return
	}
	supported := false
	for _, me := range modelEndpoints {
		if me.EndpointID == endpoint.ID {
			supported = true
			break
		}
	}
	if !supported {
		common.SysErrorf("[ImageGenerate] 模型不支持此端点: model=%s, endpoint=%s", modelName, endpoint.Path)
		c.JSON(http.StatusBadRequest, gin.H{"code": "fail", "message": "该模型不支持图像生成端点"})
		return
	}

	// 4. 查找可用供应商
	vendorModels, err := model.GetVendorModelsByModelID(m.ID)
	if err != nil || len(vendorModels) == 0 {
		common.SysErrorf("[ImageGenerate] 无可用供应商: model=%s, err=%v", modelName, err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": "fail", "message": "无可用供应商"})
		return
	}

	// 取第一个可用供应商
	vendorModel := vendorModels[0]
	vendor, err := model.GetVendorByID(vendorModel.VendorID)
	if err != nil {
		common.SysErrorf("[ImageGenerate] 供应商信息获取失败: vendorID=%d, err=%v", vendorModel.VendorID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "fail", "message": "供应商信息获取失败"})
		return
	}

	common.SysLogf("[ImageGenerate] 选择供应商: %s (vendorID=%d), 供应商模型ID: %s", vendor.Name, vendor.ID, vendorModel.VendorModelID)

	// 5. 解密 API Key
	apiKey, err := common.DecryptSecret(vendor.APIKey)
	if err != nil {
		common.SysErrorf("[ImageGenerate] 供应商密钥解密失败: vendor=%s, err=%v", vendor.Name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "fail", "message": "供应商密钥解密失败"})
		return
	}

	// 6. 调用供应商 API
	cfg := supplier.Config{
		BaseURL: vendor.BaseURL,
		APIKey:  apiKey,
	}
	s := supplier.NewSupplier(vendor.Name, cfg)
	if s == nil {
		common.SysErrorf("[ImageGenerate] 不支持的供应商类型: %s", vendor.Name)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "fail", "message": "不支持的供应商类型"})
		return
	}

	// 将请求体中的 model 替换为供应商侧的模型ID
	reqBody["model"] = vendorModel.VendorModelID

	result := s.ImageGenerate(supplier.ImageGenerateRequest{
		Body: reqBody,
	})

	// 7. 调用失败直接返回
	if result.Code != "success" {
		common.SysErrorf("[ImageGenerate] 供应商调用失败: vendor=%s, model=%s", vendor.Name, modelName)
		c.JSON(http.StatusOK, gin.H{"code": "fail", "message": "供应商调用失败"})
		return
	}

	// 8. 计算积分消耗
	quotaAmount := int64(0)
	quotaRule, err := model.GetQuotaRuleByModelID(m.ID)
	if err == nil && quotaRule != nil {
		quotaAmount = int64(quotaRule.BasePrice)
		// 检查是否有参数差异化定价
		if len(quotaRule.Items) > 0 {
			for _, item := range quotaRule.Items {
				if paramVal, ok := reqBody[item.ParamPath].(string); ok && paramVal == item.ParamValue {
					quotaAmount = int64(item.Price)
					break
				}
			}
		}
	}

	// 9. 获取当前用户ID
	userID := c.GetInt("userID")
	if userID == 0 {
		// 从 Token 获取用户ID
		tokenKey := c.GetHeader("Authorization")
		if tokenKey != "" {
			tokenKey = tokenKey[7:] // 移除 "Bearer " 前缀
			token, err := model.GetTokenByKey(tokenKey)
			if err == nil {
				userID = token.UserID
			}
		}
	}

	// 10. 扣除积分
	if quotaAmount > 0 && userID > 0 {
		if err := model.DeductCredits(userID, "", quotaAmount, "图像生成任务"); err != nil {
			common.SysErrorf("[ImageGenerate] 积分扣除失败: userID=%d, amount=%d, err=%v", userID, quotaAmount, err)
			c.JSON(http.StatusPaymentRequired, gin.H{"code": "fail", "message": "积分不足"})
			return
		}
	}

	// 11. 调用成功，写入任务表
	vendorRespJSON, _ := json.Marshal(result.Data)
	task := model.Task{
		TaskID:         model.GenerateTaskID(),
		UserID:         userID,
		VendorID:       vendor.ID,
		ModelID:        m.ID,
		EndpointID:     endpoint.ID,
		Status:         "submitted",
		QuotaAmount:    quotaAmount,
		VendorResponse: string(vendorRespJSON),
	}
	if err := model.CreateTask(&task); err != nil {
		common.SysErrorf("[ImageGenerate] 任务创建失败: %v", err)
		// 退还积分
		if quotaAmount > 0 && userID > 0 {
			model.RefundCredits(userID, "", quotaAmount, "任务创建失败退还")
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": "fail", "message": "任务创建失败"})
		return
	}

	// 更新积分日志的任务ID
	if quotaAmount > 0 && userID > 0 {
		model.UpdateQuotaLogTaskID(userID, task.TaskID)
	}

	common.SysLogf("[ImageGenerate] 任务创建成功: taskId=%s, vendor=%s, model=%s", task.TaskID, vendor.Name, modelName)

	// 9. 启动后台轮询任务状态
	go pollTaskStatus(task.TaskID, task.VendorResponse, vendor.Name, cfg)

	// 10. 返回系统 taskId
	c.JSON(http.StatusOK, gin.H{
		"code":   "success",
		"taskId": task.TaskID,
	})
}

// GetTask 查询任务状态
// GET /v1/tasks/:task_id
func GetTask(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"taskId": "", "status": "fail", "message": "缺少 task_id"})
		return
	}

	task, err := model.GetTaskByTaskID(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"taskId": taskID, "status": "fail", "message": "任务不存在"})
		return
	}

	// 解析 query_response 作为 data
	var data map[string]interface{}
	if task.QueryResponse != "" {
		json.Unmarshal([]byte(task.QueryResponse), &data)
	}

	c.JSON(http.StatusOK, gin.H{
		"taskId": task.TaskID,
		"status": task.Status,
		"data":   data,
	})
}

// pollTaskStatus 后台轮询供应商任务状态
func pollTaskStatus(taskID string, vendorResponse string, vendorName string, cfg supplier.Config) {
	s := supplier.NewSupplier(vendorName, cfg)
	if s == nil {
		common.SysErrorf("[TaskPoll] 不支持的供应商: %s, taskID=%s", vendorName, taskID)
		model.UpdateTaskStatusWithRefund(taskID, "call_fail", "")
		return
	}

	maxAttempts := 120 // 最多轮询120次
	interval := 5 * time.Second

	for i := 0; i < maxAttempts; i++ {
		time.Sleep(interval)

		result := s.TaskQuery(vendorResponse)
		common.SysLogf("[TaskPoll] 轮询 #%d: taskID=%s, vendorStatus=%s", i+1, taskID, result.Status)

		switch result.Status {
		case "completed":
			dataJSON, _ := json.Marshal(result.Data)
			if err := model.UpdateTaskStatus(taskID, "completed", string(dataJSON)); err != nil {
				common.SysErrorf("[TaskPoll] 更新任务状态失败: taskID=%s, err=%v", taskID, err)
			} else {
				common.SysLogf("[TaskPoll] 任务完成: taskID=%s", taskID)
			}
			return

		case "failed", "cancelled":
			if err := model.UpdateTaskStatusWithRefund(taskID, result.Status, ""); err != nil {
				common.SysErrorf("[TaskPoll] 更新任务状态失败: taskID=%s, err=%v", taskID, err)
			} else {
				common.SysLogf("[TaskPoll] 任务终止: taskID=%s, status=%s", taskID, result.Status)
			}
			return

		case "call_fail":
			common.SysErrorf("[TaskPoll] 查询调用失败，继续重试: taskID=%s", taskID)
			continue

		default:
			// pending / processing，继续轮询
			continue
		}
	}

	// 超过最大轮询次数
	common.SysErrorf("[TaskPoll] 轮询超时: taskID=%s, 已轮询%d次", taskID, maxAttempts)
	model.UpdateTaskStatusWithRefund(taskID, "call_fail", `{"error":"poll timeout"}`)
}

// RecoverPendingTasks 启动时恢复未完成任务的轮询
func RecoverPendingTasks() {
	tasks, err := model.GetPendingTasks()
	if err != nil {
		common.SysErrorf("[Recover] 查询未完成任务失败: %v", err)
		return
	}

	if len(tasks) == 0 {
		common.SysLogf("[Recover] 无未完成任务")
		return
	}

	common.SysLogf("[Recover] 发现 %d 个未完成任务，开始恢复轮询", len(tasks))

	for _, task := range tasks {
		if task.VendorResponse == "" {
			common.SysErrorf("[Recover] 任务缺少供应商响应，跳过: taskID=%s", task.TaskID)
			model.UpdateTaskStatus(task.TaskID, "call_fail", `{"error":"missing vendor_response"}`)
			continue
		}

		// 获取供应商信息
		vendor, err := model.GetVendorByID(task.VendorID)
		if err != nil {
			common.SysErrorf("[Recover] 供应商不存在，跳过: taskID=%s, vendorID=%d, err=%v", task.TaskID, task.VendorID, err)
			model.UpdateTaskStatus(task.TaskID, "call_fail", `{"error":"vendor not found"}`)
			continue
		}

		// 解密 API Key
		apiKey, err := common.DecryptSecret(vendor.APIKey)
		if err != nil {
			common.SysErrorf("[Recover] 密钥解密失败，跳过: taskID=%s, vendor=%s, err=%v", task.TaskID, vendor.Name, err)
			model.UpdateTaskStatus(task.TaskID, "call_fail", `{"error":"decrypt failed"}`)
			continue
		}

		cfg := supplier.Config{
			BaseURL: vendor.BaseURL,
			APIKey:  apiKey,
		}

		common.SysLogf("[Recover] 恢复轮询: taskID=%s, vendor=%s", task.TaskID, vendor.Name)
		go pollTaskStatus(task.TaskID, task.VendorResponse, vendor.Name, cfg)
	}
}
