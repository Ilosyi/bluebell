package middlewares

import (
	"bluebell/controller"
	"bluebell/pkg/jwt"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware 是“需要登录接口”前面的鉴权中间件。
// 它做的事情很明确：
// 1. 从 Authorization 请求头里取 token
// 2. 校验 token 格式和签名
// 3. 把 userID / username 写入 Gin Context
// 4. 让后续 handler 可以直接读取当前登录用户
func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		// 当前项目约定 token 放在请求头 Authorization 中，
		// 并且格式必须是：Bearer <token>。
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			controller.ResponseError(c, controller.CodeNeedLogin)
			c.Abort()
			return
		}

		// 例如 "Bearer abc.def.ghi" 会被分成两段：
		// parts[0] = "Bearer"
		// parts[1] = "abc.def.ghi"
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			controller.ResponseError(c, controller.CodeInvalidToken)
			c.Abort()
			return
		}
		// 调用 pkg/jwt 解析并校验 token。
		mc, err := jwt.ParseToken(parts[1])
		if err != nil {
			controller.ResponseError(c, controller.CodeInvalidToken)
			c.Abort()
			return
		}

		// 鉴权成功后，把当前用户信息写进 Gin Context。
		// 后续 controller 可以通过 c.Get(...) 拿到这些值。
		c.Set(controller.CtxUserIDKey, mc.UserID)
		c.Set(controller.CtxUsernameKey, mc.Username)

		// 放行到下一个中间件或最终 handler。
		c.Next()
	}
}
