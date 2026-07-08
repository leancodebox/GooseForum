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
	TopicId    uint64
	ReporterId uint64
	Reason     string
}

func handleHttpNotifyTopicPublished(ctx context.Context, event *TopicPublishedEvent) error {
	if !httpnotifyservice.ShouldNotify(httpnotifyservice.EventTopicPublished) {
		return nil
	}
	httpnotifyservice.Notify(httpnotifyservice.EventTopicPublished, topicEventNotifyPayload(event))
	return nil
}

func handleHttpNotifyTopicUpdated(ctx context.Context, event *TopicUpdatedEvent) error {
	if !httpnotifyservice.ShouldNotify(httpnotifyservice.EventTopicUpdated) {
		return nil
	}
	httpnotifyservice.Notify(httpnotifyservice.EventTopicUpdated, topicUpdatedEventNotifyPayload(event))
	return nil
}

func handleHttpNotifyCommentCreated(ctx context.Context, event *CommentCreatedEvent) error {
	if !httpnotifyservice.ShouldNotify(httpnotifyservice.EventCommentCreated) {
		return nil
	}
	topic := topics.GetSimple(event.TopicId)
	topicPayload := topicNotifyPayloadFromSmall(topic)
	commenter := userNotifyPayload(event.UserId)
	post := posts.Get(event.PostId)
	postNo := uint64(0)
	if post.Id > 0 {
		postNo = post.PostNo
	}

	postPayload := notifyPost{
		ID:                  event.PostId,
		PostNo:              postNo,
		UserID:              event.UserId,
		User:                commenter,
		ReplyToPostID:       event.ReplyToPostId,
		ReplyToPostAuthorID: event.ReplyToPostAuthorId,
		URL:                 postURL(event.TopicId, event.PostId),
	}
	payload := notifyEventData{
		BaseURI:        baseURI(),
		ContentPreview: TakeUpTo64Chars(event.Content),
		Topic:          &topicPayload,
		User:           &commenter,
		Post:           &postPayload,
	}
	if event.ReplyToPostAuthorId > 0 {
		parentAuthor := userNotifyPayload(event.ReplyToPostAuthorId)
		payload.Post.ReplyToPostAuthor = &parentAuthor
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
	topic := topics.GetSimple(event.TopicId)
	payload := notifyEventData{
		BaseURI:       baseURI(),
		ReportID:      new(event.ReportId),
		TargetType:    event.TargetType,
		TargetID:      new(event.TargetId),
		ReporterID:    new(event.ReporterId),
		Reason:        new(event.Reason),
		Topic:         new(topicNotifyPayloadFromSmall(topic)),
		Reporter:      new(userNotifyPayload(event.ReporterId)),
		ModerationURL: moderationTargetURL(event),
	}
	if event.TargetType == "reply" || event.TargetType == "post" {
		post := posts.Get(event.TargetId)
		commenter := userNotifyPayload(post.UserId)
		payload.Post = &notifyPost{
			ID:     post.Id,
			PostNo: post.PostNo,
			UserID: post.UserId,
			User:   commenter,
			URL:    postURL(post.TopicId, post.Id),
		}
	}
	httpnotifyservice.Notify(httpnotifyservice.EventReportCreated, payload)
	return nil
}

type notifyEventData struct {
	BaseURI        string       `json:"baseUri"`
	ContentPreview string       `json:"contentPreview,omitempty"`
	Topic          *notifyTopic `json:"topic,omitempty"`
	User           *notifyUser  `json:"user,omitempty"`
	Post           *notifyPost  `json:"post,omitempty"`
	ReportID       *uint64      `json:"reportId,omitempty"`
	TargetType     string       `json:"targetType,omitempty"`
	TargetID       *uint64      `json:"targetId,omitempty"`
	ReporterID     *uint64      `json:"reporterId,omitempty"`
	Reason         *string      `json:"reason,omitempty"`
	Reporter       *notifyUser  `json:"reporter,omitempty"`
	ModerationURL  string       `json:"moderationUrl,omitempty"`
}

type notifyTopic struct {
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

type notifyPost struct {
	ID                  uint64      `json:"id"`
	PostNo              uint64      `json:"postNo"`
	UserID              uint64      `json:"userId"`
	User                notifyUser  `json:"user"`
	ReplyToPostID       uint64      `json:"replyToPostId,omitempty"`
	ReplyToPostAuthorID uint64      `json:"replyToPostAuthorId,omitempty"`
	ReplyToPostAuthor   *notifyUser `json:"replyToPostAuthor,omitempty"`
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

func topicEventNotifyPayload(event *TopicPublishedEvent) notifyEventData {
	if event != nil && event.Topic != nil {
		return topicNotifyPayload(event.Topic)
	}
	return notifyEventData{BaseURI: baseURI()}
}

func topicUpdatedEventNotifyPayload(event *TopicUpdatedEvent) notifyEventData {
	if event != nil && event.Topic != nil {
		return topicNotifyPayload(event.Topic)
	}
	return notifyEventData{BaseURI: baseURI()}
}

func topicNotifyPayload(topic *topics.Entity) notifyEventData {
	if topic == nil {
		return notifyEventData{BaseURI: baseURI()}
	}
	summary := notifyTopic{
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
		Topic:   &summary,
		User:    &user,
	}
}

func topicNotifyPayloadFromSmall(topic topics.Entity) notifyTopic {
	if topic.Id == 0 {
		return notifyTopic{}
	}
	return notifyTopic{
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

func postURL(topicID uint64, postID uint64) string {
	if topicID == 0 {
		return ""
	}
	if postID == 0 {
		return urlconfig.PostDetail(topicID)
	}
	return urlconfig.PostDetail(topicID) + "#post-" + uintToString(postID)
}

func moderationTargetURL(event *ReportCreatedEvent) string {
	if event.TopicId == 0 {
		return ""
	}
	if (event.TargetType == "reply" || event.TargetType == "post") && event.TargetId > 0 {
		return postURL(event.TopicId, event.TargetId)
	}
	return urlconfig.PostDetail(event.TopicId)
}

func uintToString(value uint64) string {
	return strconv.FormatUint(value, 10)
}

func baseURI() string {
	return strings.TrimRight(hotdataserve.GetSiteSettingsConfigCache().SiteUrl, "/")
}
