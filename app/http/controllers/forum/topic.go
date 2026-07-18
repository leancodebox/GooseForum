package forum

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/i18n"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/moderationservice"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/leancodebox/GooseForum/app/service/topicunseenservice"
	"github.com/leancodebox/GooseForum/app/service/topicviewservice"
	"github.com/leancodebox/GooseForum/app/service/userservice"
	"github.com/spf13/cast"
)

const postWindowLimit = 20

func TopicDetail(c *gin.Context) {
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
	if loginUser.UserId > 0 {
		if err := topicunseenservice.MarkVisited(loginUser.UserId, topic.Id, topic.LastPostId, time.Now()); err != nil {
			slog.Warn("mark topic visited failed", "userId", loginUser.UserId, "topicId", topic.Id, "error", err)
		}
	}
	props := buildTopicDetailProps(c, &topic, &firstPost)
	payload := PagePayload{
		Component: "topic.detail",
		Props:     props,
		Meta:      buildTopicMeta(c, props.Topic, props.PostStream.Posts),
		Layout:    buildLayout(c, activeKeyForTopic(props.Topic)),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	if topic.Status != 1 {
		payload.Meta.Robots = "noindex"
	}

	renderPage(c, "topic.gohtml", payload)
	if shouldCountTopicView(&topic) {
		topicviewservice.RecordView(topic.Id)
	}
}

type PostWindowReq struct {
	TopicID      uint64 `form:"topicId"`
	AnchorPostID uint64 `form:"anchorPostId"`
	AnchorPostNo uint64 `form:"anchorPostNo"`
	BeforePostNo uint64 `form:"beforePostNo"`
	AfterPostNo  uint64 `form:"afterPostNo"`
	Limit        int    `form:"limit"`
}

func PostWindow(req component.BetterRequest[PostWindowReq]) component.Response {
	topicID := req.Params.TopicID
	if topicID == 0 {
		return component.FailResponseCode(component.MessageTopicNotFound, nil)
	}

	topicEntity := topics.GetSimple(topicID)
	if topicEntity.Id == 0 {
		return component.FailResponseCode(component.MessageTopicNotFound, nil)
	}
	if !canViewTopicSimple(&topicEntity, req.UserId) {
		return component.FailResponseCode(component.MessageTopicNotFound, nil)
	}

	limit := req.Params.Limit
	if limit <= 0 || limit > 50 {
		limit = postWindowLimit
	}

	var postEntities []*posts.Entity
	hasBefore := false
	var hasAfter bool

	switch {
	case req.Params.AnchorPostNo > 0:
		anchor, ok := posts.GetByTopicPostNoAtOrAfter(topicID, req.Params.AnchorPostNo)
		if !ok {
			anchor, ok = posts.GetByTopicPostNoAtOrBefore(topicID, req.Params.AnchorPostNo)
		}
		if !ok || anchor.Id == 0 || anchor.TopicId != topicID || anchor.PostNo < 1 {
			return component.FailResponseCode(component.MessagePostNotFound, nil)
		}
		beforeLimit := min(5, limit/2)
		afterLimit := limit - beforeLimit - 1
		beforePosts := posts.GetByTopicPostNoBefore(topicID, anchor.PostNo, beforeLimit+1)
		afterPosts := posts.GetByTopicPostNoAfter(topicID, anchor.PostNo, afterLimit+1)
		hasBefore = len(beforePosts) > beforeLimit
		hasAfter = len(afterPosts) > afterLimit
		if hasBefore {
			beforePosts = beforePosts[1:]
		}
		if hasAfter {
			afterPosts = afterPosts[:afterLimit]
		}
		postEntities = append(postEntities, beforePosts...)
		postEntities = append(postEntities, &anchor)
		postEntities = append(postEntities, afterPosts...)
	case req.Params.AnchorPostID > 0:
		anchor := posts.Get(req.Params.AnchorPostID)
		if anchor.Id == 0 || anchor.TopicId != topicID || anchor.PostNo < 1 {
			return component.FailResponseCode(component.MessagePostNotFound, nil)
		}
		beforeLimit := min(5, limit/2)
		afterLimit := limit - beforeLimit - 1
		beforePosts := posts.GetByTopicIdBefore(topicID, anchor.Id, beforeLimit+1)
		afterPosts := posts.GetByTopicIdAfter(topicID, anchor.Id, afterLimit+1)
		hasBefore = len(beforePosts) > beforeLimit
		hasAfter = len(afterPosts) > afterLimit
		if hasBefore {
			beforePosts = beforePosts[1:]
		}
		if hasAfter {
			afterPosts = afterPosts[:afterLimit]
		}
		postEntities = append(postEntities, beforePosts...)
		postEntities = append(postEntities, &anchor)
		postEntities = append(postEntities, afterPosts...)
	case req.Params.BeforePostNo > 0:
		postEntities = posts.GetByTopicPostNoBefore(topicID, req.Params.BeforePostNo, limit+1)
		hasBefore = len(postEntities) > limit
		if hasBefore {
			postEntities = postEntities[1:]
		}
		hasAfter = true
	case req.Params.AfterPostNo > 0:
		postEntities = posts.GetByTopicPostNoAfter(topicID, req.Params.AfterPostNo, limit+1)
		hasAfter = len(postEntities) > limit
		if hasAfter {
			postEntities = postEntities[:limit]
		}
		hasBefore = true
	default:
		postEntities = posts.GetByTopicPostNoAfter(topicID, 0, limit+1)
		hasAfter = len(postEntities) > limit
		if hasAfter {
			postEntities = postEntities[:limit]
		}
	}

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
	canModeratePosts := moderationservice.CanModerateAnyCategory(req.UserId, topicEntity.CategoryIds)
	maxPostNo := uint64(0)
	if topicEntity.PostSeq > 0 {
		maxPostNo = topicEntity.PostSeq
	}
	if maxPostNo == 0 && topicEntity.ReplyCount > 0 {
		maxPostSeq := posts.GetMaxPostNoByTopicId(topicID)
		if maxPostSeq > 0 {
			maxPostNo = maxPostSeq
		}
	}

	return component.SuccessResponse(buildPostWindowPayloadFromEntities(
		postEntities,
		userMap,
		req.UserId,
		canModeratePosts,
		hasBefore,
		hasAfter,
		int64(maxPostNo),
		maxPostNo,
		req.Params.AnchorPostID,
	))
}

func canViewTopic(entity *topics.Entity, userID uint64) bool {
	if entity.Status != 1 {
		return userID != 0 && userID == entity.UserId
	}
	if entity.ProcessStatus != 0 && !currentUserCanViewProcessedTopic(userID) && !moderationservice.CanModerateAnyCategory(userID, entity.CategoryIds) {
		return false
	}
	return true
}

func canViewTopicSimple(entity *topics.Entity, userID uint64) bool {
	if entity.Status != 1 {
		return userID != 0 && userID == entity.UserId
	}
	if entity.ProcessStatus != 0 && !currentUserCanViewProcessedTopic(userID) && !moderationservice.CanModerateAnyCategory(userID, entity.CategoryIds) {
		return false
	}
	return true
}

func currentUserCanViewProcessedTopic(userID uint64) bool {
	if userID == 0 {
		return false
	}
	roleID, ok := userservice.GetUserRoleId(userID)
	return ok && permission.CheckRole(roleID, permission.TopicsManager)
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

func activeKeyForTopic(topic TopicDetailPayload) string {
	if len(topic.Categories) > 0 {
		return "category_" + cast.ToString(topic.Categories[0].ID)
	}
	return "topics"
}
