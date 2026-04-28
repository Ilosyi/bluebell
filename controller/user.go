package controller

import (
	"bluebell/logic"
	"bluebell/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// 用户注册
func SignUpHandler(c *gin.Context) {
	//1.获取参数和参数校验
	//json:{"username":"xxx","password":"xxx","re_password":"xxx"}
	p := new(models.SignUpParam)
	if err := c.ShouldBindJSON(p); err != nil {
		//请求参数有误，直接返回响应，记录日志
		zap.L().Error("注册请求参数有误", zap.Error(err))
		//判断err是不是validator.ValidationErrors类型
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			//翻译错误
			"error": removeTopStruct(errs.Translate(trans)),
		})
		return
	}
	//2.业务处理
	logic.SignUp()
	//3.返回响应
	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}
