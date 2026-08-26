package router

import (
	"github.com/FutureAI/token-hub/controller"
	"github.com/gin-gonic/gin"
)

func SetRouter(server *gin.Engine) {
	// API 路由组
	apiRouter := server.Group("/v1")
	{
		// 模型列表接口
		apiRouter.GET("/models", controller.ListModels)
	}

	// 健康检查接口
	server.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
}
