package eventhandlers

import (
	"context"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/leancodebox/GooseForum/app/models/forum/dailyStats"
)

// NewStatsSignUpHandler 记录注册统计
func NewStatsSignUpHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		"StatsSignUpHandler",
		func(ctx context.Context, event *UserSignUpEvent) error {
			return dailyStats.Increment(time.Now(), dailyStats.StatTypeRegCount, 1)
		},
	)
}

// NewStatsPostHandler 记录发帖统计
func NewStatsPostHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		"StatsPostHandler",
		func(ctx context.Context, event *ArticlePublishedEvent) error {
			return dailyStats.Increment(time.Now(), dailyStats.StatTypeArticleCount, 1)
		},
	)
}

// NewStatsReplyHandler 记录回复统计
func NewStatsReplyHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		"StatsReplyHandler",
		func(ctx context.Context, event *CommentCreatedEvent) error {
			return dailyStats.Increment(time.Now(), dailyStats.StatTypeReplyCount, 1)
		},
	)
}
