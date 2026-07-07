package eventnotice

import (
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/leancodebox/GooseForum/app/service/unreadservice"
	"github.com/spf13/cast"
)

// SendCommentNotification 发送评论通知
func SendCommentNotification(userId uint64, topicId uint64, commentContent string, commenterId uint64, postId uint64) error {
	payload := eventNotification.NotificationPayload{
		Content:     commentContent,
		TemplateKey: eventNotification.TemplateComment,
		TemplateParams: eventNotification.NotificationTemplateParams{
			Preview: commentContent,
		},
		ActorId: commenterId,
		TopicId: topicId,
		PostId:  postId,
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

// SendPostReplyNotification 发送 post 回复通知
func SendPostReplyNotification(userId uint64, postId uint64, topicId uint64, replyContent string, replierId uint64) error {
	payload := eventNotification.NotificationPayload{
		Content:     replyContent,
		TemplateKey: eventNotification.TemplatePostReply,
		TemplateParams: eventNotification.NotificationTemplateParams{
			Preview: replyContent,
		},
		ActorId: replierId,
		TopicId: topicId,
		PostId:  postId,
	}

	notification := &eventNotification.Entity{
		UserId:    userId,
		EventType: eventNotification.EventTypePostReply,
		Payload:   payload,
	}

	err := eventNotification.Create(notification)
	if err == nil {
		unreadservice.Invalidate(userId)
	}
	return err
}

func SendTopicPostNotifications(userIds []uint64, topicId uint64, postId uint64, commentContent string, commenterId uint64) error {
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
			EventType: eventNotification.EventTypeTopicPost,
			Payload: eventNotification.NotificationPayload{
				Content:     commentContent,
				TemplateKey: eventNotification.TemplateTopicPost,
				TemplateParams: eventNotification.NotificationTemplateParams{
					Preview: commentContent,
				},
				ActorId: commenterId,
				TopicId: topicId,
				PostId:  postId,
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
