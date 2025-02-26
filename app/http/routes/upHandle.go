package routes

import (
	"bytes"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/spf13/cast"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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

// ginUpP  支持params 参数
func ginUpP[T any](action func(request T) component.Response) func(c *gin.Context) {
	return func(c *gin.Context) {
		var params T
		_ = c.ShouldBind(&params)
		c.Set("requestData", params)
		err := validate.Struct(params)
		if err != nil {
			c.JSON(http.StatusBadRequest, component.FailData(formatError(err)))
			return
		}
		response := action(params)
		c.JSON(response.Code, response.Data)
	}
}

// ginUpNP  支持空参数
func ginUpNP(action func() component.Response) func(c *gin.Context) {
	return func(c *gin.Context) {
		response := action()
		c.JSON(response.Code, response.Data)
	}
}

func UpButterReq[T any](action func(ctx component.BetterRequest[T]) component.Response) func(c *gin.Context) {
	return func(c *gin.Context) {
		userIdData, _ := c.Get("userId")
		userId := cast.ToUint64(userIdData)
		var params T
		_ = c.ShouldBind(&params)
		c.Set("requestData", params)
		err := validate.Struct(params)
		if err != nil {
			c.JSON(http.StatusBadRequest, component.FailData(formatError(err)))
			return
		}
		response := action(component.BetterRequest[T]{
			Params: params,
			UserId: userId,
		})
		c.JSON(response.Code, response.Data)
	}
}

func formatError(err error) string {
	var msg bytes.Buffer
	for _, errItem := range err.(validator.ValidationErrors) {
		// 输出中文错误信息
		msg.WriteString(errItem.Translate(trans))
	}
	return msg.String()
}
