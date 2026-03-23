package eventhandlers

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

// Handlers 返回所有事件处理器
func Handlers() []cqrs.EventHandler {
	return []cqrs.EventHandler{
		NewCommentCreatedHandler(),
		NewUserFollowedHandler(),
		NewArticlePublishedHandler(),
		NewArticleUpdatedHandler(),
		NewPointArticlePublishedHandler(),
		NewPointCommentCreatedHandler(),
		NewUserLastActiveUpdatedHandler(),

		// 用户行为记录处理器
		NewActivitySignUpHandler(),
		NewActivityPostHandler(),
		NewActivityLikeHandler(),
		NewActivityFollowHandler(),
		NewActivityReplyHandler(),

		// 每日统计处理器
		NewStatsSignUpHandler(),
		NewStatsPostHandler(),
		NewStatsReplyHandler(),
	}
}
