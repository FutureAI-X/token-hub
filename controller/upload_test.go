package controller

import (
	"strings"
	"testing"
)

// ── sanitizeFilename：阻断路径穿越与 multipart 头注入 ──

func TestSanitizeFilenameStripsCRLF(t *testing.T) {
	// 这是 M2 的实际攻击载荷：RFC 2231 解码后 CR/LF 会进入文件名
	evil := "evil\r\nX-Injected: pwned\r\n"
	got := sanitizeFilename(evil)

	if strings.ContainsAny(got, "\r\n") {
		t.Errorf("结果仍含 CR/LF，可造成 multipart 头注入: %q", got)
	}
	if strings.Contains(got, ":") {
		t.Errorf("结果仍含冒号，可构造伪头部: %q", got)
	}
}

func TestSanitizeFilenameStripsPathTraversal(t *testing.T) {
	cases := []string{
		"../../etc/passwd",
		"..\\..\\windows\\system32\\config",
		"/absolute/path/image.png",
		"C:\\Users\\evil\\image.png",
	}
	for _, in := range cases {
		got := sanitizeFilename(in)
		if strings.ContainsAny(got, `/\`) {
			t.Errorf("sanitizeFilename(%q) = %q，仍含路径分隔符", in, got)
		}
		if strings.HasPrefix(got, ".") {
			t.Errorf("sanitizeFilename(%q) = %q，仍为隐藏/相对路径", in, got)
		}
	}
}

func TestSanitizeFilenameHandlesDegenerateInput(t *testing.T) {
	cases := map[string]string{
		"":          "image",
		"..":        "image",
		".":         "image",
		"...":       "image",
		"\r\n":      "image",
		"   ":       "image",
		"../../../": "image",
	}
	for in, want := range cases {
		if got := sanitizeFilename(in); got != want {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSanitizeFilenameTruncatesLongNames(t *testing.T) {
	long := strings.Repeat("a", 500) + ".png"
	got := sanitizeFilename(long)

	if len([]rune(got)) > maxFilenameLen {
		t.Errorf("文件名长度 %d 超过上限 %d", len([]rune(got)), maxFilenameLen)
	}
}

func TestSanitizeFilenamePreservesNormalNames(t *testing.T) {
	for _, name := range []string{"photo.png", "my-image_01.jpg", "截图2026.png"} {
		if got := sanitizeFilename(name); got != name {
			t.Errorf("正常文件名不应被改动: sanitizeFilename(%q) = %q", name, got)
		}
	}
}

// ── enforceImageExtension：强制扩展名与真实类型一致 ──

// TestEnforceImageExtensionBlocksHTMLDisguise 覆盖一个真实风险：
// 用户可以把 HTML/GIF 多态文件命名为 .html，借助上游 CDN 的按扩展名分发
// 策略制造存储型 XSS。
func TestEnforceImageExtensionBlocksHTMLDisguise(t *testing.T) {
	cases := []struct {
		name        string
		contentType string
		wantExt     string
	}{
		{"evil.html", "image/png", ".png"},
		{"evil.svg", "image/png", ".png"},
		{"evil.htm", "image/jpeg", ".jpg"},
		{"evil.js", "image/gif", ".gif"},
		{"noext", "image/webp", ".webp"},
	}
	for _, tc := range cases {
		got := enforceImageExtension(tc.name, tc.contentType)
		if !strings.HasSuffix(got, tc.wantExt) {
			t.Errorf("enforceImageExtension(%q, %q) = %q, 期望以 %s 结尾",
				tc.name, tc.contentType, got, tc.wantExt)
		}
		if strings.Contains(got, ".html") || strings.Contains(got, ".svg") {
			t.Errorf("危险扩展名未被替换: %q", got)
		}
	}
}

func TestEnforceImageExtensionReplacesExistingExtension(t *testing.T) {
	got := enforceImageExtension("photo.bmp", "image/png")
	if got != "photo.png" {
		t.Errorf("got %q, want %q", got, "photo.png")
	}
}

func TestEnforceImageExtensionKeepsBaseName(t *testing.T) {
	got := enforceImageExtension("holiday-photo", "image/jpeg")
	if got != "holiday-photo.jpg" {
		t.Errorf("got %q, want %q", got, "holiday-photo.jpg")
	}
}

// ── detectImageType：magic bytes 识别 ──

func TestDetectImageType(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		want string
	}{
		{"JPEG", []byte{0xFF, 0xD8, 0xFF, 0xE0, 0, 0, 0, 0, 0, 0, 0, 0}, "image/jpeg"},
		{"PNG", []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A, 0, 0, 0, 0}, "image/png"},
		{"GIF87a", []byte("GIF87a\x00\x00\x00\x00\x00\x00"), "image/gif"},
		{"GIF89a", []byte("GIF89a\x00\x00\x00\x00\x00\x00"), "image/gif"},
		{"HTML", []byte("<!DOCTYPE html><html>"), ""},
		{"SVG", []byte("<svg xmlns=\"http://www.w3.org/2000/svg\">"), ""},
		{"too short", []byte{0xFF, 0xD8}, ""},
		{"empty", nil, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := detectImageType(tc.data); got != tc.want {
				t.Errorf("detectImageType() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDetectImageTypeWebP(t *testing.T) {
	// RIFF....WEBP
	data := []byte("RIFF\x00\x00\x00\x00WEBPVP8 ")
	if got := detectImageType(data); got != "image/webp" {
		t.Errorf("got %q, want %q", got, "image/webp")
	}
}
