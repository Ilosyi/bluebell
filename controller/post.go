package controller

import (
	"bluebell/logic"
	"bluebell/models"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	// 这里同样使用“函数变量”而不是直接硬编码调用 logic 包，
	// 方便在单测中替换为假实现。
	createPost     = logic.CreatePost
	createDraft    = logic.CreateDraft
	getPostByID    = logic.GetPostById
	getPostList    = logic.GetPostList
	getPostListNew = logic.GetPostListNew
	getMyPostList  = logic.GetMyPostList
	getManagePost  = logic.GetManagePostByID
	updatePost     = logic.UpdatePost
	updateDraft    = logic.UpdateDraft
	publishDraft   = logic.PublishDraft
	deletePost     = logic.DeletePost
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
	// 第一步：读取请求体。
	// 这里复用了 models.Post 作为请求结构，前端只需要传 title/content/community_id。
	p := new(models.Post)
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Debug("c.ShouldBindJSON failed", zap.Any("err", err))
		zap.L().Error("CreatePost with invalid param")
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 第二步：从 JWT 中间件写入的 Context 中拿到当前登录用户 ID。
	userID, err := getCurrentUserID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}

	// 第三步：把当前用户设置成帖子作者。
	// 这里不信任前端传来的 author_id，而是以后端登录态为准。
	p.AuthorID = userID

	// 第四步：进入 logic 层创建帖子。
	if err := createPost(p); err != nil {
		zap.L().Error("logic.CreatePost failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	// 第五步：返回成功响应。
	ResponseSuccess(c, nil)
}

// CreateDraftHandler 创建草稿。
// 与 CreatePost 不同，草稿不会立即进入公开帖子流。
func CreateDraftHandler(c *gin.Context) {
	p := new(models.ParamPostDraft)
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Debug("CreateDraft with invalid param", zap.Any("err", err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	userID, err := getCurrentUserID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	data, err := createDraft(userID, p)
	if err != nil {
		zap.L().Error("logic.CreateDraft failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
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
	// 第二步：调用 logic 层。如果帖子不存在，返回明确的业务码，
	// 而不是把所有错误都包装成“服务繁忙”。
	data, err := getPostByID(pid)
	if err != nil {
		zap.L().Error("logic.GetPostById(pid) failed", zap.Error(err))
		if errors.Is(err, logic.ErrPostNotFound) {
			ResponseError(c, CodePostNotFound)
			return
		}
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
	// 先设置默认值，保证即使前端不传 page/size/order，也能得到稳定结果。
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

	// 进入 logic 层：
	// 1. 如果带 keyword，走 MySQL 搜索。
	// 2. 如果没带 keyword，走 Redis 排序榜单 + MySQL 聚合详情。
	data, err := getPostListNew(p)
	if err != nil {
		zap.L().Error("logic.GetPostListNew failed", zap.Error(err))
		if errors.Is(err, logic.ErrCommunityNotFound) {
			ResponseError(c, CodeCommunityNotFound)
			return
		}
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
	data, err := getPostList(page, size)
	if err != nil {
		zap.L().Error("logic.GetPostList failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	//返回响应
	ResponseSuccess(c, data)
}

func GetMyPostListHandler(c *gin.Context) {
	// 这个接口是“我的帖子/草稿箱”，因此必须先拿登录用户。
	userID, err := getCurrentUserID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	p := &models.ParamMyPostList{
		Page:   1,
		Size:   10,
		Status: models.PostStatusPublished,
	}
	// ShouldBindQuery 会把 URL 参数绑定到结构体，例如 ?page=2&status=0。
	if err := c.ShouldBindQuery(p); err != nil {
		zap.L().Error("GetMyPostListHandler with invalid param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	data, err := getMyPostList(userID, p)
	if err != nil {
		zap.L().Error("logic.GetMyPostList failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}

// GetManagePostHandler 获取“当前用户自己的某篇帖子”详情。
// 它和公开帖子详情的区别：
// 1. 可以查看草稿。
// 2. 要检查当前用户是否是作者。
func GetManagePostHandler(c *gin.Context) {
	userID, pid, ok := getUserAndPostID(c)
	if !ok {
		return
	}
	data, err := getManagePost(userID, pid)
	if err != nil {
		writePostManageError(c, err)
		return
	}
	ResponseSuccess(c, data)
}

// UpdatePostHandler 编辑已发布帖子。
func UpdatePostHandler(c *gin.Context) {
	userID, pid, ok := getUserAndPostID(c)
	if !ok {
		return
	}
	p := new(models.ParamPostEdit)
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Debug("UpdatePost with invalid param", zap.Any("err", err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	data, err := updatePost(userID, pid, p)
	if err != nil {
		writePostManageError(c, err)
		return
	}
	ResponseSuccess(c, data)
}

// UpdateDraftHandler 编辑草稿。
func UpdateDraftHandler(c *gin.Context) {
	userID, pid, ok := getUserAndPostID(c)
	if !ok {
		return
	}
	p := new(models.ParamPostDraft)
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Debug("UpdateDraft with invalid param", zap.Any("err", err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	data, err := updateDraft(userID, pid, p)
	if err != nil {
		writePostManageError(c, err)
		return
	}
	ResponseSuccess(c, data)
}

// PublishDraftHandler 把草稿发布出去。
func PublishDraftHandler(c *gin.Context) {
	userID, pid, ok := getUserAndPostID(c)
	if !ok {
		return
	}
	data, err := publishDraft(userID, pid)
	if err != nil {
		writePostManageError(c, err)
		return
	}
	ResponseSuccess(c, data)
}

// DeletePostHandler 删除帖子或草稿。
func DeletePostHandler(c *gin.Context) {
	userID, pid, ok := getUserAndPostID(c)
	if !ok {
		return
	}
	if err := deletePost(userID, pid); err != nil {
		writePostManageError(c, err)
		return
	}
	ResponseSuccess(c, nil)
}

// getUserAndPostID 是多个“帖子管理接口”共用的小工具函数。
// 它一次性完成两件事：
// 1. 从登录态拿当前用户 ID。
// 2. 从路由参数里解析帖子 ID。
func getUserAndPostID(c *gin.Context) (int64, int64, bool) {
	userID, err := getCurrentUserID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return 0, 0, false
	}
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ResponseError(c, CodeInvalidParam)
		return 0, 0, false
	}
	return userID, pid, true
}

// writePostManageError 把 logic 层返回的领域错误统一映射成前端可识别的业务码。
func writePostManageError(c *gin.Context, err error) {
	zap.L().Error("post manage operation failed", zap.Error(err))
	switch {
	case errors.Is(err, logic.ErrPostNotFound):
		ResponseError(c, CodePostNotFound)
	case errors.Is(err, logic.ErrForbidden):
		ResponseError(c, CodeForbidden)
	case errors.Is(err, logic.ErrPostNotReady):
		ResponseError(c, CodePostNotReady)
	default:
		ResponseError(c, CodeServerBusy)
	}
}
