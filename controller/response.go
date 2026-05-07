package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 是项目统一的接口响应格式。
// 绝大多数接口最终都返回：
//
//	{
//	  "code": 1000,
//	  "msg": "success",
//	  "data": ...
//	}
type Response struct {
	Code ResCode     `json:"code"` // 业务状态码
	Msg  interface{} `json:"msg"`  // 提示信息
	Data interface{} `json:"data"` // 响应数据
}

// ResponseError 返回“标准错误响应”。
// 它会使用 ResCode 自带的默认中文消息。
func ResponseError(c *gin.Context, code ResCode) {
	ResponseErrorWithMsg(c, code, code.Msg())
}

// ResponseErrorWithMsg 返回“自定义错误消息”。
// 适合参数校验失败这种场景：msg 可以直接传字段错误详情 map。
func ResponseErrorWithMsg(c *gin.Context, code ResCode, msg interface{}) {
	c.JSON(http.StatusOK, &Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

// ResponseSuccess 返回成功响应。
// data 可以是任意结构体、切片、map，也可以是 nil。
func ResponseSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, &Response{
		Code: CodeSuccess,
		Msg:  CodeSuccess.Msg(),
		Data: data,
	})
}
