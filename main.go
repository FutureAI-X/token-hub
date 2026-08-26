package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/FutureAI/token-hub/router"
	"github.com/gin-gonic/gin"
)

func main() {
	// 设置 Gin 模式
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建 Gin 引擎
	server := gin.New()

	// 添加 Recovery 中间件
	server.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		log.Printf("panic detected: %v", err)
		c.JSON(500, gin.H{
			"error": gin.H{
				"message": fmt.Sprintf("Internal server error: %v", err),
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
		port = "3000"
	}

	// 验证端口
	if _, err := strconv.Atoi(port); err != nil {
		log.Fatalf("invalid PORT value: %s", port)
	}

	// 启动服务器
	log.Printf("Token Hub started on port %s", port)
	if err := server.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
