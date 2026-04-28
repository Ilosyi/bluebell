package controller

import (
	"bluebell/logic"
	"bluebell/models"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// SignUpHandler 用户注册
func SignUpHandler(c *gin.Context) {
	//1.获取参数和参数校验
	//json:{"username":"xxx","password":"xxx","re_password":"xxx"}
	p := new(models.SignUpParam)
	if err := c.ShouldBindJSON(p); err != nil {
		// 请求参数有误，直接返回响应，记录日志
		zap.L().Error("注册请求参数有误", zap.Error(err))

		// Go1.13+ 的 errors 包支持错误链（wrapped errors），
		// 直接做类型断言在包装错误上会失败，应使用 errors.As 来判断是否包含 validator.ValidationErrors
		if err, ok := errors.AsType[validator.ValidationErrors](err); ok {
			c.JSON(http.StatusBadRequest, gin.H{
				// 翻译错误
				"error": removeTopStruct(err.Translate(trans)),
			})
			return
		}

		// 不是校验错误，直接返回原始错误信息
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	//2.业务处理
	if err := logic.SignUp(p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	//3.返回响应
	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}
