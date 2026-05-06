package controller

import (
	"bluebell/logic"
	"bluebell/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CreatePostHandler 创建帖子
// @Summary 创建帖子
// @Description 创建帖子接口，需要登录
// @Tags 帖子模块
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param object body models.Post true "帖子参数（title、content、community_id）"
// @Success 200 {object} Response "成功"
// @Failure 200 {object} Response "参数错误或未登录"
// @Router /post [post]
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
		return
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
// @Summary 获取帖子详情
// @Description 根据帖子ID获取帖子详情（含作者、社区信息、投票数）
// @Tags 帖子模块
// @Produce json
// @Param id path int64 true "帖子ID"
// @Success 200 {object} Response{data=models.ApiPostDetail} "成功"
// @Failure 200 {object} Response "参数错误或服务繁忙"
// @Router /post/{id} [get]
func GetPostDetailHandler(c *gin.Context) {
	//1.获取参数（帖子id）
	pidStr := c.Param("id")
	//字符串转换成数字
	pid, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil {
		//参数有问题
		zap.L().Error("get post detail with invalid param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
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

// GetPostListHandler2 按时间或分数排序的帖子列表
// @Summary 帖子列表（支持排序和社区筛选）
// @Description 按时间或分数排序获取帖子列表，支持社区ID筛选
// @Tags 帖子模块
// @Produce json
// @Param page query int64 true "页码" default(1)
// @Param size query int64 true "每页数量" default(10)
// @Param order query string true "排序方式: time 或 score" default(time) Enums(time, score)
// @Param community_id query int64 false "社区ID（0或不传表示全局）"
// @Success 200 {object} Response{data=[]models.ApiPostDetail} "成功"
// @Failure 200 {object} Response "参数错误或服务繁忙"
// @Router /posts2 [get]
func GetPostListHandler2(c *gin.Context) {
	//解析query参数到ParamPostList
	p := &models.ParamPostList{
		Page:  1,
		Size:  10,
		Order: "time",
	}
	if err := c.ShouldBindQuery(p); err != nil {
		zap.L().Error("GetPostListHandler2 with invalid param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	//调用logic层获取数据
	data, err := logic.GetPostListNew(p)
	if err != nil {
		zap.L().Error("logic.GetPostListNew failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}

// GetPostListHandler 获取帖子列表（旧版，按创建时间倒序）
// @Summary 帖子列表（按时间倒序）
// @Description 分页获取帖子列表，按创建时间倒序
// @Tags 帖子模块
// @Produce json
// @Param page query int64 true "页码" default(1)
// @Param size query int64 true "每页数量" default(10)
// @Success 200 {object} Response{data=[]models.ApiPostDetail} "成功"
// @Failure 200 {object} Response "服务繁忙"
// @Router /posts [get]
func GetPostListHandler(c *gin.Context) {
	//获取分页参数
	page, size := getPageInfo(c)
	//获取数据
	data, err := logic.GetPostList(page, size)
	if err != nil {
		zap.L().Error("logic.GetPostList failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	//返回响应
	ResponseSuccess(c, data)
}
