package models

// User 用户数据模型
type User struct {
	UserID   int64  `db:"user_id"`           // 用户ID
	Username string `db:"username"`          // 用户名
	Password string `db:"password"`          // 密码（加密存储）
	Token    string `db:"-" json:"token"`    // JWT Token（不入库）
}
