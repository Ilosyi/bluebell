package controller

import (
	"bluebell/logic"
	"bluebell/models"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var voteForPost = logic.VoteForPost

// PostVoteHandler 帖子投票
// @Summary 帖子投票
// @Description 对帖子进行投票（赞成=1，反对=-1，取消=0），需要登录
// @Tags 投票模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param object body models.ParamVoteData true "投票参数"
// @Success 200 {object} Response{data=models.VoteData} "成功"
// @Failure 200 {object} Response "参数错误、未登录、重复投票或超出投票时间"
// @Router /vote [post]
func PostVoteHandler(c *gin.Context) {
	//参数校验
	p := new(models.ParamVoteData)
	if err := c.ShouldBindJSON(p); err != nil {
		//validator断言
		var errs validator.ValidationErrors
		ok := errors.As(err, &errs)
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}
		errData := removeTopStruct(errs.Translate(trans))  //翻译并去除错误提示中的结构体标识
		ResponseErrorWithMsg(c, CodeInvalidParam, errData) //翻译并返回响应
		return
	}
	//获取用户id
	userID, err := getCurrentUserID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	if err := voteForPost(userID, p); err != nil {
		ResponseErrorWithMsg(c, CodeServerBusy, err.Error())
		return
	}
	ResponseSuccess(c, nil)
}
