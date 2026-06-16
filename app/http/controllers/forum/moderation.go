package forum

import (
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/moderatorservice"
	"github.com/leancodebox/GooseForum/app/service/optlogger"
	"github.com/leancodebox/GooseForum/app/service/searchservice"
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
	statusCode := "unblocked"
	if nextStatus == 1 {
		statusCode = "blocked"
	}
	optlogger.UserOptCode(req.UserId, optlogger.EditArticle, article.Id, "moderator.opt.article.statusChanged", optlogger.MessageParams{
		"title":  article.Title,
		"status": statusCode,
	})
	return component.SuccessResponse(true)
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
