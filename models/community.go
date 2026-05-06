package models

import "time"

// Community 社区简要信息（列表展示用）
type Community struct {
	ID   int64  `json:"id" db:"community_id"`   // 社区ID
	Name string `json:"name" db:"community_name"` // 社区名称
}

// CommunityDetail 社区详细信息
type CommunityDetail struct {
	ID           int64     `json:"id" db:"community_id"`               // 社区ID
	Name         string    `json:"name" db:"community_name"`           // 社区名称
	Introduction string    `json:"introduction,omitempty" db:"introduction"` // 社区简介
	CreateTime   time.Time `json:"create_time" db:"create_time"`       // 创建时间
}
