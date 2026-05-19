package forum

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/http/controllers/transform"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/articleBookmark"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
	"github.com/leancodebox/GooseForum/app/models/forum/articleLike"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/userActivities"
	"github.com/leancodebox/GooseForum/app/models/forum/userFollow"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/chatservice"
	"github.com/leancodebox/GooseForum/app/service/notificationservice"
	"github.com/leancodebox/GooseForum/app/service/searchservice"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"
	"github.com/samber/lo"
)

const payloadVersion = "1.0"

type PagePayload struct {
	Component string        `json:"component"`
	Props     any           `json:"props"`
	Meta      PageMeta      `json:"meta"`
	Layout    LayoutPayload `json:"layout"`
	URL       string        `json:"url"`
	Version   string        `json:"version"`
}

type PageMeta struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Canonical   string `json:"canonical,omitempty"`
	Robots      string `json:"robots,omitempty"`
}

type ErrorPageProps struct {
	Code    string `json:"code"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

type LoginPageProps struct {
	InitialMode string `json:"initialMode"`
	RedirectURL string `json:"redirectUrl"`
	GitHubURL   string `json:"githubUrl"`
	GoogleReady bool   `json:"googleReady"`
}

type ResetPasswordPageProps struct {
	Token string `json:"token"`
}

type LayoutPayload struct {
	Site    SitePayload    `json:"site"`
	Viewer  ViewerPayload  `json:"viewer"`
	Sidebar SidebarPayload `json:"sidebar"`
	Footer  FooterPayload  `json:"footer"`
}

type SitePayload struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Logo          string `json:"logo"`
	Favicon       string `json:"favicon"`
	ExternalLinks string `json:"externalLinks,omitempty"`
	BrandType     string `json:"brandType"`
	BrandText     string `json:"brandText"`
	BrandImage    string `json:"brandImage"`
}

type ViewerPayload struct {
	ID              uint64 `json:"id"`
	Username        string `json:"username"`
	AvatarURL       string `json:"avatarUrl"`
	IsAuthenticated bool   `json:"isAuthenticated"`
	IsAdmin         bool   `json:"isAdmin"`
}

type NavItemPayload struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Icon   string `json:"icon"`
	URL    string `json:"url"`
	Active bool   `json:"active"`
}

type CategoryNavPayload struct {
	NavItemPayload
	ID    uint64 `json:"id"`
	Color string `json:"color"`
}

type SidebarPayload struct {
	Main       []NavItemPayload     `json:"main"`
	Resources  []NavItemPayload     `json:"resources"`
	Categories []CategoryNavPayload `json:"categories"`
	ActiveKey  string               `json:"activeKey"`
}

type FooterPayload struct {
	Links   []pageConfig.FooterItem `json:"links"`
	Primary []string                `json:"primary"`
}

type HomeProps struct {
	Sort         string              `json:"sort"`
	Tabs         []TabPayload        `json:"tabs"`
	Topics       []TopicPayload      `json:"topics"`
	Pagination   PaginationPayload   `json:"pagination"`
	Announcement AnnouncementPayload `json:"announcement"`
}

type TabPayload struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	URL    string `json:"url"`
	Active bool   `json:"active"`
}

type PaginationPayload struct {
	Page     int    `json:"page"`
	NextPage int    `json:"nextPage"`
	HasNext  bool   `json:"hasNext"`
	NextURL  string `json:"nextUrl"`
}

type AnnouncementPayload struct {
	Enabled bool   `json:"enabled"`
	HTML    string `json:"html"`
}

type TopicPayload struct {
	ID             uint64                 `json:"id"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	URL            string                 `json:"url"`
	Author         TopicAuthorPayload     `json:"author"`
	Participants   []TopicAuthorPayload   `json:"participants"`
	Categories     []TopicCategoryPayload `json:"categories"`
	ReplyCount     uint64                 `json:"replyCount"`
	ViewCount      uint64                 `json:"viewCount"`
	ActivityText   string                 `json:"activityText"`
	LastUpdateTime string                 `json:"lastUpdateTime"`
}

type TopicAuthorPayload struct {
	ID        uint64 `json:"id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatarUrl"`
}

type TopicCategoryPayload struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	URL   string `json:"url"`
	Color string `json:"color"`
}

type ArticleDetailProps struct {
	Article     ArticlePayload     `json:"article"`
	Replies     []ReplyPayload     `json:"replies"`
	HotTopics   []TopicPayload     `json:"hotTopics"`
	Permissions ArticlePermissions `json:"permissions"`
}

type ArticlePayload struct {
	ID           uint64                 `json:"id"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	URL          string                 `json:"url"`
	HTML         string                 `json:"html"`
	Author       TopicAuthorPayload     `json:"author"`
	Participants []TopicAuthorPayload   `json:"participants"`
	Categories   []TopicCategoryPayload `json:"categories"`
	ReplyCount   uint64                 `json:"replyCount"`
	ViewCount    uint64                 `json:"viewCount"`
	LikeCount    uint64                 `json:"likeCount"`
	IsLiked      bool                   `json:"isLiked"`
	IsBookmarked bool                   `json:"isBookmarked"`
	CreatedAt    string                 `json:"createdAt"`
	UpdatedAt    string                 `json:"updatedAt"`
}

type ReplyPayload struct {
	ID              uint64             `json:"id"`
	ArticleID       uint64             `json:"articleId"`
	Content         string             `json:"content"`
	Author          TopicAuthorPayload `json:"author"`
	CreatedAt       string             `json:"createdAt"`
	ReplyToID       uint64             `json:"replyToId,omitempty"`
	ReplyToUserID   uint64             `json:"replyToUserId,omitempty"`
	ReplyToUsername string             `json:"replyToUsername,omitempty"`
	IsOwnReply      bool               `json:"isOwnReply"`
}

type ArticlePermissions struct {
	IsOwnArticle bool `json:"isOwnArticle"`
	CanReply     bool `json:"canReply"`
}

type UserProfileProps struct {
	User         *vo.UserCard            `json:"user"`
	Topics       []TopicPayload          `json:"topics"`
	Activities   []UserActivityPayload   `json:"activities"`
	Following    []UserConnectionPayload `json:"following"`
	Followers    []UserConnectionPayload `json:"followers"`
	IsOwnProfile bool                    `json:"isOwnProfile"`
	CanMessage   bool                    `json:"canMessage"`
	CanFollow    bool                    `json:"canFollow"`
	MessageURL   string                  `json:"messageUrl"`
	SettingsURL  string                  `json:"settingsUrl"`
}

type UserActivityPayload struct {
	ID             uint64 `json:"id"`
	Action         int    `json:"action"`
	SubjectType    string `json:"subjectType"`
	SubjectID      uint64 `json:"subjectId"`
	ContentPreview string `json:"contentPreview"`
	URL            string `json:"url"`
	Label          string `json:"label"`
	CreatedAt      string `json:"createdAt"`
}

type UserConnectionPayload struct {
	ID        uint64 `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatarUrl"`
	Bio       string `json:"bio"`
	URL       string `json:"url"`
}

type CategoryPageProps struct {
	Category   CategoryHeaderPayload `json:"category"`
	Sort       string                `json:"sort"`
	Tabs       []TabPayload          `json:"tabs"`
	Topics     []TopicPayload        `json:"topics"`
	Pagination PaginationPayload     `json:"pagination"`
}

type CategoryHeaderPayload struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
	URL         string `json:"url"`
}

type LinksPageProps struct {
	Groups     []LinkGroupPayload `json:"groups"`
	TotalCount int                `json:"totalCount"`
}

type LinkGroupPayload struct {
	Name  string              `json:"name"`
	Emoji string              `json:"emoji"`
	Color string              `json:"color"`
	Links []FriendLinkPayload `json:"links"`
}

type FriendLinkPayload struct {
	Name    string `json:"name"`
	Desc    string `json:"desc"`
	URL     string `json:"url"`
	LogoURL string `json:"logoUrl"`
}

type SponsorsPageProps struct {
	Sections   []SponsorSectionPayload `json:"sections"`
	TotalCount int                     `json:"totalCount"`
}

type SponsorSectionPayload struct {
	Key      string           `json:"key"`
	Label    string           `json:"label"`
	Tone     string           `json:"tone"`
	Sponsors []SponsorPayload `json:"sponsors"`
}

type SponsorPayload struct {
	Name      string `json:"name"`
	Message   string `json:"message"`
	Link      string `json:"link"`
	AvatarURL string `json:"avatarUrl"`
}

type NotificationsPageProps struct {
	Total         int64                 `json:"total"`
	UnreadCount   int64                 `json:"unreadCount"`
	Notifications []NotificationPayload `json:"notifications"`
	Pagination    PaginationPayload     `json:"pagination"`
}

type NotificationPayload struct {
	ID        uint64                                `json:"id"`
	EventType string                                `json:"eventType"`
	IsRead    bool                                  `json:"isRead"`
	CreatedAt string                                `json:"createdAt"`
	Title     string                                `json:"title"`
	Content   string                                `json:"content"`
	Actor     TopicAuthorPayload                    `json:"actor"`
	Article   *NotificationArticlePayload           `json:"article,omitempty"`
	Payload   eventNotification.NotificationPayload `json:"payload"`
}

type NotificationArticlePayload struct {
	ID    uint64 `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type MessagesPageProps struct {
	Conversations  []MessageConversationPayload `json:"conversations"`
	SuggestedUsers []UserConnectionPayload      `json:"suggestedUsers"`
}

type MessageConversationPayload struct {
	ID           uint64 `json:"id"`
	PeerID       uint64 `json:"peerId"`
	PeerUsername string `json:"peerUsername"`
	PeerAvatar   string `json:"peerAvatar"`
	LastMsg      string `json:"lastMsg"`
	LastMsgTime  string `json:"lastMsgTime"`
	UnreadCount  uint   `json:"unreadCount"`
	ConvID       uint64 `json:"convId"`
	PeerURL      string `json:"peerUrl"`
}

type SettingsPageProps struct {
	User  *vo.UserDetailedVo   `json:"user"`
	Stats SettingsStatsPayload `json:"stats"`
	Tabs  []TabPayload         `json:"tabs"`
}

type SettingsStatsPayload struct {
	ArticleCount      uint   `json:"articleCount"`
	ReplyCount        uint   `json:"replyCount"`
	FollowerCount     uint   `json:"followerCount"`
	FollowingCount    uint   `json:"followingCount"`
	LikeReceivedCount uint   `json:"likeReceivedCount"`
	CreatedAt         string `json:"createdAt"`
}

type PublishPageProps struct {
	ArticleID  uint64                   `json:"articleId"`
	IsEditing  bool                     `json:"isEditing"`
	Categories []PublishCategoryPayload `json:"categories"`
	Types      []PublishTypePayload     `json:"types"`
	Article    PublishArticlePayload    `json:"article"`
}

type PublishCategoryPayload struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type PublishTypePayload struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type PublishArticlePayload struct {
	Title       string   `json:"title"`
	Content     string   `json:"content"`
	Type        int8     `json:"type"`
	CategoryIDs []uint64 `json:"categoryIds"`
}

type SearchPageProps struct {
	Query      string            `json:"query"`
	Topics     []TopicPayload    `json:"topics"`
	Total      int64             `json:"total"`
	TotalPages int               `json:"totalPages"`
	Pagination PaginationPayload `json:"pagination"`
}

func buildLayout(c *gin.Context, activeKey string) LayoutPayload {
	siteConfig := hotdataserve.GetSiteSettingsConfigCache()
	currentUser := component.GetLoginUser(c)
	viewer := ViewerPayload{}
	if currentUser != nil {
		viewer = ViewerPayload{
			ID:              currentUser.UserId,
			Username:        currentUser.Username,
			AvatarURL:       currentUser.AvatarUrl,
			IsAuthenticated: currentUser.UserId > 0,
			IsAdmin:         currentUser.IsAdmin,
		}
	}

	footerPrimary := make([]string, 0, len(siteConfig.FooterInfo.Primary))
	for _, item := range siteConfig.FooterInfo.Primary {
		footerPrimary = append(footerPrimary, item.Content)
	}

	return LayoutPayload{
		Site: SitePayload{
			Name:          siteConfig.SiteName,
			Description:   siteConfig.SiteDescription,
			Logo:          siteConfig.SiteLogo,
			Favicon:       siteConfig.SiteLogo,
			ExternalLinks: siteConfig.ExternalLinks,
			BrandType:     siteConfig.BrandType,
			BrandText:     siteConfig.BrandText,
			BrandImage:    siteConfig.BrandImage,
		},
		Viewer: viewer,
		Sidebar: buildSidebarPayload(
			hotdataserve.GetArticleCategory(),
			activeKey,
			viewer.IsAuthenticated,
		),
		Footer: FooterPayload{
			Links:   siteConfig.FooterInfo.List,
			Primary: footerPrimary,
		},
	}
}

func buildSidebarPayload(categories []*articleCategory.Entity, activeKey string, isLoggedIn bool) SidebarPayload {
	main := []NavItemPayload{
		{Key: "topics", Label: "主题", Icon: "💬", URL: "/", Active: activeKey == "topics"},
		{Key: "hot", Label: "热门", Icon: "🔥", URL: "/?sort=hot", Active: activeKey == "hot"},
		{Key: "popular", Label: "流行", Icon: "📈", URL: "/?sort=popular", Active: activeKey == "popular"},
	}
	if isLoggedIn {
		main = append(main,
			NavItemPayload{Key: "messages", Label: "私信", Icon: "✉", URL: "/messages", Active: activeKey == "messages"},
			NavItemPayload{Key: "notifications", Label: "通知", Icon: "🔔", URL: "/notifications", Active: activeKey == "notifications"},
		)
	}

	categoryItems := make([]CategoryNavPayload, 0, len(categories))
	for _, category := range categories {
		if category == nil {
			continue
		}
		key := "category_" + strconv.FormatUint(category.Id, 10)
		categoryItems = append(categoryItems, CategoryNavPayload{
			NavItemPayload: NavItemPayload{
				Key:    key,
				Label:  category.Category,
				Icon:   category.Icon,
				URL:    categoryURL(category),
				Active: activeKey == key,
			},
			ID:    category.Id,
			Color: category.Color,
		})
	}

	return SidebarPayload{
		Main: main,
		Resources: []NavItemPayload{
			{Key: "links", Label: "链接", Icon: "🔗", URL: "/links", Active: activeKey == "links"},
			{Key: "sponsors", Label: "赞助", Icon: "♥", URL: "/sponsors", Active: activeKey == "sponsors"},
		},
		Categories: categoryItems,
		ActiveKey:  activeKey,
	}
}

func buildHomeProps(page int, sort string, topics []*vo.ArticlesSimpleVo) HomeProps {
	nextPage := 0
	if len(topics) == 20 {
		nextPage = page + 1
	}

	announcement := hotdataserve.GetAnnouncementConfigCache()
	return HomeProps{
		Sort:   sort,
		Tabs:   buildHomeTabs(sort),
		Topics: buildTopicPayloads(topics),
		Pagination: PaginationPayload{
			Page:     page,
			NextPage: nextPage,
			HasNext:  nextPage > 0,
			NextURL:  buildHomePageURL(sort, nextPage),
		},
		Announcement: AnnouncementPayload{
			Enabled: announcement.Enabled,
			HTML:    announcement.GetHtmlContent(),
		},
	}
}

func buildLoginPageProps(c *gin.Context) LoginPageProps {
	mode := "login"
	if c.Query("register") == "true" || c.Query("model") == "register" {
		mode = "register"
	}
	redirectURL := c.Query("redirect")
	githubURL := "/api/auth/github"
	if redirectURL != "" {
		githubURL += "?redirect=" + url.QueryEscape(redirectURL)
	}
	return LoginPageProps{
		InitialMode: mode,
		RedirectURL: redirectURL,
		GitHubURL:   githubURL,
		GoogleReady: false,
	}
}

func buildHomeTabs(sort string) []TabPayload {
	return []TabPayload{
		{Key: "latest", Label: "最新", URL: "/", Active: sort == "latest" || sort == ""},
		{Key: "hot", Label: "热门", URL: "/?sort=hot", Active: sort == "hot"},
		{Key: "popular", Label: "流行", URL: "/?sort=popular", Active: sort == "popular"},
	}
}

func buildTopicPayloads(topics []*vo.ArticlesSimpleVo) []TopicPayload {
	categoryMap := hotdataserve.ArticleCategoryMap()
	res := make([]TopicPayload, 0, len(topics))
	for _, topic := range topics {
		if topic == nil {
			continue
		}
		categories := make([]TopicCategoryPayload, 0, len(topic.CategoriesId))
		for i, categoryID := range topic.CategoriesId {
			name := ""
			color := "#9ca3af"
			if i < len(topic.Categories) {
				name = topic.Categories[i]
			}
			if category, ok := categoryMap[categoryID]; ok && category != nil {
				name = category.Category
				color = category.Color
			}
			if name == "" {
				continue
			}
			categories = append(categories, TopicCategoryPayload{
				ID:    categoryID,
				Name:  name,
				URL:   "/c/" + url.PathEscape(name) + "/" + strconv.FormatUint(categoryID, 10),
				Color: color,
			})
		}

		res = append(res, TopicPayload{
			ID:          topic.Id,
			Title:       topic.Title,
			Description: topic.Description,
			URL:         urlconfig.PostDetail(topic.Id),
			Author: TopicAuthorPayload{
				ID:        topic.AuthorId,
				Username:  topic.Username,
				AvatarURL: topic.AvatarUrl,
			},
			Participants:   buildParticipants(topic),
			Categories:     categories,
			ReplyCount:     topic.CommentCount,
			ViewCount:      topic.ViewCount,
			ActivityText:   topic.LastUpdateTime,
			LastUpdateTime: topic.LastUpdateTime,
		})
	}
	return res
}

func buildParticipants(topic *vo.ArticlesSimpleVo) []TopicAuthorPayload {
	participants := make([]TopicAuthorPayload, 0, len(topic.Posters)+1)
	seen := map[uint64]bool{}
	add := func(user TopicAuthorPayload) {
		if user.ID == 0 || seen[user.ID] {
			return
		}
		seen[user.ID] = true
		participants = append(participants, user)
	}
	for _, poster := range topic.Posters {
		add(TopicAuthorPayload{ID: poster.Id, Username: poster.Username, AvatarURL: poster.AvatarUrl})
	}
	add(TopicAuthorPayload{ID: topic.AuthorId, Username: topic.Username, AvatarURL: topic.AvatarUrl})
	if len(participants) > 4 {
		return participants[:4]
	}
	return participants
}

func buildHomePageURL(sort string, page int) string {
	if page <= 0 {
		return ""
	}
	values := url.Values{}
	if sort != "" && sort != "latest" {
		values.Set("sort", sort)
	}
	values.Set("page", strconv.Itoa(page))
	return "/?" + values.Encode()
}

func categoryURL(category *articleCategory.Entity) string {
	slug := category.Slug
	if slug == "" {
		slug = category.Category
	}
	return "/c/" + url.PathEscape(slug) + "/" + strconv.FormatUint(category.Id, 10)
}

func activeKeyForHome(sort string) string {
	switch sort {
	case "hot":
		return "hot"
	case "popular":
		return "popular"
	default:
		return "topics"
	}
}

func buildPageURL(c *gin.Context) string {
	return component.GetBaseUri(c) + c.Request.URL.String()
}

func buildHomeMeta(c *gin.Context) PageMeta {
	info := controllers.GetGooseForumInfo()
	return PageMeta{
		Title:       "GooseForum",
		Description: info.Desc,
		Canonical:   buildPageURL(c),
	}
}

func buildArticleDetailProps(c *gin.Context, entity *articles.Entity) ArticleDetailProps {
	currentUserID := component.LoginUserId(c)
	replyEntities := reply.GetByArticleId(entity.Id)
	userIDs := []uint64{entity.UserId}
	userIDs = append(userIDs, lo.Map(replyEntities, func(item *reply.Entity, _ int) uint64 { return item.UserId })...)
	userMap := users.GetMapByIds(lo.Uniq(userIDs))
	replyMap := lo.KeyBy(replyEntities, func(item *reply.Entity) uint64 { return item.Id })

	return ArticleDetailProps{
		Article: buildArticlePayload(c, entity, userMap),
		Replies: lo.Map(replyEntities, func(item *reply.Entity, _ int) ReplyPayload {
			author := userPayload(item.UserId, userMap)
			replyToName, replyToUserID := "", uint64(0)
			if item.ReplyId > 0 {
				if parent, ok := replyMap[item.ReplyId]; ok && parent != nil {
					parentAuthor := userPayload(parent.UserId, userMap)
					replyToName = parentAuthor.Username
					replyToUserID = parentAuthor.ID
				}
			}
			return ReplyPayload{
				ID:              item.Id,
				ArticleID:       item.ArticleId,
				Content:         item.Content,
				Author:          author,
				CreatedAt:       item.CreatedAt.Format(time.DateTime),
				ReplyToID:       item.ReplyId,
				ReplyToUserID:   replyToUserID,
				ReplyToUsername: replyToName,
				IsOwnReply:      currentUserID == item.UserId,
			}
		}),
		HotTopics: buildArticleHotTopics(entity.Id),
		Permissions: ArticlePermissions{
			IsOwnArticle: currentUserID == entity.UserId,
			CanReply:     currentUserID > 0,
		},
	}
}

func buildArticleHotTopics(currentArticleID uint64) []TopicPayload {
	topics := hotdataserve.GetLatestArticlesSimpleVoPaginated(1, "hot")
	filtered := make([]*vo.ArticlesSimpleVo, 0, 6)
	for _, topic := range topics {
		if topic == nil || topic.Id == currentArticleID {
			continue
		}
		filtered = append(filtered, topic)
		if len(filtered) >= 6 {
			break
		}
	}
	return buildTopicPayloads(filtered)
}

func buildArticlePayload(c *gin.Context, entity *articles.Entity, userMap map[uint64]*users.EntityComplete) ArticlePayload {
	participants := make([]TopicAuthorPayload, 0, len(userMap))
	seen := map[uint64]bool{}
	addParticipant := func(userID uint64) {
		if userID == 0 || seen[userID] {
			return
		}
		seen[userID] = true
		participants = append(participants, userPayload(userID, userMap))
	}
	addParticipant(entity.UserId)
	for userID := range userMap {
		addParticipant(userID)
	}
	if len(participants) > 12 {
		participants = participants[:12]
	}

	currentUserID := component.LoginUserId(c)
	isLiked := false
	isBookmarked := false
	if currentUserID > 0 {
		isLiked = articleLike.GetByArticleId(currentUserID, entity.Id).Status == 1
		isBookmarked = articleBookmark.GetByArticleId(currentUserID, entity.Id).Status == 1
	}

	return ArticlePayload{
		ID:           entity.Id,
		Title:        entity.Title,
		Description:  entity.Description,
		URL:          urlconfig.PostDetail(entity.Id),
		HTML:         entity.RenderedHTML,
		Author:       userPayload(entity.UserId, userMap),
		Participants: participants,
		Categories:   categoryPayloads(entity.CategoryId),
		ReplyCount:   entity.ReplyCount,
		ViewCount:    entity.ViewCount,
		LikeCount:    entity.LikeCount,
		IsLiked:      isLiked,
		IsBookmarked: isBookmarked,
		CreatedAt:    entity.CreatedAt.Format(time.DateTime),
		UpdatedAt:    entity.UpdatedAt.Format(time.DateTime),
	}
}

func categoryPayloads(ids []uint64) []TopicCategoryPayload {
	categoryMap := hotdataserve.ArticleCategoryMap()
	res := make([]TopicCategoryPayload, 0, len(ids))
	for _, id := range ids {
		category, ok := categoryMap[id]
		if !ok || category == nil {
			continue
		}
		res = append(res, TopicCategoryPayload{
			ID:    category.Id,
			Name:  category.Category,
			URL:   categoryURL(category),
			Color: category.Color,
		})
	}
	return res
}

func userPayload(userID uint64, userMap map[uint64]*users.EntityComplete) TopicAuthorPayload {
	user, ok := userMap[userID]
	if !ok || user == nil {
		return TopicAuthorPayload{ID: userID, Username: "匿名用户", AvatarURL: urlconfig.GetDefaultAvatar()}
	}
	return TopicAuthorPayload{ID: userID, Username: user.Username, AvatarURL: user.GetWebAvatarUrl()}
}

func ensureRenderedHTML(entity *articles.Entity) {
	if entity.RenderedVersion >= markdown2html.GetVersion() && entity.RenderedHTML != "" {
		return
	}
	entity.RenderedHTML = markdown2html.MarkdownToHTML(entity.Content)
	entity.RenderedVersion = markdown2html.GetVersion()
	_ = articles.SaveNoUpdate(entity)
}

func buildArticleMeta(c *gin.Context, article ArticlePayload) PageMeta {
	return PageMeta{
		Title:       article.Title + " - GooseForum",
		Description: article.Description,
		Canonical:   component.GetBaseUri(c) + article.URL,
	}
}

func buildUserProfileProps(c *gin.Context, user users.EntityComplete) UserProfileProps {
	currentUserID := component.LoginUserId(c)
	stats := userStatistics.Get(user.Id)
	isFollowing := userFollow.IsFollowing(currentUserID, user.Id)
	userCard := transform.User2UserCard(user, stats, isFollowing, currentUserID)

	latestArticles, _ := articles.GetLatestArticlesByUserId(user.Id, 8)
	activities, _ := userActivities.GetUserTimeline(user.Id, 0, 20)

	return UserProfileProps{
		User:         userCard,
		Topics:       buildTopicPayloads(hotdataserve.ArticlesSmallEntity2Vo(latestArticles)),
		Activities:   buildUserActivities(activities),
		Following:    buildUserConnections(userFollow.GetFollowingList(user.Id, 1, 12)),
		Followers:    buildUserConnections(userFollow.GetFollowerList(user.Id, 1, 12)),
		IsOwnProfile: currentUserID == user.Id,
		CanMessage:   currentUserID > 0 && currentUserID != user.Id,
		CanFollow:    currentUserID > 0 && currentUserID != user.Id,
		MessageURL:   "/messages?userId=" + strconv.FormatUint(user.Id, 10) + "&username=" + url.QueryEscape(userCard.Nickname) + "&avatar=" + url.QueryEscape(userCard.AvatarUrl),
		SettingsURL:  "/settings",
	}
}

func buildUserActivities(activities []*userActivities.Entity) []UserActivityPayload {
	res := make([]UserActivityPayload, 0, len(activities))
	for _, activity := range activities {
		if activity == nil {
			continue
		}
		res = append(res, UserActivityPayload{
			ID:             activity.Id,
			Action:         activity.Action,
			SubjectType:    activity.SubjectType,
			SubjectID:      activity.SubjectId,
			ContentPreview: activity.ContentPreview,
			URL:            userActivityURL(activity),
			Label:          userActivityLabel(activity.Action),
			CreatedAt:      activity.CreatedAt.Format(time.DateTime),
		})
	}
	return res
}

func userActivityURL(activity *userActivities.Entity) string {
	switch activity.SubjectType {
	case userActivities.SubjectTopic, userActivities.SubjectPost:
		if activity.SubjectId > 0 {
			return urlconfig.PostDetail(activity.SubjectId)
		}
	case userActivities.SubjectUser:
		if activity.SubjectId > 0 {
			return "/u/" + strconv.FormatUint(activity.SubjectId, 10)
		}
	}
	return ""
}

func userActivityLabel(action int) string {
	switch userActivities.ActionType(action) {
	case userActivities.ActionSignUp:
		return "加入论坛"
	case userActivities.ActionPost:
		return "发布主题"
	case userActivities.ActionLike:
		return "点赞内容"
	case userActivities.ActionFollow:
		return "关注用户"
	case userActivities.ActionComment:
		return "参与回复"
	default:
		return "活动"
	}
}

func buildUserConnections(list []*users.EntityComplete) []UserConnectionPayload {
	res := make([]UserConnectionPayload, 0, len(list))
	for _, user := range list {
		if user == nil || user.Id == 0 {
			continue
		}
		res = append(res, UserConnectionPayload{
			ID:        user.Id,
			Username:  user.Username,
			Nickname:  user.Nickname,
			AvatarURL: user.GetWebAvatarUrl(),
			Bio:       user.Bio,
			URL:       "/u/" + strconv.FormatUint(user.Id, 10),
		})
	}
	return res
}

func buildUserMeta(c *gin.Context, user *vo.UserCard) PageMeta {
	description := user.Bio
	if description == "" {
		description = user.Signature
	}
	if description == "" {
		description = "查看 " + user.Username + " 在 GooseForum 的主题、动态和社区关系。"
	}
	return PageMeta{
		Title:       user.Username + " - GooseForum",
		Description: description,
		Canonical:   component.GetBaseUri(c) + "/u/" + strconv.FormatUint(user.UserId, 10),
	}
}

func buildCategoryPageProps(category *articleCategory.Entity, page int, sort string, topics []*vo.ArticlesSimpleVo) CategoryPageProps {
	nextPage := 0
	if len(topics) == 20 {
		nextPage = page + 1
	}
	return CategoryPageProps{
		Category: CategoryHeaderPayload{
			ID:          category.Id,
			Name:        category.Category,
			Description: category.Desc,
			Icon:        category.Icon,
			Color:       category.Color,
			URL:         categoryURL(category),
		},
		Sort:   sort,
		Tabs:   buildCategoryTabs(category, sort),
		Topics: buildTopicPayloads(topics),
		Pagination: PaginationPayload{
			Page:     page,
			NextPage: nextPage,
			HasNext:  nextPage > 0,
			NextURL:  buildCategoryPageURL(category, sort, nextPage),
		},
	}
}

func buildCategoryTabs(category *articleCategory.Entity, sort string) []TabPayload {
	return []TabPayload{
		{Key: "latest", Label: "最新回复", URL: categorySortURL(category, "latest"), Active: sort == "" || sort == "latest"},
		{Key: "new", Label: "最新发布", URL: categorySortURL(category, "new"), Active: sort == "new"},
	}
}

func categorySortURL(category *articleCategory.Entity, sort string) string {
	if sort == "" || sort == "latest" {
		return categoryURL(category)
	}
	return categoryURL(category) + "/l/" + url.PathEscape(sort)
}

func buildCategoryPageURL(category *articleCategory.Entity, sort string, page int) string {
	if page <= 0 {
		return ""
	}
	values := url.Values{}
	values.Set("page", strconv.Itoa(page))
	return categorySortURL(category, sort) + "?" + values.Encode()
}

func buildCategoryMeta(c *gin.Context, category *articleCategory.Entity) PageMeta {
	description := category.Desc
	if description == "" {
		description = "浏览 " + category.Category + " 分类下的主题。"
	}
	return PageMeta{
		Title:       category.Category + " - GooseForum",
		Description: description,
		Canonical:   component.GetBaseUri(c) + categoryURL(category),
	}
}

func buildLinksPageProps(groups []pageConfig.FriendLinksGroup) LinksPageProps {
	res := make([]LinkGroupPayload, 0, len(groups))
	total := 0
	for _, group := range groups {
		links := make([]FriendLinkPayload, 0, len(group.Links))
		for _, link := range group.Links {
			if link.Status != 1 {
				continue
			}
			links = append(links, FriendLinkPayload{
				Name:    link.Name,
				Desc:    link.Desc,
				URL:     link.Url,
				LogoURL: link.LogoUrl,
			})
		}
		if len(links) == 0 {
			continue
		}
		total += len(links)
		res = append(res, LinkGroupPayload{
			Name:  group.Name,
			Emoji: group.Emoji,
			Color: group.Color,
			Links: links,
		})
	}
	return LinksPageProps{Groups: res, TotalCount: total}
}

func buildLinksMeta(c *gin.Context) PageMeta {
	return PageMeta{
		Title:       "友情链接 - GooseForum",
		Description: "GooseForum 友情链接与社区伙伴。",
		Canonical:   component.GetBaseUri(c) + "/links",
	}
}

func buildSponsorsPageProps(config pageConfig.SponsorsConfig) SponsorsPageProps {
	sections := []SponsorSectionPayload{
		{Key: "diamond", Label: "Diamond Partners", Tone: "diamond", Sponsors: buildSponsorPayloads(config.Sponsors.Level0)},
		{Key: "gold", Label: "Gold Sponsors", Tone: "gold", Sponsors: buildSponsorPayloads(config.Sponsors.Level1)},
		{Key: "silver", Label: "Silver Sponsors", Tone: "silver", Sponsors: buildSponsorPayloads(config.Sponsors.Level2)},
		{Key: "supporter", Label: "Supporters", Tone: "supporter", Sponsors: buildSponsorPayloads(config.Sponsors.Level3)},
	}
	visibleSections := make([]SponsorSectionPayload, 0, len(sections))
	total := 0
	for _, section := range sections {
		if len(section.Sponsors) == 0 {
			continue
		}
		total += len(section.Sponsors)
		visibleSections = append(visibleSections, section)
	}
	return SponsorsPageProps{
		Sections:   visibleSections,
		TotalCount: total,
	}
}

func buildSponsorPayloads(items []pageConfig.SponsorItem) []SponsorPayload {
	res := make([]SponsorPayload, 0, len(items))
	for _, item := range items {
		if item.Name == "" {
			continue
		}
		res = append(res, SponsorPayload{
			Name:      item.Name,
			Message:   item.Message,
			Link:      item.Link,
			AvatarURL: sponsorAvatar(item.AvatarUrl),
		})
	}
	return res
}

func sponsorAvatar(avatar string) string {
	if avatar != "" {
		return avatar
	}
	return "/static/pic/default-avatar.webp"
}

func buildSponsorsMeta(c *gin.Context) PageMeta {
	return PageMeta{
		Title:       "赞助 - GooseForum",
		Description: "感谢支持 GooseForum 的赞助者与社区伙伴。",
		Canonical:   component.GetBaseUri(c) + "/sponsors",
	}
}

func buildNotificationsPageProps(c *gin.Context) NotificationsPageProps {
	userID := component.LoginUserId(c)
	total, notifications := notificationservice.GetNotificationItemList(userID, 20, 0, false)
	unreadCount, _ := eventNotification.GetUnreadCount(userID)
	items := make([]NotificationPayload, 0, len(notifications))
	for _, notification := range notifications {
		if notification == nil {
			continue
		}
		items = append(items, buildNotificationPayload(notification))
	}
	return NotificationsPageProps{
		Total:         total,
		UnreadCount:   unreadCount,
		Notifications: items,
		Pagination: PaginationPayload{
			Page:     1,
			NextPage: lo.Ternary(len(notifications) >= 20, 2, 0),
			HasNext:  len(notifications) >= 20,
			NextURL:  "",
		},
	}
}

func buildNotificationPayload(notification *eventNotification.Entity) NotificationPayload {
	payload := notification.Payload
	item := NotificationPayload{
		ID:        notification.Id,
		EventType: notification.EventType,
		IsRead:    notification.IsRead,
		CreatedAt: notification.CreatedAt.Format(time.DateTime),
		Title:     notificationTitle(notification.EventType, payload),
		Content:   payload.Content,
		Actor: TopicAuthorPayload{
			ID:       payload.ActorId,
			Username: payload.ActorName,
		},
		Payload: payload,
	}
	if payload.ArticleId > 0 {
		item.Article = &NotificationArticlePayload{
			ID:    payload.ArticleId,
			Title: payload.ArticleTitle,
			URL:   urlconfig.PostDetail(payload.ArticleId),
		}
	}
	return item
}

func notificationTitle(eventType string, payload eventNotification.NotificationPayload) string {
	if payload.Title != "" {
		return payload.Title
	}
	switch eventType {
	case eventNotification.EventTypeComment:
		return "有人评论了你的主题"
	case eventNotification.EventTypeReply:
		return "有人回复了你"
	case eventNotification.EventTypeLike:
		return "有人喜欢了你的内容"
	case eventNotification.EventTypeFollow:
		return "有新的关注者"
	default:
		return "系统通知"
	}
}

func buildMessagesPageProps(c *gin.Context) MessagesPageProps {
	userID := component.LoginUserId(c)
	list, _ := chatservice.GetChatList(userID)
	conversations := make([]MessageConversationPayload, 0, len(list))
	for _, item := range list {
		if item == nil {
			continue
		}
		conversations = append(conversations, MessageConversationPayload{
			ID:           item.Id,
			PeerID:       item.PeerId,
			PeerUsername: item.PeerUsername,
			PeerAvatar:   item.PeerAvatar,
			LastMsg:      item.LastMsg,
			LastMsgTime:  item.LastMsgTime,
			UnreadCount:  item.UnreadCount,
			ConvID:       item.ConvId,
			PeerURL:      "/u/" + strconv.FormatUint(item.PeerId, 10),
		})
	}
	return MessagesPageProps{
		Conversations:  conversations,
		SuggestedUsers: buildUserConnections(userFollow.GetFollowingList(userID, 1, 12)),
	}
}

func buildSettingsPageProps(user users.EntityComplete) SettingsPageProps {
	stats := userStatistics.Get(user.Id)
	return SettingsPageProps{
		User: transform.User2UserDetailedVo(user),
		Stats: SettingsStatsPayload{
			ArticleCount:      stats.ArticleCount,
			ReplyCount:        stats.ReplyCount,
			FollowerCount:     stats.FollowerCount,
			FollowingCount:    stats.FollowingCount,
			LikeReceivedCount: stats.LikeReceivedCount,
			CreatedAt:         user.CreatedAt.Format(time.DateTime),
		},
		Tabs: []TabPayload{
			{Key: "profile", Label: "资料", URL: "/settings", Active: true},
			{Key: "account", Label: "账号", URL: "/settings?tab=account"},
			{Key: "privacy", Label: "隐私", URL: "/settings?tab=privacy"},
			{Key: "binding", Label: "绑定", URL: "/settings?tab=binding"},
		},
	}
}

func buildPublishPageProps(c *gin.Context, articleID uint64) (PublishPageProps, error) {
	props := PublishPageProps{
		ArticleID:  articleID,
		IsEditing:  articleID > 0,
		Categories: buildPublishCategories(),
		Types:      buildPublishTypes(),
		Article: PublishArticlePayload{
			Type: defaultPublishType(),
		},
	}
	if articleID == 0 {
		return props, nil
	}

	entity := articles.Get(articleID)
	if entity.Id == 0 || entity.UserId != component.LoginUserId(c) {
		return props, fmt.Errorf("article not found")
	}
	props.Article = PublishArticlePayload{
		Title:       entity.Title,
		Content:     entity.Content,
		Type:        entity.Type,
		CategoryIDs: entity.CategoryId,
	}
	return props, nil
}

func buildPublishCategories() []PublishCategoryPayload {
	categories := hotdataserve.GetArticleCategory()
	res := make([]PublishCategoryPayload, 0, len(categories))
	for _, category := range categories {
		if category == nil {
			continue
		}
		res = append(res, PublishCategoryPayload{
			ID:    category.Id,
			Name:  category.Category,
			Color: category.Color,
		})
	}
	return res
}

func buildPublishTypes() []PublishTypePayload {
	items := hotdataserve.GetArticlesType()
	res := make([]PublishTypePayload, 0, len(*items))
	for _, item := range *items {
		res = append(res, PublishTypePayload{Name: item.Name, Value: item.Value})
	}
	return res
}

func defaultPublishType() int8 {
	items := hotdataserve.GetArticlesType()
	if items != nil && len(*items) > 0 {
		return int8((*items)[0].Value)
	}
	return 0
}

func buildSearchPageProps(query string, page int) SearchPageProps {
	const pageSize = 10
	props := SearchPageProps{
		Query:  query,
		Topics: []TopicPayload{},
		Pagination: PaginationPayload{
			Page: page,
		},
	}
	if query == "" {
		return props
	}

	if page < 1 {
		page = 1
	}
	result, err := searchservice.SearchArticles(searchservice.SearchRequest{
		Query:  query,
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	})
	if err != nil || result == nil {
		return props
	}

	ids := lo.Map(result.Results, func(item searchservice.SearchResult, _ int) uint64 {
		return item.ID
	})
	articleMap := articles.GetMapByIds(ids)
	orderedArticles := lo.FilterMap(ids, func(id uint64, _ int) (*articles.SmallEntity, bool) {
		article, ok := articleMap[id]
		return article, ok && article != nil
	})
	totalPageCount := totalPages(result.Total, pageSize)
	nextPage := 0
	if page < totalPageCount {
		nextPage = page + 1
	}

	props.Topics = buildTopicPayloads(hotdataserve.ArticlesSmallEntity2Vo(orderedArticles))
	props.Total = result.Total
	props.TotalPages = totalPageCount
	props.Pagination = PaginationPayload{
		Page:     page,
		NextPage: nextPage,
		HasNext:  nextPage > 0,
		NextURL:  buildSearchURL(query, nextPage),
	}
	return props
}

func buildSearchURL(query string, page int) string {
	values := url.Values{}
	if query != "" {
		values.Set("q", query)
	}
	if page > 1 {
		values.Set("page", strconv.Itoa(page))
	}
	encoded := values.Encode()
	if encoded == "" {
		return "/search"
	}
	return "/search?" + encoded
}

func parsePositiveInt(value string, fallback int) int {
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}
