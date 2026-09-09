package main

import (
	"log"
	"os"
	"strconv"

	"github.com/FutureAI/token-hub/common"
	"github.com/FutureAI/token-hub/controller"
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

	// 初始化数据库
	if err := model.InitDB(); err != nil {
		common.FatalLog("failed to initialize database: " + err.Error())
	}
	defer model.CloseDB()

	// 安全提示：检测弱/默认密钥
	if s := os.Getenv("JWT_SECRET"); s == "" || s == "token-hub-jwt-secret-change-me" {
		common.SysErrorf("[安全] JWT_SECRET 未设置或为默认值，JWT 可被伪造，请设置强随机密钥!")
	}
	if s := os.Getenv("SECRET_KEY"); s == "" || s == "token-hub-secret-change-me" {
		common.SysErrorf("[安全] SECRET_KEY 未设置或为默认值，敏感数据加密可被破解，请设置强随机密钥!")
	}

	// 恢复未完成任务的轮询
	go controller.RecoverPendingTasks()

	// 设置 Gin 模式
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建 Gin 引擎
	server := gin.New()

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
	server.Use(gin.Logger())

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
