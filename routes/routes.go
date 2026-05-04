package routes

import (
	"bluebell/controller"
	"bluebell/middlewares"
	"bluebell/settings"

	"bluebell/logger"

	"github.com/gin-gonic/gin"
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
	v1 := r.Group("/api/v1")
	//注册路由
	v1.POST("/signup", controller.SignUpHandler)
	//登录路由
	v1.POST("/login", controller.LoginHandler)
	v1.Use(middlewares.JWTAuthMiddleware())
	{
		//社区
		v1.GET("/community", controller.CommunityHandler)
		v1.GET("/community/:id", controller.CommunityDetailHandler)

		//帖子
		v1.POST("/post", controller.CreatePostHandler)
		v1.GET("/post/:id", controller.GetPostDetailHandler)
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
