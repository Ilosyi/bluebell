package redis

import (
	"bluebell/models"
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// oneWeekInSeconds 表示帖子投票有效期：7 天。
	oneWeekInSeconds = 7 * 24 * 60 * 60
	// scorePerVote 表示每一票对帖子热度分的影响。
	// 当前教学项目里约定：一票 = 432 分。
	scorePerVote = 432.0 //每票的分数
)

var (
	// ErrVoteTimeExpire 表示帖子已经超过可投票时间。
	ErrVoteTimeExpire = errors.New("超出投票时间")
	// ErrVoteRepeated 表示用户当前操作没有产生变化，例如重复投相同方向。
	ErrVoteRepeated = errors.New("请勿重复投票")
)

// VoteForPost 在 Redis 中执行投票。
// Redis 负责的内容有两类：
// 1. 帖子总分（用于热榜）
// 2. 用户对某个帖子的投票记录（用于判断重复投票和改票）
func VoteForPost(userID int64, postID string, value float64) error {
	// 第一步：检查帖子是否还在可投票时间内。
	// 发布时间存在 bluebell:post:time 这个 zset 中。
	postTime := rdb.ZScore(context.Background(), getRediskey(KeyPostTimeZset), postID).Val()
	if float64(time.Now().Unix())-postTime > oneWeekInSeconds {
		return ErrVoteTimeExpire
	}

	// 第二步：准备读取“这个用户之前对这篇帖子投了什么”。
	votedKey := getRediskey(KeyPostVotedZsetPrefix + postID)
	scoreKey := getRediskey(KeyPostScoreZset)
	userIDStr := strconv.FormatInt(userID, 10)

	// oldValue 表示用户旧投票值：1 / -1 / 0。
	oldValue, err := rdb.ZScore(context.Background(), votedKey, userIDStr).Result()
	if errors.Is(err, redis.Nil) {
		// redis.Nil 表示这个用户之前没有投过票。
		if value == 0 {
			// 没投过票却要“取消投票”，本质是无效操作。
			return ErrVoteRepeated
		}
		oldValue = 0
	} else if err != nil {
		return err
	} else if value == oldValue {
		// 新旧方向完全一样，说明是重复投票。
		return ErrVoteRepeated
	}

	// 第三步：开启事务管道，一次性发送多条命令，减少网络往返。
	pipeline := rdb.TxPipeline()
	// delta 的计算方式非常关键：
	// - 0 -> 1    : +432
	// - 1 -> -1   : -864
	// - -1 -> 0   : +432
	// 核心公式就是 (新值 - 旧值) * 每票分值。
	delta := (value - oldValue) * scorePerVote
	pipeline.ZIncrBy(context.Background(), scoreKey, delta, postID)

	// 第四步：同步更新用户投票记录。
	if value == 0 {
		// 取消投票时，直接把用户从 voted zset 中删掉。
		pipeline.ZRem(context.Background(), votedKey, userIDStr)
	} else {
		// 否则把最新投票值写进去。
		pipeline.ZAdd(context.Background(), votedKey, redis.Z{
			Score:  value,
			Member: userIDStr,
		})
	}

	// 第五步：真正提交事务管道里的命令。
	_, err = pipeline.Exec(context.Background())
	return err
}

// SavePostTimeAndScore 将帖子创建时间和初始分数写入redis，并加入社区SET
func SavePostTimeAndScore(postID int64, communityID int64, t time.Time) error {
	postIDStr := strconv.FormatInt(postID, 10)
	pipeline := rdb.TxPipeline()
	// 时间 zset：用于“最新帖子”排序。
	pipeline.ZAdd(context.Background(), getRediskey(KeyPostTimeZset), redis.Z{
		Score:  float64(t.Unix()),
		Member: postIDStr,
	})
	// 分数 zset：用于“热榜”排序。
	// 初始分数先使用发布时间戳，保证新帖在没有投票时也有基本排序值。
	pipeline.ZAdd(context.Background(), getRediskey(KeyPostScoreZset), redis.Z{
		Score:  float64(t.Unix()),
		Member: postIDStr,
	})
	// 社区 set：记录“某个社区下有哪些帖子”。
	communityKey := getRediskey(KeyCommunitySetPF + strconv.FormatInt(communityID, 10))
	pipeline.SAdd(context.Background(), communityKey, postIDStr)
	// 同时删除该社区旧的交集缓存，避免之后列表页读到旧结果。
	deleteCommunityOrderCache(pipeline, communityID)
	_, err := pipeline.Exec(context.Background())
	return err
}

// DeletePostIndex 清理帖子在排序、社区和投票相关 Redis 结构中的索引。
func DeletePostIndex(postID int64, communityID int64) error {
	postIDStr := strconv.FormatInt(postID, 10)
	pipeline := rdb.TxPipeline()
	pipeline.ZRem(context.Background(), getRediskey(KeyPostTimeZset), postIDStr)
	pipeline.ZRem(context.Background(), getRediskey(KeyPostScoreZset), postIDStr)
	pipeline.Del(context.Background(), getRediskey(KeyPostVotedZsetPrefix+postIDStr))
	pipeline.SRem(context.Background(), getRediskey(KeyCommunitySetPF+strconv.FormatInt(communityID, 10)), postIDStr)
	deleteCommunityOrderCache(pipeline, communityID)
	_, err := pipeline.Exec(context.Background())
	return err
}

// MovePostCommunity 在帖子编辑变更社区时同步 Redis 社区集合。
func MovePostCommunity(postID int64, oldCommunityID int64, newCommunityID int64) error {
	// 社区没变时不需要做任何事。
	if oldCommunityID == newCommunityID {
		return nil
	}
	postIDStr := strconv.FormatInt(postID, 10)
	pipeline := rdb.TxPipeline()
	pipeline.SRem(context.Background(), getRediskey(KeyCommunitySetPF+strconv.FormatInt(oldCommunityID, 10)), postIDStr)
	pipeline.SAdd(context.Background(), getRediskey(KeyCommunitySetPF+strconv.FormatInt(newCommunityID, 10)), postIDStr)
	deleteCommunityOrderCache(pipeline, oldCommunityID)
	deleteCommunityOrderCache(pipeline, newCommunityID)
	_, err := pipeline.Exec(context.Background())
	return err
}

// deleteCommunityOrderCache 删除“社区帖子交集排序”的临时缓存 key。
// 因为 GetCommunityPostIDsInOrder 会把 ZINTERSTORE 结果缓存 60 秒，
// 所以帖子创建/删除/挪社区后必须手动删缓存。
func deleteCommunityOrderCache(pipeline redis.Pipeliner, communityID int64) {
	communityIDStr := strconv.FormatInt(communityID, 10)
	pipeline.Del(context.Background(), getRediskey(KeyPostTimeZset)+communityIDStr)
	pipeline.Del(context.Background(), getRediskey(KeyPostScoreZset)+communityIDStr)
}

// GetPostScore 获取帖子分数
func GetPostScore(postID int64) float64 {
	return rdb.ZScore(context.Background(), getRediskey(KeyPostScoreZset), strconv.FormatInt(postID, 10)).Val()
}

// GetUserVoteScore 获取用户对帖子的投票分数
func GetUserVoteScore(userID, postID int64) float64 {
	return rdb.ZScore(context.Background(), getRediskey(KeyPostVotedZsetPrefix+strconv.FormatInt(postID, 10)), strconv.FormatInt(userID, 10)).Val()
}

// GetPostIDsInOrder 按分数或时间从redis获取帖子ID列表（全局榜单）
func GetPostIDsInOrder(p *models.ParamPostList) ([]string, error) {
	// 根据排序参数决定读取哪个 zset。
	key := getRediskey(KeyPostTimeZset)
	if p.Order == "score" {
		key = getRediskey(KeyPostScoreZset)
	}
	// Redis zset 的分页是按下标截取，所以先算 start/end。
	start := (p.Page - 1) * p.Size
	end := start + p.Size - 1
	// Rev=true 表示倒序：时间最新或分数最高排最前。
	return rdb.ZRangeArgs(context.Background(), redis.ZRangeArgs{
		Key:   key,
		Start: start,
		Stop:  end,
		Rev:   true,
	}).Result()
}

// GetCommunityPostIDsInOrder 按社区+排序获取帖子ID列表（社区榜单）
// 使用 ZINTERSTORE 将社区SET与排序ZSET求交集，缓存60秒，减少重复计算
func GetCommunityPostIDsInOrder(p *models.ParamPostList) ([]string, error) {
	//1.根据排序参数选择zset key
	orderKey := getRediskey(KeyPostTimeZset)
	if p.Order == "score" {
		orderKey = getRediskey(KeyPostScoreZset)
	}
	//2.社区SET的key: bluebell:community:<communityID>
	cKey := getRediskey(KeyCommunitySetPF + strconv.FormatInt(p.CommunityID, 10))
	// 3. 缓存 key：用“排序 key + 社区 ID”区分不同社区的榜单缓存。
	key := orderKey + strconv.FormatInt(p.CommunityID, 10)
	// 4. 如果缓存不存在，就重新计算“社区集合 与 排序 zset”的交集。
	if rdb.Exists(context.Background(), key).Val() < 1 {
		pipeline := rdb.Pipeline()
		// ZINTERSTORE 的结果是一个新的 zset：
		// 只保留“既在社区集合里、又在排序 zset 里”的帖子。
		pipeline.ZInterStore(context.Background(), key, &redis.ZStore{
			Keys:      []string{cKey, orderKey},
			Aggregate: "MAX",
		})
		// 只缓存 60 秒，避免缓存长期不一致。
		pipeline.Expire(context.Background(), key, 60*time.Second)
		if _, err := pipeline.Exec(context.Background()); err != nil {
			return nil, err
		}
	}
	// 5. 从缓存 zset 中按页取帖子 ID。
	return getIDsFromKey(key, p.Page, p.Size)
}

// CountPostsInCommunity 返回某个社区下帖子总数。
// 前端分页条需要真实的 total / total_pages，不能再靠当前页条数猜测。
func CountPostsInCommunity(communityID int64) (int64, error) {
	key := getRediskey(KeyCommunitySetPF + strconv.FormatInt(communityID, 10))
	return rdb.SCard(context.Background(), key).Result()
}

// getIDsFromKey 从Zset中按分页获取帖子ID列表（倒序）
func getIDsFromKey(key string, page, size int64) ([]string, error) {
	start := (page - 1) * size
	end := start + size - 1
	return rdb.ZRangeArgs(context.Background(), redis.ZRangeArgs{
		Key:   key,
		Start: start,
		Stop:  end,
		Rev:   true,
	}).Result()
}

// GetPostVoteData 批量获取帖子的赞成票数（pipeline批量执行，减少RTT）
func GetPostVoteData(ids []string) (data []int64, err error) {
	data = make([]int64, 0, len(ids))
	// 用 pipeline 一次性批量发送多条 ZCOUNT，减少网络往返次数。
	pipeline := rdb.Pipeline()
	cmds := make([]*redis.IntCmd, 0, len(ids))
	for _, id := range ids {
		// 每个帖子自己的投票记录都存在独立 zset 中。
		key := getRediskey(KeyPostVotedZsetPrefix + id)
		// 只统计 score=1 的成员数量，也就是“赞成票数”。
		cmds = append(cmds, pipeline.ZCount(context.Background(), key, "1", "1"))
	}
	// 真正执行所有命令。
	if _, err = pipeline.Exec(context.Background()); err != nil {
		return
	}
	// cmds 的顺序和 ids 的顺序一致，因此可以按原顺序组装结果。
	for _, cmd := range cmds {
		data = append(data, cmd.Val())
	}
	return
}
