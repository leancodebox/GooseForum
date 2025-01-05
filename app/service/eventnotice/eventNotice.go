package eventnotice

import (
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
)

// SendCommentNotification 发送评论通知
func SendCommentNotification(userId uint64, articleId uint64, commentContent string, commenterName string, commenterId uint64, articleTitle string) error {
	payload := eventNotification.NotificationPayload{
		Title:        "收到新评论",
		Content:      commentContent,
		ActorName:    commenterName,
		ActorId:      commenterId,
		ArticleId:    articleId,
		ArticleTitle: articleTitle,
	}

	notification := &eventNotification.Entity{
		UserId:    userId,
		EventType: eventNotification.EventTypeComment,
		Payload:   payload,
	}

	return eventNotification.Create(notification)
}

// SendReplyNotification 发送回复通知
func SendReplyNotification(userId uint64, commentId uint64, articleId uint64, replyContent string, replierName string, replierId uint64, articleTitle string) error {
	payload := eventNotification.NotificationPayload{
		Title:        "收到新回复",
		Content:      replyContent,
		ActorName:    replierName,
		ActorId:      replierId,
		ArticleId:    articleId,
		ArticleTitle: articleTitle,
		CommentId:    commentId,
	}

	notification := &eventNotification.Entity{
		UserId:    userId,
		EventType: eventNotification.EventTypeReply,
		Payload:   payload,
	}

	return eventNotification.Create(notification)
}

// SendSystemNotification 发送系统通知
func SendSystemNotification(userId uint64, title string, content string, extra map[string]interface{}) error {
	payload := eventNotification.NotificationPayload{
		Title:   title,
		Content: content,
		Extra:   extra,
	}

	notification := &eventNotification.Entity{
		UserId:    userId,
		EventType: eventNotification.EventTypeSystem,
		Payload:   payload,
	}

	return eventNotification.Create(notification)
}

// SendFollowNotification 发送关注通知
func SendFollowNotification(userId uint64, followerName string) error {
	payload := eventNotification.NotificationPayload{
		Title:   "新增关注者",
		Content: followerName + " 关注了你",
		Extra: map[string]interface{}{
			"followerName": followerName,
		},
	}

	notification := &eventNotification.Entity{
		UserId:    userId,
		EventType: eventNotification.EventTypeFollow,
		Payload:   payload,
	}

	return eventNotification.Create(notification)
}
