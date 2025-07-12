package controllers

import (
	_ "embed"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	array "github.com/leancodebox/GooseForum/app/bundles/collectionopt"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/http/controllers/viewrender"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategoryRs"
	"github.com/leancodebox/GooseForum/app/models/forum/articleLike"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/userFollow"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"
	"github.com/spf13/cast"
)

func Home(c *gin.Context) {
	last := getLatestArticles()
	viewrender.Render(c, "index.gohtml", map[string]any{
		"IsProduction":        setting.IsProduction(),
		"CanonicalHref":       buildCanonicalHref(c),
		"User":                GetLoginUser(c),
		"Title":               "GooseForum - 自由漫谈的江湖茶馆",
		"ArticleCategoryList": articleCategoryLabel(),
		"Description":         "GooseForum's home",
		"LatestArticles":      articlesSmallEntity2Dto(last), // 最新的文章
		"Stats":               GetSiteStatisticsData(),
		"RecommendedArticles": getRecommendedArticles(),
		"GooseForumInfo":      GetGooseForumInfo(),
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
	replyEntities := reply.GetByArticleId(entity.Id)
	userIds := array.Map(replyEntities, func(item *reply.Entity) uint64 {
		return item.UserId
	})
	userIds = append(userIds, entity.UserId)
	userMap := users.GetMapByIds(userIds)
	author := "陶渊明"
	avatarUrl := urlconfig.GetDefaultAvatar()
	authorUserInfo := users.EntityComplete{}
	authorInfoStatistics := userStatistics.Get(entity.UserId)
	if user, ok := userMap[entity.UserId]; ok {
		author = user.Username
		avatarUrl = user.GetWebAvatarUrl()
		authorUserInfo = *user
	}
	replyList := array.Map(replyEntities, func(item *reply.Entity) ReplyDto {
		username := "陶渊明"
		userAvatarUrl := urlconfig.GetDefaultAvatar()
		if user, ok := userMap[item.UserId]; ok {
			username = user.Username
			userAvatarUrl = user.GetWebAvatarUrl()
		}
		// 获取被回复评论的用户名
		replyToUsername := ""
		var replyToUserId uint64 = 0
		if item.ReplyId > 0 {
			if replyTo := reply.Get(item.ReplyId); replyTo.Id > 0 {
				if replyUser, ok := userMap[replyTo.UserId]; ok {
					replyToUsername = replyUser.Username
					replyToUserId = replyTo.UserId
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
			ReplyToUserId:   replyToUserId,
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
	isFollowing := false
	loginUser := GetLoginUser(c)
	if loginUser.UserId != 0 {
		iLike = articleLike.GetByArticleId(loginUser.UserId, entity.Id).Status == 1
		// 检查是否已关注作者
		if loginUser.UserId != entity.UserId {
			followEntity := userFollow.GetByUserId(loginUser.UserId, entity.UserId)
			isFollowing = followEntity.Status == 1
		}
	}
	// 构建模板数据
	viewrender.Render(c, "detail.gohtml", map[string]any{
		"IsProduction":         setting.IsProduction(),
		"ArticleId":            id,
		"AuthorId":             authorId,
		"Title":                entity.Title + " - GooseForum",
		"Description":          entity.Description,
		"OgType":               "article",
		"Year":                 time.Now().Year(),
		"ArticleTitle":         entity.Title,
		"ArticleContent":       template.HTML(entity.RenderedHTML),
		"LikeCount":            entity.LikeCount,
		"ViewCount":            entity.ViewCount,
		"CreateTime":           entity.CreatedAt.Format(time.DateTime),
		"ILike":                iLike,
		"Username":             author,
		"CommentList":          replyList,
		"AvatarUrl":            avatarUrl,
		"User":                 loginUser,
		"CanonicalHref":        buildCanonicalHref(c),
		"AuthorArticles":       authorArticles,
		"ArticleCategory":      acMap[id],
		"Keywords":             strings.Join(acMap[id], ","),
		"Website":              authorUserInfo.Website,
		"WebsiteName":          authorUserInfo.WebsiteName,
		"ExternalInformation":  authorUserInfo.ExternalInformation,
		"Bio":                  authorUserInfo.Bio,
		"Signature":            authorUserInfo.Signature,
		"AuthorInfoStatistics": authorInfoStatistics,
		"IsFollowing":          isFollowing,
		"IsOwnArticle":         loginUser.UserId == entity.UserId,
		"ArticleCategoryList":  articleCategoryLabel(),
		"ArticleJSONLD":        generateArticleJSONLD(c, entity, author),
		"PublishedTime":        entity.CreatedAt.Format(time.RFC3339),
		"ModifiedTime":         entity.UpdatedAt.Format(time.RFC3339),
	})
}

type ForumInfo struct {
	Title        string
	Desc         string
	Independence bool
}

func GetGooseForumInfo() ForumInfo {
	return ForumInfo{
		Title:        "GooseForum",
		Desc:         "🦢 大鹅栖息地 | 自由漫谈的江湖茶馆",
		Independence: false,
	}
}

// generateArticleJSONLD 生成文章的JSON-LD结构化数据
func generateArticleJSONLD(c *gin.Context, entity articles.Entity, author string) template.JS {
	jsonLD := map[string]any{
		"@context": "https://schema.org",
		"@type":    "Article",
		"headline": entity.Title,
		"author": map[string]any{
			"@type": "Person",
			"name":  author,
			"url":   fmt.Sprintf("%s/user/%d", getBaseUri(c), entity.UserId),
		},
		"publisher": map[string]any{
			"@type": "Organization",
			"name":  "GooseForum",
			"url":   getBaseUri(c),
		},
		"datePublished": entity.CreatedAt.Format(time.RFC3339),
		"url":           buildCanonicalHref(c),
		"interactionStatistic": map[string]any{
			"@type":                "InteractionCounter",
			"interactionType":      "https://schema.org/ViewAction",
			"userInteractionCount": entity.ViewCount,
		},
	}
	jsonString := jsonopt.EncodeFormat(jsonLD)
	return template.JS(jsonString)
}

func Post(c *gin.Context) {
	filters := c.DefaultQuery("filters", "")
	categories := array.Filter(array.Map(strings.Split(filters, "-"), func(t string) int {
		return cast.ToInt(t)
	}), func(i int) bool {
		return i > 0
	})
	page := cast.ToInt(c.DefaultQuery("page", "1"))
	pageSize := cast.ToInt(c.DefaultQuery("pageSize", "20"))
	var forumInfo ForumInfo = GetGooseForumInfo()
	pageData := articles.Page[articles.SmallEntity](
		articles.PageQuery{
			Page:         max(page, 1),
			PageSize:     pageSize,
			FilterStatus: true,
			Categories:   categories,
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
	categoryMap := articleCategoryMap()
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
	totalPages := (cast.ToInt(pageData.Total) + pageSize - 1) / pageSize
	articleCategoryList := articleCategoryLabel()
	var pagination []PageButton
	start := max(pageData.Page-3, 1)
	for i := 1; i <= 7; i++ {
		if start > totalPages {
			break
		}
		pagination = append(pagination, PageButton{Index: i, Page: start})
		start += 1
	}

	title := "GooseForum - 自由漫谈的江湖茶馆"
	description := "🦢 大鹅栖息地 | 自由漫谈的江湖茶馆"
	if len(categories) == 1 {
		if category, ok := categoryMap[cast.ToUint64(categories[0])]; ok {
			forumInfo.Independence = true
			forumInfo.Desc = category.Desc
			forumInfo.Title = category.Category

			title = category.Category + " 社区 | GooseForum"
			description = category.Desc
		}
	}

	viewrender.Render(c, "list.gohtml", map[string]any{
		"IsProduction":        setting.IsProduction(),
		"Title":               title,
		"Description":         description,
		"Year":                time.Now().Year(),
		"ArticleList":         articleList,
		"Page":                pageData.Page,
		"PageSize":            pageSize,
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
		"Stats":               GetSiteStatisticsData(),
		"ForumInfo":           forumInfo,
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
	authorInfoStatistics := userStatistics.Get(id)

	// 获取关注和粉丝列表（默认显示前10个）
	followingList, _ := userFollow.GetFollowingList(id, 1, 10)
	followerList, _ := userFollow.GetFollowerList(id, 1, 10)

	// 获取当前登录用户信息
	currentUser := GetLoginUser(c)

	// 检查当前用户是否关注了列表中的用户
	var followingStatusMap map[uint64]bool
	var followerStatusMap map[uint64]bool
	var isFollowingAuthor bool

	if currentUser.UserId > 0 {
		followingStatusMap = make(map[uint64]bool)
		followerStatusMap = make(map[uint64]bool)

		// 检查是否关注了页面作者
		if currentUser.UserId != id {
			authorFollowEntity := userFollow.GetByUserId(currentUser.UserId, id)
			isFollowingAuthor = authorFollowEntity.Id > 0 && authorFollowEntity.Status == 1
		}

		// 检查关注列表中的用户状态
		for _, user := range followingList {
			if user.Id != currentUser.UserId {
				followEntity := userFollow.GetByUserId(currentUser.UserId, user.Id)
				followingStatusMap[user.Id] = followEntity.Id > 0 && followEntity.Status == 1
			}
		}

		// 检查粉丝列表中的用户状态
		for _, user := range followerList {
			if user.Id != currentUser.UserId {
				followEntity := userFollow.GetByUserId(currentUser.UserId, user.Id)
				followerStatusMap[user.Id] = followEntity.Id > 0 && followEntity.Status == 1
			}
		}
	}

	viewrender.Render(c, "user.gohtml", map[string]any{
		"IsProduction":         setting.IsProduction(),
		"Articles":             articlesSmallEntity2Dto(last),
		"ArticlesCount":        articles.GetUserCount(showUser.UserId),
		"Author":               showUser,
		"AuthorInfoStatistics": authorInfoStatistics,
		"FollowingList":        followingList,
		"FollowerList":         followerList,
		"FollowingStatusMap":   followingStatusMap,
		"FollowerStatusMap":    followerStatusMap,
		"IsFollowingAuthor":    isFollowingAuthor,
		"User":                 currentUser,
		"Title":                showUser.Username + " - GooseForum",
		"Description":          showUser.Username + " 的个人简介 ",
		"OgType":               "profile",
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

//go:embed docs/terms-of-service.md
var termsOfServiceMD string

//go:embed docs/privacy-policy.md
var privacyPolicyMD string

// TermsOfService 用户协议页面
func TermsOfService(c *gin.Context) {
	htmlContent := markdown2html.MarkdownToHTML(termsOfServiceMD)
	viewrender.Render(c, "markdown-page.gohtml", map[string]any{
		"IsProduction": setting.IsProduction(),
		"User":         GetLoginUser(c),
		"Title":        "用户协议 - GooseForum",
		"Subtitle":     "Terms of Service",
		"Description":  "GooseForum 用户服务协议",
		"Content":      template.HTML(htmlContent),
	})
}

// PrivacyPolicy 隐私政策页面
func PrivacyPolicy(c *gin.Context) {
	htmlContent := markdown2html.MarkdownToHTML(privacyPolicyMD)
	viewrender.Render(c, "markdown-page.gohtml", map[string]any{
		"IsProduction": setting.IsProduction(),
		"User":         GetLoginUser(c),
		"Title":        "隐私政策 - GooseForum",
		"Subtitle":     "Privacy Policy",
		"Description":  "GooseForum 隐私保护政策",
		"Content":      template.HTML(htmlContent),
	})
}

type LinkStatistics struct {
	Name       string
	Counter    int
	Proportion int
}
type LinkStatisticsInfo struct {
	Community LinkStatistics
	Blog      LinkStatistics
	Tool      LinkStatistics
}

func LinksView(c *gin.Context) {
	configEntity := pageConfig.GetByPageType(pageConfig.FriendShipLinks)
	res := jsonopt.Decode[[]pageConfig.FriendLinksGroup](configEntity.Config)
	totalCounter := 0
	var statistics []LinkStatistics
	for _, group := range res {
		counter := len(group.Links)
		totalCounter += len(group.Links)
		statistics = append(statistics, LinkStatistics{
			Name:       group.Name,
			Counter:    counter,
			Proportion: 0,
		})
	}
	var linkStatisticsInfo LinkStatisticsInfo
	for i, item := range statistics {
		statistics[i].Proportion = item.Counter * 100 / totalCounter
		if item.Name == `tool` {
			linkStatisticsInfo.Tool = statistics[i]
		} else if item.Name == `blog` {
			linkStatisticsInfo.Blog = statistics[i]
		} else if item.Name == `community` {
			linkStatisticsInfo.Community = statistics[i]
		}
	}
	viewrender.Render(c, "links.gohtml", map[string]any{
		"IsProduction":        setting.IsProduction(),
		"User":                GetLoginUser(c),
		"Title":               "友情链接 - GooseForum",
		"FriendLinksGroup":    res,
		"Description":         "GooseForum's links",
		"TotalCounter":        totalCounter,
		"RecommendedArticles": getRecommendedArticles(),
		"LinkStatisticsInfo":  linkStatisticsInfo,
	})
}

func Profile(c *gin.Context) {
	viewrender.Render(c, "profile.gohtml", map[string]any{
		"IsProduction": setting.IsProduction(),
		"User":         GetLoginUser(c),
		"Title":        "个人中心 - GooseForum",
	})
}
func PublishV3(c *gin.Context) {
	viewrender.Render(c, "publish-v3.gohtml", map[string]any{
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

func Admin(c *gin.Context) {
	viewrender.Render(c, "admin.gohtml", map[string]any{
		"IsProduction": setting.IsProduction(),
		"User":         GetLoginUser(c),
		"Title":        "管理 - GooseForum",
		"Description":  "GooseForum's Admin",
	})
}
