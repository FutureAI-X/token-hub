package controller

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/FutureAI/token-hub/common"
	"github.com/FutureAI/token-hub/model"
	"github.com/FutureAI/token-hub/supplier"
	"github.com/gin-gonic/gin"
)

// maxUploadSize 最大上传文件大小（10MB）
const maxUploadSize = 10 * 1024 * 1024

// uploadTimeout 上传调用超时
const uploadTimeout = 30 * time.Second

// UploadImage 图片上传端点
// POST /v1/uploads/images
// 仅一个参数：file；返回 url(来自第三方)/filename/content_type/bytes(程序解析)
func UploadImage(c *gin.Context) {
	// 1. 解析上传的文件字段 file
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "fail", "message": "缺少 file 参数"})
		return
	}

	// 2. 大小限制
	if fileHeader.Size > maxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"code": "fail", "message": "文件大小超过 10MB 限制"})
		return
	}

	// 3. 读取文件内容
	f, err := fileHeader.Open()
	if err != nil {
		common.SysErrorf("[UploadImage] 打开文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "fail", "message": "文件读取失败"})
		return
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		common.SysErrorf("[UploadImage] 读取文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "fail", "message": "文件读取失败"})
		return
	}

	// 4. 识别图片类型（仅支持 JPEG/PNG/WebP/GIF）
	contentType := detectImageType(data)
	if contentType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "fail", "message": "仅支持 JPEG/PNG/WebP/GIF 图片"})
		return
	}

	// 5. 根据供应商名称（忽略大小写）获取 APIMart
	vendor, err := model.GetEnabledVendorByNameInsensitive("APIMart")
	if err != nil {
		common.SysErrorf("[UploadImage] 未找到启用的 APIMart 供应商: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": "fail", "message": "图片上传服务暂不可用"})
		return
	}

	// 6. 解密供应商 API Key
	apiKey, err := common.DecryptSecret(vendor.APIKey)
	if err != nil {
		common.SysErrorf("[UploadImage] 供应商密钥解密失败: vendor=%s, err=%v", vendor.Name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "fail", "message": "服务暂时不可用，请稍后再试"})
		return
	}

	// 7. 调用上传服务
	cfg := supplier.Config{
		BaseURL: vendor.BaseURL,
		APIKey:  apiKey,
	}
	uploader := supplier.NewUploader(vendor.Name, cfg)
	if uploader == nil {
		common.SysErrorf("[UploadImage] 不支持的图片上传服务: %s", vendor.Name)
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": "fail", "message": "图片上传服务暂不可用"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), uploadTimeout)
	defer cancel()
	result, err := uploader.UploadImage(ctx, fileHeader.Filename, contentType, data)
	if err != nil {
		common.SysErrorf("[UploadImage] 上传失败: vendor=%s, err=%v", vendor.Name, err)
		c.JSON(http.StatusBadGateway, gin.H{"code": "fail", "message": "上传失败，请稍后再试"})
		return
	}

	// 8. url 来自第三方，filename/content_type/bytes 由程序解析
	c.JSON(http.StatusOK, gin.H{
		"url":          result.URL,
		"filename":     fileHeader.Filename,
		"content_type": contentType,
		"bytes":        int64(len(data)),
	})
}

// detectImageType 通过文件头（magic bytes）识别图片 MIME 类型，不支持则返回空
func detectImageType(data []byte) string {
	if len(data) < 12 {
		return ""
	}
	switch {
	case len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF:
		// JPEG
		return "image/jpeg"
	case len(data) >= 8 && bytes.HasPrefix(data, []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}):
		// PNG
		return "image/png"
	case len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP":
		// WebP
		return "image/webp"
	case len(data) >= 6 && (string(data[:6]) == "GIF87a" || string(data[:6]) == "GIF89a"):
		// GIF
		return "image/gif"
	default:
		return ""
	}
}
