package supplier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/FutureAI/token-hub/common"
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
		common.SysErrorf("[APIMart] 请求体序列化失败: %v", err)
		return fail
	}

	// 构建 HTTP 请求
	url := fmt.Sprintf("%s/v1/images/generations", a.cfg.BaseURL)
	common.SysLogf("[APIMart] 发起图像生成请求: POST %s", url)

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		common.SysErrorf("[APIMart] 构建HTTP请求失败: %v", err)
		return fail
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.cfg.APIKey))

	// 发送请求
	resp, err := a.client.Do(httpReq)
	if err != nil {
		common.SysErrorf("[APIMart] 请求发送失败: %v", err)
		return fail
	}
	defer resp.Body.Close()

	common.SysLogf("[APIMart] 收到响应: HTTP %d", resp.StatusCode)

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		common.SysErrorf("[APIMart] 读取响应体失败: %v", err)
		return fail
	}

	// HTTP 状态码非 200 直接失败
	if resp.StatusCode != http.StatusOK {
		common.SysErrorf("[APIMart] HTTP状态码异常: %d, 响应: %s", resp.StatusCode, truncate(string(respBody), 500))
		return fail
	}

	// 解析响应
	var apiResp apimartResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		common.SysErrorf("[APIMart] 响应JSON解析失败: %v, 原始响应: %s", err, truncate(string(respBody), 500))
		return fail
	}

	// code 非 200 视为失败
	if apiResp.Code != 200 {
		common.SysErrorf("[APIMart] 业务code异常: %d, 响应: %s", apiResp.Code, truncate(string(respBody), 500))
		return fail
	}

	// 解析 data 数组，提取第一个元素的 task_id
	var tasks []apimartTaskItem
	if err := json.Unmarshal(apiResp.Data, &tasks); err != nil {
		common.SysErrorf("[APIMart] data字段解析失败: %v, data: %s", err, string(apiResp.Data))
		return fail
	}
	if len(tasks) == 0 {
		common.SysErrorf("[APIMart] data数组为空")
		return fail
	}

	common.SysLogf("[APIMart] 图像生成成功, taskId: %s", tasks[0].TaskID)

	return ImageGenerateResponse{
		Code: "success",
		Data: map[string]interface{}{
			"taskId": tasks[0].TaskID,
		},
	}
}

// apimartTaskQueryResponse APIMart 任务查询响应结构
type apimartTaskQueryResponse struct {
	Code int                  `json:"code"`
	Data apimartTaskQueryData `json:"data"`
}

type apimartTaskQueryData struct {
	ID       string              `json:"id"`
	Status   string              `json:"status"`
	Progress int                 `json:"progress"`
	Result   apimartTaskResult   `json:"result"`
}

type apimartTaskResult struct {
	Images []apimartTaskImage `json:"images"`
}

type apimartTaskImage struct {
	URL      []string `json:"url"`
	B64JSON  string   `json:"b64_json"`
}

// TaskQuery 查询 APIMart 任务状态
// vendorResponse 为提交任务时供应商返回的原始 JSON，APIMart 从中提取 taskId
func (a *apimart) TaskQuery(vendorResponse string) TaskQueryResponse {
	// 从 vendorResponse 中提取 taskId
	var respData map[string]interface{}
	if err := json.Unmarshal([]byte(vendorResponse), &respData); err != nil {
		common.SysErrorf("[APIMart] vendorResponse 解析失败: %v", err)
		return TaskQueryResponse{Status: "call_fail"}
	}
	taskID, _ := respData["taskId"].(string)
	if taskID == "" {
		common.SysErrorf("[APIMart] vendorResponse 中缺少 taskId")
		return TaskQueryResponse{Status: "call_fail"}
	}

	failResp := TaskQueryResponse{TaskID: taskID, Status: "call_fail"}

	// 构建请求
	url := fmt.Sprintf("%s/v1/tasks/%s?language=zh", a.cfg.BaseURL, taskID)
	common.SysLogf("[APIMart] 查询任务状态: GET %s", url)

	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		common.SysErrorf("[APIMart] 构建查询请求失败: %v", err)
		return failResp
	}
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.cfg.APIKey))

	// 发送请求
	resp, err := a.client.Do(httpReq)
	if err != nil {
		common.SysErrorf("[APIMart] 查询请求发送失败: %v", err)
		return failResp
	}
	defer resp.Body.Close()

	common.SysLogf("[APIMart] 查询响应: HTTP %d", resp.StatusCode)

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		common.SysErrorf("[APIMart] 读取查询响应失败: %v", err)
		return failResp
	}

	// HTTP 状态码非 200
	if resp.StatusCode != http.StatusOK {
		common.SysErrorf("[APIMart] 查询HTTP状态异常: %d, 响应: %s", resp.StatusCode, truncate(string(respBody), 500))
		return failResp
	}

	// 解析响应
	var apiResp apimartTaskQueryResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		common.SysErrorf("[APIMart] 查询响应JSON解析失败: %v, 响应: %s", err, truncate(string(respBody), 500))
		return failResp
	}

	if apiResp.Code != 200 {
		common.SysErrorf("[APIMart] 查询业务code异常: %d, 响应: %s", apiResp.Code, truncate(string(respBody), 500))
		return failResp
	}

	taskData := apiResp.Data
	common.SysLogf("[APIMart] 任务状态: %s (progress=%d)", taskData.Status, taskData.Progress)

	// 终态：completed
	if taskData.Status == "completed" {
		result := TaskQueryResponse{
			TaskID: taskID,
			Status: "completed",
			Data:   map[string]interface{}{"url": "", "b64_json": ""},
		}
		if len(taskData.Result.Images) > 0 {
			img := taskData.Result.Images[0]
			if len(img.URL) > 0 {
				result.Data["url"] = img.URL[0]
			}
			result.Data["b64_json"] = img.B64JSON
		}
		return result
	}

	// 终态：failed / cancelled
	if taskData.Status == "failed" || taskData.Status == "cancelled" {
		return TaskQueryResponse{TaskID: taskID, Status: taskData.Status}
	}

	// 中间态：pending / processing
	return TaskQueryResponse{TaskID: taskID, Status: taskData.Status}
}

// truncate 截断字符串到指定长度
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
