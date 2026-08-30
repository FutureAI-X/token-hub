package controller

import (
	"encoding/json"
	"net/http"

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
		c.JSON(http.StatusBadRequest, gin.H{"code": "fail", "message": "请求体格式错误"})
		return
	}

	// 提取 model 字段
	modelName, _ := reqBody["model"].(string)
	if modelName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "fail", "message": "缺少 model 参数"})
		return
	}

	// 1. 根据端点路径查找端点
	endpoint, err := model.GetEndpointByPath("/v1/images/generations")
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "fail", "message": "端点不存在"})
		return
	}

	// 2. 根据 model 名称查找模型
	m, err := model.GetModelByName(modelName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "fail", "message": "模型不存在或已禁用"})
		return
	}

	// 3. 验证模型是否支持此端点
	modelEndpoints, err := model.GetModelEndpoints(m.ID)
	if err != nil {
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
		c.JSON(http.StatusBadRequest, gin.H{"code": "fail", "message": "该模型不支持图像生成端点"})
		return
	}

	// 4. 查找可用供应商
	vendorModels, err := model.GetVendorModelsByModelID(m.ID)
	if err != nil || len(vendorModels) == 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": "fail", "message": "无可用供应商"})
		return
	}

	// 取第一个可用供应商
	vendorModel := vendorModels[0]
	vendor, err := model.GetVendorByID(vendorModel.VendorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "fail", "message": "供应商信息获取失败"})
		return
	}

	// 5. 解密 API Key
	apiKey, err := common.DecryptSecret(vendor.APIKey)
	if err != nil {
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
		c.JSON(http.StatusOK, gin.H{"code": "fail", "message": "供应商调用失败"})
		return
	}

	// 8. 调用成功，写入任务表
	vendorRespJSON, _ := json.Marshal(result.Data)
	task := model.Task{
		TaskID:         model.GenerateTaskID(),
		VendorID:       vendor.ID,
		ModelID:        m.ID,
		EndpointID:     endpoint.ID,
		Status:         "submitted",
		VendorResponse: string(vendorRespJSON),
	}
	if err := model.CreateTask(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "fail", "message": "任务创建失败"})
		return
	}

	// 9. 返回系统 taskId
	c.JSON(http.StatusOK, gin.H{
		"code":   "success",
		"taskId": task.TaskID,
	})
}
