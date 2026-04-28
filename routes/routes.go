package routes

import (
	"bluebell/controller"
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

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, World!",
		})
	})
	return r
}
