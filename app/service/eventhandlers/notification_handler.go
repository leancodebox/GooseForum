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
	ArticleId           uint64
	CommentId           uint64 // 新创建的评论ID
	TopicId             uint64
	PostId              uint64 // 新创建的正文ID
	UserId              uint64 // 评论者ID
	Content             string // 评论内容
	ArticleAuthorId     uint64 // 文章作者ID
	ParentReplyId       uint64 // 父评论ID（如果是回复评论）
	ParentReplyAuthorId uint64 // 父评论作者ID
}

func (event *CommentCreatedEvent) topicID() uint64 {
	if event.TopicId > 0 {
		return event.TopicId
	}
	return event.ArticleId
}

func (event *CommentCreatedEvent) postID() uint64 {
	if event.PostId > 0 {
		return event.PostId
	}
	return event.CommentId
}

// handleCommentCreated 发送评论/回复通知
func handleCommentCreated(ctx context.Context, event *CommentCreatedEvent) error {
	contentPreview := TakeUpTo64Chars(event.Content)
	// 如果不是文章作者自己评论，通知文章作者
	if shouldNotifyArticleAuthor(event) {
		_ = eventnotice.SendCommentNotification(event.ArticleAuthorId, event.topicID(), contentPreview, event.UserId, event.postID())
	}
	// 如果是回复评论，且不是回复自己，通知原评论作者
	if shouldNotifyParentReplyAuthor(event) {
		_ = eventnotice.SendReplyNotification(event.ParentReplyAuthorId, event.postID(), event.topicID(), contentPreview, event.UserId)
	}
	notifyArticleWatchers(event, contentPreview)
	return nil
}

func shouldNotifyArticleAuthor(event *CommentCreatedEvent) bool {
	if event.ArticleAuthorId == 0 || event.ArticleAuthorId == event.UserId {
		return false
	}
	return event.ParentReplyId == 0 || event.ArticleAuthorId != event.ParentReplyAuthorId
}

func shouldNotifyParentReplyAuthor(event *CommentCreatedEvent) bool {
	return event.ParentReplyId > 0 && event.ParentReplyAuthorId > 0 && event.ParentReplyAuthorId != event.UserId
}

func notifyArticleWatchers(event *CommentCreatedEvent, contentPreview string) {
	excludeUserIds := commentNotificationExcludeUserIds(event)
	afterUserId := uint64(0)
	for {
		userIds := topicUserAction.ListActiveWatchUserIDsAfter(event.topicID(), afterUserId, excludeUserIds, articleWatchNotifyBatchSize)
		if len(userIds) == 0 {
			return
		}
		_ = eventnotice.SendArticleCommentNotifications(userIds, event.topicID(), event.postID(), contentPreview, event.UserId)
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
	add(event.ArticleAuthorId)
	add(event.ParentReplyAuthorId)

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
