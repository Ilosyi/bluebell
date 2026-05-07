package middlewares

import (
	"net/http"

	"bluebell/controller"
	"bluebell/settings"

	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
)

// RateLimitMiddleware 创建一个全局令牌桶限流中间件。
//
// “全局”指的是：整个 Gin 服务共享同一个 bucket。
// 也就是说，不管请求来自哪个用户、哪个 IP、哪个接口，都会一起消耗这个桶里的令牌。
// 这种方式实现简单，适合保护整体服务不被瞬时流量压垮。
//
// 如果以后想做得更细，可以扩展成：
// - 每个 IP 一个 bucket
// - 每个登录用户一个 bucket
// - 登录、发帖、搜索等不同接口使用不同 bucket
func RateLimitMiddleware(cfg *settings.RateLimitConfig) gin.HandlerFunc {
	// 如果配置为空，或者显式关闭限流，就返回一个“直接放行”的中间件。
	// 这样测试环境或临时调试时不会因为缺少配置导致 panic。
	if cfg == nil || !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Rate 和 Capacity 必须大于 0，否则令牌桶没有明确意义。
	// 这里选择“配置不合法时放行”，避免因为配置写错导致服务完全不可用。
	if cfg.Rate <= 0 || cfg.Capacity <= 0 {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// NewBucketWithRate 会创建令牌桶。
	// cfg.Rate 表示每秒补充多少令牌，cfg.Capacity 表示桶容量上限。
	bucket := ratelimit.NewBucketWithRate(cfg.Rate, cfg.Capacity)

	return func(c *gin.Context) {
		// 每个请求都尝试拿 1 个令牌。
		// TakeAvailable 不会阻塞：拿得到就返回拿到的数量，拿不到就立刻返回 0。
		if bucket.TakeAvailable(1) == 0 {
			// AbortWithStatusJSON 会终止后续中间件和 handler。
			// 这里使用 HTTP 429，语义是 Too Many Requests。
			c.AbortWithStatusJSON(http.StatusTooManyRequests, &controller.Response{
				Code: controller.CodeTooManyRequests,
				Msg:  controller.CodeTooManyRequests.Msg(),
				Data: nil,
			})
			return
		}

		// 拿到令牌后继续执行后续中间件和真正的业务 handler。
		c.Next()
	}
}
