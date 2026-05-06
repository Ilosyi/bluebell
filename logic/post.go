package logic

import (
	"bluebell/dao/mysql"
	"bluebell/dao/redis"
	"bluebell/models"
	"bluebell/pkg/snowflake"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"go.uber.org/zap"
)

var (
	genPostID                  = snowflake.GenID
	createPostInMySQL          = mysql.CreatePost
	savePostTimeAndScore       = redis.SavePostTimeAndScore
	// 详情页走“聚合查询 + 单独补票数”的链路，减少多次查询。
	getPostBundleByIDFromMySQL = mysql.GetPostBundleByID
	// 旧版列表逻辑仍保留，因此旧依赖也暂时保留，方便平滑迁移。
	getUserByIDFromMySQL       = mysql.GetUserById
	getCommunityDetailByID     = mysql.GetCommunityDetailByID
	getPostIDsInOrder          = redis.GetPostIDsInOrder
	getCommunityPostIDsInOrder = redis.GetCommunityPostIDsInOrder
	// 列表页改成批量聚合查询，避免逐条查作者/社区。
	getPostBundlesByIDsFromMySQL = mysql.GetPostBundlesByIDs
	getPostVoteData            = redis.GetPostVoteData
	getPostListFromMySQL       = mysql.GetPostList
	countPostsFromMySQL        = mysql.CountPosts
	countPostsInCommunity      = redis.CountPostsInCommunity
)

func CreatePost(p *models.Post) error {
	//1.生成post_id,使用雪花算法
	p.ID = genPostID()
	//2.保存到数据库
	if err := createPostInMySQL(p); err != nil {
		return err
	}
	//3.将帖子创建时间和分数写入redis，并加入社区SET
	return savePostTimeAndScore(p.ID, p.CommunityID, time.Now())
}

func GetPostById(pid int64) (data *models.ApiPostDetail, err error) {

	// 第一步：一次性取出帖子主体、作者和社区信息。
	data, err = getPostBundleByIDFromMySQL(pid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrPostNotFound
			return
		}
		zap.L().Error("mysql.GetPostById(pid) failed", zap.Int64("pid", pid), zap.Error(err))
		return
	}
	// 第二步：票数仍由 Redis 维护，因此单独补一条当前帖子的投票统计。
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

// GetPostListNew 合并版帖子列表，根据CommunityID决定走全局榜单还是社区榜单
func GetPostListNew(p *models.ParamPostList) (data *models.ApiPostList, err error) {
	//根据CommunityID决定走全局榜单还是社区榜单
	if p.CommunityID == 0 {
		//全局榜单
		return GetPostList2(p)
	}
	//社区榜单
	return GetCommunityPostList(p)
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
		//根据帖子id查询作者信息
		user, err := getUserByIDFromMySQL(post.AuthorID)
		if err != nil {
			zap.L().Error("mysql.GetUserById(post.AuthorID) failed",
				zap.Int64("author_id", post.AuthorID),
				zap.Error(err))
			continue //查询作者失败，跳过该帖子
		}
		//根据社区id查询社区详情
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
