package controller

import (
	"bluebell/logic"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	// 通过包级变量接 logic 函数，便于测试时替换为 mock/stub。
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
	// 调用 logic 层获取所有社区。
	// data 的类型是 []*models.Community，也就是“社区切片”。
	data, err := getCommunityList()
	if err != nil {
		zap.L().Error("logic.GetCommunityList failed", zap.Error(err))
		// 不把底层数据库错误直接暴露给前端，统一返回服务繁忙。
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
	// 先从路由参数中取出 id，例如 /community/3 中的 3。
	strId := c.Param("id")

	// 把字符串 id 转成 int64。
	// ParseInt 的第二个参数 10 表示按十进制解析。
	id, err := strconv.ParseInt(strId, 10, 64)
	if err != nil {
		// 只要 URL 里的 id 不是合法数字，就直接判定为参数错误。
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
