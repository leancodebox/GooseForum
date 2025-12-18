package routes

import (
	"net/http"

	"github.com/leancodebox/GooseForum/app/bundles/validate"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"

	"github.com/gin-gonic/gin"
)

// ginUpNP  支持空参数
func ginUpNP(action func() component.Response) func(c *gin.Context) {
	return func(c *gin.Context) {
		response := action()
		c.JSON(response.Code, response.Data)
	}
}

func UpButterReq[T any](action func(ctx component.BetterRequest[T]) component.Response) func(c *gin.Context) {
	return func(c *gin.Context) {
		bindAndExecute(c, c.ShouldBind, action, false)
	}
}

// UpJsonReq 强制 JSON 绑定
func UpJsonReq[T any](action func(ctx component.BetterRequest[T]) component.Response) func(c *gin.Context) {
	return func(c *gin.Context) {
		bindAndExecute(c, c.ShouldBindJSON, action, true)
	}
}

// UpQueryReq 强制 Query 绑定
func UpQueryReq[T any](action func(ctx component.BetterRequest[T]) component.Response) func(c *gin.Context) {
	return func(c *gin.Context) {
		bindAndExecute(c, c.ShouldBindQuery, action, true)
	}
}

// UpUriReq 强制 URI 绑定 (Path params)
func UpUriReq[T any](action func(ctx component.BetterRequest[T]) component.Response) func(c *gin.Context) {
	return func(c *gin.Context) {
		bindAndExecute(c, c.ShouldBindUri, action, true)
	}
}

// UpFormReq 强制 Form 绑定
func UpFormReq[T any](action func(ctx component.BetterRequest[T]) component.Response) func(c *gin.Context) {
	return func(c *gin.Context) {
		// c.ShouldBind 在 Content-Type 为 multipart/form-data 或 application/x-www-form-urlencoded 时会自动处理
		// 但为了语义明确，我们也可以使用 ShouldBindWith，不过 ShouldBind 足够智能
		bindAndExecute(c, c.ShouldBind, action, true)
	}
}

// bindAndExecute 内部通用处理逻辑
func bindAndExecute[T any](c *gin.Context, binder func(any) error, action func(component.BetterRequest[T]) component.Response, strict bool) {
	userId := c.GetUint64("userId")
	var params T

	// 执行绑定
	if err := binder(&params); err != nil {
		// 如果是严格模式，绑定失败直接返回 400
		if strict {
			c.JSON(http.StatusBadRequest, component.FailData("参数解析失败: "+err.Error()))
			return
		}
		// 非严格模式下（UpButterReq），保持原有逻辑，忽略绑定错误，依赖后续的 validate
	}

	c.Set("requestData", params)
	if err := validate.Valid(params); err != nil {
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
