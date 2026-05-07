package models

import "time"

const (
	// PostStatusDraft 表示草稿。
	// 草稿只存在于数据库和“我的草稿箱”，不会进入公开帖子列表。
	PostStatusDraft int32 = 0
	// PostStatusPublished 表示已发布。
	// 只有这个状态的帖子才会进入公开列表、详情页和搜索结果。
	PostStatusPublished int32 = 1
)

// Post 对应 post 表中的主体数据。
// 它是项目里最核心的数据模型之一：
// 1. 发帖时 controller 会把请求体绑定到这个结构体。
// 2. dao/mysql 查询 post 表时也会映射到这个结构体。
// 3. logic 层会基于它继续拼接作者名、社区名、投票数等额外信息。
type Post struct {
	// ID 使用雪花算法生成，不是数据库自增主键。
	// json:",string" 的作用：把 int64 按字符串返回，避免前端 JavaScript 精度丢失。
	ID int64 `json:"id,string" db:"post_id"` // 帖子ID（雪花算法生成）
	// AuthorID 表示发帖人。
	AuthorID int64 `json:"author_id,string" db:"author_id"` // 作者用户ID
	// CommunityID 表示帖子属于哪个社区。
	CommunityID int64 `json:"community_id" db:"community_id" binding:"required"` // 所属社区ID
	// Status 表示帖子当前状态：草稿/已发布。
	Status int32 `json:"status" db:"status"` // 帖子状态
	// Title 是帖子标题。
	Title string `json:"title" db:"title" binding:"required"` // 帖子标题
	// Content 是帖子正文。
	Content string `json:"content" db:"content" binding:"required"` // 帖子内容
	// CreateTime 是数据库记录的创建时间。
	CreateTime time.Time `json:"create_time" db:"create_time"` // 创建时间
}

// ApiPostDetail 是给前端返回的“完整帖子详情”。
// 它不是单纯的 post 表数据，而是“帖子主体 + 作者名 + 社区详情 + 票数”的聚合结果。
type ApiPostDetail struct {
	AuthorName       string             `json:"author_name"` // 作者用户名
	VoteNum          int64              `json:"vote_num"`    // 赞成票数
	*Post            `json:"post"`      // 帖子详情
	*CommunityDetail `json:"community"` // 所属社区信息
}

// Pagination 统一描述分页元信息。
// 前端分页组件会根据这些字段决定是否显示下一页、总页数等。
type Pagination struct {
	Page       int64 `json:"page"`
	Size       int64 `json:"size"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
	HasMore    bool  `json:"has_more"`
}

// ApiPostList 统一帖子列表响应结构：
// Items 放当前页数据，Pagination 放分页元信息。
type ApiPostList struct {
	Items      []*ApiPostDetail `json:"items"`
	Pagination Pagination       `json:"pagination"`
}
