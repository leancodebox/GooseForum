package eventhandlers

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/leancodebox/GooseForum/app/service/pointservice"
)

// NewPointArticlePublishedHandler 发帖获得积分
func NewPointArticlePublishedHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		"PointArticlePublishedHandler",
		func(ctx context.Context, event *ArticlePublishedEvent) error {
			pointservice.RewardPoints(event.Article.UserId, 10, pointservice.RewardPoints4WriteArticles)
			return nil
		},
	)
}

// NewPointCommentCreatedHandler 评论获得积分
func NewPointCommentCreatedHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		"PointCommentCreatedHandler",
		func(ctx context.Context, event *CommentCreatedEvent) error {
			pointservice.RewardPoints(event.UserId, 2, pointservice.RewardPoints4Reply)
			return nil
		},
	)
}
