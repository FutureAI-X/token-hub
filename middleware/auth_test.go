package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestContext(target string, headers map[string]string) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", target, nil)
	for k, v := range headers {
		c.Request.Header.Set(k, v)
	}
	return c
}

func TestExtractTokenFromBearerHeader(t *testing.T) {
	c := newTestContext("/", map[string]string{"Authorization": "Bearer sk-abc123"})
	if got := extractToken(c); got != "sk-abc123" {
		t.Errorf("got %q, want %q", got, "sk-abc123")
	}
}

func TestExtractTokenIsCaseInsensitiveOnScheme(t *testing.T) {
	for _, scheme := range []string{"Bearer", "bearer", "BEARER", "BeArEr"} {
		c := newTestContext("/", map[string]string{"Authorization": scheme + " sk-abc"})
		if got := extractToken(c); got != "sk-abc" {
			t.Errorf("scheme %q: got %q, want %q", scheme, got, "sk-abc")
		}
	}
}

// TestExtractTokenIgnoresQueryParameter 是本项目的一处重要修复：
// 旧实现会回落到 c.Query("token")，使 JWT / API Key 出现在 URL 中，
// 进而被访问日志、反向代理日志、浏览器历史与 Referer 头记录下来。
func TestExtractTokenIgnoresQueryParameter(t *testing.T) {
	c := newTestContext("/v1/models?token=sk-leaked-in-url", nil)
	if got := extractToken(c); got != "" {
		t.Errorf("不得从查询参数提取凭证，实际得到 %q", got)
	}
}

func TestExtractTokenRejectsMalformedHeader(t *testing.T) {
	cases := map[string]string{
		"缺少 scheme":  "sk-abc123",
		"scheme 不匹配": "Basic dXNlcjpwYXNz",
		"只有 scheme":  "Bearer",
		"空 header":   "",
	}
	for name, header := range cases {
		t.Run(name, func(t *testing.T) {
			c := newTestContext("/", map[string]string{"Authorization": header})
			if got := extractToken(c); got != "" {
				t.Errorf("应返回空字符串，实际得到 %q", got)
			}
		})
	}
}

func TestExtractTokenTrimsWhitespace(t *testing.T) {
	c := newTestContext("/", map[string]string{"Authorization": "Bearer   sk-abc   "})
	if got := extractToken(c); got != "sk-abc" {
		t.Errorf("got %q, want %q", got, "sk-abc")
	}
}

// TestExtractTokenKeepsEmbeddedSpaces 确保只按第一个空格切分，
// 不会把密钥中被误加的空格截断成无效 Key。
func TestExtractTokenDoesNotSplitBeyondFirstSpace(t *testing.T) {
	c := newTestContext("/", map[string]string{"Authorization": "Bearer sk-a b"})
	if got := extractToken(c); got != "sk-a b" {
		t.Errorf("got %q, want %q", got, "sk-a b")
	}
}
