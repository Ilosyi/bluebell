package models

import "time"

// User 对应用户表 user。
// 这个结构体同时承担两种角色：
// 1. dao 层查表/写表时的数据承载体。
// 2. controller 返回用户资料时的响应结构。
type User struct {
	// UserID 是业务用户 ID，不是数据库自增主键。
	UserID int64 `json:"user_id,string" db:"user_id"` // 用户ID
	// Username 是登录账号，注册后通常不再修改。
	Username string `json:"username" db:"username"` // 用户名
	// Password 存的是加密后的密码哈希。
	// json:"-" 表示返回给前端时绝不输出这个字段。
	Password string `json:"-" db:"password"` // 密码（加密存储）
	// Nickname 是用户对外展示的昵称。
	Nickname string `json:"nickname" db:"nickname"` // 昵称
	// AvatarURL 是头像地址。
	AvatarURL string `json:"avatar_url" db:"avatar_url"` // 头像地址
	// Bio 是个人简介。
	Bio string `json:"bio" db:"bio"` // 个人简介
	// CreateTime 是数据库里用户记录的创建时间。
	CreateTime time.Time `json:"create_time" db:"create_time"` // 创建时间
	// Token 只在登录成功时临时返回给前端，不会写入数据库。
	Token string `json:"token" db:"-"` // JWT Token（不入库）
}

// UpdateUserProfileParam 是“更新当前用户资料”接口的请求体。
// 注意：这里只允许改昵称、头像、简介，不允许改账号和密码。
type UpdateUserProfileParam struct {
	Nickname  string `json:"nickname" binding:"required,min=2,max=32"`
	AvatarURL string `json:"avatar_url" binding:"omitempty,max=512"`
	Bio       string `json:"bio" binding:"omitempty,max=160"`
}
