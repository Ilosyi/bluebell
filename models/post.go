package models

import "time"

// Post 帖子数据模型
type Post struct {
	ID          int64     `json:"id,string" db:"post_id"`           // 帖子ID（雪花算法生成）
	AuthorID    int64     `json:"author_id,string" db:"author_id"` // 作者用户ID
	CommunityID int64     `json:"community_id" db:"community_id" binding:"required"` // 所属社区ID
	Status      int32     `json:"status" db:"status"`               // 帖子状态
	Title       string    `json:"title" db:"title" binding:"required"`   // 帖子标题
	Content     string    `json:"content" db:"content" binding:"required"` // 帖子内容
	CreateTime  time.Time `json:"create_time" db:"create_time"`     // 创建时间
}

// ApiPostDetail 帖子详情接口响应结构体
type ApiPostDetail struct {
	AuthorName       string             `json:"author_name"` // 作者用户名
	VoteNum          int64              `json:"vote_num"`    // 赞成票数
	*Post            `json:"post"`      // 帖子详情
	*CommunityDetail `json:"community"` // 所属社区信息
}
