package eventhandlers

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/leancodebox/GooseForum/app/service/eventnotice"
)

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
	UserId              uint64 // 评论者ID
	Content             string // 评论内容
	ArticleAuthorId     uint64 // 文章作者ID
	ParentReplyId       uint64 // 父评论ID（如果是回复评论）
	ParentReplyAuthorId uint64 // 父评论作者ID
}

// CommentCreatedHandler 评论/回复创建处理器
func NewCommentCreatedHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		"CommentCreatedHandler",
		func(ctx context.Context, event *CommentCreatedEvent) error {
			contentPreview := TakeUpTo64Chars(event.Content)
			// 如果不是文章作者自己评论，通知文章作者
			if event.ArticleAuthorId != event.UserId {
				_ = eventnotice.SendCommentNotification(event.ArticleAuthorId, event.ArticleId, contentPreview, event.UserId, event.CommentId)
			}
			// 如果是回复评论，且不是回复自己，通知原评论作者
			if event.ParentReplyId > 0 && event.ParentReplyAuthorId != event.UserId {
				_ = eventnotice.SendReplyNotification(event.ParentReplyAuthorId, event.ParentReplyId, event.ArticleId, contentPreview, event.UserId)
			}
			return nil
		},
	)
}

// UserFollowedEvent 用户关注事件
type UserFollowedEvent struct {
	UserId       uint64
	FollowerId   uint64
	FollowerName string
}

// UserFollowedHandler 用户关注处理器
func NewUserFollowedHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		"UserFollowedHandler",
		func(ctx context.Context, event *UserFollowedEvent) error {
			return eventnotice.SendFollowNotification(event.UserId, event.FollowerName)
		},
	)
}

// ArticleLikedEvent 文章点赞事件
type ArticleLikedEvent struct {
	UserId    uint64
	ArticleId uint64
	Title     string
	LikierId  uint64
}

// ArticleLikedHandler 文章点赞处理器
func NewArticleLikedHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		"ArticleLikedHandler",
		func(ctx context.Context, event *ArticleLikedEvent) error {
			// 目前点赞暂不发送通知，仅预留
			return nil
		},
	)
}
