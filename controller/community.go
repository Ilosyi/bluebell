package controller

import (
	"bluebell/logic"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func CommunityHandler(c *gin.Context) {
	//查询所有社区（id、name），以列表（切片）形式返回
	data, err := logic.GetCommunityList()
	if err != nil {
		zap.L().Error("logic.GetCommunityList failed", zap.Error(err))
		//不轻易把服务端报错暴露给外面
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}

func CommunityDetailHandler(c *gin.Context) {
	//获取id
	strId := c.Param("id") //获取url参数

	id, err := strconv.ParseInt(strId, 10, 64) //格式转换
	if err != nil {
		//zap.L().Error("strconv.ParseInt failed", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	//根据id获取详情
	data, err := logic.GetCommunityDetail(id)
	if err != nil {
		zap.L().Error("logic.GetCommunityDetail failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}
