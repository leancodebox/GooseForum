package eventhandlers

import (
	"context"

	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserAction"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"github.com/leancodebox/GooseForum/app/service/notificationservice"
)

const topicWatchNotifyBatchSize = 500

// TakeUpTo64Chars 按字符数截取字符串，最多取 64 个字符
func TakeUpTo64Chars(s string) string {
	return markdown2html.ExtractPreview(s, 64)
}

// CommentCreatedEvent 评论/回复创建事件
type CommentCreatedEvent struct {
	TopicId             uint64
	PostId              uint64 // 新创建的 post ID
	PostNo              uint64 // 新创建的 post 楼层号
	UserId              uint64 // 发表评论者 ID
	Content             string
	TopicAuthorId       uint64 // 主题作者 ID
	ReplyToPostId       uint64 // 被回复的 post ID
	ReplyToPostAuthorId uint64 // 被回复的 post 作者 ID
}

// handleCommentCreated 发送评论/回复通知
func handleCommentCreated(ctx context.Context, event *CommentCreatedEvent) error {
	contentPreview := TakeUpTo64Chars(event.Content)
	topic := topics.GetSimple(event.TopicId)
	plan := buildCommentNotificationPlan(event)
	// 如果不是主题作者自己发表评论，通知主题作者
	if plan.notifyTopicAuthor && canReceiveTopicNotification(event.TopicAuthorId, topic) {
		_ = notificationservice.SendCommentNotification(event.TopicAuthorId, event.TopicId, contentPreview, event.UserId, event.PostId, event.PostNo)
	}
	// 如果是回复 post，且不是回复自己，通知原 post 作者
	if plan.notifyParentReplyAuthor && canReceiveTopicNotification(event.ReplyToPostAuthorId, topic) {
		_ = notificationservice.SendPostReplyNotification(event.ReplyToPostAuthorId, event.PostId, event.PostNo, event.TopicId, contentPreview, event.UserId)
	}
	notifyTopicWatchers(event, topic, contentPreview, plan.watcherExcludeUserIDs)
	return nil
}

type commentNotificationPlan struct {
	notifyTopicAuthor       bool
	notifyParentReplyAuthor bool
	watcherExcludeUserIDs   []uint64
}

// buildCommentNotificationPlan gives direct notifications priority over the
// broader watched-topic notification. A user present in both audiences gets
// only the more specific notification.
func buildCommentNotificationPlan(event *CommentCreatedEvent) commentNotificationPlan {
	plan := commentNotificationPlan{
		notifyTopicAuthor:       shouldNotifyTopicAuthor(event),
		notifyParentReplyAuthor: shouldNotifyParentReplyAuthor(event),
	}
	excludeSet := map[uint64]struct{}{}
	add := func(userID uint64) {
		if userID > 0 {
			excludeSet[userID] = struct{}{}
		}
	}
	add(event.UserId)
	if plan.notifyTopicAuthor {
		add(event.TopicAuthorId)
	}
	if plan.notifyParentReplyAuthor {
		add(event.ReplyToPostAuthorId)
	}
	plan.watcherExcludeUserIDs = make([]uint64, 0, len(excludeSet))
	for userID := range excludeSet {
		plan.watcherExcludeUserIDs = append(plan.watcherExcludeUserIDs, userID)
	}
	return plan
}

func shouldNotifyTopicAuthor(event *CommentCreatedEvent) bool {
	if event.TopicAuthorId == 0 || event.TopicAuthorId == event.UserId {
		return false
	}
	return event.ReplyToPostId == 0 || event.TopicAuthorId != event.ReplyToPostAuthorId
}

func shouldNotifyParentReplyAuthor(event *CommentCreatedEvent) bool {
	return event.ReplyToPostId > 0 && event.ReplyToPostAuthorId > 0 && event.ReplyToPostAuthorId != event.UserId
}

func notifyTopicWatchers(event *CommentCreatedEvent, topic topics.Entity, contentPreview string, excludeUserIDs []uint64) {
	if topic.Id == 0 || topic.Status != 1 || topic.ProcessStatus != 0 {
		return
	}
	afterUserId := uint64(0)
	for {
		userIds := topicUserAction.ListActiveWatchUserIDsAfter(event.TopicId, afterUserId, excludeUserIDs, topicWatchNotifyBatchSize)
		if len(userIds) == 0 {
			return
		}
		readableUserIDs, err := accesscontrol.FilterReadableUserIDs(userIds, topic.MainCategoryId)
		if err != nil {
			return
		}
		_ = notificationservice.SendTopicPostNotifications(readableUserIDs, event.TopicId, event.PostId, event.PostNo, contentPreview, event.UserId)
		afterUserId = userIds[len(userIds)-1]
		if len(userIds) < topicWatchNotifyBatchSize {
			return
		}
	}
}

func canReceiveTopicNotification(userID uint64, topic topics.Entity) bool {
	if userID == 0 || topic.Id == 0 || topic.Status != 1 || topic.ProcessStatus != 0 {
		return false
	}
	snapshot, err := accesscontrol.Resolve(userID)
	return err == nil && snapshot.CanReadCategory(topic.MainCategoryId)
}

// UserFollowedEvent 用户关注事件
type UserFollowedEvent struct {
	UserId       uint64
	FollowerId   uint64
	FollowerName string
}

// handleUserFollowed 发送关注通知
func handleUserFollowed(ctx context.Context, event *UserFollowedEvent) error {
	return notificationservice.SendFollowNotification(event.UserId, event.FollowerId, event.FollowerName)
}

// TopicLikedEvent 主题点赞事件
type TopicLikedEvent struct {
	UserId  uint64
	TopicId uint64
	Title   string
	LikerId uint64
}
