package mysql

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"bluebell/settings"

	_ "github.com/go-sql-driver/mysql"
)

var db *sqlx.DB

func GetDB() *sqlx.DB {
	return db
}

// Init 初始化数据库连接
func Init() (err error) {
	if settings.GlobalConfig == nil {
		return fmt.Errorf("config is nil")
	}

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

	if err = db.Ping(); err != nil {
		return err
	}

	// 设置最大打开连接数
	db.SetMaxOpenConns(settings.GlobalConfig.MySQL.MaxOpenConns)
	// 设置最大空闲连接数
	db.SetMaxIdleConns(settings.GlobalConfig.MySQL.MaxIdleConns)

	return nil
}
