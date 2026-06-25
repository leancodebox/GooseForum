package forum

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/articleviewservice"
	"github.com/leancodebox/GooseForum/app/service/moderatorservice"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/leancodebox/GooseForum/app/service/userservice"
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
	if !canViewArticle(&entity, loginUser.UserId) {
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

	renderPage(c, "article.gohtml", payload)
	if shouldCountArticleView(&entity) {
		articleviewservice.RecordView(entity.Id)
	}
}

type ArticleRepliesWindowReq struct {
	ArticleID     uint64 `form:"articleId"`
	AnchorReplyID uint64 `form:"anchorReplyId"`
	AnchorReplyNo uint64 `form:"anchorReplyNo"`
	Before        uint64 `form:"before"`
	After         uint64 `form:"after"`
	BeforeReplyNo uint64 `form:"beforeReplyNo"`
	AfterReplyNo  uint64 `form:"afterReplyNo"`
	Limit         int    `form:"limit"`
	Tail          bool   `form:"tail"`
}

func ArticleRepliesWindow(req component.BetterRequest[ArticleRepliesWindowReq]) component.Response {
	articleID := req.Params.ArticleID
	if articleID == 0 {
		return component.FailResponseCode(component.MessageArticleNotFound, nil)
	}

	articleEntity := articles.GetSimple(articleID)
	if articleEntity.Id == 0 {
		return component.FailResponseCode(component.MessageArticleNotFound, nil)
	}
	if !canViewArticleSimple(&articleEntity, req.UserId) {
		return component.FailResponseCode(component.MessageArticleNotFound, nil)
	}

	limit := req.Params.Limit
	if limit <= 0 || limit > 50 {
		limit = replyWindowLimit
	}

	var replyEntities []*reply.Entity
	hasBefore := false
	var hasAfter bool

	switch {
	case req.Params.AnchorReplyNo > 0:
		anchor, ok := reply.GetByArticleReplyNoAtOrAfter(articleID, req.Params.AnchorReplyNo)
		if !ok {
			anchor, ok = reply.GetByArticleReplyNoAtOrBefore(articleID, req.Params.AnchorReplyNo)
		}
		if !ok || anchor.Id == 0 || anchor.ArticleId != articleID {
			return component.FailResponseCode(component.MessageReplyNotFound, nil)
		}
		beforeLimit := min(5, limit/2)
		afterLimit := limit - beforeLimit - 1
		beforeReplies := reply.GetByArticleReplyNoBefore(articleID, anchor.ReplyNo, beforeLimit+1)
		afterReplies := reply.GetByArticleReplyNoAfter(articleID, anchor.ReplyNo, afterLimit+1)
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
	case req.Params.Tail:
		replyEntities = reply.GetByArticleReplyNoDesc(articleID, limit+1)
		hasBefore = len(replyEntities) > limit
		if hasBefore {
			replyEntities = replyEntities[1:]
		}
	case req.Params.AnchorReplyID > 0:
		anchor := reply.Get(req.Params.AnchorReplyID)
		if anchor.Id == 0 || anchor.ArticleId != articleID {
			return component.FailResponseCode(component.MessageReplyNotFound, nil)
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
	case req.Params.BeforeReplyNo > 0:
		replyEntities = reply.GetByArticleReplyNoBefore(articleID, req.Params.BeforeReplyNo, limit+1)
		hasBefore = len(replyEntities) > limit
		if hasBefore {
			replyEntities = replyEntities[1:]
		}
		hasAfter = true
	case req.Params.AfterReplyNo > 0:
		replyEntities = reply.GetByArticleReplyNoAfter(articleID, req.Params.AfterReplyNo, limit+1)
		hasAfter = len(replyEntities) > limit
		if hasAfter {
			replyEntities = replyEntities[:limit]
		}
		hasBefore = true
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
		replyEntities = reply.GetByArticleReplyNoAsc(articleID, limit+1)
		hasAfter = len(replyEntities) > limit
		if hasAfter {
			replyEntities = replyEntities[:limit]
		}
	}

	userIDs := make([]uint64, 0, len(replyEntities))
	seenUserIDs := make(map[uint64]struct{}, len(replyEntities))
	for _, item := range replyEntities {
		if item == nil {
			continue
		}
		if _, seen := seenUserIDs[item.UserId]; seen {
			continue
		}
		seenUserIDs[item.UserId] = struct{}{}
		userIDs = append(userIDs, item.UserId)
	}
	userMap := users.GetMapByIds(userIDs)
	canModerateReplies := moderatorservice.CanModerateAnyCategory(req.UserId, articleEntity.CategoryId)
	payloadReplies := buildReplyPayloads(replyEntities, userMap, req.UserId, canModerateReplies)

	var beforeCursor uint64
	var afterCursor uint64
	var beforeReplyNo uint64
	var afterReplyNo uint64
	if len(replyEntities) > 0 {
		beforeCursor = replyEntities[0].Id
		afterCursor = replyEntities[len(replyEntities)-1].Id
		beforeReplyNo = replyEntities[0].ReplyNo
		afterReplyNo = replyEntities[len(replyEntities)-1].ReplyNo
	}
	maxReplyNo := articleEntity.ReplySeq
	if maxReplyNo == 0 && articleEntity.ReplyCount > 0 {
		maxReplyNo = reply.GetMaxReplyNoByArticleId(articleID)
	}

	return component.SuccessResponse(ReplyWindowPayload{
		Replies:       payloadReplies,
		AnchorReplyID: req.Params.AnchorReplyID,
		BeforeCursor:  beforeCursor,
		AfterCursor:   afterCursor,
		BeforeReplyNo: beforeReplyNo,
		AfterReplyNo:  afterReplyNo,
		HasBefore:     hasBefore,
		HasAfter:      hasAfter,
		Total:         int64(articleEntity.ReplyCount),
		MaxReplyNo:    maxReplyNo,
	})
}

func canViewArticle(entity *articles.Entity, userID uint64) bool {
	if entity.ArticleStatus != 1 {
		return userID != 0 && userID == entity.UserId
	}
	if entity.ProcessStatus != 0 && !currentUserCanViewProcessedArticle(userID) && !moderatorservice.CanModerateAnyCategory(userID, entity.CategoryId) {
		return false
	}
	return true
}

func canViewArticleSimple(entity *articles.SmallEntity, userID uint64) bool {
	if entity.ArticleStatus != 1 {
		return userID != 0 && userID == entity.UserId
	}
	if entity.ProcessStatus != 0 && !currentUserCanViewProcessedArticle(userID) && !moderatorservice.CanModerateAnyCategory(userID, entity.CategoryId) {
		return false
	}
	return true
}

func currentUserCanViewProcessedArticle(userID uint64) bool {
	if userID == 0 {
		return false
	}
	roleID, ok := userservice.GetUserRoleId(userID)
	return ok && permission.CheckRole(roleID, permission.ArticlesManager)
}

func shouldCountArticleView(entity *articles.Entity) bool {
	return entity.ArticleStatus == 1 && entity.ProcessStatus == 0
}

func renderNotFound(c *gin.Context) {
	renderNotFoundWithMessage(c, component.MessagePageNotFound)
}

func renderNotFoundWithMessage(c *gin.Context, messageCode component.MessageCode) {
	payload := PagePayload{
		Component: "error.notFound",
		Props: ErrorPageProps{
			Code:        "404",
			Title:       "页面不存在",
			MessageCode: messageCode,
		},
		Meta: PageMeta{
			Title:  pageTitle("页面不存在"),
			Robots: "noindex",
		},
		Layout:  buildLayout(c, "topics"),
		URL:     buildPageURL(c),
		Version: payloadVersion,
	}

	renderPageWithStatus(c, http.StatusNotFound, "error.gohtml", payload)
}

func activeKeyForArticle(article ArticlePayload) string {
	if len(article.Categories) > 0 {
		return "category_" + cast.ToString(article.Categories[0].ID)
	}
	return "topics"
}
