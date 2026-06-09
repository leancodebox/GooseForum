package controllers

import (
	"context"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/eventbus"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/articleBookmark"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategoryRs"
	"github.com/leancodebox/GooseForum/app/models/forum/articleLike"
	"github.com/leancodebox/GooseForum/app/models/forum/articleWatch"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/userFollow"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/eventhandlers"
	"github.com/leancodebox/GooseForum/app/service/replyservice"
	"github.com/leancodebox/GooseForum/app/service/userservice"
	"github.com/samber/lo"
)

func GetSiteStatistics() component.Response {
	return component.SuccessResponse(hotdataserve.GetSiteStatisticsData())
}

type WriteArticleReq struct {
	Id            int64    `json:"id"`
	Content       string   `json:"content" validate:"required"`
	Title         string   `json:"title" validate:"required"`
	Type          int8     `json:"type"`
	CategoryId    []uint64 `json:"categoryId" validate:"min=1,max=3"`
	ArticleStatus int8     `json:"articleStatus" validate:"oneof=0 1"`
}

// WriteArticles 创建或更新文章。
func WriteArticles(req component.BetterRequest[WriteArticleReq]) component.Response {
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
			component.MessageArticleTitleTooShort,

			component.MessageParams{"minLength": minLength})

	}

	if len(req.Params.Title) > postingConfig.TextControl.MaxTitleLength {
		maxLength := postingConfig.TextControl.MaxTitleLength
		return component.FailResponseCode(
			component.MessageArticleTitleTooLong,

			component.MessageParams{"maxLength": maxLength})

	}

	if len(req.Params.Content) < postingConfig.TextControl.MinPostLength {
		minLength := postingConfig.TextControl.MinPostLength
		return component.FailResponseCode(
			component.MessageArticleContentTooShort,

			component.MessageParams{"minLength": minLength})

	}

	if len(req.Params.Content) > postingConfig.TextControl.MaxPostLength {
		maxLength := postingConfig.TextControl.MaxPostLength
		return component.FailResponseCode(
			component.MessageArticleContentTooLong,

			component.MessageParams{"maxLength": maxLength})

	}

	// 检查新用户冷却时间
	if postingConfig.TextControl.NewUserPostCooldownMinutes > 0 {
		cooldownTime := userEntity.CreatedAt.Add(time.Duration(postingConfig.TextControl.NewUserPostCooldownMinutes) * time.Minute)
		if time.Now().Before(cooldownTime) {
			minutes := postingConfig.TextControl.NewUserPostCooldownMinutes
			availableAt := cooldownTime.Format("2006-01-02 15:04:05")
			return component.FailResponseCode(
				component.MessageArticlePostCooldown,

				component.MessageParams{"minutes": minutes, "availableAt": availableAt})

		}
	}

	if articles.CantWriteNew(req.UserId, 10) {
		return component.FailResponseCode(component.MessageArticleDailyLimit, nil)
	}
	var article articles.Entity
	if req.Params.Id != 0 {
		article = articles.Get(req.Params.Id)
		if article.UserId != req.UserId {
			return component.FailResponseCode(component.MessageArticleOwnerMismatch, nil)
		}
	} else {
		article.UserId = req.UserId
		article.Type = req.Params.Type
	}
	article.CategoryId = req.Params.CategoryId
	article.ArticleStatus = req.Params.ArticleStatus
	article.Content = req.Params.Content
	article.Title = req.Params.Title
	// 自动生成文章描述
	article.Description = markdown2html.ExtractDescription(req.Params.Content, 200)
	article.FirstImageURL = markdown2html.ExtractFirstImageURL(req.Params.Content)
	article.RenderedVersion = markdown2html.GetVersion()
	article.RenderedHTML = "" // 用户提交后不用渲染，避免提交时间过长。
	if article.Id > 0 {
		if err := articles.Save(&article); err != nil {
			return component.FailResponseCode(component.MessageOperationFailed, nil)
		}
		categoryIDMap := lo.SliceToMap(req.Params.CategoryId, func(id uint64) (uint64, bool) {
			return id, true
		})

		for _, item := range articleCategoryRs.GetByArticleId(article.Id) {
			if _, ok := categoryIDMap[item.ArticleCategoryId]; ok {
				item.Effective = 1
				articleCategoryRs.SaveOrCreateById(item)
				// 如果已经存在，从 map 中删除，避免重复插入
				delete(categoryIDMap, item.ArticleCategoryId)
			} else {
				item.Effective = 0
				articleCategoryRs.SaveOrCreateById(item)
			}
		}
		// 插入新的条目
		for id := range categoryIDMap {
			rs := &articleCategoryRs.Entity{ArticleId: article.Id, ArticleCategoryId: id, Effective: 1}
			articleCategoryRs.SaveOrCreateById(rs)
		}
		if article.ArticleStatus == 1 {
			eventbus.Publish(context.Background(), &eventhandlers.ArticleUpdatedEvent{Article: &article})
		}
	} else {
		articles.Create(&article)
		if article.ArticleStatus == 1 {
			userStatistics.WriteArticle(req.UserId)
		}
		userservice.InvalidateUserPublicProfileCache(req.UserId)
		lo.ForEach(req.Params.CategoryId, func(item uint64, _ int) {
			rs := articleCategoryRs.Entity{ArticleId: article.Id, ArticleCategoryId: item, Effective: 1}
			articleCategoryRs.SaveOrCreateById(&rs)
		})
		if article.ArticleStatus == 1 {
			eventbus.Publish(context.Background(), &eventhandlers.ArticlePublishedEvent{Article: &article})
		}
	}
	return component.SuccessResponse(article.Id)
}

type ArticleStatusReq struct {
	Id            uint64 `json:"id" validate:"required"`
	ArticleStatus int8   `json:"articleStatus" validate:"oneof=0 1"`
}

func UpdateArticleStatus(req component.BetterRequest[ArticleStatusReq]) component.Response {
	article := articles.Get(req.Params.Id)
	if article.Id == 0 {
		return component.FailResponseCode(component.MessageArticleNotFound, nil)
	}
	if article.UserId != req.UserId {
		return component.FailResponseCode(component.MessageArticleOperationDenied, nil)
	}
	if article.ArticleStatus == req.Params.ArticleStatus {
		return component.SuccessResponse(true)
	}
	article.ArticleStatus = req.Params.ArticleStatus
	if err := articles.Save(&article); err != nil {
		return component.FailResponseCode(component.MessageArticleSaveFailed, nil)
	}
	if article.ArticleStatus == 1 {
		eventbus.Publish(context.Background(), &eventhandlers.ArticlePublishedEvent{Article: &article})
	}
	return component.SuccessResponse(true)
}

type ArticleReplyId struct {
	ArticleId uint64 `json:"articleId"`
	Content   string `json:"content"`
	ReplyId   uint64 `json:"replyId"`
}

func ArticleReply(req component.BetterRequest[ArticleReplyId]) component.Response {
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

	articleEntity := articles.GetSimple(req.Params.ArticleId)
	if articleEntity.Id == 0 {
		return component.FailResponseCode(component.MessageArticleNotFound, nil)
	}

	var parentReply reply.Entity
	if req.Params.ReplyId > 0 {
		parentReply = reply.Get(req.Params.ReplyId)
		if parentReply.Id == 0 || parentReply.ArticleId != req.Params.ArticleId {
			return component.FailResponseCode(component.MessageCommentReplyTargetMissed, nil)
		}
	}

	replyEntity := &reply.Entity{
		ArticleId:       req.Params.ArticleId,
		Content:         content,
		RenderedHTML:    markdown2html.CommentMarkdownToHTML(content),
		RenderedVersion: markdown2html.GetCommentVersion(),
		UserId:          req.UserId,
		ReplyId:         req.Params.ReplyId,
	}

	err = replyservice.CreateArticleReply(replyEntity, articleEntity)
	if err != nil {
		return component.FailResponseCode(
			component.MessageCommentCreateFailed,

			component.MessageParams{"error": err.Error()})

	}
	userStatistics.WriteComment(req.UserId)
	userservice.InvalidateUserPublicProfileCache(req.UserId)

	// 获取父评论作者ID
	var parentReplyAuthorId uint64
	if req.Params.ReplyId > 0 {
		parentReplyAuthorId = parentReply.UserId
	}

	// 发布统一的评论创建事件
	eventbus.Publish(context.Background(), &eventhandlers.CommentCreatedEvent{
		ArticleId:           articleEntity.Id,
		CommentId:           replyEntity.Id,
		UserId:              req.UserId,
		Content:             req.Params.Content,
		ArticleAuthorId:     articleEntity.UserId,
		ParentReplyId:       req.Params.ReplyId,
		ParentReplyAuthorId: parentReplyAuthorId,
	})

	return component.SuccessResponse(map[string]any{
		"id":              replyEntity.Id,
		"replyNo":         replyEntity.ReplyNo,
		"renderedContent": replyEntity.RenderedHTML,
	})
}

type DeleteReplyId struct {
	ReplyId uint64 `json:"replyId"`
}

type UpdateReplyReq struct {
	ReplyId uint64 `json:"replyId"`
	Content string `json:"content"`
}

func UpdateReply(req component.BetterRequest[UpdateReplyReq]) component.Response {
	postingConfig := hotdataserve.GetPostingSettingsConfigCache()
	replyEntity := reply.Get(req.Params.ReplyId)
	if replyEntity.Id == 0 {
		return component.FailResponseCode(component.MessageReplyNotFound, nil)
	}
	if replyEntity.UserId != req.UserId {
		return component.FailResponseCode(component.MessageArticleOperationDenied, nil)
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

	replyEntity.Content = content
	replyEntity.RenderedHTML = markdown2html.CommentMarkdownToHTML(content)
	replyEntity.RenderedVersion = markdown2html.GetCommentVersion()

	if err := reply.Save(&replyEntity); err != nil {
		return component.FailResponseCode(
			component.MessageReplyUpdateFailed,

			component.MessageParams{"error": err.Error()})

	}

	return component.SuccessResponse(map[string]any{
		"id":              replyEntity.Id,
		"replyNo":         replyEntity.ReplyNo,
		"content":         replyEntity.Content,
		"renderedContent": replyEntity.RenderedHTML,
		"updatedAt":       replyEntity.UpdatedAt.Format(time.DateTime),
	})
}

func DeleteReply(req component.BetterRequest[DeleteReplyId]) component.Response {
	replyEntity := reply.Get(req.Params.ReplyId)
	if replyEntity.Id == 0 {
		return component.FailResponseCode(component.MessageReplyNotFound, nil)
	}
	if replyEntity.UserId != req.UserId {
		return component.FailResponseCode(component.MessageArticleOperationDenied, nil)
	}
	reply.DeleteEntity(&replyEntity)
	articleEntity := articles.GetSimple(replyEntity.ArticleId)
	if articleEntity.Id > 0 {
		replyservice.SyncArticleReplyStats(articleEntity, req.UserId, true)
	}
	return component.SuccessResponse(true)
}

type LikeArticleReq struct {
	Id     uint64 `json:"id"`
	Action int    `json:"action" validate:"min=1,max=2"` // 1 点赞，2 取消
}

func LikeArticle(req component.BetterRequest[LikeArticleReq]) component.Response {
	articleEntity := articles.Get(req.Params.Id)
	if articleEntity.Id == 0 {
		return component.FailResponseCode(component.MessageArticleNotFound, nil)
	}
	oldLike := articleLike.GetByArticleId(req.UserId, articleEntity.Id)
	var targetStatus int
	if req.Params.Action == 1 {
		if oldLike.Id == 0 {
			oldLike.UserId = req.UserId
			oldLike.ArticleId = articleEntity.Id
		}
		targetStatus = 1
	} else {
		if oldLike.Id == 0 {
			return component.SuccessResponse(true)
		}
		targetStatus = 0
	}
	if oldLike.Status == targetStatus {
		return component.SuccessResponse(true)
	}
	oldLike.Status = targetStatus
	if articleLike.SaveOrCreateById(&oldLike) > 0 {
		if req.Params.Action == 1 {
			articles.IncrementLike(articleEntity)
			userStatistics.LikeArticle(articleEntity.UserId)
			userStatistics.GivenLike(req.UserId)
			userservice.InvalidateUserPublicProfileCache(articleEntity.UserId)
			userservice.InvalidateUserPublicProfileCache(req.UserId)

			// 发送点赞事件
			eventbus.Publish(context.Background(), &eventhandlers.ArticleLikedEvent{
				UserId:    articleEntity.UserId,
				ArticleId: articleEntity.Id,
				Title:     articleEntity.Title,
				LikierId:  req.UserId,
			})
		} else {
			articles.DecrementLike(articleEntity)
			userStatistics.CancelLikeArticle(articleEntity.UserId)
			userStatistics.CancelGivenLike(req.UserId)
			userservice.InvalidateUserPublicProfileCache(articleEntity.UserId)
			userservice.InvalidateUserPublicProfileCache(req.UserId)
		}
	}
	return component.SuccessResponse(true)
}

type BookmarkArticleReq struct {
	Id     uint64 `json:"id"`
	Action int    `json:"action" validate:"min=1,max=2"` // 1 收藏，2 取消
}

func BookmarkArticle(req component.BetterRequest[BookmarkArticleReq]) component.Response {
	articleEntity := articles.Get(req.Params.Id)
	if articleEntity.Id == 0 {
		return component.FailResponseCode(component.MessageArticleNotFound, nil)
	}

	bookmark := articleBookmark.GetByArticleId(req.UserId, articleEntity.Id)
	targetStatus := bookmarkTargetStatus(req.Params.Action)
	if bookmark.Id == 0 && targetStatus == 0 {
		return component.SuccessResponse(true)
	}
	if bookmark.Id != 0 && bookmark.Status == targetStatus {
		return component.SuccessResponse(true)
	}

	bookmark.UserId = req.UserId
	bookmark.ArticleId = articleEntity.Id
	bookmark.Status = targetStatus
	articleBookmark.SaveOrCreateById(&bookmark)
	updateBookmarkStats(req.UserId, targetStatus)
	userservice.InvalidateUserPublicProfileCache(req.UserId)
	return component.SuccessResponse(true)
}

func bookmarkTargetStatus(action int) int {
	if action == 1 {
		return 1
	}
	return 0
}

func updateBookmarkStats(userID uint64, targetStatus int) {
	if targetStatus == 1 {
		userStatistics.Collection(userID)
		return
	}
	userStatistics.CancelCollection(userID)
}

type WatchArticleReq struct {
	Id     uint64 `json:"id"`
	Action int    `json:"action" validate:"min=1,max=2"` // 1 关注，2 取消
}

func WatchArticle(req component.BetterRequest[WatchArticleReq]) component.Response {
	articleEntity := articles.Get(req.Params.Id)
	if articleEntity.Id == 0 {
		return component.FailResponseCode(component.MessageArticleNotFound, nil)
	}

	watchEntity := articleWatch.GetByArticleId(req.UserId, articleEntity.Id)
	if watchEntity.Id == 0 {
		watchEntity.UserId = req.UserId
		watchEntity.ArticleId = articleEntity.Id
	}

	targetStatus := 0
	if req.Params.Action == 1 {
		targetStatus = 1
	}
	if watchEntity.Status == targetStatus {
		return component.SuccessResponse(true)
	}

	watchEntity.Status = targetStatus
	articleWatch.SaveOrCreateById(&watchEntity)
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
