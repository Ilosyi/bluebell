package controller

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 上下文键名：在 Gin 的 Context 中存取用户信息时使用（统一命名以便复用）
const (
	// CtxUserIDKey 在 Context 中保存/读取用户 ID（int64）
	CtxUserIDKey = "userID"
	// CtxUsernameKey 在 Context 中保存/读取用户名（string）
	CtxUsernameKey = "username"
)

// ErrorUserNotLogin 表示当前请求未登录或无法从 Context 中读取用户信息
var ErrorUserNotLogin = errors.New("用户未登录")

// getCurrentUserID 从 gin.Context 中提取当前登录用户的 ID
// 返回值：
// - userID: 成功时返回用户 ID（int64）
// - err: 如果 Context 中没有该键或者类型断言失败，返回 ErrorUserNotLogin
// 注意：该函数仅做简单提取与类型检查，不做额外权限校验。
func getCurrentUserID(c *gin.Context) (userID int64, err error) {
	uid, ok := c.Get(CtxUserIDKey)
	if !ok {
		// 上下文中未设置用户 ID，视为未登录
		err = ErrorUserNotLogin
		return
	}
	// 类型断言为 int64（中间件写入时应保证类型一致）
	userID, ok = uid.(int64)
	if !ok {
		// 类型不匹配也视为未登录
		err = ErrorUserNotLogin
		return
	}
	return
}

// getPageInfo 从请求中解析分页参数 page 和 size
// 行为说明：
// - 如果解析失败或参数为空，page 默认为 1，size 默认为 10
// - 返回值顺序为 (page, size)
// 该方法方便在业务 handler 中统一获取分页参数并进行默认值处理。
func getPageInfo(c *gin.Context) (int64, int64) {
	pageStr := c.Query("page")
	sizeStr := c.Query("size")

	var (
		page int64
		size int64
		err  error
	)

	// 解析 page，若失败则使用默认值 1
	page, err = strconv.ParseInt(pageStr, 10, 64)
	if err != nil {
		page = 1
	}
	// 解析 size，若失败则使用默认值 10
	size, err = strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		size = 10
	}
	return page, size
}
