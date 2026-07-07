package eventhandlers

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

// Handlers 返回所有事件处理器
func Handlers() []cqrs.EventHandler {
	return []cqrs.EventHandler{
		cqrs.NewEventHandler("CommentCreatedHandler", handleCommentCreated),
		cqrs.NewEventHandler("UserFollowedHandler", handleUserFollowed),
		cqrs.NewEventHandler("TopicPublishedHandler", handleTopicPublished),
		cqrs.NewEventHandler("TopicUpdatedHandler", handleTopicUpdated),
		cqrs.NewEventHandler("PointTopicPublishedHandler", handlePointTopicPublished),
		cqrs.NewEventHandler("PointCommentCreatedHandler", handlePointCommentCreated),
		cqrs.NewEventHandler("UserLastActiveUpdatedHandler", handleUserLastActiveUpdated),

		// 用户行为记录处理器
		cqrs.NewEventHandler("ActivitySignUpHandler", handleActivitySignUp),
		cqrs.NewEventHandler("ActivityPostHandler", handleActivityPost),
		cqrs.NewEventHandler("ActivityLikeHandler", handleActivityLike),
		cqrs.NewEventHandler("ActivityFollowHandler", handleActivityFollow),
		cqrs.NewEventHandler("ActivityReplyHandler", handleActivityReply),

		// 每日统计处理器
		cqrs.NewEventHandler("StatsSignUpHandler", handleStatsSignUp),
		cqrs.NewEventHandler("StatsPostHandler", handleStatsPost),
		cqrs.NewEventHandler("StatsReplyHandler", handleStatsReply),

		// 徽章自动授予处理器
		cqrs.NewEventHandler("BadgePostHandler", handleBadgePost),
		cqrs.NewEventHandler("BadgeCommentHandler", handleBadgeComment),
		cqrs.NewEventHandler("BadgeLikeHandler", handleBadgeLike),
		cqrs.NewEventHandler("BadgeFollowHandler", handleBadgeFollow),

		// HTTP 事件通知
		cqrs.NewEventHandler("HttpNotifyTopicPublishedHandler", handleHttpNotifyTopicPublished),
		cqrs.NewEventHandler("HttpNotifyTopicUpdatedHandler", handleHttpNotifyTopicUpdated),
		cqrs.NewEventHandler("HttpNotifyCommentCreatedHandler", handleHttpNotifyCommentCreated),
		cqrs.NewEventHandler("HttpNotifyUserSignUpHandler", handleHttpNotifyUserSignUp),
		cqrs.NewEventHandler("HttpNotifyReportCreatedHandler", handleHttpNotifyReportCreated),
	}
}
