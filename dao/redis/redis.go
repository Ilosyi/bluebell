// Package redis 负责 Redis 客户端初始化，以及 Redis DAO 实现。
// 这里的 rdb 是全局客户端，其他 redis DAO 文件共享它。
package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"bluebell/settings"
)

// rdb 是全局 Redis 客户端。
// 它由 redis.Init 初始化，后续所有 Redis 操作都通过它执行。
var rdb *redis.Client

// GetRDB 获取全局 Redis 客户端
func GetRDB() *redis.Client {
	return rdb
}

// Init 初始化 Redis 连接。
// 与 mysql.Init 类似，这里只负责“连上 Redis”，不负责初始化业务 key。
func Init() (err error) {
	if settings.GlobalConfig == nil {
		return nil
	}

	// NewClient 不会立刻验证连接，真正的连通性检查在后面的 Ping。
	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", settings.GlobalConfig.Redis.Host, settings.GlobalConfig.Redis.Port),
		Password: "",
		DB:       settings.GlobalConfig.Redis.DB,
	})

	// Ping 一次，确认 Redis 服务确实可用。
	_, err = rdb.Ping(context.Background()).Result()
	return err
}
