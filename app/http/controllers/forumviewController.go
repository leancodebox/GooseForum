package controllers

import (
	_ "embed"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/datastruct"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/articleBookmark"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"

	"github.com/gin-gonic/gin"
	array "github.com/leancodebox/GooseForum/app/bundles/collectionopt"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/http/controllers/transform"
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

//go:embed docs/terms-of-service.md
var termsOfServiceMD string

//go:embed docs/privacy-policy.md
var privacyPolicyMD string

type PageButton struct {
	Index int
	Page  int
}

type HomeData struct {
	ArticleCategoryList []datastruct.Option[string, uint64]
	LatestArticles      []vo.ArticlesSimpleDto
	Stats               vo.SiteStats
	RecommendedArticles []articles.SmallEntity
	Announcement        pageConfig.AnnouncementConfig
	GooseForumInfo      ForumInfo
}

func Home(c *gin.Context) {
	pageMeta := viewrender.NewPageMetaBuilder().
		SetTitle("GooseForum - 自由漫谈的江湖茶馆").
		SetDescription("GooseForum's home").
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build()

	viewrender.SafeRender(c, "index.gohtml", HomeData{
		ArticleCategoryList: hotdataserve.ArticleCategoryLabel(),
		LatestArticles:      hotdataserve.GetLatestArticleSimpleDto(),
		Stats:               hotdataserve.GetSiteStatisticsData(),
		RecommendedArticles: hotdataserve.GetRecommendedArticles(),
		Announcement:        hotdataserve.GetAnnouncementConfigCache(),
		GooseForumInfo:      GetGooseForumInfo(),
	}, pageMeta)
}

func LoginView(c *gin.Context) {
	viewrender.SafeRender[any](c, "login.gohtml", nil, viewrender.NewPageMetaBuilder().
		SetTitle("登录/注册").
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build())
}

func ResetPasswordView(c *gin.Context) {
	viewrender.SafeRender[any](c, "reset-password.gohtml", nil, viewrender.NewPageMetaBuilder().
		SetTitle("重置密码").
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build())
}

type PostDetailData struct {
	Article              articles.Entity
	Username             string
	CommentList          []ReplyDto
	AvatarUrl            string
	AuthorArticles       []articles.SmallEntity
	ArticleCategory      []string
	AuthorUserInfo       users.EntityComplete
	AuthorInfoStatistics userStatistics.Entity
	AuthorCard           *vo.UserCard
	IsOwnArticle         bool
	ArticleCategoryList  []datastruct.Option[string, uint64]
	ILike                bool
	IsFollowing          bool
	IsBookmarked         bool
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
	replyMap := array.Slice2Map(replyEntities, func(item *reply.Entity) uint64 {
		return item.Id
	})
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
			if replyTo, ok := replyMap[item.ReplyId]; ok {
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

	authorCard := transform.User2UserCard(authorUserInfo, authorInfoStatistics, isFollowing, currentUserId == entity.UserId, currentUserId > 0, false)

	// 构建模板数据
	viewrender.SafeRender(c, "detail.gohtml", PostDetailData{
		Article:              entity,
		Username:             author,
		CommentList:          replyList,
		AvatarUrl:            avatarUrl,
		AuthorArticles:       authorArticles,
		ArticleCategory:      articleCategory,
		AuthorUserInfo:       authorUserInfo,
		AuthorInfoStatistics: authorInfoStatistics,
		AuthorCard:           authorCard,
		IsOwnArticle:         currentUserId == entity.UserId,
		ArticleCategoryList:  hotdataserve.ArticleCategoryLabel(),
		ILike:                iLike,
		IsFollowing:          isFollowing,
		IsBookmarked:         isBookmarked,
	}, viewrender.NewPageMetaBuilder().
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
		Build())

	articles.IncrementView(entity)
}

type ForumInfo struct {
	Title        string
	Desc         string
	Independence bool
}

func GetGooseForumInfo() ForumInfo {
	siteData := hotdataserve.GetSiteSettingsConfigCache()
	return ForumInfo{
		Title:        siteData.SiteName,
		Desc:         siteData.SiteDescription,
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

type PostListData struct {
	ArticleList         []vo.ArticlesSimpleDto
	Page                int
	PageSize            int
	Total               int64
	TotalPages          int
	PrevPage            int
	NextPage            int
	ArticleCategoryList []datastruct.Option[string, uint64]
	RecommendedArticles []articles.SmallEntity
	CanonicalHref       string
	Filters             string
	FilterIds           []int
	NoFilter            bool
	Pagination          []PageButton
	Stats               vo.SiteStats
	ForumInfo           ForumInfo
}

func Post(c *gin.Context) {
	filters := c.DefaultQuery("filters", "")
	categories := array.Filter(array.Map(strings.Split(filters, "-"), func(t string) int {
		return cast.ToInt(t)
	}), func(i int) bool {
		return i > 0
	})
	page := min(60, cast.ToInt(c.DefaultQuery("page", "1")))
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
	defaultForumInfo := GetGooseForumInfo()
	title := defaultForumInfo.Title
	description := defaultForumInfo.Desc
	if len(categories) == 1 {
		if category, ok := categoryMap[cast.ToUint64(categories[0])]; ok {
			forumInfo.Independence = true
			forumInfo.Desc = category.Desc
			forumInfo.Title = category.Category

			title = category.Category + " 社区 | GooseForum"
			description = category.Desc
		}
	}

	viewrender.SafeRender(c, "list.gohtml", PostListData{
		ArticleList:         articleList,
		Page:                pageData.Page,
		PageSize:            pageSize,
		Total:               pageData.Total,
		TotalPages:          totalPages,
		PrevPage:            max(pageData.Page-1, 1),
		NextPage:            min(max(pageData.Page, 1)+1, totalPages),
		ArticleCategoryList: articleCategoryList,
		RecommendedArticles: hotdataserve.GetRecommendedArticles(),
		CanonicalHref:       component.BuildCanonicalHref(c),
		Filters:             filters,
		FilterIds:           categories,
		NoFilter:            len(categories) == 0,
		Pagination:          pagination,
		Stats:               hotdataserve.GetSiteStatisticsData(),
		ForumInfo:           forumInfo,
	}, viewrender.NewPageMetaBuilder().
		SetTitle(title).
		SetDescription(description).
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build())
}

type UserData struct {
	Articles       []vo.ArticlesSimpleDto
	UserCard       *vo.UserCard
	FollowingList  []*users.EntityComplete
	FollowerList   []*users.EntityComplete
	MyFollowingIds []uint64
}

func User(c *gin.Context) {
	id := cast.ToUint64(c.Param("userId"))
	showUser := component.GetUserShowByUserId(id)
	if showUser.UserId == 0 {
		errorPage(c, "用户不存在", "用户不存在")
		return
	}
	last, _ := articles.GetLatestArticlesByUserId(id, 5)

	// 获取关注和粉丝列表（默认显示前10个）
	followingList, _ := userFollow.GetFollowingList(id, 1, 10)
	followerList, _ := userFollow.GetFollowerList(id, 1, 10)

	// 获取当前登录用户信息
	currentUserId := component.LoginUserId(c)

	// 检查当前用户是否关注了列表中的用户
	var isFollowingAuthor bool

	if currentUserId > 0 && currentUserId != id {
		isFollowingAuthor = userFollow.IsFollowing(currentUserId, id)
	}
	user, _ := users.Get(id)
	stats := userStatistics.Get(id)
	userCard := transform.User2UserCard(user, stats, isFollowingAuthor, currentUserId == id, true, true)

	var myFollowingIds []uint64
	if currentUserId > 0 {
		myFollowingIds = userFollow.GetAllFollowingIds(currentUserId)
	}

	viewrender.SafeRender(c, "user.gohtml", UserData{
		Articles:       hotdataserve.ArticlesSmallEntity2Dto(last),
		UserCard:       userCard,
		FollowingList:  followingList,
		FollowerList:   followerList,
		MyFollowingIds: myFollowingIds,
	}, viewrender.NewPageMetaBuilder().
		SetUserProfile(showUser.Username, showUser.Bio).
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build())
}

func About(c *gin.Context) {
	viewrender.SafeRender[any](c, "about.gohtml", nil, viewrender.NewPageMetaBuilder().
		SetTitle(`关于`).
		SetDescription(`GooseForum's about`).
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build())
}

type SponsorsData struct {
	SponsorsInfo pageConfig.SponsorsConfig
}

func SponsorsView(c *gin.Context) {
	sponsorsInfo := hotdataserve.SponsorsConfigCache()
	viewrender.SafeRender(c, "sponsors.gohtml", SponsorsData{
		SponsorsInfo: sponsorsInfo,
	}, viewrender.NewPageMetaBuilder().
		SetTitle(`赞助商`).
		SetDescription(`GooseForum's sponsors`).
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build())
}

type MarkdownPageData struct {
	Title       string
	Subtitle    string
	Description string
	Content     template.HTML
}

// TermsOfService 用户协议页面
func TermsOfService(c *gin.Context) {
	htmlContent := markdown2html.MarkdownToHTML(termsOfServiceMD)
	viewrender.SafeRender(c, "markdown-page.gohtml", MarkdownPageData{
		Title:       "用户协议 - GooseForum",
		Subtitle:    "Terms of Service",
		Description: "GooseForum 用户服务协议",
		Content:     template.HTML(htmlContent),
	}, viewrender.NewPageMetaBuilder().
		SetTitle(`用户协议`).
		SetDescription(`GooseForum 用户服务协议`).
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build())
}

// PrivacyPolicy 隐私政策页面
func PrivacyPolicy(c *gin.Context) {
	htmlContent := markdown2html.MarkdownToHTML(privacyPolicyMD)
	viewrender.SafeRender(c, "markdown-page.gohtml", MarkdownPageData{
		Title:       "隐私政策 - GooseForum",
		Subtitle:    "Privacy Policy",
		Description: "GooseForum 隐私保护政策",
		Content:     template.HTML(htmlContent),
	}, viewrender.NewPageMetaBuilder().
		SetTitle(`隐私政策`).
		SetDescription(`GooseForum 隐私保护政策`).
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build())
}

type LinkStatistics struct {
	Name       string
	Counter    int
	Proportion int
}

type LinkStatisticsInfo struct {
	Tool      LinkStatistics
	Blog      LinkStatistics
	Community LinkStatistics
}

type LinksData struct {
	FriendLinksGroup    []pageConfig.FriendLinksGroup
	TotalCounter        int
	RecommendedArticles []articles.SmallEntity
	LinkStatisticsInfo  LinkStatisticsInfo
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
	viewrender.SafeRender(c, "links.gohtml", LinksData{
		FriendLinksGroup:    res,
		TotalCounter:        totalCounter,
		RecommendedArticles: hotdataserve.GetRecommendedArticles(),
		LinkStatisticsInfo:  linkStatisticsInfo,
	}, viewrender.NewPageMetaBuilder().
		SetTitle(`友情链接`).
		SetDescription(`GooseForum's links`).
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build())
}

type ProfileData struct {
	UserCard *vo.UserCard
	FullUser *users.EntityComplete
	Stats    userStatistics.Entity
}

func Profile(c *gin.Context) {
	userId := component.LoginUserId(c)
	if userId == 0 {
		c.Redirect(302, "/login")
		return
	}
	user, err := users.Get(userId)
	if err != nil {
		errorPage(c, "用户不存在", "用户不存在")
		return
	}
	stats := userStatistics.Get(userId)
	userCard := transform.User2UserCard(user, stats, false, true, true, true)

	viewrender.SafeRender(c, "profile.gohtml", ProfileData{
		UserCard: userCard,
		FullUser: &user,
		Stats:    stats,
	}, viewrender.NewPageMetaBuilder().
		SetTitle(`个人中心`).
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build())
}

func UserSetting(c *gin.Context) {
	userId := component.LoginUserId(c)
	if userId == 0 {
		c.Redirect(302, "/login")
		return
	}
	user, err := users.Get(userId)
	if err != nil {
		errorPage(c, "用户不存在", "用户不存在")
		return
	}
	stats := userStatistics.Get(userId)
	userCard := transform.User2UserCard(user, stats, false, true, true, true)

	viewrender.SafeRender(c, "settings.gohtml", ProfileData{
		UserCard: userCard,
		FullUser: &user,
		Stats:    stats,
	}, viewrender.NewPageMetaBuilder().
		SetTitle(`个人中心`).
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build())
}

func Publish(c *gin.Context) {
	viewrender.SafeRender[any](c, "publish.gohtml", nil, viewrender.NewPageMetaBuilder().
		SetTitle(`发布中心`).
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build())
}

func Notifications(c *gin.Context) {
	viewrender.SafeRender[any](c, "notifications.gohtml", nil, viewrender.NewPageMetaBuilder().
		SetTitle(`通知中心`).
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build())
}

func Admin(c *gin.Context) {
	viewrender.SafeRender[any](c, "admin.gohtml", nil, viewrender.NewPageMetaBuilder().
		SetTitle(`管理`).
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build())
}
