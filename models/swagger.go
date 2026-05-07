package models

// swagger.go 里的结构体只用于生成 Swagger 文档。
// 它们的存在是为了让 API 文档展示更清晰的返回结构，
// 不参与真正的业务计算。

// SignUpData 表示注册成功后的返回 data。
type SignUpData struct {
	Message string `json:"message" example:"ok"` // 提示信息
}

// LoginData 表示登录成功后的返回 data。
type LoginData struct {
	UserID   string `json:"user_id" example:"1234567890"`  // 用户ID（字符串格式，避免JS精度丢失）
	Username string `json:"user_name" example:"zhangsan"`  // 用户名
	Token    string `json:"token" example:"eyJhbGciOi..."` // JWT Token
}

// VoteData 目前没有额外字段，只是为了在 Swagger 中表达“这是投票接口的成功返回”。
type VoteData struct{}
