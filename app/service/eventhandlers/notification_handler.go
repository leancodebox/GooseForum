package eventhandlers

import (
	"context"

	"github.com/leancodebox/GooseForum/app/models/forum/topicUserAction"
	"github.com/leancodebox/GooseForum/app/service/eventnotice"
)

const articleWatchNotifyBatchSize = 500

// TakeUpTo64Chars 按字符数截取字符串，最多取 64 个字符
func TakeUpTo64Chars(s string) string {
	runes := []rune(s)
	if len(runes) > 64 {
		return string(runes[:64])
	}
	return s
}

// CommentCreatedEvent 评论/回复创建事件
type CommentCreatedEvent struct {
	TopicId             uint64
	PostId              uint64 // 新创建的 post ID
	UserId              uint64 // 发表评论者 ID
	Content             string
	TopicAuthorId       uint64 // 主题作者 ID
	ReplyToPostId       uint64 // 被回复的 post ID
	ReplyToPostAuthorId uint64 // 被回复的 post 作者 ID
}

// handleCommentCreated 发送评论/回复通知
func handleCommentCreated(ctx context.Context, event *CommentCreatedEvent) error {
	contentPreview := TakeUpTo64Chars(event.Content)
	// 如果不是主题作者自己发表评论，通知主题作者
	if shouldNotifyArticleAuthor(event) {
		_ = eventnotice.SendCommentNotification(event.TopicAuthorId, event.TopicId, contentPreview, event.UserId, event.PostId)
	}
	// 如果是回复 post，且不是回复自己，通知原 post 作者
	if shouldNotifyParentReplyAuthor(event) {
		_ = eventnotice.SendPostReplyNotification(event.ReplyToPostAuthorId, event.PostId, event.TopicId, contentPreview, event.UserId)
	}
	notifyArticleWatchers(event, contentPreview)
	return nil
}

func shouldNotifyArticleAuthor(event *CommentCreatedEvent) bool {
	if event.TopicAuthorId == 0 || event.TopicAuthorId == event.UserId {
		return false
	}
	return event.ReplyToPostId == 0 || event.TopicAuthorId != event.ReplyToPostAuthorId
}

func shouldNotifyParentReplyAuthor(event *CommentCreatedEvent) bool {
	return event.ReplyToPostId > 0 && event.ReplyToPostAuthorId > 0 && event.ReplyToPostAuthorId != event.UserId
}

func notifyArticleWatchers(event *CommentCreatedEvent, contentPreview string) {
	excludeUserIds := commentNotificationExcludeUserIds(event)
	afterUserId := uint64(0)
	for {
		userIds := topicUserAction.ListActiveWatchUserIDsAfter(event.TopicId, afterUserId, excludeUserIds, articleWatchNotifyBatchSize)
		if len(userIds) == 0 {
			return
		}
		_ = eventnotice.SendTopicPostNotifications(userIds, event.TopicId, event.PostId, contentPreview, event.UserId)
		afterUserId = userIds[len(userIds)-1]
		if len(userIds) < articleWatchNotifyBatchSize {
			return
		}
	}
}

func commentNotificationExcludeUserIds(event *CommentCreatedEvent) []uint64 {
	excludeSet := map[uint64]struct{}{}
	add := func(userId uint64) {
		if userId > 0 {
			excludeSet[userId] = struct{}{}
		}
	}
	add(event.UserId)
	add(event.TopicAuthorId)
	add(event.ReplyToPostAuthorId)

	userIds := make([]uint64, 0, len(excludeSet))
	for userId := range excludeSet {
		userIds = append(userIds, userId)
	}
	return userIds
}

// UserFollowedEvent 用户关注事件
type UserFollowedEvent struct {
	UserId       uint64
	FollowerId   uint64
	FollowerName string
}

// handleUserFollowed 发送关注通知
func handleUserFollowed(ctx context.Context, event *UserFollowedEvent) error {
	return eventnotice.SendFollowNotification(event.UserId, event.FollowerId, event.FollowerName)
}

// ArticleLikedEvent 文章点赞事件
type ArticleLikedEvent struct {
	UserId    uint64
	ArticleId uint64
	Title     string
	LikierId  uint64
}
