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
	oneWeekInSeconds = 7 * 24 * 60 * 60
	scorePerVote     = 432.0 //每票的分数
)

var (
	ErrVoteTimeExpire = errors.New("超出投票时间")
	ErrVoteRepeated   = errors.New("请勿重复投票")
)

func VoteForPost(userID int64, postID string, value float64) error {
	//1.判断投票限制
	//去redis取帖子发布时间
	postTime := rdb.ZScore(context.Background(), getRediskey(KeyPostTimeZset), postID).Val()
	if float64(time.Now().Unix())-postTime > oneWeekInSeconds {
		return ErrVoteTimeExpire
	}
	//2.更新帖子分数
	//根据用户投票方向计算分数变化
	votedKey := getRediskey(KeyPostVotedZsetPrefix + postID)
	scoreKey := getRediskey(KeyPostScoreZset)
	userIDStr := strconv.FormatInt(userID, 10)

	oldValue, err := rdb.ZScore(context.Background(), votedKey, userIDStr).Result()
	if errors.Is(err, redis.Nil) {
		//用户未投过票
		if value == 0 {
			return ErrVoteRepeated //未投票不能取消
		}
		oldValue = 0
	} else if err != nil {
		return err
	} else if value == oldValue {
		return ErrVoteRepeated //重复投票
	}

	pipeline := rdb.TxPipeline()
	//计算分数变化量：新值 - 旧值，每票432分
	delta := (value - oldValue) * scorePerVote
	pipeline.ZIncrBy(context.Background(), scoreKey, delta, postID)

	//3.记录用户的投票数据
	if value == 0 {
		//取消投票，从voted set中移除
		pipeline.ZRem(context.Background(), votedKey, userIDStr)
	} else {
		//赞成或反对，更新voted set
		pipeline.ZAdd(context.Background(), votedKey, redis.Z{
			Score:  value,
			Member: userIDStr,
		})
	}

	_, err = pipeline.Exec(context.Background())
	return err
}

// SavePostTimeAndScore 将帖子创建时间和初始分数写入redis，并加入社区SET
func SavePostTimeAndScore(postID int64, communityID int64, t time.Time) error {
	postIDStr := strconv.FormatInt(postID, 10)
	pipeline := rdb.TxPipeline()
	//写入时间zset
	pipeline.ZAdd(context.Background(), getRediskey(KeyPostTimeZset), redis.Z{
		Score:  float64(t.Unix()),
		Member: postIDStr,
	})
	//写入分数zset
	pipeline.ZAdd(context.Background(), getRediskey(KeyPostScoreZset), redis.Z{
		Score:  float64(t.Unix()),
		Member: postIDStr,
	})
	//加入社区SET: bluebell:community:<communityID>
	communityKey := getRediskey(KeyCommunitySetPF + strconv.FormatInt(communityID, 10))
	pipeline.SAdd(context.Background(), communityKey, postIDStr)
	_, err := pipeline.Exec(context.Background())
	return err
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
	//根据排序参数选择redis key
	key := getRediskey(KeyPostTimeZset) //默认按时间排序
	if p.Order == "score" {
		key = getRediskey(KeyPostScoreZset)
	}
	//计算分页起止位置
	start := (p.Page - 1) * p.Size
	end := start + p.Size - 1
	//ZRangeArgs+Rev: 从大到小排序（分数最高/时间最新在前）
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
	//3.缓存key：orderKey + communityID，用于缓存ZINTERSTORE的结果
	key := orderKey + strconv.FormatInt(p.CommunityID, 10)
	//4.如果缓存不存在，执行ZINTERSTORE计算
	//rdb.Exists返回key是否存在，返回1表示存在，0表示不存在
	if rdb.Exists(context.Background(), key).Val() < 1 {
		pipeline := rdb.Pipeline()
		//ZINTERSTORE: 社区SET与排序Zset求交集，聚合方式取MAX
		pipeline.ZInterStore(context.Background(), key, &redis.ZStore{
			Keys:      []string{cKey, orderKey},
			Aggregate: "MAX",
		})
		//设置60秒过期
		pipeline.Expire(context.Background(), key, 60*time.Second)
		if _, err := pipeline.Exec(context.Background()); err != nil {
			return nil, err
		}
	}
	//5.从缓存的Zset中分页获取帖子ID
	return getIDsFromKey(key, p.Page, p.Size)
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
	//使用pipeline批量发送ZCOUNT命令，一次网络往返
	pipeline := rdb.Pipeline()
	cmds := make([]*redis.IntCmd, 0, len(ids))
	for _, id := range ids {
		key := getRediskey(KeyPostVotedZsetPrefix + id)
		cmds = append(cmds, pipeline.ZCount(context.Background(), key, "1", "1"))
	}
	//批量执行
	if _, err = pipeline.Exec(context.Background()); err != nil {
		return
	}
	//按顺序读取结果
	for _, cmd := range cmds {
		data = append(data, cmd.Val())
	}
	return
}
