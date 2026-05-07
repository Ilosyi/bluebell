package routes

import (
	"bluebell/controller"
	"bluebell/middlewares"
	"bluebell/settings"

	"bluebell/logger"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "bluebell/docs"
)

func Setup() *gin.Engine {
	// 如果配置是 release 模式，就把 Gin 切到发布模式。
	// 发布模式会减少调试输出，更适合线上环境。
	if settings.GlobalConfig.App.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 让控制台日志带颜色，开发时更容易读。
	gin.ForceConsoleColor()

	// 创建一个“空白”的 Gin 引擎。
	// gin.New() 不会自动挂载 Logger/Recovery，所以我们下面自己挂自定义版本。
	r := gin.New()
	var rateLimitCfg *settings.RateLimitConfig
	if settings.GlobalConfig != nil {
		rateLimitCfg = &settings.GlobalConfig.RateLimit
	}
	r.Use(
		logger.GinLogger(),
		logger.GinRecovery(true),
		middlewares.RateLimitMiddleware(rateLimitCfg))

	// Swagger 文档入口。
	// 生产环境通常不应该暴露接口文档，因此这里通过配置开关控制。
	if settings.GlobalConfig != nil && settings.GlobalConfig.App.EnableSwagger {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// 所有业务接口统一挂在 /api/v1 下。
	v1 := r.Group("/api/v1")
	// 最简单的健康检查接口。
	v1.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// 公开接口：注册、登录、社区、公开帖子列表/详情。
	v1.POST("/signup", controller.SignUpHandler)
	v1.POST("/login", controller.LoginHandler)
	v1.GET("/community", controller.CommunityHandler)
	v1.GET("/community/:id", controller.CommunityDetailHandler)
	v1.GET("/post/:id", controller.GetPostDetailHandler)
	v1.GET("/posts", controller.GetPostListHandler)
	v1.GET("/posts2", controller.GetPostListHandler2)

	// 下面开始挂登录后才能访问的接口。
	// 注意：对同一个路由组调用 Use 后，之后定义的接口都会自动经过 JWT 中间件。
	v1.Use(middlewares.JWTAuthMiddleware())
	{
		// 当前登录用户资料。
		v1.GET("/me", controller.GetMeHandler)
		v1.PUT("/me", controller.UpdateMeHandler)

		// 帖子管理。
		v1.POST("/post", controller.CreatePostHandler)
		v1.POST("/post/draft", controller.CreateDraftHandler)
		v1.GET("/my/posts", controller.GetMyPostListHandler)
		v1.GET("/my/posts/:id", controller.GetManagePostHandler)
		v1.PUT("/post/:id", controller.UpdatePostHandler)
		v1.PUT("/post/:id/draft", controller.UpdateDraftHandler)
		v1.POST("/post/:id/publish", controller.PublishDraftHandler)
		v1.DELETE("/post/:id", controller.DeletePostHandler)

		// 投票。
		v1.POST("/vote", controller.PostVoteHandler)
	}

	return r
}
