package controller

import (
	"bluebell/logic"
	"bluebell/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func CreatePostHandler(c *gin.Context) {
	// 获取参数和参数校验
	//c.ShouldBindJSON
	p := new(models.Post)
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Debug("c.ShouldBindJSON failed", zap.Any("err", err))
		zap.L().Error("CreatePost with invalid param")
		ResponseError(c, CodeInvalidParam)
		return
	}
	//从上下文中获取用户id
	userID, err := getCurrentUserID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
	}
	p.AuthorID = userID
	//2.创建帖子
	if err := logic.CreatePost(p); err != nil {
		zap.L().Error("logic.CreatePost failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	//3,返回响应
	ResponseSuccess(c, nil)
}

// GetPostDetailHandler 获取帖子详情的处理函数
func GetPostDetailHandler(c *gin.Context) {
	//1.获取参数（帖子id）
	pidStr := c.Param("id")
	//字符串转换成数字
	pid, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil {
		//参数有问题
		zap.L().Error("get post detail with invalid param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
	}
	//2.根据id取出帖子数据
	data, err := logic.GetPostById(pid)
	if err != nil {
		zap.L().Error("logic.GetPostById(pid) failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	//3.返回响应
	ResponseSuccess(c, data)
}
