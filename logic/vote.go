package logic

import (
	"bluebell/dao/redis"
	"bluebell/models"

	"go.uber.org/zap"
)

/*投票功能
投票的几种情况
direction=1时
	1.之前没有投过票，现在投赞成票
	2.之前投过反对票，现在改投赞成票
direction=-1时
	1.之前没有投过票，现在投反对票
	2.之前投过赞成票，现在改投反对票
direction=0时
	1.之前投过赞成票，现在取消投票
	2.之前投过反对票，现在取消投票

投票限制：
只有一个星期内的帖子允许投票，超过则不允许投票，到期之后就不需要redis key了，存储到mysql表中
*/

func VoteForPost(userID int64, p *models.ParamVoteData) error {
	err := redis.VoteForPost(userID, p.PostId, float64(p.Direction))
	if err != nil {
		//Debug日志
		zap.L().Debug("redis.VoteForPost failed", zap.Int64("userID", userID), zap.String("postID", p.PostId), zap.Int8("direction", p.Direction), zap.Error(err))
		zap.L().Error("redis.VoteForPost failed", zap.Error(err))
		return err
	}
	return nil
}
