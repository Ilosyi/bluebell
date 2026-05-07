package controller

// ResCode 是项目里的“业务状态码”类型。
// 注意：这里的 code 和 HTTP 状态码不是一回事。
// 项目约定大多数接口都返回 HTTP 200，
// 然后通过 JSON 里的 code 字段表达业务成功/失败。
type ResCode int64

const (
	// iota 会从 0 开始自动递增，因此 CodeSuccess = 1000，后面依次 +1。
	CodeSuccess ResCode = 1000 + iota
	CodeInvalidParam
	CodeUserExist
	CodeUserNotExist
	CodeInvalidPassword
	CodeServerBusy
	CodeNeedLogin
	CodeInvalidToken
	CodePostNotFound
	CodeCommunityNotFound
	CodeForbidden
	CodePostNotReady
	CodeNicknameExist
)

// codeMsgMap 用来把业务码映射成默认消息。
// controller.ResponseError(code) 最终就是从这里取默认中文提示。
var codeMsgMap = map[ResCode]string{
	CodeSuccess:           "success",
	CodeInvalidParam:      "请求参数错误",
	CodeUserExist:         "用户已存在",
	CodeUserNotExist:      "用户不存在",
	CodeInvalidPassword:   "账号或密码错误",
	CodeServerBusy:        "服务繁忙",
	CodeNeedLogin:         "需要登录",
	CodeInvalidToken:      "无效的token",
	CodePostNotFound:      "帖子不存在",
	CodeCommunityNotFound: "社区不存在",
	CodeForbidden:         "无权限操作",
	CodePostNotReady:      "草稿内容不完整，不能发布",
	CodeNicknameExist:     "昵称已存在",
}

// Msg 根据业务码返回默认消息。
// 如果传入了未知业务码，则统一降级为“服务繁忙”。
func (code ResCode) Msg() string {
	msg, ok := codeMsgMap[code]
	if !ok {
		return codeMsgMap[CodeServerBusy]
	}
	return msg
}
