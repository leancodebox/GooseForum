package forum

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

const replyWindowLimit = 20

func ArticleDetail(c *gin.Context) {
	id := cast.ToUint64(c.Param("id"))
	if id == 0 {
		renderNotFound(c)
		return
	}

	entity := articles.Get(id)
	if entity.Id == 0 {
		renderNotFound(c)
		return
	}
	loginUser := component.GetLoginUser(c)
	if !canViewArticle(&entity, loginUser.UserId, loginUser.IsAdmin) {
		renderNotFound(c)
		return
	}

	ensureRenderedHTML(&entity)
	props := buildArticleDetailProps(c, &entity)
	payload := PagePayload{
		Component: "article.detail",
		Props:     props,
		Meta:      buildArticleMeta(c, props.Article),
		Layout:    buildLayout(c, activeKeyForArticle(props.Article)),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	if entity.ArticleStatus != 1 {
		payload.Meta.Robots = "noindex"
	}

	c.Header("Vary", "X-Goose-Page, Accept")
	c.Status(http.StatusOK)
	renderPage(c, "article.gohtml", payload)
	if shouldCountArticleView(&entity) {
		articles.IncrementView(entity)
	}
}

type ArticleRepliesWindowReq struct {
	ArticleID     uint64 `form:"articleId"`
	AnchorReplyID uint64 `form:"anchorReplyId"`
	Before        uint64 `form:"before"`
	After         uint64 `form:"after"`
	Limit         int    `form:"limit"`
}

func ArticleRepliesWindow(req component.BetterRequest[ArticleRepliesWindowReq]) component.Response {
	articleID := req.Params.ArticleID
	if articleID == 0 {
		return component.FailResponse("文章不存在")
	}

	articleEntity := articles.GetSimple(articleID)
	if articleEntity.Id == 0 {
		return component.FailResponse("文章不存在")
	}
	if !canViewArticleSimple(&articleEntity, req.UserId, false) {
		return component.FailResponse("文章不存在")
	}

	limit := req.Params.Limit
	if limit <= 0 || limit > 50 {
		limit = replyWindowLimit
	}

	var replyEntities []*reply.Entity
	hasBefore := false
	hasAfter := false

	switch {
	case req.Params.AnchorReplyID > 0:
		anchor := reply.Get(req.Params.AnchorReplyID)
		if anchor.Id == 0 || anchor.ArticleId != articleID {
			return component.FailResponse("回复不存在")
		}
		beforeLimit := min(5, limit/2)
		afterLimit := limit - beforeLimit - 1
		beforeReplies := reply.GetByArticleIdBefore(articleID, anchor.Id, beforeLimit+1)
		afterReplies := reply.GetByArticleIdAfter(articleID, anchor.Id, afterLimit+1)
		hasBefore = len(beforeReplies) > beforeLimit
		hasAfter = len(afterReplies) > afterLimit
		if hasBefore {
			beforeReplies = beforeReplies[1:]
		}
		if hasAfter {
			afterReplies = afterReplies[:afterLimit]
		}
		replyEntities = append(replyEntities, beforeReplies...)
		replyEntities = append(replyEntities, &anchor)
		replyEntities = append(replyEntities, afterReplies...)
	case req.Params.Before > 0:
		replyEntities = reply.GetByArticleIdBefore(articleID, req.Params.Before, limit+1)
		hasBefore = len(replyEntities) > limit
		if hasBefore {
			replyEntities = replyEntities[1:]
		}
		hasAfter = true
	case req.Params.After > 0:
		replyEntities = reply.GetByArticleIdAfter(articleID, req.Params.After, limit+1)
		hasAfter = len(replyEntities) > limit
		if hasAfter {
			replyEntities = replyEntities[:limit]
		}
		hasBefore = true
	default:
		replyEntities = reply.GetByArticleIdAsc(articleID, limit+1)
		hasAfter = len(replyEntities) > limit
		if hasAfter {
			replyEntities = replyEntities[:limit]
		}
	}

	userIDs := lo.Map(replyEntities, func(item *reply.Entity, _ int) uint64 {
		return item.UserId
	})
	userMap := users.GetMapByIds(lo.Uniq(userIDs))
	payloadReplies := buildReplyPayloads(replyEntities, userMap, req.UserId)

	var beforeCursor uint64
	var afterCursor uint64
	if len(replyEntities) > 0 {
		beforeCursor = replyEntities[0].Id
		afterCursor = replyEntities[len(replyEntities)-1].Id
	}

	return component.SuccessResponse(ReplyWindowPayload{
		Replies:       payloadReplies,
		AnchorReplyID: req.Params.AnchorReplyID,
		BeforeCursor:  beforeCursor,
		AfterCursor:   afterCursor,
		HasBefore:     hasBefore,
		HasAfter:      hasAfter,
		Total:         reply.CountByArticleId(articleID),
	})
}

func canViewArticle(entity *articles.Entity, userID uint64, isAdmin bool) bool {
	if entity.ArticleStatus != 1 {
		return userID != 0 && userID == entity.UserId
	}
	if entity.ProcessStatus != 0 && !isAdmin {
		return false
	}
	return true
}

func canViewArticleSimple(entity *articles.SmallEntity, userID uint64, isAdmin bool) bool {
	if entity.ArticleStatus != 1 {
		return userID != 0 && userID == entity.UserId
	}
	if entity.ProcessStatus != 0 && !isAdmin {
		return false
	}
	return true
}

func shouldCountArticleView(entity *articles.Entity) bool {
	return entity.ArticleStatus == 1 && entity.ProcessStatus == 0
}

func renderNotFound(c *gin.Context) {
	renderNotFoundWithMessage(c, "这个主题不存在，或已经被删除。")
}

func renderNotFoundWithMessage(c *gin.Context, message string) {
	payload := PagePayload{
		Component: "error.notFound",
		Props: ErrorPageProps{
			Code:    "404",
			Title:   "页面不存在",
			Message: message,
		},
		Meta: PageMeta{
			Title:  pageTitle("页面不存在"),
			Robots: "noindex",
		},
		Layout:  buildLayout(c, "topics"),
		URL:     buildPageURL(c),
		Version: payloadVersion,
	}

	c.Header("Vary", "X-Goose-Page, Accept")
	c.Status(http.StatusNotFound)
	renderPage(c, "error.gohtml", payload)
}

func activeKeyForArticle(article ArticlePayload) string {
	if len(article.Categories) > 0 {
		return "category_" + cast.ToString(article.Categories[0].ID)
	}
	return "topics"
}
