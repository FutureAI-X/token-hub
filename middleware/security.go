package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders 为所有响应添加基础安全响应头。
// 服务本身不提供静态文件托管，因此这里以「收紧浏览器行为」为主。
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		// 禁止浏览器对响应体做 MIME 类型嗅探
		h.Set("X-Content-Type-Options", "nosniff")
		// 禁止被嵌入 iframe，防点击劫持
		h.Set("X-Frame-Options", "DENY")
		// 不向外部站点泄露完整 URL（含路径中的资源 ID）
		h.Set("Referrer-Policy", "no-referrer")
		// API 服务不需要任何浏览器特性
		h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		// API 响应不应被缓存（可能包含账户数据）
		if c.GetHeader("Cache-Control") == "" {
			h.Set("Cache-Control", "no-store")
		}

		c.Next()
	}
}

// HSTS 仅在 HTTPS 部署时才有意义，因此由反向代理（Nginx/CDN）负责下发，
// 不在此处硬编码，避免本地 HTTP 开发时把浏览器锁死到 HTTPS。
