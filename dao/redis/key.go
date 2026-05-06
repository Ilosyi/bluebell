package redis

// redis key
const (
	KeyPrefix              = "bluebell:"
	KeyPostTimeZset        = "post:time"           //帖子时间线 zset, member=post_id, score=发布时间戳
	KeyPostScoreZset       = "post:score"          //帖子投票分数 zset, member=post_id, score=帖子分数
	KeyPostVotedZsetPrefix = "post:voted:"         //参数是帖子id，记录用户投票类型 zset, member=user_id, score=投票值(1/-1)
	KeyCommunitySetPF      = "community:"          //社区帖子集合 set, member=post_id
)

// 给redis key加上前缀
func getRediskey(key string) string {
	return KeyPrefix + key
}
