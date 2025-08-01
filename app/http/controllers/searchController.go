package controllers

import (
	"fmt"
	array "github.com/leancodebox/GooseForum/app/bundles/collectionopt"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"
	"math"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/controllers/viewrender"
	"github.com/leancodebox/GooseForum/app/service/searchservice"
	"github.com/spf13/cast"
)

// SearchArticlesRequest 搜索文章请求结构
type SearchArticlesRequest struct {
	Query      string   `json:"query" form:"query" validate:"required,min=1,max=100"` // 搜索关键词
	Categories []uint64 `json:"categories" form:"categories"`                         // 分类ID列表
	Page       int      `json:"page" form:"page"`                                     // 页码，从1开始
	PageSize   int      `json:"pageSize" form:"pageSize"`                             // 每页数量
}

// SearchArticles 搜索文章接口
// 通过名字和类别查询文章，先查Meilisearch索引，再查文章表获取具体数据
func SearchArticles(req component.BetterRequest[SearchArticlesRequest]) component.Response {
	// 参数验证
	if req.Params.Query == "" {
		return component.FailResponse("搜索关键词不能为空")
	}

	// 设置默认值
	if req.Params.Page <= 0 {
		req.Params.Page = 1
	}
	if req.Params.PageSize <= 0 {
		req.Params.PageSize = 20
	}
	if req.Params.PageSize > 100 {
		req.Params.PageSize = 100
	}

	// 计算偏移量
	offset := (req.Params.Page - 1) * req.Params.PageSize

	// 构建搜索请求
	searchReq := searchservice.SearchRequest{
		Query:      req.Params.Query,
		Categories: req.Params.Categories,
		Limit:      req.Params.PageSize,
		Offset:     offset,
	}

	// 执行搜索
	result, err := searchservice.SearchArticles(searchReq)
	if err != nil {
		return component.FailResponse("搜索失败: " + err.Error())
	}

	// 构建响应数据
	responseData := map[string]interface{}{
		"results":  result.Results,
		"total":    result.Total,
		"page":     req.Params.Page,
		"pageSize": req.Params.PageSize,
		"query":    req.Params.Query,
	}

	if len(req.Params.Categories) > 0 {
		responseData["categories"] = req.Params.Categories
	}

	return component.SuccessResponse(responseData)
}

// SearchPage 搜索页面
func SearchPage(c *gin.Context) {
	query := c.Query("q")
	page := cast.ToInt(c.DefaultQuery("page", "1"))
	pageSize := 10 // 页面显示固定每页10条

	// 构建模板数据
	templateData := map[string]any{
		"Query":       query,
		"CurrentPage": page,
		"ShowSearch":  true, // 控制导航栏搜索框显示
	}

	// 如果有搜索关键词，执行搜索
	if query != "" {
		// 计算偏移量
		offset := (page - 1) * pageSize

		// 构建搜索请求
		searchReq := searchservice.SearchRequest{
			Query:  query,
			Limit:  pageSize,
			Offset: offset,
		}

		// 执行搜索
		result, err := searchservice.SearchArticles(searchReq)
		if err == nil {

			templateData["SearchResponse"] = result
			ids := array.Map(result.Results, func(t searchservice.SearchResult) uint64 {
				return t.ID
			})
			articleEntityList := articles.GetByIds(ids)
			userIds := array.Map(articleEntityList, func(t *articles.SmallEntity) uint64 {
				return t.UserId
			})
			userMap := users.GetMapByIds(userIds)

			categoryMap := hotdataserve.ArticleCategoryMap()

			articleList := array.Map(articleEntityList, func(t *articles.SmallEntity) vo.ArticlesSimpleDto {
				categoryNames := array.Map(t.CategoryId, func(item uint64) string {
					if category, ok := categoryMap[item]; ok {
						return category.Category
					}
					return ""
				})
				username := ""
				avatarUrl := urlconfig.GetDefaultAvatar()
				if user, ok := userMap[t.UserId]; ok {
					username = user.Username
					avatarUrl = user.GetWebAvatarUrl()
				}
				return vo.ArticlesSimpleDto{
					Id:             t.Id,
					Title:          t.Title,
					LastUpdateTime: t.UpdatedAt.Format(time.DateTime),
					Username:       username,
					AuthorId:       t.UserId,
					AvatarUrl:      avatarUrl,
					ViewCount:      t.ViewCount,
					CommentCount:   t.ReplyCount,
					Category:       array.FirstOr(categoryNames, "未分类"),
					Categories:     categoryNames,
					CategoriesId:   t.CategoryId,
					Type:           t.Type,
					TypeStr:        hotdataserve.GetArticlesTypeName(int(t.Type)),
				}
			})

			templateData["ArticleList"] = articleList
			// 计算分页信息
			totalPages := int(math.Ceil(float64(result.Total) / float64(pageSize)))
			templateData["TotalPages"] = totalPages

			// 生成页码列表（显示当前页前后2页）
			var pageNumbers []int
			start := page - 2
			if start < 1 {
				start = 1
			}
			end := page + 2
			if end > totalPages {
				end = totalPages
			}
			for i := start; i <= end; i++ {
				pageNumbers = append(pageNumbers, i)
			}
			templateData["PageNumbers"] = pageNumbers
		}
	}

	templateData["PageMeta"] = viewrender.NewPageMetaBuilder().
		SetTitle(fmt.Sprintf("%v 的搜索结果", query)).
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build()
	// 渲染模板
	viewrender.Render(c, "search.gohtml",
		templateData,
	)
}
