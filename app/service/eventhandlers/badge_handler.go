package eventhandlers

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/leancodebox/GooseForum/app/service/badgeservice"
	"github.com/leancodebox/GooseForum/app/service/userservice"
)

func NewBadgePostHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		"BadgePostHandler",
		func(ctx context.Context, event *ArticlePublishedEvent) error {
			checkAndInvalidateUserBadges(event.Article.UserId, badgeservice.TriggerPost)
			return nil
		},
	)
}

func checkAndInvalidateUserBadges(userID uint64, trigger badgeservice.Trigger) {
	before := len(badgeservice.GetUserBadges(userID))
	badgeservice.CheckAndGrant(userID, trigger)
	if len(badgeservice.GetUserBadges(userID)) != before {
		userservice.InvalidateUserPublicProfileCache(userID)
	}
}

func NewBadgeCommentHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		"BadgeCommentHandler",
		func(ctx context.Context, event *CommentCreatedEvent) error {
			checkAndInvalidateUserBadges(event.UserId, badgeservice.TriggerComment)
			return nil
		},
	)
}

func NewBadgeLikeHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		"BadgeLikeHandler",
		func(ctx context.Context, event *ArticleLikedEvent) error {
			checkAndInvalidateUserBadges(event.LikierId, badgeservice.TriggerLike)
			checkAndInvalidateUserBadges(event.UserId, badgeservice.TriggerLike)
			return nil
		},
	)
}

func NewBadgeFollowHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		"BadgeFollowHandler",
		func(ctx context.Context, event *UserFollowedEvent) error {
			checkAndInvalidateUserBadges(event.UserId, badgeservice.TriggerFollow)
			return nil
		},
	)
}
