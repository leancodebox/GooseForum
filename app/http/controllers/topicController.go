package controllers

import (
	"context"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/eventbus"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserAction"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/userFollow"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/eventhandlers"
	"github.com/leancodebox/GooseForum/app/service/fileusageservice"
	"github.com/leancodebox/GooseForum/app/service/postservice"
	"github.com/leancodebox/GooseForum/app/service/userservice"
)

func GetSiteStatistics() component.Response {
	return component.SuccessResponse(hotdataserve.GetSiteStatisticsData())
}

type WriteTopicReq struct {
	TopicId     uint64   `json:"topicId"`
	Content     string   `json:"content" validate:"required"`
	Title       string   `json:"title" validate:"required"`
	Type        int8     `json:"type"`
	CategoryId  []uint64 `json:"categoryId" validate:"min=1,max=3"`
	TopicStatus int8     `json:"topicStatus" validate:"oneof=0 1"`
}

// WriteTopic creates or updates a topic and its first post.
func WriteTopic(req component.BetterRequest[WriteTopicReq]) component.Response {
	// 获取发布设置
	postingConfig := hotdataserve.GetPostingSettingsConfigCache()

	userEntity, err := req.GetUser()
	if err != nil || userEntity.Id == 0 {
		return component.FailResponseCode(component.MessageUserFetchFailed, nil)
	}

	// 统一权限检查
	if _, err := component.CheckUserPermission(&userEntity, component.PermissionActionPost); err != nil {
		return component.FailResponseError(err)
	}

	if len(req.Params.Title) < postingConfig.TextControl.MinTitleLength {
		minLength := postingConfig.TextControl.MinTitleLength
		return component.FailResponseCode(
			component.MessageTopicTitleTooShort,

			component.MessageParams{"minLength": minLength})

	}

	if len(req.Params.Title) > postingConfig.TextControl.MaxTitleLength {
		maxLength := postingConfig.TextControl.MaxTitleLength
		return component.FailResponseCode(
			component.MessageTopicTitleTooLong,

			component.MessageParams{"maxLength": maxLength})

	}

	if len(req.Params.Content) < postingConfig.TextControl.MinPostLength {
		minLength := postingConfig.TextControl.MinPostLength
		return component.FailResponseCode(
			component.MessageTopicContentTooShort,

			component.MessageParams{"minLength": minLength})

	}

	if len(req.Params.Content) > postingConfig.TextControl.MaxPostLength {
		maxLength := postingConfig.TextControl.MaxPostLength
		return component.FailResponseCode(
			component.MessageTopicContentTooLong,

			component.MessageParams{"maxLength": maxLength})

	}

	// 检查新用户冷却时间
	if postingConfig.TextControl.NewUserPostCooldownMinutes > 0 {
		cooldownTime := userEntity.CreatedAt.Add(time.Duration(postingConfig.TextControl.NewUserPostCooldownMinutes) * time.Minute)
		if time.Now().Before(cooldownTime) {
			minutes := postingConfig.TextControl.NewUserPostCooldownMinutes
			availableAt := cooldownTime.Format("2006-01-02 15:04:05")
			return component.FailResponseCode(
				component.MessageTopicPostCooldown,

				component.MessageParams{"minutes": minutes, "availableAt": availableAt})

		}
	}

	if topics.CantWriteNew(req.UserId, 10) {
		return component.FailResponseCode(component.MessageTopicDailyLimit, nil)
	}
	var topic topics.Entity
	var firstPost posts.Entity
	if req.Params.TopicId != 0 {
		topic = topics.Get(req.Params.TopicId)
		if topic.UserId != req.UserId {
			return component.FailResponseCode(component.MessageTopicOwnerMismatch, nil)
		}
		firstPost = posts.Get(topic.FirstPostId)
		if firstPost.Id == 0 {
			firstPost, _ = posts.GetByTopicPostNoAtOrAfter(topic.Id, 1)
		}
	} else {
		topic.UserId = req.UserId
	}
	topic.CategoryIds = req.Params.CategoryId
	topic.Status = req.Params.TopicStatus
	topic.Title = req.Params.Title
	topic.Excerpt = markdown2html.ExtractDescription(req.Params.Content, 200)
	topic.FirstImageURL = markdown2html.ExtractFirstImageURL(req.Params.Content)
	if topic.Id > 0 {
		if firstPost.Id == 0 {
			return component.FailResponseCode(component.MessageTopicNotFound, nil)
		}
		firstPost.Content = req.Params.Content
		firstPost.RenderedHTML = ""
		firstPost.RenderedVersion = markdown2html.GetVersion()
		if err := topics.Save(&topic); err != nil {
			return component.FailResponseCode(component.MessageOperationFailed, nil)
		}
		if err := posts.Save(&firstPost); err != nil {
			return component.FailResponseCode(component.MessageOperationFailed, nil)
		}
		if err := topicCategoryIndex.ReplaceTopicCategories(topic.Id, req.Params.CategoryId); err != nil {
			return component.FailResponseCode(component.MessageOperationFailed, nil)
		}
		fileusageservice.ReplaceTopic(topic.Id, req.UserId, firstPost.Content)
		hotdataserve.ClearTopicListCache()
		if topic.Status == 1 {
			eventbus.Publish(context.Background(), &eventhandlers.TopicUpdatedEvent{Topic: &topic, FirstPost: &firstPost})
		}
	} else {
		topic.PostCount = 1
		topic.PostSeq = 1
		topic.Posters = []topics.Poster{{UserID: req.UserId}}
		if err := topics.Create(&topic); err != nil {
			return component.FailResponseCode(component.MessageOperationFailed, nil)
		}
		firstPost = posts.Entity{
			TopicId:         topic.Id,
			PostNo:          1,
			UserId:          req.UserId,
			Content:         req.Params.Content,
			RenderedHTML:    "",
			RenderedVersion: markdown2html.GetVersion(),
		}
		if err := posts.Create(&firstPost); err != nil {
			return component.FailResponseCode(component.MessageOperationFailed, nil)
		}
		topic.FirstPostId = firstPost.Id
		topic.LastPostId = firstPost.Id
		now := time.Now()
		topic.LastPostedAt = &now
		if err := topics.Save(&topic); err != nil {
			return component.FailResponseCode(component.MessageOperationFailed, nil)
		}
		fileusageservice.ReplaceTopic(topic.Id, req.UserId, firstPost.Content)
		if topic.Status == 1 {
			userStatistics.WriteTopic(req.UserId)
		}
		userservice.InvalidateUserPublicProfileCache(req.UserId)
		if err := topicCategoryIndex.ReplaceTopicCategories(topic.Id, req.Params.CategoryId); err != nil {
			return component.FailResponseCode(component.MessageOperationFailed, nil)
		}
		hotdataserve.ClearTopicListCache()
		if topic.Status == 1 {
			eventbus.Publish(context.Background(), &eventhandlers.TopicPublishedEvent{Topic: &topic, FirstPost: &firstPost})
		}
	}
	return component.SuccessResponse(topic.Id)
}

func normalizeWriteTopicType(topicType int8) int8 {
	switch topicType {
	case 1, 2:
		return topicType
	default:
		return 1
	}
}

type TopicStatusReq struct {
	TopicId     uint64 `json:"topicId" validate:"required"`
	TopicStatus int8   `json:"topicStatus" validate:"oneof=0 1"`
}

func UpdateTopicStatus(req component.BetterRequest[TopicStatusReq]) component.Response {
	topic := topics.Get(req.Params.TopicId)
	if topic.Id == 0 {
		return component.FailResponseCode(component.MessageTopicNotFound, nil)
	}
	if topic.UserId != req.UserId {
		return component.FailResponseCode(component.MessageTopicOperationDenied, nil)
	}
	nextStatus := req.Params.TopicStatus
	if topic.Status == nextStatus {
		return component.SuccessResponse(true)
	}
	topic.Status = nextStatus
	if err := topics.Save(&topic); err != nil {
		return component.FailResponseCode(component.MessageTopicSaveFailed, nil)
	}
	firstPost := posts.Get(topic.FirstPostId)
	hotdataserve.ClearTopicListCache()
	if topic.Status == 1 {
		eventbus.Publish(context.Background(), &eventhandlers.TopicPublishedEvent{Topic: &topic, FirstPost: &firstPost})
	}
	return component.SuccessResponse(true)
}

type CreatePostReq struct {
	TopicId       uint64 `json:"topicId"`
	Content       string `json:"content"`
	ReplyToPostId uint64 `json:"replyToPostId"`
}

func CreatePost(req component.BetterRequest[CreatePostReq]) component.Response {
	// 获取发布设置
	postingConfig := hotdataserve.GetPostingSettingsConfigCache()

	userEntity, err := req.GetUser()
	if err != nil || userEntity.Id == 0 {
		return component.FailResponseCode(component.MessageUserFetchFailed, nil)
	}

	// 统一权限检查
	if _, err := component.CheckUserPermission(&userEntity, component.PermissionActionComment); err != nil {
		return component.FailResponseError(err)
	}

	content := strings.TrimSpace(req.Params.Content)
	if len(content) < postingConfig.TextControl.MinPostLength {
		minLength := postingConfig.TextControl.MinPostLength
		return component.FailResponseCode(
			component.MessageCommentContentTooShort,

			component.MessageParams{"minLength": minLength})

	}

	if len(content) > postingConfig.TextControl.MaxPostLength {
		maxLength := postingConfig.TextControl.MaxPostLength
		return component.FailResponseCode(
			component.MessageCommentContentTooLong,

			component.MessageParams{"maxLength": maxLength})

	}

	// 评论也受发帖冷却限制
	if postingConfig.TextControl.NewUserPostCooldownMinutes > 0 {
		cooldownTime := userEntity.CreatedAt.Add(time.Duration(postingConfig.TextControl.NewUserPostCooldownMinutes) * time.Minute)
		if time.Now().Before(cooldownTime) {
			minutes := postingConfig.TextControl.NewUserPostCooldownMinutes
			availableAt := cooldownTime.Format("2006-01-02 15:04:05")
			return component.FailResponseCode(
				component.MessageCommentPostCooldown,

				component.MessageParams{"minutes": minutes, "availableAt": availableAt})

		}
	}

	topicEntity := topics.GetSimple(req.Params.TopicId)
	if topicEntity.Id == 0 {
		return component.FailResponseCode(component.MessageTopicNotFound, nil)
	}

	var parentPost posts.Entity
	if req.Params.ReplyToPostId > 0 {
		parentPost = posts.Get(req.Params.ReplyToPostId)
		if parentPost.Id == 0 || parentPost.TopicId != req.Params.TopicId || parentPost.PostNo <= 1 {
			return component.FailResponseCode(component.MessageCommentReplyTargetMissed, nil)
		}
	}

	postEntity := &posts.Entity{
		TopicId:         req.Params.TopicId,
		Content:         content,
		RenderedHTML:    markdown2html.CommentMarkdownToHTML(content),
		RenderedVersion: markdown2html.GetCommentVersion(),
		UserId:          req.UserId,
		ReplyToPostId:   req.Params.ReplyToPostId,
	}

	err = postservice.CreateTopicPost(postEntity, topicEntity)
	if err != nil {
		return component.FailResponseCode(
			component.MessageCommentCreateFailed,

			component.MessageParams{"error": err.Error()})

	}
	fileusageservice.ReplacePost(postEntity.Id, req.UserId, postEntity.Content)
	userStatistics.WriteComment(req.UserId)
	userservice.InvalidateUserPublicProfileCache(req.UserId)
	hotdataserve.ClearTopicListCache()

	// 获取父评论作者ID
	var parentReplyAuthorId uint64
	if req.Params.ReplyToPostId > 0 {
		parentReplyAuthorId = parentPost.UserId
	}

	// 发布统一的评论创建事件
	eventbus.Publish(context.Background(), &eventhandlers.CommentCreatedEvent{
		TopicId:             topicEntity.Id,
		PostId:              postEntity.Id,
		UserId:              req.UserId,
		Content:             req.Params.Content,
		TopicAuthorId:       topicEntity.UserId,
		ReplyToPostId:       req.Params.ReplyToPostId,
		ReplyToPostAuthorId: parentReplyAuthorId,
	})

	return component.SuccessResponse(map[string]any{
		"id":              postEntity.Id,
		"postNo":          postEntity.PostNo - 1,
		"renderedContent": postEntity.RenderedHTML,
	})
}

type DeletePostReq struct {
	PostId uint64 `json:"postId"`
}

type UpdatePostReq struct {
	PostId  uint64 `json:"postId"`
	Content string `json:"content"`
}

func UpdatePost(req component.BetterRequest[UpdatePostReq]) component.Response {
	postingConfig := hotdataserve.GetPostingSettingsConfigCache()
	postEntity := posts.Get(req.Params.PostId)
	if postEntity.Id == 0 || postEntity.PostNo <= 1 {
		return component.FailResponseCode(component.MessagePostNotFound, nil)
	}
	if postEntity.UserId != req.UserId {
		return component.FailResponseCode(component.MessageTopicOperationDenied, nil)
	}

	content := strings.TrimSpace(req.Params.Content)
	if len(content) < postingConfig.TextControl.MinPostLength {
		minLength := postingConfig.TextControl.MinPostLength
		return component.FailResponseCode(
			component.MessageCommentContentTooShort,

			component.MessageParams{"minLength": minLength})

	}

	if len(content) > postingConfig.TextControl.MaxPostLength {
		maxLength := postingConfig.TextControl.MaxPostLength
		return component.FailResponseCode(
			component.MessageCommentContentTooLong,

			component.MessageParams{"maxLength": maxLength})

	}

	postEntity.Content = content
	postEntity.RenderedHTML = markdown2html.CommentMarkdownToHTML(content)
	postEntity.RenderedVersion = markdown2html.GetCommentVersion()

	if err := posts.Save(&postEntity); err != nil {
		return component.FailResponseCode(
			component.MessagePostUpdateFailed,

			component.MessageParams{"error": err.Error()})

	}
	fileusageservice.ReplacePost(postEntity.Id, req.UserId, postEntity.Content)

	return component.SuccessResponse(map[string]any{
		"id":              postEntity.Id,
		"postNo":          postEntity.PostNo - 1,
		"content":         postEntity.Content,
		"renderedContent": postEntity.RenderedHTML,
		"updatedAt":       postEntity.UpdatedAt.Format(time.DateTime),
	})
}

func DeletePost(req component.BetterRequest[DeletePostReq]) component.Response {
	postEntity := posts.Get(req.Params.PostId)
	if postEntity.Id == 0 || postEntity.PostNo <= 1 {
		return component.FailResponseCode(component.MessagePostNotFound, nil)
	}
	if postEntity.UserId != req.UserId {
		return component.FailResponseCode(component.MessageTopicOperationDenied, nil)
	}
	posts.DeleteEntity(&postEntity)
	topicEntity := topics.GetSimple(postEntity.TopicId)
	if topicEntity.Id > 0 {
		postservice.SyncTopicPostStats(topicEntity, req.UserId, true)
		hotdataserve.ClearTopicListCache()
	}
	return component.SuccessResponse(true)
}

type LikeTopicReq struct {
	TopicId uint64 `json:"topicId"`
	Action  int    `json:"action" validate:"min=1,max=2"` // 1 点赞，2 取消
}

func LikeTopic(req component.BetterRequest[LikeTopicReq]) component.Response {
	topicEntity := topics.Get(req.Params.TopicId)
	if topicEntity.Id == 0 {
		return component.FailResponseCode(component.MessageTopicNotFound, nil)
	}
	state := topicUserAction.GetByTopicId(req.UserId, topicEntity.Id)
	targetLiked := req.Params.Action == 1
	if state.Id == 0 && !targetLiked {
		return component.SuccessResponse(true)
	}
	if state.Id != 0 && (state.LikedAt != nil) == targetLiked {
		return component.SuccessResponse(true)
	}
	if topicUserAction.SetLiked(req.UserId, topicEntity.Id, targetLiked) {
		if req.Params.Action == 1 {
			topics.IncrementLike(topicEntity)
			userStatistics.LikeTopic(topicEntity.UserId)
			userStatistics.GivenLike(req.UserId)
			userservice.InvalidateUserPublicProfileCache(topicEntity.UserId)
			userservice.InvalidateUserPublicProfileCache(req.UserId)
			hotdataserve.ClearTopicListCache()

			// 发送点赞事件
			eventbus.Publish(context.Background(), &eventhandlers.TopicLikedEvent{
				UserId:  topicEntity.UserId,
				TopicId: topicEntity.Id,
				Title:   topicEntity.Title,
				LikerId: req.UserId,
			})
		} else {
			topics.DecrementLike(topicEntity)
			userStatistics.CancelLikeTopic(topicEntity.UserId)
			userStatistics.CancelGivenLike(req.UserId)
			userservice.InvalidateUserPublicProfileCache(topicEntity.UserId)
			userservice.InvalidateUserPublicProfileCache(req.UserId)
			hotdataserve.ClearTopicListCache()
		}
	}
	return component.SuccessResponse(true)
}

type BookmarkTopicReq struct {
	TopicId uint64 `json:"topicId"`
	Action  int    `json:"action" validate:"min=1,max=2"` // 1 收藏，2 取消
}

func BookmarkTopic(req component.BetterRequest[BookmarkTopicReq]) component.Response {
	topicEntity := topics.Get(req.Params.TopicId)
	if topicEntity.Id == 0 {
		return component.FailResponseCode(component.MessageTopicNotFound, nil)
	}

	state := topicUserAction.GetByTopicId(req.UserId, topicEntity.Id)
	targetBookmarked := req.Params.Action == 1
	if state.Id == 0 && !targetBookmarked {
		return component.SuccessResponse(true)
	}
	if state.Id != 0 && (state.BookmarkedAt != nil) == targetBookmarked {
		return component.SuccessResponse(true)
	}

	if topicUserAction.SetBookmarked(req.UserId, topicEntity.Id, targetBookmarked) {
		updateBookmarkStats(req.UserId, targetBookmarked)
		userservice.InvalidateUserPublicProfileCache(req.UserId)
	}
	return component.SuccessResponse(true)
}

func updateBookmarkStats(userID uint64, bookmarked bool) {
	if bookmarked {
		userStatistics.Collection(userID)
		return
	}
	userStatistics.CancelCollection(userID)
}

type WatchTopicReq struct {
	TopicId uint64 `json:"topicId"`
	Action  int    `json:"action" validate:"min=1,max=2"` // 1 关注，2 取消
}

func WatchTopic(req component.BetterRequest[WatchTopicReq]) component.Response {
	topicEntity := topics.Get(req.Params.TopicId)
	if topicEntity.Id == 0 {
		return component.FailResponseCode(component.MessageTopicNotFound, nil)
	}

	state := topicUserAction.GetByTopicId(req.UserId, topicEntity.Id)
	targetWatched := req.Params.Action == 1
	if state.Id == 0 && !targetWatched {
		return component.SuccessResponse(true)
	}
	if state.Id != 0 && (state.WatchedAt != nil) == targetWatched {
		return component.SuccessResponse(true)
	}

	topicUserAction.SetWatched(req.UserId, topicEntity.Id, targetWatched)
	return component.SuccessResponse(true)
}

type FollowUserReq struct {
	Id     uint64 `json:"id"`
	Action int    `json:"action" validate:"min=1,max=2"` // 1 关注，2 取消
}

func FollowUser(req component.BetterRequest[FollowUserReq]) component.Response {
	userEntity, _ := users.Get(req.Params.Id)
	if userEntity.Id == 0 {
		return component.FailResponseCode(component.MessageUserNotFound, nil)
	}
	userFollowEntity := userFollow.GetByUserId(req.UserId, req.Params.Id)
	if userFollowEntity.Id == 0 {
		userFollowEntity.UserId = req.UserId
		userFollowEntity.FollowUserId = req.Params.Id
	}
	var targetStatus int
	if req.Params.Action == 1 {
		targetStatus = 1
	} else {
		targetStatus = 0
	}

	if userFollowEntity.Status == targetStatus {
		return component.SuccessResponse(true)
	}
	userFollowEntity.Status = targetStatus
	if userFollow.SaveOrCreateById(&userFollowEntity) > 0 {
		if req.Params.Action == 1 {
			userStatistics.Following(req.UserId)
			userStatistics.Follower(req.Params.Id)
			userservice.InvalidateUserPublicProfileCache(req.UserId)
			userservice.InvalidateUserPublicProfileCache(req.Params.Id)

			// 发送关注通知
			followerUser, _ := req.GetUser()
			eventbus.Publish(context.Background(), &eventhandlers.UserFollowedEvent{
				UserId:       req.Params.Id,
				FollowerId:   req.UserId,
				FollowerName: followerUser.Username,
			})
		} else {
			userStatistics.CancelFollowing(req.UserId)
			userStatistics.CancelFollower(req.Params.Id)
			userservice.InvalidateUserPublicProfileCache(req.UserId)
			userservice.InvalidateUserPublicProfileCache(req.Params.Id)
		}
	}
	return component.SuccessResponse(true)
}
