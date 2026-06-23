package eventhandlers

import (
	"context"
	"strconv"
	"strings"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
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
	ReporterId uint64
	Reason     string
}

func NewHttpNotifyArticlePublishedHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler("HttpNotifyArticlePublishedHandler", func(ctx context.Context, event *ArticlePublishedEvent) error {
		httpnotifyservice.Notify(httpnotifyservice.EventArticlePublished, articleNotifyPayload(event.Article))
		return nil
	})
}

func NewHttpNotifyArticleUpdatedHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler("HttpNotifyArticleUpdatedHandler", func(ctx context.Context, event *ArticleUpdatedEvent) error {
		httpnotifyservice.Notify(httpnotifyservice.EventArticleUpdated, articleNotifyPayload(event.Article))
		return nil
	})
}

func NewHttpNotifyCommentCreatedHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler("HttpNotifyCommentCreatedHandler", func(ctx context.Context, event *CommentCreatedEvent) error {
		article := articles.GetSimple(event.ArticleId)
		articlePayload := articleNotifyPayloadFromSmall(article)
		commenter := userNotifyPayload(event.UserId)
		comment := reply.Get(event.CommentId)

		commentPayload := notifyComment{
			ID:                  event.CommentId,
			ReplyNo:             comment.ReplyNo,
			UserID:              event.UserId,
			User:                commenter,
			ParentReplyID:       event.ParentReplyId,
			ParentReplyAuthorID: event.ParentReplyAuthorId,
			URL:                 commentURL(event.ArticleId, event.CommentId),
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
	})
}

func NewHttpNotifyUserSignUpHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler("HttpNotifyUserSignUpHandler", func(ctx context.Context, event *UserSignUpEvent) error {
		user := userNotifyPayload(event.UserId)
		httpnotifyservice.Notify(httpnotifyservice.EventUserSignup, notifyEventData{
			BaseURI: baseURI(),
			User:    &user,
		})
		return nil
	})
}

func NewHttpNotifyReportCreatedHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler("HttpNotifyReportCreatedHandler", func(ctx context.Context, event *ReportCreatedEvent) error {
		article := articles.GetSimple(event.ArticleId)
		articlePayload := articleNotifyPayloadFromSmall(article)
		reporter := userNotifyPayload(event.ReporterId)
		payload := notifyEventData{
			BaseURI:       baseURI(),
			ReportID:      new(event.ReportId),
			TargetType:    event.TargetType,
			TargetID:      new(event.TargetId),
			ReporterID:    new(event.ReporterId),
			Reason:        new(event.Reason),
			Article:       &articlePayload,
			Reporter:      &reporter,
			ModerationURL: moderationTargetURL(event),
		}
		if event.TargetType == "reply" {
			comment := reply.Get(event.TargetId)
			commenter := userNotifyPayload(comment.UserId)
			payload.Comment = &notifyComment{
				ID:      comment.Id,
				ReplyNo: comment.ReplyNo,
				UserID:  comment.UserId,
				User:    commenter,
				URL:     commentURL(comment.ArticleId, comment.Id),
			}
		}
		httpnotifyservice.Notify(httpnotifyservice.EventReportCreated, payload)
		return nil
	})
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

func articleNotifyPayload(article *articles.Entity) notifyEventData {
	if article == nil {
		return notifyEventData{BaseURI: baseURI()}
	}
	summary := notifyArticle{
		ID:            article.Id,
		Title:         article.Title,
		URL:           urlconfig.PostDetail(article.Id),
		Description:   article.Description,
		FirstImageURL: article.FirstImageURL,
		UserID:        article.UserId,
		User:          userNotifyPayload(article.UserId),
		CategoryIDs:   article.CategoryId,
		Categories:    categoryNotifyPayloads(article.CategoryId),
	}
	user := summary.User
	return notifyEventData{
		BaseURI: baseURI(),
		Article: &summary,
		User:    &user,
	}
}

func articleNotifyPayloadFromSmall(article articles.SmallEntity) notifyArticle {
	if article.Id == 0 {
		return notifyArticle{}
	}
	return notifyArticle{
		ID:            article.Id,
		Title:         article.Title,
		URL:           urlconfig.PostDetail(article.Id),
		Description:   article.Description,
		FirstImageURL: article.FirstImageURL,
		UserID:        article.UserId,
		User:          userNotifyPayload(article.UserId),
		CategoryIDs:   article.CategoryId,
		Categories:    categoryNotifyPayloads(article.CategoryId),
	}
}

func categoryNotifyPayloads(categoryIDs []uint64) []notifyCategory {
	if len(categoryIDs) == 0 {
		return []notifyCategory{}
	}
	categories := articleCategory.All()
	categoryByID := make(map[uint64]*articleCategory.Entity, len(categories))
	for _, category := range categories {
		categoryByID[category.Id] = category
	}
	payloads := make([]notifyCategory, 0, len(categoryIDs))
	for _, categoryID := range categoryIDs {
		category, ok := categoryByID[categoryID]
		if !ok {
			continue
		}
		payloads = append(payloads, notifyCategory{
			ID:   category.Id,
			Name: category.Category,
			Slug: category.Slug,
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
	if event.ArticleId == 0 {
		return ""
	}
	if event.TargetType == "reply" && event.TargetId > 0 {
		return commentURL(event.ArticleId, event.TargetId)
	}
	return urlconfig.PostDetail(event.ArticleId)
}

func uintToString(value uint64) string {
	return strconv.FormatUint(value, 10)
}


func baseURI() string {
	return strings.TrimRight(hotdataserve.GetSiteSettingsConfigCache().SiteUrl, "/")
}
