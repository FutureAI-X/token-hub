package main

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// clientIPVia 构造一个最小 gin 引擎，返回它在给定配置与请求下解析出的客户端 IP。
func clientIPVia(t *testing.T, trustedProxies string, remoteAddr string, xff string) string {
	t.Helper()
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	if err := engine.SetTrustedProxies(parseTrustedProxies(trustedProxies)); err != nil {
		t.Fatalf("SetTrustedProxies(%q) 失败: %v", trustedProxies, err)
	}

	var got string
	engine.GET("/ip", func(c *gin.Context) { got = c.ClientIP() })

	req := httptest.NewRequest("GET", "/ip", nil)
	req.RemoteAddr = remoteAddr
	if xff != "" {
		req.Header.Set("X-Forwarded-For", xff)
	}
	engine.ServeHTTP(httptest.NewRecorder(), req)
	return got
}

// TestNoTrustedProxiesIgnoresForwardedHeader 是本项目的关键安全属性：
// 未显式配置 TRUSTED_PROXIES 时，客户端自带的 X-Forwarded-For 必须被完全忽略，
// 否则攻击者可以为每个请求伪造一个不同的 IP，从而完全绕过登录限流。
func TestNoTrustedProxiesIgnoresForwardedHeader(t *testing.T) {
	got := clientIPVia(t, "", "203.0.113.7:12345", "1.2.3.4")
	if got != "203.0.113.7" {
		t.Errorf("未配置信任代理时不得采信 X-Forwarded-For：got %q, want %q", got, "203.0.113.7")
	}
}

// TestNoTrustedProxiesIgnoresSpoofedChain 伪造 XFF 链同样必须被忽略
func TestNoTrustedProxiesIgnoresSpoofedChain(t *testing.T) {
	got := clientIPVia(t, "", "203.0.113.7:12345", "10.0.0.1, 10.0.0.2, 10.0.0.3")
	if got != "203.0.113.7" {
		t.Errorf("got %q, want %q", got, "203.0.113.7")
	}
}

// TestLoopbackProxyHonorsForwardedHeader 配置 127.0.0.1/32 后
// （即 Nginx 同机部署的标准配置），应能从 XFF 还原真实客户端 IP。
func TestLoopbackProxyHonorsForwardedHeader(t *testing.T) {
	got := clientIPVia(t, "127.0.0.1/32", "127.0.0.1:5000", "203.0.113.7")
	if got != "203.0.113.7" {
		t.Errorf("got %q, want %q", got, "203.0.113.7")
	}
}

// TestLoopbackProxyIgnoresClientSuppliedPrefix 是最重要的一条：
// Nginx 使用 $proxy_add_x_forwarded_for 会把真实客户端 IP 追加到链尾，
// 客户端预置的伪造值排在链首。Gin 从右向左扫描，遇到第一个不在信任列表中的
// IP 就返回——因此伪造值必须被忽略，取到的是 Nginx 追加的真实 IP。
func TestLoopbackProxyIgnoresClientSuppliedPrefix(t *testing.T) {
	got := clientIPVia(t, "127.0.0.1/32", "127.0.0.1:5000", "1.2.3.4, 203.0.113.7")
	if got != "203.0.113.7" {
		t.Errorf("客户端预置的伪造值未被忽略：got %q, want %q", got, "203.0.113.7")
	}
}

// TestUntrustedSourceCannotSpoofEvenWhenProxiesConfigured
// 反向代理之外的直连来源（非回环地址）即使配置了信任代理也不能伪造 IP，
// 这保证了一旦 3001 端口意外暴露，攻击者仍无法绕过限流。
func TestUntrustedSourceCannotSpoofEvenWhenProxiesConfigured(t *testing.T) {
	got := clientIPVia(t, "127.0.0.1/32", "203.0.113.99:12345", "1.2.3.4")
	if got != "203.0.113.99" {
		t.Errorf("非受信来源不得采信 XFF：got %q, want %q", got, "203.0.113.99")
	}
}

// TestParseTrustedProxies 校验环境变量解析
func TestParseTrustedProxies(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", []string{}},
		{"   ", []string{}},
		{",", []string{}},
		{"127.0.0.1/32", []string{"127.0.0.1/32"}},
		{"127.0.0.1/32,10.0.0.0/8", []string{"127.0.0.1/32", "10.0.0.0/8"}},
		{" 127.0.0.1/32 , 10.0.0.0/8 ", []string{"127.0.0.1/32", "10.0.0.0/8"}},
		{"127.0.0.1/32,,10.0.0.0/8,", []string{"127.0.0.1/32", "10.0.0.0/8"}},
	}
	for _, tc := range cases {
		got := parseTrustedProxies(tc.in)
		if len(got) != len(tc.want) {
			t.Errorf("parseTrustedProxies(%q) = %v, want %v", tc.in, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("parseTrustedProxies(%q)[%d] = %q, want %q", tc.in, i, got[i], tc.want[i])
			}
		}
	}
}

// TestInvalidTrustedProxiesIsRejected 非法网段必须报错，
// 以便 main 中的 FatalLog 能拦住配置错误，而不是静默降级。
func TestInvalidTrustedProxiesIsRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	for _, bad := range []string{"not-an-ip", "999.999.999.999", "10.0.0.0/99"} {
		if err := engine.SetTrustedProxies(parseTrustedProxies(bad)); err == nil {
			t.Errorf("非法网段 %q 应被拒绝", bad)
		}
	}
}
