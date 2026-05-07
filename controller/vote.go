package controller

import (
	"bluebell/logic"
	"bluebell/models"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// voteForPost 指向 logic 层投票函数。
// 做成变量是为了测试时可以替换掉真实实现。
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
	// 第一步：把请求 JSON 绑定到 ParamVoteData。
	p := new(models.ParamVoteData)
	if err := c.ShouldBindJSON(p); err != nil {
		// 第二步：尝试把错误识别为 validator 字段校验错误。
		// 如果不是 validator 错误，就直接返回通用参数错误。
		var errs validator.ValidationErrors
		ok := errors.As(err, &errs)
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}
		// Translate(trans) 把规则错误翻译成中文。
		// removeTopStruct 去掉结构体名前缀，让前端拿到更干净的字段名。
		errData := removeTopStruct(errs.Translate(trans))
		ResponseErrorWithMsg(c, CodeInvalidParam, errData)
		return
	}

	// 第三步：从 JWT 中间件写入的上下文里取出当前登录用户 ID。
	userID, err := getCurrentUserID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}

	// 第四步：进入 logic 层执行投票。
	// logic 里会继续调用 Redis，并处理重复投票、投票过期等规则。
	if err := voteForPost(userID, p); err != nil {
		ResponseErrorWithMsg(c, CodeServerBusy, err.Error())
		return
	}

	// 第五步：返回成功响应。
	ResponseSuccess(c, nil)
}
