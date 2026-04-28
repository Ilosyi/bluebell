package logic

import (
	"bluebell/dao/mysql"
	"bluebell/models"
	"bluebell/pkg/snowflake"
	"errors"
)

func SignUp(p *models.SignUpParam) (err error) {
	//用户是否存在
	exist, err := mysql.CheckUserExist(p.Username)
	if err != nil {
		//数据库查询错误
		return err
	}
	if exist {
		return errors.New("用户已存在")
	}
	//用户密码加密SignUp
	userID := snowflake.GenID()
	//构造User实例

	//保存到数据库

}
