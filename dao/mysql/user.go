package mysql

import (
	"bluebell/models"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
)

// GetUserById 根据业务用户 ID 查询用户资料。
// 这里刻意没有查 password，避免无意义地把密码哈希带到上层。
func GetUserById(userId int64) (user *models.User, err error) {
	user = new(models.User)
	sqlstr := "select user_id, username, coalesce(nickname, '') as nickname, coalesce(avatar_url, '') as avatar_url, coalesce(bio, '') as bio, create_time from user where user_id=?"
	err = db.Get(user, sqlstr, userId)
	return
}

// InsertUser 向数据库插入一条新的用户记录
func InsertUser(user *models.User) error {
	// 入库前先把明文密码转换成哈希。
	user.Password, _ = encryptPassword(user.Password)
	// 如果没有显式设置昵称，默认用用户名做初始昵称。
	if user.Nickname == "" {
		user.Nickname = user.Username
	}
	// 执行 insert 入库。
	sqlstr := "insert into user (user_id,username,password,nickname) values (?,?,?,?)"
	_, err := db.Exec(sqlstr, user.UserID, user.Username, user.Password, user.Nickname)
	return err
}

// QueryUserById 目前未使用，保留作占位。
func QueryUserById() {

}

// CheckUserExist 检查用户名是否已存在。
// 如果已存在，返回中文业务错误“用户已存在”。
func CheckUserExist(username string) error {
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

// globalSalt 是项目当前用于密码哈希的固定盐。
// 这只适合教学项目，生产环境应改用 bcrypt/argon2 等更安全的方案。
const globalSalt = "losyi"

// encryptPassword 对密码做一次“加盐 + MD5”。
// 返回值是十六进制字符串。
func encryptPassword(password string) (string, error) {
	h := md5.New()
	// 先写入盐，再把 password 作为 Sum 的初始数据拼接进去。
	h.Write([]byte(globalSalt))
	return hex.EncodeToString(h.Sum([]byte(password))), nil
}

// Login 按“用户名或昵称 + 密码”执行登录校验。
// user 参数既是输入参数，也是输出参数：
// - 调用前只需要填 Username 和 Password
// - 调用后会回填 UserID、Username、Nickname 等数据库字段
func Login(user *models.User) error {
	// 先保留用户输入的原始密码，避免后面 db.Get 把 user.Password 覆盖掉后丢失。
	oPassword := user.Password
	// 支持“用户名登录”或“昵称登录”。
	sqlstr := "select user_id, username, password, coalesce(nickname, '') as nickname, coalesce(avatar_url, '') as avatar_url, coalesce(bio, '') as bio from user where username=? or nickname=? limit 1"
	err := db.Get(user, sqlstr, user.Username, user.Username)
	if err != nil {
		// 用户不存在时统一返回“账号或密码错误”，避免暴露“这个账号是否存在”。
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("账号或密码错误")
		}
		return err
	}
	// 把输入密码按同样规则加密后与数据库哈希做比较。
	password, err := encryptPassword(oPassword)
	if err != nil {
		return err
	}
	if user.Password != password {
		return errors.New("账号或密码错误")
	}
	return nil
}

// CheckNicknameAvailable 校验昵称是否可用。
// 规则：
// 1. 不能与其他人的昵称重复
// 2. 也不能与其他人的账号名冲突
func CheckNicknameAvailable(userID int64, nickname string) error {
	var count int
	sqlstr := "select count(user_id) from user where (nickname=? or username=?) and user_id<>?"
	if err := db.Get(&count, sqlstr, nickname, nickname, userID); err != nil {
		return err
	}
	if count > 0 {
		return errors.New("昵称已存在")
	}
	return nil
}

// UpdateUserProfile 更新用户资料。
// 这里只更新 nickname/avatar_url/bio，不允许修改 username/password。
func UpdateUserProfile(userID int64, p *models.UpdateUserProfileParam) error {
	sqlstr := "update user set nickname=?, avatar_url=?, bio=? where user_id=?"
	_, err := db.Exec(sqlstr, p.Nickname, p.AvatarURL, p.Bio, userID)
	return err
}
