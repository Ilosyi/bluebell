package logic

import (
	"bluebell/dao/redis"
	"bluebell/models"

	"go.uber.org/zap"
)

// voteForPostInRedis 指向 Redis 层真实投票实现。
// 单元测试时会替换成假函数，从而不依赖真实 Redis。
var voteForPostInRedis = redis.VoteForPost

/*
VoteForPost 背后的业务规则总结：

1. direction = 1
   - 之前没投票，现在投赞成票
   - 之前投反对票，现在改投赞成票

2. direction = -1
   - 之前没投票，现在投反对票
   - 之前投赞成票，现在改投反对票

3. direction = 0
   - 之前投过赞成票，现在取消
   - 之前投过反对票，现在取消

4. 时间限制
   - 只有发帖后一周内允许投票
   - 一周后即使前端继续请求，后端也会拒绝
*/

// VoteForPost 执行投票。
// 这个函数当前主要起“业务入口 + 日志记录”作用，真正的投票计算在 dao/redis 中完成。
func VoteForPost(userID int64, p *models.ParamVoteData) error {
	err := voteForPostInRedis(userID, p.PostId, float64(p.Direction))
	if err != nil {
		// Debug 日志保留更多上下文，排查问题时更方便。
		zap.L().Debug("redis.VoteForPost failed", zap.Int64("userID", userID), zap.String("postID", p.PostId), zap.Int8("direction", p.Direction), zap.Error(err))
		// Error 日志用于提醒线上/控制台真正出现了一次失败。
		zap.L().Error("redis.VoteForPost failed", zap.Error(err))
		return err
	}
	return nil
}
