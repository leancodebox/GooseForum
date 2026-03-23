package eventhandlers

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/leancodebox/GooseForum/app/models/forum/userActivities"
)

// NewActivitySignUpHandler 记录注册行为
func NewActivitySignUpHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		"ActivitySignUpHandler",
		func(ctx context.Context, event *UserSignUpEvent) error {
			return userActivities.Record(event.UserId, userActivities.ActionSignUp, userActivities.SubjectUser, event.UserId, "注册了账号")
		},
	)
}

// NewActivityPostHandler 记录发帖行为
func NewActivityPostHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		"ActivityPostHandler",
		func(ctx context.Context, event *ArticlePublishedEvent) error {
			return userActivities.Record(event.Article.UserId, userActivities.ActionPost, userActivities.SubjectTopic, event.Article.Id, event.Article.Title)
		},
	)
}

// NewActivityLikeHandler 记录点赞行为
func NewActivityLikeHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		"ActivityLikeHandler",
		func(ctx context.Context, event *ArticleLikedEvent) error {
			return userActivities.Record(event.LikierId, userActivities.ActionLike, userActivities.SubjectTopic, event.ArticleId, event.Title)
		},
	)
}

// NewActivityFollowHandler 记录关注行为
func NewActivityFollowHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		"ActivityFollowHandler",
		func(ctx context.Context, event *UserFollowedEvent) error {
			return userActivities.Record(event.FollowerId, userActivities.ActionFollow, userActivities.SubjectUser, event.UserId, "Followed a user")
		},
	)
}

// NewActivityReplyHandler 记录回复行为
func NewActivityReplyHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		"ActivityReplyHandler",
		func(ctx context.Context, event *CommentCreatedEvent) error {
			return userActivities.Record(event.UserId, userActivities.ActionComment, userActivities.SubjectPost, event.CommentId, TakeUpTo64Chars(event.Content))
		},
	)
}
