// Package models 放的是“数据结构定义”。
// 这些结构体会在多个层之间传递：
// 1. controller 层用来接收/返回 JSON。
// 2. dao 层用来接收数据库查询结果。
// 3. logic 层用来组织业务数据。
package models

import "time"

// Community 表示“社区列表页”里的简要社区信息。
// 这里只保留列表真正需要的字段：ID 和名称。
type Community struct {
	// json tag 决定返回给前端时字段叫什么。
	// db tag 决定 sqlx 从哪一列把值映射到这个字段。
	ID   int64  `json:"id" db:"community_id"`     // 社区ID
	Name string `json:"name" db:"community_name"` // 社区名称
}

// CommunityDetail 表示社区详情。
// 相比 Community，它额外包含简介和创建时间。
type CommunityDetail struct {
	ID           int64     `json:"id" db:"community_id"`                     // 社区ID
	Name         string    `json:"name" db:"community_name"`                 // 社区名称
	Introduction string    `json:"introduction,omitempty" db:"introduction"` // 社区简介
	CreateTime   time.Time `json:"create_time" db:"create_time"`             // 创建时间
}
