package logic

import (
	"bluebell/dao/mysql"
	"bluebell/models"
	"bluebell/pkg/jwt"
	"bluebell/pkg/snowflake"
	"strings"
)

var (
	// 这些包级变量默认指向 dao/pkg 层真实实现。
	// 这么写的目的主要是方便测试时替换依赖。
	checkUserExist = mysql.CheckUserExist
	insertUser     = mysql.InsertUser
	loginUser      = mysql.Login
	getUserProfile = mysql.GetUserById
	updateProfile  = mysql.UpdateUserProfile
	checkNickname  = mysql.CheckNicknameAvailable
	genUserID      = snowflake.GenID
	genToken       = jwt.GenToken
)

// SignUp 执行用户注册。
// 主要流程：
// 1. 检查用户名是否已存在
// 2. 生成业务用户 ID
// 3. 组装 User 结构体
// 4. 交给 dao 层入库（dao 层会负责密码加密）
func SignUp(p *models.SignUpParam) (err error) {
	// 第一步：校验用户名是否已被占用。
	if err := checkUserExist(p.Username); err != nil {
		return err
	}
	// 第二步：生成业务用户 ID。
	userID := genUserID()
	// 第三步：组装用户对象。
	// 注意这里的 Password 仍然是明文，真正入库前会在 dao 层加密。
	user := &models.User{
		UserID:   userID,
		Username: p.Username,
		Password: p.Password,
	}

	// 第四步：保存到数据库。
	return insertUser(user)
}

// Login 执行登录。
// 主要流程：
// 1. 让 dao 层根据“账号或昵称 + 密码”校验用户
// 2. 登录成功后生成 JWT
// 3. 把 token 挂到 user 结构体上返回给 controller
func Login(p *models.LoginParam) (*models.User, error) {
	// 去掉前后空格，避免“用户输入多打空格”导致本来正确的账号无法登录。
	user := &models.User{
		Username: strings.TrimSpace(p.Username),
		Password: p.Password,
	}
	// dao 层会负责查库和密码比对。
	if err := loginUser(user); err != nil {
		return nil, err
	}

	// 登录成功后生成 JWT，后续前端会把它放到 Authorization 请求头中。
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

// GetUserProfile 获取某个用户的资料。
// 当前主要用于“当前登录用户资料”接口。
func GetUserProfile(userID int64) (*models.User, error) {
	return getUserProfile(userID)
}

// UpdateUserProfile 更新用户资料。
// 这里会先做“输入清洗 + 昵称唯一校验”，再进入 dao 层真正更新数据库。
func UpdateUserProfile(userID int64, p *models.UpdateUserProfileParam) (*models.User, error) {
	// 先去掉前后空格，避免出现“看起来一样但实际上带空格”的昵称。
	p.Nickname = strings.TrimSpace(p.Nickname)
	p.AvatarURL = strings.TrimSpace(p.AvatarURL)
	p.Bio = strings.TrimSpace(p.Bio)

	// 先从业务层做一次显式昵称唯一校验，给前端返回更明确的错误。
	if err := checkNickname(userID, p.Nickname); err != nil {
		if strings.Contains(err.Error(), "昵称已存在") {
			return nil, ErrNicknameExist
		}
		return nil, err
	}

	// 即使前面已经查过，数据库层面仍然可能因为并发写入触发唯一索引冲突。
	// 所以这里还要把 Duplicate entry 再映射成统一的昵称冲突错误。
	if err := updateProfile(userID, p); err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, ErrNicknameExist
		}
		return nil, err
	}
	return getUserProfile(userID)
}
