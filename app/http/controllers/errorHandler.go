package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/viewrender"
	"net/http"
	"time"
)

// ErrorPageData 错误页面数据结构
type ErrorPageData struct {
	Title      string
	Message    string
	ErrorCode  string
	StatusCode int
}

// RenderErrorPage 渲染错误页面的通用函数
func RenderErrorPage(c *gin.Context, data ErrorPageData) {
	if data.StatusCode == 0 {
		data.StatusCode = http.StatusInternalServerError
	}

	if data.Title == "" {
		switch data.StatusCode {
		case http.StatusNotFound:
			data.Title = "页面不存在"
		case http.StatusForbidden:
			data.Title = "访问被拒绝"
		case http.StatusUnauthorized:
			data.Title = "需要登录"
		case http.StatusInternalServerError:
			data.Title = "服务器错误"
		default:
			data.Title = "出错了"
		}
	}

	if data.Message == "" {
		switch data.StatusCode {
		case http.StatusNotFound:
			data.Message = "抱歉，您访问的页面不存在或已被删除"
		case http.StatusForbidden:
			data.Message = "您没有权限访问此页面"
		case http.StatusUnauthorized:
			data.Message = "请先登录后再访问此页面"
		case http.StatusInternalServerError:
			data.Message = "服务器内部错误，请稍后重试"
		default:
			data.Message = "发生了未知错误"
		}
	}

	c.Status(data.StatusCode)
	viewrender.Render(c, "error.gohtml", map[string]any{
		"User": GetLoginUser(c),
		"PageMeta": viewrender.NewPageMetaBuilder().
			SetTitle(data.Title).
			SetCanonicalURL(buildCanonicalHref(c)).
			Build(),
		"message":   data.Message,
		"errorCode": data.ErrorCode,
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"requestId": c.GetHeader("X-Request-ID"),
	})
}

// NotFoundPage 404页面
func NotFoundPage(c *gin.Context) {
	RenderErrorPage(c, ErrorPageData{
		StatusCode: http.StatusNotFound,
		ErrorCode:  "404",
	})
}

// ForbiddenPage 403页面
func ForbiddenPage(c *gin.Context) {
	RenderErrorPage(c, ErrorPageData{
		StatusCode: http.StatusForbidden,
		ErrorCode:  "403",
	})
}

// UnauthorizedPage 401页面
func UnauthorizedPage(c *gin.Context) {
	RenderErrorPage(c, ErrorPageData{
		StatusCode: http.StatusUnauthorized,
		ErrorCode:  "401",
	})
}

// InternalServerErrorPage 500页面
func InternalServerErrorPage(c *gin.Context) {
	RenderErrorPage(c, ErrorPageData{
		StatusCode: http.StatusInternalServerError,
		ErrorCode:  "500",
	})
}

// CustomErrorPage 自定义错误页面
func CustomErrorPage(c *gin.Context, title, message string) {
	RenderErrorPage(c, ErrorPageData{
		Title:      title,
		Message:    message,
		StatusCode: http.StatusNotFound,
		ErrorCode:  "CUSTOM",
	})
}

func errorPage(c *gin.Context, title, message string) {
	CustomErrorPage(c, title, message)
}
