package forum

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/i18n"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
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

	topic := topics.Get(id)
	if topic.Id == 0 {
		renderNotFound(c)
		return
	}
	loginUser := component.GetLoginUser(c)
	if !canViewTopic(&topic, loginUser.UserId) {
		renderNotFound(c)
		return
	}

	firstPost := posts.Get(topic.FirstPostId)
	if firstPost.Id == 0 {
		firstPost, _ = posts.GetByTopicPostNoAtOrAfter(topic.Id, 1)
	}
	ensurePostRenderedHTML(&firstPost)
	props := buildTopicDetailProps(c, &topic, &firstPost)
	payload := PagePayload{
		Component: "article.detail",
		Props:     props,
		Meta:      buildArticleMeta(c, props.Article),
		Layout:    buildLayout(c, activeKeyForArticle(props.Article)),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	if topic.Status != 1 {
		payload.Meta.Robots = "noindex"
	}

	renderPage(c, "article.gohtml", payload)
	if shouldCountTopicView(&topic) {
		articleviewservice.RecordView(topic.Id)
	}
}

type ArticleRepliesWindowReq struct {
	TopicID      uint64 `form:"topicId"`
	AnchorPostID uint64 `form:"anchorPostId"`
	AnchorPostNo uint64 `form:"anchorPostNo"`
	Before       uint64 `form:"before"`
	After        uint64 `form:"after"`
	BeforePostNo uint64 `form:"beforePostNo"`
	AfterPostNo  uint64 `form:"afterPostNo"`
	Limit        int    `form:"limit"`
	Tail         bool   `form:"tail"`
}

func ArticleRepliesWindow(req component.BetterRequest[ArticleRepliesWindowReq]) component.Response {
	topicID := req.Params.TopicID
	if topicID == 0 {
		return component.FailResponseCode(component.MessageArticleNotFound, nil)
	}

	topicEntity := topics.GetSimple(topicID)
	if topicEntity.Id == 0 {
		return component.FailResponseCode(component.MessageArticleNotFound, nil)
	}
	if !canViewTopicSimple(&topicEntity, req.UserId) {
		return component.FailResponseCode(component.MessageArticleNotFound, nil)
	}

	limit := req.Params.Limit
	if limit <= 0 || limit > 50 {
		limit = replyWindowLimit
	}

	var postEntities []*posts.Entity
	hasBefore := false
	var hasAfter bool

	switch {
	case req.Params.AnchorPostNo > 0:
		anchor, ok := posts.GetByTopicPostNoAtOrAfter(topicID, req.Params.AnchorPostNo+1)
		if !ok {
			anchor, ok = posts.GetByTopicPostNoAtOrBefore(topicID, req.Params.AnchorPostNo+1)
		}
		if !ok || anchor.Id == 0 || anchor.TopicId != topicID || anchor.PostNo <= 1 {
			return component.FailResponseCode(component.MessageReplyNotFound, nil)
		}
		beforeLimit := min(5, limit/2)
		afterLimit := limit - beforeLimit - 1
		beforeReplies := posts.GetByTopicPostNoBefore(topicID, anchor.PostNo, beforeLimit+1)
		afterReplies := posts.GetByTopicPostNoAfter(topicID, anchor.PostNo, afterLimit+1)
		hasBefore = len(beforeReplies) > beforeLimit
		hasAfter = len(afterReplies) > afterLimit
		if hasBefore {
			beforeReplies = beforeReplies[1:]
		}
		if hasAfter {
			afterReplies = afterReplies[:afterLimit]
		}
		postEntities = append(postEntities, beforeReplies...)
		postEntities = append(postEntities, &anchor)
		postEntities = append(postEntities, afterReplies...)
	case req.Params.Tail:
		postEntities = posts.GetByTopicPostNoDesc(topicID, limit+1)
		hasBefore = len(postEntities) > limit
		if hasBefore {
			postEntities = postEntities[1:]
		}
	case req.Params.AnchorPostID > 0:
		anchor := posts.Get(req.Params.AnchorPostID)
		if anchor.Id == 0 || anchor.TopicId != topicID || anchor.PostNo <= 1 {
			return component.FailResponseCode(component.MessageReplyNotFound, nil)
		}
		beforeLimit := min(5, limit/2)
		afterLimit := limit - beforeLimit - 1
		beforeReplies := posts.GetByTopicIdBefore(topicID, anchor.Id, beforeLimit+1)
		afterReplies := posts.GetByTopicIdAfter(topicID, anchor.Id, afterLimit+1)
		hasBefore = len(beforeReplies) > beforeLimit
		hasAfter = len(afterReplies) > afterLimit
		if hasBefore {
			beforeReplies = beforeReplies[1:]
		}
		if hasAfter {
			afterReplies = afterReplies[:afterLimit]
		}
		postEntities = append(postEntities, beforeReplies...)
		postEntities = append(postEntities, &anchor)
		postEntities = append(postEntities, afterReplies...)
	case req.Params.BeforePostNo > 0:
		postEntities = posts.GetByTopicPostNoBefore(topicID, req.Params.BeforePostNo+1, limit+1)
		hasBefore = len(postEntities) > limit
		if hasBefore {
			postEntities = postEntities[1:]
		}
		hasAfter = true
	case req.Params.AfterPostNo > 0:
		postEntities = posts.GetByTopicPostNoAfter(topicID, req.Params.AfterPostNo+1, limit+1)
		hasAfter = len(postEntities) > limit
		if hasAfter {
			postEntities = postEntities[:limit]
		}
		hasBefore = true
	case req.Params.Before > 0:
		postEntities = posts.GetByTopicIdBefore(topicID, req.Params.Before, limit+1)
		hasBefore = len(postEntities) > limit
		if hasBefore {
			postEntities = postEntities[1:]
		}
		hasAfter = true
	case req.Params.After > 0:
		postEntities = posts.GetByTopicIdAfter(topicID, req.Params.After, limit+1)
		hasAfter = len(postEntities) > limit
		if hasAfter {
			postEntities = postEntities[:limit]
		}
		hasBefore = true
	default:
		postEntities = posts.GetByTopicPostNoAfter(topicID, 1, limit+1)
		hasAfter = len(postEntities) > limit
		if hasAfter {
			postEntities = postEntities[:limit]
		}
	}
	postEntities = filterReplyPosts(postEntities)

	userIDs := make([]uint64, 0, len(postEntities))
	seenUserIDs := make(map[uint64]struct{}, len(postEntities))
	for _, item := range postEntities {
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
	canModerateReplies := moderatorservice.CanModerateAnyCategory(req.UserId, topicEntity.CategoryIds)
	payloadReplies := buildPostPayloads(postEntities, userMap, req.UserId, canModerateReplies)

	var beforeCursor uint64
	var afterCursor uint64
	var beforeReplyNo uint64
	var afterReplyNo uint64
	if len(postEntities) > 0 {
		beforeCursor = postEntities[0].Id
		afterCursor = postEntities[len(postEntities)-1].Id
		beforeReplyNo = postEntities[0].PostNo - 1
		afterReplyNo = postEntities[len(postEntities)-1].PostNo - 1
	}
	maxReplyNo := uint64(0)
	if topicEntity.PostSeq > 0 {
		maxReplyNo = topicEntity.PostSeq - 1
	}
	if maxReplyNo == 0 && topicEntity.ReplyCount > 0 {
		maxPostNo := posts.GetMaxPostNoByTopicId(topicID)
		if maxPostNo > 0 {
			maxReplyNo = maxPostNo - 1
		}
	}

	return component.SuccessResponse(ReplyWindowPayload{
		Replies:       payloadReplies,
		AnchorReplyID: req.Params.AnchorPostID,
		BeforeCursor:  beforeCursor,
		AfterCursor:   afterCursor,
		BeforeReplyNo: beforeReplyNo,
		AfterReplyNo:  afterReplyNo,
		HasBefore:     hasBefore,
		HasAfter:      hasAfter,
		Total:         int64(topicEntity.ReplyCount),
		MaxReplyNo:    maxReplyNo,
	})
}

func filterReplyPosts(postEntities []*posts.Entity) []*posts.Entity {
	res := postEntities[:0]
	for _, item := range postEntities {
		if item == nil || item.PostNo <= 1 {
			continue
		}
		res = append(res, item)
	}
	return res
}

func canViewTopic(entity *topics.Entity, userID uint64) bool {
	if entity.Status != 1 {
		return userID != 0 && userID == entity.UserId
	}
	if entity.ProcessStatus != 0 && !currentUserCanViewProcessedArticle(userID) && !moderatorservice.CanModerateAnyCategory(userID, entity.CategoryIds) {
		return false
	}
	return true
}

func canViewTopicSimple(entity *topics.SmallEntity, userID uint64) bool {
	if entity.Status != 1 {
		return userID != 0 && userID == entity.UserId
	}
	if entity.ProcessStatus != 0 && !currentUserCanViewProcessedArticle(userID) && !moderatorservice.CanModerateAnyCategory(userID, entity.CategoryIds) {
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

func shouldCountTopicView(entity *topics.Entity) bool {
	return entity.Status == 1 && entity.ProcessStatus == 0
}

func renderNotFound(c *gin.Context) {
	renderNotFoundWithMessage(c, component.MessagePageNotFound)
}

func renderNotFoundWithMessage(c *gin.Context, messageCode component.MessageCode) {
	payload := PagePayload{
		Component: "error.notFound",
		Props: ErrorPageProps{
			Code:        "404",
			Title:       i18n.T(requestLang(c), "meta.notFound"),
			MessageCode: messageCode,
		},
		Meta: PageMeta{
			Title:  pageTitle(i18n.T(requestLang(c), "meta.notFound")),
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
