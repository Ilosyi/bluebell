package controller

import (
	"bluebell/logic"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	getCommunityList   = logic.GetCommunityList
	getCommunityDetail = logic.GetCommunityDetail
)

// CommunityHandler 获取社区列表
// @Summary 获取所有社区
// @Description 查询所有社区（id、name），以列表形式返回
// @Tags 社区模块
// @Produce json
// @Success 200 {object} Response{data=[]models.Community} "成功"
// @Failure 200 {object} Response "服务繁忙"
// @Router /community [get]
func CommunityHandler(c *gin.Context) {
	//查询所有社区（id、name），以列表（切片）形式返回
	data, err := getCommunityList()
	if err != nil {
		zap.L().Error("logic.GetCommunityList failed", zap.Error(err))
		//不轻易把服务端报错暴露给外面
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}

// CommunityDetailHandler 获取社区详情
// @Summary 获取社区详情
// @Description 根据社区ID获取社区详情
// @Tags 社区模块
// @Produce json
// @Param id path int64 true "社区ID"
// @Success 200 {object} Response{data=models.CommunityDetail} "成功"
// @Failure 200 {object} Response "参数错误或服务繁忙"
// @Router /community/{id} [get]
func CommunityDetailHandler(c *gin.Context) {
	//获取id
	strId := c.Param("id") //获取url参数

	id, err := strconv.ParseInt(strId, 10, 64) //格式转换
	if err != nil {
		//zap.L().Error("strconv.ParseInt failed", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 根据 id 获取详情；如果 logic 明确告诉我们“社区不存在”，
	// 这里返回稳定的业务码，方便前端区分“404 资源不存在”和“服务异常”。
	data, err := getCommunityDetail(id)
	if err != nil {
		zap.L().Error("logic.GetCommunityDetail failed", zap.Error(err))
		if errors.Is(err, logic.ErrCommunityNotFound) {
			ResponseError(c, CodeCommunityNotFound)
			return
		}
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}
