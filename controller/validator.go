package controller

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

// trans 是全局翻译器对象。
// validator 返回的是结构化错误，配合 trans 才能翻译成中文提示。
var trans ut.Translator

// InitTrans 初始化参数校验错误的翻译器。
// locale 传 "zh" 时，前端就能收到中文的字段错误提示。
func InitTrans(locale string) (err error) {
	// binding.Validator.Engine() 返回 Gin 内部正在使用的 validator 引擎。
	// 类型断言成功后，说明我们拿到了真实的 *validator.Validate，可以继续自定义行为。
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {

		// 默认情况下 validator 报错会使用结构体字段名，例如 SignUpParam.Username。
		// 这里改成优先读取 json tag，这样错误提示里会显示 username、password 等前端字段名。
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		// 创建中文/英文翻译器实例。
		zhT := zh.New()
		enT := en.New()

		// UniversalTranslator 负责“多语言管理”。
		// 第一个参数是默认语言，后续参数是支持的语言列表。
		uni := ut.New(enT, zhT, enT)

		// 根据调用方指定的 locale 取出对应翻译器。
		var ok bool
		trans, ok = uni.GetTranslator(locale)
		if !ok {
			return fmt.Errorf("uni.GetTranslator(%s) failed", locale)
		}

		// 把 validator 的默认错误翻译模板注册进去。
		// 例如 required/min/max/oneof 等规则都会有对应的人类可读消息。
		switch locale {
		case "en":
			err = enTranslations.RegisterDefaultTranslations(v, trans)
		case "zh":
			err = zhTranslations.RegisterDefaultTranslations(v, trans)
		default:
			err = enTranslations.RegisterDefaultTranslations(v, trans)
		}
		return
	}
	return
}

// removeTopStruct 去掉 validator 返回的字段路径中的结构体前缀。
// 例如把 "SignUpParam.username" 变成 "username"。
// 这样前端收到的错误 map 更短、更稳定。
func removeTopStruct(fields map[string]string) map[string]string {
	res := map[string]string{}
	for field, err := range fields {
		res[field[strings.Index(field, ".")+1:]] = err
	}
	return res
}
