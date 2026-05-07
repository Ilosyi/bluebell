package logic

import "errors"

var (
	// ErrPostNotFound 用于在 logic 层统一表达“帖子不存在”，
	// controller 再把它映射成稳定的业务码，而不是一律返回服务繁忙。
	ErrPostNotFound = errors.New("post not found")
	// ErrCommunityNotFound 用于在 logic 层统一表达“社区不存在”。
	ErrCommunityNotFound = errors.New("community not found")
	// ErrForbidden 表示当前用户没有权限操作该资源。
	ErrForbidden = errors.New("forbidden")
	// ErrPostNotReady 表示草稿内容不完整，不能发布。
	ErrPostNotReady = errors.New("post not ready")
	// ErrNicknameExist 表示昵称已被其他用户占用。
	ErrNicknameExist = errors.New("nickname exists")
)
