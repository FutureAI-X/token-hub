package router

import (
	"github.com/FutureAI/token-hub/controller"
	"github.com/FutureAI/token-hub/middleware"
	"github.com/gin-gonic/gin"
)

func SetRouter(server *gin.Engine) {
	// 健康检查接口
	server.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// 认证相关接口（无需登录）
	authRouter := server.Group("/api/auth")
	{
		authRouter.POST("/login", controller.Login)
	}

	// 公开接口（无需登录）
	pricingRouter := server.Group("/api/pricing")
	{
		pricingRouter.GET("", controller.GetPricing)
	}

	// 需要登录的接口
	userRouter := server.Group("/api/user")
	userRouter.Use(middleware.UserAuth())
	{
		userRouter.GET("/info", controller.GetUserInfo)
		userRouter.PUT("/email", controller.UpdateEmail)
	}

	// API 路由组（OpenAI 兼容格式）
	apiRouter := server.Group("/v1")
	{
		// 模型列表接口
		apiRouter.GET("/models", controller.ListModels)
	}

	// 管理员接口
	adminRouter := server.Group("/api/admin")
	adminRouter.Use(middleware.RootAuth())
	{
		adminRouter.GET("/users", controller.AdminGetUsers)
		adminRouter.POST("/users", controller.AdminCreateUser)
		adminRouter.DELETE("/users/:id", controller.AdminDeleteUser)
		adminRouter.PUT("/users/:id/status", controller.AdminUpdateUserStatus)
		adminRouter.PUT("/users/:id/quota", controller.AdminAdjustUserQuota)
	}
}
