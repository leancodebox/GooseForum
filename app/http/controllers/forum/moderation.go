package forum

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/moderationLog"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/moderationlogservice"
	"github.com/leancodebox/GooseForum/app/service/moderatorservice"
	"github.com/leancodebox/GooseForum/app/service/searchservice"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"
)

const moderationPageSize = 20

func Moderation(c *gin.Context) {
	userID := component.LoginUserId(c)
	if !moderatorservice.CanAccessModeration(userID) {
		renderNotFound(c)
		return
	}
	page := moderationPage(c)
	global, categoryIDs := moderatorservice.ScopeForUser(userID)
	if !global && len(categoryIDs) == 0 {
		renderNotFound(c)
		return
	}
	availableCategories := moderationCategories(global, categoryIDs)
	if len(availableCategories) == 0 {
		renderNotFound(c)
		return
	}
	categoryID := moderationCategoryID(c, availableCategories)

	pageData := articles.PageForModeration(articles.ModerationPageQuery{
		Page:                page,
		PageSize:            moderationPageSize,
		FilterProcessStatus: true,
		ProcessStatus:       1,
		CategoryIDs:         []uint64{categoryID},
	})
	hasNext := pageData.HasNext
	nextPage := 0
	nextURL := ""
	if hasNext {
		nextPage = pageData.Page + 1
		nextURL = buildModerationPageURL(categoryID, nextPage)
	}
	payload := PagePayload{
		Component: "moderation.index",
		Props: ModerationPageProps{
			CategoryTabs: buildModerationCategoryTabs(availableCategories, categoryID),
			Topics:       buildTopicPayloads(hotdataserve.ArticlesSmallEntity2Vo(moderationEntityPointers(pageData.Data))),
			Pagination: PaginationPayload{
				Page:     pageData.Page,
				NextPage: nextPage,
				HasNext:  hasNext,
				NextURL:  nextURL,
			},
		},
		Meta: PageMeta{
			Title:       pageTitle("版主管理"),
			Description: "处理你负责范围内的帖子封禁与解封。",
			Robots:      "noindex",
		},
		Layout:  buildLayout(c, "moderation"),
		URL:     buildPageURL(c),
		Version: payloadVersion,
	}
	c.Header("Vary", "X-Goose-Page, Accept")
	c.Status(http.StatusOK)
	renderPage(c, "home.gohtml", payload)
}

func moderationPage(c *gin.Context) int {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		return 1
	}
	return page
}

type ModerationArticleStatusReq struct {
	Id     uint64 `json:"id" validate:"required"`
	Action string `json:"action" validate:"oneof=ban unban"`
}

type ModerationLogListReq struct {
	Cursor   uint64 `json:"cursor"`
	PageSize int    `json:"pageSize"`
}

type ModerationLogListResponse struct {
	Items      []ModerationLogItem `json:"items"`
	NextCursor uint64              `json:"nextCursor"`
	HasNext    bool                `json:"hasNext"`
}

type ModerationLogItem struct {
	ID          uint64                 `json:"id"`
	Action      string                 `json:"action"`
	Actor       TopicAuthorPayload     `json:"actor"`
	Subject     ModerationLogSubject   `json:"subject"`
	Categories  []TopicCategoryPayload `json:"categories"`
	MessageCode string                 `json:"messageCode"`
	Params      map[string]any         `json:"params"`
	CreatedAt   string                 `json:"createdAt"`
}

type ModerationLogSubject struct {
	Type    string `json:"type"`
	ID      uint64 `json:"id"`
	Title   string `json:"title"`
	URL     string `json:"url,omitempty"`
	Excerpt string `json:"excerpt,omitempty"`
}

func UpdateModerationArticleStatus(req component.BetterRequest[ModerationArticleStatusReq]) component.Response {
	article := articles.Get(req.Params.Id)
	if article.Id == 0 {
		return component.FailResponseCode(component.MessageArticleNotFound, nil)
	}
	if !moderatorservice.CanModerateAnyCategory(req.UserId, article.CategoryId) {
		return component.FailResponseCode(component.MessagePermissionDenied, nil)
	}
	nextStatus := moderationTargetStatus(req.Params.Action)
	if article.ProcessStatus == nextStatus {
		return component.SuccessResponse(true)
	}
	if err := articles.UpdateProcessStatus(article.Id, nextStatus); err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	article.ProcessStatus = nextStatus
	hotdataserve.ClearArticleListCache()
	if _, err := searchservice.BuildSingleArticleSearchDocument(&article); err != nil {
		slog.Error("failed to rebuild article search document", "articleId", article.Id, "err", err)
	}
	moderationlogservice.ArticleStatusChanged(req.UserId, article.Id, article.Title, nextStatus == 1)
	return component.SuccessResponse(true)
}

func ModerationLogList(req component.BetterRequest[ModerationLogListReq]) component.Response {
	if !moderatorservice.CanAccessModeration(req.UserId) {
		return component.FailResponseCode(component.MessagePermissionDenied, nil)
	}
	pageSize := component.BoundPageSizeWithRange(req.Params.PageSize, 10, 50)
	records := moderationLog.CursorPage(moderationLog.CursorPageQuery{
		Cursor:   req.Params.Cursor,
		PageSize: uint64(pageSize + 1),
	})
	hasNext := len(records) > pageSize
	if hasNext {
		records = records[:pageSize]
	}
	nextCursor := uint64(0)
	if hasNext && len(records) > 0 {
		nextCursor = records[len(records)-1].Id
	}
	return component.SuccessResponse(ModerationLogListResponse{
		Items:      buildModerationLogItems(records),
		NextCursor: nextCursor,
		HasNext:    hasNext,
	})
}

func moderationTargetStatus(actionType string) int8 {
	if actionType == "unban" {
		return 0
	}
	return 1
}

func moderationCategoryID(c *gin.Context, categories []TopicCategoryPayload) uint64 {
	requested, err := strconv.ParseUint(c.Query("category"), 10, 64)
	if err == nil && requested > 0 {
		for _, category := range categories {
			if category.ID == requested {
				return requested
			}
		}
	}
	return categories[0].ID
}

func buildModerationCategoryTabs(categories []TopicCategoryPayload, activeID uint64) []TabPayload {
	tabs := make([]TabPayload, 0, len(categories))
	for _, category := range categories {
		tabs = append(tabs, TabPayload{
			Key:    strconv.FormatUint(category.ID, 10),
			Label:  category.Name,
			URL:    buildModerationPageURL(category.ID, 1),
			Active: category.ID == activeID,
		})
	}
	return tabs
}

func buildModerationPageURL(categoryID uint64, page int) string {
	values := url.Values{}
	if categoryID > 0 {
		values.Set("category", strconv.FormatUint(categoryID, 10))
	}
	if page > 1 {
		values.Set("page", strconv.Itoa(page))
	}
	if encoded := values.Encode(); encoded != "" {
		return "/moderation?" + encoded
	}
	return "/moderation"
}

func moderationCategories(global bool, categoryIDs []uint64) []TopicCategoryPayload {
	categories := hotdataserve.GetArticleCategory()
	allowed := map[uint64]bool{}
	if global {
		for _, category := range categories {
			if category != nil {
				allowed[category.Id] = true
			}
		}
	} else {
		for _, categoryID := range categoryIDs {
			allowed[categoryID] = true
		}
	}
	res := make([]TopicCategoryPayload, 0, len(allowed))
	for _, category := range categories {
		if category == nil || !allowed[category.Id] {
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

func moderationEntityPointers(data []articles.SmallEntity) []*articles.SmallEntity {
	res := make([]*articles.SmallEntity, 0, len(data))
	for i := range data {
		res = append(res, &data[i])
	}
	return res
}

func buildModerationLogItems(records []moderationLog.Entity) []ModerationLogItem {
	if len(records) == 0 {
		return []ModerationLogItem{}
	}
	actorIDs := make([]uint64, 0, len(records))
	articleIDs := make([]uint64, 0, len(records))
	for _, record := range records {
		actorIDs = appendUniqueUint64(actorIDs, record.ActorUserId)
		if record.SubjectType != moderationLog.SubjectArticle {
			continue
		}
		articleIDs = appendUniqueUint64(articleIDs, record.SubjectId)
	}
	userMap := users.GetMapByIds(actorIDs)
	articleMap := articles.GetMapByIds(articleIDs)
	items := make([]ModerationLogItem, 0, len(records))
	for _, record := range records {
		payload := record.Payload
		if payload.Params == nil {
			payload.Params = map[string]any{}
		}
		subject := moderationLogSubject(record, payload.Params, articleMap)
		items = append(items, ModerationLogItem{
			ID:          record.Id,
			Action:      record.Action,
			Actor:       userPayload(record.ActorUserId, userMap),
			Subject:     subject,
			Categories:  moderationLogCategories(subject.ID, articleMap),
			MessageCode: payload.MessageCode,
			Params:      payload.Params,
			CreatedAt:   record.CreatedAt.Format(time.DateTime),
		})
	}
	return items
}

func moderationLogSubject(record moderationLog.Entity, params map[string]any, articleMap map[uint64]*articles.SmallEntity) ModerationLogSubject {
	switch record.SubjectType {
	case moderationLog.SubjectArticle:
		subject := ModerationLogSubject{Type: record.SubjectType, ID: record.SubjectId, Title: fmt.Sprint(params["title"])}
		if article := articleMap[record.SubjectId]; article != nil {
			subject.Title = article.Title
			subject.URL = urlconfig.PostDetail(article.Id)
			subject.Excerpt = article.Description
		} else if subject.ID > 0 {
			subject.URL = urlconfig.PostDetail(subject.ID)
		}
		if subject.Title == "" || subject.Title == "<nil>" {
			subject.Title = fmt.Sprintf("#%d", subject.ID)
		}
		return subject
	case moderationLog.SubjectCategory:
		title := fmt.Sprint(params["categoryName"])
		if title == "" || title == "<nil>" {
			title = fmt.Sprintf("#%d", record.SubjectId)
		}
		return ModerationLogSubject{Type: record.SubjectType, ID: record.SubjectId, Title: title}
	case moderationLog.SubjectUser:
		title := fmt.Sprint(params["username"])
		if title == "" || title == "<nil>" {
			title = fmt.Sprintf("#%d", record.SubjectId)
		}
		return ModerationLogSubject{Type: record.SubjectType, ID: record.SubjectId, Title: title}
	default:
		return ModerationLogSubject{Type: moderationLog.SubjectSystem, ID: record.SubjectId, Title: fmt.Sprintf("#%d", record.SubjectId)}
	}
}

func moderationLogCategories(articleID uint64, articleMap map[uint64]*articles.SmallEntity) []TopicCategoryPayload {
	article := articleMap[articleID]
	if article == nil {
		return []TopicCategoryPayload{}
	}
	return categoryPayloads(article.CategoryId)
}

func appendUniqueUint64(items []uint64, item uint64) []uint64 {
	if item == 0 {
		return items
	}
	if slices.Contains(items, item) {
		return items
	}
	return append(items, item)
}
