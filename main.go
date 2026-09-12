package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/FutureAI/token-hub/common"
	"github.com/FutureAI/token-hub/controller"
	"github.com/FutureAI/token-hub/middleware"
	"github.com/FutureAI/token-hub/model"
	"github.com/FutureAI/token-hub/router"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		common.SysLog("No .env file found, using environment variables")
	}

	// 安全前置校验：密钥必须存在且足够强，否则拒绝启动（fail-closed）。
	// 绝不能回落到可预测的默认值——那会让 JWT 可被伪造、供应商密钥可被解密。
	for _, name := range []string{"JWT_SECRET", "SECRET_KEY"} {
		if _, err := common.RequireSecret(name); err != nil {
			common.FatalLog("[安全] " + err.Error())
		}
	}

	// 初始化数据库
	if err := model.InitDB(); err != nil {
		common.FatalLog("failed to initialize database: " + err.Error())
	}
	defer model.CloseDB()

	// 恢复未完成任务的轮询
	go controller.RecoverPendingTasks()

	// 设置 Gin 模式（默认 release；仅在显式 GIN_MODE=debug 时开启调试）
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建 Gin 引擎
	server := gin.New()

	// 信任代理配置。
	// gin 默认信任所有代理（0.0.0.0/0），导致 c.ClientIP() 无条件采信客户端自带的
	// X-Forwarded-For，登录限流可被伪造 IP 绕过。因此默认不信任任何代理头，
	// 仅在显式配置 TRUSTED_PROXIES 时才信任指定 CIDR。
	trustedProxies := parseTrustedProxies(os.Getenv("TRUSTED_PROXIES"))
	if err := server.SetTrustedProxies(trustedProxies); err != nil {
		common.FatalLog("TRUSTED_PROXIES 配置无效: " + err.Error())
	}
	if len(trustedProxies) == 0 {
		common.SysLog("[安全] TRUSTED_PROXIES 未配置，将直接使用 TCP 对端地址作为客户端 IP")
	} else {
		common.SysLogf("[安全] 信任的代理网段: %s", strings.Join(trustedProxies, ", "))
	}

	// 安全响应头
	server.Use(middleware.SecurityHeaders())

	// 添加 Recovery 中间件（不向客户端泄露 panic 细节）
	server.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		log.Printf("[PANIC] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		c.JSON(500, gin.H{
			"error": gin.H{
				"message": "Internal server error",
				"type":    "server_error",
			},
		})
	}))

	// 添加 Logger 中间件
	// SkipQueryString 必须为 true：token / data_key 等敏感值历史上曾通过查询参数传递，
	// 若记录查询串会把凭证写进访问日志、代理日志与日志聚合平台。
	server.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipQueryString: true,
	}))

	// 设置路由
	router.SetRouter(server)

	// 获取端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	// 验证端口
	if _, err := strconv.Atoi(port); err != nil {
		log.Fatalf("invalid PORT value: %s", port)
	}

	// 启动服务器
	common.SysLogf("Token Hub started on port %s", port)
	if err := server.Run(":" + port); err != nil {
		common.FatalLog("failed to start server: " + err.Error())
	}
}

// parseTrustedProxies 解析 TRUSTED_PROXIES（逗号分隔的 IP 或 CIDR）。
// 返回空切片表示不信任任何代理头，此时 gin 使用 TCP 对端地址。
func parseTrustedProxies(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			result = append(result, p)
		}
	}
	return result
}
