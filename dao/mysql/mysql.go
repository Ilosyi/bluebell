// Package mysql 负责 MySQL 连接初始化与 MySQL DAO 实现。
// 这里的 db 是一个全局 sqlx.DB 连接池，其他 mysql DAO 文件共享它。
package mysql

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"bluebell/settings"

	_ "github.com/go-sql-driver/mysql"
)

// db 是全局数据库连接池。
// 它由 mysql.Init 初始化，后续所有 DAO 查询都通过它执行。
var db *sqlx.DB

// GetDB 返回全局连接池，主要给 main.go 在退出时 Close 使用。
func GetDB() *sqlx.DB {
	return db
}

// Init 初始化数据库连接池。
// 这个函数只负责“连上数据库并配置连接池”，不负责建表。
func Init() (err error) {
	if settings.GlobalConfig == nil {
		return fmt.Errorf("config is nil")
	}

	// DSN 是 MySQL 驱动要求的连接字符串格式。
	// parseTime=True 表示把 datetime/timestamp 自动解析成 Go 的 time.Time。
	// loc=Local 表示按本地时区解析时间字段。
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		settings.GlobalConfig.MySQL.User,
		settings.GlobalConfig.MySQL.Password,
		settings.GlobalConfig.MySQL.Host,
		settings.GlobalConfig.MySQL.Port,
		settings.GlobalConfig.MySQL.DBName,
	)

	db, err = sqlx.Open("mysql", dsn)
	if err != nil {
		return err
	}

	// Ping 用于真正验证连接是否可用。
	// sqlx.Open 本身只是创建对象，不一定立即连上数据库。
	if err = db.Ping(); err != nil {
		return err
	}

	// 设置最大打开连接数
	db.SetMaxOpenConns(settings.GlobalConfig.MySQL.MaxOpenConns)
	// 设置最大空闲连接数
	db.SetMaxIdleConns(settings.GlobalConfig.MySQL.MaxIdleConns)

	return nil
}
