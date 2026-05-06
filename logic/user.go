package logic

import (
	"bluebell/dao/mysql"
	"bluebell/models"
	"bluebell/pkg/jwt"
	"bluebell/pkg/snowflake"
)

var (
	checkUserExist = mysql.CheckUserExist
	insertUser     = mysql.InsertUser
	loginUser      = mysql.Login
	genUserID      = snowflake.GenID
	genToken       = jwt.GenToken
)

func SignUp(p *models.SignUpParam) (err error) {
	//用户是否存在
	if err := checkUserExist(p.Username); err != nil {
		return err
	}
	//生成UID
	userID := genUserID()
	//构造User实例
	user := &models.User{
		UserID:   userID,
		Username: p.Username,
		Password: p.Password,
	}

	//保存到数据库
	return insertUser(user)
}

func Login(p *models.LoginParam) (*models.User, error) {
	user := &models.User{
		Username: p.Username,
		Password: p.Password,
	}
	if err := loginUser(user); err != nil {
		return nil, err
	}

	token, err := genToken(&jwt.Myclaims{
		UserID:   user.UserID,
		Username: user.Username,
	})
	if err != nil {
		return nil, err
	}
	user.Token = token
	return user, nil
}
