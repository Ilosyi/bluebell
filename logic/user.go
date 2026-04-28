package logic

import (
	"bluebell/dao/mysql"
	"bluebell/models"
	"bluebell/pkg/snowflake"
)

func SignUp(p *models.SignUpParam) (err error) {
	//用户是否存在
	if err := mysql.CheckUserExist(p.Username); err != nil {
		return err
	}
	//生成UID
	userID := snowflake.GenID()
	//构造User实例
	user := &models.User{
		UserID:   userID,
		Username: p.Username,
		Password: p.Password,
	}

	//保存到数据库
	return mysql.InsertUser(user)
}
