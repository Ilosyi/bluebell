package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构体
type Response struct {
	Code ResCode     `json:"code"` // 业务状态码
	Msg  interface{} `json:"msg"`  // 提示信息
	Data interface{} `json:"data"` // 响应数据
}

func ResponseError(c *gin.Context, code ResCode) {
	ResponseErrorWithMsg(c, code, code.Msg())
}

func ResponseErrorWithMsg(c *gin.Context, code ResCode, msg interface{}) {
	c.JSON(http.StatusOK, &Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

func ResponseSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, &Response{
		Code: CodeSuccess,
		Msg:  CodeSuccess.Msg(),
		Data: data,
	})
}
