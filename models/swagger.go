package models

// swagger 专用响应模型，仅用于生成 API 文档，不参与业务逻辑

// SignUpData 注册成功返回数据
type SignUpData struct {
	Message string `json:"message" example:"ok"` // 提示信息
}

// LoginData 登录成功返回数据
type LoginData struct {
	UserID   string `json:"user_id" example:"1234567890"`    // 用户ID（字符串格式，避免JS精度丢失）
	Username string `json:"user_name" example:"zhangsan"`    // 用户名
	Token    string `json:"token" example:"eyJhbGciOi..."`   // JWT Token
}

// VoteData 投票成功返回数据（无额外数据）
type VoteData struct{}
