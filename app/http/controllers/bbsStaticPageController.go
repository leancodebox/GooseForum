package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/spf13/cast"
	"html/template"
	"net/http"
	"time"
)

// 添加新的服务端渲染的控制器方法
func markdownToHTML(md string) template.HTML {
	output := markdown.ToHTML([]byte(md), nil, nil)
	return template.HTML(output)
}

// RenderArticlesPage 渲染文章列表页面
func RenderArticlesPage(c *gin.Context) {
	param := GetArticlesPageRequest{
		Page:     cast.ToInt(c.DefaultQuery("page", "1")),
		PageSize: cast.ToInt(c.DefaultQuery("pageSize", "20")),
		Search:   c.Query("search"),
	}

	// 复用现有的数据获取逻辑
	response := GetArticlesPage(param)
	if response.Code != 200 {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"message": "获取文章列表失败",
		})
		return
	}
	result := response.Data.Result.(component.Page[ArticlesSimpleDto])
	// 计算总页数
	totalPages := (cast.ToInt(result.Total) + param.PageSize - 1) / param.PageSize

	// 构建模板数据
	templateData := gin.H{
		"title":       "文章列表",
		"description": "GooseForum的文章列表页面",
		"year":        time.Now().Year(),
		"Data":        result.List,
		"Page":        result.Page,
		"PageSize":    param.PageSize,
		"Total":       result.Total,
		"TotalPages":  totalPages,
		"PrevPage":    max(result.Page-1, 1),
		"NextPage":    min(max(result.Page, 1)+1, totalPages),
	}
	c.HTML(http.StatusOK, "list.gohtml", templateData)
}

// RenderArticleDetail 渲染文章详情页面
func RenderArticleDetail(c *gin.Context) {
	id := cast.ToUint64(c.Param("id"))
	if id == 0 {
		c.HTML(http.StatusNotFound, "error.gohtml", gin.H{
			"title":   "页面不存在",
			"message": "文章不存在",
			"year":    time.Now().Year(),
		})
		return
	}

	req := GetArticlesDetailRequest{
		Id:           id,
		MaxCommentId: 0,
		PageSize:     50,
	}

	// 复用现有的数据获取逻辑
	response := GetArticlesDetail(req)
	data, _ := response.Data.Result.(GetArticlesDetailRequest)
	if response.Code != 200 || data.Id == 0 {
		c.HTML(http.StatusNotFound,
			"error.gohtml",
			gin.H{
				"title":   "页面不存在",
				"message": "文章不存在",
				"year":    time.Now().Year(),
			})
		return
	}
	result := response.Data.Result.(map[string]any)

	// 构建模板数据
	templateData := gin.H{
		"articleId":      id,
		"title":          cast.ToString(result["articleTitle"]),
		"description":    TakeUpTo64Chars(cast.ToString(result["articleContent"])),
		"year":           time.Now().Year(),
		"articleTitle":   cast.ToString(result["articleTitle"]),
		"articleContent": markdownToHTML(cast.ToString(result["articleContent"])),
		"username":       cast.ToString(result["username"]),
		"commentList":    result["commentList"],
	}

	c.HTML(http.StatusOK, "detail.gohtml", templateData)
}
