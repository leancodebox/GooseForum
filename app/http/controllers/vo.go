package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/searchservice"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"
)

type null struct {
}

// SidebarItem describes one navigation entry rendered by the shared sidebar.
type SidebarItem struct {
	Key        string
	Label      string
	Icon       string
	Color      string
	Url        string
	Active     bool
	IsCategory bool
	CategoryId uint64
}

// SidebarMenu groups sidebar navigation entries and tracks their active state.
type SidebarMenu struct {
	MainMenu      []*SidebarItem
	ResourceItems []*SidebarItem
	CategoryItems []*SidebarItem
}

// SetActive marks the sidebar item with the given key as active.
func (s *SidebarMenu) SetActive(key string) {
	for _, item := range s.MainMenu {
		item.Active = item.Key == key
	}
	for _, item := range s.ResourceItems {
		item.Active = item.Key == key
	}
	for _, item := range s.CategoryItems {
		item.Active = item.Key == key
	}
}

// SetActiveCategory marks the category item with the given category ID as active.
func (s *SidebarMenu) SetActiveCategory(id uint64) {
	for _, item := range s.MainMenu {
		item.Active = false
	}
	for _, item := range s.ResourceItems {
		item.Active = false
	}
	for _, item := range s.CategoryItems {
		item.Active = item.CategoryId == id
	}
}

// NewSidebarMenu builds the default sidebar menu for the current visitor.
func NewSidebarMenu(categories []*articleCategory.Entity, isLoggedIn bool) SidebarMenu {
	mainMenu := []*SidebarItem{
		{Key: "topics", Label: "nav_topics", Icon: "💬", Url: urlconfig.Home()},
		{Key: "hot", Label: "nav_more", Icon: "🔥", Url: urlconfig.Home() + "?sort=hot"},
		{Key: "popular", Label: "nav_recent", Icon: "📅", Url: urlconfig.Home() + "?sort=popular"},
	}

	if isLoggedIn {
		mainMenu = append(mainMenu,
			&SidebarItem{Key: "messages", Label: "nav_messages", Icon: "📨", Url: urlconfig.Messages()},
			&SidebarItem{Key: "notifications", Label: "nav_notifications", Icon: "🔔", Url: urlconfig.Notifications()},
		)
	}

	menu := SidebarMenu{
		MainMenu: mainMenu,
		ResourceItems: []*SidebarItem{
			{Key: "links", Label: "nav_links", Icon: "🔗", Url: urlconfig.Links()},
			{Key: "sponsors", Label: "nav_acknowledgements", Icon: "❤️", Url: urlconfig.Sponsors()},
			//{Key: "docs", Label: "nav_docs", Icon: "📘", Url: urlconfig.Docs()},
		},
	}

	for _, cat := range categories {
		menu.CategoryItems = append(menu.CategoryItems, &SidebarItem{
			Key:        "category_" + strconv.FormatUint(cat.Id, 10),
			Label:      cat.Category,
			Icon:       cat.Icon,
			Color:      cat.Color,
			Url:        "/c/" + cat.Category + "/" + strconv.FormatUint(cat.Id, 10),
			IsCategory: true,
			CategoryId: cat.Id,
		})
	}

	return menu
}

// CommonDataVo carries data shared by V3 template-rendered pages.
type CommonDataVo struct {
	ArticleCategoryList []*articleCategory.Entity
	Stats               *vo.SiteStats
	RecommendedArticles []*articles.SmallEntity
	Announcement        pageConfig.AnnouncementConfig
	GooseForumInfo      ForumInfo
	Category            *articleCategory.Entity // Category is the current page category, if one should be highlighted.
	Sidebar             SidebarMenu             // Sidebar is rebuilt per request so active state never leaks from cache.
}

// GetCommonData loads cached V3 shared page data and adds request-scoped sidebar state.
func GetCommonData(c *gin.Context) CommonDataVo {
	data := hotdataserve.GetOrLoad("common_data", func() (CommonDataVo, error) {
		return CommonDataVo{
			ArticleCategoryList: hotdataserve.GetArticleCategory(),
			Stats:               hotdataserve.GetSiteStatisticsData(),
			RecommendedArticles: hotdataserve.GetRecommendedArticles(),
			Announcement:        hotdataserve.GetAnnouncementConfigCache(),
			GooseForumInfo:      GetGooseForumInfo(),
		}, nil
	})

	currentUserId := component.LoginUserId(c)
	data.Sidebar = NewSidebarMenu(data.ArticleCategoryList, currentUserId > 0)
	return data
}

// HomeData is the V3 home page view object.
type HomeData struct {
	CommonDataVo
	LatestArticles []*vo.ArticlesSimpleVo
	CurrentPage    int
	NextPage       int
	Sort           string
}

// CategoryData is the V3 category page view object.
type CategoryData struct {
	CommonDataVo
	Category          *articleCategory.Entity
	CategoryFirstChar string
	Articles          []*vo.ArticlesSimpleVo
	CurrentPage       int
	NextPage          int
	Sort              string
}

// PostDetailData contains the article detail fields shared by old and V3 detail views.
type PostDetailData struct {
	Article              articles.Entity
	Username             string
	CommentList          []ReplyVo
	AvatarUrl            string
	AuthorArticles       []*articles.SmallEntity
	ArticleCategory      []string
	AuthorUserInfo       users.EntityComplete
	AuthorInfoStatistics userStatistics.Entity
	IsOwnArticle         bool
	ArticleCategoryList  []*articleCategory.Entity
	ILike                bool
	IsFollowing          bool
	IsBookmarked         bool
	Posters              []vo.PosterVo
}

// PostDetailDataVo is the V3 post detail page view object.
type PostDetailDataVo struct {
	CommonDataVo
	PostDetailData
	LatestArticles      []*vo.ArticlesSimpleVo
	ArticleCategoryList []*articleCategory.Entity // ArticleCategoryList resolves the embedded field name conflict.
}

// UserDataVo is the V3 user profile page view object.
type UserDataVo struct {
	CommonDataVo
	UserData
}

// MessagesData is the V3 private messages page view object.
type MessagesData struct {
	CommonDataVo
}

// SettingsData is the V3 account settings page view object.
type SettingsData struct {
	CommonDataVo
	User  *vo.UserDetailedVo
	Stats userStatistics.Entity
}

// NewTopicData is the V3 topic editor page view object.
type NewTopicData struct {
	CommonDataVo
	ArticleId uint64
}

// SearchData is the V3 search results page view object.
type SearchData struct {
	CommonDataVo
	Query          string
	CurrentPage    int
	TotalPages     int
	PageNumbers    []int
	SearchResponse *searchservice.SearchResponse
	ArticleList    []*vo.ArticlesSimpleVo
}
