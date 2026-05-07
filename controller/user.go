package controller

import (
	"bluebell/logic"
	"bluebell/models"
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

var (
	// 这些变量默认指向 logic 层真实实现。
	// 测试时可以把它们替换成假函数，从而隔离数据库/Redis 依赖。
	signUp            = logic.SignUp
	login             = logic.Login
	getProfile        = logic.GetUserProfile
	updateUserProfile = logic.UpdateUserProfile
)

// SignUpHandler 用户注册
// @Summary 用户注册
// @Description 用户注册接口，传入用户名、密码、确认密码
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param object body models.SignUpParam true "注册参数"
// @Success 200 {object} Response{data=models.SignUpData} "成功"
// @Failure 200 {object} Response "参数错误或用户已存在"
// @Router /signup [post]
func SignUpHandler(c *gin.Context) {
	// 第一步：读取并校验请求体。
	// 请求 JSON 形如：
	// {"username":"xxx","password":"xxx","re_password":"xxx"}
	p := new(models.SignUpParam)
	if err := c.ShouldBindJSON(p); err != nil {
		// ShouldBindJSON 失败时，可能是 JSON 格式错，也可能是字段校验没通过。
		zap.L().Error("注册请求参数有误", zap.Error(err))

		// 这里优先判断是不是 validator 的字段校验错误。
		if verr, ok := errors.AsType[validator.ValidationErrors](err); ok {
			detail := removeTopStruct(verr.Translate(trans))
			zap.L().Debug("注册参数校验失败", zap.Any("detail", detail))
			ResponseErrorWithMsg(c, CodeInvalidParam, detail)
			return
		}

		// 否则直接返回原始错误信息，通常是 JSON 格式错误等。
		ResponseErrorWithMsg(c, CodeInvalidParam, err.Error())
		return
	}

	// 第二步：进入 logic 层执行注册。
	if err := signUp(p); err != nil {
		switch {
		case strings.Contains(err.Error(), "用户已存在"):
			ResponseError(c, CodeUserExist)
		case errors.Is(err, logic.ErrNicknameExist), strings.Contains(err.Error(), "昵称已存在"):
			ResponseError(c, CodeNicknameExist)
		default:
			ResponseError(c, CodeServerBusy)
		}
		return
	}

	// 第三步：返回成功响应。
	ResponseSuccess(c, gin.H{"message": "ok"})
}

// LoginHandler 处理登录请求
// @Summary 用户登录
// @Description 用户登录接口，返回 user_id、user_name 和 JWT token
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param object body models.LoginParam true "登录参数"
// @Success 200 {object} Response{data=models.LoginData} "成功"
// @Failure 200 {object} Response "参数错误或用户名密码错误"
// @Router /login [post]
func LoginHandler(c *gin.Context) {
	// 第一步：绑定并校验登录请求。
	p := new(models.LoginParam)
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Error("登录请求参数有误", zap.Error(err))
		if verr, ok := errors.AsType[validator.ValidationErrors](err); ok {
			detail := removeTopStruct(verr.Translate(trans))
			zap.L().Debug("登录参数校验失败", zap.Any("detail", detail))
			ResponseErrorWithMsg(c, CodeInvalidParam, detail)
			return
		}
		// 不是 validator 错误时，通常是 JSON 格式问题。
		ResponseErrorWithMsg(c, CodeInvalidParam, err.Error())
		return
	}

	// 第二步：进入 logic 层登录。
	user, err := login(p)
	if err != nil {
		zap.L().Error("登录失败", zap.Error(err))
		switch {
		case strings.Contains(err.Error(), "用户名或密码错误"), strings.Contains(err.Error(), "账号或密码错误"):
			ResponseError(c, CodeInvalidPassword)
		default:
			ResponseError(c, CodeServerBusy)
		}
		return
	}

	// 第三步：返回登录成功数据。
	// user_id 用字符串返回，避免前端 JavaScript 处理大整数时精度丢失。
	ResponseSuccess(c, gin.H{
		"user_id":    fmt.Sprintf("%d", user.UserID), // id值大于1<<53-1  int64类型的最大值是1<<63-1
		"user_name":  user.Username,
		"username":   user.Username,
		"nickname":   user.Nickname,
		"avatar_url": user.AvatarURL,
		"bio":        user.Bio,
		"token":      user.Token,
	})
}

// GetMeHandler 返回“当前登录用户”的资料。
// 它依赖 JWT 中间件先把 userID 放到 Gin Context 中。
func GetMeHandler(c *gin.Context) {
	userID, err := getCurrentUserID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	user, err := getProfile(userID)
	if err != nil {
		zap.L().Error("logic.GetUserProfile failed", zap.Int64("userID", userID), zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, user)
}

// UpdateMeHandler 更新当前登录用户资料。
// 这里只允许修改 nickname、avatar_url、bio。
func UpdateMeHandler(c *gin.Context) {
	userID, err := getCurrentUserID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	p := new(models.UpdateUserProfileParam)
	if err := c.ShouldBindJSON(p); err != nil {
		// 这里直接把校验错误消息返回给前端，便于表单显示。
		ResponseErrorWithMsg(c, CodeInvalidParam, err.Error())
		return
	}

	// 头像地址当前只接受 http/https 外链。
	// 这样可以避免把明显错误的地址写进数据库。
	if p.AvatarURL != "" && !(strings.HasPrefix(p.AvatarURL, "https://") || strings.HasPrefix(p.AvatarURL, "http://")) {
		ResponseErrorWithMsg(c, CodeInvalidParam, "头像地址仅支持 http 或 https")
		return
	}

	// 调用 logic 层执行昵称唯一校验和资料更新。
	user, err := updateUserProfile(userID, p)
	if err != nil {
		zap.L().Error("logic.UpdateUserProfile failed", zap.Int64("userID", userID), zap.Error(err))
		if errors.Is(err, logic.ErrNicknameExist) {
			ResponseError(c, CodeNicknameExist)
			return
		}
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, user)
}
