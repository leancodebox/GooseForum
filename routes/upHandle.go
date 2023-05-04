package routes

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"io/fs"
	"net/http"
	"path"

	"github.com/spf13/cast"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type resultMap map[string]any

var validate = validator.New()

type fsFunc func(name string) (fs.File, error)

func (f fsFunc) Open(name string) (fs.File, error) {
	return f(name)
}

func upFsHandle(pPath string, fSys fs.FS) fsFunc {
	return func(name string) (fs.File, error) {
		assetPath := path.Join(pPath, name)
		// If we can't find the asset, fs can handle the error
		file, err := fSys.Open(assetPath)
		if err != nil {
			fmt.Println(err, "出错了")
			return nil, err
		}
		return file, err
	}
}

func PFilSystem(pPath string, fSys fs.FS) http.FileSystem {
	return http.FS(upFsHandle(pPath, fSys))
}

// ginUpP  支持params 参数
func ginUpP[T any](action func(request T) component.Response) func(c *gin.Context) {
	return func(c *gin.Context) {
		var params T
		_ = c.ShouldBind(&params)
		err := validate.Struct(params)
		if err != nil {
			c.JSON(http.StatusBadRequest, component.DataMap{
				"msg": err.Error(),
			})
			return
		}
		response := action(params)
		c.JSON(response.Code, response.Data)
	}
}

// ginUpP  支持params 参数
func ginUpJP[T any](action func(request T) component.Response) func(c *gin.Context) {
	return func(c *gin.Context) {
		var params T
		_ = c.BindJSON(&params)
		err := validate.Struct(params)
		if err != nil {
			c.JSON(http.StatusBadRequest, component.DataMap{
				"msg": err.Error(),
			})
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
		err := validate.Struct(params)
		if err != nil {
			c.JSON(http.StatusBadRequest, resultMap{
				"msg": err.Error(),
			})
		}
		response := action(component.BetterRequest[T]{
			Params: params,
			UserId: userId,
		})
		c.JSON(response.Code, response.Data)
	}
}

//// 普通无参数
//// 普通有参数
//// 普通用 request
//// 普通无 request
//// 登录 无参数
//// 登录 有参数
//
//type unParams func() component.Response
//type paramsType[T any] func(request T) component.Response
//
//func UpButterReqV2[T any](varParams any) func(c *gin.Context) {
//	switch action := varParams.(type) {
//	case unParams:
//		return func(c *gin.Context) {
//			response := action()
//			c.JSON(response.Code, response.Data)
//		}
//	case paramsType[T]:
//		return func(c *gin.Context) {
//			var params T
//			_ = c.BindJSON(&params)
//			err := validate.Struct(params)
//			if err != nil {
//				c.JSON(http.StatusBadRequest, component.DataMap{
//					"msg": err.Error(),
//				})
//				return
//			}
//			response := action(params)
//			c.JSON(response.Code, response.Data)
//		}
//	default:
//		return func(c *gin.Context) {
//			c.JSON(http.StatusInternalServerError, "handle不规范")
//		}
//	}
//}
//
//func ginUpSuper[R any, T func() component.Response | func(Request R) component.Response](action T) func(c *gin.Context) {
//	return func(actionItem any) func(c *gin.Context) {
//		switch op := actionItem.(type) {
//		case func() component.Response:
//			return func(c *gin.Context) {
//				response := op()
//				c.JSON(response.Code, response.Data)
//			}
//		case func(Request R) component.Response:
//			return func(c *gin.Context) {
//				var params R
//				_ = c.ShouldBind(&params)
//				err := validate.Struct(params)
//				if err != nil {
//					c.JSON(http.StatusBadRequest, component.DataMap{
//						"msg": err.Error(),
//					})
//					return
//				}
//				response := op(params)
//				c.JSON(response.Code, response.Data)
//			}
//		default:
//			return func(c *gin.Context) {
//				c.JSON(http.StatusInternalServerError, map[string]string{"msg": "action type un support"})
//			}
//		}
//	}(action)
//}
