package logic

import (
	"bluebell/dao/mysql"
	"bluebell/models"
	"bluebell/pkg/jwt"
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

func LoginUp(p *models.LoginParam) (*models.User, error) {
	user := &models.User{
		Username: p.Username,
		Password: p.Password,
	}
	if err := mysql.Login(user); err != nil {
		return nil, err
	}

	token, err := jwt.GenToken(&jwt.Myclaims{
		UserID:   user.UserID,
		Username: user.Username,
	})
	if err != nil {
		return nil, err
	}
	user.Token = token
	return user, nil
}
