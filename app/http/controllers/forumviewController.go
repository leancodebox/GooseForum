package controllers

import (
	"github.com/gin-gonic/gin"
	array "github.com/leancodebox/GooseForum/app/bundles/collectionopt"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/http/controllers/viewrender"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategoryRs"
	"github.com/leancodebox/GooseForum/app/models/forum/articleLike"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"
	"github.com/spf13/cast"
	"html/template"
	"strings"
	"time"
)

func Home(c *gin.Context) {
	last := getLatestArticles()
	viewrender.Render(c, "index.gohtml", map[string]any{
		"IsProduction":        setting.IsProduction(),
		"CanonicalHref":       buildCanonicalHref(c),
		"User":                GetLoginUser(c),
		"Title":               "GooseForum",
		"ArticleCategoryList": articleCategory.Label(),
		//"FeaturedArticles":    articlesSmallEntity2Dto(getRecommendedArticles()), //回复最多的文章
		"Description":    "GooseForum's home",
		"LatestArticles": articlesSmallEntity2Dto(last), // 最新的文章
		"Stats":          GetSiteStatisticsData(),
	})
}

func LoginView(c *gin.Context) {
	viewrender.Render(c, "login-vue.gohtml", map[string]any{
		"IsProduction": setting.IsProduction(),
		"User":         GetLoginUser(c),
		"Title":        "登录/注册 - GooseForum",
	})
}

func PostDetail(c *gin.Context) {
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
	entity := articles.Get(req.Id)
	if entity.Id == 0 {
		errorPage(c, "文章不存在", "文章不存在")
		return
	}
	replyEntities := reply.GetByMaxIdPage(req.Id, req.MaxCommentId, boundPageSizeWithRange(req.PageSize, 10, 100))
	userIds := array.Map(replyEntities, func(item reply.Entity) uint64 {
		return item.UserId
	})
	userIds = append(userIds, entity.UserId)
	userMap := users.GetMapByIds(userIds)
	author := "陶渊明"
	avatarUrl := urlconfig.GetDefaultAvatar()
	authorUserInfo := users.Entity{}
	if user, ok := userMap[entity.UserId]; ok {
		author = user.Username
		avatarUrl = user.GetWebAvatarUrl()
		authorUserInfo = *user
	}
	replyList := array.Map(replyEntities, func(item reply.Entity) ReplyDto {
		username := "陶渊明"
		userAvatarUrl := urlconfig.GetDefaultAvatar()
		if user, ok := userMap[item.UserId]; ok {
			username = user.Username
			userAvatarUrl = user.GetWebAvatarUrl()
		}
		// 获取被回复评论的用户名
		replyToUsername := ""
		if item.ReplyId > 0 {
			if replyTo := reply.Get(item.ReplyId); replyTo.Id > 0 {
				if replyUser, ok := userMap[replyTo.UserId]; ok {
					replyToUsername = replyUser.Username
				}
			}
		}
		return ReplyDto{
			Id:              item.Id,
			ArticleId:       item.ArticleId,
			UserId:          item.UserId,
			UserAvatarUrl:   userAvatarUrl,
			Username:        username,
			Content:         item.Content,
			CreateTime:      item.CreatedAt.Format(time.DateTime),
			ReplyToUsername: replyToUsername,
		}
	})
	articles.IncrementView(entity)
	// 复用现有的数据获取逻辑
	authorId := entity.UserId

	if entity.RenderedVersion < markdown2html.GetVersion() || entity.RenderedHTML == "" {
		mdInfo := markdown2html.MarkdownToHTML(entity.Content)
		entity.RenderedHTML = mdInfo
		entity.RenderedVersion = markdown2html.GetVersion()
		articles.SaveNoUpdate(&entity)
	}

	authorArticles, _ := articles.GetRecommendedArticlesByAuthorId(cast.ToUint64(authorId), 5)
	acMap := articleCategoryMapList([]uint64{id})
	iLike := false
	loginUser := GetLoginUser(c)
	if loginUser.UserId != 0 {
		iLike = articleLike.GetByArticleId(loginUser.UserId, entity.Id).Status == 1
	}
	// 构建模板数据
	viewrender.Render(c, "detail.gohtml", map[string]any{
		"IsProduction":        setting.IsProduction(),
		"ArticleId":           id,
		"AuthorId":            authorId,
		"Title":               entity.Title + " - GooseForum",
		"Description":         TakeUpTo64Chars(entity.Content),
		"Year":                time.Now().Year(),
		"ArticleTitle":        entity.Title,
		"ArticleContent":      template.HTML(entity.RenderedHTML),
		"LikeCount":           entity.LikeCount,
		"ViewCount":           entity.ViewCount,
		"CreateTime":          entity.CreatedAt.Format(time.DateTime),
		"ILike":               iLike,
		"Username":            author,
		"CommentList":         replyList,
		"AvatarUrl":           avatarUrl,
		"User":                loginUser,
		"CanonicalHref":       buildCanonicalHref(c),
		"AuthorArticles":      authorArticles,
		"ArticleCategory":     acMap[id],
		"Keywords":            strings.Join(acMap[id], ","),
		"Website":             authorUserInfo.Website,
		"WebsiteName":         authorUserInfo.WebsiteName,
		"ExternalInformation": authorUserInfo.GetExternalInformation(),
	})
}

func PostV2(c *gin.Context) {
	filters := c.DefaultQuery("filters", "")

	categories := array.Filter(array.Map(strings.Split(filters, "-"), func(t string) int {
		return cast.ToInt(t)
	}), func(i int) bool {
		return i > 0
	})
	param := GetArticlesPageRequest{
		Page:       cast.ToInt(c.DefaultQuery("page", "1")),
		PageSize:   cast.ToInt(c.DefaultQuery("pageSize", "20")),
		Search:     c.Query("search"),
		Categories: categories,
	}
	pageData := articles.Page[articles.SmallEntity](
		articles.PageQuery{
			Page:         max(param.Page, 1),
			PageSize:     param.PageSize,
			FilterStatus: true,
			Categories:   param.Categories,
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

	articleList := array.Map(pageData.Data, func(t articles.SmallEntity) ArticlesSimpleDto {
		categoryNames := array.Map(categoriesGroup[t.Id], func(rs *articleCategoryRs.Entity) string {
			if category, ok := categoryMap[rs.ArticleCategoryId]; ok {
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
		return ArticlesSimpleDto{
			Id:             t.Id,
			Title:          t.Title,
			LastUpdateTime: t.UpdatedAt.Format(time.DateTime),
			Username:       username,
			AuthorId:       t.UserId,
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
	// 计算总页数
	totalPages := (cast.ToInt(pageData.Total) + param.PageSize - 1) / param.PageSize
	articleCategoryList := articleCategory.Label()
	pagination := []PageButton{}
	start := max(pageData.Page-3, 1)
	for i := 1; i <= 7; i++ {
		pagination = append(pagination, PageButton{Index: i, Page: start})
		start += 1
	}

	viewrender.Render(c, "list.gohtml", map[string]any{
		"IsProduction":        setting.IsProduction(),
		"Title":               "GooseForum",
		"Description":         "知无不言,言无不尽",
		"Year":                time.Now().Year(),
		"ArticleList":         articleList,
		"Page":                pageData.Page,
		"PageSize":            param.PageSize,
		"Total":               pageData.Total,
		"TotalPages":          totalPages,
		"PrevPage":            max(pageData.Page-1, 1),
		"NextPage":            min(max(pageData.Page, 1)+1, totalPages),
		"User":                GetLoginUser(c),
		"ArticleCategoryList": articleCategoryList,
		"RecommendedArticles": getRecommendedArticles(),
		"CanonicalHref":       buildCanonicalHref(c),
		"Filters":             filters,
		"FilterIds":           categories,
		"NoFilter":            len(categories) == 0,
		"Pagination":          pagination,
	})
}

func User(c *gin.Context) {
	id := cast.ToUint64(c.Param("id"))
	showUser := GetUserShowByUserId(id)
	if showUser.UserId == 0 {
		errorPage(c, "用户不存在", "用户不存在")
		return
	}
	last, _ := articles.GetLatestArticlesByUserId(id, 5)
	viewrender.Render(c, "user.gohtml", map[string]any{
		"IsProduction": setting.IsProduction(),
		"Articles":     articlesSmallEntity2Dto(last),
		"Author":       showUser,
		"User":         GetLoginUser(c),
		"Title":        showUser.Username + " - GooseForum",
		"Description":  showUser.Username + " 的个人简介 ",
	})
}

func About(c *gin.Context) {
	viewrender.Render(c, "about.gohtml", map[string]any{
		"IsProduction": setting.IsProduction(),
		"User":         GetLoginUser(c),
		"Title":        "关于 - GooseForum",
		"Description":  "GooseForum's about",
	})
}

func SponsorsView(c *gin.Context) {
	viewrender.Render(c, "sponsors.gohtml", map[string]any{
		"IsProduction": setting.IsProduction(),
		"User":         GetLoginUser(c),
		"Title":        "赞助商 - GooseForum",
		"Description":  "GooseForum's sponsors",
	})
}

func LinksView(c *gin.Context) {
	configEntity := pageConfig.GetByPageType(FriendShipLinks)
	res := jsonopt.Decode[[]FriendLinksGroup](configEntity.Config)
	viewrender.Render(c, "links.gohtml", map[string]any{
		"IsProduction":     setting.IsProduction(),
		"User":             GetLoginUser(c),
		"Title":            "友情链接 - GooseForum",
		"FriendLinksGroup": res,
		"Description":      "GooseForum's links",
	})
}

func Profile(c *gin.Context) {
	viewrender.Render(c, "profile.gohtml", map[string]any{
		"IsProduction": setting.IsProduction(),
		"User":         GetLoginUser(c),
		"Title":        "个人中心 - GooseForum",
	})
}

func Publish(c *gin.Context) {
	viewrender.Render(c, "publish.gohtml", map[string]any{
		"IsProduction": setting.IsProduction(),
		"User":         GetLoginUser(c),
		"Title":        "发布中心 - GooseForum",
	})
}

func Notifications(c *gin.Context) {
	viewrender.Render(c, "notifications.gohtml", map[string]any{
		"IsProduction": setting.IsProduction(),
		"User":         GetLoginUser(c),
		"Title":        "通知中心 - GooseForum",
	})
}

func SubmitLink(c *gin.Context) {
	viewrender.Render(c, "submit-link.gohtml", map[string]any{
		"IsProduction": setting.IsProduction(),
		"User":         GetLoginUser(c),
		"Title":        "友情链接申请 - GooseForum",
		"Description":  "友情链接申请",
	})
}
