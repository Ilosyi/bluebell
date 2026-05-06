package logic

import (
	"bluebell/dao/mysql"
	"bluebell/dao/redis"
	"bluebell/models"
	"bluebell/pkg/snowflake"
	"strconv"
	"time"

	"go.uber.org/zap"
)

var (
	genPostID                  = snowflake.GenID
	createPostInMySQL          = mysql.CreatePost
	savePostTimeAndScore       = redis.SavePostTimeAndScore
	getPostByIDFromMySQL       = mysql.GetPostById
	getUserByIDFromMySQL       = mysql.GetUserById
	getCommunityDetailByID     = mysql.GetCommunityDetailByID
	getPostIDsInOrder          = redis.GetPostIDsInOrder
	getCommunityPostIDsInOrder = redis.GetCommunityPostIDsInOrder
	getPostListByIDsFromMySQL  = mysql.GetPostListByIDs
	getPostVoteData            = redis.GetPostVoteData
	getPostListFromMySQL       = mysql.GetPostList
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

	//查询并组合我们想要的数据
	post, err := getPostByIDFromMySQL(pid)
	if err != nil {
		zap.L().Error("mysql.GetPostById(pid) failed", zap.Int64("pid", pid), zap.Error(err))
		return
	}
	//根据帖子id查询作者信息
	user, err := getUserByIDFromMySQL(post.AuthorID)
	if err != nil {
		zap.L().Error("mysql.GetUserById(post.AuthorID) failed",
			zap.Int64("author_id", post.AuthorID),
			zap.Error(err))
		return
	}
	//根据社区id查询社区详情
	community, err := getCommunityDetailByID(post.CommunityID)
	if err != nil {
		zap.L().Error("mysql.GetCommunityDetailByID(post.CommunityID) failed",
			zap.Int64("community_id", post.CommunityID),
			zap.Error(err))
		return
	}
	voteCounts, err := getPostVoteData([]string{strconv.FormatInt(pid, 10)})
	if err != nil {
		zap.L().Error("redis.GetPostVoteData failed", zap.Int64("pid", pid), zap.Error(err))
		return
	}

	var voteNum int64
	if len(voteCounts) > 0 {
		voteNum = voteCounts[0]
	}
	//拼装数据
	data = &models.ApiPostDetail{
		AuthorName:      user.Username,
		VoteNum:         voteNum,
		Post:            post,
		CommunityDetail: community,
	}
	return
}

// GetPostListNew 合并版帖子列表，根据CommunityID决定走全局榜单还是社区榜单
func GetPostListNew(p *models.ParamPostList) (data []*models.ApiPostDetail, err error) {
	//根据CommunityID决定走全局榜单还是社区榜单
	if p.CommunityID == 0 {
		//全局榜单
		return GetPostList2(p)
	}
	//社区榜单
	return GetCommunityPostList(p)
}

// GetPostList2 全局榜单：从Redis获取ID列表，从MySQL批量查询，从Redis获取投票数，拼装详情
func GetPostList2(p *models.ParamPostList) (data []*models.ApiPostDetail, err error) {
	//从Redis获取按时间或分数排序的帖子ID列表
	ids, err := getPostIDsInOrder(p)
	if err != nil {
		zap.L().Error("redis.GetPostIDsInOrder failed", zap.Error(err))
		return
	}
	return getPostListByIDs(ids)
}

// GetCommunityPostList 社区榜单：与全局类似，但ID来源为社区帖子集合
func GetCommunityPostList(p *models.ParamPostList) (data []*models.ApiPostDetail, err error) {
	//从Redis获取社区帖子ID列表（使用ZInterStore交集）
	ids, err := getCommunityPostIDsInOrder(p)
	if err != nil {
		zap.L().Error("redis.GetCommunityPostIDsInOrder failed", zap.Error(err))
		return
	}
	return getPostListByIDs(ids)
}

// getPostListByIDs 根据帖子ID列表查询详情并拼装（公共逻辑）
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
	//根据ID列表从MySQL批量查询帖子主体
	posts, err := getPostListByIDsFromMySQL(idInt64s)
	if err != nil {
		zap.L().Error("mysql.GetPostListByIDs failed", zap.Error(err))
		return
	}
	//预先从Redis查询每篇帖子的赞成票数
	voteCounts, err := getPostVoteData(ids)
	if err != nil {
		zap.L().Error("redis.GetPostVoteData failed", zap.Error(err))
		return
	}
	//依次拼装ApiPostDetail（作者名、社区信息、VoteNum、帖子内容）
	data = make([]*models.ApiPostDetail, 0, len(posts))
	for idx, post := range posts {
		user, err := getUserByIDFromMySQL(post.AuthorID)
		if err != nil {
			zap.L().Error("mysql.GetUserById failed", zap.Int64("author_id", post.AuthorID), zap.Error(err))
			continue
		}
		community, err := getCommunityDetailByID(post.CommunityID)
		if err != nil {
			zap.L().Error("mysql.GetCommunityDetailByID failed", zap.Int64("community_id", post.CommunityID), zap.Error(err))
			continue
		}
		data = append(data, &models.ApiPostDetail{
			AuthorName:      user.Username,
			VoteNum:         voteCounts[idx],
			Post:            post,
			CommunityDetail: community,
		})
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
