package controller

import (
	"bluebell/logic"
	"bluebell/models"
	"errors"
	"fmt"
	"strings"

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
		if verr, ok := errors.AsType[validator.ValidationErrors](err); ok {
			detail := removeTopStruct(verr.Translate(trans))
			zap.L().Debug("注册参数校验失败", zap.Any("detail", detail))
			ResponseErrorWithMsg(c, CodeInvalidParam, detail)
			return
		}

		// 不是校验错误，直接返回原始错误信息
		ResponseErrorWithMsg(c, CodeInvalidParam, err.Error())
		return
	}
	//2.业务处理
	if err := logic.SignUp(p); err != nil {
		switch {
		case strings.Contains(err.Error(), "用户已存在"):
			ResponseError(c, CodeUserExist)
		default:
			ResponseError(c, CodeServerBusy)
		}
		return
	}
	//3.返回响应
	ResponseSuccess(c, gin.H{"message": "ok"})
}

// LoginHandler  处理登录请求
func LoginHandler(c *gin.Context) {
	//1.获取参数和参数校验
	p := new(models.LoginParam)
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Error("登录请求参数有误", zap.Error(err))
		if verr, ok := errors.AsType[validator.ValidationErrors](err); ok {
			detail := removeTopStruct(verr.Translate(trans))
			zap.L().Debug("登录参数校验失败", zap.Any("detail", detail))
			ResponseErrorWithMsg(c, CodeInvalidParam, detail)
			return
		}
		//不是校验错误
		ResponseErrorWithMsg(c, CodeInvalidParam, err.Error())
		return
	}
	//2.业务处理
	user, err := logic.LoginUp(p)
	if err != nil {
		zap.L().Error("登录失败", zap.Error(err))
		switch {
		case strings.Contains(err.Error(), "用户名或密码错误"):
			ResponseError(c, CodeInvalidPassword)
		default:
			ResponseError(c, CodeServerBusy)
		}
		return
	}
	//3.返回成功响应
	ResponseSuccess(c, gin.H{
		"user_id":   fmt.Sprintf("%d", user.UserID), // id值大于1<<53-1  int64类型的最大值是1<<63-1
		"user_name": user.Username,
		"token":     user.Token,
	})
}
