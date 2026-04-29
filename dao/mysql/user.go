package mysql

import (
	"bluebell/models"
	"crypto/md5"
	"encoding/hex"
	"errors"
)

// InsertUser 向数据库插入一条新的用户记录
func InsertUser(user *models.User) error {
	//对密码进行加密
	user.Password, _ = encryptPassword(user.Password)
	//执行SQL语句入库
	sqlstr := "insert into user (user_id,username,password) values (?,?,?)"
	_, err := db.Exec(sqlstr, user.UserID, user.Username, user.Password)
	return err
}

// QueryUserById 根据Id查询用户
func QueryUserById() {

}

// CheckUserExist 查询数据库，判断用户是否存在
func CheckUserExist(username string) error {
	//查询数据库，判断用户是否存在
	sqlstr := "select count(user_id) from user where username=?"
	var count int
	if err := db.Get(&count, sqlstr, username); err != nil {
		return err
	}
	if count > 0 {
		return errors.New("用户已存在")
	}
	return nil

}

// 全局固定盐
const globalSalt = "losyi"

func encryptPassword(password string) (string, error) {
	h := md5.New()
	h.Write([]byte(globalSalt))
	return hex.EncodeToString(h.Sum([]byte(password))), nil
}

func Login(user *models.User) error {
	//先记录原始密码
	oPassword := user.Password
	sqlstr := "select user_id, username, password from user where username=?"
	err := db.Get(user, sqlstr, user.Username)
	if err != nil {
		return err
	}
	//判断密码是否正确
	password, err := encryptPassword(oPassword)
	if err != nil {
		return err
	}
	if user.Password != password {
		return errors.New("用户名或密码错误")
	}
	return nil
}
