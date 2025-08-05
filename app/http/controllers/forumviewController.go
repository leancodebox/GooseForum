package controllers

import (
	_ "embed"
	"fmt"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/articleBookmark"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"html/template"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	array "github.com/leancodebox/GooseForum/app/bundles/collectionopt"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/http/controllers/viewrender"
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

type PageButton struct {
	Index int
	Page  int
}

func Home(c *gin.Context) {
	pageMeta := viewrender.NewPageMetaBuilder().
		SetTitle("GooseForum - 自由漫谈的江湖茶馆").
		SetDescription("GooseForum's home").
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build()
	viewrender.Render(c, "index.gohtml", map[string]any{
		"PageMeta":            pageMeta,
		"CanonicalHref":       component.BuildCanonicalHref(c),
		"Title":               "GooseForum - 自由漫谈的江湖茶馆",
		"ArticleCategoryList": hotdataserve.ArticleCategoryLabel(),
		"Description":         "GooseForum's home",
		"LatestArticles":      hotdataserve.GetLatestArticleSimpleDto(), // 最新的文章
		"Stats":               hotdataserve.GetSiteStatisticsData(),
		"RecommendedArticles": hotdataserve.GetRecommendedArticles(),
		"Announcement":        hotdataserve.GetAnnouncementConfigCache(),
		"GooseForumInfo":      GetGooseForumInfo(),
	})
}

func LoginView(c *gin.Context) {
	viewrender.Render(c, "login-vue.gohtml", map[string]any{
		"PageMeta": viewrender.NewPageMetaBuilder().
			SetTitle("登录/注册").
			SetCanonicalURL(component.BuildCanonicalHref(c)).
			Build(),
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

	// 复用现有的数据获取逻辑
	authorId := entity.UserId

	if entity.RenderedVersion < markdown2html.GetVersion() || entity.RenderedHTML == "" {
		mdInfo := markdown2html.MarkdownToHTML(entity.Content)
		entity.RenderedHTML = mdInfo
		entity.RenderedVersion = markdown2html.GetVersion()
		articles.SaveNoUpdate(&entity)
	}

	authorArticles := hotdataserve.GetOrLoad(fmt.Sprintf("authorId:hot:%v", authorId), func() ([]articles.SmallEntity, error) {
		return articles.GetRecommendedArticlesByAuthorId(cast.ToUint64(authorId), 5)
	})

	categoryMap := hotdataserve.ArticleCategoryMap()
	articleCategory := array.Map(entity.CategoryId, func(item uint64) string {
		if cateItem, ok := categoryMap[item]; ok {
			return cateItem.Category
		} else {
			return ""
		}
	})
	iLike := false
	isFollowing := false
	isBookmarked := false
	currentUserId := component.LoginUserId(c)
	if currentUserId != 0 {
		iLike = articleLike.GetByArticleId(currentUserId, entity.Id).Status == 1
		// 检查是否已关注作者
		if currentUserId != entity.UserId {
			followEntity := userFollow.GetByUserId(currentUserId, entity.UserId)
			isFollowing = followEntity.Status == 1
		}
		isBookmarked = articleBookmark.GetByArticleId(currentUserId, entity.Id).Status == 1
	}
	// 构建模板数据
	viewrender.Render(c, "detail.gohtml", map[string]any{
		"PageMeta": viewrender.NewPageMetaBuilder().
			SetArticle(
				entity.Title,
				entity.Description,
				author,
				articleCategory,
				&entity.CreatedAt,
				&entity.UpdatedAt,
			).
			SetCanonicalURL(component.BuildCanonicalHref(c)).
			SetSchemaOrg(generateArticleJSONLD(c, entity, author)).
			Build(),
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
		"CanonicalHref":        component.BuildCanonicalHref(c),
		"AuthorArticles":       authorArticles,
		"ArticleCategory":      articleCategory,
		"Keywords":             strings.Join(articleCategory, ","),
		"Website":              authorUserInfo.Website,
		"WebsiteName":          authorUserInfo.WebsiteName,
		"ExternalInformation":  authorUserInfo.ExternalInformation,
		"Bio":                  authorUserInfo.Bio,
		"Signature":            authorUserInfo.Signature,
		"AuthorInfoStatistics": authorInfoStatistics,
		"IsFollowing":          isFollowing,
		"IsOwnArticle":         currentUserId == entity.UserId,
		"ArticleCategoryList":  hotdataserve.ArticleCategoryLabel(),
		"IsBookmarked":         isBookmarked,
	})
	articles.IncrementView(entity)
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
			"url":   fmt.Sprintf("%s/user/%d", component.GetBaseUri(c), entity.UserId),
		},
		"publisher": map[string]any{
			"@type": "Organization",
			"name":  "GooseForum",
			"url":   component.GetBaseUri(c),
		},
		"datePublished": entity.CreatedAt.Format(time.RFC3339),
		"url":           component.BuildCanonicalHref(c),
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

	categoryMap := hotdataserve.ArticleCategoryMap()

	articleList := array.Map(pageData.Data, func(t articles.SmallEntity) vo.ArticlesSimpleDto {
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
	// 计算总页数
	totalPages := (cast.ToInt(pageData.Total) + pageSize - 1) / pageSize
	articleCategoryList := hotdataserve.ArticleCategoryLabel()
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
		"PageMeta": viewrender.NewPageMetaBuilder().
			SetTitle(title).
			SetDescription(description).
			SetCanonicalURL(component.BuildCanonicalHref(c)).
			Build(),
		"Year":                time.Now().Year(),
		"ArticleList":         articleList,
		"Page":                pageData.Page,
		"PageSize":            pageSize,
		"Total":               pageData.Total,
		"TotalPages":          totalPages,
		"PrevPage":            max(pageData.Page-1, 1),
		"NextPage":            min(max(pageData.Page, 1)+1, totalPages),
		"ArticleCategoryList": articleCategoryList,
		"RecommendedArticles": hotdataserve.GetRecommendedArticles(),
		"CanonicalHref":       component.BuildCanonicalHref(c),
		"Filters":             filters,
		"FilterIds":           categories,
		"NoFilter":            len(categories) == 0,
		"Pagination":          pagination,
		"Stats":               hotdataserve.GetSiteStatisticsData(),
		"ForumInfo":           forumInfo,
	})
}

func User(c *gin.Context) {
	id := cast.ToUint64(c.Param("id"))
	showUser := component.GetUserShowByUserId(id)
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
	currentUserId := component.LoginUserId(c)

	// 检查当前用户是否关注了列表中的用户
	var followingStatusMap map[uint64]bool
	var followerStatusMap map[uint64]bool
	var isFollowingAuthor bool

	if currentUserId > 0 {
		// 收集所有需要查询关注状态的用户ID
		var targetUserIds []uint64

		// 添加页面作者ID（如果不是自己）
		if currentUserId != id {
			targetUserIds = append(targetUserIds, id)
		}

		// 添加关注列表中的用户ID（排除自己）
		for _, user := range followingList {
			if user.Id != currentUserId {
				targetUserIds = append(targetUserIds, user.Id)
			}
		}

		// 添加粉丝列表中的用户ID（排除自己）
		for _, user := range followerList {
			if user.Id != currentUserId {
				targetUserIds = append(targetUserIds, user.Id)
			}
		}

		// 一次性批量查询所有关注状态
		allFollowStatusMap := userFollow.GetFollowStatusMap(currentUserId, targetUserIds)

		// 设置页面作者的关注状态
		if currentUserId != id {
			isFollowingAuthor = allFollowStatusMap[id]
		}

		// 构建关注列表的状态映射
		followingStatusMap = make(map[uint64]bool)
		for _, user := range followingList {
			if user.Id != currentUserId {
				followingStatusMap[user.Id] = allFollowStatusMap[user.Id]
			}
		}

		// 构建粉丝列表的状态映射
		followerStatusMap = make(map[uint64]bool)
		for _, user := range followerList {
			if user.Id != currentUserId {
				followerStatusMap[user.Id] = allFollowStatusMap[user.Id]
			}
		}
	}

	viewrender.Render(c, "user.gohtml", map[string]any{
		"PageMeta": viewrender.NewPageMetaBuilder().
			SetUserProfile(showUser.Username, showUser.Bio).
			SetCanonicalURL(component.BuildCanonicalHref(c)).
			Build(),
		"Articles":             hotdataserve.ArticlesSmallEntity2Dto(last),
		"Author":               showUser,
		"AuthorInfoStatistics": authorInfoStatistics,
		"FollowingList":        followingList,
		"FollowerList":         followerList,
		"FollowingStatusMap":   followingStatusMap,
		"FollowerStatusMap":    followerStatusMap,
		"IsFollowingAuthor":    isFollowingAuthor,
		"ExternalInformation":  showUser.ExternalInformation,
	})
}

func About(c *gin.Context) {
	viewrender.Render(c, "about.gohtml", map[string]any{
		"PageMeta": viewrender.NewPageMetaBuilder().
			SetTitle(`关于`).
			SetDescription(`GooseForum's about`).
			SetCanonicalURL(component.BuildCanonicalHref(c)).
			Build(),
	})
}

func SponsorsView(c *gin.Context) {
	sponsorsInfo := hotdataserve.SponsorsConfigCache()
	viewrender.Render(c, "sponsors.gohtml", map[string]any{
		"PageMeta": viewrender.NewPageMetaBuilder().
			SetTitle(`赞助商`).
			SetDescription(`GooseForum's sponsors`).
			SetCanonicalURL(component.BuildCanonicalHref(c)).
			Build(),
		"SponsorsInfo": sponsorsInfo,
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
		"PageMeta": viewrender.NewPageMetaBuilder().
			SetTitle(`用户协议`).
			SetDescription(`GooseForum 用户服务协议`).
			SetCanonicalURL(component.BuildCanonicalHref(c)).
			Build(),
		"Title":       "用户协议 - GooseForum",
		"Subtitle":    "Terms of Service",
		"Description": "GooseForum 用户服务协议",
		"Content":     template.HTML(htmlContent),
	})
}

// PrivacyPolicy 隐私政策页面
func PrivacyPolicy(c *gin.Context) {
	htmlContent := markdown2html.MarkdownToHTML(privacyPolicyMD)
	viewrender.Render(c, "markdown-page.gohtml", map[string]any{
		"PageMeta": viewrender.NewPageMetaBuilder().
			SetTitle(`隐私政策`).
			SetDescription(`GooseForum 隐私保护政策`).
			SetCanonicalURL(component.BuildCanonicalHref(c)).
			Build(),
		"Title":       "隐私政策 - GooseForum",
		"Subtitle":    "Privacy Policy",
		"Description": "GooseForum 隐私保护政策",
		"Content":     template.HTML(htmlContent),
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
		"PageMeta": viewrender.NewPageMetaBuilder().
			SetTitle(`友情链接`).
			SetDescription(`GooseForum's links`).
			SetCanonicalURL(component.BuildCanonicalHref(c)).
			Build(),
		"FriendLinksGroup":    res,
		"TotalCounter":        totalCounter,
		"RecommendedArticles": hotdataserve.GetRecommendedArticles(),
		"LinkStatisticsInfo":  linkStatisticsInfo,
	})
}

func Profile(c *gin.Context) {
	viewrender.Render(c, "profile.gohtml", map[string]any{
		"PageMeta": viewrender.NewPageMetaBuilder().
			SetTitle(`个人中心`).
			SetCanonicalURL(component.BuildCanonicalHref(c)).
			Build(),
	})
}
func PublishV3(c *gin.Context) {
	viewrender.Render(c, "publish-v3.gohtml", map[string]any{
		"PageMeta": viewrender.NewPageMetaBuilder().
			SetTitle(`发布中心`).
			SetCanonicalURL(component.BuildCanonicalHref(c)).
			Build(),
	})
}

func Notifications(c *gin.Context) {
	viewrender.Render(c, "notifications.gohtml", map[string]any{
		"PageMeta": viewrender.NewPageMetaBuilder().
			SetTitle(`通知中心`).
			SetCanonicalURL(component.BuildCanonicalHref(c)).
			Build(),
	})
}

func Admin(c *gin.Context) {
	viewrender.Render(c, "admin.gohtml", map[string]any{
		"PageMeta": viewrender.NewPageMetaBuilder().
			SetTitle(`管理`).
			SetCanonicalURL(component.BuildCanonicalHref(c)).
			Build(),
	})
}
