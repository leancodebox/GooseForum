package controllers

import (
	"fmt"
	"math"
	"time"

	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"
	"github.com/samber/lo"

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
	responseData := map[string]any{
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

// SearchPage 搜索页面 V3
func SearchPage(c *gin.Context) {
	query := c.Query("q")
	page := cast.ToInt(c.DefaultQuery("page", "1"))
	pageSize := 10 // 页面显示固定每页10条

	// 构建模板数据
	data := SearchData{
		CommonDataVo: GetCommonData(c),
		Query:        query,
		CurrentPage:  page,
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
			data.SearchResponse = result
			ids := lo.Map(result.Results, func(t searchservice.SearchResult, _ int) uint64 {
				return t.ID
			})
			articleEntityList := articles.GetByIds(ids)
			// 收集所有需要查询的用户ID (作者 + 回复者)
			userIds := lo.FlatMap(articleEntityList, func(t *articles.SmallEntity, _ int) []uint64 {
				return append([]uint64{t.UserId}, lo.Map(t.GetPosters(), func(p articles.Poster, _ int) uint64 {
					return p.UserID
				})...)
			})
			userIds = lo.Uniq(userIds)

			userMap := users.GetMapByIds(userIds)
			categoryMap := hotdataserve.ArticleCategoryMap()
			articleList := lo.Map(articleEntityList, func(t *articles.SmallEntity, _ int) *vo.ArticlesSimpleVo {
				categoryNames := lo.Map(t.CategoryId, func(item uint64, _ int) string {
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

				// 构建 Posters 列表
				posters := lo.FilterMap(t.GetPosters(), func(p articles.Poster, _ int) (vo.PosterVo, bool) {
					if u, ok := userMap[p.UserID]; ok {
						return vo.PosterVo{
							Id:        u.Id,
							Username:  u.Username,
							AvatarUrl: u.GetWebAvatarUrl(),
						}, true
					}
					return vo.PosterVo{}, false
				})

				return &vo.ArticlesSimpleVo{
					Id:             t.Id,
					Title:          t.Title,
					Description:    t.Description,
					LastUpdateTime: t.UpdatedAt.Format(time.DateTime),
					Username:       username,
					AuthorId:       t.UserId,
					AvatarUrl:      avatarUrl,
					ViewCount:      t.ViewCount,
					CommentCount:   t.ReplyCount,
					Categories:     categoryNames,
					CategoriesId:   t.CategoryId,
					Type:           t.Type,
					TypeStr:        hotdataserve.GetArticlesTypeName(int(t.Type)),
					Posters:        posters,
				}
			})

			data.ArticleList = articleList
			// 计算分页信息
			totalPages := int(math.Ceil(float64(result.Total) / float64(pageSize)))
			data.TotalPages = totalPages
			// 生成页码列表（显示当前页前后2页）
			data.PageNumbers = lo.RangeFrom(max(page-2, 1), min(page+2, totalPages)-max(page-2, 1)+1)
		}
	}

	pageMeta := viewrender.NewPageMetaBuilder().
		SetTitle(fmt.Sprintf("%v - Search Results - GooseForum", query)).
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build()

	// 渲染模板
	viewrender.SafeRender(c, "search.gohtml", data, pageMeta)
}
