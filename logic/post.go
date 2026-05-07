package logic

import (
	"bluebell/dao/mysql"
	"bluebell/dao/redis"
	"bluebell/models"
	"bluebell/pkg/snowflake"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

var (
	// 这些变量默认指向 dao/pkg 层真实实现。
	// 单测里会把它们替换成假函数，从而专注验证业务编排逻辑。
	genPostID              = snowflake.GenID
	createPostInMySQL      = mysql.CreatePost
	savePostTimeAndScore   = redis.SavePostTimeAndScore
	deletePostIndex        = redis.DeletePostIndex
	movePostCommunityIndex = redis.MovePostCommunity
	// 详情页走“聚合查询 + 单独补票数”的链路，减少多次查询。
	getPostBundleByIDFromMySQL = mysql.GetPostBundleByID
	// 旧版列表逻辑仍保留，因此旧依赖也暂时保留，方便平滑迁移。
	getUserByIDFromMySQL       = mysql.GetUserById
	getCommunityDetailByID     = mysql.GetCommunityDetailByID
	getPostIDsInOrder          = redis.GetPostIDsInOrder
	getCommunityPostIDsInOrder = redis.GetCommunityPostIDsInOrder
	// 列表页改成批量聚合查询，避免逐条查作者/社区。
	getPostBundlesByIDsFromMySQL = mysql.GetPostBundlesByIDs
	getPostVoteData              = redis.GetPostVoteData
	getPostListFromMySQL         = mysql.GetPostList
	countPostsFromMySQL          = mysql.CountPosts
	searchPostBundlesFromMySQL   = mysql.SearchPostBundles
	countSearchPostsFromMySQL    = mysql.CountSearchPosts
	countPostsInCommunity        = redis.CountPostsInCommunity
	getPostForManageByID         = mysql.GetPostForManageByID
	getMyPostBundles             = mysql.GetMyPostBundles
	countMyPosts                 = mysql.CountMyPosts
	updatePostInMySQL            = mysql.UpdatePost
	publishPostInMySQL           = mysql.PublishPost
	deletePostInMySQL            = mysql.DeletePost
)

// CreatePost 创建一篇“已发布帖子”。
// 流程非常典型：
// 1. 生成业务帖子 ID
// 2. 设为“已发布”状态
// 3. 写入 MySQL
// 4. 把时间线/热度/社区索引写入 Redis
func CreatePost(p *models.Post) error {
	// 生成全局唯一的帖子 ID。
	p.ID = genPostID()
	// 新帖子默认直接发布，而不是草稿。
	p.Status = models.PostStatusPublished
	// 先写 MySQL，确保帖子主体数据持久化成功。
	if err := createPostInMySQL(p); err != nil {
		return err
	}
	// 再写 Redis，用于后续时间排序、热度排序、社区内排序。
	return savePostTimeAndScore(p.ID, p.CommunityID, time.Now())
}

// CreateDraft 创建草稿。
// 草稿和已发布帖子的区别主要是 Status=Draft，且不会立刻进入 Redis 公共索引。
func CreateDraft(userID int64, p *models.ParamPostDraft) (*models.ApiPostDetail, error) {
	post := &models.Post{
		ID:          genPostID(),
		AuthorID:    userID,
		CommunityID: p.CommunityID,
		Status:      models.PostStatusDraft,
		Title:       p.Title,
		Content:     p.Content,
	}
	if err := createPostInMySQL(post); err != nil {
		return nil, err
	}
	return getPostForManageByID(post.ID)
}

// GetPostById 获取公开帖子详情。
// 这里的“公开”很重要：草稿不会通过这个函数暴露出去。
func GetPostById(pid int64) (data *models.ApiPostDetail, err error) {
	// 先从 MySQL 一次性取出帖子、作者、社区三部分信息。
	data, err = getPostBundleByIDFromMySQL(pid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrPostNotFound
			return
		}
		zap.L().Error("mysql.GetPostById(pid) failed", zap.Int64("pid", pid), zap.Error(err))
		return
	}
	// 投票数最新值保存在 Redis，因此需要单独补一条票数查询。
	voteCounts, err := getPostVoteData([]string{strconv.FormatInt(pid, 10)})
	if err != nil {
		zap.L().Error("redis.GetPostVoteData failed", zap.Int64("pid", pid), zap.Error(err))
		return
	}

	var voteNum int64
	if len(voteCounts) > 0 {
		voteNum = voteCounts[0]
	}
	data.VoteNum = voteNum
	return
}

// GetManagePostByID 获取“当前用户自己管理的帖子”。
// 相比公开详情，它允许查看草稿，但必须验证作者身份。
func GetManagePostByID(userID, pid int64) (*models.ApiPostDetail, error) {
	post, err := getPostForManageByID(pid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}
	if post.AuthorID != userID {
		return nil, ErrForbidden
	}
	voteCounts, err := getPostVoteData([]string{strconv.FormatInt(pid, 10)})
	if err != nil {
		return nil, err
	}
	if len(voteCounts) > 0 {
		post.VoteNum = voteCounts[0]
	}
	return post, nil
}

// GetMyPostList 获取当前用户自己的帖子列表或草稿列表。
// status=1 表示“我的帖子”，status=0 表示“草稿箱”。
func GetMyPostList(userID int64, p *models.ParamMyPostList) (*models.ApiPostList, error) {
	items, err := getMyPostBundles(userID, p.Status, p.Page, p.Size)
	if err != nil {
		return nil, err
	}
	// 已发布帖子才需要补票数；草稿没有公开投票。
	if p.Status == models.PostStatusPublished {
		ids := make([]string, 0, len(items))
		for _, item := range items {
			ids = append(ids, strconv.FormatInt(item.Post.ID, 10))
		}
		votes, err := getPostVoteData(ids)
		if err != nil {
			return nil, err
		}
		for i, vote := range votes {
			if i < len(items) {
				items[i].VoteNum = vote
			}
		}
	}
	total, err := countMyPosts(userID, p.Status)
	if err != nil {
		return nil, err
	}
	return buildPostList(items, p.Page, p.Size, total), nil
}

// UpdatePost 编辑已发布帖子。
// 如果社区发生变化，还要同步 Redis 里的社区索引。
func UpdatePost(userID, pid int64, p *models.ParamPostEdit) (*models.ApiPostDetail, error) {
	// 先确认帖子存在且属于当前用户。
	old, err := GetManagePostByID(userID, pid)
	if err != nil {
		return nil, err
	}

	// 只把允许修改的字段组装出来。
	next := &models.Post{
		CommunityID: p.CommunityID,
		Title:       p.Title,
		Content:     p.Content,
	}
	// 更新 MySQL 主体数据。
	if err := updatePostInMySQL(pid, next); err != nil {
		return nil, err
	}
	// 如果已发布帖子换了社区，还要调整 Redis 的社区归属集合。
	if old.Status == models.PostStatusPublished && old.CommunityID != p.CommunityID {
		if err := movePostCommunityIndex(pid, old.CommunityID, p.CommunityID); err != nil {
			return nil, err
		}
	}
	return GetManagePostByID(userID, pid)
}

// UpdateDraft 编辑草稿。
// 和 UpdatePost 的区别是：这里只允许处理草稿状态。
func UpdateDraft(userID, pid int64, p *models.ParamPostDraft) (*models.ApiPostDetail, error) {
	old, err := GetManagePostByID(userID, pid)
	if err != nil {
		return nil, err
	}
	if old.Status != models.PostStatusDraft {
		return nil, ErrForbidden
	}
	next := &models.Post{
		CommunityID: p.CommunityID,
		Title:       p.Title,
		Content:     p.Content,
	}
	if err := updatePostInMySQL(pid, next); err != nil {
		return nil, err
	}
	return GetManagePostByID(userID, pid)
}

// PublishDraft 发布草稿。
// 发布时除了改数据库状态，还要把帖子写入 Redis 排序索引。
func PublishDraft(userID, pid int64) (*models.ApiPostDetail, error) {
	post, err := GetManagePostByID(userID, pid)
	if err != nil {
		return nil, err
	}
	if post.Status != models.PostStatusDraft {
		return nil, ErrForbidden
	}
	// 草稿要发布，至少要有社区、标题和正文。
	if post.CommunityID <= 0 || post.Title == "" || post.Content == "" {
		return nil, ErrPostNotReady
	}
	// 先把数据库状态改成已发布。
	if err := publishPostInMySQL(pid); err != nil {
		return nil, err
	}
	// 再把公开排序相关索引写入 Redis。
	if err := savePostTimeAndScore(pid, post.CommunityID, time.Now()); err != nil {
		return nil, err
	}
	return GetManagePostByID(userID, pid)
}

// DeletePost 删除帖子。
// 如果删的是已发布帖子，还要删除 Redis 里的各种公开索引。
func DeletePost(userID, pid int64) error {
	post, err := GetManagePostByID(userID, pid)
	if err != nil {
		return err
	}
	if err := deletePostInMySQL(pid); err != nil {
		return err
	}
	if post.Status == models.PostStatusPublished {
		return deletePostIndex(pid, post.CommunityID)
	}
	return nil
}

// GetPostListNew 合并版帖子列表，根据CommunityID决定走全局榜单还是社区榜单
func GetPostListNew(p *models.ParamPostList) (data *models.ApiPostList, err error) {
	// 搜索是一个独立逻辑入口，优先级高于社区榜单/全局榜单。
	p.Keyword = strings.TrimSpace(p.Keyword)
	if p.Keyword != "" {
		return SearchPostList(p)
	}
	//根据CommunityID决定走全局榜单还是社区榜单
	if p.CommunityID == 0 {
		//全局榜单
		return GetPostList2(p)
	}
	//社区榜单
	return GetCommunityPostList(p)
}

// SearchPostList 走 MySQL 搜索公开帖子。
// 因为搜索条件涉及标题、正文、作者、社区名，单纯依赖 Redis 排序结构不合适。
func SearchPostList(p *models.ParamPostList) (*models.ApiPostList, error) {
	// 先查当前页帖子主体数据。
	items, err := searchPostBundlesFromMySQL(p)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, strconv.FormatInt(item.Post.ID, 10))
	}
	// 再补 Redis 里的最新投票数。
	votes, err := getPostVoteData(ids)
	if err != nil {
		return nil, err
	}
	for idx, vote := range votes {
		if idx < len(items) {
			items[idx].VoteNum = vote
		}
	}
	total, err := countSearchPostsFromMySQL(p)
	if err != nil {
		return nil, err
	}
	return buildPostList(items, p.Page, p.Size, total), nil
}

// GetPostList2 全局榜单：从Redis获取ID列表，从MySQL批量查询，从Redis获取投票数，拼装详情
func GetPostList2(p *models.ParamPostList) (data *models.ApiPostList, err error) {
	//从Redis获取按时间或分数排序的帖子ID列表
	ids, err := getPostIDsInOrder(p)
	if err != nil {
		zap.L().Error("redis.GetPostIDsInOrder failed", zap.Error(err))
		return
	}
	// 再用批量聚合查询把当前页帖子渲染所需的字段一次性查出来。
	items, err := getPostListByIDs(ids)
	if err != nil {
		return nil, err
	}
	// 全局榜单总数来自 MySQL，便于返回稳定分页元数据。
	total, err := countPostsFromMySQL()
	if err != nil {
		return nil, err
	}
	return buildPostList(items, p.Page, p.Size, total), nil
}

// GetCommunityPostList 社区榜单：与全局类似，但ID来源为社区帖子集合
func GetCommunityPostList(p *models.ParamPostList) (data *models.ApiPostList, err error) {
	//从Redis获取社区帖子ID列表（使用ZInterStore交集）
	ids, err := getCommunityPostIDsInOrder(p)
	if err != nil {
		zap.L().Error("redis.GetCommunityPostIDsInOrder failed", zap.Error(err))
		return
	}
	items, err := getPostListByIDs(ids)
	if err != nil {
		return nil, err
	}
	// 社区榜单总数来自 Redis community set 的基数。
	total, err := countPostsInCommunity(p.CommunityID)
	if err != nil {
		return nil, err
	}
	return buildPostList(items, p.Page, p.Size, total), nil
}

// getPostListByIDs 根据帖子 ID 列表批量查询详情，并把 Redis 中的票数补回去。
func getPostListByIDs(ids []string) (data []*models.ApiPostDetail, err error) {
	if len(ids) == 0 {
		return
	}
	//将string类型的ID转为int64
	idInt64s := make([]int64, 0, len(ids))
	for _, idStr := range ids {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			zap.L().Error("strconv.ParseInt failed", zap.String("id", idStr), zap.Error(err))
			continue
		}
		idInt64s = append(idInt64s, id)
	}
	// 这里不再只查 post 表，而是直接查聚合后的详情结构。
	posts, err := getPostBundlesByIDsFromMySQL(idInt64s)
	if err != nil {
		zap.L().Error("mysql.GetPostBundlesByIDs failed", zap.Error(err))
		return
	}
	// Redis 中的投票统计仍然是最新值，因此列表页票数以 Redis 为准。
	voteCounts, err := getPostVoteData(ids)
	if err != nil {
		zap.L().Error("redis.GetPostVoteData failed", zap.Error(err))
		return
	}
	// MySQL 返回的帖子顺序已经和 Redis 一致，这里只需要顺序补票数。
	data = make([]*models.ApiPostDetail, 0, len(posts))
	for idx, post := range posts {
		if idx < len(voteCounts) {
			post.VoteNum = voteCounts[idx]
		}
		data = append(data, post)
	}
	return
}

func GetPostList(page, size int64) (data []*models.ApiPostDetail, err error) {
	posts, err := getPostListFromMySQL(page, size)
	data = make([]*models.ApiPostDetail, 0, len(posts))
	for _, post := range posts {
		// 旧版列表没有走聚合查询，所以这里会逐条查作者、逐条查社区。
		// 这也是典型的 N+1 查询问题来源。
		user, err := getUserByIDFromMySQL(post.AuthorID)
		if err != nil {
			zap.L().Error("mysql.GetUserById(post.AuthorID) failed",
				zap.Int64("author_id", post.AuthorID),
				zap.Error(err))
			continue //查询作者失败，跳过该帖子
		}
		community, err := getCommunityDetailByID(post.CommunityID)
		if err != nil {
			zap.L().Error("mysql.GetCommunityDetailByID(post.CommunityID) failed",
				zap.Int64("community_id", post.CommunityID),
				zap.Error(err))
			continue //查询社区失败，跳过该帖子
		}
		postsDetail := &models.ApiPostDetail{
			AuthorName:      user.Username,
			Post:            post,
			CommunityDetail: community,
		}
		data = append(data, postsDetail)
	}
	return
}

// buildPostList 把帖子项和分页元信息组装成统一响应结构。
func buildPostList(items []*models.ApiPostDetail, page, size, total int64) *models.ApiPostList {
	totalPages := int64(0)
	if total > 0 && size > 0 {
		totalPages = (total + size - 1) / size
	}

	// 统一封装分页元数据，前端分页栏就不需要“猜”下一页是否存在了。
	return &models.ApiPostList{
		Items: items,
		Pagination: models.Pagination{
			Page:       page,
			Size:       size,
			Total:      total,
			TotalPages: totalPages,
			HasMore:    page < totalPages,
		},
	}
}
