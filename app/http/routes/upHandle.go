package routes

import (
	"github.com/leancodebox/GooseForum/app/bundles/validate"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ginUpP  支持params 参数
func ginUpP[T any](action func(request T) component.Response) func(c *gin.Context) {
	return func(c *gin.Context) {
		var params T
		_ = c.ShouldBind(&params)
		c.Set("requestData", params)
		err := validate.Valid(params)
		if err != nil {
			c.JSON(http.StatusBadRequest, component.FailData(validate.FormatError(err)))
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
		userId := c.GetUint64("userId")
		var params T
		_ = c.ShouldBind(&params)
		c.Set("requestData", params)
		err := validate.Valid(params)
		if err != nil {
			c.JSON(http.StatusBadRequest, component.FailData(validate.FormatError(err)))
			return
		}
		response := action(component.BetterRequest[T]{
			Params:     params,
			UserId:     userId,
			GinContext: c,
		})
		c.JSON(response.Code, response.Data)
	}
}
