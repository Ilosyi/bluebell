package snowflake

import (
	"time"

	"github.com/bwmarrin/snowflake"
	sf "github.com/bwmarrin/snowflake"
)

var node *sf.Node

// Init 初始化雪花算法节点。
// 雪花算法的核心思想是：在分布式环境下也能生成趋势递增且全局唯一的 int64 ID。
// startTime 会被设置成算法纪元（Epoch），machineID 用于区分不同机器节点。
func Init(startTime string, machineID int64) (err error) {
	var st time.Time
	// 项目里约定 startTime 格式为 "2006-01-02"。
	st, err = time.Parse("2006-01-02", startTime)
	if err != nil {
		return
	}
	// snowflake.Epoch 需要毫秒时间戳。
	snowflake.Epoch = st.UnixNano() / 1000000
	node, err = snowflake.NewNode(machineID)
	return
}

// GenID 生成一个全局唯一业务 ID。
// 常用于用户 ID、帖子 ID 等需要暴露给前端的主键。
func GenID() int64 {
	return node.Generate().Int64()
}
