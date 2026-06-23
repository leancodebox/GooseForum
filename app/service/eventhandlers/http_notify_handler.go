package eventhandlers

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/leancodebox/GooseForum/app/service/httpnotifyservice"
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
		httpnotifyservice.Notify(httpnotifyservice.EventArticlePublished, map[string]any{
			"articleId": event.Article.Id,
			"title":     event.Article.Title,
			"userId":    event.Article.UserId,
		})
		return nil
	})
}

func NewHttpNotifyArticleUpdatedHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler("HttpNotifyArticleUpdatedHandler", func(ctx context.Context, event *ArticleUpdatedEvent) error {
		httpnotifyservice.Notify(httpnotifyservice.EventArticleUpdated, map[string]any{
			"articleId": event.Article.Id,
			"title":     event.Article.Title,
			"userId":    event.Article.UserId,
		})
		return nil
	})
}

func NewHttpNotifyCommentCreatedHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler("HttpNotifyCommentCreatedHandler", func(ctx context.Context, event *CommentCreatedEvent) error {
		httpnotifyservice.Notify(httpnotifyservice.EventCommentCreated, map[string]any{
			"articleId": event.ArticleId,
			"commentId": event.CommentId,
			"userId":    event.UserId,
		})
		return nil
	})
}

func NewHttpNotifyUserSignUpHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler("HttpNotifyUserSignUpHandler", func(ctx context.Context, event *UserSignUpEvent) error {
		httpnotifyservice.Notify(httpnotifyservice.EventUserSignup, map[string]any{
			"userId":   event.UserId,
			"username": event.Username,
		})
		return nil
	})
}

func NewHttpNotifyReportCreatedHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler("HttpNotifyReportCreatedHandler", func(ctx context.Context, event *ReportCreatedEvent) error {
		httpnotifyservice.Notify(httpnotifyservice.EventReportCreated, map[string]any{
			"reportId":   event.ReportId,
			"targetType": event.TargetType,
			"targetId":   event.TargetId,
			"articleId":  event.ArticleId,
			"reporterId": event.ReporterId,
			"reason":     event.Reason,
		})
		return nil
	})
}
