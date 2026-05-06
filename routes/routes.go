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
	if settings.GlobalConfig.App.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	gin.ForceConsoleColor() // 强制开启颜色
	r := gin.New()
	r.Use(
		logger.GinLogger(),
		logger.GinRecovery(true))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	v1 := r.Group("/api/v1")
	//注册路由
	v1.POST("/signup", controller.SignUpHandler)
	//登录路由
	v1.POST("/login", controller.LoginHandler)
	//社区
	v1.GET("/community", controller.CommunityHandler)
	v1.GET("/community/:id", controller.CommunityDetailHandler)

	//帖子

	v1.GET("/post/:id", controller.GetPostDetailHandler)
	//分页展示帖子列表
	v1.GET("/posts", controller.GetPostListHandler)
	//按时间或分数排序的帖子列表
	v1.GET("/posts2", controller.GetPostListHandler2)

	v1.Use(middlewares.JWTAuthMiddleware())
	{
		v1.POST("/post", controller.CreatePostHandler)
		//投票
		v1.POST("/vote", controller.PostVoteHandler)
	}

	v1.GET("/ping", middlewares.JWTAuthMiddleware(), func(c *gin.Context) {
		userID, _ := c.Get(controller.CtxUserIDKey)
		username, _ := c.Get(controller.CtxUsernameKey)
		controller.ResponseSuccess(c, gin.H{
			"message":  "pong",
			"user_id":  userID,
			"username": username,
		})
	})
	return r
}
