package eventhandlers

import (
	"context"
	"strconv"
	"strings"

	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/httpnotifyservice"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"
)

type ReportCreatedEvent struct {
	ReportId   uint64
	TargetType string
	TargetId   uint64
	ArticleId  uint64
	TopicId    uint64
	ReporterId uint64
	Reason     string
}

func handleHttpNotifyArticlePublished(ctx context.Context, event *ArticlePublishedEvent) error {
	if !httpnotifyservice.ShouldNotify(httpnotifyservice.EventArticlePublished) {
		return nil
	}
	httpnotifyservice.Notify(httpnotifyservice.EventArticlePublished, articleEventNotifyPayload(event))
	return nil
}

func handleHttpNotifyArticleUpdated(ctx context.Context, event *ArticleUpdatedEvent) error {
	if !httpnotifyservice.ShouldNotify(httpnotifyservice.EventArticleUpdated) {
		return nil
	}
	httpnotifyservice.Notify(httpnotifyservice.EventArticleUpdated, articleUpdatedEventNotifyPayload(event))
	return nil
}

func handleHttpNotifyCommentCreated(ctx context.Context, event *CommentCreatedEvent) error {
	if !httpnotifyservice.ShouldNotify(httpnotifyservice.EventCommentCreated) {
		return nil
	}
	topic := topics.GetSimple(event.topicID())
	articlePayload := topicNotifyPayloadFromSmall(topic)
	commenter := userNotifyPayload(event.UserId)
	post := posts.Get(event.postID())
	replyNo := uint64(0)
	if post.Id > 0 {
		replyNo = post.PostNo - 1
	}

	commentPayload := notifyComment{
		ID:                  event.postID(),
		ReplyNo:             replyNo,
		UserID:              event.UserId,
		User:                commenter,
		ParentReplyID:       event.ParentReplyId,
		ParentReplyAuthorID: event.ParentReplyAuthorId,
		URL:                 commentURL(event.topicID(), event.postID()),
	}
	payload := notifyEventData{
		BaseURI:        baseURI(),
		ContentPreview: TakeUpTo64Chars(event.Content),
		Article:        &articlePayload,
		User:           &commenter,
		Comment:        &commentPayload,
	}
	if event.ParentReplyAuthorId > 0 {
		parentAuthor := userNotifyPayload(event.ParentReplyAuthorId)
		payload.Comment.ParentReplyAuthor = &parentAuthor
	}
	httpnotifyservice.Notify(httpnotifyservice.EventCommentCreated, payload)
	return nil
}

func handleHttpNotifyUserSignUp(ctx context.Context, event *UserSignUpEvent) error {
	if !httpnotifyservice.ShouldNotify(httpnotifyservice.EventUserSignup) {
		return nil
	}
	user := userNotifyPayload(event.UserId)
	httpnotifyservice.Notify(httpnotifyservice.EventUserSignup, notifyEventData{
		BaseURI: baseURI(),
		User:    &user,
	})
	return nil
}

func handleHttpNotifyReportCreated(ctx context.Context, event *ReportCreatedEvent) error {
	if !httpnotifyservice.ShouldNotify(httpnotifyservice.EventReportCreated) {
		return nil
	}
	topic := topics.GetSimple(event.topicID())
	payload := notifyEventData{
		BaseURI:       baseURI(),
		ReportID:      new(event.ReportId),
		TargetType:    event.TargetType,
		TargetID:      new(event.TargetId),
		ReporterID:    new(event.ReporterId),
		Reason:        new(event.Reason),
		Article:       new(topicNotifyPayloadFromSmall(topic)),
		Reporter:      new(userNotifyPayload(event.ReporterId)),
		ModerationURL: moderationTargetURL(event),
	}
	if event.TargetType == "reply" || event.TargetType == "post" {
		post := posts.Get(event.TargetId)
		commenter := userNotifyPayload(post.UserId)
		payload.Comment = &notifyComment{
			ID:      post.Id,
			ReplyNo: post.PostNo - 1,
			UserID:  post.UserId,
			User:    commenter,
			URL:     commentURL(post.TopicId, post.Id),
		}
	}
	httpnotifyservice.Notify(httpnotifyservice.EventReportCreated, payload)
	return nil
}

type notifyEventData struct {
	BaseURI        string         `json:"baseUri"`
	ContentPreview string         `json:"contentPreview,omitempty"`
	Article        *notifyArticle `json:"article,omitempty"`
	User           *notifyUser    `json:"user,omitempty"`
	Comment        *notifyComment `json:"comment,omitempty"`
	ReportID       *uint64        `json:"reportId,omitempty"`
	TargetType     string         `json:"targetType,omitempty"`
	TargetID       *uint64        `json:"targetId,omitempty"`
	ReporterID     *uint64        `json:"reporterId,omitempty"`
	Reason         *string        `json:"reason,omitempty"`
	Reporter       *notifyUser    `json:"reporter,omitempty"`
	ModerationURL  string         `json:"moderationUrl,omitempty"`
}

type notifyArticle struct {
	ID            uint64           `json:"id"`
	Title         string           `json:"title"`
	URL           string           `json:"url"`
	Description   string           `json:"description"`
	FirstImageURL string           `json:"firstImageUrl"`
	UserID        uint64           `json:"userId"`
	User          notifyUser       `json:"user"`
	CategoryIDs   []uint64         `json:"categoryIds"`
	Categories    []notifyCategory `json:"categories"`
}

type notifyComment struct {
	ID                  uint64      `json:"id"`
	ReplyNo             uint64      `json:"replyNo"`
	UserID              uint64      `json:"userId"`
	User                notifyUser  `json:"user"`
	ParentReplyID       uint64      `json:"parentReplyId,omitempty"`
	ParentReplyAuthorID uint64      `json:"parentReplyAuthorId,omitempty"`
	ParentReplyAuthor   *notifyUser `json:"parentReplyAuthor,omitempty"`
	URL                 string      `json:"url"`
}

type notifyCategory struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type notifyUser struct {
	ID          uint64 `json:"id"`
	Username    string `json:"username"`
	Nickname    string `json:"nickname"`
	DisplayName string `json:"displayName"`
	AvatarURL   string `json:"avatarUrl"`
	URL         string `json:"url"`
}

func articleEventNotifyPayload(event *ArticlePublishedEvent) notifyEventData {
	if event != nil && event.Topic != nil {
		return topicNotifyPayload(event.Topic)
	}
	return notifyEventData{BaseURI: baseURI()}
}

func articleUpdatedEventNotifyPayload(event *ArticleUpdatedEvent) notifyEventData {
	if event != nil && event.Topic != nil {
		return topicNotifyPayload(event.Topic)
	}
	return notifyEventData{BaseURI: baseURI()}
}

func topicNotifyPayload(topic *topics.Entity) notifyEventData {
	if topic == nil {
		return notifyEventData{BaseURI: baseURI()}
	}
	summary := notifyArticle{
		ID:            topic.Id,
		Title:         topic.Title,
		URL:           urlconfig.PostDetail(topic.Id),
		Description:   topic.Excerpt,
		FirstImageURL: topic.FirstImageURL,
		UserID:        topic.UserId,
		User:          userNotifyPayload(topic.UserId),
		CategoryIDs:   topic.CategoryIds,
		Categories:    topicCategoryNotifyPayloads(topic.CategoryIds),
	}
	user := summary.User
	return notifyEventData{
		BaseURI: baseURI(),
		Article: &summary,
		User:    &user,
	}
}

func topicNotifyPayloadFromSmall(topic topics.SmallEntity) notifyArticle {
	if topic.Id == 0 {
		return notifyArticle{}
	}
	return notifyArticle{
		ID:            topic.Id,
		Title:         topic.Title,
		URL:           urlconfig.PostDetail(topic.Id),
		Description:   topic.Excerpt,
		FirstImageURL: topic.FirstImageURL,
		UserID:        topic.UserId,
		User:          userNotifyPayload(topic.UserId),
		CategoryIDs:   topic.CategoryIds,
		Categories:    topicCategoryNotifyPayloads(topic.CategoryIds),
	}
}

func topicCategoryNotifyPayloads(categoryIDs []uint64) []notifyCategory {
	if len(categoryIDs) == 0 {
		return []notifyCategory{}
	}
	categories := category.All()
	categoryByID := make(map[uint64]*category.Entity, len(categories))
	for _, item := range categories {
		categoryByID[item.Id] = item
	}
	payloads := make([]notifyCategory, 0, len(categoryIDs))
	for _, categoryID := range categoryIDs {
		item, ok := categoryByID[categoryID]
		if !ok {
			continue
		}
		payloads = append(payloads, notifyCategory{
			ID:   item.Id,
			Name: item.Name,
			Slug: item.Slug,
		})
	}
	return payloads
}

func userNotifyPayload(userID uint64) notifyUser {
	if userID == 0 {
		return notifyUser{}
	}
	user, err := users.Get(userID)
	if err != nil || user.Id == 0 {
		return notifyUser{ID: userID, URL: urlconfig.User(userID)}
	}
	displayName := user.Nickname
	if displayName == "" {
		displayName = user.Username
	}
	return notifyUser{
		ID:          user.Id,
		Username:    user.Username,
		Nickname:    user.Nickname,
		DisplayName: displayName,
		AvatarURL:   user.GetWebAvatarUrl(),
		URL:         urlconfig.User(user.Id),
	}
}

func commentURL(articleID uint64, commentID uint64) string {
	if articleID == 0 {
		return ""
	}
	if commentID == 0 {
		return urlconfig.PostDetail(articleID)
	}
	return urlconfig.PostDetail(articleID) + "#reply-" + uintToString(commentID)
}

func moderationTargetURL(event *ReportCreatedEvent) string {
	topicID := event.topicID()
	if topicID == 0 {
		return ""
	}
	if (event.TargetType == "reply" || event.TargetType == "post") && event.TargetId > 0 {
		return commentURL(topicID, event.TargetId)
	}
	return urlconfig.PostDetail(topicID)
}

func (event *ReportCreatedEvent) topicID() uint64 {
	if event.TopicId > 0 {
		return event.TopicId
	}
	return event.ArticleId
}

func uintToString(value uint64) string {
	return strconv.FormatUint(value, 10)
}

func baseURI() string {
	return strings.TrimRight(hotdataserve.GetSiteSettingsConfigCache().SiteUrl, "/")
}
