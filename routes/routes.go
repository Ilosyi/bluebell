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
	r.Use(logger.GinLogger(), logger.GinRecovery(true))

	//注册路由
	r.POST("/signup", controller.SignUpHandler)
	//登录路由
	r.POST("/login", controller.LoginHandler)
	r.GET("/ping", middlewares.JWTAuthMiddleware(), func(c *gin.Context) {
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
