package redis

// 这里统一定义项目里用到的 Redis key 名称。
// 这样做的好处是：
// 1. 所有 key 都集中管理，不容易写错。
// 2. 后续如果要改 key 命名规则，只需要改这里。
const (
	KeyPrefix              = "bluebell:"
	KeyPostTimeZset        = "post:time"   //帖子时间线 zset, member=post_id, score=发布时间戳
	KeyPostScoreZset       = "post:score"  //帖子投票分数 zset, member=post_id, score=帖子分数
	KeyPostVotedZsetPrefix = "post:voted:" //参数是帖子id，记录用户投票类型 zset, member=user_id, score=投票值(1/-1)
	KeyCommunitySetPF      = "community:"  //社区帖子集合 set, member=post_id
)

// getRediskey 给业务 key 补上统一前缀。
// 例如传入 "post:time"，返回 "bluebell:post:time"。
func getRediskey(key string) string {
	return KeyPrefix + key
}
