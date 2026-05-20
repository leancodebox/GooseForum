package eventnotice

import (
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/leancodebox/GooseForum/app/service/unreadservice"
	"github.com/spf13/cast"
)

// SendCommentNotification 发送评论通知
func SendCommentNotification(userId uint64, articleId uint64, commentContent string, commenterId uint64, replyId uint64) error {
	payload := eventNotification.NotificationPayload{
		Title:     "收到新评论",
		Content:   commentContent,
		ActorId:   commenterId,
		ArticleId: articleId,
		CommentId: replyId,
	}

	notification := &eventNotification.Entity{
		UserId:    userId,
		EventType: eventNotification.EventTypeComment,
		Payload:   payload,
	}

	err := eventNotification.Create(notification)
	if err == nil {
		unreadservice.Invalidate(userId)
	}
	return err
}

// SendReplyNotification 发送回复通知
func SendReplyNotification(userId uint64, commentId uint64, articleId uint64, replyContent string, replierId uint64) error {
	payload := eventNotification.NotificationPayload{
		Title:     "收到新回复",
		Content:   replyContent,
		ActorId:   replierId,
		ArticleId: articleId,
		CommentId: commentId,
	}

	notification := &eventNotification.Entity{
		UserId:    userId,
		EventType: eventNotification.EventTypeReply,
		Payload:   payload,
	}

	err := eventNotification.Create(notification)
	if err == nil {
		unreadservice.Invalidate(userId)
	}
	return err
}

// SendSystemNotification 发送系统通知
func SendSystemNotification(userId uint64, title string, content string, extra map[string]any) error {
	payload := eventNotification.NotificationPayload{
		Title:   title,
		Content: content,
		Extra:   eventNotification.Extra{},
	}

	notification := &eventNotification.Entity{
		UserId:    userId,
		EventType: eventNotification.EventTypeSystem,
		Payload:   payload,
	}

	err := eventNotification.Create(notification)
	if err == nil {
		unreadservice.Invalidate(userId)
	}
	return err
}

func SendBadgeNotification(userId uint64, badgeCode string, badgeName string, badgeIconURL string) error {
	payload := eventNotification.NotificationPayload{
		Title:   "获得新徽章",
		Content: "你获得了「" + badgeName + "」徽章",
		ActorId: userId,
		Extra: eventNotification.Extra{
			BadgeCode:    badgeCode,
			BadgeName:    badgeName,
			BadgeIconURL: badgeIconURL,
			ProfileURL:   "/u/" + cast.ToString(userId),
		},
	}

	notification := &eventNotification.Entity{
		UserId:    userId,
		EventType: eventNotification.EventTypeBadge,
		Payload:   payload,
	}

	err := eventNotification.Create(notification)
	if err == nil {
		unreadservice.Invalidate(userId)
	}
	return err
}

// SendFollowNotification 发送关注通知
func SendFollowNotification(userId uint64, followerId uint64, followerName string) error {
	payload := eventNotification.NotificationPayload{
		Title:   "新增关注者",
		Content: followerName + " 关注了你",
		ActorId: followerId,
		Extra:   eventNotification.Extra{FollowerName: followerName},
	}

	notification := &eventNotification.Entity{
		UserId:    userId,
		EventType: eventNotification.EventTypeFollow,
		Payload:   payload,
	}

	err := eventNotification.Create(notification)
	if err == nil {
		unreadservice.Invalidate(userId)
	}
	return err
}
