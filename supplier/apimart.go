package supplier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// apimart APIMart 供应商实现
type apimart struct {
	cfg    Config
	client *http.Client
}

// newAPIMart 创建 APIMart 供应商实例
func newAPIMart(cfg Config) *apimart {
	return &apimart{
		cfg: cfg,
		client: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
	}
}

// apimartResponse APIMart API 原始响应结构
type apimartResponse struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

// apimartTaskItem APIMart 任务数据项
type apimartTaskItem struct {
	Status string `json:"status"`
	TaskID string `json:"task_id"`
}

// ImageGenerate 调用 APIMart 图像生成 API
func (a *apimart) ImageGenerate(req ImageGenerateRequest) ImageGenerateResponse {
	fail := ImageGenerateResponse{Code: "fail"}

	// 构建请求体
	bodyBytes, err := json.Marshal(req.Body)
	if err != nil {
		return fail
	}

	// 构建 HTTP 请求
	url := fmt.Sprintf("%s/v1/images/generations", a.cfg.BaseURL)
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fail
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.cfg.APIKey))

	// 发送请求
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return fail
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fail
	}

	// HTTP 状态码非 200 直接失败
	if resp.StatusCode != http.StatusOK {
		return fail
	}

	// 解析响应
	var apiResp apimartResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fail
	}

	// code 非 200 视为失败
	if apiResp.Code != 200 {
		return fail
	}

	// 解析 data 数组，提取第一个元素的 task_id
	var tasks []apimartTaskItem
	if err := json.Unmarshal(apiResp.Data, &tasks); err != nil {
		return fail
	}
	if len(tasks) == 0 {
		return fail
	}

	return ImageGenerateResponse{
		Code: "success",
		Data: map[string]interface{}{
			"taskId": tasks[0].TaskID,
		},
	}
}
