package models

// SignUpParam 对应注册接口的请求体。
// binding tag 会在 controller.ShouldBindJSON 时自动触发校验。
type SignUpParam struct {
	Username   string `json:"username" binding:"required,min=2,max=20"`        // 用户名，2-20个字符
	Password   string `json:"password" binding:"required,min=6,max=20"`        // 密码，6-20个字符
	RePassword string `json:"re_password" binding:"required,eqfield=Password"` // 确认密码，需与密码一致
}

// LoginParam 对应登录接口的请求体。
// Username 字段虽然叫 Username，但实际允许传“账号或昵称”。
type LoginParam struct {
	Username string `json:"username" binding:"required,min=2,max=32"` // 用户名或昵称
	Password string `json:"password" binding:"required,min=6,max=20"` // 密码
}

// ParamPostList 对应公开帖子列表接口 `/posts2` 的查询参数。
// Query string 会绑定到这个结构体，例如：
// /posts2?page=1&size=10&order=time&community_id=2&keyword=gin
type ParamPostList struct {
	CommunityID int64  `json:"community_id" form:"community_id"`                              // 社区ID（0或不传表示全局）
	Page        int64  `json:"page" form:"page" binding:"required,gte=1" example:"1"`         // 页码，从1开始
	Size        int64  `json:"size" form:"size" binding:"required,gte=1" example:"10"`        // 每页数量
	Order       string `json:"order" form:"order" binding:"oneof=time score" example:"score"` // 排序方式：time 或 score
	Keyword     string `json:"keyword" form:"keyword" binding:"omitempty,max=64"`             // 搜索关键词
}

// ParamMyPostList 对应“我的帖子/草稿箱”列表接口。
type ParamMyPostList struct {
	Page   int64 `json:"page" form:"page" binding:"required,gte=1" example:"1"`
	Size   int64 `json:"size" form:"size" binding:"required,gte=1" example:"10"`
	Status int32 `json:"status" form:"status" binding:"oneof=0 1" example:"1"` // 0=草稿，1=已发布
}

// ParamPostEdit 对应“编辑已发布帖子”的请求体。
// 它要求字段完整，因为编辑已发布帖子后仍然必须保持内容完整。
type ParamPostEdit struct {
	CommunityID int64  `json:"community_id" binding:"required"`
	Title       string `json:"title" binding:"required,max=128"`
	Content     string `json:"content" binding:"required,max=8192"`
}

// ParamPostDraft 对应“保存草稿”的请求体。
// 标题和内容可以为空，因为草稿允许暂存未完成内容。
type ParamPostDraft struct {
	CommunityID int64  `json:"community_id" binding:"required,gte=1"`
	Title       string `json:"title" binding:"omitempty,max=128"`
	Content     string `json:"content" binding:"omitempty,max=8192"`
}

// ParamVoteData 对应投票接口的请求体。
// direction 通过 `,string` 告诉 Gin：即使前端传的是字符串，也要转换成 int8。
type ParamVoteData struct {
	PostId    string `json:"post_id" binding:"required"`              // 帖子ID
	Direction int8   `json:"direction,string" binding:"oneof=1 -1 0"` // 投票方向：1=赞成，-1=反对，0=取消
}
