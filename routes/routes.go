package routes

import (
	"bluebell/controller"
	"bluebell/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Setup() *gin.Engine {
	//初始化gin框架内置校验器使用的翻译器
	if err:=controller.InitTrans("zh");err!=nil{
		zap.L().Error("初始化翻译器失败",zap.Error(err))
	}
	r := gin.New()
	logger.Init()
	r.Use(logger.GinLogger(), logger.GinRecovery(true))

	//注册路由
	r.POST("/signup",controller.SignUpHandler)


	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, World!",
		})
	})
	return r
}
