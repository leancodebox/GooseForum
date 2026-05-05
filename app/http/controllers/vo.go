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

// SidebarItem represents a single item in the sidebar
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

// SidebarMenu controls the display and selection of sidebar items
type SidebarMenu struct {
	MainMenu      []*SidebarItem
	ResourceItems []*SidebarItem
	CategoryItems []*SidebarItem
}

// SetActive sets the active item by key
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

// SetActiveCategory sets the active category by ID
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

// NewSidebarMenu creates a new SidebarMenu with default items and categories
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

// CommonDataVo 包含所有 V3 页面共有的基础数据（如侧边栏、站点统计等）
type CommonDataVo struct {
	ArticleCategoryList []*articleCategory.Entity
	Stats               *vo.SiteStats
	RecommendedArticles []*articles.SmallEntity
	Announcement        pageConfig.AnnouncementConfig
	GooseForumInfo      ForumInfo
	Category            *articleCategory.Entity // 当前页面所属的分类（如果有），用于侧边栏高亮
	Sidebar             SidebarMenu             // 侧边栏菜单控制
}

// GetCommonData 获取 V3 页面共有的基础数据（使用 hotdataserve.GetOrLoad 缓存机制）
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

	// Rebuild SidebarMenu for each request to allow per-request state (Active)
	currentUserId := component.LoginUserId(c)
	data.Sidebar = NewSidebarMenu(data.ArticleCategoryList, currentUserId > 0)
	return data
}

// HomeData V3 首页专用的数据结构
type HomeData struct {
	CommonDataVo
	LatestArticles []*vo.ArticlesSimpleVo
	CurrentPage    int
	NextPage       int
	Sort           string
}

// CategoryData V3 分类页专用的数据结构
type CategoryData struct {
	CommonDataVo
	Category          *articleCategory.Entity
	CategoryFirstChar string
	Articles          []*vo.ArticlesSimpleVo
	CurrentPage       int
	NextPage          int
	Sort              string
}

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

// PostDetailDataVo V3 帖子详情页专用的数据结构
type PostDetailDataVo struct {
	CommonDataVo
	PostDetailData
	LatestArticles      []*vo.ArticlesSimpleVo
	ArticleCategoryList []*articleCategory.Entity // 显式定义以解决嵌入结构体字段冲突
}

type UserDataVo struct {
	CommonDataVo
	UserData
}

type MessagesData struct {
	CommonDataVo
}

type SettingsData struct {
	CommonDataVo
	User  *vo.UserDetailedVo
	Stats userStatistics.Entity
}

type NewTopicData struct {
	CommonDataVo
	ArticleId uint64
}

type SearchData struct {
	CommonDataVo
	Query          string
	CurrentPage    int
	TotalPages     int
	PageNumbers    []int
	SearchResponse *searchservice.SearchResponse
	ArticleList    []*vo.ArticlesSimpleVo
}
