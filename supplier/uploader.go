package supplier

import (
	"context"
	"strings"
)

// UploadResult 图片上传结果
type UploadResult struct {
	URL         string `json:"url"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Bytes       int64  `json:"bytes"`
}

// Uploader 图片上传供应商接口，每个实现对应一个图片上传服务
type Uploader interface {
	// UploadImage 上传图片，返回托管 URL
	UploadImage(ctx context.Context, filename string, contentType string, data []byte) (*UploadResult, error)
}

// NewUploader 工厂函数，根据供应商名称返回对应上传实现（名称不区分大小写）
// 新增图片上传服务时，在此注册即可
func NewUploader(vendorName string, cfg Config) Uploader {
	switch strings.ToLower(vendorName) {
	case "apimart":
		return newAPIMartUploader(cfg)
	default:
		return nil
	}
}
