package controllers

import (
	"bytes"
	"github.com/gin-gonic/gin"
	array "github.com/leancodebox/GooseForum/app/bundles/goose/collectionopt"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategoryRs"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/spf13/cast"
	"github.com/yuin/goldmark"
	"html/template"
	"log/slog"
	"net/http"
	"time"
)

// 添加新的服务端渲染的控制器方法
func markdownToHTML(markdown string) template.HTML {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(markdown), &buf); err != nil {
		slog.Error("转化失败", "err", err)
	}
	return template.HTML(buf.String())
}
func RenderIndex(c *gin.Context) {
	pageData := articles.Page[articles.SmallEntity](
		articles.PageQuery{
			Page:         1,
			PageSize:     6,
			FilterStatus: true,
		})
	userIds := array.Map(pageData.Data, func(t articles.SmallEntity) uint64 {
		return t.UserId
	})
	userMap := users.GetMapByIds(userIds)

	//获取文章的分类信息
	articleIds := array.Map(pageData.Data, func(t articles.SmallEntity) uint64 {
		return t.Id
	})
	categoryRs := articleCategoryRs.GetByArticleIdsEffective(articleIds)
	categoryIds := array.Map(categoryRs, func(t *articleCategoryRs.Entity) uint64 {
		return t.ArticleCategoryId
	})
	categoryMap := articleCategory.GetMapByIds(categoryIds)
	// 获取文章的分类和标签
	categoriesGroup := array.GroupBy(categoryRs, func(rs *articleCategoryRs.Entity) uint64 {
		return rs.ArticleId
	})

	res := array.Map(pageData.Data, func(t articles.SmallEntity) ArticlesSimpleDto {
		categoryNames := array.Map(categoriesGroup[t.Id], func(rs *articleCategoryRs.Entity) string {
			if category, ok := categoryMap[rs.ArticleCategoryId]; ok {
				return category.Category
			}
			return ""
		})
		username := ""
		avatarUrl := ""
		if user, ok := userMap[t.UserId]; ok {
			username = user.Username
			if user.AvatarUrl != "" {
				avatarUrl = component.FilePath(user.AvatarUrl)
			}
		}
		return ArticlesSimpleDto{
			Id:             t.Id,
			Title:          t.Title,
			LastUpdateTime: t.UpdatedAt.Format("2006-01-02 15:04:05"),
			Username:       username,
			AvatarUrl:      avatarUrl,
			ViewCount:      t.ViewCount,
			CommentCount:   t.ReplyCount,
			Category:       FirstOr(categoryNames, "未分类"),
			Categories:     categoryNames,
			CategoriesId: array.Map(categoriesGroup[t.Id], func(rs *articleCategoryRs.Entity) uint64 {
				return rs.ArticleCategoryId
			}),
			Type:    t.Type,
			TypeStr: articlesTypeMap[int(t.Type)].Name,
		}
	})

	templateData := gin.H{
		"FeaturedArticles": res,
		"LatestArticles":   res,
		"Stats":            GetSiteStatisticsData(),
	}
	c.HTML(http.StatusOK, "home.gohtml", templateData)
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
		errorPage(c, "获取文章列表失败", "获取文章列表失败")
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
		errorPage(c, "页面不存在", "页面不存在")
		return
	}

	req := GetArticlesDetailRequest{
		Id:           id,
		MaxCommentId: 0,
		PageSize:     50,
	}

	// 复用现有的数据获取逻辑
	response := GetArticlesDetail(req)
	result := response.Data.Result.(map[string]any)

	if _, ok := result["id"]; !ok {
		errorPage(c, "页面不存在", "文章不存在")
		return
	}
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

func errorPage(c *gin.Context, title, message string) {
	c.HTML(http.StatusNotFound, "error.gohtml", gin.H{
		"title":   title,
		"message": message,
		"year":    time.Now().Year(),
	})
}

func LoginPage(c *gin.Context) {
	c.HTML(http.StatusNotFound, "login.gohtml", gin.H{"title": "login.gohtml"})
}
func Notifications(c *gin.Context) {
	c.HTML(http.StatusNotFound, "notifications.gohtml", gin.H{"title": "notifications.gohtml"})
}
func PostEdit(c *gin.Context) {
	c.HTML(http.StatusNotFound, "post_edit.gohtml", gin.H{"title": "post_edit.gohtml"})
}
func UserProfile(c *gin.Context) {
	c.HTML(http.StatusNotFound, "user_profile.gohtml", gin.H{"title": "user_profile.gohtml"})
}
