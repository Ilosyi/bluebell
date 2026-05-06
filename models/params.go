package models

// SignUpParam 用户注册请求参数
type SignUpParam struct {
	Username   string `json:"username" binding:"required,min=2,max=20"`     // 用户名，2-20个字符
	Password   string `json:"password" binding:"required,min=6,max=20"`     // 密码，6-20个字符
	RePassword string `json:"re_password" binding:"required,eqfield=Password"` // 确认密码，需与密码一致
}

// LoginParam 用户登录请求参数
type LoginParam struct {
	Username string `json:"username" binding:"required,min=2,max=20"` // 用户名
	Password string `json:"password" binding:"required,min=6,max=20"` // 密码
}

// ParamPostList 帖子列表请求参数
type ParamPostList struct {
	CommunityID int64  `json:"community_id" form:"community_id"`                                  // 社区ID（0或不传表示全局）
	Page        int64  `json:"page" form:"page" binding:"required,gte=1" example:"1"`             // 页码，从1开始
	Size        int64  `json:"size" form:"size" binding:"required,gte=1" example:"10"`            // 每页数量
	Order       string `json:"order" form:"order" binding:"oneof=time score" example:"score"`     // 排序方式：time 或 score
}

// ParamVoteData 帖子投票请求参数
type ParamVoteData struct {
	PostId    string `json:"post_id" binding:"required"`              // 帖子ID
	Direction int8   `json:"direction,string" binding:"oneof=1 -1 0"` // 投票方向：1=赞成，-1=反对，0=取消
}
