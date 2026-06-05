// Package validate wraps go-playground/validator with Chinese translations.
package validate

import (
	"bytes"
	"log/slog"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
	"github.com/spf13/cast"
)

var validate = validator.New()

var trans ut.Translator

func init() {
	// 注册中文翻译器
	zhEntity := zh.New()
	uni := ut.New(zhEntity, zhEntity)
	trans, _ = uni.GetTranslator("zh")
	err := zhTranslations.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		slog.Error(cast.ToString(err))
	}
}

// Valid validates params using struct tags from go-playground/validator.
func Valid(params any) error {
	return validate.Struct(params)
}

// FormatError returns translated text for validator validation errors.
func FormatError(err error) string {
	var msg bytes.Buffer
	for _, errItem := range err.(validator.ValidationErrors) {
		// 输出中文错误信息
		msg.WriteString(errItem.Translate(trans))
	}
	return msg.String()
}
