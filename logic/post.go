package logic

import (
	"bluebell/dao/mysql"
	"bluebell/models"
	"bluebell/pkg/snowflake"

	"go.uber.org/zap"
)

func CreatePost(p *models.Post) error {
	//1.生成post_id,使用雪花算法
	p.ID = int64(snowflake.GenID())
	//2.保存到数据库
	return mysql.CreatePost(p)
}

func GetPostById(pid int64) (data *models.ApiPostDetail, err error) {

	//查询并组合我们想要的数据
	post, err := mysql.GetPostById(pid)
	if err != nil {
		zap.L().Error("mysql.GetPostById(pid) failed", zap.Int64("pid", pid), zap.Error(err))
		return
	}
	//根据帖子id查询作者信息
	user, err := mysql.GetUserById(post.AuthorID)
	if err != nil {
		zap.L().Error("mysql.GetUserById(post.AuthorID) failed",
			zap.Int64("author_id", post.AuthorID),
			zap.Error(err))
		return
	}
	//根据社区id查询社区详情
	community, err := mysql.GetCommunityDetailByID(post.CommunityID)
	if err != nil {
		zap.L().Error("mysql.GetCommunityDetailByID(post.CommunityID) failed",
			zap.Int64("community_id", post.CommunityID),
			zap.Error(err))
		return
	}
	//拼装数据
	data = &models.ApiPostDetail{
		AuthorName:      user.Username,
		Post:            post,
		CommunityDetail: community,
	}
	return
}
