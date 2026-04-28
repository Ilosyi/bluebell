package main

import (
	"bluebell/controller"
	"bluebell/dao/mysql"
	"bluebell/logger"
	"bluebell/pkg/snowflake"
	"bluebell/routes"
	"bluebell/settings"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.uber.org/zap"
)

// Go Web开发通用脚手架末班
func main() {
	//1.加载配置
	if err := settings.Init(); err != nil {
		fmt.Printf("init settings failed, err:%v\n", err)
		return
	}
	//2。初始化验证器
	if err := controller.InitTrans("zh"); err != nil {
		fmt.Printf("init validator failed, err:%v\n", err)
		return
	}
	//3.初始化日志
	if err := logger.Init(&settings.GlobalConfig.Log, settings.GlobalConfig.App.Mode); err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	}
	defer zap.L().Sync()
	zap.L().Debug("logger init success")
	//3.初始化MySQL连接
	if err := mysql.Init(); err != nil {
		fmt.Printf("init mysql failed, err:%v\n", err)
		return
	}
	defer mysql.GetDB().Close()
	//4,初始化Redis连接
	// if err := redis.Init(); err != nil {
	// 	fmt.Printf("init redis failed, err:%v\n", err)
	// 	return
	// }
	// defer redis.GetRDB().Close()
	//初始化雪花算法生成器
	if err := snowflake.Init(settings.GlobalConfig.App.StartTime, settings.GlobalConfig.App.MachineID); err != nil {
		fmt.Printf("init snowflake failed, err:%v\n", err)
		return
	}
	//5.注册路由
	r := routes.Setup()

	//6.启动服务（优雅关机）
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", settings.GlobalConfig.App.Port),
		Handler: r,
	}
	go func() {
		//开启一个goroutine启动服务
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	//等待中断信号来优雅关闭服务器，设置五秒超时
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	zap.L().Info("shutdown server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server shutdown failed, err:%v\n", err)
	}
	log.Println("server shutdown success.")

}
