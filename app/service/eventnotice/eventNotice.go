package eventnotice

import (
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/leancodebox/GooseForum/app/service/unreadservice"
	"github.com/spf13/cast"
)

// SendCommentNotification 发送评论通知
func SendCommentNotification(userId uint64, articleId uint64, commentContent string, commenterId uint64, replyId uint64) error {
	payload := eventNotification.NotificationPayload{
		Content:     commentContent,
		TemplateKey: eventNotification.TemplateComment,
		TemplateParams: eventNotification.NotificationTemplateParams{
			Preview: commentContent,
		},
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
		Content:     replyContent,
		TemplateKey: eventNotification.TemplateReply,
		TemplateParams: eventNotification.NotificationTemplateParams{
			Preview: replyContent,
		},
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

func SendArticleCommentNotifications(userIds []uint64, articleId uint64, commentId uint64, commentContent string, commenterId uint64) error {
	if len(userIds) == 0 {
		return nil
	}

	notifications := make([]*eventNotification.Entity, 0, len(userIds))
	for _, userId := range userIds {
		if userId == 0 {
			continue
		}
		notifications = append(notifications, &eventNotification.Entity{
			UserId:    userId,
			EventType: eventNotification.EventTypeArticleComment,
			Payload: eventNotification.NotificationPayload{
				Content:     commentContent,
				TemplateKey: eventNotification.TemplateArticleComment,
				TemplateParams: eventNotification.NotificationTemplateParams{
					Preview: commentContent,
				},
				ActorId:   commenterId,
				ArticleId: articleId,
				CommentId: commentId,
			},
		})
	}
	if len(notifications) == 0 {
		return nil
	}

	err := eventNotification.CreateBatch(notifications, 100)
	if err == nil {
		for _, userId := range userIds {
			unreadservice.Invalidate(userId)
		}
	}
	return err
}

func SendBadgeNotification(userId uint64, badgeCode string, badgeName string, badgeIconURL string) error {
	payload := eventNotification.NotificationPayload{
		TemplateKey: eventNotification.TemplateBadge,
		TemplateParams: eventNotification.NotificationTemplateParams{
			BadgeCode: badgeCode,
			BadgeName: badgeName,
		},
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
		TemplateKey: eventNotification.TemplateFollow,
		TemplateParams: eventNotification.NotificationTemplateParams{
			FollowerName: followerName,
		},
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
