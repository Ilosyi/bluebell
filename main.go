package main

import (
	"bluebell/controller"
	"bluebell/dao/mysql"
	"bluebell/dao/redis"
	"bluebell/logger"
	"bluebell/pkg/snowflake"
	"bluebell/routes"
	"bluebell/settings"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// @title bluebell 社区论坛 API
// @version 1.0
// @description bluebell 社区论坛后端接口文档
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	// 第一步：读取配置文件。
	// settings.Init 会优先读取 settings/config.yaml，如果不存在再读取 config.example.yaml。
	// 后续 MySQL、Redis、JWT、日志、服务端口等模块都依赖 settings.GlobalConfig。
	if err := settings.Init(); err != nil {
		fmt.Printf("init settings failed, err:%v\n", err)
		return
	}

	// 第二步：初始化请求参数校验器的中文翻译。
	// Gin 的 ShouldBindJSON/ShouldBindQuery 底层使用 validator；
	// 初始化后，参数错误可以返回更容易理解的中文字段提示。
	if err := controller.InitTrans("zh"); err != nil {
		fmt.Printf("init validator failed, err:%v\n", err)
		return
	}

	// 第三步：初始化 zap 日志。
	// 开发模式输出到控制台，生产模式输出到配置文件并支持日志切割。
	if err := logger.Init(&settings.GlobalConfig.Log, settings.GlobalConfig.App.Mode); err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	}
	// 程序退出前把缓冲区里的日志刷到目标输出。
	// zap.L() 表示获取全局 logger。
	defer func(l *zap.Logger) {
		err := l.Sync()
		if err != nil {
			fmt.Printf("sync logger failed, err:%v\n", err)
		}
	}(zap.L())
	zap.L().Debug("logger init success")

	// 第四步：初始化 MySQL 连接池。
	// 这里不会每次请求都新建连接，而是复用连接池里的连接。
	if err := mysql.Init(); err != nil {
		fmt.Printf("init mysql failed, err:%v\n", err)
		return
	}
	// 程序退出时关闭 MySQL 连接池。
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			fmt.Printf("close mysql failed, err:%v\n", err)
		}
	}(mysql.GetDB())

	// 第五步：初始化 Redis 客户端。
	// 项目里帖子排序、热度、投票记录都依赖 Redis。
	if err := redis.Init(); err != nil {
		fmt.Printf("init redis failed, err:%v\n", err)
		return
	}
	defer redis.GetRDB().Close()

	// 第六步：初始化雪花算法 ID 生成器。
	// 用户 ID 和帖子 ID 都不是数据库自增 ID，而是通过雪花算法生成的业务 ID。
	if err := snowflake.Init(settings.GlobalConfig.App.StartTime, settings.GlobalConfig.App.MachineID); err != nil {
		fmt.Printf("init snowflake failed, err:%v\n", err)
		return
	}

	// 第七步：注册 Gin 路由和中间件。
	r := routes.Setup()

	// 第八步：创建 HTTP Server。
	// 不直接调用 r.Run，是为了下面能用 srv.Shutdown 做优雅关机。
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", settings.GlobalConfig.App.Port),
		Handler: r,
	}

	// 在单独的 goroutine 里启动 HTTP 服务。
	// 主 goroutine 会继续往下等待系统中断信号。
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待 Ctrl+C 等中断信号。
	// make(chan os.Signal, 1) 中的 1 是缓冲区大小，避免信号发送时阻塞。
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	zap.L().Info("shutdown server...")

	// 收到退出信号后，最多等待 5 秒让正在处理的请求完成。
	// 如果 5 秒内请求还没结束，Shutdown 会返回超时错误。
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server shutdown failed, err:%v\n", err)
	}
	log.Println("server shutdown success.")

}
