package controller

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/FutureAI/token-hub/common"
	"github.com/FutureAI/token-hub/model"
	"github.com/FutureAI/token-hub/supplier"
	"github.com/gin-gonic/gin"
)

// maxUploadSize 最大上传文件大小（10MB）
const maxUploadSize = 10 * 1024 * 1024

// multipartOverhead multipart 边界、头字段等额外的字节余量
const multipartOverhead = 64 * 1024

// uploadTimeout 上传调用超时
const uploadTimeout = 30 * time.Second

// maxFilenameLen 文件名最大长度
const maxFilenameLen = 100

// UploadImage 图片上传端点
// POST /v1/uploads/images
// 仅一个参数：file；返回 url(来自第三方)/filename/content_type/bytes(程序解析)
func UploadImage(c *gin.Context) {
	// 1. 先限制请求体大小，再解析 multipart。
	// 顺序至关重要：c.FormFile 会触发 multipart 解析，超过内存阈值（默认 32MB）
	// 的部分会被写入临时文件；只有在解析前加上 MaxBytesReader 才能真正拦住
	// 超大请求打满磁盘/内存的情况（原来的 Size 检查发生在解析之后，为时已晚）。
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize+multipartOverhead)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"code": "fail", "message": "文件大小超过 10MB 限制"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": "fail", "message": "缺少 file 参数"})
		return
	}

	// 2. 大小限制（maxUploadSize 之外的兜底）
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

	// 4.5 清洗文件名并强制扩展名与真实类型一致。
	// 文件名会被转发给上游供应商，且可能被其用于存储路径与内容分发，
	// 必须防住 CRLF 头注入、目录穿越，以及用 .html 等扩展名做内容伪装。
	safeFilename := enforceImageExtension(sanitizeFilename(fileHeader.Filename), contentType)

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
	result, err := uploader.UploadImage(ctx, safeFilename, contentType, data)
	if err != nil {
		common.SysErrorf("[UploadImage] 上传失败: vendor=%s, err=%v", vendor.Name, err)
		c.JSON(http.StatusBadGateway, gin.H{"code": "fail", "message": "上传失败，请稍后再试"})
		return
	}

	// 8. url 来自第三方，filename/content_type/bytes 由程序解析
	c.JSON(http.StatusOK, gin.H{
		"url":          result.URL,
		"filename":     safeFilename,
		"content_type": contentType,
		"bytes":        int64(len(data)),
	})
}

// sanitizeFilename 清洗上传文件名，仅保留安全字符。
// 防住三类问题：路径分隔符（目录穿越）、控制字符与 CRLF（multipart 头注入）、
// 超长名字（上游存储截断/异常）。
func sanitizeFilename(name string) string {
	// 统一分隔符后只取最后一段
	name = strings.ReplaceAll(name, "\\", "/")
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}

	cleaned := make([]rune, 0, len(name))
	for _, r := range name {
		switch {
		case r < 0x20 || r == 0x7F: // 控制字符，含 CR / LF / TAB
			continue
		case strings.ContainsRune(`"'<>\:|?*`, r):
			continue
		default:
			cleaned = append(cleaned, r)
		}
	}

	// 去掉前导点，避免产生 "." / ".." / 隐藏文件
	result := strings.TrimLeft(strings.TrimSpace(string(cleaned)), ".")
	if result == "" {
		result = "image"
	}
	if runes := []rune(result); len(runes) > maxFilenameLen {
		result = string(runes[:maxFilenameLen])
	}
	return result
}

// enforceImageExtension 用探测到的真实 MIME 类型强制扩展名。
// 防止用户把图片命名为 .html/.svg 等内容类型，借助上游 CDN 的按扩展名
// 分发策略制造存储型 XSS。
func enforceImageExtension(name, contentType string) string {
	ext, ok := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
		"image/gif":  ".gif",
	}[contentType]
	if !ok {
		return name
	}
	if idx := strings.LastIndex(name, "."); idx > 0 {
		name = name[:idx]
	}
	return name + ext
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
