package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"bluebell/settings"
)

var rdb *redis.Client

// GetRDB 获取全局 Redis 客户端
func GetRDB() *redis.Client {
	return rdb
}

// Init 初始化 Redis 连接
func Init() (err error) {
	if settings.GlobalConfig == nil {
		return nil
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", settings.GlobalConfig.Redis.Host, settings.GlobalConfig.Redis.Port),
		Password: "",
		DB:       settings.GlobalConfig.Redis.DB,
	})

	_, err = rdb.Ping(context.Background()).Result()
	return err
}
